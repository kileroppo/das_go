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
	"time"
)

/*
*	处理APP发送过来的命令消息
*
 */
func ProcAppMsg(appMsg string) error {
	log.Debug("ProcAppMsg process msg from app: ", appMsg)
	// 1、解析消息
	//json str 转struct(部份字段)
	var head entity.Header
	if err := json.Unmarshal([]byte(appMsg), &head); err != nil {
		log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		return err
	}

	// 若为远程开锁流程且查询redis能查到random，则需要进行SM2加签
	switch head.Cmd {
	case constant.Remote_open: {
		//1. 先判断是否为亿速码加签名，查询redis，若为远程开锁流程且能查到random，则需要加签名
		random, err0 := redis.GetDeviceYisumaRandomfromPool(head.DevId)
		if err0 == nil {
			//2. 亿速码加签
			if "" != random {
				signRandom, err0 := util.AddYisumaRandomSign(head, appMsg, random) // 加上签名
				if err0 != nil {
					log.Error("ProcAppMsg util.AddYisumaRandomSign error, err=", err0)
					return err0
				}
				appMsg = signRandom
			}
		}
	}
	case constant.Add_dev_user: {	// 添加锁用户，用户名
		var addDevUser entity.AddDevUser
		if err := json.Unmarshal([]byte(appMsg), &addDevUser); err != nil {
			log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		}

		userNoteTag := strconv.FormatInt(time.Now().Unix(), 10)
		addDevUser.UserNote = userNoteTag
		if uNoteErr := redis.SetDevUserNotePool(addDevUser.DevId, addDevUser.UserNote, userNoteTag); uNoteErr != nil {
			log.Error("ProcAppMsg redis.SetDevUserNotePool error, err=", uNoteErr)
			return uNoteErr
		}
		addDevUserStr, err1 := json.Marshal(addDevUser)
		if err1 != nil {
			log.Error("Get addDevUser json.Marshal failed, err=", err1)
			return err1
		}
		appMsg = string(addDevUserStr)
		log.Debug("ProcAppMsg , appMsg=", appMsg)
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
	if strToDevData, err = util.ECBEncrypt([]byte(appMsg), myKey); err == nil {
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
	case constant.ONENET_PLATFORM:
		{
			respStr, err := httpgo.Http2OneNET_write(imei, strToDevData, strconv.Itoa(head.Cmd))
			if "" != respStr && nil == err {
				var respOneNET entity.RespOneNET
				if err := json.Unmarshal([]byte(respStr), &respOneNET); err != nil {
					log.Error("ProcAppMsg json.Unmarshal RespOneNET error, err=", err)
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
						log.Info("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, ", string(toApp_str))
						producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
					} else {
						log.Error("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
					}

					//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
					redis.SetActTimePool(devAct.DevId, devAct.Time)
				}
			}
		}
	case constant.TELECOM_PLATFORM:
		{

		}
	case constant.ANDLINK_PLATFORM:
		{

		}
	case constant.WIFI_PLATFORM:
		{
			producer.SendMQMsg2Device(imei, strToDevData, strconv.Itoa(head.Cmd))
		}
	default:
		{
			log.Error("ProcAppMsg::Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	// time.Sleep(time.Second)
	return nil
}
