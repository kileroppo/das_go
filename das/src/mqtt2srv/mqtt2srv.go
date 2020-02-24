package mqtt2srv

import (
	"das/core/jobque"
	"das/core/wlprotocol"
	"github.com/dlintw/goconf"
	"github.com/eclipse/paho.mqtt.golang"
	"das/core/constant"
	"das/core/log"
	"das/core/redis"
	"fmt"
	"time"
)

//订阅回调函数；收到消息后会执行它
var fcallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s--", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())

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

func MqttInit(conf *goconf.ConfigFile) (mqtt.Client) {
	url, err := conf.GetString("mqtt2srv", "url")
	if err != nil {
		log.Error("get-mqtt2srv-url error = ", err)
		return nil
	}
	user, err := conf.GetString("mqtt2srv", "user")
	if err != nil {
		log.Error("get-mqtt2srv-user error = ", err)
		return nil
	}
	pwd, err := conf.GetString("mqtt2srv", "pwd")
	if err != nil {
		log.Error("get-mqtt2srv-pwd error = ", err)
		return nil
	}
	cid, err := conf.GetString("mqtt2srv", "cid")
	if err != nil {
		log.Error("get-mqtt2srv-cid error = ", err)
		return nil
	}

	opts := mqtt.NewClientOptions().AddBroker(url)
	opts.SetUsername(user)
	opts.SetPassword(pwd)
	opts.SetClientID(cid)
	opts.SetKeepAlive(15 * time.Second)
	opts.SetDefaultPublishHandler(fcallback)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetCleanSession(true)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
		return c
	}

	// 订阅
	log.Info("mqtt Subscribe wonly/things/smartlock/srv\n")
	if token := c.Subscribe("wonly/things/smartlock/srv", 2, nil); token.Wait() && token.Error() != nil {
		log.Error(token.Error())
		return c
	}

	return c
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