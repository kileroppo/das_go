package tuya2srv

import (
	"context"
	"das/core/entity"
	"das/core/rabbitmq"
	"encoding/base64"
	"encoding/json"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tyutils"
	"github.com/tidwall/gjson"

	pulsar "github.com/TuyaInc/tuya_pulsar_sdk_go"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tylog"
	"github.com/sirupsen/logrus"
	_ "github.com/tidwall/gjson"

	"das/core/log"
)

var consumer pulsar.Consumer

func Tuya2SrvStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Warningf("Tuya2SrvStart > %s", err)
		}
	}()

	log.Info("Tuya2SrvStart...")

	pulsar.SetInternalLogLevel(logrus.WarnLevel)
	opt := tylog.WithMaxSizeOption(10)
	tylog.SetGlobalLog("sdk", true, opt)

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

	//log.Infof("TuyaHandle.HandlePayload > recv: %s", jsonData)

	bizCode := gjson.GetBytes(jsonData, "bizCode").String()
	if len(bizCode) > 0 {
		t.sendOnOffLineMsg(jsonData)
	}

	devStatus := gjson.GetBytes(jsonData, "status").Array()

	for i,_ := range devStatus {
		switch devStatus[i].Get("code").String() {
		case "electricity_left":
			t.sendBattMsg(jsonData)
		case "clean_record":
			t.sendCleanRecordMsg(jsonData)
		}
	}

	return nil
}

func (t *TuyaHandle) sendBattMsg(jsonData []byte) {
    msg := entity.AlarmMsgBatt{}
    msg.Cmd = 0x2a
    msg.DevId = gjson.GetBytes(jsonData, "devId").String()
    msg.Vendor = "tuya"
    msg.DevType = "TYRobotCleaner"
    msg.Value = int(gjson.GetBytes(jsonData, "status").Array()[0].Get("value").Int())
    msg.Time = int32(gjson.GetBytes(jsonData, "status").Array()[0].Get("t").Int()/1000)

    data, err := json.Marshal(msg)
    if err != nil {
    	log.Warningf("TuyaHandle.sendBattMsg > json.Marshal > %s", err)
    } else {
    	rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaHandle) sendCleanRecordMsg(jsonData []byte) {
    msg := entity.TuyaMsg{}
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

func (t *TuyaHandle) sendOnOffLineMsg(jsonData []byte) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = gjson.GetBytes(jsonData, "devId").String()
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	onOff := gjson.GetBytes(jsonData, "bizCode").String()

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

func (t *TuyaHandle) decryptData(payload []byte) (jsonData []byte, err error) {
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
    consumer.Stop()
    log.Info("Tuya2Srv close")
}
