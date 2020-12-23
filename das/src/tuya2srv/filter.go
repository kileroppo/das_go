package tuya2srv

import (
	"time"

	"das/filter"
)

func tyAlarmMsgFilter(devId, code string, val interface{}) bool {
	if _,ok := tyAlarmDataFilterMap[code]; ok {
		res := filter.AlarmMsgFilter(devId + "_" + code, val, time.Hour*1)
		return res
	} else {
		return true
	}
}

