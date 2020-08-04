package redis

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/go-redis/redis"

	"das/core/log"
)

var (
	ErrConn = errors.New("DB Connection failed")
	ErrRedisPipeExec = errors.New("redis pipe exec failed")

	RedisDevPool  *RedisPool
)

type RedisPool struct {
	cli      *redis.Client
	redisUrl string

	maxPoolSize int

	ctxP       context.Context
	cancelP    context.CancelFunc
}

func InitRedis() {
    uri, err := log.Conf.GetString("redisPool", "redis_uri_dev")
    if err != nil {
    	panic(err)
	}
	maxActive, err :=  log.Conf.GetInt("redisPool", "maxActive")
	if err != nil {
		panic(err)
	}

	RedisDevPool = newRedisPool(uri, maxActive)
}

func newRedisPool(redisUrl string, maxPoolSize int) *RedisPool {
	mongoCli := newRedisClient(redisUrl, maxPoolSize)
	if mongoCli == nil {
		panic(ErrConn)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisPool{
		cli:         mongoCli,
		redisUrl:    redisUrl,
		maxPoolSize: maxPoolSize,
		ctxP:        ctx,
		cancelP:     cancel,
	}
}

func (self *RedisPool) GetCli() (*redis.Client) {
	return self.cli
}

func (self *RedisPool) Get(dbIndex int, key string) (res string, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.Get(key)
	cmder,err := pipe.Exec()
	if err != nil {
		return "", err
	}
	if len(cmder) != 2 {
		return "", ErrRedisPipeExec
	}
	strCmder,ok := cmder[1].(*redis.StringCmd)
	if !ok {
		return "", ErrRedisPipeExec
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) Set(dbIndex int, key string, value interface{}, expiration time.Duration) (res string, err error) {
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.Set(key, value, expiration)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.StatusCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HGet(dbIndex int, key,field string) (res string, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HGet(key, field)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.StringCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HExists(dbIndex int, key, field string) (res bool, err error) {
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HExists(key, field)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.BoolCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HSet(dbIndex int, key,field string, value interface{}) (res bool, err error) {
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HSet(key, field, value)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.BoolCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HVals(dbIndex int, key string) (res []string, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HVals(key)
	cmder,err := pipe.Exec()
	if err != nil {
		return nil, err
	}
	if len(cmder) != 2 {
		return nil, ErrRedisPipeExec
	}
	strCmder,ok := cmder[1].(*redis.StringSliceCmd)
	if !ok {
		return nil, ErrRedisPipeExec
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HDel(dbIndex int, key string, fields ...string) (res int64, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HDel(key, fields...)
	cmder,err := pipe.Exec()
	if err != nil {
		return 0, err
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.IntCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HMSet(dbIndex int, key string, fields map[string]interface{}) (res string, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HMSet(key, fields)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.StatusCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HGetAll(dbIndex int, key string) (res map[string]string, err error) {
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HGetAll(key)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.StringStringMapCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) HMGet(dbIndex int, key string, fields ...string) (res []interface{}, err error) {
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.HMGet(key, fields...)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.SliceCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) Expire(dbIndex int, key string, expiration time.Duration) (res bool, err error){
	pipe := self.cli.Pipeline()
	pipe.Select(dbIndex)
	pipe.Expire(key, expiration)
	cmder,errR := pipe.Exec()
	if errR != nil {
		err = errR
		return
	}
	if len(cmder) != 2 {
		err = ErrRedisPipeExec
		return
	}
	strCmder,ok := cmder[1].(*redis.BoolCmd)
	if !ok {
		err = ErrRedisPipeExec
		return
	}
	res, err = strCmder.Result()
	return
}

func (self *RedisPool) Close() {
	if !isNil(self.cli) {
		self.cli.Close()
	}
	self.cancelP()
}

func newRedisClient(url string, maxPoolSize int) *redis.Client {
	cli := redis.NewClient(
		&redis.Options{
			Addr:     url,
			Password: "",
			DB:       0,
			PoolSize: maxPoolSize,
		},
	)

	_, err := cli.Ping().Result()
	if err != nil {
		log.Error("redis NewClient error = ", err)
		return nil
	}

	return cli
}

func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}

func Close() {
	RedisDevPool.Close()
}
