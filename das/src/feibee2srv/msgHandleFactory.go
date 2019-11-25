package feibee2srv

import (
	"errors"
	"strconv"
	"time"

	"../core/entity"
	"../core/log"
	"../core/rabbitmq"
)

type MsgType int32

const (
	NewDev    MsgType = iota + 1 //设备入网
	DevOnline                    //设备离上线
	DevDelete                    //设备删除
	DevRename                    //设备重命名
	GtwOnline                    //网关离上线
	DevDegree                    //设备控制程度通知（窗帘开关程度）

	//设备操作消息
	ManualOpDev //手动操作设备
	RemoteOpDev //远程操作设备

	//特殊设备
	SensorAlarm      //传感器报警
	InfraredTreasure //红外宝
	WonlyLGuard      //小卫士
	SceneSwitch      //情景开关
	ZigbeeLock       //飞比zigbee锁
)

var (
	ErrMsgStruct = errors.New("Feibee Msg structure was inValid")
)

type MsgHandler interface {
	PushMsg()
}

func MsgHandleFactory(data entity.FeibeeData) (msgHandle MsgHandler) {
	typ := getMsgType(data)
	switch typ {

	case NewDev, DevOnline, DevRename, DevDelete, ManualOpDev, DevDegree:
		msgHandle = NormalMsgHandle{
			data:    data,
			msgType: typ,
		}

	case GtwOnline, RemoteOpDev:
		msgHandle = GtwMsgHandle{
			data:    data,
			msgType: typ,
		}

	case SensorAlarm:
		msgHandle = SensorMsgHandle{
			data: data,
		}

	case InfraredTreasure:
		msgHandle = InfraredTreasureHandle{
			data:    data,
			msgType: typ,
		}

	case WonlyLGuard:
		msgHandle = WonlyGuardHandle{
			data:    data,
			msgType: typ,
		}

	case SceneSwitch:
		msgHandle = SceneSwitchHandle{
			data: data,
		}
	case ZigbeeLock:

	default:
		msgHandle = nil
	}

	return
}

