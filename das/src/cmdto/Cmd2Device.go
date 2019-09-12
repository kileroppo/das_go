package cmdto

import (
	"../core/constant"
	"../core/httpgo"
	"../core/log"
	"../core/redis"
	"../rmq/producer"
	"errors"
)

func Cmd2Device(imei string, data string, cmd string) error {
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
	case constant.ONENET_PLATFORM:
		{
			httpgo.Http2OneNET_write(imei, data, cmd)
		}
	case constant.TELECOM_PLATFORM:
		{
			httpgo.HttpCmd2DeviceTelecom(imei, data)
		}
	case constant.ANDLINK_PLATFORM:
		{

		}
	case constant.WIFI_PLATFORM:
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
