package procLock

import (
	"bytes"
	"das/core/entity"
	"das/core/util"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strings"

	"das/core/constant"
	"das/core/httpgo"
	"das/core/log"
	"das/core/mqtt"
	"das/core/rabbitmq"
	"das/core/redis"
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
	case constant.MQTT_PAD_PLATFORM: // WiFi平板，MQTT通道
	{
		data, ok := mydata.(string)
		if ok {
			var strToDevData string
			var err error

			// 加密数据
			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = constant.SERVICE_TYPE

			myKey := util.MD52Bytes(uuid)
			if strToDevData, err = util.ECBEncrypt([]byte(data), myKey); err == nil {
				toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
				toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, toDevHead)
				strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
			}

			log.Debug("[", uuid, "] Cmd2Device resp to device, WlMqttPublishPad ", strToDevData)
			mqtt.WlMqttPublishPad(uuid, strToDevData)
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