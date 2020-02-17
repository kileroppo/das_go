package feibee2srv

import (
	"strconv"
	"time"

	"github.com/json-iterator/go"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type BaseSensorAlarm struct {
	feibeeMsg         *entity.FeibeeData
	alarmType         string
	alarmVal          string
	removalAlarmValue string

	devType string
	devid   string
	time    int

	alarmFlag int
	bindid    string
}

func (b *BaseSensorAlarm) parseAlarmMsg() {
	b.devType = devTypeConv(b.feibeeMsg.Records[0].Deviceid, b.feibeeMsg.Records[0].Zonetype)
	b.devid = b.feibeeMsg.Records[0].Uuid
	b.time = int(time.Now().Unix())
	b.bindid = b.feibeeMsg.Records[0].Bindid

	removalAlarmFlag, alarmType, alarmVal := alarmMsgParse(b.feibeeMsg.Records[0])

	b.alarmVal = alarmVal
	b.alarmType = alarmType
	b.removalAlarmValue = removalAlarmFlag

	return
}

func (self *BaseSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	self.pushMsg2mns()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
}

func (self *BaseSensorAlarm) pushMsg2app() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2app() error = ", err)
		return
	}

	//producer.SendMQMsg2APP(self.bindid, string(data))
	rabbitmq.Publish2app(data, self.devid)
}

func (self *BaseSensorAlarm) pushMsg2mns() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2mns() error = ", err)
		return
	}
	//producer.SendMQMsg2Db(string(data))
	if len(self.alarmType) > 0 && len(self.alarmVal) > 0 {
		rabbitmq.Publish2mns(data, "")
	}

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
	//producer.SendMQMsg2PMS(string(data))
	if len(self.alarmVal) >0 && len(self.alarmType) > 0 {
		rabbitmq.Publish2pms(data, "")
	}
}

func (self *BaseSensorAlarm) pushMsg2pmsForSceneTrigger() {
	msg := self.createMsg2pmsForSence()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2pmsForSceneTrigger() error = ", err)
		return
	}
	//producer.SendMQMsg2PMS(string(data))
	if len(self.alarmVal) >0 && len(self.alarmType) > 0 {
		rabbitmq.Publish2pms(data, "")
	}
}

func (self *BaseSensorAlarm) pushForcedBreakMsg() {
	if self.removalAlarmValue == "1" {
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

type InfraredSensorAlarm struct {
	BaseSensorAlarm
}

func (self *InfraredSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "有人"
		self.alarmType = "infrared"
		self.alarmFlag = 1
		self.pushMsg2mns()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无人"
		self.alarmType = "infrared"
		self.alarmFlag = 0
	} else {
		self.pushMsg2mns()
	}

	self.pushMsg2pmsForSave()
    self.pushForcedBreakMsg()
}

type DoorMagneticSensorAlarm struct {
	BaseSensorAlarm
}

func (self *DoorMagneticSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "开启"
		self.alarmType = "doorContact"
		self.alarmFlag = 1
	} else if self.alarmType == "0" {
		self.alarmVal = "关闭"
		self.alarmType = "doorContact"
		self.alarmFlag = 0
	}
	self.pushMsg2mns()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
	self.pushForcedBreakMsg()
}

type GasSensorAlarm struct {
	BaseSensorAlarm
}

func (self *GasSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "有气体"
		self.alarmType = "gas"
		self.alarmFlag = 1
		self.pushMsg2mns()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无气体"
		self.alarmType = "gas"
		self.alarmFlag = 0
	} else {
		self.pushMsg2mns()
	}
	self.pushMsg2pmsForSave()
	self.pushForcedBreakMsg()
}

type FloodSensorAlarm struct {
	BaseSensorAlarm
}

func (self *FloodSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "有水"
		self.alarmType = "flood"
		self.alarmFlag = 1
		self.pushMsg2mns()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无水"
		self.alarmType = "flood"
		self.alarmFlag = 0
	} else {
		self.pushMsg2mns()
	}

	self.pushMsg2pmsForSave()
	self.pushForcedBreakMsg()
}

type SmokeSensorAlarm struct {
	BaseSensorAlarm
}

func (self *SmokeSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "有烟"
		self.alarmType = "smoke"
		self.alarmFlag = 1
		self.pushMsg2mns()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无烟"
		self.alarmType = "smoke"
		self.alarmFlag = 0
	} else {
		self.pushMsg2mns()
	}

	self.pushMsg2pmsForSave()
	self.pushForcedBreakMsg()
}

