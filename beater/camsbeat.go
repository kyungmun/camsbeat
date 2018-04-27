package beater

import (
	"fmt"
	"time"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"net/http"
	"net"
	"log"

	"github.com/kyungmun/camsbeat/config"
	"log/syslog"
	//golang 제공 라이브러리는 windows os는 빌드예외 처리되어 사용을 못하게 되는데, syslog.go 파일을 열어 상단의 빌드예외처리를 제거하고 사용. // +build windows,!nacl,!plan9
	//"github.com/elastic/beats/libbeat/cfgfile"
	//"github.com/kyungmun/camsbeat/mb/module"
	//"github.com/Masterminds/glide/cfg"
	"strings"
	"net/url"
	"strconv"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"github.com/Shopify/sarama"
)

type Camsbeat struct {
	done          chan struct{}
	config        config.Config
	client        beat.Client
	period        time.Duration
	lastIndexTime time.Time
	logWriter2    *syslog.Writer
	cfg           *common.Config
	kafkaProducer sarama.SyncProducer
}

/*
var (
	brokers    = kingpin.Flag("brokers", "List of brokers to connect").Default("10.165.60.52:9092").Strings()
	topic      = kingpin.Flag("topic", "Topic name").Default("camsbeat").String()
)
*/


// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	log.Println("start camsbeat")

	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	logp.Info("config.name :" + config.Name);
	logp.Info("config.kinds :" + config.Kinds);
	logp.Info("config.period : " + config.Period.String());
	logp.Info("config.values : " + strings.Join(config.Values, ","));

	logWriter, err := syslog.Dial("udp", "10.165.60.64:514", syslog.LOG_ERR, "camsbeat Logger")
	defer logWriter.Close()
	if err != nil {
		logp.Err("syslog create error")
	} else {
		logWriter.Alert("alert")
		logWriter.Crit("critical")
		logWriter.Err("error")
		logWriter.Warning("warning")
		logWriter.Notice("notice")
		logWriter.Info("information")
		logWriter.Debug("debug")
		logWriter.Write([]byte("Hello Logger!"))
	}

	/*/kafka direct send (producer) -------------------------------//
	kingpin.Parse()
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 5
	kafkaConfig.Producer.Return.Successes = true

	for _, v := range *brokerList {
		fmt.Println("kafka broker List : " + v)
		logp.Info("kafka broker List : %s", v)
	}
	fmt.Println("kafka topic : " + *topic)

	producer, err := sarama.NewSyncProducer(*brokerList, kafkaConfig)

	if err != nil {
		logp.Err("producer create err : %s", err)
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			logp.Err("close err : %s", err)
			panic(err)
		}
	}()
	*/


	bt := &Camsbeat{
		done:   make(chan struct{}),
		config: config,
		logWriter2: logWriter,
		cfg: cfg,
		kafkaProducer: newDataCollector(config),
	}

	return bt, nil
}

func (bt *Camsbeat) Close() error {
	logp.Info("close camsbeat")
	fmt.Println("close camsbeat")

	if err := bt.client.Close(); err != nil {
		logp.Err("Failed to shut down beat client cleanly : %s", err)
	}

	if bt.kafkaProducer != nil {
		if  err := bt.kafkaProducer.Close(); err != nil {
			logp.Err("Failed to shut down kafkaProducer cleanly : %s", err)
		}
	}

	if err := bt.logWriter2.Close(); err != nil {
		logp.Err("Failed to shut down logWriter2 cleanly : %s", err)
	}

	return nil
}

