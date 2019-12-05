package rabbitmq

import (
	"sync"

	"github.com/dlintw/goconf"

	"../log"
)

var (
	producer2devMQ      *baseMq
	producer2appMQ      *baseMq
	producer2mnsMQ      *baseMq
	producer2pmsMQ      *baseMq
	producerGuard2appMQ *baseMq
	Consumer2devMQ      *baseMq
	Consumer2appMQ      *baseMq
	Consumer2aliMQ      *baseMq

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
		initProducerGuard2appMQ(conf)
		log.Info("RabbitMQ init")
	})
}

func Publish2app(data []byte, routingKey string) {
	if err := producer2appMQ.Publish(data, routingKey); err != nil {
		log.Warning("Publish2app error = ", err)
	} else {
		log.Debugf("RoutingKey = '%s', Publish2app msg: %s", routingKey, string(data))
	}
}

func Publish2dev(data []byte, routingKey string) {
	if err := producer2devMQ.Publish(data, routingKey); err != nil {
		log.Warning("Publish2dev error = ", err)
	} else {
		log.Debugf("RoutingKey = '%s', Publish2dev msg: %s", routingKey, string(data))
	}
}

func Publish2mns(data []byte, routingKey string) {
	if err := producer2mnsMQ.Publish(data, routingKey); err != nil {
		log.Warning("Publish2mns error = ", err)
	} else {
		log.Debug("Publish2mns msg: ", string(data))
	}
}

func Publish2pms(data []byte, routingKey string) {
	if err := producer2pmsMQ.Publish(data, routingKey); err != nil {
		log.Warning("Publish2pms error = ", err)
	} else {
		log.Debug("Publish2pms msg: ", string(data))
	}
}

func PublishGuard2app(data []byte, routingKey string) {
	if err := producerGuard2appMQ.Publish(data, routingKey); err != nil {
		log.Warning("PublishGuard2app error = ", err)
	} else {
		log.Debugf("RoutingKey = '%s', PublishGuard2app msg: %s", routingKey, string(data))
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
	}
	Consumer2aliMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
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
	}
	Consumer2appMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum: 5,
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
	}
	producer2appMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
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
	exchange, err := conf.GetString("rabbitmq", "device2db_ex")
	if err != nil {
		panic("initProducer2mnsMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "device2db_ex_type")
	if err != nil {
		panic("initProducer2mnsMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
	}
	producer2mnsMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
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
	}
	producer2pmsMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
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
	}
	Consumer2devMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
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

func initProducerGuard2appMQ(conf *goconf.ConfigFile) {
	uri, err := conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		panic("initProducerGuard2appMQ load uri conf error")
	}
	exchange, err := conf.GetString("rabbitmq", "guard2app_ex")
	if err != nil {
		panic("initProducerGuard2appMQ load exchange conf error")
	}
	exchangeType, err := conf.GetString("rabbitmq", "guard2app_ex_type")
	if err != nil {
		panic("initProducerGuard2appMQ load exchangeType conf error")
	}

	channelCtx := ChannelContext{
		Exchange:     exchange,
		ExchangeType: exchangeType,
		Durable:      false,
		AutoDelete:   false,
	}
	producerGuard2appMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
	}

	if err := producerGuard2appMQ.init(); err != nil {
		panic(err)
	}

	if err := producerGuard2appMQ.initExchange(); err != nil {
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
	}
	producer2devMQ = &baseMq{
		mqUri:      uri,
		channelCtx: channelCtx,
		reConnNum:  5,
	}

	if err := producer2devMQ.init(); err != nil {
		panic(err)
	}

	if err := producer2devMQ.initExchange(); err != nil {
		panic(err)
	}
}

func Close() {
	producer2appMQ.Close()
	producer2pmsMQ.Close()
	producer2mnsMQ.Close()
	producerGuard2appMQ.Close()
	producer2devMQ.Close()
	Consumer2devMQ.Close()
	Consumer2appMQ.Close()
	Consumer2aliMQ.Close()
	log.Info("RabbitMQ close")
}
