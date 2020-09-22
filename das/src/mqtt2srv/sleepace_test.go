package mqtt2srv

import "testing"

var (
	sleepaceMsg = `
[{"timeStamp":1600753692,"dataKey":"inBedStatus","data":{"inbedStatus":1},"deviceId":"f7o6bn7jtlusr"}]
`
)

func TestSleepaceHandler(t *testing.T) {
    SleepaceHandler([]byte(sleepaceMsg))
}