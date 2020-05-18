package feibee2srv

import (
	"das/core/redis"
	"errors"
	"strconv"
	"time"

	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)


var (
	ErrMsgStruct = errors.New("Feibee Msg structure was inValid")
	ErrLGuardVal = errors.New("LGuard value not support")
)

type MsgHandler interface {
	PushMsg()
}

//设备入网 设备离上线 设备删除 设备重命名
type NormalMsgHandle struct {
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *NormalMsgHandle) createMsg2App() (res entity.Feibee2DevMsg, routingKey,bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())
	res.Battery = 0xff

	res.DevType = devTypeConv(self.data.Msg[0].Deviceid, self.data.Msg[0].Zonetype)
	res.DevId = self.data.Msg[0].Uuid
	res.Note = self.data.Msg[0].Name
	res.Deviceuid = self.data.Msg[0].Deviceuid
	res.Online = self.data.Msg[0].Online
	res.Battery = self.data.Msg[0].Battery
	res.Bindid = self.data.Msg[0].Bindid
	res.Snid = self.data.Msg[0].Snid

	bindid = self.data.Msg[0].Bindid

	switch self.msgType {
	case NewDev:
		res.OpType = "newDevice"
		//todo: 若online=0，则该设备可能已经在其他网关下
		//if res.Online <= 0 {
		//	res.Online = 1
		//}
	case DevOnline:
		res.OpType = "newOnline"
	case DevDelete:
		res.OpType = "devDelete"
	case DevRename:
		res.OpType = "devNewName"
	case DevDegree:
		res.OpType = "devDegree"
		res.OpValue = strconv.Itoa(self.data.Msg[0].DevDegree)
	case RemoteOpDev:
		if self.data.Msg[0].Onoff == 1 {
			res.OpType = "devRemoteOn"
		} else if self.data.Msg[0].Onoff == 0 {
			res.OpType = "devRemoteOff"
		} else if self.data.Msg[0].Onoff == 2 {
			res.OpType = "devRemoteStop"
		}

		if res.Online <= 0 {
			res.Online = 1
		}
	}

	routingKey = self.data.Msg[0].Uuid

	return
}

func (self *NormalMsgHandle) createMsg2pms() (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.FeibeeData = *self.data

	res.DevType = devTypeConv(self.data.Msg[0].Deviceid, self.data.Msg[0].Zonetype)
	res.DevId = self.data.Msg[0].Uuid
	res.Msg = []entity.FeibeeDevMsg{self.data.Msg[0]}
	res.Msg[0].Devicetype = res.DevType

	return
}

func (self *NormalMsgHandle) PushMsg() {
	res, routingKey, bindid := self.createMsg2App()

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
		rabbitmq.Publish2mns(data2app, "")
	}

	//情景开关以ieee作为routingKey推送
	if self.msgType == DevOnline && self.data.Msg[0].Deviceid == 0x0004 {
        routingKey = self.data.Msg[0].IEEE
        rabbitmq.Publish2app(data2app, routingKey)
	}

	if self.msgType == NewDev && res.Online < 1 {
		log.Warningf("设备'%s'在网关'%s'下入网，但该设备已绑定其他网关", res.DevId, bindid)
	} else {
		//发送给PMS
		data2pms, err := json.Marshal(self.createMsg2pms())
		if err != nil {
			log.Error("One Msg push2pms() error = ", err)
		} else {
			//producer.SendMQMsg2PMS(string(data2pms))
			rabbitmq.Publish2pms(data2pms, "")
		}
	}
}

type ManualOpMsgHandle struct {
	data *entity.FeibeeData
}

func (self *ManualOpMsgHandle) PushMsg() {
	res, routingKey, _ := self.createMsg2App()
	//发送给APP
	data2app, err := json.Marshal(res)
	if err != nil {
		log.Error("One Msg push2app() error = ", err)
	} else {
		rabbitmq.Publish2app(data2app, routingKey)
		rabbitmq.Publish2mns(data2app, "")
	}

	//发送给PMS
	data2pms, err := json.Marshal(self.createMsg2pms())
	if err != nil {
		log.Error("One Msg push2pms() error = ", err)
	} else {
		rabbitmq.Publish2pms(data2pms, "")
	}
}

func (self *ManualOpMsgHandle) createMsg2pms() (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.FeibeeData = *self.data
	res.DevType = devTypeConv(self.data.Records[0].Deviceid, self.data.Records[0].Zonetype)
	res.DevId = self.data.Records[0].Uuid
	res.Records = []entity.FeibeeRecordsMsg{
		self.data.Records[0],
	}
	res.Records[0].Devicetype = res.DevType
	return
}

func (self *ManualOpMsgHandle) createMsg2App() (res entity.Feibee2DevMsg, routingKey,bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())
	res.Battery = 0xff
	res.DevType = devTypeConv(self.data.Records[0].Deviceid, self.data.Records[0].Zonetype)
	res.DevId = self.data.Records[0].Uuid
	res.Deviceuid = self.data.Records[0].Deviceuid
	bindid = self.data.Records[0].Bindid
	res.Bindid = bindid
	if self.data.Records[0].Value == "00" {
		res.OpType = "devOff"
	} else if self.data.Records[0].Value == "01" {
		res.OpType = "devOn"
	} else if self.data.Records[0].Value == "02" {
		res.OpType = "devStop"
	}
	routingKey = res.DevId
	return
}

type GtwMsgHandle struct {
	data    *entity.FeibeeData
}

