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

func (self *BaseSensorAlarm) initData() {
	self.devType = devTypeConv(self.feibeeMsg.Records[0].Deviceid, self.feibeeMsg.Records[0].Zonetype)
	self.devid = self.feibeeMsg.Records[0].Uuid
	self.time = int(time.Now().Unix())
	self.bindid = self.feibeeMsg.Records[0].Bindid
	self.alarmMsgType = getSpMsgKey(self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid)
}

func (self *BaseSensorAlarm) parseAlarmMsg() error {
	parse := parseFunc(nil)
	ok := false

	parse, ok = alarmMsgTyp[self.alarmMsgType]
	if !ok {
		return ErrAlarmMsg
	}

	self.removalFlag, self.alarmFlag, self.alarmVal, self.alarmType = parse(self.feibeeMsg.Records[0].Value, self.msgType, self.alarmMsgType)

	if self.alarmFlag < 0 {
		return ErrAlarmMsg
	}

	return nil
}

func (self *BaseSensorAlarm) PushMsg() {
	self.initData()
	if err := self.parseAlarmMsg(); err != nil {
		//log.Warning("BaseSensorAlarm PushMsg > %s", err)
		return
	}
	//传感器正常消息不通知不存储 门磁除外
	if self.alarmType == "doorContact" {
		self.pushMsg2mns()
	} else if self.alarmType == "lowPower" {
		self.pushStatusMsg2app("power")
		if (self.alarmFlag < 30) {
			self.pushMsg2mns()
		}
	} else {
		if (self.alarmFlag > 0) {
			self.pushMsg2mns()
		}
	}
    self.pushMsg2pmsForSave()
	if self.alarmMsgType == sensorAlarm {
		self.pushMsg2pmsForSceneTrigger()
	}
	self.pushForcedBreakMsg()
}

func (self *BaseSensorAlarm) pushStatusMsg2app(opType string) {
	msg := self.createStatusMsg(opType)

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm.pushStatusMsg2app > %s", err)
		return
	}
	rabbitmq.Publish2app(data, self.devid)
}

func (self *BaseSensorAlarm) pushAlarmMsg2app() {
	msg := self.createAlarmMsg()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm.pushAlarmMsg2app > %s", err)
		return
	}
	rabbitmq.Publish2app(data, self.devid)
}

func (self *BaseSensorAlarm) pushMsg2mns() {
	msg := self.createAlarmMsg()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm.pushMsg2mns > %s", err)
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

	msg.AlarmFlag = self.alarmFlag
	msg.AlarmType = self.alarmType

	return msg
}

func (self *BaseSensorAlarm) createStatusMsg(opType string) entity.Feibee2DevMsg {
	var msg entity.Feibee2DevMsg

	msg.Cmd = 0xfb
	msg.Ack = 0
	msg.DevType = self.devType
	msg.DevId = self.devid
	msg.Vendor = "feibee"
	msg.SeqId = 1

	msg.OpType = opType
	msg.OpValue = strconv.Itoa(self.alarmFlag)
	msg.Time = self.time

	return msg
}

func (self *BaseSensorAlarm) createAlarmMsg() entity.Feibee2AlarmMsg {
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

func (self *BaseSensorAlarm) pushMsg2pmsForSave() {
	msg := self.createAlarmMsg()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm.pushMsg2pmsForSave > %s", err)
		return
	}
	rabbitmq.Publish2pms(data, "")
}

func (self *BaseSensorAlarm) pushMsg2pmsForSceneTrigger() {
	msg := self.createMsg2pmsForSence()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm.pushMsg2pmsForSceneTrigger > %s", err)
		return
	}
	rabbitmq.Publish2pms(data, "")
}

func (self *BaseSensorAlarm) pushForcedBreakMsg() {
	if self.removalFlag > 0 {
		msg := self.createAlarmMsg()

		msg.AlarmType = "forcedBreak"
		msg.AlarmValue = "传感器被强拆"
		msg.AlarmFlag = 1

		data, err := json.Marshal(msg)
		if err != nil {
			log.Error("BaseSensorAlarm.pushForcedBreakMsg > %s", err)
			return
		}
		rabbitmq.Publish2mns(data, "")
		rabbitmq.Publish2pms(data, "")
	}
}

type ContinuousSensor struct {
	BaseSensorAlarm
}

func (c *ContinuousSensor) PushMsg() {
	c.initData()
	if err := c.parseAlarmMsg(); err != nil {
		log.Warning("ContinuousSensor.PushMsg > %s", err)
		return
	}
	//todo: 其他类型暂不推送mns
	//c.pushMsg2mns()
	if c.msgType == FloorHeat || c.msgType == Airer {
		c.pushStatusMsg2app(c.alarmType)
	}

	if c.msgType == TemperAndHumiditySensor || c.msgType == IlluminanceSensor || c.msgType == Airer {
		c.pushMsg2pmsForSave()

	}

	if c.msgType == TemperAndHumiditySensor || c.msgType == IlluminanceSensor {
		c.pushMsg2pmsForSceneTrigger()
	}

	//if c.msgType == FloorHeat {
	//	c.pushStatusMsg2app()
	//} else {
	//	c.pushMsg2pmsForSave()
	//	if c.msgType != Airer {
	//		c.pushMsg2pmsForSceneTrigger()
	//	}
	//}
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
	alarmVal = strconv.FormatUint(v64, 10)
	//alarmVal = getAlarmValName(msgType, valType, alarmFlag)
	removalAlarmFlag = -1
	alarmName = varAlarmName[valType]
	return
}

func parseFixVal(val string, msgType MsgType, valType int) (removalAlarmFlag, alarmFlag int, alarmVal, alarmName string) {
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

	alarmVal = getAlarmValName(msgType, valType, alarmFlag)
	if alarmVal == "" {
		alarmFlag = -1
		return
	} else {
		alarmFlag = int(v64)
		alarmName = varAlarmName[valType]
		return
	}
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

	//todo:周期上报数据不透传（暂定）
	if cycleFlag := (bitFlagInt & 0b0001_0000); cycleFlag > 0 {
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

	alarmVal = "电量过低"
	alarmFlag = int(valInt) / 2
	//if int(valInt)/2 <= 30 {
	//
	//} else {
	//	alarmFlag = -1
	//}
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
