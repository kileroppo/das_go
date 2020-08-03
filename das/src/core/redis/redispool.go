package redis

import (
	"fmt"
	"github.com/dlintw/goconf"
	"github.com/go-redis/redis"
	"time"
)

var (
	redisCli *redis.Client
	redisSrv string
	redisPwd string
	maxIdle int
	maxActive int
	idleTimeout int
)

//初始化redis连接池
func InitRedisPool(conf *goconf.ConfigFile) {
	redisSrv, _ = conf.GetString("redisPool", "redis_uri_dev")
	if redisSrv == "" {
		fmt.Println("未启用redis")
		return
	}
	redisPwd, _ = conf.GetString("redisPool", "redis_password")
	maxIdle, _ = conf.GetInt("redisPool", "maxIdle")
	maxActive, _ = conf.GetInt("redisPool", "maxActive")
	idleTimeout, _ = conf.GetInt("redisPool", "idleTimeout")

	redisCli = newRedisCliPool()
}

func newRedisCliPool() *redis.Client {
	redisdb := redis.NewClient(
		&redis.Options{
			Addr:         redisSrv,
			Password:     redisPwd, // no password set
			DB:           0,  // use default DB
			IdleTimeout:  time.Duration(idleTimeout) * time.Second,
			MinIdleConns: maxIdle,
			PoolSize:     maxActive,
		})

	pong, err := redisdb.Ping().Result()
	if err != nil {
		fmt.Println(pong, err)
		return nil
	}

	return redisdb
}

func GetRedisCli() *redis.Client {
	return redisCli
}

func CloseRedisCli() {
	if nil != redisCli {
		redisCli.Close()
	}
}