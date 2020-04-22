package httpgo

import (
	"das/core/log"
	"testing"
)

func init() {
	log.Init()
}

var (
	WonlyLGuard = `
{"ack":0,"bindid":"5233647","bindstr":"9799ee283c7721135f522bb27db32fda","cmd":251,"devId":"00158d0004623f11_01","devType":"WonlyLGuard","seqId":2,"value":"AB012301A930","vendor":"WonlyLGuard"}
`
)

func TestHttp2FeibeeWonlyLGuard(t *testing.T) {
	Http2FeibeeWonlyLGuard(WonlyLGuard)
}
