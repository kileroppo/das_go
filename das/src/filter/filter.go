package filter

import (
	"reflect"
	"strconv"
	"sync"
	"time"

	"das/core/redis"
)

var (
	_ MsgFilter = &SyncMapFilter{}
	_ MsgFilter = &RedisFilter{}
	alarmMsgFilter = RedisFilter{}
)

const (
	filterKey = "_msgFilter"
	filterDur = time.Duration(time.Hour)
)

type MsgFilter interface {
	IsValid(key string, val interface{}, duration time.Duration) (res bool)
}

type SyncMapFilter struct {
	m sync.Map
}

func (s *SyncMapFilter) IsValid(key string, val interface{}, duration time.Duration) (res bool) {
	_, res = s.m.LoadOrStore(key, val)
	return !res
}

type RedisFilter struct{}

func (r *RedisFilter) IsValid(key string, val interface{}, duration time.Duration) (res bool) {
	return r.isValid(key, val, duration)
}

func (r *RedisFilter) isValid(key string, val interface{}, duration time.Duration) (res bool) {
	cli := redis.GetRedisDB(1)
	oldVal, err := cli.Get(key).Result()
	if err != nil {
		res = true
	}
	if oldVal == "" {
		res = true
	} else {
		if redisValCompare(val, oldVal) {
			res = false
		} else {
			res = true
		}
	}
	if res {
		r.updKey(key, val, duration)
	}
	return
}

func redisValCompare(newVal interface{}, oldVal string) bool {
	valType := reflect.TypeOf(newVal)
	switch valType.Kind() {
	case reflect.Int64:
		spcNew := newVal.(int64)
		spcOld, err := strconv.ParseInt(oldVal, 10, 64)
		if err != nil {
			return false
		} else {
			return spcNew == spcOld
		}
	case reflect.Int:
		spcNew := newVal.(int)
		spcOld, err := strconv.ParseInt(oldVal, 10, 64)
		if err != nil {
			return false
		} else {
			return int64(spcNew) == spcOld
		}
	case reflect.String:
		spcNew := newVal.(string)
		return oldVal == spcNew
	case reflect.Bool:
		spcNew := newVal.(bool)
		spcOld := true
		if oldVal == "0" {
			spcOld = false
		}
		return spcNew == spcOld
	default:
		return false
	}
}

func (r *RedisFilter) updKey(key string, val interface{}, duration time.Duration) {
	cli := redis.GetRedisDB(1)
	cli.Set(key, val, duration)
}

func AlarmMsgFilter(key string, val interface{}) bool {
	res := alarmMsgFilter.IsValid(key + filterKey, val, filterDur)
	return res
}
