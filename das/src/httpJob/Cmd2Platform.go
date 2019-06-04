package httpJob

import (
	"../core/log"
	"../core/redis"
	"../core/httpgo"
	"errors"
)

func Cmd2Platform(imei string, data string, cmd string) error {
	if "" == imei {
		err := errors.New("imei is null")
		log.Error("Get Platform from redis failed, imei is null, err=", err)
		return err
	}
	platform, errPlat := redis.GetDevicePlatformPool(imei)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch platform {
	case "onenet":
		{
			httpgo.Http2OneNET_write(imei, data, cmd)
		}
	case "telecom":
		{
			httpgo.HttpCmd2DeviceTelecom(imei, data)
		}
	case "andlink":
		{

		}
	default:
		{
			log.Error("Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	return nil

}
