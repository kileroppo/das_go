package rabbitmq

import (
	"sync"
	"time"

	"github.com/json-iterator/go"
	"github.com/robertkowalski/graylog-golang"
	"github.com/streadway/amqp"
	"github.com/tidwall/gjson"
	"github.com/valyala/bytebufferpool"

	"das/core/entity"
	"das/core/log"
	"das/core/redis"
	"das/core/util"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	producerMQ *baseMq
	consumerMQ *baseMq
	grayCli    *gelf.Gelf

	OnceInitMQ sync.Once

	exSli  []string
	formatLog = `
{
      "version": "3.0",
      "host": "das",
      "full_message": "%s"
}
`
)

const (
	ExApp2Dev_Index     = 0
	ExDev2App_Index     = 1
	Ex2Mns_Index        = 2
	Ex2Pms_Index        = 3
	ExDev2Srv_Index     = 4
	ExSrv2Dev_Index     = 5
	ExAli2Srv_Index     = 6
	Ex2Log_Index        = 7
	ExSrv2Wonlyms_Index = 8
	Ex2PmsBeta_Index    = 9
)

func Init() {
	OnceInitMQ.Do(func() {
		initProducerMQ()
		initConsumerMQ()
		initGraylogConn()
		log.Info("RabbitMQ init")
	})
}

func initGraylogConn() {
	url,err := log.Conf.GetString("graylog", "url")
	if err != nil {
		panic(err)
	}
	port,err := log.Conf.GetInt("graylog", "port")
	if err != nil {
		panic(err)
	}


	cfg := gelf.Config{
		GraylogPort:     port,
		GraylogHostname: url,
	}
    grayCli = gelf.New(cfg)
    log.Info("Graylog init")
}

func initProducerMQ() {
	uri, err := log.Conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		log.Errorf("initProducerMQ > get rabbitmq_uri > %s", err)
		panic(err)
	}

	producerMQ = &baseMq{
		mqUri: uri,
	}

	if err = producerMQ.initConn(); err != nil {
		panic(err)
	}

	if err = producerMQ.initChannels(publishNum, 0); err != nil {
		panic(err)
	}

	initMQCfg()
}

func initConsumerMQ() {
	uri, err := log.Conf.GetString("rabbitmq", "rabbitmq_uri")
	if err != nil {
		log.Errorf("initProducerMQ > get rabbitmq_uri > %s", err)
		panic(err)
	}

	consumerMQ = &baseMq{
		mqUri: uri,
	}

	if err = consumerMQ.initConn(); err != nil {
		panic(err)
	}

	if err = consumerMQ.initChannels(0, consumerNum); err != nil {
		panic(err)
	}
}

