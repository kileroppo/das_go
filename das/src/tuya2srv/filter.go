package tuya2srv

import (
	"sync"
	"time"

	"das/core/log"
	"das/core/mysql"
	"das/filter"
)

const (
	LowBatteryThreshold = 30
)

var (
	mu = sync.Mutex{}
	sqlQueryFilterRules = `
SELECT CODE 
FROM
	ty_notify_filter_rule
`
)

func tyStatusPriorityFilter(devId string, timestamp int64, status string) bool {
	return filter.MsgNormalPriorityFilter(devId, timestamp, status, time.Minute*1)
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