func getMsgType(data entity.FeibeeData) (typ MsgType) {
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
		} else if data.Records[0].Deviceid == 0xa {
			//飞比zigbee锁
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

//设备入网 设备离上线 设备删除 设备重命名
type NormalMsgHandle struct {
	data    entity.FeibeeData
	msgType MsgType
}

func (self NormalMsgHandle) PushMsg() {
	res, bindid := createMsg2App(self.data, self.msgType)

	//发送给APP
	data2app, err := json.Marshal(res)
	if err != nil {
		log.Error("One Msg push2app() error = ", err)
	} else {
		//producer.SendMQMsg2APP(bindid, string(data2app))
		rabbitmq.Publish2app(data2app, bindid)
	}

	//发送给ums(设备入网不发?)
	if self.msgType == NewDev {

	} else {
		data2ums, err := json.Marshal(entity.Feibee2DBMsg{
			res,
			bindid,
		})
		if err != nil {
			log.Error("One Msg push2db() error = ", err)
		} else {
			//producer.SendMQMsg2Db(string(data2db))
			rabbitmq.Publish2mns(data2ums, "")
		}
	}

	//发送给PMS
	data2pms, err := json.Marshal(createMsg2pms(self.data, self.msgType))
	if err != nil {
		log.Error("One Msg push2pms() error = ", err)
	} else {
		//producer.SendMQMsg2PMS(string(data2pms))
		rabbitmq.Publish2pms(data2pms, "")
	}
}

type GtwMsgHandle struct {
	data    entity.FeibeeData
	msgType MsgType
}

func (self GtwMsgHandle) PushMsg() {
	//发送给PMS
	data2pms, err := json.Marshal(createMsg2pms(self.data, self.msgType))
	if err != nil {
		log.Error("One Msg push2pms() error = ", err)
	} else {
		//producer.SendMQMsg2PMS(string(data2pms))
		rabbitmq.Publish2pms(data2pms, "")
	}

	//发送给app
	msg, bindId := createMsg2App(self.data, self.msgType)
	data2app, err := json.Marshal(msg)
	if err != nil {
		log.Error("One Msg push2app(0 error = ", err)
	} else {
		//producer.SendMQMsg2APP(bindId, string(data2app))
		rabbitmq.Publish2app(data2app, bindId)
	}

	//发送给ums
	data2ums, err := json.Marshal(entity.Feibee2DBMsg{
		msg,
		bindId,
	})
	if err != nil {
		log.Error("One Msg push2db() error = ", err)
	} else {
		//producer.SendMQMsg2Db(string(data2db))
		rabbitmq.Publish2mns(data2ums, "")
	}

}

type SensorMsgHandle struct {
	data entity.FeibeeData
}

func (self SensorMsgHandle) PushMsg() {
	devAlarm := DevAlarmFactory(self.data)
	if devAlarm == nil {
		log.Error("该报警设备类型未支持")
		return
	}

	devAlarm.PushMsg()
}

type RemoteOpMsgHandle struct {
	data    entity.FeibeeData
	msgType MsgType
}

type InfraredTreasureHandle struct {
	data    entity.FeibeeData
	msgType MsgType
}

func (self InfraredTreasureHandle) PushMsg() {
	self.pushMsgByType()
}

func (self InfraredTreasureHandle) pushMsgByType() {
	if len(self.data.Records[0].Value) < 2 {
		log.Error("InfraredTreasureHandle.pushMsgByType() error = msg type parse error")
		return
	}
	flag, err := strconv.ParseInt(self.data.Records[0].Value[0:2], 16, 64)
	if err != nil {
		log.Warning("InfraredTreasureHandle.pushMsgByType() error = ", err)
		return
	}

	data, bindid := createMsg2App(self.data, self.msgType)

	switch flag {
	case 10: //红外宝固件版本上报
		log.Debug("红外宝 固件版本上报")

	case 5: //码组上传上报
		log.Debug("红外宝 码组上传上报")

	default:
		if len(self.data.Records[0].Value) < 24 {
			log.Warning("InfraredTreasureHandle.pushMsgByType() error = msg type parse error")
			return
		}

		funcCode, err := strconv.ParseInt(self.data.Records[0].Value[20:24], 16, 64)
		if err != nil {
			log.Warning("InfraredTreasureHandle.pushMsgByType() error = ", err)
			return
		}

		switch funcCode {
		case 0x8100: //匹配上报
			log.Debug("红外宝 匹配上报")
			data.OpType = "devMatch"
			data.OpValue = self.getMatchResult()
			if err := self.push2app(data, bindid); err != nil {
				log.Error("InfraredTreasureHandle.pushMsgByType() error = ", err)
			}
			return
		case 0x8200: //控制上报
			log.Debug("红外宝 控制上报")
		case 0x8300: //学习上报
			log.Debug("红外宝 学习上报")
			data.OpType = "devTrain"
			data.OpValue = self.getTrainResult()
			if err := self.push2app(data, bindid); err != nil {
				log.Error("InfraredTreasureHandle.pushMsgByType() error = ", err)
			}
			return
		case 0x8700: //码库更新通知上报
			log.Debug("红外宝 码库更新通知上报")
		case 0x8800: //码库保存上报
			log.Debug("红外宝 码库保存上报")
		}
	}
	return
}

func (self InfraredTreasureHandle) push2app(data entity.Feibee2AppMsg, bindid string) error {
	if len(data.OpType) <= 0 || len(data.OpValue) <= 0 {
		err := errors.New("optype or opvalue error")
		return err
	}

	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//producer.SendMQMsg2APP(bindid, string(rawData))
	rabbitmq.Publish2app(rawData, bindid)
	return nil
}

func (self InfraredTreasureHandle) getFirmwareVer() (ver string) {
	if len(self.data.Records[0].Value) == 22 {
		ver = self.data.Records[0].Value[8:20]
	}
	return
}

func (self InfraredTreasureHandle) getMatchResult() (res string) {
	if len(self.data.Records[0].Value) < 30 {
		return
	}
	return self.parseValue(26, 28)
}

func (self InfraredTreasureHandle) getTrainResult() (res string) {
	if len(self.data.Records[0].Value) < 34 {
		return
	}
	return self.parseValue(30, 32)
}

func (self InfraredTreasureHandle) parseValue(start, end int) (res string) {
	if start < 0 || end > len(self.data.Records[0].Value) {
		return
	}

	flag, err := strconv.ParseInt(self.data.Records[0].Value[start:end], 16, 64)
	if err != nil {
		log.Error("InfraredTreasureHandle.getMatchResult() error = ", err)
		return
	}
	switch flag {
	case 0:
		res = "0"
	case 1:
		res = "1"
	case 2:
		res = "2"
	}
	return
}

type WonlyGuardHandle struct {
	data    entity.FeibeeData
	msgType MsgType
}

func (self WonlyGuardHandle) pushMsgByType() {

	switch self.data.Code {
	//设备入网
	case 3:
		msg2pms := self.createMsg2PMS()
		data2pms, err := json.Marshal(msg2pms)
		if err != nil {
			log.Warning("WonlyGuardHandle msg2pms json.Marshal() error = ", err)
		} else {
			//producer.SendMQMsg2PMS(string(data2pms))
			rabbitmq.Publish2pms(data2pms, "")
		}

		msg2app,routingKey,err := self.createNewDevMsg2App()
		if err != nil {
			log.Warning(err)
			return
		}

		data2app, err := json.Marshal(msg2app)
		if err != nil {
			log.Warning("WonlyGuardHandle msg2pms json.Marshal() error = ", err)
		} else {
			//producer.SendGuardMsg2APP(routingKey, data2app)
			rabbitmq.PublishGuard2app(data2app, routingKey)
		}

	case 2:
		msg2ums,routingKey,err := self.createOtherMsg2App()
		if err != nil {
			log.Warning(err)
			return
		}
		data2ums, err := json.Marshal(msg2ums)
		if err != nil {
			log.Warning("WonlyGuardHandle msg2db json.Marshal() error = ", err)
		} else {
			//producer.SendMQMsg2Db(string(data2db))
			//rabbitmq.Publish2mns(data2ums, "")
			rabbitmq.Publish2app(data2ums, routingKey)
		}

	}
}

func (self WonlyGuardHandle) PushMsg() {
	self.pushMsgByType()
}

func (self WonlyGuardHandle) createMsg2PMS() (res entity.Feibee2PMS) {
	return createMsg2pms(self.data, self.msgType)
}

func (self WonlyGuardHandle) createOtherMsg2App() (res entity.Feibee2DBMsg, routingKey string, err error) {
	if len(self.data.Records) <= 0 {
		err = ErrMsgStruct
		return
	}

	res.Cmd = 0xfb
	res.Ack = 1
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())

	res.Devid = self.data.Records[0].Uuid
	res.Deviceuid = self.data.Records[0].Deviceuid

	res.OpType = "WonlyLGuard"
	res.OpValue = self.data.Records[0].Value

	res.Bindid = self.data.Records[0].Bindid
	res.DevType = devTypeConv(self.data.Records[0].Deviceid, self.data.Records[0].Zonetype)

	routingKey = self.data.Records[0].Uuid
	return
}

