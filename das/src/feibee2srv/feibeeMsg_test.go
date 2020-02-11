package feibee2srv

import "testing"

var (
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
)

func TestProcessFeibeeMsg(t *testing.T) {
	var tests = []struct {
		msgName string
		msgValue string
	}{
		{"devAdd", devAdd},
		{"sceneAdd", sceneAdd},
		{"sceneDel", sceneDel},
		{"sceneRename", sceneRename},
	}

	for _,ts := range tests {
		if ProcessFeibeeMsg([]byte(ts.msgValue)) != nil {
			t.Errorf("Process %s error", ts.msgName)
		}
	}
}

func BenchmarkFeibeeProc(b *testing.B) {
	for i:=0;i<b.N;i++ {
		ProcessFeibeeMsg([]byte(sceneAdd))
	}
}

