package tuya2srv

import (
	"context"
	"encoding/base64"
	"encoding/json"

	pulsar "github.com/TuyaInc/tuya_pulsar_sdk_go"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tylog"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tyutils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)

var consumer pulsar.Consumer

func Init() {
	go Tuya2SrvStart()
}

func Tuya2SrvStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Warningf("Tuya2SrvStart > %s", err)
		}
	}()

	log.Info("Tuya2SrvStart...")

	pulsar.SetInternalLogLevel(logrus.DebugLevel)
	opt := tylog.WithMaxSizeOption(10)
	tylog.SetGlobalLog("tuyaSDK", true, opt)

	accessId, err := log.Conf.GetString("tuya", "accessID")
	if err != nil {
		log.Warningf("Tuya2SrvStart > log.Conf.GetString > %s", err)
		return
	}
	accessKey, err := log.Conf.GetString("tuya", "accessKey")
	if err != nil {
		log.Warningf("Tuya2SrvStart > log.Conf.GetString > %s", err)
		return
	}

	topic := pulsar.TopicForAccessID(accessId)
	cli := pulsar.NewClient(pulsar.ClientConfig{PulsarAddr: pulsar.PulsarAddrCN})

	csmConf := pulsar.ConsumerConfig{
		Topic: topic,
		Auth:  pulsar.NewAuthProvider(accessId, accessKey),
	}

	consumer, err = cli.NewConsumer(csmConf)
	if err != nil {
		log.Warningf("Tuya2SrvStart > cli.NewConsumer > %s", err)
		return
	}

	consumer.ReceiveAndHandle(context.Background(), &TuyaHandle{secret: accessKey[8:24]})
}

type TuyaHandle struct {
	secret string
}

func (t *TuyaHandle) HandlePayload(ctx context.Context, msg *pulsar.Message, payload []byte) error {
	jsonData, err := t.decryptData(payload)
	if err != nil {
		return err
	}

	rabbitmq.SendGraylogByMQ("DAS receive from tuyaServer: %s", jsonData)
    devId := gjson.GetBytes(jsonData, "devId").String()
    rabbitmq.Publish2app(payload, devId)
	bizCode := gjson.GetBytes(jsonData, "bizCode").String()
	if len(bizCode) > 0 {
		t.sendOnOffLineMsg(devId, bizCode)
	}

	devStatus := gjson.GetBytes(jsonData, "status").Array()

	for i,_ := range devStatus {
		switch devStatus[i].Get("code").String() {
		case "electricity_left":
			t.sendBattMsg(devId, devStatus[i])
		case "clean_record":
			t.sendCleanRecordMsg(jsonData)
		case "power":
			t.sendOfflineMsg(devId, devStatus[i])
		case "status":
			t.sendNotifyAct(devId, devStatus[i])
		}
	}

	return nil
}

func (t *TuyaHandle) sendOfflineMsg(devId string, res gjson.Result) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	if res.Get("value").Bool() {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaHandle.sendOnOffLineMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaHandle) sendBattMsg(devId string, res gjson.Result) {
    msg := entity.AlarmMsgBatt{}
    msg.Cmd = 0x2a
    msg.DevId = devId
    msg.Vendor = "tuya"
    msg.DevType = "TYRobotCleaner"
    msg.Value = int(res.Get("value").Int())
    msg.Time = int32(res.Get("t").Int()/1000)

    data, err := json.Marshal(msg)
    if err != nil {
    	log.Warningf("TuyaHandle.sendBattMsg > json.Marshal > %s", err)
    } else {
    	rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaHandle) sendCleanRecordMsg(jsonData []byte) {
    msg := entity.OtherVendorDevMsg{}
    msg.Cmd = 0x1200
    msg.DevId = gjson.GetBytes(jsonData, "devId").String()
	msg.DevType = "TYRobotCleaner"
    msg.Vendor = "tuya"
    msg.OriData = string(jsonData)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaHandle.sendCleanRecordMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}

func (t *TuyaHandle) sendOnOffLineMsg(devId, jsonData string) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	onOff := jsonData

	if onOff == "online" {
		msg.Time = 1
	} else if onOff == "offline" {
		msg.Time = 0
	} else {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaHandle.sendOnOffLineMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaHandle) sendNotifyAct(devId string, res gjson.Result) {
    msg := entity.Feibee2DevMsg{}
    msg.Cmd = 0xfb
    msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	msg.OpType = res.Get("code").String()
	msg.OpValue = res.Get("value").String()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaHandle.sendNotifyAct > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaHandle) decryptData(payload []byte) (jsonData []byte, err error) {
	//log.Infof("TuyaHandle rawData > recv: %s", payload)

	val := gjson.GetBytes(payload, "data")

	jsonData, err = base64.StdEncoding.DecodeString(val.String())
	if err != nil {
		log.Warningf("TuyaHandle.decryptData > base64.StdEncoding.DecodeString > %s", err)
		return
	}

	jsonData = tyutils.EcbDecrypt(jsonData, []byte(t.secret))
	return
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
    log.Info("Tuya2Srv close")
}

//type tuyaClientImpl struct {
//	pool *manage.ClientPool
//	Addr string
//}
//
//func (c *tuyaClientImpl) NewConsumer(config manage.ConsumerConfig) (pulsar.Consumer, error) {
//	tylog.Info("start creating consumer",
//		tylog.String("pulsar", c.Addr),
//		tylog.String("topic", config.Topic),
//	)
//
//	errs := make(chan error, 10)
//	go func() {
//		for err := range errs {
//			tylog.Error("async errors", tylog.ErrorField(err))
//		}
//	}()
//	cfg := manage.ConsumerConfig{
//		ClientConfig: manage.ClientConfig{
//			Addr:       c.Addr,
//			AuthData:   config.Auth.AuthData(),
//			AuthMethod: config.Auth.AuthMethod(),
//			TLSConfig: &tls.Config{
//				InsecureSkipVerify: true,
//			},
//			Errs: errs,
//		},
//		Topic:              config.Topic,
//		SubMode:            manage.SubscriptionModeFailover,
//		Name:               subscriptionName(config.Topic),
//		NewConsumerTimeout: time.Minute,
//	}
//	p := c.GetPartition(config.Topic, cfg.ClientConfig)
//
//	// partitioned topic
//	if p > 0 {
//		list := make([]*consumerImpl, 0, p)
//		originTopic := cfg.Topic
//		for i := 0; i < p; i++ {
//			cfg.Topic = fmt.Sprintf("%s-partition-%d", originTopic, i)
//			mc := manage.NewManagedConsumer(c.pool, cfg)
//			list = append(list, &consumerImpl{
//				csm:     mc,
//				topic:   cfg.Topic,
//				stopped: make(chan struct{}),
//			})
//		}
//		consumerList := &ConsumerList{
//			list:             list,
//			FlowPeriodSecond: DefaultFlowPeriodSecond,
//			FlowPermit:       DefaultFlowPermit,
//			Topic:            config.Topic,
//			Stopped:          make(chan struct{}),
//		}
//		return consumerList, nil
//	}
//
//	// single topic
//	mc := manage.NewManagedConsumer(c.pool, config)
//	tylog.Info("create consumer success",
//		tylog.String("pulsar", c.Addr),
//		tylog.String("topic", config.Topic),
//	)
//	return &consumerImpl{
//		csm:     mc,
//		topic:   cfg.Topic,
//		stopped: make(chan struct{}),
//	}, nil
//
//}