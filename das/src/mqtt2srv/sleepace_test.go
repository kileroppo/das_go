package mqtt2srv

import (
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"fmt"
	"testing"
)

func init() {
	log.Init()
	redis.InitRedis()
	rabbitmq.Init()
}

var (
	sleepaceMsg = `
[{"timeStamp":1600753692,"dataKey":"inBedStatus","data":{"inbedStatus":2},"deviceId":"f7o6bn7jtlusr"}]
`
)

func TestSleepaceHandler(t *testing.T) {
    SleepaceHandler([]byte(sleepaceMsg))
}

func TestSleepStageMsgFilter(t *testing.T) {
	if sleepStageMsgFilter("f7o6bn7jtlusr", "sleepStatus", 6) {
		fmt.Println("valid")
	} else {
		fmt.Println("invalid")
	}
}
