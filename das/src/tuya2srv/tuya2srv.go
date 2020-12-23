package tuya2srv

import (
	"context"
	"das/filter"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

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
	alarmFilter filter.MsgFilter
)

func Init() {
	alarmFilter = &filter.RedisFilter{}
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
}

func (t *TuyaMsgHandle) MsgHandle() {
	devId := gjson.GetBytes(t.data, "devId").String()

	//rabbitmq.Publish2app(t.data, devId)
	//t.send2Others(devId, t.data)
	TyDataFilterAndNotify(devId, t.data)

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

	var esLog entity.EsLogEntiy // 记录日志
	esLog.DeviceId = devId
	esLog.Vendor = "tuya"
	esLog.Operation = "涂鸦pulsar推送"
	esLog.RetMsg = TyDevEventOperZh[bizCode]
	esLog.ThirdPlatform = "涂鸦pulsar"
	esLog.RawData = util.Bytes2Str(t.data)
	esData, err := json.Marshal(esLog)
	if err != nil {
		log.Warningf("MsgHandle > json.Marshal > %s", err)
		return
	}
	// rabbitmq.SendGraylogByMQ("涂鸦Server-pulsar->DAS: %s", t.data)
	rabbitmq.SendGraylogByMQ("%s", esData)
}

func TyEventOnlineHandle(devId, tyEvent string, rawJsonData gjson.Result) {
	msg := entity.DeviceActive{}
	msg.Cmd = constant.Upload_lock_active
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
	msg2app.Cmd = constant.Device_Normal_Msg
	msg2app.Vendor = "tuya"
	msg2app.DevId = devId
	msg2app.OpType = "newOnline"

	if tyEvent == Ty_Event_Online {
		msg2app.OpValue = "1"
		msg2app.Online = 1
	} else {
		msg2app.OpValue = "0"
	}

	feibee2srv.RecordDevOnlineStatus(msg.DevId, "tuya", msg2app.Online)
	data, err = json.Marshal(msg2app)
	if err == nil {
		rabbitmq.Publish2app(data, devId)
		rabbitmq.Publish2mns(data, "")
	}
}

func TyEventDeleteHandle(devId, tyEvent string, rawJsonData gjson.Result) {
    msg := entity.Feibee2DevMsg{}
    msg.Cmd = constant.Device_Normal_Msg
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
    msg.Cmd = constant.Other_Vendor_Msg
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
	msg.Cmd = constant.Upload_lock_active
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
	msg.Cmd = constant.Low_battery_alarm
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Value = int(rawJsonData.Get("value").Int())
	msg.Time = int32(rawJsonData.Get("t").Int() / 1000)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("TuyaCallback.TyStatusRobotCleanerBattHandle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg.DevId)
	}
}

func TyStatusNormalHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.Feibee2DevMsg{}
	msg.Cmd = constant.Device_Normal_Msg
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Time = int(correctSensorMillTimestamp(rawJsonData.Get("t").Int()) / 1000)

	msg.OpType = Ty_Status
	val := rawJsonData.Get("value").String()
	msg.OpValue = val
	note, ok := TyCleanerStatusNote[val]
	if ok {
		msg.Note = note
	} else {
		return
	}

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

func correctSensorMillTimestamp(millTimestamp int64) int64 {
	curr := time.Now().UnixNano() / 1000_000
	if millTimestamp > curr {
		return curr
	} else {
		return millTimestamp
	}
}

func TyStatusEnvSensorHandle(devId string, rawJsonData gjson.Result) {
	tyAlarmType := rawJsonData.Get("code").String()
	timestamp := rawJsonData.Get("t").Int()
	alarmFlag := rawJsonData.Get("value").Int()

	tySensorDataNotify(devId, tyAlarmType, int(alarmFlag), timestamp)
}

