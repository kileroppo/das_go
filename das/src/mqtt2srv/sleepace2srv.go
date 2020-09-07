package mqtt2srv

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tidwall/gjson"

	"das/core/log"
	dasMqtt "das/core/mqtt"
	"das/core/entity"
	"das/core/rabbitmq"
)

var (
	sleepaceMqttCli mqtt.Client
)

const (
	Sleepace_Event_SleepStage = "sleepStageEvent"
)

func InitMqtt2Srv() {
	cfg := &dasMqtt.MqttCfg{}
	if err := initSleepaceMqttCfg(cfg); err != nil {
		log.Errorf("InitSleepaceMqtt > %s", err)
		panic(err)
	}

	sleepaceMqttCli = dasMqtt.NewMqttCli(cfg)
}

func initSleepaceMqttCfg(cfg *dasMqtt.MqttCfg) (err error) {
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
	rabbitmq.SendGraylogByMQ("Receive from sleepace: %s", oriData)
	msgTyp := gjson.GetBytes(oriData, "dataKey").String()

	switch msgTyp {
	case Sleepace_Event_SleepStage:
		sendSleepStageForSceneTrigger(oriData)
	}
}

func sendSleepStageForSceneTrigger(oriData []byte) {
	msg2pms := entity.Feibee2AutoSceneMsg{
		Header:      entity.Header{
			Cmd:     241,
			Ack:     0,
			DevType: "",
			DevId:   gjson.GetBytes(oriData, "deviceId").String(),
			Vendor:  "sleepace",
			SeqId:   0,
		},
		Time:        int(gjson.GetBytes(oriData, "timeStamp").Int()),
		TriggerType: 0,
		AlarmFlag:   int(gjson.GetBytes(oriData, "data").Get("sleepStage").Int()),
		AlarmType:   "sleepStage",
		AlarmValue:  "",
		SceneId:     "",
		Zone:        "",
	}

	data,err := json.Marshal(msg2pms)
	if err != nil {
		log.Errorf("sendSleepStageForSceneTrigger > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}

func CloseMqtt2Srv() {
	dasMqtt.CloseMqttCli(sleepaceMqttCli)
}