func newDataCollector(cfg  config.Config) sarama.SyncProducer {

	var brokers = kingpin.Flag("brokers", "List of brokers to connect").Default(cfg.KafkaBroker).Strings()
	var topic = kingpin.Flag("topic", "Topic name").Default(cfg.KafkaTopic).String()

	kingpin.Parse()

	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	if cfg.KafkaAuthEnabled {
		//config.Net.SASL.Handshake = false
		config.Net.SASL.Enable = true
		config.Net.SASL.User = cfg.KafkaSaslUser
		config.Net.SASL.Password = cfg.KafkaSaslPass
		fmt.Println("kafka user : " + cfg.KafkaSaslUser)
		fmt.Println("kafka pass : " + cfg.KafkaSaslPass)
	}

	/*
	tlsConfig := createTlsConfiguration()
	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}
	*/

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	for _, v := range *brokers {
		fmt.Println("kafka broker List : " + v)
		logp.Info("kafka broker List : %s", v)
	}
	fmt.Println("kafka topic : " + *topic)
	logp.Info("kafka topic : %s", *topic)

	producer, err := sarama.NewSyncProducer(*brokers, config)
	if err != nil {
		fmt.Println("Failed to create direct producer:", err)
		logp.Err("Failed to create direct  producer: %s", err)
		producer = nil
	}

	return producer
}

func (bt *Camsbeat) Run(b *beat.Beat) error {
	logp.Info("camsbeat is running! Hit CTRL-C to stop it. %s", bt.lastIndexTime)
	var err error

	bt.client, err = b.Publisher.Connect()

	if err != nil {
		logp.Err("outout config client create error")
	}

	if bt.kafkaProducer == nil {
		fmt.Println("Kafka direct producer : nil")
		logp.Err("Kafka direct producer : nil")
	}

	ticker := time.NewTicker(bt.config.Period)

	//counter := 1
	logp.Info("start ... %s", bt.config.Kinds)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
			logp.Info("ticket ... ")
			/*
			config := config.DefaultConfig
			if err := bt.cfg.Unpack(&config); err != nil {
				logp.Info("config reload  ..fail")
			} else {
				bt.config = config
				logp.Info("reload enable true  %s", bt.config.Path)
			}
			*/
		}

		for i, url := range bt.config.Values {
			if bt.kafkaProducer != nil {
				go bt.DirectKafkaSend(url, bt.config.ObjectId[i], bt.config.AccountId[i])
			} else {
				go bt.OutputConfigSend(url, bt.config.ObjectId[i], bt.config.AccountId[i])
			}
		}
	}

	return nil
}

func (bt *Camsbeat) Stop() {
	bt.Close()
	close(bt.done)
}

func (bt *Camsbeat) OutputConfigSend(urlValue string, objectId string, accountId string) {

	t := time.Now().UTC()
	localTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	//timeValue := localTime.Round(time.Millisecond).Format("2006-01-02T15:04:05.999Z")

	logp.Info("OutputConfigSend Check url : [%s]", urlValue)
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	//시작시간
	var res *http.Response
	time_start := time.Now()
	res, err := client.Head(urlValue)

	var responseTime time.Duration
	var resultValue bool = true
	var errorMsg string = ""
	var errorKind string = ""
	if err != nil {
		logp.Info("[err] %s", err);
		errorMsg = err.Error()
		errorKind = "socket"
		responseTime = time.Since(time_start);
		res = &http.Response{
			StatusCode : 408,
		}
	} else {
		responseTime = time.Since(time_start);
		logp.Info("[response] %s %s %s %d", res.Request.URL.Host, res.Request.URL.Port(), responseTime.String(), res.StatusCode)
	}

	resultValue = (res.StatusCode == 200)

	errorInfo := common.MapStr{
		"message" : errorMsg,
		"kinds": errorKind,
	};

	agentInfo := common.MapStr{
		"name" : bt.config.Name,
		"version": "v1",
		"country": "KR",
	};

	respInfo := common.MapStr{
		//"us" : responseTime / time.Microsecond,
		"ms" : responseTime / time.Millisecond,
		"status": res.StatusCode,
	};

	u, err := url.Parse(urlValue)
	if err != nil {
		logp.Err("%s", err)
	}
	hostIp, _ := net.LookupIP(u.Hostname())

	portValue := 80
	if u.Port() != "" {
		portValue, _ = strconv.Atoi(u.Port())
	}

	event := beat.Event{
		Timestamp: time.Now(),
		Fields: common.MapStr{
			"type_name":  "url",
			"object_id":  objectId,
			"account_id":  strings.ToLower(accountId), //소문자로 치환된 account_id 값
			"createtime":   localTime,
			"kinds":      "url",
			"url":  urlValue,
			"host":  u.Hostname(),
			"ip" : hostIp[0],
			"port" : portValue,
			"up" : resultValue,
			"agent" : agentInfo,
			"response" : respInfo,
			"error" : errorInfo,
		},
	}
	//config output send
	bt.client.Publish(event)

	//syslog send
	//bt.logWriter2.Write([]byte(event.Fields.String()))
}


