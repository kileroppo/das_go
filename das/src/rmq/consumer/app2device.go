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
	go consume()
}

func consume() {
	msgs, err := rabbitmq.Consumer2appMQ.Consumer()
	if err != nil {
		log.Error("Consumer2appMQ.Consumer() error = ", err)
		if err = rabbitmq.Consumer2appMQ.ReConn(); err != nil {
			log.Warningf("Consumer2appMQ Reconnection Failed")
			return
		}
		log.Debug("Consumer2appMQ Reconnection Successful")
		msgs, err = rabbitmq.Consumer2appMQ.Consumer()
	}

	for msg := range msgs {
		log.Info("Consumer ReceiveMQMsgFromAPP: ", string(msg.Body))
		jobque.JobQueue <- NewConsumerJob(string(msg.Body))
	}

	select {
	case <-ctx.Done():
		log.Info("ReceiveMQMsgFromAPP Close")
		return
	default:
		go consume()
		return
	}
}

func Close() {
	cancel()
}
