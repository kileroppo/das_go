package tuya2srv

import (
	"time"

	"das/core/log"
	"das/core/mysql"
	"das/filter"
)

var (
	sqlQueryFilterRules = `
SELECT CODE 
FROM
	ty_notify_filter_rule
`
)

func tyAlarmMsgFilter(devId, code string, val interface{}) bool {
	if _,ok := tyAlarmDataFilterMap[code]; ok {
		res := filter.AlarmMsgFilter(devId + "_" + code, val, time.Hour*1)
		return res
	} else {
		return true
	}
}

func loadFilterRulesFromMySql() {
	rows,err := mysql.DoMysqlQuery(sqlQueryFilterRules)
	if err != nil {
		log.Errorf("loadFilterRulesFromMySql > %s", err)
		return
	}
	code := ""
	for rows.Next() {
		if err := rows.Scan(&code); err == nil {
			TyStatusDataFilterMap[code] = struct{}{}
		}
	}
	log.Info("load filter rules done")
}
