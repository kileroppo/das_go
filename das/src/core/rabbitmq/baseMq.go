package rabbitmq

import (
	"../log"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
	"time"
)

type MqConnection struct {
	Lock       sync.RWMutex
	Connection *amqp.Connection
	MqUri      string
}

type ChannelContext struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool
	ChannelId    string
	Channel      *amqp.Channel
	ReSendNum    int32 // 重发次数
}

type BaseMq struct {
	MqConnection *MqConnection
}

func (bmq *BaseMq) Init() {
	bmq.refreshConnectionAndChannel()
}

// One would typically keep a channel of publishings, a sequence number, and a
// set of unacknowledged sequence numbers and loop until the publishing channel
// is closed.
func (bmq *BaseMq) confirmOne(confirms <-chan amqp.Confirmation) {
	log.Info("waiting for confirmation of one publishing")
	if confirmed := <-confirms; confirmed.Ack {
		log.Info("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Error("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}

/*
 * get md5 from channel context
 */

/*
1. use old connection to generate channel
2. update connection then channel
*/
func (bmq *BaseMq) refreshConnectionAndChannel() (err error) {
	log.Debug("refreshConnectionAndChannel mq conn")
	bmq.MqConnection.Connection, err = amqp.Dial(bmq.MqConnection.MqUri)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

/*
*	publish message
*
*	发给APP的消息
 */
func (bmq *BaseMq) Publish2App(channelContext *ChannelContext, body []byte) error {
	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	defer channelContext.Channel.Close()

	if err != nil {
		log.Error(err)
		return err
	}

	queue_name, qerr := channelContext.Channel.QueueDeclare(
		"",    // name, leave empty to generate a unique name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		amqp.Table{
			/*"x-message-ttl": int32(5000),*/
			"x-expires": int32(1000)}, // arguments
	)
	if nil != qerr {
		log.Error("Publish2App, channelContext.Channel.QueueDeclare, err: ", qerr)
		//bmq.refreshConnectionAndChannel()
		return qerr
	}

	qbinderr := channelContext.Channel.QueueBind(
		queue_name.Name,           // name of the queue
		channelContext.RoutingKey, // bindingKey
		channelContext.Exchange,   // sourceExchange
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != qbinderr {
		log.Error("Publish2App, channelContext.Channel.QueueBind, err: ", qbinderr)
		//bmq.refreshConnectionAndChannel()
		return qbinderr
	}

	if err := channelContext.Channel.Publish(
		channelContext.Exchange,   // publish to an exchange
		channelContext.RoutingKey, // routing to 0 or more queues
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Error("send message failed refresh connection", err)
		time.Sleep(10 * time.Second)
		log.Error("Publish2App() 2-bmq.ChannelContexts[" + channelContext.ChannelId + "] is nil, refreshConnectionAndChannel()")
		if atomic.LoadInt32(&channelContext.ReSendNum) > 0 {
			log.Error("Publish2App ReSend message=", body, ", num=", channelContext.ReSendNum)
			atomic.AddInt32(&channelContext.ReSendNum, -1)
			bmq.Publish2App(channelContext, body)
		}

		return err
	}

	return nil
}

/*
*	publish message
*
*	存到mongodb数据库
 */
func (bmq *BaseMq) Publish2Db(channelContext *ChannelContext, body []byte) error {
	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	defer channelContext.Channel.Close()

	if err != nil {
		log.Error(err)
		return err
	}

	queue_name, qerr := channelContext.Channel.QueueDeclare(
		channelContext.RoutingKey, // name, leave empty to generate a unique name
		true,                      // durable
		false,                     // delete when usused
		false,                     // exclusive
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != qerr {
		log.Error("Publish2Db, channelContext.Channel.QueueDeclare, err: ", qerr)
		//bmq.refreshConnectionAndChannel(channelContext)
		return qerr
	}

	qbinderr := channelContext.Channel.QueueBind(
		queue_name.Name,         // name of the queue
		"",                      // bindingKey
		channelContext.Exchange, // sourceExchange
		false,                   // noWait
		nil,                     // arguments
	)
	if nil != qbinderr {
		log.Error("Publish2Db, channelContext.Channel.QueueBind, err: ", qbinderr)
		//bmq.refreshConnectionAndChannel(channelContext)
		return qbinderr
	}

	if err := channelContext.Channel.Publish(
		channelContext.Exchange, // publish to an exchange
		queue_name.Name,         // routing to 0 or more queues
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Error("send message failed refresh connection", err)
		time.Sleep(10 * time.Second)
		log.Error("Publish2Db() 2-bmq.ChannelContexts[" + channelContext.ChannelId + "] is nil, refreshConnectionAndChannel()")
		//recon_err := bmq.refreshConnectionAndChannel(channelContext)
		if atomic.LoadInt32(&channelContext.ReSendNum) > 0 {
			log.Error("Publish2App ReSend message=", body, ", num=", channelContext.ReSendNum)
			atomic.AddInt32(&channelContext.ReSendNum, -1)
			bmq.Publish2Db(channelContext, body)
		}
	}

	return nil
}

/*
*	publish message
*
*	存到mongodb数据库 -2
 */
func (bmq *BaseMq) Publish2Db2(channelContext *ChannelContext, body []byte) error {

	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	defer channelContext.Channel.Close()

	if err != nil {
		log.Error(err)
		return err
	}

	queue_name, qerr := channelContext.Channel.QueueDeclare(
		channelContext.RoutingKey, // name, leave empty to generate a unique name
		true,                      // durable
		false,                     // delete when usused
		false,                     // exclusive
		false,                     // noWait
		nil,                       // arguments
	)

	if nil != qerr {
		log.Error("Publish2Db2, channelContext.Channel.QueueDeclare, err: ", qerr)
		return qerr
		//bmq.refreshConnectionAndChannel()
		//return qerr
	}

	qbinderr := channelContext.Channel.QueueBind(
		queue_name.Name,           // name of the queue
		channelContext.RoutingKey, // bindingKey
		channelContext.Exchange,   // sourceExchange
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != qbinderr {
		log.Error("Publish2Db2, channelContext.Channel.QueueBind, err: ", qbinderr)
		//bmq.refreshConnectionAndChannel()
		return qbinderr
	}

	if err := channelContext.Channel.Publish(
		channelContext.Exchange,   // publish to an exchange
		channelContext.RoutingKey, // routing to 0 or more queues
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Error("Publish2Db2() send message failed refresh connection", err)
		time.Sleep(10 * time.Second)
		log.Error("Publish2Db2() 2-bmq.ChannelContexts[" + channelContext.ChannelId + "] is nil, refreshConnectionAndChannel()")
		if atomic.LoadInt32(&channelContext.ReSendNum) > 0 {
			log.Error("Publish2App ReSend message=", body, ", num=", channelContext.ReSendNum)
			atomic.AddInt32(&channelContext.ReSendNum, -1)
			bmq.Publish2Db2(channelContext, body)
		}
	}

	return nil
}

/*
*	publish message
*
*	发给平板设备的消息
 */
func (bmq *BaseMq) Publish2Device(channelContext *ChannelContext, body []byte) error {
	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	defer channelContext.Channel.Close()

	if err != nil {
		log.Error(err)
		return err
	}

	queue_name, qerr := channelContext.Channel.QueueDeclare(
		"",    // name, leave empty to generate a unique name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		amqp.Table{
			/*"x-message-ttl": int32(5000),*/
			"x-expires": int32(2000)}, // arguments
	)
	if nil != qerr {
		log.Error("Publish2Device, channelContext.Channel.QueueDeclare, err: ", qerr)
		return qerr
	}

	qbinderr := channelContext.Channel.QueueBind(
		queue_name.Name,           // name of the queue
		channelContext.RoutingKey, // bindingKey
		channelContext.Exchange,   // sourceExchange
		false,                     // noWait
		nil,                       // arguments
	)
	if nil != qbinderr {
		log.Error("Publish2Device, channelContext.Channel.QueueBind, err: ", qbinderr)
		//bmq.refreshConnectionAndChannel(channelContext)
		return qbinderr
	}

	if err := channelContext.Channel.Publish(
		channelContext.Exchange,   // publish to an exchange
		channelContext.RoutingKey, // routing to 0 or more queues
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Error("send message failed refresh connection, err: ", err)
		time.Sleep(10 * time.Second)
		log.Error("Publish2Device() 2-bmq.ChannelContexts[" + channelContext.ChannelId + "] is nil, refreshConnectionAndChannel()")
		if atomic.LoadInt32(&channelContext.ReSendNum) > 0 {
			log.Error("Publish2App ReSend message=", body, ", num=", channelContext.ReSendNum)
			atomic.AddInt32(&channelContext.ReSendNum, -1)
			bmq.Publish2Device(channelContext, body)
		}
	}

	return nil
}

/*
*	QueueDeclare
 */
func (bmq *BaseMq) QueueDeclare(channelContext *ChannelContext) error {
	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()

	if err != nil {
		log.Error(err)
		return err
	}

	queue_name, err := channelContext.Channel.QueueDeclare(
		channelContext.RoutingKey, // name, leave empty to generate a unique name
		true,                      // durable
		false,                     // delete when usused
		false,                     // exclusive
		false,                     // noWait
		nil,                       // arguments
	)

	err = channelContext.Channel.QueueBind(
		queue_name.Name,         // name of the queue
		"",                      // bindingKey
		channelContext.Exchange, // sourceExchange
		false,                   // noWait
		nil,                     // arguments
	)

	if err != nil {
		log.Error(err)
		log.Error("Failed to register a consumer")
		bmq.refreshConnectionAndChannel()
	}
	return nil
}

/*
*	consumer message
 */
func (bmq *BaseMq) Consumer(channelContext *ChannelContext) (<-chan amqp.Delivery, error) {
	if bmq.MqConnection.Connection == nil || bmq.MqConnection.Connection.IsClosed() {
		bmq.refreshConnectionAndChannel()
	}

	var err error
	channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	defer channelContext.Channel.Close()

	if err != nil {
		log.Error(err)
		return nil, err
	}
	//for {
	msgs, err := channelContext.Channel.Consume(
		channelContext.RoutingKey, // queue
		"",                        // consumer
		true,                      // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)
	if err != nil {
		log.Error(err)
		log.Error("Failed to register a consumer")
		//bmq.refreshConnectionAndChannel()
		return nil, err
	}
	return msgs, nil
	//}
}

/*func (bmq *BaseMq) Consumer(channelContext *ChannelContext, calllback func(string) bool) error {
	channelContext.ChannelId = bmq.generateChannelId(channelContext)
	if bmq.ChannelContexts[channelContext.ChannelId] == nil {
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		channelContext = bmq.ChannelContexts[channelContext.ChannelId]
	}
	if msgs, err := channelContext.Channel.Consume(
		channelContext.RoutingKey, // routing to 0 or more queues
		"",    // consumer
		true, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	); err != nil {
		log.Error(err)
		log.Error("consumer message failed refresh connection")
		time.Sleep(10 * time.Second)
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		//创建一个channel
		forever := make(chan bool)

		//调用gorountine
		go func() {
			for d := range msgs {
				result := calllback(string(d.Body))
				if result {
					d.Ack(false)
				} else {
					d.Nack(false, true)
				}
			}
		}()
		<-forever
	}
	return nil
}*/
