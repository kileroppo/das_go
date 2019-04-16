package consumer

import (
	"../../core/httpgo"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"../../core/log"
	"../../core/constant"
	"../producer"
	"../../core/entity"
	"../../core/redis"
	"strings"
	"../../core/util"
)


type AppMsg struct {
	pri string
}

/*
*	处理APP发送过来的命令消息
*
*/
func (p *AppMsg) ProcessAppMsg() error {
	log.Debug("ProcessAppMsg process msg from app: ", p.pri)

	// 1、解析消息
	//json str 转struct(部份字段)
	var head entity.Header
	if err := json.Unmarshal([]byte(p.pri), &head); err != nil {
		log.Error("ProcessAppMsg json.Unmarshal Header error, err=", err)
		return err
	}

	// 将命令发到OneNET
	imei := head.DevId

	// 加密数据
	var toDevHead entity.MyHeader
	toDevHead.ApiVersion = constant.API_VERSION
	toDevHead.ServiceType = constant.SERVICE_TYPE

	myKey := util.MD52Bytes(head.DevId)
	var strToDevData string
	var err error
	// if strToDevData, toDevHead.CheckSum, err = util.ECBEncrypt([]byte(p.pri), myKey); err == nil {
	if strToDevData, err = util.ECBEncrypt([]byte(p.pri), myKey); err == nil {
		toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
		toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, toDevHead)
		strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
	}


	// respStr, err := httpgo.Http2OneNET_write(imei, p.pri)
	respStr, err := httpgo.Http2OneNET_write(imei, strToDevData)
	if "" != respStr && nil == err {
		var respOneNET entity.RespOneNET
		if err := json.Unmarshal([]byte(respStr), &respOneNET); err != nil {
			log.Error("ProcessAppMsg json.Unmarshal RespOneNET error, err=", err)
			return err
		}

		if 0 != respOneNET.RespErrno {
			var devAct entity.DeviceActive
			devAct.Cmd = constant.Upload_lock_active
			devAct.Ack = 0
			devAct.DevType = head.DevType
			devAct.DevId = head.DevId
			devAct.Vendor = head.Vendor
			devAct.SeqId = 0
			devAct.Time = 0

			//1. 回复APP，设备离线状态
			if toApp_str, err := json.Marshal(devAct); err == nil {
				log.Info("[", head.DevId, "] ProcessAppMsg() device timeout, resp to APP, ", string(toApp_str))
				producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
			} else {
				log.Error("[", head.DevId, "] ProcessAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
			}

			//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
			redis.SetData(devAct.DevId, devAct.Time)
		}
	}

	// time.Sleep(time.Second)
	return nil
}
