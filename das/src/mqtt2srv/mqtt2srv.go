package mqtt2srv

import (
	"das/core/constant"
	"das/core/jobque"
	"das/core/log"
	"das/core/redis"
	"das/core/wlprotocol"
	"github.com/dlintw/goconf"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)
var (
	mqttcli mqtt.Client
	strTopic string
)

//订阅回调函数；收到消息后会执行它
var fcallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Debug("topic: ", msg.Topic(), ", msg: ", msg.Payload())

	//1. 检验数据是否合法
	var wlMsg wlprotocol.WlMessage
	_, err0 := wlMsg.PkDecode(msg.Payload())
	if err0 != nil {
		log.Error("mqtt.MessageHandler wlMsg.PkDecode, err0=", err0)
		return
	}

	//3. 锁对接的平台，存入redis
	mymap := make(map[string]interface{})
	mymap["from"] = constant.MQTT_PLATFORM
	redis.SetDevicePlatformPool(wlMsg.DevId.Uuid, mymap)

	//4. fetch job
	jobque.JobQueue <- NewMqttJob(msg.Payload())
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
	strTopic, err = conf.GetString("mqtt2srv", "subtopic")
	if err != nil {
		log.Error("get-mqtt2srv-subtopic error = ", err)
		return
	}

	opts := mqtt.NewClientOptions().AddBroker(url)
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetClientID(cid)
	opts.SetKeepAlive(15 * time.Second)
	opts.SetDefaultPublishHandler(fcallback)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetCleanSession(true)

	mqttcli = mqtt.NewClient(opts)
	if token := mqttcli.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

	// 订阅
	log.Info("mqtt Subscribe ", strTopic)
	if token := mqttcli.Subscribe(strTopic, 2, nil); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
	}

}

// 释放
func MqttRelease() {
	// 取消订阅
	log.Debug("mqtt Unsubscribe")
	if token := mqttcli.Unsubscribe(strTopic); token.Wait() && token.Error() != nil {
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
	// TODO:jhhe 增加wifi锁数据的处理
	log.Debug(o.rawData)
}