func (self WonlyGuardHandle) createNewDevMsg2App() (res entity.Feibee2AppMsg,routingKey string, err error) {
	if len(self.data.Msg) <= 0 {
		err = ErrMsgStruct
		return
	}
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())

	res.Battery = self.data.Msg[0].Battery
	res.Devid = self.data.Msg[0].Uuid
	res.Deviceuid = self.data.Msg[0].Deviceuid
	res.Online = self.data.Msg[0].Online
	res.DevType = devTypeConv(self.data.Msg[0].Deviceid, self.data.Msg[0].Zonetype)

	res.OpType = "newDevice"
	routingKey = self.data.Msg[0].Uuid

	return
}

type SceneSwitchHandle struct {
	data entity.FeibeeData
}

func (self SceneSwitchHandle) PushMsg() {

	sceneMsg2pms := self.createSceneMsg2pms()
	sceneData2pms, err := json.Marshal(sceneMsg2pms)
	if err != nil {
		log.Warning("SceneSwitchHandle sceneMsg2pms json.Marshal() error = ", err)
	} else {
		//producer.SendMQMsg2PMS(string(sceneData2pms))
		rabbitmq.Publish2pms(sceneData2pms, "")
	}
}

func (self SceneSwitchHandle) createSceneMsg2pms() (res entity.FeibeeAutoScene2pmsMsg) {
	//情景开关作为无触发值的触发设备
	res = createSceneMsg2pms(self.data, "", "sceneSwitch")
	return
}

