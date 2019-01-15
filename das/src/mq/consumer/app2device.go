package consumer

import (
	"../../core/rabbitmq"
	"github.com/dlintw/goconf"
	"../../core/log"
)

var rmq_uri string;
var exchange string;		// = "App2OneNET"
var exchangeType string;	// = "direct"
var routingKey string;		// = "wonlycloud"

//初始化RabbitMQ交换器，消息队列名称
func InitRmq_Ex_Que_Name(conf *goconf.ConfigFile) {
	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
	if rmq_uri == "" {
		log.Error("未启用RabbitMq")
		return
	}
	exchange, _ = conf.GetString("rabbitmq", "app2device_ex")
	exchangeType, _ = conf.GetString("rabbitmq", "app2device_ex_type")
	routingKey, _ = conf.GetString("rabbitmq", "app2device_que")
}

func ReceiveMQMsgFromAPP() {
	log.Info("start ReceiveMQMsgFromAPP......")

	//初始化rabbitmq
	if rabbitmq.ConsumerRabbitMq == nil {
		log.Error("ReceiveMQMsgFromAPP: rabbitmq.ConsumerRabbitMq is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: routingKey, Reliable: true, Durable: true}

	rabbitmq.ConsumerRabbitMq.QueueDeclare(&channleContxt)

	log.Info("Consumer ReceiveMQMsgFromAPP......")
	// go程循环去读消息，并放到Job去处理
	for {
		msgs := rabbitmq.ConsumerRabbitMq.Consumer(&channleContxt)
		log.Debug("Consumer 2 ReceiveMQMsgFromAPP......")
		forever := make(chan bool)
		go func() {
			for d := range msgs {
				log.Debug("process 3 ReceiveMQMsgFromAPP: ", string(d.Body))
				// fetch job
				work := Job{appMsg: AppMsg{pri: string(d.Body)}}
				JobQueue <- work
			}
		}()
		<-forever
	}
}

