package tuya2srv

import (
	"strconv"
	"sync"
	"time"

	"das/core/constant"
	"das/core/log"
	"das/core/mysql"
	"das/filter"
)

var (
	mu = sync.Mutex{}
	sqlQueryFilterRules = `
SELECT CODE 
FROM
	ty_notify_filter_rule
`
)

func TyFilter(devId, sensorType, sensorVal string, val interface{}) (notifyFlag bool, triggerFlag bool) {
	notifyFlag, triggerFlag = true, false
	if !tyAlarmMsgFilter(devId, sensorType, sensorVal, val) {
		if sensorType == constant.Wonly_Status_Sensor_Infrared {
			notifyFlag = false
			triggerFlag = true
		} else {
			if _,ok := tyEnvAlarmDataFilterMap[sensorType]; ok {
				notifyFlag = true
			} else {
				notifyFlag = false
			}
		}
	} else {
		notifyFlag, triggerFlag = true, true
	}
	return
}

func tyAlarmMsgFilter(devId, code, sensorVal string, val interface{}) bool {
	if _,ok := tyAlarmDataFilterMap[code]; ok {
		return filter.AlarmMsgFilter(devId + "_" + code, val, -1)
	} else {
		level,ok := GetEnvSensorLevel(code, sensorVal)
		if ok {
			return filter.AlarmMsgFilter(devId + "_" + level, val, -1)
		}
		return true
	}
}

func tyStatusPriorityFilter(devId string, timestamp int64, status string) bool {
	return filter.MsgPriorityFilter(devId, timestamp, status, time.Minute*1)
}

func loadFilterRulesFromMySql() {
	rows,err := mysql.DoMysqlQuery(sqlQueryFilterRules)
	if err != nil {
		log.Errorf("loadFilterRulesFromMySql > %s", err)
		return
	}
	code := ""
	mu.Lock()
	defer mu.Unlock()
	for rows.Next() {
		if err := rows.Scan(&code); err == nil {
			TyStatusDataFilterMap[code] = struct{}{}
		}
	}
	log.Info("load filter rules done")
}

func GetEnvSensorLevel(sensorType, sensorVal string) (levelVal string, ok bool)  {
	_, ok = ReflectGrade[sensorType]
	if ok {
		levelVal = envSensorLevelTable(sensorType, sensorVal)
	} else {
		levelVal = envSensorVOCLevelTable(sensorType, sensorVal)
	}
	if len(levelVal) == 0 {
		ok = false
	} else {
		ok = true
	}
	return
}

func envSensorLevelTable(sensorType, sensorVal string) (levelVal string) {
	rangeLimit,ok := ReflectGrade[sensorType]
	if !ok {
		return
	}
	val,err := strconv.ParseFloat(sensorVal, 64)
	if err != nil {
		return
	}

	level := 0
	for i := 0; i < len(rangeLimit); i++ {
		if val < rangeLimit[i] {
			level++
			break
		} else {
			level++
			if i == 1 && val == rangeLimit[i] {
				break
			}
		}
	}
	return GradeEnv[level-1]
}

func envSensorVOCLevelTable(sensorType, sensorVal string) (levelVal string) {
	limit,ok := ReflectLimit[sensorType]
	if !ok {
		return
	}
	val,err := strconv.ParseFloat(sensorVal, 64)
	if err != nil {
		return
	}
	if val <= limit {
		return GradeEnv[0]
	} else {
		return GradeEnv[1]
	}
}
