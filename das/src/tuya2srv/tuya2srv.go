package tuya2srv

import (
	"context"
	"das/core/util"
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

var (
	consumer       pulsar.Consumer
	tuyaMsgHandler TuyaMsgHandle
)

func Init() {
	go Tuya2SrvStart()
	tuyaMsgHandler = TuyaMsgHandle{}
	tuyaMsgHandler.InitHeader()
}

func Tuya2SrvStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Warningf("Tuya2SrvStart > %s", err)
		}
	}()
	log.Info("Tuya2SrvStart...")
	pulsar.SetInternalLogLevel(logrus.DebugLevel)
	tylog.SetGlobalLog("tuyaSDK", true)

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

	consumer.ReceiveAndHandle(context.Background(), &TuyaCallback{secret: accessKey[8:24]})
}

type TuyaCallback struct {
	secret string
}

func (t *TuyaCallback) HandlePayload(ctx context.Context, msg *pulsar.Message, payload []byte) error {
	jsonData, err := t.decryptData(payload)
	if err != nil {
		return err
	}

	tuyaMsgHandler.UpdateData(jsonData)
	tuyaMsgHandler.MsgHandle()
	return nil
}

func (t *TuyaCallback) decryptData(payload []byte) (jsonData []byte, err error) {
	//log.Infof("TuyaCallback rawData > recv: %s", payload)

	val := gjson.GetBytes(payload, "data")

	jsonData, err = base64.StdEncoding.DecodeString(val.String())
	if err != nil {
		log.Warningf("TuyaCallback.decryptData > base64.StdEncoding.DecodeString > %s", err)
		return
	}

	jsonData = tyutils.EcbDecrypt(jsonData, []byte(t.secret))
	return
}

type TuyaMsgHandle struct {
	data    []byte
	msg2msn entity.OtherVendorDevMsg
}

func (t *TuyaMsgHandle) UpdateData(data []byte) {
	t.data = data
}

func (t *TuyaMsgHandle) InitHeader() {
	t.msg2msn = entity.OtherVendorDevMsg{
		Header: entity.Header{
			Cmd:     0x1200,
			Ack:     0,
			DevType: "",
			DevId:   "",
			Vendor:  "tuya",
			SeqId:   0,
		},
		OriData: "",
	}
}

func (t *TuyaMsgHandle) MsgHandle() {
	rabbitmq.SendGraylogByMQ("tuyaServer-pulsar->DAS: %s", t.data)
	devId := gjson.GetBytes(t.data, "devId").String()
	rabbitmq.Publish2app(t.data, devId)
	t.send2MNS(devId, t.data)

	bizCode := gjson.GetBytes(t.data, "bizCode").String()
	if len(bizCode) > 0 {
		t.sendOnOffLineMsg(devId, bizCode)
	}

	devStatus := gjson.GetBytes(t.data, "status").Array()

	for i, _ := range devStatus {
		switch devStatus[i].Get("code").String() {
		case "electricity_left":
			t.sendBattMsg(devId, devStatus[i])
		case "clean_record":
			t.sendCleanRecordMsg(t.data)
		case "power":
			t.sendOfflineMsg(devId, devStatus[i])
		case "status":
			t.sendNotifyAct(devId, devStatus[i])
		}
	}
}

func (t *TuyaMsgHandle) send2MNS(devId string, oriData []byte) {
	t.msg2msn.DevId = devId
	t.msg2msn.OriData = util.Bytes2Str(oriData)

	data, err := json.Marshal(t.msg2msn)
	if err == nil {
		rabbitmq.Publish2mns(data, "")
	}
}

func (t *TuyaMsgHandle) sendOfflineMsg(devId string, res gjson.Result) {
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
		log.Warningf("TuyaCallback.sendOnOffLineMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaMsgHandle) sendBattMsg(devId string, res gjson.Result) {
	msg := entity.AlarmMsgBatt{}
	msg.Cmd = 0x2a
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Value = int(res.Get("value").Int())
	msg.Time = int32(res.Get("t").Int() / 1000)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.sendBattMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaMsgHandle) sendCleanRecordMsg(jsonData []byte) {
	msg := entity.OtherVendorDevMsg{}
	msg.Cmd = 0x1200
	msg.DevId = gjson.GetBytes(jsonData, "devId").String()
	msg.DevType = "TYRobotCleaner"
	msg.Vendor = "tuya"
	msg.OriData = string(jsonData)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.sendCleanRecordMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}

func (t *TuyaMsgHandle) sendOnOffLineMsg(devId, jsonData string) {
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
		log.Warningf("TuyaCallback.sendOnOffLineMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func (t *TuyaMsgHandle) sendNotifyAct(devId string, res gjson.Result) {
	msg := entity.Feibee2DevMsg{}
	msg.Cmd = 0xfb
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	msg.OpType = res.Get("code").String()
	msg.OpValue = res.Get("value").String()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.sendNotifyAct > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
	log.Info("Tuya2Srv close")
}