type IlluminanceSensorAlarm struct {
	BaseSensorAlarm
}

func (self *IlluminanceSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	self.alarmVal = self.getIlluminance()
	if len(self.alarmVal) <= 0 {
		log.Warning("IlluminanceSensorAlarm getIlluminance() error")
		return
	}
	self.alarmType = "illuminance"
	self.pushMsg2mns()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
	self.pushForcedBreakMsg()
}

func (self *IlluminanceSensorAlarm) getIlluminance() string {
	if len(self.feibeeMsg.Records[0].Value) != 4 {
		return ""
	}

	value := self.feibeeMsg.Records[0].Value[2:4] + self.feibeeMsg.Records[0].Value[0:2]
	illuminance, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return ""
	}
	return (strconv.Itoa(int(illuminance)))
}

type TemperAndHumiditySensorAlarm struct {
	BaseSensorAlarm
}

func (self *TemperAndHumiditySensorAlarm) PushMsg() {
	self.parseAlarmMsg()

	cid, aid := self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid

	if cid == 1026 && aid == 0 {
		self.alarmType = "temperature"
		self.alarmVal = self.getTemper()
		if len(self.alarmVal) <= 0 {
			log.Warning("TemperAndHumiditySensorAlarm getTemper() error")
			return
		}
	} else if cid == 1029 && aid == 0 {
		self.alarmType = "humidity"
		self.alarmVal = self.getHumidity()
		if len(self.alarmVal) <= 0 {
			log.Warning("TemperAndHumiditySensorAlarm getHumidity() error")
			return
		}
	} else {
		return
	}

	self.pushMsg2mns()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
	self.pushForcedBreakMsg()
}

func (self *TemperAndHumiditySensorAlarm) getTemper() string {
	if len(self.feibeeMsg.Records[0].Value) != 4 {
		return ""
	}

	value := self.feibeeMsg.Records[0].Value[2:4] + self.feibeeMsg.Records[0].Value[0:2]
	temper, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(float64(temper)/100, 'f', 2, 64)
}

func (self *TemperAndHumiditySensorAlarm) getHumidity() string {
	if len(self.feibeeMsg.Records[0].Value) != 4 {
		return ""
	}

	value := self.feibeeMsg.Records[0].Value[2:4] + self.feibeeMsg.Records[0].Value[0:2]
	humidity, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(float64(humidity)/100, 'f', 2, 64)
}

type ZigbeeAlarm struct {
	BaseSensorAlarm
}

func (self *ZigbeeAlarm) PushMsg() {

}

func alarmMsgParse(msg entity.FeibeeRecordsMsg) (removalAlarmFlag, alarmFlag, alarmValue string) {

	if msg.Cid == 1280 && msg.Aid == 128 {
		bitFlagInt, err := strconv.ParseInt(msg.Value[0:2], 16, 64)
		if err != nil {
			log.Error("strconv.ParseInt() error = ", err)
			return
		}

		if int(bitFlagInt)&1 > 0 {
			alarmFlag = "1"
		} else {
			alarmFlag = "0"
		}

		if int(bitFlagInt)&4 > 0 {
			removalAlarmFlag = "1"
		} else {
			removalAlarmFlag = "0"
		}
		return
	}

	if (msg.Cid == 1 && msg.Aid == 33) || (msg.Cid == 1 && msg.Aid == 53) {

		alarmFlag = "lowPower"
		if msg.Aid == 33 {
			alarmValue = batteryValueParse(msg.Value)
		} else {
			alarmValue = ""
		}
		return
	}

	if (msg.Cid == 1 && msg.Aid == 32) || (msg.Cid == 1 && msg.Aid == 62) {
		alarmFlag = "lowVoltage"
		if msg.Aid == 32 {
			alarmValue = voltageValueParse(msg.Value)
		} else {
			alarmValue = ""
		}
	}

	return
}

func batteryValueParse(val string) string {
	valInt, err := strconv.ParseInt(val, 16, 64)

	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return ""
	}

	res := strconv.Itoa(int(valInt) / 2)
	return res
}

func voltageValueParse(val string) string {
	valInt, err := strconv.ParseInt(val, 16, 64)

	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return ""
	}

	res := strconv.Itoa(int(valInt) / 10)
	return res
}
