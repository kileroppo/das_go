package mqtt

import (
	"das/core/log"
	"github.com/dlintw/goconf"
	"github.com/eclipse/paho.mqtt.golang"
	"time"
)

var (
	mqttcli mqtt.Client
	topic2Dev string = "wonly/things/smartlock/"
)

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
}

func WlMqttPublish(uuid string, data []byte) error {
	if token := mqttcli.Publish(topic2Dev + uuid, 0, false, data); token.Wait() && token.Error() != nil {
		log.Error("WlMqttPublish failed, err: ", token.Error())
		return token.Error()
	}

	return nil
}

// 释放
func MqttRelease() {
	// 关闭链接
	mqttcli.Disconnect(250)
}