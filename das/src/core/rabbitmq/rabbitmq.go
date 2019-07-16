package rabbitmq

import (
	"github.com/dlintw/goconf"
	"sync"
	"../log"
)

var ProducerRabbitMq *BaseMq
var producerRabbitMqonce sync.Once
var ProducerRabbitMq2Db *BaseMq
var producerRabbitMqonce2Db sync.Once
var ProducerRabbitMq2Device *BaseMq
var producerRabbitMqonce2Device sync.Once
var ConsumerRabbitMq *BaseMq
var consumerRabbitMqonce sync.Once

func getProducerRabbitMq(uri string) *BaseMq {
	producerRabbitMqonce.Do(func() {
		ProducerRabbitMq = &BaseMq{
			MqConnection: &MqConnection{MqUri: uri},
		}
		ProducerRabbitMq.Init()
	})
	return ProducerRabbitMq
}

func getProducerRabbitMq2Db(uri string) *BaseMq {
	producerRabbitMqonce2Db.Do(func() {
		ProducerRabbitMq2Db = &BaseMq{
			MqConnection: &MqConnection{MqUri: uri},
		}
		ProducerRabbitMq2Db.Init()
	})
	return ProducerRabbitMq2Db
}

func getConsumerRabbitMq(uri string) *BaseMq {
	consumerRabbitMqonce.Do(func() {
		ConsumerRabbitMq = &BaseMq{
			MqConnection: &MqConnection{MqUri: uri},
		}
		ConsumerRabbitMq.Init()
	})
	return ConsumerRabbitMq
}

func getProducerRabbitMq2Device(uri string) *BaseMq {
	producerRabbitMqonce2Device.Do(func() {
		ProducerRabbitMq2Device = &BaseMq{
			MqConnection: &MqConnection{MqUri: uri},
		}
		ProducerRabbitMq2Device.Init()
	})
	return ProducerRabbitMq2Device
}

func InitProducerMqConnection(conf *goconf.ConfigFile) *BaseMq {
	uri, _ := conf.GetString("rabbitmq", "rabbitmq_uri")
	if uri == "" {
		log.Error("未启用rabbimq")
		return nil
	}
	return getProducerRabbitMq(uri)
}

func InitProducerMqConnection2Db(conf *goconf.ConfigFile) *BaseMq {
	uri, _ := conf.GetString("rabbitmq", "rabbitmq_uri")
	if uri == "" {
		log.Error("未启用rabbimq")
		return nil
	}
	return getProducerRabbitMq2Db(uri)
}

func InitConsumerMqConnection(conf *goconf.ConfigFile) *BaseMq {
	uri, _ := conf.GetString("rabbitmq", "rabbitmq_uri")
	if uri == "" {
		log.Error("未启用rabbimq")
		return nil
	}
	return getConsumerRabbitMq(uri)
}

func InitProducerMqConnection2Device(conf *goconf.ConfigFile) *BaseMq {
	uri, _ := conf.GetString("rabbitmq", "rabbitmq_uri")
	if uri == "" {
		log.Error("未启用rabbimq")
		return nil
	}
	return getProducerRabbitMq2Device(uri)
}
