package consumer

import (
	"../../core/constant"
	"../../core/entity"
	"../../core/httpgo"
	"../../core/log"
	"../../core/redis"
	"../../core/util"
	"../producer"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
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

	// 若为远程开锁流程且查询redis能查到random，则需要进行SM2加签
	if head.Cmd == constant.Remote_open {
		//1. 先判断是否为亿速码加签名，查询redis，若为远程开锁流程且能查到random，则需要加签名
		random, err0 := redis.GetDeviceYisumaRandomfromPool(head.DevId)
		if err0 != nil {
			log.Error("Get YisumaRandom from redis failed, err=", err0)
			return err0
		}

		//2. 亿速码加签
		if "" != random {
			signRandom, err0 := util.AddYisumaRandomSign(head, p.pri, random) // 加上签名
			if err0 != nil {
				log.Error("ProcessAppMsg util.AddYisumaRandomSign error, err=", err0)
				return err0
			}
			p.pri = signRandom
		}
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

	platform, errPlat := redis.GetDevicePlatformPool(imei)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch platform {
	case "onenet":
		{
			respStr, err := httpgo.Http2OneNET_write(imei, strToDevData, strconv.Itoa(head.Cmd))
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
					redis.SetActTimePool(devAct.DevId, devAct.Time)
				}
			}
		}
	case "telecom":
		{

		}
	case "andlink":
		{

		}
	default:
		{
			log.Error("Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	// time.Sleep(time.Second)
	return nil
}
