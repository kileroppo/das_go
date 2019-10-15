package feibee2srv

import (
	"strconv"
	"encoding/json"

	"../core/entity"
	"../core/log"
	"../rmq/producer"
)

type MsgType int32

const (
	NewDev    MsgType = iota + 1 //设备入网
	DevOnline                    //设备离上线
	DevDelete                    //设备删除
	DevRename                    //设备重命名
	GtwOnline                    //网关离上线

	//设备操作消息
	ManualOpDev //手动操作设备
	RemoteOpDev //远程操作设备

	//特殊设备
	SensorAlarm      //传感器报警
	InfraredTreasure //红外宝
)

type MsgHandler interface {
	PushMsg()
}

func MsgHandleFactory(data entity.FeibeeData) (msgHandle MsgHandler) {
	typ := getMsgType(data)
	switch typ {

	case NewDev, DevOnline, DevRename, DevDelete, ManualOpDev:
		msgHandle = NormalMsgHandle{
			data:data,
			msgType:typ,
		}

	case GtwOnline, RemoteOpDev:
		msgHandle = GtwMsgHandle{
			data:data,
			msgType:typ,
		}

	case SensorAlarm:
		msgHandle = SensorMsgHandle{
			data:data,
		}

	case InfraredTreasure:

	default:
		return nil
	}

	return
}

func getMsgType(data entity.FeibeeData) (typ MsgType) {
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

		if (data.Records[0].Cid == 1280 && data.Records[0].Aid == 128) {
			typ = SensorAlarm
		}

		if (data.Records[0].Cid == 1 && data.Records[0].Aid == 33) || (data.Records[0].Cid == 1 && data.Records[0].Aid == 53) {
			typ = SensorAlarm
		}

		if (data.Records[0].Cid == 1 && data.Records[0].Aid == 32) || (data.Records[0].Cid == 1 && data.Records[0].Aid == 62) {
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
	data    entity.FeibeeData
	msgType MsgType
}

func (self NormalMsgHandle) PushMsg() {
		res,bindid := createMsg2App(self.data, self.msgType)

		//发送给APP
		data2app, err := json.Marshal(res)
		if err != nil {
			log.Error("One Msg push2app() error = ", err)
		} else {
			producer.SendMQMsg2APP(bindid, string(data2app))
		}

		//发送给DB
		data2db, err := json.Marshal(entity.Feibee2DBMsg{
			res,
			bindid,
		})
		if err != nil {
			log.Error("One Msg push2db() error = ", err)
		} else {
			producer.SendMQMsg2Db(string(data2db))
		}

		//发送给PMS
		data2pms, err := json.Marshal(createMsg2pms(self.data, self.msgType))
		if err != nil {
			log.Error("One Msg push2pms() error = ", err)
		} else {
			producer.SendMQMsg2PMS(string(data2pms))
		}
}

type GtwMsgHandle struct {
	data entity.FeibeeData
	msgType MsgType
}

func (self GtwMsgHandle) PushMsg() {
		//发送给PMS
		data2pms, err := json.Marshal(createMsg2pms(self.data, self.msgType))
		if err != nil {
			log.Error("One Msg push2pms() error = ", err)
		} else {
			producer.SendMQMsg2PMS(string(data2pms))
		}
}

type SensorMsgHandle struct {
	data entity.FeibeeData
}

func (self SensorMsgHandle) PushMsg() {
	devAlarm := NewDevAlarm(self.data, 0)
	if devAlarm == nil {
		log.Error("该报警设备类型未支持")
		return
	}

	datas, err := devAlarm.GetMsg2app(0)
	if err != nil {
		log.Error("alarmMsg2app error = ", err)
		return
	}

	if len(datas) <= 0 {
		return
	}

	for _, data := range datas {
		if len(data) > 0 {
			producer.SendMQMsg2APP(self.data.Records[0].Bindid, string(data))
			producer.SendMQMsg2Db(string(data))
		}
	}

	data2pms, err := json.Marshal(createSceneMsg2pms(self.data))
	if err != nil {
		log.Error("One Msg push2pms() error = ", err)
	} else {
		producer.SendMQMsg2PMS(string(data2pms))
	}

	return
}

type RemoteOpMsgHandle struct {
	data entity.FeibeeData
	msgType MsgType
}

func createMsg2App(data entity.FeibeeData, msgType MsgType) (res entity.Feibee2AppMsg, bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = -1

	switch msgType {
	case NewDev, DevOnline, DevDelete, DevRename:
		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.Devid = data.Msg[0].Uuid
		res.Note = data.Msg[0].Name
		res.Deviceuid = data.Msg[0].Deviceuid
		res.Online = data.Msg[0].Online
		res.Battery = data.Msg[0].Battery

		bindid = data.Msg[0].Bindid

		switch msgType {
		case NewDev:
			res.OpType = "newDevice"
		case DevOnline:
			res.OpType = "newOnline"
			res.Battery = 0xff
		case DevDelete:
			res.OpType = "devDelete"
			res.Battery = 0xff
		case DevRename:
			res.OpType = "devNewName"
			res.Battery = 0xff
		}

	case ManualOpDev:
		res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
		res.Devid = data.Records[0].Uuid
		res.Deviceuid = data.Records[0].Deviceuid

		if data.Records[0].Value == "00" {
			res.OpType = "devOff"
		} else if data.Records[0].Value == "01" {
			res.OpType = "devOn"
		} else if data.Records[0].Value == "02" {
			res.OpType = "devStop"
		}

		bindid = data.Records[0].Bindid

	case RemoteOpDev:
		if data.Msg[0].Onoff == 1 {
			res.OpType = "devRemoteOn"
		} else if data.Msg[0].Onoff == 0 {
			res.OpType = "devRemoteOff"
		} else if data.Msg[0].Onoff == 2 {
			res.OpType = "devRemoteStop"
		}

		bindid = data.Records[0].Bindid
	}

	return
}

func createMsg2pms(data entity.FeibeeData, msgType MsgType) (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.FeibeeData = data

	switch msgType {

	case NewDev,DevOnline,DevDelete,DevRename,RemoteOpDev:
		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.DevId = data.Msg[0].Uuid
		res.Msg = []entity.FeibeeDevMsg{data.Msg[0]}
		res.Msg[0].Devicetype = res.DevType

	case ManualOpDev:
		res.DevType = devTypeConv(data.Records[0].Deviceuid, data.Records[0].Zonetype)
		res.DevId = data.Records[0].Uuid
		res.Records = []entity.FeibeeRecordsMsg{
			data.Records[0],
		}
		res.Records[0].Devicetype = res.DevType

	case GtwOnline:
		res.Gateway = []entity.FeibeeGatewayMsg{data.Gateway[0]}
	}

	return
}

func createSceneMsg2pms(data entity.FeibeeData) (res entity.FeibeeAutoScene2pmsMsg) {
	res.Cmd = 0xf1
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
	res.Devid = data.Records[0].Uuid
	res.TriggerType = 0
	res.Zone = "hz"

	return
}

func devTypeConv(devId, zoneType int) string {
	pre := strconv.FormatInt(int64(devId), 16)
	tail := strconv.FormatInt(int64(zoneType), 16)
	lenPre := len(pre)
	lenTail := len(tail)

	if lenPre < 4 {
		for i := 0; i < 4-lenPre; i++ {
			pre = "0" + pre
		}
	}

	if lenTail < 4 {
		for i := 0; i < 4-lenTail; i++ {
			tail = "0" + tail
		}
	}
	return "0x" + pre + tail
}