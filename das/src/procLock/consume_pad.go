package procLock

import (
	"strings"

	"github.com/tidwall/gjson"

	"das/core/constant"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/util"
)


//初始化RabbitMQ交换器，消息队列名称
//func InitRmq_Ex_Que_Name(conf *goconf.ConfigFile) {
//	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
//	if rmq_uri == "" {
//		log.Error("未启用RabbitMq")
//		return
//	}
//	exchange, _ = conf.GetString("rabbitmq", "device2srv_ex")
//	exchangeType, _ = conf.GetString("rabbitmq", "device2srv_ex_type")
//	routingKey, _ = conf.GetString("rabbitmq", "device2srv_que")
//}

type WifiPlatJob struct {
	rawData string
	devID   string
}

func NewWifiPlatJob(rawData string, devID string) WifiPlatJob {
	return WifiPlatJob{
		rawData: rawData,
		devID:   devID,
	}
}

func (w WifiPlatJob) Handle() {
	ProcessJsonMsg(w.rawData, w.devID)
}

func consumePadDoor() {
	log.Info("start ReceiveMQMsgFromPadDoor......")
	msgs, err := rabbitmq.ConsumeDev()
	if err != nil {
		log.Errorf("consumePadDoor > %s", err)
	}

	for d := range msgs {
		log.SendGraylogByUDP("Pad设备-mq->DAS: %s", d.Body)
		//log.Info("Consumer ReceiveMQMsgFromDevice: ", string(d.Body))

		//1. 检验数据是否合法
		getData := util.Bytes2Str(d.Body)
		if !strings.Contains(getData, "#") {
			log.Error("Pad设备-mq->DAS: rabbitmq.ConsumerRabbitMq error msg: ", getData)
			continue
		}

		//2. 获取设备编号
		prData := strings.Split(getData, "#")
		var devID string
		var devData string
		if len(prData) > 1 {
			devID = prData[0]
			devData = prData[1]
		} else {
			return
		}

		//来自中控平板的消息不做处理
		if !IsDataFromZKGtw(devData) {
			//3. 锁对接的平台，存入redis
			mymap := make(map[string]interface{})
			mymap["from"] = constant.PAD_DEVICE_PLATFORM
			go redis.SetDevicePlatformPool(devID, mymap)
		}

		//4. fetch job
		// work := httpJob.Job { Serload: httpJob.Serload { DValue: devData, Imei:devID, MsgFrom:constant.NBIOT_MSG }}
		jobque.JobQueue <- NewWifiPlatJob(devData, devID)
	}

	select {
	case <-ctx.Done():
		log.Info("ReceiveMQMsgFromDevice Close")
		return
	default:
		go consumePadDoor()
		return
	}
}

func IsDataFromZKGtw(data string) bool {
	devType := gjson.Get(data, "devType").String()
	cmd := gjson.Get(data, "cmd").Int()
	return devType == constant.Device_Type_ZKGtw || cmd == constant.Scene_Trigger || cmd == constant.Set_AIPad_Reboot_Time
}
