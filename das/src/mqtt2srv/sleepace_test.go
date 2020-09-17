package mqtt2srv

import "testing"

var (
	sleepaceMsg = `
[
    {
        "dataKey": "sleepStageEvent",
        "deviceId": "002s7dbcssyxc",
        "data": {
            "sleepStage": 1
        },
        "timeStamp": 1502069838
    },
    {
        "dataKey": "sleepStageEvent",
        "deviceId": "0010y81hgof88",
        "data": {
            "sleepStage": 3
        },
        "timeStamp": 1502069356
    }
]
`
)

func TestSleepaceHandler(t *testing.T) {
    SleepaceHandler([]byte(sleepaceMsg))
}