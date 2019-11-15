package consumer

import (
	"context"

	"../../core/jobque"
	"../../core/log"
	"../../core/rabbitmq"
)

var rmq_uri string
var exchange string     // = "App2OneNET"
var exchangeType string // = "direct"
var routingKey string   // = "wonlycloud"

var (
	ctx, cancel = context.WithCancel(context.Background())
)

//初始化RabbitMQ交换器，消息队列名称
//func InitRmq_Ex_Que_Name(conf *goconf.ConfigFile) {
//	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
//	if rmq_uri == "" {
//		log.Error("未启用RabbitMq")
//		return
//	}
//	exchange, _ = conf.GetString("rabbitmq", "app2device_ex")
//	exchangeType, _ = conf.GetString("rabbitmq", "app2device_ex_type")
//	routingKey, _ = conf.GetString("rabbitmq", "app2device_que")
//}

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

func Run() {
	log.Info("start ReceiveMQMsgFromAPP......")

	//channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: routingKey, Reliable: true, Durable: true, ReSendNum: 0}
	//
	//rabbitmq.ConsumerRabbitMq.QueueDeclare(channleContxt)

	log.Info("Consumer ReceiveMQMsgFromAPP......")
	msgs, err := rabbitmq.Consumer2appMQ.Consumer()
	if nil != err {
		log.Error("Consumer2appMQ.Consumer() error = ", err)
		panic(err)
	}
	// go程循环去读消息，并放到Job去处理
	for {
		//msgs, err := rabbitmq.ConsumerRabbitMq.Consumer(&channleContxt)
		select {
		case <-ctx.Done():
			log.Info("ReceiveMQMsgFromAPP Close")
			return
		case msg := <-msgs:
			log.Error("Consumer ReceiveMQMsgFromAPP: ", string(msg.Body))
			jobque.JobQueue <- NewConsumerJob(string(msg.Body))
		}

		//go func() {
		//	for d := range msgs {
		//		log.Error("Consumer ReceiveMQMsgFromAPP: ", string(d.Body))
		//		// fetch job
		//		// work := Job{appMsg: AppMsg{pri: string(d.Body)}}
		//		jobque.JobQueue <- NewConsumerJob(string(d.Body))
		//	}
		//	forever <- true // 退出
		//}()
		//<-forever
	}
}

func Close() {
	cancel()
}
