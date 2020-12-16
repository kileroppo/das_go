package mqtt2srv

import (
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
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
