package feibee2srv

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"../core/entity"
	"../core/log"
	"../rmq/producer"
)

var (
	SensorMsgTypeErr = errors.New("sensorAlarmMsg type error")
)

type DevAlarmer interface {
	PushMsg()
}

func DevAlarmFactory(feibeeData entity.FeibeeData) (res DevAlarmer) {
	res = nil

	if len(feibeeData.Records) <= 0 {
		return
	}

	switch feibeeData.Records[0].Deviceid {
	case 0x0106:
		//光照度传感器
		res = &IlluminanceSensorAlarm{
			BaseSensorAlarm{
				feibeeMsg: feibeeData,
			},
		}
	case 0x0302:
		//温湿度传感器
		res = &TemperAndHumiditySensorAlarm{
			BaseSensorAlarm{
				feibeeMsg: feibeeData,
			},
		}
	case 0x0402:
		//飞比传感器
		switch feibeeData.Records[0].Zonetype {
		case 0x000d:
			//人体红外传感器
			res = &InfraredSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}
		case 0x0015:
			//门磁传感器
			res = &DoorMagneticSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}
		case 0x0028:
			//烟雾传感器
			res = &SmokeSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}
		case 0x002A:
			//水浸传感器
			res = &FloodSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}
		case 0x002B:
			//可燃气体传感器
			res = &GasSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}
		case 0x8001:
			//一氧化碳传感器
		}
	default:
		if (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 33) || (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 53) {
			//电量上报
			res = &BaseSensorAlarm{
				feibeeMsg: feibeeData,
			}

		}

		if (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 32) || (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 62) {
			//电压上报
			res = &BaseSensorAlarm{
				feibeeMsg: feibeeData,
			}
		}
	}
	return
}

type BaseSensorAlarm struct {
	feibeeMsg         entity.FeibeeData
	alarmType         string
	alarmVal          string
	removalAlarmValue string

	devType string
	devid   string
	time    int

	alarmFlag int
	bindid    string
}

func (b *BaseSensorAlarm) parseAlarmMsg() (err error) {
	b.devType = devTypeConv(b.feibeeMsg.Records[0].Deviceid, b.feibeeMsg.Records[0].Zonetype)
	b.devid = b.feibeeMsg.Records[0].Uuid
	b.time = int(time.Now().Unix())
	b.bindid = b.feibeeMsg.Records[0].Bindid

	removalAlarmFlag, alarmType, alarmVal := alarmMsgParse(b.feibeeMsg.Records[0])
	if alarmType == "" && alarmVal == "" && removalAlarmFlag == "" {
		err = errors.New("alarmMsgParse() error")
		return
	}

	b.alarmVal = alarmVal
	b.alarmType = alarmType
	b.removalAlarmValue = removalAlarmFlag

	return
}

func (self *BaseSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	self.pushMsg2db()
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

	producer.SendMQMsg2APP(self.bindid, string(data))
}

func (self *BaseSensorAlarm) pushMsg2db() {
	msg := self.createMsg2app()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2db() error = ", err)
		return
	}
	producer.SendMQMsg2Db(string(data))

	if self.removalAlarmValue == "1" {
		msg.AlarmType = "forcedBreak"
		msg.AlarmValue = "传感器被强拆"
		msg.AlarmFlag = 1

		data, err = json.Marshal(msg)
		if err != nil {
			log.Error("BaseSensorAlarm pushMsg2db() error = ", err)
			return
		}
		producer.SendMQMsg2Db(string(data))
	}

}

func (self *BaseSensorAlarm) createMsg2app() entity.FeibeeAlarm2AppMsg {
	var msg entity.FeibeeAlarm2AppMsg

	msg.Cmd = 0xfc
	msg.Ack = 0
	msg.DevType = self.devType
	msg.Devid = self.devid
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
	producer.SendMQMsg2PMS(string(data))

	if self.removalAlarmValue == "1" {
		msg.AlarmType = "forcedBreak"
		msg.AlarmValue = "传感器被强拆"

		data, err = json.Marshal(msg)
		if err != nil {
			log.Error("BaseSensorAlarm pushMsg2db() error = ", err)
			return
		}
		producer.SendMQMsg2PMS(string(data))
	}

}

func (self *BaseSensorAlarm) pushMsg2pmsForSceneTrigger() {
	var msg entity.FeibeeAutoScene2pmsMsg

	msg.Cmd = 0xf1
	msg.Ack = 0
	msg.Vendor = "feibee"
	msg.SeqId = 1

	msg.DevType = self.devType
	msg.Devid = self.devid

	msg.TriggerType = 0
	msg.Time = self.time

	msg.AlarmValue = self.alarmVal
	msg.AlarmType = self.alarmType

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("BaseSensorAlarm pushMsg2pmsForSceneTrigger() error = ", err)
		return
	}
	producer.SendMQMsg2PMS(string(data))

	if self.removalAlarmValue == "1" {
		msg.AlarmType = "forcedBreak"
		msg.AlarmValue = "传感器被强拆"

		data, err = json.Marshal(msg)
		if err != nil {
			log.Error("BaseSensorAlarm pushMsg2db() error = ", err)
			return
		}
		producer.SendMQMsg2PMS(string(data))
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
		self.pushMsg2db()
		self.pushMsg2pmsForSave()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无人"
		self.alarmType = "infrared"
		self.alarmFlag = 0
		//self.pushMsg2pmsForSave()
	}

}

type DoorMagneticSensorAlarm struct {
	BaseSensorAlarm
}

func (self *DoorMagneticSensorAlarm) PushMsg() {
	self.parseAlarmMsg()
	if self.alarmType == "1" {
		self.alarmVal = "开门"
		self.alarmType = "doorContact"
		self.alarmFlag = 1
	} else if self.alarmType == "0" {
		self.alarmVal = "关门"
		self.alarmType = "doorContact"
		self.alarmFlag = 0
	}
	self.pushMsg2db()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
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
		self.pushMsg2db()
		self.pushMsg2pmsForSave()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无气体"
		self.alarmType = "gas"
		self.alarmFlag = 0
	}
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
		self.pushMsg2db()
		self.pushMsg2pmsForSave()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无水"
		self.alarmType = "flood"
		self.alarmFlag = 0
	}
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
		self.pushMsg2db()
		self.pushMsg2pmsForSave()
		self.pushMsg2pmsForSceneTrigger()
	} else if self.alarmType == "0" {
		self.alarmVal = "无烟"
		self.alarmType = "smoke"
		self.alarmFlag = 0
	}
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
	self.pushMsg2db()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
}

func (self IlluminanceSensorAlarm) getIlluminance() string {
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

	self.pushMsg2db()
	self.pushMsg2pmsForSave()
	self.pushMsg2pmsForSceneTrigger()
}

func (self TemperAndHumiditySensorAlarm) getTemper() string {
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

func (self TemperAndHumiditySensorAlarm) getHumidity() string {
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
			alarmValue = batteryValueParse(msg.Orgdata[30:32])
		}
		return
	}

	if (msg.Cid == 1 && msg.Aid == 32) || (msg.Cid == 1 && msg.Aid == 62) {
		alarmFlag = "lowVoltage"
		if msg.Aid == 32 {
			alarmValue = voltageValueParse(msg.Value)
		} else {
			alarmValue = voltageValueParse(msg.Orgdata[22:24])
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