func initMQCfg() {
	var err error
	exApp2dev, err := log.Conf.GetString("rabbitmq", "app2device_ex")
	exDev2App, err := log.Conf.GetString("rabbitmq", "device2app_ex")
	ex2Mns, err := log.Conf.GetString("rabbitmq", "device2mns_ex")
	ex2Pms, err := log.Conf.GetString("rabbitmq", "das2pms_ex")
	exDev2Srv, err := log.Conf.GetString("rabbitmq", "device2srv_ex")
	exSrv2Dev, err := log.Conf.GetString("rabbitmq", "srv2device_ex")
	exAli2Srv, err := log.Conf.GetString("rabbitmq", "ali2srv_ex")
	ex2Log, err := log.Conf.GetString("rabbitmq", "logSave_ex")
	exSrv2Wonlyms, err := log.Conf.GetString("rabbitmq", "srv2wonlyms_ex")
	ex2PmsBeta, err := log.Conf.GetString("rabbitmq_beta", "das2pms_ex")

	exTypeApp2dev, err := log.Conf.GetString("rabbitmq", "app2device_ex_type")
	exTypeDev2App, err := log.Conf.GetString("rabbitmq", "device2app_ex_type")
	exType2Mns, err := log.Conf.GetString("rabbitmq", "device2mns_ex_type")
	exType2Pms, err := log.Conf.GetString("rabbitmq", "das2pms_ex_type")
	exTypeDev2Srv, err := log.Conf.GetString("rabbitmq", "device2srv_ex_type")
	exTypeSrv2Dev, err := log.Conf.GetString("rabbitmq", "srv2device_ex_type")
	exTypeAli2Srv, err := log.Conf.GetString("rabbitmq", "ali2srv_ex_type")
	exType2Log, err := log.Conf.GetString("rabbitmq", "logSave_ex_type")
	exTypeSrv2Wonlyms, err := log.Conf.GetString("rabbitmq", "srv2wonlyms_ex_type")
	exType2PmsBeta, err := log.Conf.GetString("rabbitmq_beta", "das2pms_ex_type")

	queApp2dev, err := log.Conf.GetString("rabbitmq", "app2device_que")
	que2Mns, err := log.Conf.GetString("rabbitmq", "device2mns_que")
	que2Pms, err := log.Conf.GetString("rabbitmq", "das2pms_que")
	queDev2Srv, err := log.Conf.GetString("rabbitmq", "device2srv_que")
	queAli2Srv, err := log.Conf.GetString("rabbitmq", "ali2srv_que")
	que2Log, err := log.Conf.GetString("rabbitmq", "logSave_que")
	queSrv2Wonlyms, err := log.Conf.GetString("rabbitmq", "srv2wonlyms_que")
	que2PmsBeta, err := log.Conf.GetString("rabbitmq_beta", "das2pms_que")

	exSli = []string{exApp2dev, exDev2App, ex2Mns, ex2Pms, exDev2Srv, exSrv2Dev, exAli2Srv, ex2Log, exSrv2Wonlyms, ex2PmsBeta}
	exTypeSli := []string{exTypeApp2dev, exTypeDev2App, exType2Mns, exType2Pms, exTypeDev2Srv, exTypeSrv2Dev, exTypeAli2Srv, exType2Log, exTypeSrv2Wonlyms, exType2PmsBeta}
	queSli := []string{queApp2dev, "", que2Mns, que2Pms, queDev2Srv, "", queAli2Srv, que2Log, queSrv2Wonlyms, que2PmsBeta}

	exCfg := exchangeCfg{
		name:       "",
		kind:       "",
		durable:    true,
		autoDelete: false,
		internal:   false,
		noWait:     false,
	}

	queCfg := queueCfg{
		name:       "consumerQueue",
		key:        "",
		exchange:   "",
		durable:    true,
		autoDelete: false,
		exclusive:  false,
		noWait:     false,
	}

	for i, _ := range exSli {
		exCfg.name = exSli[i]
		exCfg.kind = exTypeSli[i]
		if err = producerMQ.initExchange(producerMQ.publishCh[0], &exCfg); err != nil {
			panic(err)
		}
		if len(queSli[i]) > 0 {
			queCfg.name = queSli[i]
			queCfg.exchange = exSli[i]
			if err = producerMQ.initQueue(producerMQ.publishCh[0], &queCfg); err != nil {
				panic(err)
			}
		}
	}
}

func publishDirect(index int, mq *baseMq, ex, routingKey string, data []byte) error {
	return mq.publishSafe(index, ex, routingKey, data)
}

func Publish2dev(data []byte, routingKey string) {
	if err := publishDirect(0, producerMQ, exSli[ExSrv2Dev_Index], routingKey, data); err != nil {
		log.Warningf("Publish2dev > %s", err)
	} else {
		//log.Debugf("RoutingKey = '%s', Publish2dev msg: %s", routingKey, string(data))
	}
}

func Publish2app(data []byte, routingKey string) {
	if err := publishDirect(1, producerMQ, exSli[ExDev2App_Index], routingKey, data); err != nil {
		log.Warningf("Publish2app > %s", err)
	} else {
		sendRabbitMQUpDataLog(data)
		//log.Debugf("RoutingKey = '%s', Publish2app msg: %s", routingKey, string(data))
	}
}

