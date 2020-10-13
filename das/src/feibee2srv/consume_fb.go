package feibee2srv

import (
	"context"

	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	ctx, cancel = context.WithCancel(context.Background())
)

func Init() {
	go consumeFb()
}

func Close() {
	cancel()
}

func consumeFb() {
	log.Info("start ReceiveDevMsgFromFeibee......")
	msgs, err := rabbitmq.ConsumeFb()
	if err != nil {
		log.Errorf("ConsumeFb > %s", err)
	}

	for msg := range msgs {
		rabbitmq.SendGraylogByMQ("飞比Server-mq->DAS: %s", msg.Body)
		jobque.JobQueue <- NewFeibeeJob(msg.Body)
	}

	select {
	case <-ctx.Done():
		log.Info("ReceiveDevMsgFromFeibee Close")
		return
	default:
		go consumeFb()
		return
	}
}