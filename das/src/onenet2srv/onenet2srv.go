package onenet2srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dlintw/goconf"
	"../core/log"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"../httpJob"
	"../core/entity"
	"../core/redis"
	"../core/constant"
	"../mq/producer"
)

func OneNET2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	var httpPort int

	// 判断是否为https协议
	isHttps, err := conf.GetBool("onenet2http", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	}

	httpPort, _ = conf.GetInt("onenet2http", "onenet2http_port")

	srv := &http.Server{Addr: ":"+strconv.Itoa(httpPort)}

	http.HandleFunc("/onenet", Entry)

	go func() {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Httpserver: ListenAndServeTLS(): %s", err_https)
			}
		} else {
			log.Debug("httpServerStart http.ListenAndServe()......")
			if err_http := srv.ListenAndServe(); err_http != nil {
				// cannot panic, because this probably is an intentional close
				log.Error("Httpserver: ListenAndServe(): %s", err_http)
			}
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func Entry(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //解析参数，默认是不会解析的
	if ("GET" == req.Method) { // 基本配置：oneNET校验第三方接口
		log.Debug("httpJob.init MaxWorker: ", httpJob.MaxWorker, ", MaxQueue: ", httpJob.MaxQueue)
		msg := req.Form.Get("msg")
		// signature := req.Form.Get("signature")
		// nonce := req.Form.Get("nonce")
		if("" != msg) { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to OneNET, ", msg)
		}
	} else if ("POST" == req.Method) { // 接收OneNET推送过来的数据
		result, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Error("get req.Body failed")
		} else {
			// 处理OneNET推送过来的消息
			log.Debug("onenet2srv.Entry() get: ", bytes.NewBuffer(result).String())

			// 1、解析OneNET消息
			var data entity.OneNETData
			if err := json.Unmarshal([]byte(result), &data); err != nil {
				log.Error("OneNETData json.Unmarshal, err=", err)
				return
			}
			//1. 锁对接的平台，存入redis
			redis.SetDevicePlatformPool(data.Msg.Imei, "onenet")

			switch data.Msg.Msgtype {
			case 2: // 设备上下线消息(type=2)
				{
					log.Info("OneNET Upload_lock_active, imei=", data.Msg.Imei, ", time=", data.Msg.At/1000)

					var nTime int64
					nTime = 0
					if 1 == data.Msg.Status { // 设备上线
						nTime = data.Msg.At / 1000
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
					toApp.Time = nTime

					if toApp_str, err := json.Marshal(toApp); err == nil {
						//2. 回复到APP
						producer.SendMQMsg2APP(data.Msg.Imei, string(toApp_str))
					} else {
						log.Error("toApp json.Marshal, err=", err)
					}
				}
			case 1: // 数据点消息(type=1)，
				{
					// fetch job
					work := httpJob.Job{Serload: httpJob.Serload{DValue: bytes.NewBuffer([]byte(data.Msg.Value)).String(), Imei:data.Msg.Imei}}
					httpJob.JobQueue <- work
				}
			}
		}
	}
}