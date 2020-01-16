package feibee2srv

import (
	"das/core/entity"
	"das/core/log"
)

func MsgHandleFactory(data *entity.FeibeeData) (msgHandle MsgHandler) {
	typ := getMsgType(data)
	switch typ {

	case NewDev, DevOnline, DevRename, DevDelete, ManualOpDev, DevDegree:
		msgHandle = &NormalMsgHandle{
			data:    data,
			msgType: typ,
		}

	case GtwOnline, RemoteOpDev:
		msgHandle = &GtwMsgHandle{
			data:    data,
			msgType: typ,
		}

	case SensorAlarm:
		msgHandle = &SensorMsgHandle{
			data: data,
		}

	case InfraredTreasure:
		msgHandle = &InfraredTreasureHandle{
			data:    data,
			msgType: typ,
		}

	case WonlyLGuard:
		msgHandle = &WonlyLGuardHandle{
			data:    data,
			msgType: typ,
		}

	case SceneSwitch:
		msgHandle = &SceneSwitchHandle{
			data: data,
		}
	case ZigbeeLock:
		msgHandle = &ZigbeeLockHandle{
			data:data,
		}

	default:
		msgHandle = nil
	}

	return
}

func getMsgType(data *entity.FeibeeData) (typ MsgType) {
	defer func() {
		if err := recover(); err != nil {
			log.Warning(ErrMsgStruct)
		}
	}()

	typ = -1
	switch data.Code {
	case 3:
		if data.Msg[0].Deviceid == 779 {
			typ = WonlyLGuard
		} else {
			typ = NewDev
		}
	case 4:
		typ = DevOnline
	case 5:
		typ = DevDelete
	case 7:
		typ = RemoteOpDev
	case 10:
		if data.Msg[0].Deviceid == 0x0202 {
			typ = DevDegree
		}
	case 12:
		typ = DevRename
	case 32:
		typ = GtwOnline
	case 2:
		if data.Records[0].Deviceid == 779 {
			//小卫士
			typ = WonlyLGuard
		} else if data.Records[0].Snid == "FTB56-AVA05JD1.4" {
			//zigbee锁
			typ = ZigbeeLock
		} else if data.Records[0].Aid == 0 && data.Records[0].Cid == 6 {
			typ = ManualOpDev
		} else if isDevAlarm(data) {
			typ = SensorAlarm
		} else if data.Records[0].Cid == 0 && data.Records[0].Aid == 16394 {
			//红外宝
			typ = InfraredTreasure
		} else if data.Records[0].Cid == 61680 && data.Records[0].Aid == 61680 {
			//情景开关触发
			typ = SceneSwitch
		}
	}
	return
}

func DevAlarmFactory(feibeeData *entity.FeibeeData) (res MsgHandler) {
	res = nil

	if len(feibeeData.Records) <= 0 {
		return
	}

	switch feibeeData.Records[0].Deviceid {
	case 0xa:
		//zigbee锁

	case 0x0106:
		//光照度传感器
		res = &IlluminanceSensorAlarm{
			BaseSensorAlarm{
				feibeeMsg: *feibeeData,
			},
		}
	case 0x0302:
		//温湿度传感器
		res = &TemperAndHumiditySensorAlarm{
			BaseSensorAlarm{
				feibeeMsg: *feibeeData,
			},
		}
	case 0x0402:
		//飞比传感器
		switch feibeeData.Records[0].Zonetype {
		case 0x000d:
			//人体红外传感器
			res = &InfraredSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: *feibeeData,
				},
			}
		case 0x0015:
			//门磁传感器
			res = &DoorMagneticSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: *feibeeData,
				},
			}
		case 0x0028:
			//烟雾传感器
			res = &SmokeSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: *feibeeData,
				},
			}
		case 0x002A:
			//水浸传感器
			res = &FloodSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: *feibeeData,
				},
			}
		case 0x002B:
			//可燃气体传感器
			res = &GasSensorAlarm{
				BaseSensorAlarm{
					feibeeMsg: *feibeeData,
				},
			}
		default:
			res = &BaseSensorAlarm{
				feibeeMsg: *feibeeData,
			}
		}
	default:
		res = &BaseSensorAlarm{
			feibeeMsg: *feibeeData,
		}
	}
	return
}
