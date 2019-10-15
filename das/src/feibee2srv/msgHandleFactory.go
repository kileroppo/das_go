package feibee2srv

import (
	"../core/entity"
)

type msgType int32

const (
	NewDev    msgType = iota + 1 //设备入网
	DevOnline                    //设备离上线
	DevDelete                    //设备删除
	DevRename                    //设备重命名
	GtwOnline                    //网关离上线

	//设备操作消息
	ManualOpDev //手动操作设备
	RemoteOpDev //远程操作设备

	//特殊设备
	SensorAlarm         //传感器报警
	InfraredTreasure //红外宝
)

type MsgHandler interface {
	PushMsg()
}

func MsgHandleFactory(data entity.FeibeeData) (msgHandle MsgHandler) {
	typ := getMsgType(data)
	switch typ {

	case NewDev, DevOnline, DevRename, DevDelete:

	case GtwOnline:

	case ManualOpDev, RemoteOpDev:

	case SensorAlarm:

	case InfraredTreasure:

	default:
		return nil
	}

	return
}

func getMsgType(data entity.FeibeeData) (typ msgType) {
	typ = -1
	switch data.Code {
	case 3:
		typ = NewDev
	case 4:
		typ = DevOnline
	case 5:
		typ = DevDelete
	case 7:
		typ = RemoteOpDev
	case 12:
		typ = DevRename
	case 32:
		typ = GtwOnline
	case 2:
		if data.Records[0].Aid == 0 && data.Records[0].Cid == 6 {
			typ = ManualOpDev
		}

		if data.Records[0].Cid == 1280 && data.Records[0].Aid == 128 {
			typ = SensorAlarm
		}

		if data.Records[0].Deviceid == 355 && data.Records[0].Zonetype == 255 {
			typ = InfraredTreasure
		}
	}
	return
}

//设备入网 设备离上线 设备删除 设备重命名
type NormalMsgHandle struct {

}

func (self NormalMsgHandle) PushMsg() {


}

func (self NormalMsgHandle) pushOneMsg(e) {

}
