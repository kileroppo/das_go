package httpgo

import (
	"errors"

	"../log"
	"../redis"
	"../../mq/producer"
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
			Http2OneNET_write(imei, data, cmd)
		}
	case "telecom":
		{
			HttpCmd2DeviceTelecom(imei, data)
		}
	case "andlink":
		{

		}
	case "wifi":
		{
			producer.SendMQMsg2Device(imei, data, cmd)
		}
	default:
		{
			log.Error("Cmd2Platform::Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	return nil

}
