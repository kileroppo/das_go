package redis

import (
	"fmt"
	"time"

	"das/core/log"
)

func SetActTimePool(devId string, data int64) error {
	ret := redisCli.Set(devId, data, time.Second * 120) // 写入值120S后过期
	if nil != ret.Err() {
		log.Error("redis SetActTimePool failed, key=", devId, ", err=", ret.Err())
		return ret.Err()
	}

	return nil
}

func SetDevicePlatformPool(devId string, fields map[string]interface{}) error {
	ret := redisCli.HMSet(devId+"_platform", fields)
	if nil != ret.Err() {
		log.Error("redis SetDevicePlatformHashPool failed, key=", devId, ", err=", ret.Err())
		return ret.Err()
	}

	// 2592000 过期时间一个月
	redisCli.Expire(devId+"_platform", time.Second * 60 * 60 * 24 * 30)

	return nil
}

func GetDevicePlatformPool(devId string) (map[string]string, error) {
	mapOut := redisCli.HGetAll(devId+"_platform")
	if nil != mapOut.Err() {
		log.Error("redis GetDevicePlatformPool failed, key=", devId, ", err=", mapOut.Err())
		return nil, mapOut.Err()
	}

	mdev := mapOut.Val()

	return mdev, nil
}

func SetDeviceYisumaRandomfromPool(devId string, random string) error {
	ret := redisCli.Set(devId+"_yisumaRandom", random, time.Second * 1800) // 写入值30分钟后过期
	if nil != ret.Err() {
		log.Error("redis SetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", ret.Err())
		return ret.Err()
	}

	return nil
}

func GetDeviceYisumaRandomfromPool(devId string) (string, error) {
	ret := redisCli.Get(devId+"_yisumaRandom")
	if nil != ret.Err() {
		log.Error("redis GetDeviceYisumaRandomfromPool failed, key=", devId, ", err=", ret.Err())
		return "", ret.Err()
	}

	return ret.Val(), nil
}

func SetDevUserNotePool(devId string, strTime string, userNote string) error {
	ret := redisCli.HSet(devId + "_usernote", strTime, userNote)
	if nil != ret.Err() {
		log.Error("redis SetDevUserNotePool failed, key=", devId, ", err=", ret.Err())
		return ret.Err()
	}

	// 过期时间10分钟
	redisCli.Expire(devId + "_usernote", time.Second * 60 * 10)

	return nil
}

func SetAliIoTtoken(token string, expireTime int64) error {
	ret := redisCli.Set("ALI_IOT_TOKEN", token, time.Second * time.Duration(expireTime))
	if nil != ret.Err() {
		log.Error("redis SetAliIoTtoken failed:", ret.Err())
		return ret.Err()
	}

	return nil
}

func GetAliIoTtoken() (string, error) {
	ret := redisCli.Get("ALI_IOT_TOKEN")
	if nil != ret.Err() {
		log.Error("redis GetAliIoTtoken failed:", ret.Err())
		return "", ret.Err()
	}

	return ret.Val(), nil
}

func GetFbLockUserId(key string) (res int, err error) {
	res,err = redisCli.Get(key).Int()
	if err != nil {
		err = fmt.Errorf("GetFbLockUserId > redisCli.Get > %w", err)
		return
	} else {
		return
	}
}

func SetFbLockUserId(key string , val interface{}) (err error){
	cmd := redisCli.Set(key, val, time.Minute*1)
	if cmd.Err() != nil {
		err = fmt.Errorf("SetFbLockUserId > redisCli.Set > %w", cmd.Err())
		return
	}
	return
}

// 设置设备燃气告警状态
func SetDevGasAlarmState(devId string, data int64) error {
	ret := redisCli.Set(devId + "_gas", data, time.Second * 68) // 写入值68S后过期
	if nil != ret.Err() {
		log.Error("redis SetDevGasAlarmState failed, key=", devId, ", err=", ret.Err())
		return ret.Err()
	}

	return nil
}