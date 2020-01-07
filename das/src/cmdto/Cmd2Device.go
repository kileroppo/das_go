package cmdto

import (
	"errors"

	"das/core/constant"
	"das/core/httpgo"
	"das/core/log"
	"das/core/redis"
	"das/rmq/producer"
)

func Cmd2Device(uuid string, data string, cmd string) error {
	if "" == uuid {
		err := errors.New("uuid is null")
		log.Error("Get Platform from redis failed, uuid is null, err=", err)
		return err
	}
	platform, errPlat := redis.GetDevicePlatformPool(uuid)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch platform {
	case constant.ONENET_PLATFORM: {
			httpgo.Http2OneNET_write(uuid, data, cmd)
		}
	case constant.TELECOM_PLATFORM: {
			httpgo.HttpCmd2DeviceTelecom(uuid, data)
		}
	case constant.ANDLINK_PLATFORM: {}
	case constant.WIFI_PLATFORM: {
			producer.SendMQMsg2Device(uuid, data, cmd)
		}
	case constant.ALIIOT_PLATFORM: {	// 阿里云飞燕平台
			httpgo.HttpSetAliPro(uuid, data, cmd)
		}
	default: {
			log.Error("Cmd2Platform::Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	return nil
}
