package httpgo

import (
	"das/core/log"
	"das/core/redis"
	"testing"
)

func init() {
	log.Init()
	redis.InitRedisPool(log.Conf)
}

var (
	WonlyLGuard = `
{"ack":0,"bindid":"5233647","bindstr":"9799ee283c7721135f522bb27db32fda","cmd":251,"devId":"00158d0004623f11_01","devType":"WonlyLGuard","seqId":2,"value":"AB012301A930","vendor":"WonlyLGuard"}
`
	appZigbeeLock = `
{"passwd":"111111","passwd2":"","time":1588909445,"ack":0,"cmd":82,"devId":"00158d0004623f11_01","devType":"WlZigbeeLock","seqId":3,"vendor":"feibee", "bindid":"5233647", "bindstr": "9799ee283c7721135f522bb27db32fda"}
`
)

func TestHttp2FeibeeWonlyLGuard(t *testing.T) {
	Http2FeibeeWonlyLGuard(WonlyLGuard)
}

func TestHttp2FeibeeZigbeeLock(t *testing.T) {
	Http2FeibeeZigbeeLock(appZigbeeLock, "5233647", "9799ee283c7721135f522bb27db32fda", "00158d0004623f11_01", "")
}
