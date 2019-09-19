package connpool

import (
	"github.com/streadway/amqp"
)

var amqpURL = "amqp://wonly:Wl2016822@139.196.221.163:5672/"

func mqFactory() (interface{}, error) {
	return amqp.Dial(amqpURL)
}

func mqClose(conn interface{}) error {
	return conn.(*amqp.Connection).Close()
}

func mqIsValid(conn interface{}) bool {
	mqConn := conn.(*amqp.Connection)
	return !mqConn.IsClosed()
}

func NewMQPool(conf ConnPoolConfig) ConnPooler {
	return NewDBPool(mqFactory, mqClose, mqIsValid, conf)
}
