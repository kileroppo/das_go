package telecom2srv

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dlintw/goconf"
	"github.com/json-iterator/go"

	"das/core/constant"
	"das/core/entity"
	"das/core/jobque"
	"das/core/log"
	"das/core/redis"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Telecom2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	// 判断是否为https协议
	var httpPort int

	// 判断是否为https协议
	isHttps, err := conf.GetBool("telecom2http", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	}

	httpPort, _ = conf.GetInt("telecom2http", "telecom2https_port")

	srv := &http.Server{Addr: ":" + strconv.Itoa(httpPort)}

	http.HandleFunc("/telecom", TelecomHandler)

	go func() {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			log.Info("Telecom2HttpSrvStart ListenAndServeTLS() start...")
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Telecom2HttpSrvStart ListenAndServeTLS() error = ", err_https)
			}
		} else {
			log.Info("Telecom2HttpSrvStart ListenAndServe() start...")
			if err_http := srv.ListenAndServe(); err_http != nil {
				// cannot panic, because this probably is an intentional close
				log.Error("Telecom2HttpSrvStart ListenAndServe() error = ", err_http)
			}
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

type TelecomJob struct {
}

func NewTelecomJob(rawData []byte) TelecomJob {
	return TelecomJob{}
}

func (t TelecomJob) Handle() {
	//telecom消息处理
}

func TelecomHandler(res http.ResponseWriter, req *http.Request) {
	req.ParseForm()          //解析参数，默认是不会解析的
	if "GET" == req.Method { // 基本配置：
		log.Debug("httpJob.init MaxWorker: ", jobque.MaxWorker, ", MaxQueue: ", jobque.MaxQueue)
		msg := req.Form.Get("msg")
		// signature := req.Form.Get("signature")
		// nonce := req.Form.Get("nonce")
		if "" != msg { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to telecom, ", msg)
		}
	} else if "POST" == req.Method { // 接收OneNET推送过来的数据
		result, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Error("get req.Body failed")
		} else {
			// 处理Telecom推送过来的消息
			log.Debug("telecom.Entry() get: ", bytes.NewBuffer(result).String())

			// 1、解析TeleCom消息
			var data entity.TelecomDeviceDataChanged
			if err := json.Unmarshal([]byte(result), &data); err != nil {
				log.Error("TelecomDeviceDataChanged json.Unmarshal, err=", err)
				return
			}

			//1. 锁对接的平台，存入redis
			redis.SetDevicePlatformPool(data.DeviceId, constant.TELECOM_PLATFORM)

			// fetch job
			//work := httpJob.Job{Serload: httpJob.Serload{DValue: data.Service.Data, Imei: data.DeviceId, MsgFrom: constant.NBIOT_MSG}}
			jobque.JobQueue <- NewTelecomJob(result)

		}
	}
}
