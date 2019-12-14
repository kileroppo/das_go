package feibee2srv

import "testing"

var pushData = `
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

func TestProcessFeibeeMsg(t *testing.T) {
	if ProcessFeibeeMsg([]byte(pushData)) != nil {
		t.Error("数据解析错误")
	}
}

func BenchmarkFeibeeProc(b *testing.B) {
	ProcessFeibeeMsg([]byte(pushData))
}
