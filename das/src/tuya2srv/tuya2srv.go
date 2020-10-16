package tuya2srv

import (
	"context"
	"das/core/util"
	"encoding/base64"
	"encoding/json"
	"strconv"

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
	consumer pulsar.Consumer
)

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

	handler := TuyaMsgHandle{
		data: jsonData,
		msg2Others: entity.OtherVendorDevMsg{
			Header: entity.Header{
				Cmd:     0x1200,
				Ack:     0,
				DevType: "",
				DevId:   "",
				Vendor:  "tuya",
				SeqId:   0,
			},
			OriData: "",
		},
	}
	handler.MsgHandle()
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
	data       []byte
	msg2Others entity.OtherVendorDevMsg
}

func (t *TuyaMsgHandle) UpdateData(data []byte) {
	t.data = data
}

func (t *TuyaMsgHandle) InitHeader() {
	t.msg2Others = entity.OtherVendorDevMsg{
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
	rabbitmq.SendGraylogByMQ("涂鸦Server-pulsar->DAS: %s", t.data)
	devId := gjson.GetBytes(t.data, "devId").String()

	rabbitmq.Publish2app(t.data, devId)
	t.send2Others(devId, t.data)

	bizCode := gjson.GetBytes(t.data, "bizCode").String()
	switch bizCode {
	case Ty_Event_Online, Ty_Event_Offline:
		tyDevOnOffHandle(devId, bizCode)
	}

	devStatus := gjson.GetBytes(t.data, "status").Array()
	for i, _ := range devStatus {
		if handle, ok := TyHandleMap[devStatus[i].Get("code").String()]; ok {
			handle(devId, devStatus[i])
		}
	}
}

func (t *TuyaMsgHandle) send2Others(devId string, oriData []byte) {
	t.msg2Others.DevId = devId
	t.msg2Others.OriData = util.Bytes2Str(oriData)

	data, err := json.Marshal(t.msg2Others)
	if err == nil {
		rabbitmq.Publish2mns(data, "")
		rabbitmq.Publish2pms(data, "")
	}
}

func tyDevOnlineHandle(devId string, res gjson.Result) {
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
		log.Warningf("TuyaCallback.tyDevOnOffHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func tyDevBattHandle(devId string, res gjson.Result) {
	msg := entity.AlarmMsgBatt{}
	msg.Cmd = 0x2a
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Value = int(res.Get("value").Int())
	msg.Time = int32(res.Get("t").Int() / 1000)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.tyDevBattHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func tyDevOnOffHandle(devId, jsonData string) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	onOff := jsonData

	if onOff == Ty_Event_Online {
		msg.Time = 1
	} else  {
		msg.Time = 0
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.tyDevOnOffHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func tyDevStatusHandle(devId string, res gjson.Result) {
	msg := entity.Feibee2DevMsg{}
	msg.Cmd = 0xfb
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	msg.OpType = res.Get("code").String()
	msg.OpValue = res.Get("value").String()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.tyDevStatusHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func tyAlarmSensorHandle(devId string, rawJsonData gjson.Result) {
	tyAlarmType := rawJsonData.Get("code").String()
	alarmFlag := 0
	rawAlarmVal := rawJsonData.Get("value").String()
	if rawAlarmVal == "alarm" || rawAlarmVal == "presence" || rawAlarmVal == "true" {
		alarmFlag = 1
	}
	if alarmFlag == 0 && tyAlarmType != Ty_Status_Doorcontact_State {
		return
	}
	timestamp := rawJsonData.Get("t").Int()
	tySensorDataNotify(devId, tyAlarmType, alarmFlag, timestamp)
}

func tyEnvSensorHandle(devId string, rawJsonData gjson.Result) {
	tyAlarmType := rawJsonData.Get("code").String()
	timestamp := rawJsonData.Get("t").Int()
	alarmFlag := rawJsonData.Get("value").Int()

	tySensorDataNotify(devId, tyAlarmType, int(alarmFlag), timestamp)
}

func tySensorDataNotify(devId, tyAlarmType string, alarmFlag int, timestamp int64) {
	var msg entity.Feibee2AlarmMsg
	msg.Cmd = 0xfc
	msg.Vendor = "tuya"
	msg.Time = int(timestamp) / 1000
	msg.MilliTimestamp = int(timestamp)
	msg.DevId = devId

	msg.AlarmType = TySensor2WonlySensor[tyAlarmType]
	msg.AlarmFlag = alarmFlag
	alarmVal,ok := SensorVal2Str[msg.AlarmType]
	if ok {
		msg.AlarmValue = alarmVal[msg.AlarmFlag]
	} else {
		msg.AlarmValue = strconv.Itoa(alarmFlag)
	}

	data,err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2mns(data, "")
	}
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
	log.Info("Tuya2Srv close")
}
