package rabbitmq

import (
	"../log"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/streadway/amqp"
	"strconv"
	"sync"
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
	ReSendNum	 int		// 重发次数
}

type BaseMq struct {
	MqConnection *MqConnection
	//rabbitMq通道缓存
	ChannelContexts map[string]*ChannelContext
}

func (bmq *BaseMq) Init() {
	bmq.ChannelContexts = make(map[string]*ChannelContext)
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
func (bmq *BaseMq) generateChannelId(channelContext *ChannelContext) string {
	stringTag := channelContext.Exchange + ":" + channelContext.ExchangeType + ":" + channelContext.RoutingKey + ":" +
		strconv.FormatBool(channelContext.Durable) + ":" + strconv.FormatBool(channelContext.Reliable)
	hasher := md5.New()
	hasher.Write([]byte(stringTag))
	return hex.EncodeToString(hasher.Sum(nil))
}

/*
1. use old connection to generate channel
2. update connection then channel
*/
func (bmq *BaseMq) refreshConnectionAndChannel(channelContext *ChannelContext) error {
	bmq.MqConnection.Lock.Lock()
	defer bmq.MqConnection.Lock.Unlock()
	var err error

	if bmq.MqConnection.Connection != nil {
		channelContext.Channel, err = bmq.MqConnection.Connection.Channel()
	} else {
		log.Error("connection not init, dial first time......")
		err = errors.New("connection nil")
	}

	// reconnect connection
	if err != nil {
		for {
			bmq.MqConnection.Connection, err = amqp.Dial(bmq.MqConnection.MqUri)
			if err != nil {
				log.Error("connect mq get connection error,retry..." + bmq.MqConnection.MqUri)
				time.Sleep(10 * time.Second)
			} else {
				log.Info("connection RabbitMQ......")
				channelContext.Channel, _ = bmq.MqConnection.Connection.Channel()
				break
			}
		}
	}

	if err = channelContext.Channel.ExchangeDeclare(
		channelContext.Exchange,     // name
		channelContext.ExchangeType, // type
		channelContext.Durable,      // durable
		false,            // auto-deleted
		false,               // internal
		false,               // noWait
		nil,                   // arguments
	); err != nil {
		log.Error("channel exchange deflare failed refreshConnectionAndChannel again", err)
		return err
	}

	//add channel to channel cache
	bmq.ChannelContexts[channelContext.ChannelId] = channelContext
	return nil
}

/*
*	publish message
*
*	发给APP的消息
*/
func (bmq *BaseMq) Publish2App(channelContext *ChannelContext, body string) error {
	channelContext.ChannelId = bmq.generateChannelId(channelContext)
	if bmq.ChannelContexts[channelContext.ChannelId] == nil {
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		channelContext = bmq.ChannelContexts[channelContext.ChannelId]
	}

	/*queue_name, _ := channelContext.Channel.QueueDeclare(
		"",  					// name, leave empty to generate a unique name
		false,  				// durable
		false, 			// delete when usused
		false, 			// exclusive
		false, 				// noWait
		amqp.Table{
			"x-message-ttl": int32(5000),
			"x-expires": int32(10000)},   // arguments
	)

	channelContext.Channel.QueueBind(
		queue_name.Name,    // name of the queue
		channelContext.RoutingKey,   // bindingKey
		channelContext.Exchange, // sourceExchange
		false,    // noWait
		nil,      	  // arguments
	)*/

	if err := channelContext.Channel.Publish(
		channelContext.Exchange,    // publish to an exchange
		channelContext.RoutingKey,  // routing to 0 or more queues
		false,            // mandatory
		false,            // immediate
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
		log.Error("send message failed refresh connection")
		time.Sleep(10 * time.Second)
		recon_err := bmq.refreshConnectionAndChannel(channelContext)
		if nil != recon_err {
			if channelContext.ReSendNum < 3 {
				log.Error("Publish2App ReSend message=", body, ", num=", channelContext.ReSendNum)
				channelContext.ReSendNum++
				bmq.Publish2App(channelContext, body)
			}
		}
	}

	return nil
}

/*
*	publish message
*
*	存到mongodb数据库
*/
func (bmq *BaseMq) Publish2Db(channelContext *ChannelContext, body string) error {
	channelContext.ChannelId = bmq.generateChannelId(channelContext)
	if bmq.ChannelContexts[channelContext.ChannelId] == nil {
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		channelContext = bmq.ChannelContexts[channelContext.ChannelId]
	}

	queue_name, _ := channelContext.Channel.QueueDeclare(
		channelContext.RoutingKey,  // name, leave empty to generate a unique name
		true,  				// durable
		false, 			// delete when usused
		false, 			// exclusive
		false, 				// noWait
		nil,   // arguments
	)

	channelContext.Channel.QueueBind(
		queue_name.Name,    // name of the queue
		"",      		// bindingKey
		channelContext.Exchange, // sourceExchange
		false,    // noWait
		nil,      	  // arguments
	)

	if err := channelContext.Channel.Publish(
		channelContext.Exchange,    // publish to an exchange
		queue_name.Name,  // routing to 0 or more queues
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "UTF-8",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		log.Error("send message failed refresh connection")
		time.Sleep(10 * time.Second)
		recon_err := bmq.refreshConnectionAndChannel(channelContext)
		if nil != recon_err {
			if channelContext.ReSendNum < 3 {
				log.Error("Publish2Db ReSend message=", body, ", num=", channelContext.ReSendNum)
				channelContext.ReSendNum++
				bmq.Publish2App(channelContext, body)
			}
		}
	}

	return nil
}

/*
*	QueueDeclare
*/
func (bmq *BaseMq) QueueDeclare(channelContext *ChannelContext) error {
	channelContext.ChannelId = bmq.generateChannelId(channelContext)
	if bmq.ChannelContexts[channelContext.ChannelId] == nil {
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		channelContext = bmq.ChannelContexts[channelContext.ChannelId]
	}

	queue_name, err := channelContext.Channel.QueueDeclare(
		channelContext.RoutingKey,  // name, leave empty to generate a unique name
		true,  				// durable
		false, 			// delete when usused
		false, 			// exclusive
		false, 				// noWait
		nil,   // arguments
	)

	err = channelContext.Channel.QueueBind(
		queue_name.Name,    // name of the queue
		"",      		// bindingKey
		channelContext.Exchange, // sourceExchange
		false,    // noWait
		nil,      	  // arguments
	)

	if err != nil {
		log.Error(err)
		log.Error("Failed to register a consumer")
		bmq.refreshConnectionAndChannel(channelContext)
	}
	return nil
}

/*
*	consumer message
*/
func (bmq *BaseMq) Consumer(channelContext *ChannelContext) <-chan amqp.Delivery {
	channelContext.ChannelId = bmq.generateChannelId(channelContext)
	if bmq.ChannelContexts[channelContext.ChannelId] == nil {
		bmq.refreshConnectionAndChannel(channelContext)
	} else {
		channelContext = bmq.ChannelContexts[channelContext.ChannelId]
	}
	//for {
	msgs, err := channelContext.Channel.Consume(
		channelContext.RoutingKey, 		// queue
		"",                   // consumer
		true,                  // auto-ack
		false,                 // exclusive
		false,                  // no-local
		false,                   // no-wait
		nil,                       // args
	)
	if err != nil {
		log.Error(err)
		log.Error("Failed to register a consumer")
		bmq.refreshConnectionAndChannel(channelContext)
	}
	return msgs
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

