package feibee2srv

import (
	"das/core/constant"
	"das/core/util"
	"errors"
	"strconv"
	"strings"
	"time"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
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
	ZigbeeLock       //zigbee锁

	//小卫士场景
	FeibeeScene      //小卫士场景消息

)

var (
	ErrMsgStruct = errors.New("Feibee Msg structure was inValid")
	ErrLGuardValLens = errors.New("LGuard value lens error")
)

type MsgHandler interface {
	PushMsg()
}

//设备入网 设备离上线 设备删除 设备重命名
type NormalMsgHandle struct {
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *NormalMsgHandle) PushMsg() {
	res, routingKey, bindid := createMsg2App(self.data, self.msgType)

	//发送给APP
	data2app, err := json.Marshal(res)
	if err != nil {
		log.Error("One Msg push2app() error = ", err)
	} else {
		if self.msgType == NewDev {
			rabbitmq.Publish2app(data2app, bindid)
		} else {
			rabbitmq.Publish2app(data2app, routingKey)
		}
	}

	data2mns, err := json.Marshal(entity.Feibee2MnsMsg{
		res,
		bindid,
	})
	if err != nil {
		log.Error("One Msg push2db() error = ", err)
	} else {
		//producer.SendMQMsg2Db(string(data2db))
		rabbitmq.Publish2mns(data2mns, "")
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
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *GtwMsgHandle) PushMsg() {
	//发送给PMS
	data2pms, err := json.Marshal(createMsg2pms(self.data, self.msgType))
	if err != nil {
		log.Error("One Msg push2pms() error = ", err)
	} else {
		//producer.SendMQMsg2PMS(string(data2pms))
		rabbitmq.Publish2pms(data2pms, "")
	}

	//发送给app
	msg, routingKey, bindid := createMsg2App(self.data, self.msgType)
	data2app, err := json.Marshal(msg)
	if err != nil {
		log.Error("One Msg push2app(0 error = ", err)
	} else {
		//producer.SendMQMsg2APP(bindId, string(data2app))
		rabbitmq.Publish2app(data2app, routingKey)
	}

	//发送给ums
	data2ums, err := json.Marshal(entity.Feibee2MnsMsg{
		msg,
		bindid,
	})
	if err != nil {
		log.Error("One Msg push2db() error = ", err)
	} else {
		//producer.SendMQMsg2Db(string(data2db))
		rabbitmq.Publish2mns(data2ums, "")
	}

}

type SensorMsgHandle struct {
	data *entity.FeibeeData
}

func (self *SensorMsgHandle) PushMsg() {
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
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *InfraredTreasureHandle) PushMsg() {
	self.pushMsgByType()
}

func (self *InfraredTreasureHandle) pushMsgByType() {
	if len(self.data.Records[0].Value) < 2 {
		log.Error("InfraredTreasureHandle.pushMsgByType() error = msg type parse error")
		return
	}
	flag, err := strconv.ParseInt(self.data.Records[0].Value[0:2], 16, 64)
	if err != nil {
		log.Warning("InfraredTreasureHandle.pushMsgByType() error = ", err)
		return
	}

	data, routingKey, _ := createMsg2App(self.data, self.msgType)

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
			if err := self.push2app(data, routingKey); err != nil {
				log.Error("InfraredTreasureHandle.pushMsgByType() error = ", err)
			}
			return
		case 0x8200: //控制上报
			log.Debug("红外宝 控制上报: ", self.getControlResult())
		case 0x8300: //学习上报
			log.Debug("红外宝 学习上报")
			data.OpType = "devTrain"
			data.OpValue = self.getTrainResult()
			if err := self.push2app(data, routingKey); err != nil {
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

func (self *InfraredTreasureHandle) push2app(data entity.Feibee2AppMsg, routingKey string) error {
	if len(data.OpType) <= 0 || len(data.OpValue) <= 0 {
		err := errors.New("optype or opvalue error")
		return err
	}

	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//producer.SendMQMsg2APP(routingKey, string(rawData))
	rabbitmq.Publish2app(rawData, routingKey)
	return nil
}

func (self *InfraredTreasureHandle) getFirmwareVer() (ver string) {
	if len(self.data.Records[0].Value) == 22 {
		ver = self.data.Records[0].Value[8:20]
	}
	return
}

func (self *InfraredTreasureHandle) getMatchResult() (res string) {
	if len(self.data.Records[0].Value) < 30 {
		return
	}
	return self.parseValue(26, 28)
}

func (self *InfraredTreasureHandle) getTrainResult() (res string) {
	if len(self.data.Records[0].Value) < 34 {
		return
	}
	return self.parseValue(30, 32)
}

func (self *InfraredTreasureHandle) getControlResult() string {
	var msg entity.InfraredTreasureControlResult
	if len(self.data.Records[0].Value) < 32 {
		return ""
	}

	msg.Uuid = self.data.Records[0].Uuid
	msg.DevType = "红外宝"
	msg.FirmVer = self.data.Records[0].Value[8:20]
	msg.ControlDevType, _ = strconv.ParseInt(self.data.Records[0].Value[24:26], 16, 64)
	msg.FunctionKey = parseFuncKey(self.data.Records[0].Value[26:30])

	data, _ := json.Marshal(msg)
	return string(data)
}

func parseFuncKey(src string) int64 {
	keyStr := src[2:4] + src[0:2]
	res, _ := strconv.ParseInt(keyStr, 16, 64)
	return res
}

func (self *InfraredTreasureHandle) parseValue(start, end int) (res string) {
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

type WonlyLGuardHandle struct {
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *WonlyLGuardHandle) PushMsg() {
	//msg2pms := createMsg2pms(self.data, ManualOpDev)
	//data2pms, err := json.Marshal(msg2pms)
	//if err != nil {
	//	log.Warning("WonlyLGuardHandle msg2pms json.Marshal() error = ", err)
	//} else {
	//	rabbitmq.Publish2pms(data2pms, "")
	//}

	msg2mns, routingKey, err := self.createOtherMsg2App()
	if err != nil {
		log.Warning(err)
		return
	}
	data2mns, err := json.Marshal(msg2mns)
	if err != nil {
		log.Warning("WonlyLGuardHandle msg2db json.Marshal() error = ", err)
	} else {
		rabbitmq.Publish2mns(data2mns, "")
		rabbitmq.Publish2app(data2mns, routingKey)
	}

	self.sendMsg2pmsForSceneTrigger()
}

func (self *WonlyLGuardHandle) sendMsg2pmsForSceneTrigger() {
	//todo: parse lguard value
	_, val, err := self.parseValue(self.data.Records[0].Value)
	if err != nil {
		log.Error("WonlyLGuardHandle sendMsg2pmsForSceneTrigger() error = ", err)
		return
	}

	msg := createSceneMsg2pms(self.data, val, "WonlyLGuard")

	data2pms,err := json.Marshal(msg)
	if err != nil {
		log.Error("WonlyLGuardHandle sendMsg2pmsForSceneTrigger() error = ", err)
	} else {
		rabbitmq.Publish2pms(data2pms, "")
	}
}

func (self *WonlyLGuardHandle) parseValue(rawVal string) (typ int64, val string, err error){
    if len(rawVal) < 10 {
    	return typ, "", ErrLGuardValLens
	}

    rawValLens,err := strconv.ParseInt(rawVal[0:2], 16, 32)
    if err != nil || rawValLens != int64(len(rawVal[2:]))/2 {
		return  typ,"", ErrLGuardValLens
	}

	lens, err := strconv.ParseInt(rawVal[4:6], 16, 64)
	if err != nil || int64(len(rawVal[6:])/2) < lens+3 {
		return typ,"", ErrLGuardValLens
	}

	typ, err = strconv.ParseInt(rawVal[6:8], 16, 64)
	if err != nil {
		return typ,"", ErrLGuardValLens
	}

	funcData := rawVal[8:8+2*lens]

	switch typ {
	case 0x23:
		if funcData == "00" {
			val = "撤防"
		} else if funcData == "01" {
			val = "布防"
		}
	}
	return
}

func (self *WonlyLGuardHandle) createOtherMsg2App() (res entity.Feibee2MnsMsg, routingKey string, err error) {
	if len(self.data.Records) <= 0 {
		err = ErrMsgStruct
		return
	}

	res.Cmd = 0xfb
	res.Ack = 1
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())

	res.DevId = self.data.Records[0].Uuid
	res.Deviceuid = self.data.Records[0].Deviceuid

	res.OpType = "WonlyLGuard"
	res.OpValue = self.data.Records[0].Value

	res.Bindid = self.data.Records[0].Bindid
	res.DevType = devTypeConv(self.data.Records[0].Deviceid, self.data.Records[0].Zonetype)

	routingKey = self.data.Records[0].Uuid
	return
}

func (self *WonlyLGuardHandle) createNewDevMsg2App() (res entity.Feibee2AppMsg, routingKey,bindid string) {
	if self.data.Code == 3 {
		res, routingKey, bindid = createMsg2App(self.data, NewDev)
	} else if self.data.Code == 4 {
		res, routingKey, bindid = createMsg2App(self.data, DevOnline)
	} else if self.data.Code == 5 {
		res, routingKey, bindid = createMsg2App(self.data, DevDelete)
	}

	return
}

type SceneSwitchHandle struct {
	data *entity.FeibeeData
}

func (self *SceneSwitchHandle) PushMsg() {

	sceneMsg2pms := self.createSceneMsg2pms()
	sceneData2pms, err := json.Marshal(sceneMsg2pms)
	if err != nil {
		log.Warning("SceneSwitchHandle sceneMsg2pms json.Marshal() error = ", err)
	} else {
		//producer.SendMQMsg2PMS(string(sceneData2pms))
		rabbitmq.Publish2pms(sceneData2pms, "")
	}
}

func (self *SceneSwitchHandle) createSceneMsg2pms() (res entity.FeibeeAutoScene2pmsMsg) {
	//情景开关作为无触发值的触发设备
	res = createSceneMsg2pms(self.data, "", "sceneSwitch")
	return
}

type ZigbeeLockHandle struct {
	data *entity.FeibeeData
}

func (self *ZigbeeLockHandle) PushMsg() {
	// 标识设备接入平台，为下行做准备
	mymap := make(map[string]interface{})
	mymap["uuid"] = self.data.Records[0].Uuid
	mymap["uid"] = self.data.Records[0].Deviceuid
	mymap["from"] = constant.FEIBEE_PLATFORM
	retUuid := make([]string, 4)
	if "" != self.data.Records[0].Uuid { // uuid不能为空
		retUuid = strings.FieldsFunc(self.data.Records[0].Uuid, util.Split) // 去掉下划线后边，如：_01
		redis.SetDevicePlatformPool(retUuid[0], mymap)
	} else {
		return
	}

	//todo: parse data and handle
    if err := ParseZlockData(self.data.Records[0].Value, "WlZigbeeLock", retUuid[0]); err != nil {
    	log.Warning("ZigbeeLockHandle PushMsg() error = ", err)
	}
}

type FeibeeSceneHandle struct {
    data *entity.FeibeeData
}

func (self *FeibeeSceneHandle) PushMsg() {
    msg2mns := self.createMsg2Mns()
    data2mns,err := json.Marshal(msg2mns)
    if err != nil {
		log.Warning("FeibeeSceneHandle PushMsg json.Marshal() error = ", err)
	} else {
		rabbitmq.Publish2mns(data2mns, "")
	}
}

func (self *FeibeeSceneHandle) createMsg2Mns() (res entity.Feibee2MnsMsg){
	res.Header.Cmd = 0xfb
	res.Header.Vendor = "feibee"
	res.Header.SeqId = 1
	res.SceneMessages = self.data.SceneMessages

    switch self.data.Code {
	case 21:
		res.OpType = "addScene"
	case 22:
		res.OpType = "removeScene"
	case 23:
		res.OpType = "sceneName"
	}

	return
}

func createMsg2App(data *entity.FeibeeData, msgType MsgType) (res entity.Feibee2AppMsg, routingKey,bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())
	res.Battery = 0xff

	switch msgType {
	case NewDev, DevOnline, DevDelete, DevRename, DevDegree:
		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.DevId = data.Msg[0].Uuid
		res.Note = data.Msg[0].Name
		res.Deviceuid = data.Msg[0].Deviceuid
		res.Online = data.Msg[0].Online
		res.Battery = data.Msg[0].Battery

		bindid = data.Msg[0].Bindid

		switch msgType {
		case NewDev:
			res.OpType = "newDevice"
			//新入网设备online字段默认为1
			if res.Online <= 0 {
				res.Online = 1
			}
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
		res.DevId = data.Records[0].Uuid
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

		if res.Online <= 0 {
			res.Online = 1
		}

		res.DevType = devTypeConv(data.Msg[0].Deviceid, data.Msg[0].Zonetype)
		res.DevId = data.Msg[0].Uuid
		res.Deviceuid = data.Msg[0].Deviceuid
		bindid = data.Msg[0].Bindid
	}

	routingKey = res.DevId
	return
}

func createMsg2pms(data *entity.FeibeeData, msgType MsgType) (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.FeibeeData = *data

	switch msgType {

	case NewDev, DevOnline, DevDelete, DevRename, RemoteOpDev:
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

func createSceneMsg2pms(data *entity.FeibeeData, alarmValue, alarmType string) (res entity.FeibeeAutoScene2pmsMsg) {
	res.Cmd = 0xf1
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
	res.DevId = data.Records[0].Uuid
	res.TriggerType = 0
	res.Time = int(time.Now().Unix())

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

func isDevAlarm(data *entity.FeibeeData) bool {

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
	if (data.Records[0].Cid == 9 && data.Records[0].Aid == 61685) {
		//zigbee锁报警
		return true
	}

	return false
}
