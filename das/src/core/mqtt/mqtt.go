package mqtt

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"

	"das/core/log"
)

type MqttCfg struct {
	Username       string
	Passwd         string
	Url            string
	ClientId       string
	CleanSession   bool
	ResumeSubs     bool
	ConnectHandler mqtt.OnConnectHandler
}

func GetUuid(cid string) string {
	uid := cid + uuid.New().String()
	log.Info("Get MQTT ClientId: ", uid)
	return uid
}

func NewMqttCli(cfg *MqttCfg) mqtt.Client {
	opts := mqtt.NewClientOptions().AddBroker(cfg.Url)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Passwd)
	opts.SetClientID(GetUuid(cfg.ClientId))
	opts.SetKeepAlive(30 * time.Second)
	opts.SetDefaultPublishHandler(nil)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(cfg.CleanSession)
	opts.SetResumeSubs(cfg.ResumeSubs)
	opts.SetOnConnectHandler(cfg.ConnectHandler)

	cli := mqtt.NewClient(opts)
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return cli
}

func CloseMqttCli(cli mqtt.Client) {
	cli.Disconnect(250)
}
