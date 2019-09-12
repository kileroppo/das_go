package consumer

import (
	"../../core/rabbitmq"
	"github.com/dlintw/goconf"
	"../../core/log"
	"../../core/jobque"
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

type ConsumerJob struct {
	rawData string
}

func NewConsumerJob(rawData string) ConsumerJob {
	return ConsumerJob{
		rawData: rawData,
	}
}

func (c ConsumerJob) Handle() {
	ProcAppMsg(c.rawData)
}

func ReceiveMQMsgFromAPP() {
	log.Info("start ReceiveMQMsgFromAPP......")

	//初始化rabbitmq
	if rabbitmq.ConsumerRabbitMq == nil {
		log.Error("ReceiveMQMsgFromAPP: rabbitmq.ConsumerRabbitMq is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: routingKey, Reliable: true, Durable: true, ReSendNum: 0}

	rabbitmq.ConsumerRabbitMq.QueueDeclare(&channleContxt)

	log.Info("Consumer ReceiveMQMsgFromAPP......")
	// go程循环去读消息，并放到Job去处理
	for {
		msgs := rabbitmq.ConsumerRabbitMq.Consumer(&channleContxt)
		forever := make(chan bool)
		go func() {
			for d := range msgs {
				log.Error("Consumer ReceiveMQMsgFromAPP 1: ", string(d.Body))
				// fetch job
				// work := Job{appMsg: AppMsg{pri: string(d.Body)}}
				jobque.JobQueue <- NewConsumerJob(string(d.Body))
			}
		}()
		<-forever
	}
}