func Publish2mns(data []byte, routingKey string) {
	if err := publishDirect(2, producerMQ, exSli[Ex2Mns_Index], routingKey, data); err != nil {
		log.Warningf("Publish2mns > %s", err)
	} else {
		//log.Debugf("Publish2mns msg: %s", data)
	}
}

func Publish2pms(data []byte, routingKey string) {
	var err error
	if redis.IsDevBeta(data) {
		err = publishDirect(3, producerMQ,exSli[Ex2PmsBeta_Index], routingKey, data)
	} else {
		err = publishDirect(3, producerMQ, exSli[Ex2Pms_Index], routingKey, data)
	}

	if err != nil {
		log.Warningf("Publish2pms > %s", err)
	} else {
		//log.Debugf("Publish2pms msg: %s", data)
	}
}

func Publish2log(data []byte, routingKey string) {
	if err := publishDirect(4, producerMQ, exSli[Ex2Log_Index], routingKey, data); err != nil {
		log.Warningf("Publish2log > %s", err)
	}
	msg := entity.GrayLog{
		Version: "3.0",
		Host:    "das",
		Message: util.Bytes2Str(data),
	}
	b,err := json.Marshal(msg)
	if err == nil {
		grayCli.Log(util.Bytes2Str(b))
	}
}

func Publish2wonlyms(data []byte, routingKey string) {
	if err := publishDirect(5, producerMQ, exSli[ExSrv2Wonlyms_Index], routingKey, data); err != nil {
		log.Warningf("Publish2wonlyms > %s", err)
	}
}

func ConsumeApp() (ch <-chan amqp.Delivery, err error){
	queName, _ := log.Conf.GetString("rabbitmq", "app2device_que")
	ch, err = consumerMQ.consume(0, queName, "")
	if err != nil {
		err = consumerMQ.reConn()
		if err != nil {
			return
		} else {
			return consumerMQ.consume(0, queName, "")
		}
	}
	return
}

func ConsumeDev() (ch <-chan amqp.Delivery, err error){
	queName, _ := log.Conf.GetString("rabbitmq", "device2srv_que")
	ch, err = consumerMQ.consume(1, queName, "")
	if err != nil {
		err = consumerMQ.reConn()
		if err != nil {
			return
		} else {
			return consumerMQ.consume(1, queName, "")
		}
	}
	return
}

func ConsumeAli() (ch <-chan amqp.Delivery, err error){
	queName, _ := log.Conf.GetString("rabbitmq", "ali2srv_que")
	ch, err = consumerMQ.consume(2, queName, "")
	if err != nil {
		err = consumerMQ.reConn()
		if err != nil {
			return
		} else {
			return consumerMQ.consume(2, queName, "")
		}
	}
	return
}

func Close() {
	producerMQ.Close()
	consumerMQ.Close()
	log.Info("RabbitMQ close")
}

func sendRabbitMQUpDataLog(byteData []byte) {
	devId := gjson.GetBytes(byteData, "devId").String()
	if len(devId) == 0 {
		return
	}

	var logMsg entity.SysLogMsg
	currT := time.Now()
	logMsg.Timestamp = currT.Unix()
	logMsg.NanoTimestamp = currT.UnixNano()
	logMsg.MsgType = 4
	logMsg.MsgName = "上行设备数据"
	logMsg.UUid = devId
	logMsg.VendorName = "RabbitMQ"

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	buf.WriteString("Json数据：")
	buf.Write(byteData)

	logMsg.RawData = buf.String()
	rawData,err := json.Marshal(logMsg)
	if err != nil {
		log.Warningf("sendRabbitMQUpDataLog > json.Marshal > %s", err)
	} else {
		Publish2log(rawData, "")
	}
}