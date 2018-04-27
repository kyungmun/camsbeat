// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type Config struct {
	Period time.Duration `config:"period"`
	Kinds string	     `config:"kinds"`   //url, icmp, tcp
	Values []string      `config:"values"`
	ObjectId []string    `config:"objectid"`
	AccountId []string   `config:"accountid"`
	Name string          `config:"name`
	KafkaBroker string   `config:"kafka-broker"`
	KafkaTopic string    `config:"kafka-topic"`
	KafkaAuthEnabled bool   `config:"kafka-auth"`
	KafkaSaslUser string    `config:"kafka-sasl-user"`
	KafkaSaslPass string    `config:"kafka-sasl-pass"`

}

var DefaultConfig = Config{
	Period: 10 * time.Second,
	Kinds: "url",
	Name: "cams-agent",
	KafkaBroker: "127.0.0.1:9092",
	KafkaTopic: "cams",
	KafkaSaslUser: "alice",
	KafkaSaslPass: "alice-secret",
	KafkaAuthEnabled: true,
}