type ZigbeeLockHandle struct {
	data entity.FeibeeData
}

func (self ZigbeeLockHandle) pushByMsgType() {
	cid, aid := self.data.Records[0].Cid, self.data.Records[0].Aid

	if cid == 257 && aid == 61685 {
		//远程开锁
	} else if cid == 257 && aid == 0 {
		//门锁状态
	} else if cid == 9 && aid == 61685 {
		//报警上报
	} else if cid == 10 && aid == 0 {
		//门锁时间
	} else if cid == 1 && aid == 33 {
		//电量
	} else if cid == 1 && aid == 62 {
		//电压
	} else if cid == 1280 && aid == 130 {
		//门铃上报
	}
}

func (self ZigbeeLockHandle) PushMsg() {

}

func createMsg2App(data entity.FeibeeData, msgType MsgType) (res entity.Feibee2AppMsg, bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())
	res.Battery = 0xff

	switch msgType {
	case NewDev, DevOnline, DevDelete, DevRename, DevDegree:
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
		case DevDelete:
			res.OpType = "devDelete"
		case DevRename:
			res.OpType = "devNewName"
		case DevDegree:
			res.OpType = "devDegree"
			res.OpValue = strconv.Itoa(data.Msg[0].DevDegree)
		}

	case ManualOpDev, InfraredTreasure:
		res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
		res.Devid = data.Records[0].Uuid
		res.Deviceuid = data.Records[0].Deviceuid
		bindid = data.Records[0].Bindid

		switch msgType {

		case ManualOpDev:
			if data.Records[0].Value == "00" {
				res.OpType = "devOff"
			} else if data.Records[0].Value == "01" {
				res.OpType = "devOn"
			} else if data.Records[0].Value == "02" {
				res.OpType = "devStop"
			}

		case InfraredTreasure:

		}

	case RemoteOpDev:
		if data.Msg[0].Onoff == 1 {
			res.OpType = "devRemoteOn"
		} else if data.Msg[0].Onoff == 0 {
			res.OpType = "devRemoteOff"
		} else if data.Msg[0].Onoff == 2 {
			res.OpType = "devRemoteStop"
		}
		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.Devid = data.Msg[0].Uuid
		res.Deviceuid = data.Msg[0].Deviceuid
		bindid = data.Msg[0].Bindid
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

	case NewDev, DevOnline, DevDelete, DevRename, RemoteOpDev, WonlyLGuard:
		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.DevId = data.Msg[0].Uuid
		res.Msg = []entity.FeibeeDevMsg{data.Msg[0]}
		res.Msg[0].Devicetype = res.DevType

	case ManualOpDev:
		res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
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

func createSceneMsg2pms(data entity.FeibeeData, alarmValue, alarmType string) (res entity.FeibeeAutoScene2pmsMsg) {
	res.Cmd = 0xf1
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
	res.Devid = data.Records[0].Uuid
	res.TriggerType = 0

	res.AlarmValue = alarmValue
	res.AlarmType = alarmType

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

func isDevAlarm(data entity.FeibeeData) bool {

	if data.Records[0].Cid == 1280 && data.Records[0].Aid == 128 {
		return true
	}

	if data.Records[0].Cid == 1024 && data.Records[0].Aid == 0 {
		//光照度上报
		return true
	}
	if data.Records[0].Cid == 1026 && data.Records[0].Aid == 0 {
		//温度上报
		return true
	}
	if data.Records[0].Cid == 1029 && data.Records[0].Aid == 0 {
		//湿度上报
		return true
	}
	if (data.Records[0].Cid == 1 && data.Records[0].Aid == 33) || (data.Records[0].Cid == 1 && data.Records[0].Aid == 53) {
		//电量上报
		return true
	}
	if (data.Records[0].Cid == 1 && data.Records[0].Aid == 32) || (data.Records[0].Cid == 1 && data.Records[0].Aid == 62) {
		//电压上报
		return true
	}

	return false
}