func TyStatusSceneHandle(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	msg.DevId = devId + rawJsonData.Get("code").String()
	msg.AlarmType = "sceneSwitch"
	msg.AlarmFlag = 1

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func TyStatusSleepStage(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	msg.DevId = devId
	msg.AlarmType = constant.Wonly_Status_Sleep_Stage
	msg.Time = int(rawJsonData.Get("t").Int()/1000)
	tyVal := rawJsonData.Get("value").String()
	if tyVal == Ty_Sleep_Stage_Awake {
		msg.AlarmFlag = 1
	} else if tyVal == Ty_Sleep_Stage_Sleep {
		msg.AlarmFlag = 4
	} else {
		return
	}
	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func TyStatusOffBed(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	msg.DevId = devId
	msg.AlarmType = constant.Wonly_Status_Sleep_Stage
	msg.Time = int(rawJsonData.Get("t").Int()/1000)
	offBed := rawJsonData.Get("value").Bool()
	if offBed {
		msg.AlarmFlag = 6 //离床
	} else {
		msg.AlarmFlag = 5 //在床
	}

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func TyStatusWakeup(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	msg.DevId = devId
	msg.AlarmType = constant.Wonly_Status_Sleep_Stage
	msg.Time = int(rawJsonData.Get("t").Int()/1000)
	wakeup := rawJsonData.Get("value").Bool()
	if wakeup {
		msg.AlarmFlag = 1 //清醒
	} else {
		return //睡眠
	}

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func TyStatus2PMSHandle(devId string, rawJsonData gjson.Result) {
	msg := entity.OtherVendorDevMsg{
		Header: entity.Header{
			Cmd:     constant.Other_Vendor_Msg,
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

func TyStatusCurtainHandle(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	msg.DevId = devId
	msg.AlarmType = constant.Wonly_Status_Curtain
	msg.Time = int(rawJsonData.Get("t").Int()/1000)

	val := rawJsonData.Get("value").String()
	if val == Ty_Cmd_Work_State_Val_Open {
		msg.AlarmFlag = 1
	} else {
		msg.AlarmFlag = 0
	}
	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func tySensorDataNotify(devId, tyAlarmType string, alarmFlag int, timestamp int64) {
	correctT := correctSensorMillTimestamp(timestamp)
	var msg entity.Feibee2AlarmMsg
	msg.Cmd = constant.Device_Sensor_Msg
	msg.Vendor = "tuya"
	msg.Time = int(correctT / 1000)
	msg.MilliTimestamp = int(correctT)
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
		multiplier, ok := TyEnvSensorValTransfer[tyAlarmType]
		if ok {
			msg.Multiplier = multiplier
			msg.AlarmValue = strconv.FormatFloat(float64(alarmFlag)*float64(multiplier), 'f', 2, 64)
		} else {
			msg.Multiplier = 1
			msg.AlarmValue = strconv.Itoa(alarmFlag)
		}
	}

	//todo: 涂鸦报警过滤
	if !tyAlarmMsgFilter(msg.DevId, tyAlarmType, msg.AlarmFlag) {
		return
	}

	data, err := json.Marshal(msg)
	if err == nil {
		if msg.AlarmType == constant.Wonly_Status_Sensor_Doorcontact {
			rabbitmq.Publish2mns(data, "")
		} else if ok && msg.AlarmFlag == 1 {
			rabbitmq.Publish2mns(data, "")
		}
		rabbitmq.Publish2pms(data, "")
	}

	msg.Cmd = constant.Scene_Trigger
	data, err = json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}

func TyDataFilterAndNotify(devId string, rawData []byte) {
	devStatus := gjson.GetBytes(rawData, "status").Array()
	var statusCode string
	for i, _ := range devStatus {
		statusCode = devStatus[i].Get("code").String()
		if tyDataFilter(statusCode) {
			msg2mns := entity.OtherVendorDevMsg{
				Header: entity.Header{
					Cmd:     constant.Other_Vendor_Msg,
					Ack:     0,
					DevType: "",
					DevId:   devId,
					Vendor:  "tuya",
					SeqId:   0,
				},
				OriData: util.Bytes2Str(rawData),
			}
			data2mns,err := json.Marshal(msg2mns)
			if err == nil {
				rabbitmq.Publish2mns(data2mns, "")
			}
			return
		}
	}
}

func tyDataFilter(statusCode string) bool {
     _,ok := TyStatusDataFilterMap[statusCode]
     return ok
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
	log.Info("Tuya2Srv close")
}
