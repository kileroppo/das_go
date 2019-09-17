package redis

import (
	"fmt"
	"github.com/dlintw/goconf"
	"github.com/garyburd/redigo/redis"
)

var (
	redisServer_s   string
	redisPassword_s string
)

//初始化redis连接池
func InitRedisSingle(conf *goconf.ConfigFile) {
	redisServer_s, _ = conf.GetString("redisPool", "redis_server")
	if redisServer_s == "" {
		fmt.Println("未启用redis")
		return
	}
	redisPassword_s, _ = conf.GetString("redisPool", "redis_password")
}

func SetActTime(devId string, time int64) {
	c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}

	defer c.Close()

	// 写入值60S后过期
	_, err = c.Do("SET", devId, time, "EX", "120")
	if err != nil {
		fmt.Println("redis set failed:", err)
	}
}

func SetDevicePlatform(devId string, platform string) {
	c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}

	defer c.Close()

	// 写入值60S后过期
	_, err = c.Do("SET", devId+"_platform", platform, "EX", "2592000")

	if err != nil {
		fmt.Println("redis set failed:", err)
	}
}

func GetDevicePlatform(devId string) (string, error) {
	c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return "", err
	}

	defer c.Close()

	var retPlat string
	retPlat, err = redis.String(c.Do("GET", devId+"_platform"))
	if err != nil {
		fmt.Println("redis get failed:", err)
		return "", err
	}

	return retPlat, nil
}
