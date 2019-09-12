package wifi2srv

import (
	"../core/rabbitmq"
	"../core/log"
	"github.com/dlintw/goconf"
	"../core/jobque"
	"../core/redis"
	"strings"
	"../core/constant"
	"../procnbmsg"
)

var rmq_uri string;
var exchange string;		// = "App2OneNET"
var exchangeType string;	// = "direct"
var routingKey string;		// = "wonlycloud"

//初始化RabbitMQ交换器，消息队列名称
func InitRmq_Ex_Que_Name(conf *goconf.ConfigFile) {
	rmq_uri, _ = conf.GetString("rabbitmq", "rabbitmq_uri")
	if rmq_uri == "" {
		log.Error("未启用RabbitMq")
		return
	}
	exchange, _ = conf.GetString("rabbitmq", "device2srv_ex")
	exchangeType, _ = conf.GetString("rabbitmq", "device2srv_ex_type")
	routingKey, _ = conf.GetString("rabbitmq", "device2srv_que")
}

type WifiPlatJob struct {
	rawData string
	devID string
}

func NewWifiPlatJob(rawData string, devID string) WifiPlatJob {
	return WifiPlatJob{
		rawData: rawData,
		devID: devID,
	}
}

func (w WifiPlatJob) Handle() {
	procnbmsg.ProcessNbMsg(w.rawData, w.devID)
}

func ReceiveMQMsgFromDevice() {
	log.Info("start ReceiveMQMsgFromDevice......")

	//初始化rabbitmq
	if rabbitmq.ConsumerRabbitMq == nil {
		log.Error("ReceiveMQMsgFromDevice: rabbitmq.ConsumerRabbitMq is nil.")
		return
	}

	channleContxt := rabbitmq.ChannelContext{Exchange: exchange, ExchangeType: exchangeType, RoutingKey: routingKey, Reliable: true, Durable: true, ReSendNum: 0}

	rabbitmq.ConsumerRabbitMq.QueueDeclare(&channleContxt)

	log.Info("Consumer ReceiveMQMsgFromDevice......")
	// go程循环去读消息，并放到Job去处理
	for {
		msgs := rabbitmq.ConsumerRabbitMq.Consumer(&channleContxt)
		forever := make(chan bool)
		go func() {
			for d := range msgs {
				log.Error("Consumer ReceiveMQMsgFromDevice 1: ", string(d.Body))

				//1. 检验数据是否合法
				getData := string(d.Body)
				if !strings.Contains(getData, "#") {
					log.Error("ReceiveMQMsgFromDevice: rabbitmq.ConsumerRabbitMq error msg: ", getData)
					continue
				}

				//2. 获取设备编号
				prData := strings.Split(getData, "#")
				var devID string
				var devData string
				devID = prData[0]
				devData = prData[1]

				//3. 锁对接的平台，存入redis
				redis.SetDevicePlatformPool(devID, constant.WIFI_PLATFORM)

				//4. fetch job
				// work := httpJob.Job { Serload: httpJob.Serload { DValue: devData, Imei:devID, MsgFrom:constant.NBIOT_MSG }}
				jobque.JobQueue <- NewWifiPlatJob(devData, devID)
			}
		}()
		<-forever
	}
}
