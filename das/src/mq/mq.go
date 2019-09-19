package mq

import (
	"../core/log"
	"sync"

	"github.com/streadway/amqp"

	"../connpool"
)

var mqpoolOnce sync.Once
var p sync.Pool
var MQConnPool connpool.ConnPooler

func init() {
	p = sync.Pool{
		New: func() interface{} {
			return new(MQChannel)
		},
	}

}

type MQConf struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool
}

type MQChannel struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool

	Channel *amqp.Channel
	Conn    *amqp.Connection
}

func NewMQChannel(conf MQConf) *MQChannel {
	ch := p.Get().(*MQChannel)
	ch.ExchangeType = conf.ExchangeType
	ch.Exchange = conf.Exchange
	ch.RoutingKey = conf.RoutingKey

	return ch

	// return &MQChannel{
	// 	Exchange:     conf.Exchange,
	// 	ExchangeType: conf.ExchangeType,
	// 	RoutingKey:   conf.RoutingKey,
	// 	Reliable:     false,
	// 	Durable:      false,
	// 	Channel:      nil,
	// 	Conn:         nil,
	// }

}

func (m *MQChannel) Init() error {

	mqpoolOnce.Do(func() {

		mqPoolConfig := connpool.ConnPoolConfig{
			80,
			40,
			0,
		}
		MQConnPool = connpool.NewMQPool(mqPoolConfig)
		MQConnPool.Begin()
	})

	conn := MQConnPool.Get().(*amqp.Connection)
	m.Conn = conn

	var err error
	m.Channel, err = m.Conn.Channel()

	if err != nil {
		log.Error("mq conn create channel failed: ", err)
		// linkpool.MQConnPool.Put(m.Conn)
		return err
	}

	err = m.Channel.ExchangeDeclare(m.Exchange, m.ExchangeType, true, false, false, false, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (m *MQChannel) InitbyConn(conn *amqp.Connection) error {
	m.Conn = conn

	var err error
	m.Channel, err = m.Conn.Channel()

	if err != nil {
		log.Error("mq conn create channel failed")
		// linkpool.MQConnPool.Put(m.Conn)
		return err
	}

	err = m.Channel.ExchangeDeclare(m.Exchange, m.ExchangeType, true, false, false, false, nil)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (m *MQChannel) Product2TempQueue(data []byte) error {

	queueName, err := m.Channel.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		amqp.Table{
			"x-expires": int32(1000),
		},
	)

	if err != nil {
		log.Error(err)
	}

	err = m.Channel.QueueBind(
		queueName.Name,
		m.RoutingKey,
		m.Exchange,
		false,
		nil,
	)

	if err != nil {
		log.Error(err)
	}

	err = m.Channel.Publish(
		m.Exchange,
		m.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "application/json",
			ContentEncoding: "",
			DeliveryMode:    amqp.Transient,
			Priority:        0,
			Body:            data,
		},
	)

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (m *MQChannel) Product2NormalQueue(data []byte) error {
	queueName, err := m.Channel.QueueDeclare(
		m.RoutingKey,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Error(err)
		return err
	}

	err = m.Channel.QueueBind(
		queueName.Name,
		m.RoutingKey,
		m.Exchange,
		false,
		nil,
	)

	if err != nil {
		log.Error(err)
		return err
	}

	err = m.Channel.Publish(
		m.Exchange,     // publish to an exchange
		queueName.Name, // routing to 0 or more queues
		false,          // mandatory
		false,          // immediate
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
		log.Error(err)
		return err
	}

	return nil
}

func (m *MQChannel) Close() {
	MQConnPool.Put(m.Conn)
	m.Channel.Close()
	//p.Put(m)
}

func (m *MQChannel) Consume() {

}

func CloseChannel(m *MQChannel) {
	p.Put(m)
}
