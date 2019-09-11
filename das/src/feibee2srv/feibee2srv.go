package feibee2srv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"../core/log"
	"../httpJob"
	"../core/redis"
	"../core/entity"
)

func Entry(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //解析参数，默认是不会解析的
	if ("GET" == req.Method) { // 基本配置：
		log.Debug("httpJob.init MaxWorker: ", httpJob.MaxWorker, ", MaxQueue: ", httpJob.MaxQueue)
		msg := req.Form.Get("msg")
		// signature := req.Form.Get("signature")
		// nonce := req.Form.Get("nonce")
		if("" != msg) { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to telecom, ", msg)
		}
	} else if ("POST" == req.Method) { // 接收OneNET推送过来的数据
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
			redis.SetDevicePlatformPool(data.DeviceId, "telecom")

			// fetch job
			work := httpJob.Job { Serload: httpJob.Serload { DValue : data.Service.Data, Imei: data.DeviceId, MsgFrom:"feibee"}}
			httpJob.JobQueue <- work
		}
	}
}
