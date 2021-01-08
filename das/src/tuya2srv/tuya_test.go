package tuya2srv

import (
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/filter"
	"fmt"
	"testing"
)

func init() {
	log.Init()
	redis.InitRedis()
	rabbitmq.Init()
}

var (
	statusDemo = `
{
    "dataId": "AAW4ExXjL5RapTq7XxcAeA",
    "devId": "test",
    "productKey": "ncdapbwy",
    "status": [
        {"code":"va_temperature","103":"2390","t":1609817485969,"value":2390},
        {
            "1": "true",
            "code": "battery_percentage",
            "t": 1609766994653,
            "value": 18
        },
        {
            "code": "temper_alarm",
            "4": "true",
            "t": 1609766994653,
            "value": true
        }
    ]
}
`
	eventDemo = `
{"bizCode":"online",
"bizData":{"uid":"ay1601195946258gejAm","time":1603684405568},"devId":"6c445498a7d1e3a410goe1","productKey":"pisltm67","ts":1603684405584,"uuid":"6c445498a7d1e3a410goe1"}
`
	tyCleanerStatus = `
{
    "dataId": "0005b4e9b52ad64b173a411a2ce70085",
    "devId": "test1",
    "productKey": "imk0pcrtmyx9cbfg",
    "status": [
{"code":"status","103":"2820","t":1609324116494,"value":"pause"},
        {
            "code": "status",
            "t": 1609324116486,
            "value": "standby"
        }
    ]
}
`
	tyLight = `
{"dataId":"0005b51217c1d61a5aa5b0bb5f240102","devId":"6c445498a7d1e3a410goe1","productKey":"pisltm67","status":[{"2":"48","code":"bright_value","t":1606464196172,"value":48}]}
`
	tySleepBand = `
{
    "dataId": "AAW2Am5SjHFn9B4HBX7kAPC",
    "devId": "test0",
    "productKey": "SXfTziaaWHlCr0o2",
    "status": [
        {
            "111": "awake",
            "code": "status",
            "t": 1607496440646,
            "value": "sleep"
        },
        {
            "code": "off_bed",
            "t": 1607585001994,
            "108": "true",
            "value": true
        },
        {
            "code": "wakeup",
            "t": 1607585001994,
            "108": "true",
            "value": true
        }
    ]
}
`
	tyScene = `
{"dataId":"AAW2eGKXR8hapQO7X8cAcA","devId":"6cec485a673df8dc0dtaho","productKey":"jecjknhr","status":[{"1":"scene","code":"scene_1","t":1608003049965,"value":"scene"}]}
`
	tyEnv = `
{
    "dataId": "AAW4OEkpFRgARAftX0MHnB",
    "devId": "test1",
    "productKey": "q3jbhxhukfm7rkfy",
    "status": [
        {
            "102": 205,
            "code": "PM2_5",
            "t": 1609926768661,
            "value": 205
        },
        {
            "102": 10,
            "code": "CO2",
            "t": 1609926768661,
            "value": 10
        },
        {
            "102": 10,
            "code": "CH2O",
            "t": 1609926768661,
            "value": 100
        },
        {
            "102": 10,
            "code": "VOC",
            "t": 1609926768661,
            "value": 200
        }
    ]
}
`
)

func TestTuyaHandle(t *testing.T) {
	h := TuyaMsgHandle{
		data: []byte(statusDemo),
	}

	h.MsgHandle()
}

func TestGetEnvSensorLevel(t *testing.T) {
	datas := [][]string {
		{"CO2", "0.01"},
		{"CO2", "0.1"},
		{"CO2", "0.2"},
		{"CO2", "0.4"},
		{"CO2", "0.5"},

		{"PM2.5", "-1"},
		{"PM2.5", "101"},
		{"PM2.5", "151"},
		{"PM2.5", "161"},
		{"PM2.5", "201"},

		{"VOC", "0.59"},
		{"VOC", "0.7"},
	}

	for _,data := range datas {
		level, ok := filter.GetEnvSensorLevel(data[0], data[1])
		if ok {
			fmt.Printf("%s:%s(%s)\n", data[0], data[1], level)
		}
	}
}
