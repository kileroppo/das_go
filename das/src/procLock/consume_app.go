package procLock

import (
	"context"

	"das/core/jobque"
	"das/core/log"
	"das/core/mqtt"
	"das/core/rabbitmq"
)

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
	go consume()
	go consumePadDoor()
	go initMqtt()
}

func consume() {
	log.Info("start ReceiveMQMsgFromAPP......")
	msgs, err := rabbitmq.ConsumeApp()
	if err != nil {
		log.Errorf("consumeApp > %s", err)
	}

	for msg := range msgs {
		//log.Info("Consumer ReceiveMQMsgFromAPP: ", string(msg.Body))
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
	mqtt.CloseMqttCli(mqttCli)
}
