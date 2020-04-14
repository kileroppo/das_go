package feibee2srv

import (
	"das/core/entity"
	"das/core/log"
)

type MsgType int32

const (
	NewDev    MsgType = iota //设备入网
	DevOnline                //设备离上线
	DevDelete                //设备删除
	DevRename                //设备重命名
	GtwOnline                //网关离上线
	GtwInfo                  //网关信息

	//设备操作消息
	ManualOpDev //手动操作设备
	RemoteOpDev //远程操作设备
	//小卫士场景
	FeibeeScene
	//设备控制程度通知（窗帘开关程度）
	DevDegree
	//其他消息
	SpecialMsg
	//特殊设备
	InfraredTreasure //红外宝
	WonlyLGuard      //小卫士
	SceneSwitch      //情景开关
	ZigbeeLock       //zigbee锁
	Airer            //晾衣架

	SensorVol
	SensorBatt

	BaseSensor
	IlluminanceSensor
	TemperAndHumiditySensor
	InfraredSensor
	DoorMagneticSensor
	SmokeSensor
	FloodSensor
	GasSensor
	SosBtnSensor
	FloorHeat
)

var (
	feibeeMsgTyp = map[int]MsgType{
		2:  SpecialMsg,
		3:  NewDev,
		4:  DevOnline,
		5:  DevDelete,
		7:  RemoteOpDev,
		10: DevDegree,
		12: DevRename,
		15: GtwInfo,
		32: GtwOnline,
		21: FeibeeScene,
		22: FeibeeScene,
		23: FeibeeScene,
	}

	spDevMsgTyp = map[int]MsgType{
		//get key by feibee: deviceuid,zonetype
		0x030b0001: WonlyLGuard,      //小卫士
		0x01630001: InfraredTreasure, //红外宝
		0x00040001: SceneSwitch,
		0x00040002: SceneSwitch,
		0x00040003: SceneSwitch,
		0x00040004: SceneSwitch,      //情景开关
		0x02040001: Airer,            //晾衣架

		0x01060001: IlluminanceSensor,       //光照度传感器
		0x03020001: TemperAndHumiditySensor, //温湿度传感器
		0x0402000d: InfraredSensor,          //红外人体传感器
		0x04020015: DoorMagneticSensor,      //门磁传感器
		0x04020028: SmokeSensor,             //烟雾传感器
		0x0402002a: FloodSensor,             //水浸传感器
		0x0402002b: GasSensor,               //可燃气体传感器
		0x0402002c: SosBtnSensor,            //紧急按钮
		0x03010001: FloorHeat,
	}

	otherMsgTyp = map[int]MsgType{
		//get key by feibee: cid,aid
		0x00060000: ManualOpDev,
		0x05000080: BaseSensor, //传感器
		0x00010020: SensorVol,  //传感器低压
		0x00010021: SensorBatt, //传感器低电量
		0x00010035: SensorBatt,
		0x0001003e: SensorVol,
	}
)

func msgHandleFactory(data *entity.FeibeeData) (msgHandle MsgHandler) {
	typ := getMsgTyp(data)
	switch typ {
	case NewDev, DevOnline, DevRename, DevDelete, DevDegree, RemoteOpDev:
		msgHandle = &NormalMsgHandle{data: data, msgType: typ}
	case ManualOpDev:
		msgHandle = &ManualOpMsgHandle{data: data}
	case GtwOnline, GtwInfo:
		msgHandle = &GtwMsgHandle{data: data}
	case InfraredTreasure:
		msgHandle = &InfraredTreasureHandle{data: data, msgType: typ}
	case WonlyLGuard:
		msgHandle = &WonlyLGuardHandle{data: data, msgType: typ}
	case SceneSwitch:
		msgHandle = &SceneSwitchHandle{data: data}
	case ZigbeeLock:
		msgHandle = &ZigbeeLockHandle{data: data}
	case FeibeeScene:
		msgHandle = &FeibeeSceneHandle{data: data}
	case SensorVol, SensorBatt, BaseSensor, InfraredSensor, DoorMagneticSensor, GasSensor, FloodSensor, SosBtnSensor, SmokeSensor:
		msgHandle = &BaseSensorAlarm{feibeeMsg: data, msgType: typ}
	case TemperAndHumiditySensor, IlluminanceSensor, Airer, FloorHeat:
		msgHandle = &ContinuousSensor{BaseSensorAlarm{feibeeMsg: data, msgType: typ}}
	default:
		log.Warning("The FeibeeMsg type was not supported")
		msgHandle = nil
	}

	return
}

func getMsgTyp(data *entity.FeibeeData) (typ MsgType) {
	var ok bool
	typ, ok = feibeeMsgTyp[data.Code]
	if !ok {
		return -1
	}

	if typ == SpecialMsg {
		if data.Records[0].Snid == "FZD56-DOR07WL2.4" {
			return ZigbeeLock
		} else {
			typ, ok = spDevMsgTyp[getSpMsgKey(data.Records[0].Deviceid, data.Records[0].Zonetype)]
			if ok {
				return
			} else {
				typ, ok = otherMsgTyp[getSpMsgKey(data.Records[0].Cid, data.Records[0].Aid)]
				if ok {
					return
				} else {
					return -1
				}
			}
		}
	} else {
		return
	}
}

func getSpMsgKey(high, low int) int {
	high = high << 16
	return high + low
}
