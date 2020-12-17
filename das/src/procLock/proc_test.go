package procLock

import (
	"das/core/log"
	"das/core/rabbitmq"
	"testing"
)

func init()  {
	log.Init()
	rabbitmq.Init()
}

var (
	userOpMsg = `
{"seqId":7,"cmd":119,"ack":0,"devType":"WL025S1","devId":"05382b4a73f575fa","vendor":"general","userType":1,"userId":3,"userId2":65535,"opType":0,"opUserPara":4,"opValue":1,"time":0}
`
)

func TestProcessJsonMsg(t *testing.T) {
	ProcessJsonMsg(userOpMsg, "05382b4a73f575fa")
}
