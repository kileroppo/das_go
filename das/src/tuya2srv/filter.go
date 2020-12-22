package tuya2srv

import "das/filter"

func tyAlarmMsgFilter(devId, code string, val interface{}) bool {
	if _,ok := tyAlarmDataFilterMap[code]; ok {
		res := filter.AlarmMsgFilter(devId + code, val, -1)
		return res
	} else {
		return true
	}
}

