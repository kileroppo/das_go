package aliIot2srv

import (
	"../core/log"
	"encoding/json"
	"strings"
	"../core/constant"
	"../core/redis"
)

type AliData struct {
	Value string		`json:"value"`
	Time int64			`json:"time"`
}
type AliItems struct {
	UserData AliData	`json:"UserData"`
}
type AliIoTData struct{
	DeviceType string	`json:"deviceType"`
	IotId string		`json:"iotId"`
	RequestId string	`json:"requestId"`
	ProductKey string	`json:"productKey"`
	GmtCreate int64		`json:"gmtCreate"`
	DeviceName string	`json:"deviceName"`
	Items AliItems		`json:"items"`
}

type AliIoTStatus struct {
	DeviceType string	`json:"deviceType"`
	IotId string		`json:"iotId"`
	Action string		`json:"action"`
	ProductKey string	`json:"productKey"`
	GmtCreate string	`json:"gmtCreate"`
	DeviceName string	`json:"deviceName"`
	Status AliData		`json:"status"`
}

var DEVICETYPE = []string{"", "WlWiFiLock"}

func ProcessAliMsg(data []byte, topic string) error {
	log.Debugf("ali-topic: %s -> \n %s", topic , string(data))
	var err error
	if strings.Contains(topic, "thing/event/property/post") { // 数据
		var aliData AliIoTData
		if err = json.Unmarshal(data, &aliData); err != nil {
			log.Error("[", aliData.DeviceName, "] AliIoTData json.Unmarshal, err=", err)
			return err // break
		}

		// 锁对接的平台，存入redis
		redis.SetDevicePlatformPool(aliData.DeviceName, constant.ALIIOT_PLATFORM)

		// 数据解析
		err = ParseData(aliData.Items.UserData.Value)
		if nil != err {
			log.Error("ProcessAliMsg parseData, err=", err)
			return err
		}
	} else if strings.Contains(topic, "mqtt/status") { // 在线|离线状态
		var aliStatus AliIoTStatus
		if err = json.Unmarshal(data, &aliStatus); err != nil {
			log.Error("[", aliStatus.DeviceName, "] AliIoTStatus json.Unmarshal, err=", err)
			return err // break
		}

		// 锁对接的平台，存入redis
		redis.SetDevicePlatformPool(aliStatus.DeviceName, constant.ALIIOT_PLATFORM)
	}

	return nil
}
