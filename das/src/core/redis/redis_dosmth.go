package redis

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strconv"
	"time"

	"das/core/log"
)

const (
	Dev_Router_Key = "Grayscale_Dev"
	User_Router_Key = "Grayscale_User"
)

func SetActTimePool(devId string, data int64) error {
	cli := GetRedisDB(0)
	cmd := cli.Set(devId, data, time.Second * 120) // 写入值120S后过期
	if nil != cmd.Err() {
		log.Error("redis SetActTimePool failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	return nil
}

func SetDevicePlatformPool(devId string, fields map[string]interface{}) error {
	cli := GetRedisDB(0)
	cmd := cli.HMSet(devId+"_platform", fields)
	if nil != cmd.Err() {
		log.Error("redis SetDevicePlatformHashPool failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	// 2592000 过期时间一个月
	cli.Expire(devId+"_platform", time.Second * 60 * 60 * 24 * 30)

	return nil
}

func GetDevicePlatformPool(devId string) (map[string]string, error) {
	cli := GetRedisDB(0)
	cmd := cli.HGetAll(devId+"_platform")
	if nil != cmd.Err() {
		log.Error("redis GetDevicePlatformPool failed, key=", devId, ", err=", cmd.Err())
		return nil, cmd.Err()
	}

	return cmd.Result()
}

func SetDeviceYisumaRandomfromPool(devId string, random string) error {
	cli := GetRedisDB(0)
	cmd := cli.Set(devId+"_yisumaRandom", random, time.Second * 1800) // 写入值30分钟后过期
	if nil != cmd.Err() {
		log.Error("redis SetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	return nil
}

func GetDeviceYisumaRandomfromPool(devId string) (string, error) {
	cli := GetRedisDB(0)
	cmd := cli.Get(devId+"_yisumaRandom")
	if nil != cmd.Err() {
		log.Error("redis GetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", cmd.Err())
		return "", cmd.Err()
	}

	return cmd.Result()
}

func SetDevUserNotePool(devId string, strTime string, userNote string) error {
	cli := GetRedisDB(0)
	cmd := cli.HSet(devId + "_usernote", strTime, userNote)
	if nil != cmd.Err() {
		log.Error("redis SetDevUserNotePool failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	// 过期时间10分钟
	cli.Expire(devId + "_usernote", time.Second * 60 * 10)

	return nil
}

func SetAppUserPool(devId string, strTime string, appUser string) error {
	cli := GetRedisDB(0)
	cmd := cli.HSet(devId + "_appuser", strTime, appUser)
	if nil != cmd.Err() {
		log.Error("redis SetAppUserPool failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	// 过期时间10分钟
	cli.Expire(devId + "_appuser", time.Second * 60 * 10)

	return nil
}

func SetAliIoTtoken(token string, expireTime int64) error {
	cli := GetRedisDB(0)
	cmd := cli.Set("ALI_IOT_TOKEN", token, time.Second * time.Duration(expireTime))
	if nil != cmd.Err() {
		log.Error("redis SetAliIoTtoken failed:", cmd.Err())
		return cmd.Err()
	}

	return nil
}

func GetAliIoTtoken() (string, error) {
	cli := GetRedisDB(0)
	cmd := cli.Get("ALI_IOT_TOKEN")
	if nil != cmd.Err() {
		log.Error("redis GetAliIoTtoken failed:", cmd.Err())
		return "", cmd.Err()
	}

	return cmd.Result()
}

func GetFbLockUserId(key string) (res int, err error) {
	cli := GetRedisDB(0)
	cmd := cli.Get(key)
	val,err := cmd.Result()
	if err != nil {
		err = fmt.Errorf("GetFbLockUserId > redisCli.Get > %w", err)
		return
	}
	v,err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		err = fmt.Errorf("GetFbLockUserId > redisCli.Get > %w", err)
		return
	}
	res = int(v)
	return
}

func SetFbLockUserId(key string , val interface{}) (err error){
	cli := GetRedisDB(0)
	cmd := cli.Set(key, val, time.Minute*1)
	if cmd.Err() != nil {
		err = fmt.Errorf("SetFbLockUserId > redisCli.Set > %w", cmd.Err())
		return
	}
	return nil
}

// 设置设备燃气告警状态
func SetDevGasAlarmState(devId string, data int64) error {
	cli := GetRedisDB(0)
	cmd := cli.Set(devId + "_gas", data, time.Second * 68) // 写入值68S后过期
	if nil != cmd.Err() {
		log.Error("redis SetDevGasAlarmState failed, key=", devId, ", err=", cmd.Err())
		return cmd.Err()
	}

	return nil
}

func IsFeibeeSpSrv(data []byte) (res bool) {
	hKey := gjson.GetBytes(data, "bindid").String() + "_platform"
	key := "from"
	cli := GetRedisDB(0)
	cmd := cli.HExists(hKey, key)
	if cmd.Err() != nil {
		return false
	}
	return cmd.Val()
}

func IsDevBeta(data []byte) (res bool) {
	hKey := gjson.GetBytes(data, "devId").String()
	cli := GetRedisDB(1)
	cmd := cli.HExists(Dev_Router_Key, hKey)

	if cmd.Err() != nil {
		return false
	}

	return cmd.Val()
}
