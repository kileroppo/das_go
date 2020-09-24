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
	_,err := RedisDevPool.Set(0, devId, data, time.Second * 120) // 写入值120S后过期
	if nil != err {
		log.Error("redis SetActTimePool failed, key=", devId, ", err=", err)
		return err
	}

	return nil
}

func SetDevicePlatformPool(devId string, fields map[string]interface{}) error {
	_,err := RedisDevPool.HMSet(0, devId+"_platform", fields)
	if nil != err {
		log.Error("redis SetDevicePlatformHashPool failed, key=", devId, ", err=", err)
		return err
	}

	// 2592000 过期时间一个月
	RedisDevPool.Expire(0, devId+"_platform", time.Second * 60 * 60 * 24 * 30)

	return nil
}

func GetDevicePlatformPool(devId string) (map[string]string, error) {
	mdev,err := RedisDevPool.HGetAll(0, devId+"_platform")
	if nil != err {
		log.Error("redis GetDevicePlatformPool failed, key=", devId, ", err=", err)
		return nil, err
	}

	return mdev, nil
}

func SetDeviceYisumaRandomfromPool(devId string, random string) error {
	_,err := RedisDevPool.Set(0, devId+"_yisumaRandom", random, time.Second * 1800) // 写入值30分钟后过期
	if nil != err {
		log.Error("redis SetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", err)
		return err
	}

	return nil
}

func GetDeviceYisumaRandomfromPool(devId string) (string, error) {
	res,err := RedisDevPool.Get(0, devId+"_yisumaRandom")
	if nil != err {
		log.Error("redis GetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", err)
		return "", err
	}

	return res, nil
}

func SetDevUserNotePool(devId string, strTime string, userNote string) error {
	_,err := RedisDevPool.HSet(0, devId + "_usernote", strTime, userNote)
	if nil != err {
		log.Error("redis SetDevUserNotePool failed, key=", devId, ", err=", err)
		return err
	}

	// 过期时间10分钟
	RedisDevPool.Expire(0, devId + "_usernote", time.Second * 60 * 10)

	return nil
}

func SetAppUserPool(devId string, strTime string, appUser string) error {
	_,err := RedisDevPool.HSet(0, devId + "_appuser", strTime, appUser)
	if nil != err {
		log.Error("redis SetAppUserPool failed, key=", devId, ", err=", err)
		return err
	}

	// 过期时间10分钟
	RedisDevPool.Expire(0, devId + "_appuser", time.Second * 60 * 10)

	return nil
}

func SetAliIoTtoken(token string, expireTime int64) error {
	_,err := RedisDevPool.Set(0,"ALI_IOT_TOKEN", token, time.Second * time.Duration(expireTime))
	if nil != err {
		log.Error("redis SetAliIoTtoken failed:", err)
		return err
	}

	return nil
}

func GetAliIoTtoken() (string, error) {
	val,err := RedisDevPool.Get(0,"ALI_IOT_TOKEN")
	if nil != err {
		log.Error("redis GetAliIoTtoken failed:", err)
		return "", err
	}

	return val, nil
}

func GetFbLockUserId(key string) (res int, err error) {
	val,err := RedisDevPool.Get(0, key)
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
	_,err = RedisDevPool.Set(0, key, val, time.Minute*1)
	if err != nil {
		err = fmt.Errorf("SetFbLockUserId > redisCli.Set > %w", err)
		return
	}
	return
}

// 设置设备燃气告警状态
func SetDevGasAlarmState(devId string, data int64) error {
	_,err := RedisDevPool.Set(0, devId + "_gas", data, time.Second * 68) // 写入值68S后过期
	if nil != err {
		log.Error("redis SetDevGasAlarmState failed, key=", devId, ", err=", err)
		return err
	}

	return nil
}

func IsFeibeeSpSrv(data []byte) (res bool) {
	hKey := gjson.GetBytes(data, "bindid").String() + "_platform"
	key := "from"
	var err error

	if RedisDevPool == nil {
		return false
	}

	res, err = RedisDevPool.HExists(0, hKey, key)
	if err != nil {
		return false
	}
	return res
}

func IsDevBeta(data []byte) (res bool) {
	hKey := gjson.GetBytes(data, "devId").String()
	var err error
	res,err = RedisDevPool.HExists(1, Dev_Router_Key, hKey)

	if err != nil {
		return false
	}

	return
}