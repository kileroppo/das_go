package tuya2srv

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	pulsar "github.com/tuya/tuya-pulsar-sdk-go"
	"github.com/tuya/tuya-pulsar-sdk-go/pkg/tylog"
	"github.com/tuya/tuya-pulsar-sdk-go/pkg/tyutils"

	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/util"
	"das/feibee2srv"
	"das/filter"
)

var (
	consumer pulsar.Consumer
)

func Init() {
	loadFilterRulesFromMySql()
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
	// 通知mns
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
	msg.Timestamp = rawJsonData.Get("time").Int()
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

	alarmFlag := rawJsonData.Get("value").Int()
	if alarmFlag >= LowBatteryThreshold {
		return
	}
	alarmType := constant.Wonly_Status_Low_Power
	alarmVal := rawJsonData.Get("value").String()

	_, notifyFlag := filter.SensorFilter(devId, alarmType, "", alarmFlag)
	if notifyFlag {
		alarmMsg := entity.Feibee2AlarmMsg{
			Header:         entity.Header{
				Cmd:     constant.Device_Sensor_Msg,
				Ack:     0,
				DevType: "",
				DevId:   devId,
				Vendor:  "tuya",
				SeqId:   0,
			},
			Time:           0,
			MilliTimestamp: 0,
			TriggerType:    0,
			AlarmType:      alarmType,
			AlarmValue:     alarmVal,
			AlarmFlag:      0,
			Bindid:         "",
			CycleFlag:      false,
			Multiplier:     0,
		}

		data,err := json.Marshal(alarmMsg)
		if err != nil {
			return
		}
		rabbitmq.Publish2mns(data, "")
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
	rawTimestamp := rawJsonData.Get("t").Int()
	msg := entity.Feibee2DevMsg{}
	msg.Cmd = constant.Device_Normal_Msg
	msg.DevId = devId
	msg.Vendor = "tuya"
	msg.DevType = "TYRobotCleaner"
	msg.Time = int(correctSensorMillTimestamp(rawTimestamp) / 1000)

	msg.OpType = Ty_Status
	val := rawJsonData.Get("value").String()

	msg.OpValue = val
	note, ok := TyCleanerStatusNote[val]
	if ok {
		msg.Note = note
	} else {
		return
	}

	if !tyStatusPriorityFilter(devId, rawTimestamp, val) {
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
	alarmFlag := 0
	// 涂鸦报警类型
	tyAlarmType := rawJsonData.Get("code").String()
	// 报警value
	rawAlarmVal := rawJsonData.Get("value").String()

	if alarmVal, ok := TySensorAlarmReflect[tyAlarmType]; ok {
		if rawAlarmVal == alarmVal {
			alarmFlag = 1
		}
	}
	timestamp := rawJsonData.Get("t").Int()
	tySensorDataNotify(devId, tyAlarmType, alarmFlag, timestamp, &rawJsonData)
}

func TyStatusAlarmStateHandle(devId string, rawJsonData gjson.Result) {
	timestamp := rawJsonData.Get("t").Int()
	val := rawJsonData.Get("value").Int()
	alarmFlag := 0
	if val != 4 {
		alarmFlag = 1
	}
	tySensorDataNotify(devId, constant.Wonly_Status_Audible_Alarm, alarmFlag, timestamp, &rawJsonData)
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

	tySensorDataNotify(devId, tyAlarmType, int(alarmFlag), timestamp, nil)
}

func TyStatusSceneHandle(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	code := rawJsonData.Get("code").String()
	sceneNum := Ty_Status_Switch_1
	switch code {
	case Ty_Status_Scene_1:
		sceneNum = Ty_Status_Switch_1
	case Ty_Status_Scene_2:
		sceneNum = Ty_Status_Switch_2
	case Ty_Status_Scene_3:
		sceneNum = Ty_Status_Switch_3
	case Ty_Status_Scene_4:
		sceneNum = Ty_Status_Switch_4
	case Ty_Status_Scene_5:
		sceneNum = Ty_Status_Switch_5
	case Ty_Status_Scene_6:
		sceneNum = Ty_Status_Switch_6
	}
	msg.DevId = devId + sceneNum
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

func TyStatusSwitchValHandle(devId string, rawJsonData gjson.Result) {
	var msg entity.Feibee2AutoSceneMsg
	msg.Cmd = constant.Scene_Trigger
	code := rawJsonData.Get("code").String()
	sceneNum := Ty_Status_Scene_1
	switch code {
	case Ty_Status_Switch1_Val:
		sceneNum = Ty_Status_Switch_1
	case Ty_Status_Switch2_Val:
		sceneNum = Ty_Status_Switch_2
	case Ty_Status_Switch3_Val:
		sceneNum = Ty_Status_Switch_3
	case Ty_Status_Switch4_Val:
		sceneNum = Ty_Status_Switch_4
	case Ty_Status_Switch5_Val:
		sceneNum = Ty_Status_Switch_5
	case Ty_Status_Switch6_Val:
		sceneNum = Ty_Status_Switch_6
	}
	msg.DevId = devId + sceneNum
	msg.AlarmType = "sceneSwitch"
	msg.AlarmFlag = 1

	data, err := json.Marshal(msg)
	if err == nil {
		rabbitmq.Publish2Scene(data, "")
	}
}
// 涂鸦场景或数据通知
func tySensorDataNotify(devId, tyAlarmType string, alarmFlag int, timestamp int64, rawJsonData *gjson.Result) {
	// 1 校验时间
	correctT := correctSensorMillTimestamp(timestamp)
	// 1.1 塞入常用字段
	var msg entity.Feibee2AlarmMsg
	msg.Cmd = constant.Device_Sensor_Msg
	msg.Vendor = "tuya"
	msg.Time = int(correctT / 1000)
	msg.MilliTimestamp = int(correctT)
	msg.DevId = devId

	var ok bool
	// 1.2 b报警类型
	msg.AlarmType, ok = TySensor2WonlySensor[tyAlarmType]
	if !ok {
		msg.AlarmType = tyAlarmType
	}
	// 1.2.1 报警 --- 涂鸦 api返回的 value  （以下简称 value）
	msg.AlarmFlag = alarmFlag
	// ty:code---[gas] ---- []string{"检测正常", "燃气浓度已超标，正在报警"}
	alarmVal, ok :=  constant.SensorVal2Str[msg.AlarmType]
	if ok {
		//  找到文字版描述 用 value 在文字版数组中选一个
		msg.AlarmValue = alarmVal[msg.AlarmFlag]
	} else {
		// 传感器元器件敏感度  type == code
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
	notifyFlag, triggerFlag := filter.SensorFilter(msg.DevId, msg.AlarmType, msg.AlarmValue, msg.AlarmFlag)
	fmt.Print("notifyFlag",notifyFlag)
	fmt.Print("triggerFlag",triggerFlag)
	if notifyFlag {
		tySensorRawDataNotify(devId, rawJsonData)
		data, err := json.Marshal(msg)
		if err == nil {
			if msg.AlarmType == constant.Wonly_Status_Sensor_Doorcontact {
				rabbitmq.Publish2mns(data, "")
			} else if ok && msg.AlarmFlag == 1 {
				rabbitmq.Publish2mns(data, "")
			}
			rabbitmq.Publish2pms(data, "")
		}
	}

	if triggerFlag {
		msg.Cmd = constant.Scene_Trigger
		data, err := json.Marshal(msg)
		if err == nil {
			rabbitmq.Publish2Scene(data, "")
		}
	}
}

func tySensorRawDataNotify(devId string, rawJsonData *gjson.Result) {
	if rawJsonData == nil {
		return
	}
	singleStatus := entity.TuyaDevStaus{
		Code:  "",
		Value: nil,
		T:     0,
	}
	err := json.Unmarshal(util.Str2Bytes(rawJsonData.String()), &singleStatus)
	if err != nil {
		return
	}

	rawMsg := entity.TuyaRawStatusMsg{
		DataId:     "",
		DevId:      devId,
		ProductKey: "",
		Status:     []entity.TuyaDevStaus{singleStatus,},
	}

	data,err := json.Marshal(rawMsg)
	if err != nil {
		return
	}

	tyRawDataNotify(devId, data)
}

func TyDataFilterAndNotify(devId string, rawData []byte) {
	bizCode := gjson.GetBytes(rawData, "bizCode").String()
	if len(bizCode) > 0 && tyEventFilter(bizCode) {
		tyRawDataNotify(devId, rawData)
		return
	}

	devStatus := gjson.GetBytes(rawData, "status").Array()
	var statusCode string
	for i, _ := range devStatus {
		statusCode = devStatus[i].Get("code").String()
		if tyStatusFilter(statusCode) {
			tyRawDataNotify(devId, rawData)
			return
		}
	}
}

func tyRawDataNotify(devId string, rawData []byte) {
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

func tyStatusFilter(statusCode string) bool {
     _,ok := TyStatusDataFilterMap[statusCode]
     return ok
}

func tyEventFilter(bizCode string) bool {
	_,ok := TyEventDataFilterMap[bizCode]
	return ok
}

func Close() {
	if consumer != nil {
		consumer.Stop()
	}
	log.Info("Tuya2Srv close")
}
