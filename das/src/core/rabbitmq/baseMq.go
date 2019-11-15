package rabbitmq

import (
	"errors"
	"sync"

	"github.com/streadway/amqp"

	"../log"
	"time"
	"reflect"
)

var (
	ErrExchangeType = errors.New("exchange type was not supported")
)

type ChannelContext struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool
	AutoDelete   bool
	ChannelId    string

	ReSendNum int32 // 重发次数
}

type baseMq struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	mqUri      string
	channelCtx ChannelContext
	initOnce sync.Once
}

func (bmq *baseMq) init() (err error) {
	bmq.connection, err = amqp.Dial(bmq.mqUri)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	bmq.channel, err = bmq.connection.Channel()
	if err != nil {
		log.Error("connection.Channel() error = ", err)
		return err
	}

	return nil
}

func (bmq *baseMq) initConsumer() (err error) {
	log.Info("baseMQ init exchange: ", bmq.channelCtx.Exchange)

	queue, err := bmq.channel.QueueDeclare(
		bmq.channelCtx.RoutingKey, // name, leave empty to generate a unique name
		bmq.channelCtx.Durable,    // durable
		bmq.channelCtx.AutoDelete, // delete when usused
		false,                     // exclusive
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != err {
		log.Error("baseMq QueueDeclare() error = ", err)
		bmq.channel.Close()
		return err
	}
	err = bmq.channel.QueueBind(
		queue.Name,                // name of the queue
		bmq.channelCtx.RoutingKey, // bindingKey
		bmq.channelCtx.Exchange,   // sourceExchange
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != err {
		log.Error("baseMq, QueueBind() error = ", err)
		bmq.channel.Close()
		return err
	}

	return nil
}

func (bmq *baseMq) initExchange() error {
	log.Info("BaseMQ init exchange")
	err := bmq.channel.ExchangeDeclare(
		bmq.channelCtx.Exchange,
		bmq.channelCtx.ExchangeType,
		true,
		false,
		false,
		false,
		amqp.Table{},
	)

	if err != nil {
		log.Error("Channel.ExchangeDeclare() error = ", err)
		return err
	}

	return nil
}

func (bmq *baseMq) Publish(data []byte, routingKey string) (err error) {
	err = bmq.channel.Publish(
		bmq.channelCtx.Exchange, // publish to an exchange
		routingKey,              // routing to 0 or more queues
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            data,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
	if err != nil {
		log.Error("bmq.channel.Publish error = ", err)
		return
	}

	return nil
}

func (bmq *baseMq) Consumer() (ch <-chan amqp.Delivery, err error) {
	reConn := func() {
		time.Sleep(10 * time.Second)
		bmq.reConn()
		ch,err = bmq.Consumer()
	}

	ch, err = bmq.channel.Consume(
		bmq.channelCtx.RoutingKey, // queue
		"",                        // consumer
		true,                      // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)

	if err != nil {
		log.Error(err)
		reConn()
	}

	return
}

func (bmq *baseMq) reConn() {
	log.Info("MQ Reconnnecting...")
	bmq.init()
}

func (bmq *baseMq) Close() error {

	bmq.channel.Close()
	if !isNil(bmq.connection) && !bmq.connection.IsClosed() {
		err := bmq.connection.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func isNil(i interface{}) bool {
	switch i.(type) {
	case *amqp.Connection:
		val := reflect.ValueOf(i)
		return val.IsNil()
	}

	return false
}
