package mqtt2srv

import (
	"das/core/constant"
	"das/core/entity"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/wlprotocol"
	"das/procwlpro"
	"encoding/hex"
	"encoding/json"
	"github.com/dlintw/goconf"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)
var (
	mqttcli mqtt.Client
	msgTopic string
	// cdTopic string = "$SYS/brokers/+/clients/#"
	conTopic = "$SYS/brokers/+/clients/+/connected"
	disTopic = "$SYS/brokers/+/clients/+/disconnected"
)

// 上线事件。当某客户端上线时，会发布该消息
type ConnectedEvent struct {
	Clientid string	`json:"clientid"`		//"clientid":"id1",
	Username string `json:"username"`		//"username":"u",
	Ipaddress string `json:"ipaddress"`		//"ipaddress":"127.0.0.1",
	Connack string	`json:"connack"`		//"connack":0,
	Ts int32 `json:"ts"`					//"ts":1554047291,
	Proto_ver int32	`json:"proto_ver"`		//"proto_ver":3,
	Proto_name string `json:"proto_name"`	//"proto_name":"MQIsdp",
	Clean_start bool `json:"clean_start"`	//"clean_start":true,
	Keepalive int32	`json:"keepalive"`		//"keepalive":60
}

// 下线事件。当某客户端离线时，会发布该消息
type DisconnectedEvent struct {
	Clientid string	`json:"clientid"`		//"clientid":"id1",
	Username string	`json:"username"`		//"username":"u",
	Reason string `json:"reason"`			//"reason":"normal",
	Ts int32 `json:"ts"`					//"ts":1554047291
}

//订阅回调函数；收到消息后会执行它
var msgCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Debug("Mqtt-Topic: ", msg.Topic(), ", strHexMsg: ", hex.EncodeToString(msg.Payload()))

	//1. 解析包头
	var wlMsg wlprotocol.WlMessage
	_, err0 := wlMsg.PkDecode(msg.Payload())
	if err0 != nil {
		log.Error("mqtt.MessageHandler wlMsg.PkDecode, err0=", err0)
		return
	}

	//2. 锁对接的平台，存入redis
	mymap := make(map[string]interface{})
	mymap["from"] = constant.MQTT_PLATFORM
	redis.SetDevicePlatformPool(wlMsg.DevId.Uuid, mymap)

	//3. fetch job
	jobque.JobQueue <- NewMqttJob(msg.Payload())
}

//订阅回调函数；设备上线消息 connected
var conCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var conMsg = msg.Payload()
	var conEvent ConnectedEvent
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
	}
}

//订阅回调函数；设备下线消息 disconnected
var disCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var disMsg = msg.Payload()
	var disEvent DisconnectedEvent
	if err := json.Unmarshal(disMsg, &disEvent); err != nil {
		log.Error("mqtt.MessageHandler disCallback json.Unmarshal Header error, err=", err)
		return
	}

	//1. 锁状态，存入redis
	redis.SetActTimePool(disEvent.Clientid, 0)

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
		log.Info("[", disEvent.Clientid, "] mqtt.MessageHandler conCallback device disconnected, resp to APP, ", string(toApp_str))
		rabbitmq.Publish2app(toApp_str, devAct.DevId)
	} else {
		log.Error("[", disEvent.Clientid, "] mqtt.MessageHandler conCallback device disconnected, resp to APP, json.Marshal, err=", err)
	}
}

func MqttInit(conf *goconf.ConfigFile) {
	url, err := conf.GetString("mqtt2srv", "url")
	if err != nil {
		log.Error("get-mqtt2srv-url error = ", err)
		return
	}
	user, err := conf.GetString("mqtt2srv", "user")
	if err != nil {
		log.Error("get-mqtt2srv-user error = ", err)
		return
	}
	pwd, err := conf.GetString("mqtt2srv", "pwd")
	if err != nil {
		log.Error("get-mqtt2srv-pwd error = ", err)
		return
	}
	cid, err := conf.GetString("mqtt2srv", "cid")
	if err != nil {
		log.Error("get-mqtt2srv-cid error = ", err)
		return
	}
	msgTopic, err = conf.GetString("mqtt2srv", "subtopic")
	if err != nil {
		log.Error("get-mqtt2srv-subtopic error = ", err)
		return
	}

	opts := mqtt.NewClientOptions().AddBroker(url)
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetClientID(cid)
	opts.SetKeepAlive(15 * time.Second)
	opts.SetDefaultPublishHandler(nil)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetCleanSession(true)

	mqttcli = mqtt.NewClient(opts)
	if token := mqttcli.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅
	log.Info("mqtt Subscribe ", msgTopic)
	if token := mqttcli.Subscribe(msgTopic, 0, msgCallback); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 设备上线消息
	log.Info("mqtt Subscribe ", conTopic)
	if token := mqttcli.Subscribe(conTopic, 0, conCallback); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅 设备下线消息
	log.Info("mqtt Subscribe ", disTopic)
	if token := mqttcli.Subscribe(disTopic, 0, disCallback); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}
}

func GetMqttClient() mqtt.Client {
	return mqttcli
}
// 释放
func MqttRelease() {
	// 取消订阅
	log.Debug("mqtt Unsubscribe ", msgTopic)
	if token := mqttcli.Unsubscribe(msgTopic); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 取消订阅
	log.Debug("mqtt Unsubscribe ", conTopic)
	if token := mqttcli.Unsubscribe(conTopic); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 取消订阅
	log.Debug("mqtt Unsubscribe ", disTopic)
	if token := mqttcli.Unsubscribe(disTopic); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 关闭链接
	mqttcli.Disconnect(250)
}

type MqttJob struct {
	rawData []byte
}

func NewMqttJob(rawData []byte) MqttJob {
	return MqttJob{
		rawData: rawData,
	}
}

func (o MqttJob) Handle() {
	procwlpro.ParseData(o.rawData)
}