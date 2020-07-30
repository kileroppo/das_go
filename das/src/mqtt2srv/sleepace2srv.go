package mqtt2srv

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"

	"das/core/log"
	dasMqtt "das/core/mqtt"
)

var (
	sleepaceMqttCli mqtt.Client
)

func InitMqtt2Srv() {
	cfg := &dasMqtt.MqttCfg{}
	if err := initMqttCfg(cfg); err != nil {
		log.Errorf("InitSleepaceMqtt > %s", err)
		panic(err)
	}

	sleepaceMqttCli = dasMqtt.NewMqttCli(cfg)
}

func initMqttCfg(cfg *dasMqtt.MqttCfg) (err error) {
	cfg.Url, err = log.Conf.GetString("sleepace", "url")
	if err != nil {
		err = fmt.Errorf("get-sleepaceMqtt-url > %w", err)
		return
	}
	cfg.Username, err = log.Conf.GetString("sleepace", "account")
	if err != nil {
		err = fmt.Errorf("get-sleepaceMqtt-user > %w", err)
		return
	}
	cfg.Passwd, err = log.Conf.GetString("sleepace", "password")
	if err != nil {
		err = fmt.Errorf("get-sleepaceMqtt-pwd > %w", err)
		return
	}

	cfg.ConnectHandler = subscribeSleepaceTopic
	cfg.ResumeSubs = true
	cfg.CleanSession = false

	return nil
}

func subscribeSleepaceTopic(cli mqtt.Client) {
	topic,err := log.Conf.GetString("sleepace", "topic")
	if err != nil {
		return
	}

	if token := cli.Subscribe(topic, 0, sleepaceCallback); token.WaitTimeout(time.Second*3) && token.Error() != nil {
		log.Errorf("subscribeSleepaceTopic > %s", token.Error())
	} else {
		log.Infof("sleepaceMqtt Subscribe Topic: %s", topic)
	}
}

func sleepaceCallback(client mqtt.Client, msg mqtt.Message) {
	oriData := msg.Payload()
	log.Info("Receive from sleepace: ", oriData)
}

func CloseMqtt2Srv() {
	dasMqtt.CloseMqttCli(sleepaceMqttCli)
}