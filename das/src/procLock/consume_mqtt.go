package procLock

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"das/core/constant"
	"das/core/entity"
	"das/core/jobque"
	"das/core/log"
	dasMqtt "das/core/mqtt"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/wlprotocol"
)

var (
	mqttCli mqtt.Client

	conTopic = "$SYS/brokers/+/clients/+/connected"
	disTopic = "$SYS/brokers/+/clients/+/disconnected"

	topic2Dev       string = "wonly/things/smartlock/"
	topic2Pad       string = "wonly/things/smartpad/"
)

type MqttJob struct {
	rawData []byte
}

func NewMqttJob(rawData []byte) MqttJob {
	return MqttJob{
		rawData: rawData,
	}
}

func (o MqttJob) Handle() {
	ParseData(o.rawData)
}

func initMqtt() {
	cfg := &dasMqtt.MqttCfg{}
	err := initWlMqttCfg(cfg)
	if err != nil {
		panic(err)
	}

	mqttCli = dasMqtt.NewMqttCli(cfg)
}

type SmartPadData struct {
	rawData string
	devID   string
}

func NewSmartPadJob(rawData string, devID string) SmartPadData {
	return SmartPadData{
		rawData: rawData,
		devID:   devID,
	}
}

func (d SmartPadData) Handle() {
	ProcessJsonMsg(d.rawData, d.devID)
}

func initWlMqttCfg(cfg *dasMqtt.MqttCfg) (err error) {
	cfg.Url, err = log.Conf.GetString("mqtt2srv", "url")
	if err != nil {
		err = fmt.Errorf("get-mqtt2srv-url > %w", err)
		return
	}
	cfg.Username, err = log.Conf.GetString("mqtt2srv", "user")
	if err != nil {
		err = fmt.Errorf("get-mqtt2srv-user > %w", err)
		return
	}
	cfg.Passwd, err = log.Conf.GetString("mqtt2srv", "pwd")
	if err != nil {
		err = fmt.Errorf("get-mqtt2srv-pwd > %w", err)
		return
	}
	cfg.ClientId, err = log.Conf.GetString("mqtt2srv", "subcid")
	if err != nil {
		err = fmt.Errorf("get-mqtt2srv-cid > %w", err)
		return
	}
	cfg.ConnectHandler = subscribeDefaultTopic
	cfg.ResumeSubs = true
	cfg.CleanSession = false

	return
}

func subscribeDefaultTopic(client mqtt.Client) {
	msgTopic, err := log.Conf.GetString("mqtt2srv", "subtopic")
	if err != nil {
		return
	}
	msgTopic_pad, err := log.Conf.GetString("mqtt2srv", "subtopic-pad")
	if err != nil {
		return
	}
	msgTopic_test, err := log.Conf.GetString("mqtt2srv", "subtopic-test")
	if err != nil {
		return
	}

	log.Info("call subscribeDefaultTopic")
	// 订阅
	log.Info("mqtt Subscribe ", msgTopic)
	if token := client.Subscribe(msgTopic, 0, msgCallback); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 设备上线消息
	log.Info("mqtt Subscribe ", conTopic)
	if token := client.Subscribe(conTopic, 0, conCallback); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 设备下线消息
	log.Info("mqtt Subscribe ", disTopic)
	if token := client.Subscribe(disTopic, 0, disCallback); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 平板智能设备
	log.Info("mqtt Subscribe ", msgTopic_pad)
	if token := client.Subscribe(msgTopic_pad, 0, msgCallbackPad); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 测试
	log.Info("mqtt Subscribe ", msgTopic_test)
	if token := client.Subscribe(msgTopic_test, 0, msgCallback_test); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Error(token.Error())
	}

	log.Info("exit subscribeDefaultTopic")
}

// 上线下线事件。当某客户端上线时，会发布该消息；当某客户端离线时，会发布该消息
type ConDisEvent struct {
	Clientid string	`json:"clientid"`		//"clientid":"id1",
}

//订阅回调函数；收到消息后会执行它
var msgCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	//log.Debug("Mqtt-Topic: ", msg.Topic(), ", strHexMsg: ", hex.EncodeToString(msg.Payload()))

	//1. 解析包头
	var wlMsg wlprotocol.WlMessage
	_, err0 := wlMsg.PkDecode(msg.Payload())
	if err0 != nil {
		log.Error("mqtt.MessageHandler wlMsg.PkDecode, err0=", err0)
		return
	}

	rabbitmq.SendGraylogByMQ("Mqtt-Topic: %s, strHexMsg: %s", msg.Topic(), hex.EncodeToString(msg.Payload()))

	//2. 锁对接的平台，存入redis
	mymap := make(map[string]interface{})
	mymap["from"] = constant.MQTT_PLATFORM
	go redis.SetDevicePlatformPool(wlMsg.DevId.Uuid, mymap)

	//3. fetch job
	jobque.JobQueue <- NewMqttJob(msg.Payload())
}

