package producer

import (
	"../../core/log"
	"../../core/rabbitmq"
	"github.com/dlintw/goconf"
)

var rmq_uri_device string
var exchange_device string
var exchangeType_device string

//初始化RabbitMQ交换器，消息队列名称
func InitRmq_Ex_Que_Name_Device(conf *goconf.ConfigFile) {
	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
	if rmq_uri == "" {
		log.Error("未启用RabbitMq")
		return
	}
	exchange, _ = conf.GetString("rabbitmq", "srv2device_ex")
	exchangeType, _ = conf.GetString("rabbitmq", "srv2device_ex_type")
}

func SendMQMsg2Device(uuid string, message string, cmd string ) {
	if rabbitmq.ProducerRabbitMq == nil {
		log.Error("SendMQMsg2Device: rabbitmq.ConsumerRabbitMq is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: uuid + "_robot", Reliable: true, Durable: true, ReSendNum: 0}

	log.Info("[ ", uuid, "_robot ] " + cmd + "Publish2Device: ", message)
	rabbitmq.ProducerRabbitMq.Publish2Device(&channleContxt, message)
}