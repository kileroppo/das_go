package feibee2srv

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"../core/entity"
	"../core/log"
)

var (
	SensorMsgTypeErr = errors.New("sensorAlarmMsg type error")
)

type DevAlarmer interface {
	GetMsg2app(int) ([][]byte, error)
	GetAlarmValue() string
}

func NewDevAlarm(feibeeData entity.FeibeeData, index int) (res DevAlarmer) {
	res = nil

	if feibeeData.Records[index].Deviceid > 0 {
		switch feibeeData.Records[index].Deviceid {
		//光照度传感器
		case 0x0106:
			res = &IlluminanceSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg:feibeeData,
				},
			}

			//温湿度传感器
		case 0x0302:
			res = &TemperAndHumiditySensorAlarm{
				BaseSensorAlarm{
					feibeeMsg:feibeeData,
				},
			}

		case 0x0402:
			switch feibeeData.Records[index].Zonetype {

			//人体红外传感器
			case 0x000d:
				res = &InfraredSensorAlarm{
					BaseSensorAlarm{
						feibeeMsg: feibeeData,
					},
				}

				//门磁传感器
			case 0x0015:
				res = &DoorMagneticSensorAlarm{
					BaseSensorAlarm{
						feibeeMsg: feibeeData,
					},
				}

				//烟雾传感器
			case 0x0028:

				//水浸传感器
			case 0x002A:
				res = &FloodSensorAlarm{
					BaseSensorAlarm{
						feibeeMsg: feibeeData,
					},
				}

				//可燃气体传感器
			case 0x002B:
				res = &GasSensorAlarm{
					BaseSensorAlarm{
						feibeeMsg: feibeeData,
					},
				}

				//一氧化碳传感器
			case 0x8001:

			}

		}
	} else {

		if (feibeeData.Records[0].Cid == 1024 && feibeeData.Records[0].Aid == 0) {
			//光照度
			res = &IlluminanceSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg:feibeeData,
				},
			}
		}

		if (feibeeData.Records[0].Cid == 1026 && feibeeData.Records[0].Aid == 0) || (feibeeData.Records[0].Cid == 1029 && feibeeData.Records[0].Aid == 0){
			//温湿度
			res = &TemperAndHumiditySensorAlarm{
				BaseSensorAlarm{
					feibeeMsg:feibeeData,
				},
			}

		}

		if (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 33) || (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 53) {
			//电量上报
			res = &BaseSensorAlarm{
				feibeeMsg:feibeeData,
			}

		}

		if (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 32) || (feibeeData.Records[0].Cid == 1 && feibeeData.Records[0].Aid == 62) {
			//电压上报
			res = &BaseSensorAlarm{
				feibeeMsg:feibeeData,
			}
		}
	}


	return
}

type BaseSensorAlarm struct {
	feibeeMsg        entity.FeibeeData
	alarmFlag        string
	alarmVal         string
	removalAlarmFlag string
}

func (b *BaseSensorAlarm) getMsg2app(index int) (msg entity.FeibeeAlarm2AppMsg, err error) {
	if index >= len(b.feibeeMsg.Records) || index < 0 {
		err = errors.New("the message is not alarm message")
		return
	}

	msg.Cmd = 0xfc
	msg.Ack = 0
	msg.DevType = devTypeConv(b.feibeeMsg.Records[index].Deviceid, b.feibeeMsg.Records[index].Zonetype)
	msg.Devid = b.feibeeMsg.Records[index].Uuid
	msg.Vendor = "feibee"
	msg.SeqId = 1
	msg.Time = int(time.Now().Unix())

	removalAlarmFlag, alarmFlag, alarmVal := alarmMsgParse(b.feibeeMsg.Records[index])
	if alarmFlag == "" && alarmVal == "" && removalAlarmFlag == "" {
		err = errors.New("alarmMsgParse() error")
		return
	}

	b.alarmVal = alarmVal
	b.alarmFlag = alarmFlag
	b.removalAlarmFlag = removalAlarmFlag

	return
}

func (self *BaseSensorAlarm) GetMsg2app(index int) ([][]byte, error) {

	alarmMsg,err := self.getMsg2app(index)
	if err != nil {
		log.Error("BaseSensorAlarm getMsg2app() error = ", err)
		return nil,err
	}

	if self.alarmVal == "" && self.alarmFlag == "" {
		log.Error("BaseSensorAlarm getMsg2app() error = ", err)
		return nil,SensorMsgTypeErr
	}

	alarmMsg.AlarmType = self.alarmFlag
	alarmMsg.AlarmValue = self.alarmVal

	alarmRawData,err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("BaseSensorAlarm GetMsg2app() error = ", err)
		return nil, err
	}

	return [][]byte{alarmRawData}, nil
}

