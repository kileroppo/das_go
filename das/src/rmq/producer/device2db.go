package producer

import (
	"../../core/log"
	"../../core/rabbitmq"
	"github.com/dlintw/goconf"
)

var rmq_uri_mgo string
var exchange_mgo string         // = "OneNET2APP"
var exchange_mgo2 string
var exchangeType_mgo string     // = "direct"
var routingKey_mgo string = ""  // 设备的uuid
var routingKey_mgo2 string = "" // 设备的uuid

//初始化RabbitMQ交换器，消息队列名称
func InitRmq_Ex_Que_Name_mongo(conf *goconf.ConfigFile) {
	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
	if rmq_uri == "" {
		log.Error("未启用RabbitMq")
		return
	}
	exchange_mgo, _ = conf.GetString("rabbitmq", "Device2Db_ex")
	exchangeType_mgo, _ = conf.GetString("rabbitmq", "device2db_ex_type")
	routingKey_mgo, _ = conf.GetString("rabbitmq", "device2db_que")

	exchange_mgo2, _ = conf.GetString("rabbitmq", "Device2Db_ex2")
	routingKey_mgo2, _ = conf.GetString("rabbitmq", "device2db_que2")
}

func SendMQMsg2Db(message string) {
	if rabbitmq.ProducerRabbitMq2Db == nil {
		log.Error("SendMQMsg2Db: rabbitmq.ProducerRabbitMq2Db is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange_mgo, ExchangeType: exchangeType_mgo, RoutingKey: routingKey_mgo, Reliable: true, Durable: true, ReSendNum: 0}

	log.Info("rabbitmq.ProducerRabbitMq.Publish2Db: ", message)
	rabbitmq.ProducerRabbitMq2Db.Publish2Db(&channleContxt, message)
}

func SendMQMsg2Db2(message string) {
	if rabbitmq.ProducerRabbitMq2Db == nil {
		log.Error("SendMQMsg2Db2: rabbitmq.ProducerRabbitMq2Db is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange_mgo2, ExchangeType: exchangeType_mgo, RoutingKey: routingKey_mgo2, Reliable: true, Durable: true, ReSendNum: 0}

	log.Info("rabbitmq.ProducerRabbitMq.Publish2Db: ", message)
	rabbitmq.ProducerRabbitMq2Db.Publish2Db(&channleContxt, message)
}
