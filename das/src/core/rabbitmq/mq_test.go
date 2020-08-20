package rabbitmq

import (
	"das/core/log"
	"das/core/redis"
	"encoding/json"
	"fmt"
	"testing"
)

var (
	rawData = `count %d`
	alarmData = `
{
    "cmd": 252,
    "ack": 1,
    "devType": "",
    "sceneId": "1591005962970cc95",
    "devId": "%d",
    "vendor": "general",
    "seqId": 2,
    "triggerT": 1,
    "alarmType": "infrared",
    "alarmFlag": 0,
    "alarmValue": "有人经过",
    "time": 666666
}
`
)

func init() {
	log.Init()
	redis.InitRedis()
	Init()
}

func TestPublish(t *testing.T) {
	for i:=0;i<100;i++ {
		data := []byte(fmt.Sprintf(rawData, i))
		Publish2pms(data, "")
		Publish2app(data, "test")
		Publish2dev(data, "test")
		Publish2mns(data, "")
		Publish2wonlyms(data, "")
	}
}

type GraylogEntity struct {
	Version string `json:"version"`
	Host string `json:"host"`
	Short_message string `json:"short_message"`
}

func sendJson(data string) {
    en := GraylogEntity{
		Version:       "3.0",
		Host:          "DAS",
		Short_message: data,
	}

	json.Marshal(en)
}

func sendRaw(data string) {
    fmt.Sprintf(formatLog, data)
}

func BenchmarkSendJson(b *testing.B) {
	for i:=0;i<b.N;i++ {
		sendJson(alarmData)
	}
}

func BenchmarkSendRaw(b *testing.B) {
	for i:=0;i<b.N;i++ {
		sendRaw(alarmData)
	}
}