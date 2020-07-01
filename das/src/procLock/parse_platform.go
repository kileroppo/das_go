package procLock

import (
	"das/core/constant"
	"das/core/httpgo"
	"das/core/log"
	"das/core/mqtt"
	"das/core/rabbitmq"
	"das/core/redis"
	"encoding/hex"
	"errors"
)

func Cmd2Device(uuid string, mydata interface{}, cmd string) error {
	if "" == uuid {
		err := errors.New("uuid is null")
		log.Error("Get Platform from redis failed, uuid is null, err=", err)
		return err
	}
	ret, errPlat := redis.GetDevicePlatformPool(uuid)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch ret["from"] {
	case constant.ONENET_PLATFORM: {
		data, ok := mydata.(string)
		if ok {
			httpgo.Http2OneNET_write(uuid, data, cmd)
		}
	}
	//case constant.TELECOM_PLATFORM: {
	//		httpgo.HttpCmd2DeviceTelecom(uuid, data)
	//	}
	case constant.ANDLINK_PLATFORM: {}
	case constant.PAD_DEVICE_PLATFORM: {
		data, ok := mydata.(string)
		if ok {
			SendMQMsg2Device(uuid, data, cmd)
		}
	}
	case constant.MQTT_PAD_PLATFORM: { // WiFi平板，MQTT通道
		data, ok := mydata.(string)
		if ok {
			log.Debug("[", uuid, "] Cmd2Device resp to device, WlMqttPublishPad ", data)
			mqtt.WlMqttPublishPad(uuid, data)
		}
	}
	case constant.ALIIOT_PLATFORM: {	// 阿里云飞燕平台
		data, ok := mydata.(string)
		if ok {
			httpgo.HttpSetAliPro(uuid, data, cmd)
		}
	}
	case constant.FEIBEE_PLATFORM: {	//飞比zigbee锁
		//TODO:JHHE 不回复设备
		log.Debug("Cmd2Platform::constant.FEIBEE_PLATFORM not reply to device")
	}
	case constant.MQTT_PLATFORM: {		// MQTT
		data, ok := mydata.([]byte)
		if ok {
			log.Debug("mqtt.WlMqttPublish, data: ", hex.EncodeToString(data))
			mqtt.WlMqttPublish(uuid, data)
		}
	}
	default: {
			log.Error("Cmd2Platform::Unknow Platform from redis, please check the platform: ", ret)
		}
	}

	return nil
}

func SendMQMsg2Device(uuid string, message string, cmd string) {
	var rkey string
	rkey = uuid + "_robot"
	log.Info("[ ", rkey, " ] "+cmd+" rabbitmq.ProducerRabbitMq2Device.Publish2Device: ", message)
	rabbitmq.Publish2dev([]byte(message), rkey)
}