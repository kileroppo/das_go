package rabbitmq

import (
	"context"
	"errors"
	"reflect"
	"sync"
		"time"
	"runtime"

	"github.com/streadway/amqp"

	"das/core/log"
)

var (
	ErrExchangeType = errors.New("exchange type was not supported")
	ErrReConn       = errors.New("baseMQ reconnection failed")
	ErrInvalidMQ    = errors.New("baseMQ was invalid")
)

type ChannelContext struct {
	Name         string
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool
	AutoDelete   bool
	ChannelId    string
	QueueName    string
}

type baseMq struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	mqUri      string
	channelCtx ChannelContext
	initOnce   sync.Once

	currReConnNum int32
	mu            sync.Mutex
	isConsumer    bool
	reconnFlag    bool

	ctx    context.Context
	cancel context.CancelFunc
}

func (bmq *baseMq) init() (err error) {
	bmq.connection, err = amqp.Dial(bmq.mqUri)
	if err != nil {
		log.Errorf("%s init() error = %s", bmq.channelCtx.Name, err)
		panic(err)
	}

	bmq.channel, err = bmq.connection.Channel()
	if err != nil {
		log.Errorf("%s init() error = %s", bmq.channelCtx.Name, err)
		panic(err)
	}

	bmq.ctx, bmq.cancel = context.WithCancel(context.Background())

	return nil
}

func (bmq *baseMq) initConsumer() (err error) {
	queue, err := bmq.channel.QueueDeclare(
		bmq.channelCtx.QueueName,  // name, leave empty to generate a unique name
		bmq.channelCtx.Durable,    // durable
		bmq.channelCtx.AutoDelete, // delete when usused
		false,                     // exclusive
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != err {
		log.Errorf("%s QueueDeclare() error = %s", bmq.channelCtx.Name, err)
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
		log.Errorf("%s QueueBind() error = %s", bmq.channelCtx.Name, err)
		bmq.channel.Close()
		return err
	}

	return nil
}

func (bmq *baseMq) initExchange() error {
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
		log.Errorf("%s initExchange() error = %s", bmq.channelCtx.Name, err)
		return err
	}

	return nil
}

func (bmq *baseMq) Publish(data []byte, routingKey string) (err error) {
	select {
	case <-bmq.ctx.Done():
		return ErrInvalidMQ
	default:
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
			log.Errorf("%s Publish() error = %s", bmq.channelCtx.Name, err)
			bmq.cancel()
			go bmq.ReConn()
			return
		}
	}
	return nil
}

func (bmq *baseMq) Consumer() (ch <-chan amqp.Delivery, err error) {
	//if isNil(bmq.channel) {
	//	bmq.ReConn()
	//}

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
		log.Errorf("%s Consumer() error = ", err)
		return nil, err
	}

	return
}

func (bmq *baseMq) ReConn() error {
	//log.Info("MQ Reconnnecting...")
	var err error
	isFirst := true
	bmq.mu.Lock()
	if bmq.reconnFlag {
		bmq.mu.Unlock()
		runtime.Goexit()
	}
	bmq.reconnFlag = true
	bmq.mu.Unlock()
	for {
		if isFirst {
			isFirst = false
		} else {
			time.Sleep(time.Second * 20)
		}
		log.Infof("%s %dth Reconnecting...", bmq.channelCtx.Name, bmq.currReConnNum+1)
		bmq.connection, err = amqp.Dial(bmq.mqUri)
		if err != nil {
			bmq.currReConnNum++
			continue
		}
		bmq.channel, err = bmq.connection.Channel()
		if err != nil {
			bmq.currReConnNum++
			continue
		}
		if bmq.isConsumer {
			if err = bmq.initConsumer(); err != nil {
				bmq.currReConnNum++
				continue
			}
		}
		bmq.currReConnNum = 0
		bmq.ctx, bmq.cancel = context.WithCancel(context.Background())
		log.Infof("%s Reconnected Successfully", bmq.channelCtx.Name)
		return nil
	}
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
