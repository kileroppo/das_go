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

	ErrAlarmMsg = errors.New("Feibee alarm message was invalid")
)

type parseFunc func(string, MsgType, int) (int, int, string, string)

type BaseSensorAlarm struct {
	feibeeMsg    *entity.FeibeeData
	msgType      MsgType
	alarmMsgType int

	alarmType string
	alarmVal  string

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
	self.alarmMsgType = getSpMsgKey(self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid)

	parse := parseFunc(nil)
	ok := false

	parse, ok = alarmMsgTyp[self.alarmMsgType]
	if !ok {
		parse = parseContinuousVal
	}

	self.removalFlag, self.alarmFlag, self.alarmVal, self.alarmType = parse(self.feibeeMsg.Records[0].Value, self.msgType, self.alarmMsgType)

	if self.alarmFlag < 0 {
		return ErrAlarmMsg
	}

	return nil
}

func (self *BaseSensorAlarm) PushMsg() {
	if err := self.parseAlarmMsg(); err != nil {
		log.Warning("BaseSensorAlarm PushMsg() error = ", err)
		return
	}
	if !(self.alarmFlag == 0) {
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

func (self *BaseSensorAlarm) createMsg2pmsForSence() entity.Feibee2AutoSceneMsg {
	var msg entity.Feibee2AutoSceneMsg

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

func (self *BaseSensorAlarm) createMsg2app() entity.Feibee2AlarmMsg {
	var msg entity.Feibee2AlarmMsg

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

func (self *BaseSensorAlarm) getDisplayName() (res string) {
	if sli, ok := alarmValueMapByTyp[self.msgType]; ok {
		if self.alarmFlag < len(sli) {
			res = sli[self.alarmFlag]
		} else {
			return res
		}
	}
	return
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

func parseTempAndHuminityVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
	alarmVal = Little2BigEndianString(val)
	if len(alarmVal) == 0 {
		return -1, -1, "", ""
	}
	v64, err := strconv.ParseUint(alarmVal, 16, 64)
	if err != nil {
		return -1, -1, "", ""
	}

	alarmVal = strconv.FormatFloat(float64(v64)/100, 'f', 2, 64)
	alarmFlag = int(v64)
	removalAlarmFlag = -1
	alarmName = varAlarmName[valType]
	return
}

func parseContinuousVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
	alarmVal = Little2BigEndianString(val)
	if len(alarmVal) == 0 {
		return -1, -1, "", ""
	}
	v64, err := strconv.ParseUint(alarmVal, 16, 64)
	if err != nil {
		return -1, -1, "", ""
	}
	if valType == illuminance {
		if v64 > 1000 {
			v64 = 1000
		}
	}

	alarmFlag = int(v64)
	alarmVal = getAlarmValName(msgType, valType, alarmFlag)
	removalAlarmFlag = -1
	alarmName = varAlarmName[valType]
	return
}

func parseSensorVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
	bitFlagInt, err := strconv.ParseInt(val[0:2], 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return -1, -1, "", ""
	}

	if msgType == SosBtnSensor {
		alarmFlag = (int(bitFlagInt) & 0b0000_0010) >> 1
	} else {
		alarmFlag = int(bitFlagInt) & 0b0000_0001
	}

	//todo: 周期上报数据不透传
	if cycleFlag := (bitFlagInt & 0b0001_0000); cycleFlag > 0{
		alarmFlag = -1
	}

	alarmVal = getAlarmValName(msgType, valType, alarmFlag)
	removalAlarmFlag = int(bitFlagInt) & 4
	alarmName = fixAlarmName[msgType]
	return
}

func parseBatteryVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
	valInt, err := strconv.ParseInt(val, 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return -1, -1, "", ""
	}

	alarmVal = strconv.Itoa(int(valInt) / 2)
	alarmFlag = -1
	alarmName = varAlarmName[valType]
	return
}

func parseVoltageVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
	valInt, err := strconv.ParseInt(val, 16, 64)
	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return -1, -1, "", ""
	}

	alarmVal = strconv.Itoa(int(valInt) / 10)
	alarmFlag = -1
	alarmName = varAlarmName[valType]
	return
}

func Little2BigEndianString(src string) (dst string) {
	if len(src)%2 != 0 || len(src) > 16 {
		return ""
	}

	v64, err := strconv.ParseUint(src, 16, 64)
	if err != nil {
		return ""
	}
	res := make([]byte, 8)
	binary.LittleEndian.PutUint64(res, v64)
	dst = hex.EncodeToString(res[:len(src)/2])
	return
}

func getAlarmValName(msgType MsgType, valType int, alarmFlag int) (res string) {
	sli, ok := alarmValueMapByTyp[msgType]
	res = strconv.Itoa(alarmFlag)
	if !ok {
		sli, ok = alarmValueMapByCid[valType]
		if !ok {
			return
		}
	}
	if alarmFlag < len(sli) && alarmFlag >= 0 {
		return sli[alarmFlag]
	} else {
		return
	}
}
