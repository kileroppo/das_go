package redis

import (
	"fmt"
	"github.com/dlintw/goconf"
	"github.com/garyburd/redigo/redis"
	goredis "github.com/go-redis/redis"
	"time"
)

var (
	RedisClient   *redis.Pool
	redisServer   string
	redisPassword string
	maxIdle       int
	maxActive     int
	idleTimeout   int
)

func NewPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func NewPool_goredis(server, password string) *goredis.Client {
	client := goredis.NewClient(
		&goredis.Options{
			Addr:         redisServer,
			Password:     "", // no password set
			DB:           0,  // use default DB
			IdleTimeout:  time.Duration(idleTimeout) * time.Second,
			MinIdleConns: maxIdle,
			PoolSize:     maxActive,
		})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	return client
}

//初始化redis连接池
func InitRedisPool(conf *goconf.ConfigFile) {
	redisServer, _ = conf.GetString("redisPool", "redis_server")
	if redisServer == "" {
		fmt.Println("未启用redis")
		return
	}
	redisPassword, _ = conf.GetString("redisPool", "redis_password")
	maxIdle, _ = conf.GetInt("redisPool", "maxIdle")
	maxActive, _ = conf.GetInt("redisPool", "maxActive")
	idleTimeout, _ = conf.GetInt("redisPool", "idleTimeout")

	RedisClient = NewPool(redisServer, redisPassword)
}

func RedisString(reply interface{}, err1 error) (value string, err2 error) {
	value, err2 = redis.String(reply, err1)
	return
}

func SetActTimePool(devId string, time int64) {
	/*c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}*/
	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	// 写入值60S后过期
	_, err := rc.Do("SET", devId, time, "EX", "120")
	if err != nil {
		fmt.Println("redis set failed:", err)
	}
}

func SetDevicePlatformPool(devId string, platform string) {
	/*c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}

	defer c.Close()*/
	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	// 写入值60S后过期
	_, err := rc.Do("SET", devId+"_platform", platform, "EX", "2592000")

	if err != nil {
		fmt.Println("redis set failed:", err)
	}
}

func GetDevicePlatformPool(devId string) (string, error) {
	/*c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return "", err
	}

	defer c.Close()*/

	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	var retPlat string
	retPlat, err := redis.String(rc.Do("GET", devId+"_platform"))
	if err != nil {
		fmt.Println("redis get failed:", err)
		return "", err
	}

	return retPlat, nil
}

func SetDeviceYisumaRandomfromPool(devId string, random string) {
	/*c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}

	defer c.Close()*/
	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	// 写入值30分钟过期
	_, err := rc.Do("SET", devId+"_yisumaRandom", random, "EX", "1800")

	if err != nil {
		fmt.Println("redis set failed:", err)
	}
}

func GetDeviceYisumaRandomfromPool(devId string) (string, error) {

	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	var retPlat string
	retPlat, err := redis.String(rc.Do("GET", devId+"_yisumaRandom"))
	if err != nil {
		fmt.Println("redis get failed:", err)
		return "", err
	}

	return retPlat, nil
}

func SetDevUserNotePool(devId string, strTime string, userNote string) error {
	/*c, err := redis.Dial("tcp", redisServer_s)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}*/
	// 从池里获取连接
	rc := RedisClient.Get()

	// 用完后将连接放回连接池
	defer rc.Close()

	// 写入值120S后过期
	_, err := rc.Do("HSET", devId + "_usernote", strTime, userNote)
	if err != nil {
		fmt.Println("redis HSET failed:", err)
		return err
	}
	_, err = rc.Do("EXPIRE", devId + "_usernote", 600)
	if err != nil {
		fmt.Println("redis EXPIRE failed:", err)
	}

	return nil
}