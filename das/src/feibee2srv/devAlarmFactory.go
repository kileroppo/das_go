package feibee2srv

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"../core/entity"
	"../core/log"
)

type DevAlarmer interface {
	GetMsg2app(int) ([][]byte, error)
}

func NewDevAlarm(feibeeData FeibeeData, index int) (res DevAlarmer) {
	res = nil

	switch feibeeData.Records[index].Deviceid {
	//光照度传感器
	case 0x0106:

	//温湿度传感器
	case 0x0302:

	case 0x0402:
		switch feibeeData.Records[index].Zonetype {

		//人体红外传感器
		case 0x000d:
			res = InfraredSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}

		//门磁传感器
		case 0x0015:
			res = DoorMagneticSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}

		//烟雾传感器
		case 0x0028:

		//水浸传感器
		case 0x002A:
			res = FloodSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}

		//可燃气体传感器
		case 0x002B:
			res = GasSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: feibeeData,
				},
			}

		//一氧化碳传感器
		case 0x8001:

		}

	}

	return
}

type BaseSensorAlarm struct {
	feibeeMsg        FeibeeData
	alarmFlag        string
	alarmVal         string
	removalAlarmFlag string
}

func (b *BaseSensorAlarm) getMsg2app(index int) (msg entity.FeibeeAlarm2AppMsg, err error) {
	if index >= len(b.feibeeMsg.Records) || index < 0 {
		err = errors.New("the message is not alarm message")
		return
	}

	removalAlarmFlag, alarmFlag, alarmVal := alarmMsgParse(b.feibeeMsg.Records[index])
	if alarmFlag == "" && alarmVal == "" && removalAlarmFlag == "" {
		err = errors.New("alarmMsgParse() error")
		return
	}

	msg.Cmd = 0xfc
	msg.Ack = 0
	msg.DevType = b.feibeeMsg.Records[index].Devicetype
	msg.Devid = b.feibeeMsg.Records[index].Uuid
	msg.Vendor = "feibee"
	msg.SeqId = 1
	msg.Time = int(time.Now().Unix())

	b.alarmVal = alarmVal
	b.alarmFlag = alarmFlag
	b.removalAlarmFlag = removalAlarmFlag

	return
}

type InfraredSensorAlarm struct {
	BaseSensorAlarm
}

func (self InfraredSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("getMsg2app() error = ", err)
		return nil, err
	}

	if self.alarmFlag == "1" {
		alarmMsg.AlarmValue = "有人"
		alarmMsg.AlarmType = "infrared"
	} else if self.alarmFlag == "0" {
		alarmMsg.AlarmValue = "无人"
		alarmMsg.AlarmType = "infrared"
	} else {
		alarmMsg.AlarmType = self.alarmFlag
		alarmMsg.AlarmValue = self.alarmVal
	}

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return nil, err
	}

	res := [][]byte{}
	res = append(res, alarmRawData)

	if self.removalAlarmFlag == "1" {
		removalAlarmMsg := alarmMsg
		removalAlarmMsg.AlarmType = "forcedBreak"
		removalAlarmMsg.AlarmValue = "传感器被强拆"

		removalRawData, err := json.Marshal(removalAlarmMsg)
		if err != nil {
			log.Error("json.Marshal() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type DoorMagneticSensorAlarm struct {
	BaseSensorAlarm
}

func (self DoorMagneticSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("getMsg2app() error = ", err)
		return nil, err
	}

	if self.alarmFlag == "1" {
		alarmMsg.AlarmValue = "开门"
		alarmMsg.AlarmType = "doorMagnet"
	} else if self.alarmFlag == "0" {
		alarmMsg.AlarmValue = "关门"
		alarmMsg.AlarmType = "doorMagnet"
	} else {
		alarmMsg.AlarmType = self.alarmFlag
		alarmMsg.AlarmValue = self.alarmVal
	}

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return nil, err
	}

	res := [][]byte{}
	res = append(res, alarmRawData)

	if self.removalAlarmFlag == "1" {
		removalAlarmMsg := alarmMsg
		removalAlarmMsg.AlarmType = "forcedBreak"
		removalAlarmMsg.AlarmValue = "传感器被强拆"

		removalRawData, err := json.Marshal(removalAlarmMsg)
		if err != nil {
			log.Error("json.Marshal() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type GasSensorAlarm struct {
	BaseSensorAlarm
}

func (self GasSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("getMsg2app() error = ", err)
		return nil, err
	}

	if self.alarmFlag == "1" {
		alarmMsg.AlarmValue = "有气体"
		alarmMsg.AlarmType = "gas"
	} else if self.alarmFlag == "0" {
		alarmMsg.AlarmValue = "无气体"
		alarmMsg.AlarmType = "gas"
	} else {
		alarmMsg.AlarmType = self.alarmFlag
		alarmMsg.AlarmValue = self.alarmVal
	}

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return nil, err
	}

	res := [][]byte{}
	res = append(res, alarmRawData)

	if self.removalAlarmFlag == "1" {
		removalAlarmMsg := alarmMsg
		removalAlarmMsg.AlarmType = "forcedBreak"
		removalAlarmMsg.AlarmValue = "传感器被强拆"

		removalRawData, err := json.Marshal(removalAlarmMsg)
		if err != nil {
			log.Error("json.Marshal() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type FloodSensorAlarm struct {
	BaseSensorAlarm
}

func (self FloodSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("getMsg2app() error = ", err)
		return nil, err
	}

	if self.alarmFlag == "1" {
		alarmMsg.AlarmValue = "有水"
		alarmMsg.AlarmType = "flood"
	} else if self.alarmFlag == "0" {
		alarmMsg.AlarmValue = "无水"
		alarmMsg.AlarmType = "flood"
	} else {
		alarmMsg.AlarmType = self.alarmFlag
		alarmMsg.AlarmValue = self.alarmVal
	}

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return nil, err
	}

	res := [][]byte{}
	res = append(res, alarmRawData)

	if self.removalAlarmFlag == "1" {
		removalAlarmMsg := alarmMsg
		removalAlarmMsg.AlarmType = "forcedBreak"
		removalAlarmMsg.AlarmValue = "传感器被强拆"

		removalRawData, err := json.Marshal(removalAlarmMsg)
		if err != nil {
			log.Error("json.Marshal() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
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

	res := strconv.Itoa(int(valInt)/2) + "%"
	return res
}

func voltageValueParse(val string) string {
	valInt, err := strconv.ParseInt(val, 16, 64)

	if err != nil {
		log.Error("strconv.ParseInt() error = ", err)
		return ""
	}

	res := strconv.Itoa(int(valInt)/10) + "V"
	return res
}
