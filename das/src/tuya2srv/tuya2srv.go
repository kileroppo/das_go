package tuya2srv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"

	pulsar "github.com/TuyaInc/tuya_pulsar_sdk_go"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tylog"
	"github.com/TuyaInc/tuya_pulsar_sdk_go/pkg/tyutils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/util"
	"das/feibee2srv"
)

var (
	consumer pulsar.Consumer
	alarmFilter MsgFilter
)

func Init() {
	alarmFilter = &RedisFilter{}
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
	eventHandle, ok := TyDevEventHandlers[bizCode]
	if ok {
		eventHandle(devId, bizCode, gjson.GetBytes(t.data, "bizData"))
	}

	devStatus := gjson.GetBytes(t.data, "status").Array()
	for i, _ := range devStatus {
		statusCode := devStatus[i].Get("code").String()
		if statusHandle, ok := TyDevStatusHandlers[statusCode]; ok {
			statusHandle(devId, devStatus[i])
		}
	}
}

func (t *TuyaMsgHandle) send2Others(devId string, oriData []byte) {
	t.msg2Others.DevId = devId
	t.msg2Others.OriData = util.Bytes2Str(oriData)

	data, err := json.Marshal(t.msg2Others)
	if err == nil {
		rabbitmq.Publish2mns(data, "")
		//rabbitmq.Publish2pms(data, "")
	}
}

func TyEventOnlineHandle(devId, tyEvent string, rawJsonData gjson.Result) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = devId
	msg.Vendor = "tuya"
	if tyEvent == Ty_Event_Online {
		msg.Time = 1
	} else {
		msg.Time = 0
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.TyEventOnlineHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
		rabbitmq.Publish2pms(data, "")
	}

	msg2app := entity.Feibee2DevMsg{}
	msg2app.Cmd = 0xfb
	msg2app.Vendor = "tuya"
	msg2app.DevId = devId
	msg2app.OpType = "newOnline"

	if tyEvent == Ty_Event_Online {
		msg2app.OpValue = "1"
		msg2app.Online = 1
	} else {
		msg2app.OpValue = "0"
	}

	feibee2srv.RecordDevOnlineStatus(msg.DevId, msg2app.Online)
	data, err = json.Marshal(msg2app)
	if err == nil {
		rabbitmq.Publish2app(data, devId)
		rabbitmq.Publish2mns(data, "")
	}
}

func TyEventDeleteHandle(devId, tyEvent string, rawJsonData gjson.Result) {
    msg := entity.Feibee2DevMsg{}
    msg.Cmd = 0xfb
    msg.DevId = devId
    msg.OpType = "devDelete"
    msg.Vendor = "tuya"

    data, err := json.Marshal(msg)
    if err == nil {
    	rabbitmq.Publish2mns(data, "")
	}
}

func TyStatusDevBatt(devId string, rawJsonData gjson.Result) {
    msg := entity.OtherVendorDevMsg{}
    msg.Cmd = 0x1200
    msg.DevId = devId
    msg.Vendor = "tuya"
    msg.OriData = rawJsonData.Raw
    data, err := json.Marshal(msg)
    if err == nil {
    	rabbitmq.Publish2pms(data, "")
	}
}

func TyStatusPowerHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.DeviceActive{}
	msg.Cmd = 0x46
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	if rawJsonData.Get("value").Bool() {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.TyStatusPowerHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func TyStatusRobotCleanerBattHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.AlarmMsgBatt{}
	msg.Cmd = 0x2a
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Value = int(rawJsonData.Get("value").Int())
	msg.Time = int64(rawJsonData.Get("t").Int() / 1000)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.TyStatusRobotCleanerBattHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func TyStatusNormalHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.Feibee2DevMsg{}
	msg.Cmd = 0xfb
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"

	msg.OpType = rawJsonData.Get("code").String()
	msg.OpValue = rawJsonData.Get("value").String()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.TyStatusNormalHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func TyStatusAlarmSensorHandle(devId string, rawJsonData gjson.Result) {
	tyAlarmType := rawJsonData.Get("code").String()
	alarmFlag := 0
	rawAlarmVal := rawJsonData.Get("value").String()
	if alarmVal, ok := TySensorAlarmReflect[tyAlarmType]; ok {
		if rawAlarmVal == alarmVal {
			alarmFlag = 1
		}
	}
	timestamp := rawJsonData.Get("t").Int()
	tySensorDataNotify(devId, tyAlarmType, alarmFlag, timestamp)
}

func TyStatusEnvSensorHandle(devId string, rawJsonData gjson.Result) {
	tyAlarmType := rawJsonData.Get("code").String()
	timestamp := rawJsonData.Get("t").Int()
	alarmFlag := rawJsonData.Get("value").Int()

	tySensorDataNotify(devId, tyAlarmType, int(alarmFlag), timestamp)
}

func TyStatusSceneHandle(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = 0xf1
	msg.DevId = devId + rawJsonData.Get("code").String()
	msg.AlarmType = "sceneSwitch"
	msg.AlarmFlag = 1

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2pms(data, "")
	}
}

func TyStatus2PMSHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.OtherVendorDevMsg{
		Header: entity.Header{
			Cmd:     0x1200,
			Ack:     0,
			DevType: "",
			DevId:   devId,
			Vendor:  "tuya",
			SeqId:   0,
		},
		OriData: "",
	}
	msg.OriData = rawJsonData.Raw

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2pms(data, "")
	}
}

func tySensorDataNotify(devId, tyAlarmType string, alarmFlag int, timestamp int64) {
	var msg entity.Feibee2AlarmMsg
	msg.Cmd = 0xfc
	msg.Vendor = "tuya"
	msg.Time = int(timestamp) / 1000
	msg.MilliTimestamp = int(timestamp)
	msg.DevId = devId

	var ok bool
	msg.AlarmType, ok = TySensor2WonlySensor[tyAlarmType]
	if !ok {
		msg.AlarmType = tyAlarmType
	}
	msg.AlarmFlag = alarmFlag
	alarmVal, ok :=  constant.SensorVal2Str[msg.AlarmType]
	if ok {
		msg.AlarmValue = alarmVal[msg.AlarmFlag]
	} else {
		divisor, ok := TyEnvSensorValDivisor[tyAlarmType]
		if ok {
			msg.Divisor = divisor
			msg.AlarmValue = strconv.FormatFloat(float64(alarmFlag)/float64(divisor), 'f', 2, 64)
		} else {
			msg.Divisor = 1
			msg.AlarmValue = strconv.Itoa(alarmFlag)
		}
	}

	//todo: 涂鸦设备上报周期为1min，是否增加设备报警过滤？
	//if alarmFilter.Exists(devId + msg.AlarmType) {
	//	return
	//} else {
	//	alarmFilter.Set(devId + msg.AlarmType, time.Minute * 30)
	//}

	data, err := json.Marshal(msg)
	if err == nil {
		if msg.AlarmFlag == 1 && ok && msg.AlarmType != constant.Wonly_Status_Sensor_Doorcontact {
			rabbitmq.Publish2mns(data, "")
		}
		rabbitmq.Publish2pms(data, "")
	}

	msg.Cmd = 0xf1
	data, err = json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2pms(data, "")
	}
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
	log.Info("Tuya2Srv close")
}
