package feibee2srv

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/json-iterator/go"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ErrAlarmMsgValue = errors.New("feibeeAlarm value was not supported")

	alarmMsgTyp = map[int]func(string) (int, int, string){
		0x00010020: parseVoltageVal,
		0x00010021: parseBatteryVal,
		0x05000080: parseSensorVal,
		0x04020000: parseTempAndHuminityVal,
		0x04050000: parseTempAndHuminityVal,
		0x04000000: parseIlluminanceVal,
	}

	alarmName = map[MsgType]string{
		InfraredSensor:     "infrared",
		DoorMagneticSensor: "doorContact",
		SmokeSensor:        "smoke",
		FloodSensor:        "flood",
		GasSensor:          "gas",
		IlluminanceSensor:  "illuminance",
		SosBtnSensor:       "sosButton",
	}

	alarmValue = map[MsgType]([]string){
		InfraredSensor:     []string{"无人", "有人"},
		DoorMagneticSensor: []string{"关闭", "开启"},
		SmokeSensor:        []string{"无烟", "有烟"},
		FloodSensor:        []string{"无水", "有水"},
		GasSensor:          []string{"无气体", "有气体"},
		SosBtnSensor:       []string{"正常", "报警"},
	}
)

type BaseSensorAlarm struct {
	feibeeMsg *entity.FeibeeData
	msgType   MsgType

	alarmType         string
	alarmVal          string
	removalAlarmValue string

	devType string
	devid   string
	time    int
	bindid  string

	alarmFlag   int
	removalFlag int
}

func (self *BaseSensorAlarm) parseAlarmMsg() error {
	self.devType = devTypeConv(self.feibeeMsg.Records[0].Deviceid, self.feibeeMsg.Records[0].Zonetype)
	self.devid = self.feibeeMsg.Records[0].Uuid
	self.time = int(time.Now().Unix())
	self.bindid = self.feibeeMsg.Records[0].Bindid

	parse, ok := alarmMsgTyp[getSpMsgKey(self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid)]
	if !ok {
		return ErrAlarmMsgValue
	}

	self.removalFlag, self.alarmFlag, self.alarmVal = parse(self.feibeeMsg.Records[0].Value)

	if self.alarmFlag == 0 || self.alarmFlag == 1 {
		if sli,ok := alarmValue[self.msgType]; ok {
			self.alarmVal = sli[self.alarmFlag]
		}
	}
	self.alarmType = alarmName[self.msgType]
	return nil
}

func (self *BaseSensorAlarm) PushMsg() {
	if err := self.parseAlarmMsg(); err != nil {
		return
	}
	if self.alarmFlag == 1 || self.msgType == DoorMagneticSensor {
		self.pushMsg2mns()
	}
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
	self.pushForcedBreakMsg()
}

func (self *BaseSensorAlarm) pushMsg2app() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2app() error = ", err)
		return
	}
	rabbitmq.Publish2app(data, self.devid)
}

func (self *BaseSensorAlarm) pushMsg2mns() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2mns() error = ", err)
		return
	}
	rabbitmq.Publish2mns(data, "")
}

func (self *BaseSensorAlarm) createMsg2pmsForSence() entity.FeibeeAutoScene2pmsMsg {
	var msg entity.FeibeeAutoScene2pmsMsg

	msg.Cmd = 0xf1
	msg.Ack = 0
	msg.Vendor = "feibee"
	msg.SeqId = 1

	msg.DevType = self.devType
	msg.DevId = self.devid

	msg.TriggerType = 0
	msg.Time = self.time

	msg.AlarmValue = self.alarmVal
	msg.AlarmType = self.alarmType

	return msg
}

func (self *BaseSensorAlarm) createMsg2app() entity.FeibeeAlarm2AppMsg {
	var msg entity.FeibeeAlarm2AppMsg

	msg.Cmd = 0xfc
	msg.Ack = 0
	msg.DevType = self.devType
	msg.DevId = self.devid
	msg.Vendor = "feibee"
	msg.SeqId = 1

	msg.AlarmType = self.alarmType
	msg.AlarmValue = self.alarmVal
	msg.Time = self.time
	msg.Bindid = self.bindid
	msg.AlarmFlag = self.alarmFlag

	return msg
}

