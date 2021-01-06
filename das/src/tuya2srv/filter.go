package tuya2srv

import (
	"strconv"
	"sync"
	"time"

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

func tyAlarmMsgFilter(devId, code, sensorVal string, val interface{}) bool {
	if _,ok := tyAlarmDataFilterMap[code]; ok {
		res := filter.AlarmMsgFilter(devId + "_" + code, val, -1)
		return res
	} else {
		if _,ok := tyEnvAlarmDataFilterMap[code]; ok {
			level := GetEnvSensorLevel(code, sensorVal)
			res := filter.AlarmMsgFilter(devId + "_" + level, val, -1)
			return res
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

func GetEnvSensorLevel(sensorType, sensorVal string) (levelVal string)  {
	_, ok := ReflectGrade[sensorType]
	if ok {
		return envSensorLevelTable(sensorType, sensorVal)
	} else {
		return envSensorVOCLevelTable(sensorType, sensorVal)
	}
}

func envSensorLevelTable(sensorType, sensorVal string) (levelVal string) {
	levelVal = "-1"
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
	levelVal = "-1"
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
