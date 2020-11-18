package tuya2srv

import (
	"das/core/redis"
	"sync"
	"time"
)

var (
	_ MsgFilter = &SyncMapFilter{}
	_ MsgFilter = &RedisFilter{}
)

type MsgFilter interface {
	Set(key string, duration time.Duration)
	Exists(key string) bool
}

type SyncMapFilter struct {
	m sync.Map
}

func (s *SyncMapFilter) Set(key string, duration time.Duration) {
	s.m.Store(key, nil)
}

func (s *SyncMapFilter) Exists(key string) bool {
    return false
}

type RedisFilter struct {}

func (r *RedisFilter) Set(key string, duration time.Duration) {
	_,_ = redis.RedisDevPool.Set(4, key, nil, duration)
}

func (r *RedisFilter) Exists(key string) bool {
    res,err := redis.RedisDevPool.Exists(4, key)
    if res > 0 && err != nil {
    	return true
	}
	return false
}