func (self *BaseSensorAlarm) GetAlarmValue() string {
    return self.alarmVal
}

type InfraredSensorAlarm struct {
	BaseSensorAlarm
}

func (self *InfraredSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("InfraredSensorAlarm getMsg2app() error = ", err)
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

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("InfraredSensorAlarm GetMsg2app() error = ", err)
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
			log.Error("InfraredSensorAlarm GetMsg2app() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type DoorMagneticSensorAlarm struct {
	BaseSensorAlarm
}

func (self *DoorMagneticSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("DoorMagneticSensorAlarm getMsg2app() error = ", err)
		return nil, err
	}

	if self.alarmFlag == "1" {
		alarmMsg.AlarmValue = "开门"
		alarmMsg.AlarmType = "doorContact"
	} else if self.alarmFlag == "0" {
		alarmMsg.AlarmValue = "关门"
		alarmMsg.AlarmType = "doorContact"
	} else {
		alarmMsg.AlarmType = self.alarmFlag
		alarmMsg.AlarmValue = self.alarmVal
	}

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("DoorMagneticSensorAlarm GetMsg2app() error = ", err)
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
			log.Error("DoorMagneticSensorAlarm GetMsg2app() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type GasSensorAlarm struct {
	BaseSensorAlarm
}

func (self *GasSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg, err := self.getMsg2app(index)
	if err != nil {
		log.Error("GasSensorAlarm getMsg2app() error = ", err)
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

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("GasSensorAlarm GetMsg2app() error = ", err)
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
			log.Error("GasSensorAlarm GetMsg2app() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type FloodSensorAlarm struct {
	BaseSensorAlarm
}

func (self *FloodSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
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

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

	alarmRawData, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("FloodSensorAlarm GetMsg2app() error = ", err)
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
			log.Error("FloodSensorAlarm GetMsg2app() error = ", err)
		} else {
			res = append(res, removalRawData)
		}
	}

	return res, nil
}

type IlluminanceSensorAlarm struct {
	BaseSensorAlarm
}

func (self *IlluminanceSensorAlarm) GetMsg2app(index int) ([][]byte, error) {
    alarmMsg,_ := self.getMsg2app(index)

    alarmMsg.AlarmValue = self.getIlluminance()
    if alarmMsg.AlarmValue == "" {
    	err := errors.New("sensor get illuminance error")
    	return nil,err
	}

    alarmMsg.AlarmType = "illuminance"
    res, err := json.Marshal(alarmMsg)
    if err != nil {
    	log.Error("IlluminanceSensorAlarm GetMsg2app() error = ", err)
    	return nil,err
	}

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

    return [][]byte{res}, nil
}

func (self IlluminanceSensorAlarm) getIlluminance() string {
	if len(self.feibeeMsg.Records[0].Value) != 4 {
		return ""
	}

	value := self.feibeeMsg.Records[0].Value[2:4] + self.feibeeMsg.Records[0].Value[0:2]
	illuminance,err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return ""
	}
	return (strconv.Itoa(int(illuminance)) + "lux")
}

type TemperAndHumiditySensorAlarm struct {
	BaseSensorAlarm
}

func (self *TemperAndHumiditySensorAlarm) GetMsg2app(index int) ([][]byte, error) {
	alarmMsg,_ := self.getMsg2app(index)

	cid,aid :=  self.feibeeMsg.Records[0].Cid, self.feibeeMsg.Records[0].Aid

	if cid == 1026 && aid == 0 {
		alarmMsg.AlarmType = "temperature"
		alarmMsg.AlarmValue = self.getTemper()
		if alarmMsg.AlarmValue == "" {
			err := errors.New("sensor get temperature error")
			return nil,err
		}
	} else if cid == 1029 && aid == 0 {
		alarmMsg.AlarmType = "humidity"
		alarmMsg.AlarmValue = self.getHumidity()
		if alarmMsg.AlarmValue == "" {
			err := errors.New("sensor get humidity error")
			return nil,err
		}
	} else {
		return nil, SensorMsgTypeErr
	}

	self.alarmVal = alarmMsg.AlarmValue
	self.alarmFlag = alarmMsg.AlarmType

	res, err := json.Marshal(alarmMsg)
	if err != nil {
		log.Error("TemperAndHumiditySensorAlarm GetMsg2app() error = ", err)
		return nil,err
	}

	return [][]byte{res}, nil
}

func (self TemperAndHumiditySensorAlarm) getTemper() string {
	if len(self.feibeeMsg.Records[0].Value) != 4 {
		return ""
	}

	value := self.feibeeMsg.Records[0].Value[2:4] + self.feibeeMsg.Records[0].Value[0:2]
	temper,err := strconv.ParseInt(value, 16, 64)
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
	humidity,err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(float64(humidity)/100, 'f', 2, 64) + "%"
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
