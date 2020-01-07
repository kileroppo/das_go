package producer

import (
	"das/core/log"
	"das/core/rabbitmq"
)

var rmq_uri_device string
var exchange_device string
var exchangeType_device string

//初始化RabbitMQ交换器，消息队列名称
//func InitRmq_Ex_Que_Name_Device(conf *goconf.ConfigFile) {
//	rmq_uri_device, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
//	if rmq_uri_device == "" {
//		log.Error("未启用RabbitMq")
//		return
//	}
//	exchange_device, _ = conf.GetString("rabbitmq", "srv2device_ex")
//	exchangeType_device, _ = conf.GetString("rabbitmq", "srv2device_ex_type")
//}

func SendMQMsg2Device(uuid string, message string, cmd string) {
	//if rabbitmq.ProducerRabbitMq2Device == nil {
	//	log.Error("SendMQMsg2Device: rabbitmq.ProducerRabbitMq2Device is nil.")
	//	return
	//}

	var rkey string
	rkey = uuid + "_robot"
	//channleContxt := rabbitmq.ChannelContext{Exchange: exchange_device, ExchangeType: exchangeType_device, RoutingKey: rkey, Reliable: true, Durable: true, ReSendNum: 0}

	log.Info("[ ", rkey, " ] "+cmd+" rabbitmq.ProducerRabbitMq2Device.Publish2Device: ", message)
	//rabbitmq.ProducerRabbitMq2Device.Publish2Device(channleContxt, []byte(message))
	rabbitmq.Publish2dev([]byte(message), rkey)
}
