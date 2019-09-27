package producer

import (
	"github.com/dlintw/goconf"

	"../../core/log"
	"../../core/rabbitmq"
)

var rmq_uri_mgo string
var exchange_mgo string // = "OneNET2APP"
var exchange_pms string
var exchangeType_mgo string    // = "direct"
var routingKey_mgo string = "" // 设备的uuid
var routingKey_pms string = "" // 设备的uuid

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

	exchange_pms, _ = conf.GetString("rabbitmq", "das2pms_ex")
	routingKey_pms, _ = conf.GetString("rabbitmq", "das2pms_que")
}

func SendMQMsg2Db(message string) {
	if rabbitmq.ProducerRabbitMq2Db == nil {
		log.Error("SendMQMsg2Db: rabbitmq.ProducerRabbitMq2Db is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange_mgo, ExchangeType: exchangeType_mgo, RoutingKey: routingKey_mgo, Reliable: true, Durable: true, ReSendNum: 0}

	log.Debug("rabbitmq.ProducerRabbitMq.Publish2Db: ", message)
	err := rabbitmq.ProducerRabbitMq2Db.Publish2Db(channleContxt, []byte(message))

	if err != nil {
		log.Warning("ProducerRabbitMq2Db.Publish2Db() error = ", err)
	}

}

func SendMQMsg2PMS(message string) {
	if rabbitmq.ProducerRabbitMq2Db == nil {
		log.Error("SendMQMsg2PMS: rabbitmq.ProducerRabbitMq2Db is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange_pms, ExchangeType: exchangeType_mgo, RoutingKey: routingKey_pms, Reliable: true, Durable: true, ReSendNum: 0}

	log.Debug("rabbitmq.ProducerRabbitMq.Publish2PMS: ", message)
	err := rabbitmq.ProducerRabbitMq2Db.Publish2PMS(channleContxt, []byte(message))

	if err != nil {
		log.Warning("ProducerRabbitMq2Db.Publish2PMS() error = ", err)
	}
}
