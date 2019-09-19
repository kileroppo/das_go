package mq

import (
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type MQConfig struct {
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Reliable     bool
	Durable      bool
}

type MQData struct {
	data []byte
	conf MQConfig
}

type MQChannelPool struct {
	conn *amqp.Connection
}

func NewMQChannelPool() MQChannelPool {
	return MQChannelPool{
		conn: nil,
	}
}

func (m *MQChannelPool) Init(amqpURL string) error {
	var err error
	m.conn, err = amqp.Dial(amqpURL)

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (m *MQChannelPool) Product(data []byte, conf MQConfig) error {

	channel, err := m.conn.Channel()

	if err != nil {
		log.Error(err)
		return err
	}

	err = channel.ExchangeDeclare(conf.Exchange, conf.ExchangeType, true, false, false, false, nil)

	if err != nil {
		log.Error(err)
		return err
	}

	queueName, err := channel.QueueDeclare(
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
		return nil
	}

	err = channel.QueueBind(
		queueName.Name,
		conf.RoutingKey,
		conf.Exchange,
		false,
		nil,
	)

	if err != nil {
		log.Error(err)
		return err
	}

	err = channel.Publish(
		conf.Exchange,
		conf.RoutingKey,
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

func (m *MQChannelPool) Close() error {
	m.conn.Close()
	return nil
}