//订阅回调函数；收到消息后会执行它
var msgCallbackPad mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Debug("Mqtt-Topic: ", msg.Topic(), ", strHexMsg: ", string(msg.Payload()))

	//1. 检验数据是否合法
	getData := string(msg.Payload())
	if !strings.Contains(getData, "#") {
		log.Error("msgCallbackPad mqtt.MessageHandler error msg: ", getData)
		return
	}

	//2. 获取设备编号
	prData := strings.Split(getData, "#")
	var devID string
	var devData string
	devID = prData[0]
	devData = prData[1]

	//3. 锁对接的平台，存入redis
	mymap := make(map[string]interface{})
	mymap["from"] = constant.MQTT_PAD_PLATFORM
	go redis.SetDevicePlatformPool(devID, mymap)

	//4. fetch job
	jobque.JobQueue <- NewSmartPadJob(devData, devID)
}

//TODO:JHHE 测试 订阅回调函数；收到消息后会执行它
var msgCallback_test mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	strMsg := string(msg.Payload())
	log.Debug("msgCallback_test Mqtt-Topic: ", msg.Topic(), ", strMsg: ", strMsg)
	if strings.Contains(strMsg, "\"") {
		nStart := strings.IndexAny(strMsg, "\"")
		nEnd := strings.LastIndexAny(strMsg, "\"")
		if -1 != nStart && -1 != nEnd {
			strMsg = strMsg[nStart+1:nEnd]
		}
	}

	log.Debug("msgCallback_test mymqtt.WlMqttPublish, ClientID: ", strMsg, ", strMsg: ", string(msg.Payload()))
	WlMqttPublish(strMsg, msg.Payload())
}

//订阅回调函数；设备上线消息 connected
var conCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Debug("conCallback Mqtt-Topic: ", msg.Topic(), ", strMsg: ", string(msg.Payload()))
	// TODO:JHHE WiFi锁去掉在线侦测
	/*var conMsg = msg.Payload()
	var conEvent ConDisEvent
	if err := json.Unmarshal(conMsg, &conEvent); err != nil {
		log.Error("mqtt.MessageHandler conCallback json.Unmarshal Header error, err=", err)
		return
	}

	//1. 锁状态，存入redis
	redis.SetActTimePool(conEvent.Clientid, 1)

	//2. 通知APP
	var devAct entity.DeviceActive
	devAct.Cmd = constant.Upload_lock_active
	devAct.Ack = 0
	devAct.DevType = ""
	devAct.DevId = conEvent.Clientid
	devAct.Vendor = ""
	devAct.SeqId = 0
	devAct.Signal = 0
	devAct.Time = 1
	if toApp_str, err := json.Marshal(devAct); err == nil {
		log.Info("[", conEvent.Clientid, "] mqtt.MessageHandler conCallback device connected, resp to APP, ", string(toApp_str))
		rabbitmq.Publish2app(toApp_str, devAct.DevId)
	} else {
		log.Error("[", conEvent.Clientid, "] mqtt.MessageHandler conCallback device connected, resp to APP, json.Marshal, err=", err)
	}*/
}

//订阅回调函数；设备下线消息 disconnected
var disCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Debug("disCallback Mqtt-Topic: ", msg.Topic(), ", strMsg: ", string(msg.Payload()))
	var disMsg = msg.Payload()
	var disEvent ConDisEvent
	if err := json.Unmarshal(disMsg, &disEvent); err != nil {
		log.Error("mqtt.MessageHandler disCallback json.Unmarshal Header error, err=", err)
		return
	}

	//1. 锁状态，存入redis
	go redis.SetActTimePool(disEvent.Clientid, 0)

	//2. 通知APP
	var devAct entity.DeviceActive
	devAct.Cmd = constant.Upload_lock_active
	devAct.Ack = 0
	devAct.DevType = ""
	devAct.DevId = disEvent.Clientid
	devAct.Vendor = ""
	devAct.SeqId = 0
	devAct.Signal = 0
	devAct.Time = 0
	if toApp_str, err := json.Marshal(devAct); err == nil {
		log.Info("[", disEvent.Clientid, "] mqtt.MessageHandler disCallback device disconnected, resp to APP, ", string(toApp_str))
		rabbitmq.Publish2app(toApp_str, devAct.DevId)
		rabbitmq.Publish2mns(toApp_str, "")
	} else {
		log.Error("[", disEvent.Clientid, "] mqtt.MessageHandler disCallback device disconnected, resp to APP, json.Marshal, err=", err)
	}
}