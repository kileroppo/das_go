package filter

import (
	"reflect"
	"strconv"
	"time"

	"das/core/redis"
)

var (
	_ MsgFilter = &RedisFilter{}
	alarmMsgFilter = RedisFilter{}
)

const (
	filterKey = "_msgFilter"
	frequentDur = time.Second*30
	filterDur = time.Hour
)

type MsgFilter interface {
	IsValid(key string, val interface{}, duration time.Duration) (res bool)
	IsFrequentValid(key string, val interface{}, dur time.Duration, frequentDur time.Duration) bool
}

type RedisFilter struct{}

func (r *RedisFilter) IsValid(key string, val interface{}, duration time.Duration) (res bool) {
	res,_ = r.isValid(key, val, duration)
	return
}

func (r *RedisFilter) IsFrequentValid(key string, val interface{}, dur time.Duration, frequentDur time.Duration) bool {
	changed, ttl := r.isValid(key, val, dur)
	if changed {
		if ttl < 0 || ttl > dur {
			return true
		} else {
			since := dur - ttl
			if since < frequentDur {
				return false
			} else {
				return true
			}
		}
	} else {
		return false
	}
}

func (r *RedisFilter) isValid(key string, val interface{}, duration time.Duration) (res bool, ttl time.Duration) {
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
		ttl = cli.PTTL(key).Val()
		r.updKey(key, val, duration)
	}
	return
}

func (r *RedisFilter) GetDuration(key string) time.Duration {
	cli := redis.GetRedisDB(1)
	durCmd := cli.PTTL(key)
	if durCmd.Err() != nil {
		return -1
	} else {
		return durCmd.Val()
	}
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

func AlarmMsgFilter(key string, val interface{}, filterDur time.Duration) bool {
	res := alarmMsgFilter.IsValid(key + filterKey, val, filterDur)
	return res
}

func AlarmFrequentFilter(key string, val interface{}, filterDur, frequentDur time.Duration) bool {
	return alarmMsgFilter.IsFrequentValid(key + filterKey, val, filterDur, frequentDur)
}
