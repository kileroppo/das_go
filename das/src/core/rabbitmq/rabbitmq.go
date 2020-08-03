package rabbitmq

import (
	"sync"

	"github.com/dlintw/goconf"

	"das/core/log"
)

var (
	producer2devMQ      *baseMq
	producer2appMQ      *baseMq
	producer2mnsMQ      *baseMq
	producer2pmsMQ      *baseMq
	producer2logMQ      *baseMq
	Consumer2devMQ      *baseMq
	Consumer2appMQ      *baseMq
	Consumer2aliMQ      *baseMq
	producer2wonlymsMQ  *baseMq

	OnceInitMQ sync.Once
)

func Init(conf *goconf.ConfigFile) {
	OnceInitMQ.Do(func() {
		initConsumer2aliMQ(conf)
		initConsumer2appMQ(conf)
		initConsumer2devMQ(conf)
		initProducer2appMQ(conf)
		initProducer2devMQ(conf)
		initProducer2pmsMQ(conf)
		initProducer2mnsMQ(conf)
		initProducer2logMQ(conf)
		initProduct2wonlymsMQ(conf)
		log.Info("RabbitMQ init")
	})
}

func Publish2app(data []byte, routingKey string) {
	if err := producer2appMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2app > %s", err)
	} else {
		//log.Debugf("RoutingKey = '%s', Publish2app msg: %s", routingKey, string(data))
	}
}

func Publish2dev(data []byte, routingKey string) {
	if err := producer2devMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2dev > %s", err)
	} else {
		//log.Debugf("RoutingKey = '%s', Publish2dev msg: %s", routingKey, string(data))
	}
}

func Publish2mns(data []byte, routingKey string) {
	if err := producer2mnsMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2mns > %s", err)
	} else {
		//log.Debugf("Publish2mns msg: %s", data)
	}
}

func Publish2pms(data []byte, routingKey string) {
	if err := producer2pmsMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2pms > %s", err)
	} else {
		//log.Debugf("Publish2pms msg: %s", data)
	}
}

func Publish2log(data []byte, routingKey string) {
	if err := producer2logMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2log > %s", err)
	}
}

func Publish2wonlyms(data []byte, routingKey string) {
	if err := producer2wonlymsMQ.Publish(data, routingKey); err != nil {
		log.Warningf("Publish2wonlyms > %s", err)
	}
}

func initConsumer2aliMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initConsumer2aliMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "ali2srv_ex")
	if err != nil {
		panic("initConsumer2aliMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "ali2srv_ex_type")
	if err != nil {
		panic("initConsumer2aliMQ load exchangeType conf error")
	}
	queueName, err := conf.GetString("rabbitmq", "ali2srv_que")
	if err != nil {
		panic("initConsumer2aliMQ load routingKey conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		RoutingKey:   "",
		QueueName:    queueName,
		Durable:      true,
		AutoDelete:   false,
		Name:         "Consumer2aliMQ",
	}
	Consumer2aliMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		isConsumer: true,
	}

	if err := Consumer2aliMQ.init(); err != nil {
		panic(err)
	}

	if err := Consumer2aliMQ.initExchange(); err != nil {
		panic(err)
	}

	if err := Consumer2aliMQ.initConsumer(); err != nil {
		panic(err)
	}
}

func initConsumer2appMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initConsumer2appMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "app2device_ex")
	if err != nil {
		panic("initConsumer2appMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "app2device_ex_type")
	if err != nil {
		panic("initConsumer2appMQ load exchangeType conf error")
	}
	queueName, err := conf.GetString("rabbitmq", "app2device_que")
	if err != nil {
		panic("initConsumer2appMQ load routingKey conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		RoutingKey:   "",
		QueueName:    queueName,
		Durable:      true,
		AutoDelete:   false,
		Name:         "Consumer2appMQ",
	}
	Consumer2appMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		isConsumer: true,
	}

	if err := Consumer2appMQ.init(); err != nil {
		panic(err)
	}

	if err := Consumer2appMQ.initExchange(); err != nil {
		panic(err)
	}

	if err := Consumer2appMQ.initConsumer(); err != nil {
		panic(err)
	}
}

func initProducer2appMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducer2appMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "device2app_ex")
	if err != nil {
		panic("initProducer2appMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "device2app_ex_type")
	if err != nil {
		panic("initProducer2appMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2appMQ",
	}
	producer2appMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2appMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2appMQ.initExchange(); err != nil {
		panic(err)
	}
}

func initProducer2mnsMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducer2mnsMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "device2mns_ex")
	if err != nil {
		panic("initProducer2mnsMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "device2mns_ex_type")
	if err != nil {
		panic("initProducer2mnsMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2mnsMQ",
	}
	producer2mnsMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2mnsMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2mnsMQ.initExchange(); err != nil {
		panic(err)
	}
}

func initProducer2pmsMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducer2pmsMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "das2pms_ex")
	if err != nil {
		panic("initProducer2pmsMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "das2pms_ex_type")
	if err != nil {
		panic("initProducer2pmsMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2pmsMQ",
	}
	producer2pmsMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2pmsMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2pmsMQ.initExchange(); err != nil {
		panic(err)
	}
}

func initConsumer2devMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initConsumer2devMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "device2srv_ex")
	if err != nil {
		panic("initConsumer2devMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "device2srv_ex_type")
	if err != nil {
		panic("initConsumer2devMQ load exchangeType conf error")
	}
	queueName, err := conf.GetString("rabbitmq", "device2srv_que")
	if err != nil {
		panic("initConsumer2devMQ load queue conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		RoutingKey:   "",
		QueueName:    queueName,
		Durable:      true,
		AutoDelete:   false,
		Name:         "Consumer2devMQ",
	}
	Consumer2devMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		isConsumer: true,
	}

	if err := Consumer2devMQ.init(); err != nil {
		panic(err)
	}

	if err := Consumer2devMQ.initExchange(); err != nil {
		panic(err)
	}

	if err := Consumer2devMQ.initConsumer(); err != nil {
		panic(err)
	}
}

func initProducer2devMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducer2devMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "srv2device_ex")
	if err != nil {
		panic("initProducer2devMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "srv2device_ex_type")
	if err != nil {
		panic("initProducer2devMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2devMQ",
	}
	producer2devMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2devMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2devMQ.initExchange(); err != nil {
		panic(err)
	}
}

func initProducer2logMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducer2logMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "logSave_ex")
	if err != nil {
		panic("initProducer2logMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "logSave_ex_type")
	if err != nil {
		panic("initProducer2logMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2logMQ",
	}
	producer2logMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2logMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2logMQ.initExchange(); err != nil {
		panic(err)
	}
}

func initProduct2wonlymsMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProduct2wonlymsMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "srv2wonlyms_ex")
	if err != nil {
		panic("initProduct2wonlymsMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "srv2wonlyms_ex_type")
	if err != nil {
		panic("initProduct2wonlymsMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
		Name:         "Producer2WonlymsMQ",
	}
	producer2wonlymsMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
	}

	if err := producer2wonlymsMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2wonlymsMQ.initExchange(); err != nil {
		panic(err)
	}
}

func Close() {
	producer2appMQ.Close()
	producer2pmsMQ.Close()
	producer2mnsMQ.Close()
	producer2devMQ.Close()
	producer2logMQ.Close()
	Consumer2devMQ.Close()
	Consumer2appMQ.Close()
	Consumer2aliMQ.Close()
	producer2wonlymsMQ.Close()
	log.Info("RabbitMQ close")
}
