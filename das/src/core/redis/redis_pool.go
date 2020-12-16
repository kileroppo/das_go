package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis"
	"reflect"

	"das/core/log"
)

var (
	ErrConn = errors.New("DB Connection failed")
	ErrRedisPipeExec = errors.New("redis pipe exec failed")

	redisDevPool0 *RedisPool
	redisDevPool1 *RedisPool
)

type RedisPool struct {
	cli      *redis.Client
	redisUrl string

	maxPoolSize int

	ctxP       context.Context
	cancelP    context.CancelFunc
}

func GetRedisDB(dbIndex int) *redis.Client {
	switch dbIndex {
	case 0:
		return redisDevPool0.GetCli()
	case 1:
		return redisDevPool1.GetCli()
	default:
		return redisDevPool0.GetCli()
	}
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

	redisDevPool0 = newRedisPool(uri, maxActive, 0)
	redisDevPool1 = newRedisPool(uri, maxActive, 1)
}

func newRedisPool(redisUrl string, maxPoolSize int, defaultDB int) *RedisPool {
	cli := newRedisClient(redisUrl, maxPoolSize, defaultDB)
	if cli == nil {
		panic(ErrConn)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisPool{
		cli:         cli,
		redisUrl:    redisUrl,
		maxPoolSize: maxPoolSize,
		ctxP:        ctx,
		cancelP:     cancel,
	}
}

func (self *RedisPool) GetCli() (*redis.Client) {
	return self.cli
}

func (self *RedisPool) Close() {
	if !isNil(self.cli) {
		self.cli.Close()
	}
	self.cancelP()
}

func newRedisClient(url string, maxPoolSize int, defaultDB int) *redis.Client {
	cli := redis.NewClient(
		&redis.Options{
			Addr:     url,
			Password: "",
			DB:       defaultDB,
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
	redisDevPool0.Close()
}
