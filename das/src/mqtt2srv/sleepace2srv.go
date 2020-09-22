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
	Sleepace_Data_Key_Sleep_Stage  = "sleepStageEvent"
	Sleepace_Data_Key_Inbed_Status = "inbedStatus"

	Sleepace_Data_Field_Sleep_Stage  = "sleepStage"
	Sleepace_Data_Field_Inbed_Status = "inbedStatus"
)

func Init() {
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
	SleepaceHandler(msg.Payload())
}

func SleepaceHandler(rawData []byte) {
	rabbitmq.SendGraylogByMQ("Receive from sleepaceServer: %s", rawData)

	if !gjson.ValidBytes(rawData) {
		log.Error("sleepaceMsg invalid json")
		return
	}

	msgs := gjson.ParseBytes(rawData).Array()
	for i := range msgs {
		msgTyp := msgs[i].Get("dataKey").String()
		switch msgTyp {
		case Sleepace_Data_Key_Sleep_Stage, Sleepace_Data_Key_Inbed_Status:
			sendSleepStageForSceneTrigger(msgTyp, msgs[i])
		}
	}
}

func sendSleepStageForSceneTrigger(msgTyp string, oriData gjson.Result) {
	alarmType := "sleepStatus"
	devId := oriData.Get("deviceId").String()
	switch msgTyp {
	case Sleepace_Data_Key_Sleep_Stage:
		alarmFlag := int(oriData.Get("data").Get(Sleepace_Data_Field_Sleep_Stage).Int())
		sendSceneTrigger(devId, alarmType, alarmFlag)
	case Sleepace_Data_Key_Inbed_Status:
		alarmFlag := int(oriData.Get("data").Get(Sleepace_Data_Field_Inbed_Status).Int())
		sendSceneTrigger(devId, alarmType, alarmFlag+5)
	}
}

func sendSceneTrigger(devId, alarmType string, alarmFlag int) {
	msg2pms := entity.Feibee2AutoSceneMsg{
		Header:      entity.Header{
			Cmd:     241,
			Ack:     0,
			DevType: "",
			DevId:   devId,
			Vendor:  "",
			SeqId:   0,
		},
		Time:        0,
		TriggerType: 0,
		AlarmFlag:   alarmFlag,
		AlarmType:   alarmType,
		AlarmValue:  "",
		SceneId:     "",
		Zone:        "",
	}
	data,err := json.Marshal(msg2pms)
	if err != nil {
		log.Errorf("sendSceneTrigger > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}

func Close() {
	dasMqtt.CloseMqttCli(sleepaceMqttCli)
}