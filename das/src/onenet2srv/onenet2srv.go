package onenet2srv

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dlintw/goconf"
	"github.com/json-iterator/go"

	"das/core/constant"
	"das/core/entity"
	"das/core/httpgo"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/procnbmsg"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func OneNET2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	var httpPort int

	// 判断是否为https协议
	isHttps, err := conf.GetBool("onenet2http", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	}

	httpgo.InitOneNETConfig(conf)
	httpPort, _ = conf.GetInt("onenet2http", "onenet2http_port")

	srv := &http.Server{Addr: ":" + strconv.Itoa(httpPort)}

	http.HandleFunc("/onenet", OnenetHandler)

	go func() {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			log.Info("OneNET2HttpSrvStart ListenAndServeTLS() start...")
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("OneNET2HttpSrvStart ListenAndServeTLS() error = ", err_https)
			}
		} else {
			log.Info("OneNET2HttpSrvStart ListenAndServe() start...")
			if err_http := srv.ListenAndServe(); err_http != nil {
				// cannot panic, because this probably is an intentional close
				log.Error("OneNET2HttpSrvStart ListenAndServe() error =", err_http)
			}
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

type OnenetJob struct {
	rawData []byte
}

func NewOnenetJob(rawData []byte) OnenetJob {
	return OnenetJob{
		rawData: rawData,
	}
}

func (o OnenetJob) Handle() {
	// log.Debug("onenet2srv.Handle() get: ", bytes.NewBuffer(o.rawData).String())

	// 1、解析OneNET消息
	var data entity.OneNETData
	if err := json.Unmarshal(o.rawData, &data); err != nil {
		log.Error("OneNETData json.Unmarshal, err=", err)
		return
	}
	//1. 锁对接的平台，存入redis
	redis.SetDevicePlatformPool(data.Msg.Imei, constant.ONENET_PLATFORM)

	switch data.Msg.Msgtype {
	case 2: // 设备上下线消息(type=2)
		{
			log.Info("OneNET Upload_lock_active, imei=", data.Msg.Imei, ", time=", data.Msg.At/1000)

			var nTime int64
			nTime = 0
			if 1 == data.Msg.Status { // 设备上线
				nTime = 1
			} else if 0 == data.Msg.Status { // 设备离线
				nTime = 0
			}

			//1. 锁状态，存入redis
			redis.SetActTimePool(data.Msg.Imei, nTime)

			//struct 到json str
			var toApp entity.DeviceActive
			toApp.Cmd = constant.Upload_lock_active
			toApp.Ack = 0
			toApp.DevType = ""
			toApp.Vendor = ""
			toApp.DevId = data.Msg.Imei
			toApp.SeqId = 0
			toApp.Signal = 0
			toApp.Time = nTime

			if toApp_str, err := json.Marshal(toApp); err == nil {
				//2. 回复到APP
				//producer.SendMQMsg2APP(data.Msg.Imei, string(toApp_str))
				rabbitmq.Publish2app(toApp_str, data.Msg.Imei)
			} else {
				log.Error("toApp json.Marshal, err=", err)
			}
		}
	case 1: // 数据点消息(type=1)，
		{
			// 处理数据点消息
			procnbmsg.ProcessNbMsg(data.Msg.Value, data.Msg.Imei)
		}
	}
}

func OnenetHandler(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()          //解析参数，默认是不会解析的
	if "GET" == req.Method { // 基本配置：oneNET校验第三方接口
		log.Debug("httpJob.init MaxWorker: ", jobque.MaxWorker, ", MaxQueue: ", jobque.MaxQueue)
		msg := req.Form.Get("msg")

		if "" != msg { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to OneNET, ", msg)
		}
	} else if "POST" == req.Method { // 接收OneNET推送过来的数据
		result, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Error("get req.Body failed")
		} else {
			// fetch job
			jobque.JobQueue <- NewOnenetJob(result)
		}
	}
}
