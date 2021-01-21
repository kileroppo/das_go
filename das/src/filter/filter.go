package filter

import (
	"reflect"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis"

	"das/core/redis"
)

var (
	_ MsgFilter = &RedisFilter{}
	alarmMsgFilter = RedisFilter{}
)

const (
	filterKey = "_msgFilter"
	priorityKey = "_priorityFilter"
	PriorityField = "prio"
	ValueField = "value"
	frequentDur = time.Second*30
	filterDur = time.Hour
)

type ValComparer func(int64, int64) bool

var NormalComparer ValComparer = func(newVal, oldVal int64) bool {
	return newVal > oldVal
}

type MsgFilter interface {
	IsPriorityValid(key string, priority int64, val interface{}, dur time.Duration, compare ValComparer) bool
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

func (r *RedisFilter) IsPriorityValid(key string, priority int64, val interface{}, dur time.Duration, compare ValComparer) bool {
	return r.isPriorityValid(key, priority, val, dur, compare)
}

func (r *RedisFilter) isPriorityValid(key string, priority int64, val interface{}, dur time.Duration, compare ValComparer) (res bool) {
	cli := redis.GetRedisDB(1)
	str,err := cli.HGet(key, PriorityField).Result()
	if len(str) == 0 || err != nil {
		r.setKey(cli, key, priority, val, dur)
		res = true
	} else {
		oldPrio,err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			res = true
			r.setKey(cli, key, priority, val, dur)
		} else {
			if compare(priority, oldPrio) {
				res = true
				r.setKey(cli, key, priority, val, dur)
			} else {
				res = false
			}
		}
	}
	return
}

func (r *RedisFilter) setKey(cli *goredis.Client, key string, priority int64, val interface{}, dur time.Duration) {
	fields := map[string]interface{} {
		ValueField: val,
		PriorityField: priority,
	}
	pipe := cli.TxPipeline()
	defer pipe.Close()
	pipe.HMSet(key, fields)
	pipe.Expire(key, dur)
	pipe.Exec()
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

func AlarmMsgFilter(key string, val interface{}, filterDur time.Duration) bool {
	res := alarmMsgFilter.IsValid(key + filterKey, val, filterDur)
	return res
}

func AlarmFrequentFilter(key string, val interface{}, filterDur, frequentDur time.Duration) bool {
	return alarmMsgFilter.IsFrequentValid(key + filterKey, val, filterDur, frequentDur)
}

func MsgNormalPriorityFilter(key string, priority int64, val interface{}, dur time.Duration) bool {
	return alarmMsgFilter.IsPriorityValid(key + priorityKey, priority, val, dur, NormalComparer)
}

func MsgPriorityFilter(key string, priority int64, val interface{}, dur time.Duration, compare ValComparer) bool {
	return alarmMsgFilter.IsPriorityValid(key + filterKey, priority, val, dur, compare)
}