func (bt *Camsbeat) DirectKafkaSend(urlValue string, objectId string, accountId string) {

	//location, _ := time.LoadLocation("Asia/Seoul")
	t := time.Now().UTC()
	localTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	timeValue := localTime.Round(time.Millisecond).Format("2006-01-02T15:04:05.999Z")

	logp.Info("DirectKafkaSend Check url : [%s]", urlValue)
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	//요청 시작시간
	var res *http.Response
	time_start := time.Now()
	res, err := client.Head(urlValue)

	var responseTime time.Duration
	var resultValue bool = true
	var errorMsg string = ""
	var errorKind string = ""
	if err != nil {
		logp.Info("[err] %s", err);
		errorMsg = err.Error()
		errorKind = "socket"
		resultValue = false
		responseTime = time.Since(time_start);
		res = &http.Response{
			StatusCode : 408,
		}
	} else {
		responseTime = time.Since(time_start);
		logp.Info("[response] %s %s %s %d", res.Request.URL.Host, res.Request.URL.Port(), responseTime.String(), res.StatusCode)
	}

	resultValue = (res.StatusCode == 200)

	errorInfo := common.MapStr{
		"message" : errorMsg,
		"kinds": errorKind,
	};

	agentInfo := common.MapStr{
		"name" : bt.config.Name,
		"version": "v1",
		"country": "KR",
	};

	respInfo := common.MapStr{
		"us" : responseTime / time.Microsecond,
		"ms" : responseTime / time.Millisecond,
		"status": res.StatusCode,
	};

	u, err := url.Parse(urlValue)
	if err != nil {
		logp.Err("%s", err)
	}

	hostIp, _ := net.LookupIP(u.Hostname())

	portValue := 80
	if u.Port() != "" {
		portValue, _ = strconv.Atoi(u.Port())
	}

	fields := common.MapStr{
		"type_name":  "url",
		"object_id":  objectId,
		"account_id":  strings.ToLower(accountId), //소문자로 치환된 account_id 값
		"createtime":  timeValue,
		"kinds": "url",
		"url":  urlValue,
		"host":  u.Hostname(),
		"ip" : hostIp[0],
		"port" : portValue,
		"up" : resultValue,
		"agent" : agentInfo,
		"response" : respInfo,
		"error" : errorInfo,
	}


	//sync방식
	sync_msg := &sarama.ProducerMessage{
		Topic: bt.config.KafkaTopic,
		Value: sarama.StringEncoder(fields.String()),
	}
	partition, offset, err := bt.kafkaProducer.SendMessage(sync_msg)
	logp.Info("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", bt.config.KafkaTopic, partition, offset)


	/*
	//async 방식
	async_msg := &sarama.ProducerMessage{
		Topic: bt.config.KafkaTopic,
		Key: sarama.StringEncoder(objectId),
		Value: sarama.StringEncoder(fields.String()),
	}
	bt.kafkaProducer.Input() <- async_msg
	logp.Info("Message is stored in topic (%s) \n", bt.config.KafkaTopic)
	*/

	if err != nil {
		logp.Err("err %s", err)
		panic(err)
	}
}