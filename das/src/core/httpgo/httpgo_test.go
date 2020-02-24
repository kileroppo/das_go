package httpgo

import (
	"das/core/log"
	"testing"
)

func init() {
	log.Init()
}

var (
	WonlyLGuard = `{"vendor":"00158d0003e8b2e3_01","bindid":"5233586","devId":"00158d0003e8b2e3_01","seqId":1,"devType":"WonlyLGuard","cmd":251,"bindstr":"fd4fdc69f43248ff4a6fb55833f5c386","ack":0,"value":"ab000e81e4"}`
)

func TestHttp2FeibeeWonlyLGuard(t *testing.T) {
	Http2FeibeeWonlyLGuard(WonlyLGuard)
}
