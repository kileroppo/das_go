package redis

import (
	"fmt"
	"github.com/dlintw/goconf"
	"github.com/go-redis/redis"
	"github.com/tidwall/gjson"
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
	redisSrv, _ = conf.GetString("redisPool", "redis_server")
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

func IsFeibeeSpSrv(data []byte) (res bool) {
	hKey := gjson.GetBytes(data, "bindid").String() + "_platform"
	key := "from"
	var err error

	if redisCli == nil {
		return false
	}

	res, err = redisCli.HExists(hKey, key).Result()
	if err != nil {
		return false
	}
	return res
}