func (self *GtwMsgHandle) createMsg2mns() (res entity.Feibee2DevMsg) {
	res.Cmd = 0xfb
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Bindid = self.data.Gateway[0].Bindid
	res.DevId = res.Bindid
	res.OpType = "gtwVer"
	res.OpValue = self.data.Gateway[0].Version
	return
}

func (self *GtwMsgHandle) createMsg2pms() (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.FeibeeData = *self.data
	res.Gateway = []entity.FeibeeGatewayMsg{self.data.Gateway[0]}
	return
}

func (self *GtwMsgHandle) PushMsg() {
	//发送给PMS
	data2pms, err := json.Marshal(self.createMsg2pms())
	if err != nil {
		log.Errorf("GtwMsgHandle.PushMsg > json.Marshal > pms > %s", err)
	} else {
		rabbitmq.Publish2pms(data2pms, "")
	}

	data2mns,err := json.Marshal(self.createMsg2mns())
	if err != nil {
		log.Errorf("GtwMsgHandle.PushMsg > json.Marshal > mns > %s", err)
	} else {
		rabbitmq.Publish2mns(data2mns, "")
	}
}

type InfraredTreasureHandle struct {
	data    *entity.FeibeeData
	msgType MsgType
}

func (self *InfraredTreasureHandle) createMsg2App() (res entity.Feibee2DevMsg, routingKey,bindid string) {
	res.Cmd = 0xfb
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.Time = int(time.Now().Unix())
	res.Battery = 0xff

	res.DevType = devTypeConv(self.data.Records[0].Deviceid, self.data.Records[0].Zonetype)
	res.DevId = self.data.Records[0].Uuid
	res.Deviceuid = self.data.Records[0].Deviceuid
	bindid = self.data.Records[0].Bindid
	routingKey = self.data.Records[0].Uuid

	return
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

	data, routingKey, _ := self.createMsg2App()

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

func (self *InfraredTreasureHandle) push2app(data entity.Feibee2DevMsg, routingKey string) error {
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
	_, val, err := self.parseValue(self.data.Records[0].Value)
	if err != nil {
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

func (self *WonlyLGuardHandle) parseValue(rawVal string) (typ int64, val int, err error){
    if len(rawVal) < 10 {
    	return typ, 0, ErrLGuardVal
	}

    rawValLens,err := strconv.ParseInt(rawVal[0:2], 16, 32)
    if err != nil || rawValLens != int64(len(rawVal[2:]))/2 {
		return  typ,0, ErrLGuardVal
	}

	lens, err := strconv.ParseInt(rawVal[4:6], 16, 64)
	if err != nil || int64(len(rawVal[6:])/2) < lens+3 {
		return typ,0, ErrLGuardVal
	}

	typ, err = strconv.ParseInt(rawVal[6:8], 16, 64)
	if err != nil {
		return typ,0, ErrLGuardVal
	}

	funcData := rawVal[8:8+2*lens]

	switch typ {
	case 0x23:
		if funcData == "00" {
			val = 0
		} else if funcData == "01" {
			val = 1
		}
	default:
		return typ, 0, ErrLGuardVal
	}
	return
}

func (self *WonlyLGuardHandle) createOtherMsg2App() (res entity.Feibee2DevMsg, routingKey string, err error) {
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
	res.Snid = self.data.Records[0].Snid

	routingKey = self.data.Records[0].Uuid
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

func (self *SceneSwitchHandle) createSceneMsg2pms() (res entity.Feibee2AutoSceneMsg) {
	//情景开关作为无触发值的触发设备
	res = createSceneMsg2pms(self.data, 1, "sceneSwitch")
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
	redis.SetDevicePlatformPool(self.data.Records[0].Uuid, mymap)

	//TODO: parse data and handle, 去掉飞比加上的长度1个字节（16进制字符串2位）
	if 22 < len(self.data.Records[0].Value) {	// 包头长度为11字节
		nStart, err := strconv.ParseInt(self.data.Records[0].Value[2:4], 16, 64) // 去掉飞比的两位长度，再取两位[2:4)为王力数据包开始位0xA5
		if err != nil {
			log.Errorf("strconv.Atoi err: ", err)
			return
		}
		if 0xA5 == nStart { // 透传的zigbee常在线锁数据
			if err := ParseZlockData(self.data.Records[0].Value[2:], "WlZigbeeLock", self.data.Records[0].Uuid); err != nil {
				log.Warning("ZigbeeLockHandle PushMsg() error = ", err)
			}
		}
	}
}

type FeibeeSceneHandle struct {
    data *entity.FeibeeData
}

func (self *FeibeeSceneHandle) PushMsg() {
    msg2mns := self.createMsg2mns()
    data2mns,err := json.Marshal(msg2mns)
    if err != nil {
		log.Warning("FeibeeSceneHandle PushMsg json.Marshal() error = ", err)
	} else {
		rabbitmq.Publish2mns(data2mns, "")
	}
}

func (self *FeibeeSceneHandle) createMsg2mns() (res entity.Feibee2DevMsg){
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

func createSceneMsg2pms(data *entity.FeibeeData, alarmFlag int, alarmType string) (res entity.Feibee2AutoSceneMsg) {
	res.Cmd = 0xf1
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.DevType = devTypeConv(data.Records[0].Deviceid, data.Records[0].Zonetype)
	res.DevId = data.Records[0].Uuid
	res.TriggerType = 0
	res.Time = int(time.Now().Unix())

	res.AlarmFlag = alarmFlag
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