func (self *BaseSensorAlarm) pushMsg2pmsForSave() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2pmsForSave() error = ", err)
		return
	}
	rabbitmq.Publish2pms(data, "")
}

func (self *BaseSensorAlarm) pushMsg2pmsForSceneTrigger() {
	msg := self.createMsg2pmsForSence()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2pmsForSceneTrigger() error = ", err)
		return
	}
	rabbitmq.Publish2pms(data, "")
}

func (self *BaseSensorAlarm) pushForcedBreakMsg() {
	if self.removalFlag > 0 {
		msg := self.createMsg2app()

		msg.AlarmType = "forcedBreak"
		msg.AlarmValue = "传感器被强拆"

		data, err := json.Marshal(msg)
		if err != nil {
			log.Error("BaseSensorAlarm pushForcedBreakMsg() error = ", err)
			return
		}
		rabbitmq.Publish2mns(data, "")
		rabbitmq.Publish2pms(data, "")

		msgForScene := self.createMsg2pmsForSence()
		msgForScene.AlarmType = "forcedBreak"
		msgForScene.AlarmValue = "传感器被强拆"

		data, err = json.Marshal(msgForScene)
		if err != nil {
			log.Error("BaseSensorAlarm pushForcedBreakMsg() error = ", err)
			return
		}
		rabbitmq.Publish2pms(data, "")
	}
}

type TemperAndHumiditySensorAlarm struct {
	BaseSensorAlarm
}

func (self *TemperAndHumiditySensorAlarm) PushMsg() {
	if err := self.parseAlarmMsg(); err != nil {
		return
	}

	cid, aid := self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid

	if cid == 1026 && aid == 0 {
		self.alarmType = "temperature"
	} else if cid == 1029 && aid == 0 {
		self.alarmType = "humidity"
	} else {
		return
	}

	self.pushMsg2mns()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
	self.pushForcedBreakMsg()
}

func parseTempAndHuminityVal(val string) (removalAlarmFlag, alarmFlag int, alarmVal string) {
	if len(val) != 4 {
		return -1, -1, ""
	}

	v64, err := strconv.ParseUint(val, 16, 64)
	if err != nil {
		return -1, -1, ""
	}
	res := make([]byte, 2)
	binary.LittleEndian.PutUint16(res, uint16(v64))

	alarmVal = hex.EncodeToString(res)
	v64, err = strconv.ParseUint(alarmVal, 16, 64)
	if err != nil {
		return -1, -1, ""
	}

	alarmVal = strconv.FormatFloat(float64(v64)/100, 'f', 2, 64)
	alarmFlag = 1
	removalAlarmFlag = -1
	return
}

func parseIlluminanceVal(val string) (removalAlarmFlag, alarmFlag int, alarmVal string) {
	if len(val) != 4 {
		return -1, -1, ""
	}

	v64, err := strconv.ParseUint(val, 16, 64)
	if err != nil {
		return -1, -1, ""
	}
	res := make([]byte, 2)
	binary.LittleEndian.PutUint16(res, uint16(v64))

	alarmVal = hex.EncodeToString(res)
	v64, err = strconv.ParseUint(alarmVal, 16, 64)
	if err != nil {
		return -1, -1, ""
	}

	alarmVal = strconv.Itoa(int(v64))
	alarmFlag = 1
	removalAlarmFlag = -1
	return
}

func parseSensorVal(val string) (removalAlarmFlag, alarmFlag int, alarmVal string) {
	bitFlagInt, err := strconv.ParseInt(val[0:2], 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return -1, -1, ""
	}

	if int(bitFlagInt)&3 > 0 {
		alarmFlag = 1
	}

	removalAlarmFlag = int(bitFlagInt) & 4
	return
}

func parseBatteryVal(val string) (removalAlarmFlag, alarmFlag int, alarmVal string) {
	valInt, err := strconv.ParseInt(val, 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return -1, -1, ""
	}

	alarmVal = strconv.Itoa(int(valInt) / 2)
	alarmFlag = 1
	return
}

func parseVoltageVal(val string) (removalAlarmFlag, alarmFlag int, alarmVal string) {
	valInt, err := strconv.ParseInt(val, 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return
	}

	alarmVal = strconv.Itoa(int(valInt) / 10)
	alarmFlag = 1
	return
}
