package feibee2srv

import (
	"das/core/log"
	"das/core/rabbitmq"
	"testing"
)

var (
	newOnline = `
{
    "code": 4,
    "status": "newonline",
    "ver": "2",
    "msg": [
        {
            "bindid": "3229",
            "deviceuid": 91169,
            "devicetype": "0x4020015",
            "uuid": "00158d0001da911a_01",
            "deviceid": 1026,
            "online": 3,
            "zonetype": 21,
            "pushstring": "烟雾传感器 上线"
        }
    ]
}`

	devDelete = `
{
    "code": 5,
    "status": "devdelete",
    "ver": "2",
    "msg": [
        {
            "bindid": "112036",
            "deviceuid": 109344,
            "devicetype": "0x4020015",
            "uuid": "00158d0001dd866a_01",
            "deviceid": 1026,
            "zonetype": 21,
            "pushstring": "未知设备 已离网"
        }
    ]
}
`
	devDegree = `
{
    "code":10,
    "status":"devbrightness",
    "ver":"2",
    "msg":[
        {
            "bindid":"21295",
            "deviceuid":759601,
            "devicetype":"0x2020002",
            "uuid": "00124b00180f8789_01",
            "zonetype": 3,
            "deviceid": 514,
            "brightness":138
        }
    ]
}
`

	devReName = `
{
    "code": 12,
    "status": "devnewname",
    "ver": "2",
    "msg": [
        {
            "bindid": "225259",
            "name": "123",
            "deviceuid": 97965,
            "devicetype": "0x20003",
            "uuid": "00158d000205e00f_01",
            "deviceid": 2,
            "zonetype": 3
        }
    ]
}
`

	manualOp = `
{
    "code": 2,
    "status": "report",
    "ver": "2",
    "records": [
        {
            "bindid": "11**26",
            "deviceuid": 71572,
            "cid": 6,
            "aid": 0,
            "value": "00",
            "orgdata": "700a17940106000100002000",
            "zonetype": 3,
            "deviceid": 2,
            "uuid": "00124b00180f8789_10",
            "devicetype": "0x20003",
            "pushstring": "",
            "uptime": 1535539330530
        }
    ]
}
`

	remoteOp = `
{
    "code": 7,
    "status": "devstate",
    "ver": "2",
    "msg": [
        {
            "bindid": "3229",
            "deviceuid": 771198,
            "devicetype": "0x20003",
            "uuid": "00158d0001dd7d27_0b",
            "deviceid": 2,
            "zonetype": 3,
            "onoff": 0
        }
    ]
}
`

	devAdd = `
{
    "code": 3,
    "status": "newdevice",
    "ver": "2",
    "msg": [
        {
            "bindid": "112e5291",
            "name": "",
            "deviceuid": 97965,
            "snid": "FNB56-DOS07FB2.7",
            "devicetype": "8888",
            "uuid": "3",
            "profileid": 260,
            "deviceid": 8888,
            "onoff": 0,
            "online": 3,
            "battery": 40,
            "lastvalue": 24,
            "zonetype": 21,
            "IEEE": "00158d000205e00f"
        }
    ]
}
`
	sceneAdd = `
{
  "code": 21,
  "status": "addscene",
  "ver": "2",
  "sceneMessages": [
    {
      "scenes": [
        {
          "sceneName": "abcdd2",
          "sceneID": 2,
          "sceneMembers": [
            {
              "deviceuid": 737925,
              "deviceID": 528,
              "data1": 1,
              "data2": 45,
              "data3": 45,
              "data4": 45,
              "IRID": 0,
              "delaytime": 0,
              "sceneFunctionID": 0,
              "uuid": "00124b0014afcbbf_0b"
            }
          ]
        }
      ],
      "bindid": "2201129"
    }
  ]
}
`
	sceneRename = `
{
  "code": 23,
  "status": "scenename",
  "ver": "2",
  "sceneMessages": [
    {
      "scenes": [
        {
          "sceneName": "abcd34",
          "sceneID": 1
        }
      ]
    }
  ]
}
`

	sceneDel = `
{
  "code": 22,
  "status": "removescene",
  "ver": "2",
  "sceneMessages": [
    {
      "scenes": [
        {
          "sceneName": "abcdd2",
          "sceneID": 1,
          "sceneMembers": [
            {
              "deviceuid": 1236014,
              "deviceID": 2,
              "data1": 0,
              "data2": 0,
              "data3": 0,
              "data4": 0,
              "IRID": 0,
              "delaytime": 0,
              "sceneFunctionID": 0,
              "uuid": "00124b00092a5de7_12"
            }
          ]
        }
      ]
    }
  ]
}
`
	lGuardScene = `
{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":0,"aid":16652,"value":"06ab01230141e7","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FNB54-HWD01WL0.7","devicetype":"0x30b0001","deviceid":779,"zonetype":1,"uptime":1581527097545}]}
`
	zigbeeMsg = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":0,"aid":16652,"value":"06ab01230141e7","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56-DOR07WL2.4","devicetype":"0x30b0001","deviceid":79,"zonetype":1,"uptime":1581527097545}]}
`
	sensorBatt = `
{
    "code": 2,
    "status": "report",
    "ver": "2",
    "records": [{
        "bindid": "112126",
        "deviceuid": 109189,
        "devicetype": "0x4020015",
        "uuid": "00158d000205e00f_01",
        "deviceid": 1026,
        "zonetype": 21,
        "cid": 1,
        "aid": 33,
        "value": "c8",
        "orgdata": "701585aa010100032000201e210020c83e001b00000000",
        "pushstring": "",
        "uptime": 1536847293202
    }, {
        "bindid": "112126",
        "deviceuid": 109189,
        "devicetype": "0x4020015",
        "uuid": "00158d000205e00f_01",
        "deviceid": 1026,
        "zonetype": 21,
        "cid": 1,
        "aid": 62,
        "value": "01000000",
        "orgdata": "701585aa010100032000201e210020c83e001b00000000",
        "pushstring": "",
        "uptime": 1536847293203
    }]
}
`
	devMatch = `
{
    "code": 2,
    "status": "report",
    "ver": "2",
    "records": [
        {
            "bindid": "225259",
            "deviceuid": 109319,
            "cid": 0,
            "aid": 16394,
            "value": "0e55550ae20701040000810002027d",
            "orgdata": "701807ab010000010a40420e55550ae20701040000810002027d",
            "pushstring": "",
            "uuid": "00158d0001dd866a_01",
            "devicetype": "0x16300ff",
            "deviceid": 355,
            "zonetype": 1,
            "uptime": 1540891316963
        }
    ]
}
`
	devTrain = `
{
    "code": 2,
    "status": "report",
    "ver": "2",
    "records": [
        {
            "bindid": "225259",
            "deviceuid": 109319,
            "cid": 0,
            "aid": 16394,
            "value": "0f55550be2070104000082000202007f",
            "orgdata": "701907ab010000010a40420f55550be2070104000082000202007f",
            "pushstring": "",
            "uuid": "00158d0001dd866a_01",
            "devicetype": "0x16300ff",
            "deviceid": 355,
            "zonetype": 1,
            "uptime": 1540892484262
        }
    ]
}
`
	sensorTemp = `
{
    "code": 2,
    "status": "report",
    "ver": "2",
    "records": [
        {
            "bindid": "29900",
            "deviceuid": 128499,
            "deviceid": 770,
            "zonetype": 1,
            "cid": 1026,
            "aid": 0,
            "value": "200B",
            "orgdata": "700B411401020401000029200B",
            "pushstring": "",
            "uptime": 1528486262762
        },
         {
            "bindid": "29900",
            "deviceuid": 128499,
            "cid": 1029,
            "aid": 0,
            "deviceid": 770,
            "zonetype": 1,
            "value": "E716",
            "orgdata": "700B411401050401000021E716",
            "pushstring": "",
            "uptime": 1528486262762
        }
    ]
}
`
	sensorBattVol = `
{
  "code": 2,
  "status": "report",
  "ver": "2",
  "records": [
    {
      "bindid": "38745",
      "deviceuid": 108400,
      "cid": 1,
      "aid": 33,
      "value": "c8",
      "orgdata": "701570a70101000320002020210020c83e001b00000000",
      "pushstring": "",
      "uuid": "00158d0001da90ae_01",
      "devicetype": "0x402002b",
      "deviceid": 11,
      "zonetype": 40,
      "uptime": 1559029207334
    },
    {
      "bindid": "38745",
      "deviceuid": 108400,
      "cid": 1,
      "aid": 62,
      "value": "01000000",
      "orgdata": "701570a70101000320002020210020c83e001b00000000",
      "pushstring": "",
      "uuid": "00158d0001da90ae_01",
      "devicetype": "0x402002b",
      "deviceid": 1026,
      "zonetype": 43,
      "uptime": 1559029207334
    }
  ]
}
`
)

func init() {
	conf := log.Init()
	rabbitmq.Init(conf)
}

func TestProcessFeibeeMsg(t *testing.T) {
	var tests = []struct {
		msgName  string
		msgValue string
	}{
		{"devAdd", devAdd},
		{"devDegree", devDegree},
		{"sceneAdd", sceneAdd},
		{"sceneDel", sceneDel},
		{"sceneRename", sceneRename},
		{"devDel", devDelete},
		{"devRename", devReName},
		{"newOnline", newOnline},
		{"manualOp", manualOp},
		{"remoteOp", remoteOp},
	}

	for _, ts := range tests {
		if ProcessFeibeeMsg([]byte(ts.msgValue)) != nil {
			t.Errorf("Process %s error", ts.msgName)
		}
	}
}

func TestPrcessOneMsg(t *testing.T) {
	ProcessFeibeeMsg([]byte(smoke))
}

func BenchmarkFeibeeProc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProcessFeibeeMsg([]byte(sceneAdd))
	}
}
