package producer

import (
	"fmt"
	"../../core/rabbitmq"
	"github.com/dlintw/goconf"
)

var rmq_uri string
var exchange string	// = "OneNET2APP"
var exchangeType string	// = "direct"

//初始化RabbitMQ交换器，消息队列名称
func InitRmq_Ex_Que_Name(conf *goconf.ConfigFile) {
	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
	if rmq_uri == "" {
		fmt.Println("未启用RabbitMq")
		return
	}
	exchange, _ = conf.GetString("rabbitmq", "device2app_ex")
	exchangeType, _ = conf.GetString("rabbitmq", "device2app_ex_type")
}

func SendMQMsg2APP(uuid string, message string) {
	if rabbitmq.ProducerRabbitMq == nil {
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: uuid, Reliable: true, Durable: true}

	fmt.Println("sending message")
	rabbitmq.ProducerRabbitMq.Publish2App(&channleContxt, message)
}