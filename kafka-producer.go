package main

import (
	"fmt"
	"time"
	 kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/Shopify/sarama"
	"github.com/elastic/beats/libbeat/common"
)

var (
	brokerList = kingpin.Flag("brokerList", "List of brokers to connect").Default("10.165.60.52:9092").Strings()
	topic      = kingpin.Flag("topic", "Topic name").Default("camsbeat").String()
)

func main() {
	kingpin.Parse()
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	for _, v := range *brokerList {
		fmt.Println("brokerList : " + v)
	}

	producer, err := sarama.NewSyncProducer(*brokerList, config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	agentInfo := common.MapStr{
		"name" : "camsbeat-66",
		"version": "v1",
		"contury": "KR",
	};

	var responsetime time.Duration
	responInfo := common.MapStr{
		"us" : responsetime / time.Microsecond,
		"ms" : responsetime / time.Millisecond,
		"status": 200,
	};

	fields := common.MapStr{
		"type_name":  "url",
		"object_id":  "IDnmNGAB-zo9g1nyY_mt",
		"createtime":   time.Now(),
		"kinds":      "url",
		"url":  "http://www.daum.net",
		"host":   "www.daum.net",
		"ip" : "1.1.1.1",
		"port" : 80,
		"up" : true,
		"agent" : agentInfo,
		"response" : responInfo,
		//"error" : errorInfo,
	}

	msg := &sarama.ProducerMessage{
		Topic: *topic,
		Value: sarama.StringEncoder(fields.String()),
	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		panic(err)
	}
	fmt.Printf(" Message is stored in topic(%s)/partition(%d)/offset(%d)\n", *topic, partition, offset)

}