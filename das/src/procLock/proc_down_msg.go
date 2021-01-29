package procLock

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"das/core/constant"
	"das/core/entity"
	"das/core/httpgo"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/util"
)

/*
 *	处理APP下行数据
 *
 */
func ProcAppMsg(appMsg string) error {
	// var strAppMsg string = appMsg
	if !strings.ContainsAny(appMsg, "{ & }") { // 判断数据中是否正确的json，不存在，则是错误数据.
		/*log.Error("ProcAppMsg() error msg : ", appMsg)
		return errors.New("error msg.")*/
		// TODO:JHHE APP下行数据解密
		if !strings.ContainsAny(appMsg, "#") {
			log.Error("ProcAppMsg() error msg")
			return errors.New("error msg.")
		}
		//1. 获取设备编号
		prData := strings.Split(appMsg, "#")
		var devID string
		var devData string
		devID = prData[0]
		devData = prData[1]

		//2. 校验数据正确性
		lens := strings.Count(devData, "") - 1
		if lens < 16 {
			log.Error("[", devID, "] ProcAppMsg() error msg : ", devData, ", len: ", lens)
			return errors.New("error msg.")
		}

		//3. 解密数据
		var myHead entity.MyHeader
		var strHead string
		strHead = devData[0:16]
		byteHead, _ := hex.DecodeString(strHead)
		myHead.ApiVersion = util.BytesToInt16(byteHead[0:2])
		myHead.ServiceType = util.BytesToInt16(byteHead[2:4])
		myHead.MsgLen = util.BytesToInt16(byteHead[4:6])
		myHead.CheckSum = util.BytesToInt16(byteHead[6:8])
		//log.Info("[", devID, "] ProcAppMsg() ApiVersion: ", myHead.ApiVersion, ", ServiceType: ", myHead.ServiceType, ", MsgLen: ", myHead.MsgLen, ", CheckSum: ", myHead.CheckSum)

		var checkSum uint16
		var strData string
		strData = devData[16:]
		checkSum = util.CheckSum([]byte(strData))
		if checkSum != myHead.CheckSum {
			log.Error("[", devID, "] ProcAppMsg() CheckSum failed, src:", myHead.CheckSum, ", dst: ", checkSum)
			return errors.New("CheckSum failed.")
		}

		myKey := util.MD52Bytes(devID + "Potato")
		var err_aes error
		appMsg, err_aes = util.ECBDecrypt(strData, myKey)
		if nil != err_aes {
			log.Error("[", devID, "] util.ECBDecrypt failed, strData:", strData, ", key: ", myKey, ", error: ", err_aes)
			return err_aes
		}
		//rabbitmq.SendGraylogByMQ("[%s] After ECBDecrypt, data.Msg.Value: %s", devID, appMsg)
		//log.Info("[", devID, "] After ECBDecrypt, data.Msg.Value: ", appMsg)
	}

	// 1、解析消息
	var head entity.Header
	if err := json.Unmarshal([]byte(appMsg), &head); err != nil {
		log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		return err
	}

	// 记录APP下行日志
	var esLog entity.EsLogEntiy // 记录日志
	esLog.Operation = "APP-rmq-DAS"
	// rabbitmq.SendGraylogByMQ("下行数据(APP-mq->DAS)：dev[%s]; %s >>> %s", head.DevId, strAppMsg, appMsg)
	//sendPadDoorUpLogMsg(head.DevId, strAppMsg + ">>>" + appMsg, "下行设备数据")

	//2. 数据干预处理
	// 若为远程开锁流程且查询redis能查到random，则需要进行SM2加签
	switch head.Cmd {
	case constant.Remote_open: {
		esLog.Operation += "远程开门"
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
		} else { // 非亿速码
			// 单双人判断
			if strings.Contains(appMsg, "passwd2") { // 双人
				var mROpenLockReq entity.MRemoteOpenLockReq
				if err := json.Unmarshal([]byte(appMsg), &mROpenLockReq); err != nil {
					log.Error("ProcAppMsg json.Unmarshal MRemoteOpenLockReq error, err=", err)
				}

				appUserTag := strconv.FormatInt(time.Now().Unix(), 16)
				//if uNoteErr := redis.SetAppUserPool(mROpenLockReq.DevId, appUserTag, mROpenLockReq.AppUser); uNoteErr != nil {
				//	log.Error("ProcAppMsg redis.SetAppUserPool error, err=", uNoteErr)
				//	return uNoteErr
				//}
				go redis.SetAppUserPool(mROpenLockReq.DevId, appUserTag, mROpenLockReq.AppUser)
				mROpenLockReq.AppUser = appUserTag // 值跟KEY交换，下发到锁端

				stringKey := strings.ToUpper(util.Md5(head.DevId))
				if len(mROpenLockReq.Passwd) > 6 {
					psw1, err_0 := hex.DecodeString(mROpenLockReq.Passwd)
					if err_0 != nil {
						log.Error("ProcAppMsg MRemoteOpenLockReq DecodeString 0 failed, err=", err_0)
						return err_0
					}

					passwd1, err0 := util.ECBDecryptByte(psw1, []byte(stringKey))
					if err0 != nil {
						log.Error("ProcAppMsg MRemoteOpenLockReq ECBDecryptByte 0 failed, err=", err0)
						return err0
					}

					mROpenLockReq.Passwd = string(passwd1)
				}

				if len(mROpenLockReq.Passwd2) > 6 {
					psw2, err_1 := hex.DecodeString(mROpenLockReq.Passwd2)
					if err_1 != nil {
						log.Error("ProcAppMsg MRemoteOpenLockReq DecodeString 1 failed, err=", err_1)
						return err_1
					}

					passwd2, err1 := util.ECBDecryptByte(psw2, []byte(stringKey))
					if err1 != nil {
						log.Error("ProcAppMsg MRemoteOpenLockReq ECBDecryptByte 1 failed, err=", err1)
						return err1
					}
					mROpenLockReq.Passwd2 = string(passwd2)
				}

				jsonStr, err2 := json.Marshal(mROpenLockReq)
				if err2 != nil {
					log.Error("ProcAppMsg MRemoteOpenLockReq json.Marshal failed, err=", err2)
					return err2
				}
				appMsg = string(jsonStr)
			} else { // 单人
				var sROpenLockReq entity.SRemoteOpenLockReq
				if err := json.Unmarshal([]byte(appMsg), &sROpenLockReq); err != nil {
					log.Error("ProcAppMsg json.Unmarshal SRemoteOpenLockReq error, err=", err)
				}

				appUserTag := strconv.FormatInt(time.Now().Unix(), 16)
				//if uNoteErr := redis.SetAppUserPool(sROpenLockReq.DevId, appUserTag, sROpenLockReq.AppUser); uNoteErr != nil {
				//	log.Error("ProcAppMsg redis.SetAppUserPool error, err=", uNoteErr)
				//	return uNoteErr
				//}
				go redis.SetAppUserPool(sROpenLockReq.DevId, appUserTag, sROpenLockReq.AppUser)
				sROpenLockReq.AppUser = appUserTag // 值跟KEY交换，下发到锁端

				if len(sROpenLockReq.Passwd) > 6 {
					psw1, err0 := hex.DecodeString(sROpenLockReq.Passwd)
					if err0 != nil {
						log.Error("ProcAppMsg SRemoteOpenLockReq DecodeString failed, err=", err0)
						return err0
					}
					stringKey := strings.ToUpper(util.Md5(head.DevId))
					passwd1, err1 := util.ECBDecryptByte(psw1, []byte(stringKey))
					if err1 != nil {
						log.Error("ProcAppMsg SRemoteOpenLockReq ECBDecryptByte failed, err=", err1)
						return err1
					}

					sROpenLockReq.Passwd = string(passwd1)
				}

				jsonStr, err2 := json.Marshal(sROpenLockReq)
				if err2 != nil {
					log.Error("ProcAppMsg SRemoteOpenLockReq json.Marshal failed, err=", err2)
					return err2
				}
				appMsg = string(jsonStr)
			}
		}
	}
	case constant.Add_dev_user: {
		esLog.Operation += "添加锁用户"
		// 添加锁用户，用户名
		var addDevUser entity.AddDevUser
		if err := json.Unmarshal([]byte(appMsg), &addDevUser); err != nil {
			log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		}

		userTag := strconv.FormatInt(time.Now().Unix(), 16)
		//if uNoteErr := redis.SetDevUserNotePool(addDevUser.DevId, userTag, addDevUser.UserNote); uNoteErr != nil {
		//	log.Error("ProcAppMsg redis.SetDevUserNotePool error, err=", uNoteErr)
		//	return uNoteErr
		//}
		go redis.SetDevUserNotePool(addDevUser.DevId, userTag, addDevUser.UserNote)
		addDevUser.UserNote = userTag // 值跟KEY交换，下发到锁端

		//if appUserErr := redis.SetAppUserPool(addDevUser.DevId, userTag, addDevUser.AppUser); appUserErr != nil {
		//	log.Error("ProcAppMsg redis.SetAppUserPool error, err=", appUserErr)
		//	return appUserErr
		//}
		go redis.SetAppUserPool(addDevUser.DevId, userTag, addDevUser.AppUser)
		addDevUser.AppUser = userTag // 值跟KEY交换，下发到锁端
		if addDevUser.Total > 0 {
			addDevUser.Count = addDevUser.Total // TODO:JHHE 2020-10-22 兼容旧版平板门
		}

		if constant.OPEN_PWD == addDevUser.MainOpen { // 主开锁方式（1-密码，2-刷卡，3-指纹，5-人脸，12-蓝牙）
			if len(addDevUser.Passwd) > 6 {
				psw1, err0 := hex.DecodeString(addDevUser.Passwd)
				if err0 != nil {
					log.Error("ProcAppMsg AddDevUser DecodeString failed, err=", err0)
					return err0
				}
				stringKey := strings.ToUpper(util.Md5(head.DevId))
				passwd1, err1 := util.ECBDecryptByte(psw1, []byte(stringKey))
				if err1 != nil {
					log.Error("ProcAppMsg AddDevUser ECBDecryptByte failed, err=", err1)
					return err1
				}

				addDevUser.Passwd = string(passwd1)
			}
		}

		// json转字符串
		addDevUserStr, err1 := json.Marshal(addDevUser)
		if err1 != nil {
			log.Error("ProcAppMsg addDevUser json.Marshal failed, err=", err1)
			return err1
		}
		appMsg = string(addDevUserStr)
		//log.Debug("ProcAppMsg , appMsg=", appMsg)
	}
	case constant.Set_dev_user_temp: {
		esLog.Operation += "设置锁临时用户"
		var setTmpDevUser entity.SetTmpDevUser
		if err := json.Unmarshal([]byte(appMsg), &setTmpDevUser); err != nil {
			log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		}
		if setTmpDevUser.Total > 0 {
			setTmpDevUser.Count = setTmpDevUser.Total // TODO:JHHE 2020-10-22 兼容旧版平板门
		}
		// json转字符串
		setTmpDevUserStr, err1 := json.Marshal(setTmpDevUser)
		if err1 != nil {
			log.Error("ProcAppMsg setTmpDevUser json.Marshal failed, err=", err1)
			return err1
		}
		appMsg = string(setTmpDevUserStr)
	}
	case constant.Del_dev_user: {
		esLog.Operation += "删除锁用户"
		// 添加锁用户，用户名
		var delDevUser entity.DelDevUser
		if err := json.Unmarshal([]byte(appMsg), &delDevUser); err != nil {
			log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		}

		userTag := strconv.FormatInt(time.Now().Unix(), 16)
		//if appUserErr := redis.SetAppUserPool(delDevUser.DevId, userTag, delDevUser.AppUser); appUserErr != nil {
		//	log.Error("ProcAppMsg redis.SetAppUserPool error, err=", appUserErr)
		//	return appUserErr
		//}
		go redis.SetAppUserPool(delDevUser.DevId, userTag, delDevUser.AppUser)
		delDevUser.AppUser = userTag // 值跟KEY交换，下发到锁端

		// json转字符串
		delDevUserStr, err1 := json.Marshal(delDevUser)
		if err1 != nil {
			log.Error("ProcAppMsg delDevUserStr json.Marshal failed, err=", err1)
			return err1
		}
		appMsg = string(delDevUserStr)
		//log.Debug("ProcAppMsg , appMsg=", appMsg)
	}
	case constant.Set_dev_para: {
		esLog.Operation += "设置锁参数"
		// 添加锁用户，用户名
		var setLockParamReq entity.SetLockParamReq
		if err := json.Unmarshal([]byte(appMsg), &setLockParamReq); err != nil {
			log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
			return err
		}

		userTag := strconv.FormatInt(time.Now().Unix(), 16)
		//if appUserErr := redis.SetAppUserPool(setLockParamReq.DevId, userTag, setLockParamReq.AppUser); appUserErr != nil {
		//log.Error("ProcAppMsg redis.SetAppUserPool error, err=", appUserErr)
		//return appUserErr
		//}
		go redis.SetAppUserPool(setLockParamReq.DevId, userTag, setLockParamReq.AppUser)
		setLockParamReq.AppUser = userTag // 值跟KEY交换，下发到锁端

		// json转字符串
		setLockParamReqStr, err1 := json.Marshal(setLockParamReq)
		if err1 != nil {
			log.Error("ProcAppMsg setLockParamReqStr json.Marshal failed, err=", err1)
			return err1
		}
		appMsg = string(setLockParamReqStr)
		//log.Debug("ProcAppMsg , appMsg=", appMsg)
	}
	case constant.Wonly_LGuard_Msg: {
		esLog.Operation += "小卫士操作"
		//小卫士消息
		go httpgo.Http2FeibeeWonlyLGuard(appMsg)
		return nil
	}
	case constant.Body_Fat_Scale:
		esLog.Operation += "将体脂秤数据上报到服务器"
		rabbitmq.Publish2pms([]byte(appMsg), "")
		return nil
	case constant.Video_Hang_Up:
		esLog.Operation += "视频锁通话挂断请求上报到服务器"
		send_data := util.Str2Bytes(appMsg)
		rkey := gjson.Get(appMsg, "devId").String()
		rabbitmq.Publish2app(send_data, rkey) //to 中控平板
		rabbitmq.Publish2dev(send_data, rkey) //to 油烟机
		return nil
	}
	//log.Debug("ProcAppMsg after, ", appMsg)

	ret, errPlat := redis.GetDevicePlatformPool(head.DevId)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch ret["from"] {
	case constant.ONENET_PLATFORM: { // 移动OneNET平台
		esLog.ThirdPlatform = "移动OneNET平台"
		var toDevHead entity.MyHeader
		toDevHead.ApiVersion = constant.API_VERSION
		toDevHead.ServiceType = constant.SERVICE_TYPE

		// 解密数据
		myKey := util.MD52Bytes(head.DevId)
		var strToDevData string
		var err error
		if strToDevData, err = util.ECBEncrypt([]byte(appMsg), myKey); err == nil {
			toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
			toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, toDevHead)
			strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
		}

		respStr, err := httpgo.Http2OneNET_write(head.DevId, strToDevData, strconv.Itoa(head.Cmd))
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
				devAct.Signal = 0
				devAct.Time = 0

				//1. 回复APP，设备离线状态
				if toApp_str, err := json.Marshal(devAct); err == nil {
					//log.Info("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, ", string(toApp_str))
					//producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
					rabbitmq.Publish2app(toApp_str, devAct.DevId)
				} else {
					log.Error("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
				}

				//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
				go redis.SetActTimePool(devAct.DevId, int64(devAct.Time))
			}
		}
	}
	case constant.TELECOM_PLATFORM: {} // 电信平台
	case constant.ANDLINK_PLATFORM: {} // 移动AndLink平台
	case constant.PAD_DEVICE_PLATFORM: {	// WiFi平板
		esLog.ThirdPlatform = "王力RabbitMQ"
		var strToDevData string
		var err error
		if constant.PadDoor_RealVideo == head.Cmd { // TODO:JHHE 临时方案，平板锁开启视频不加密
			strToDevData = appMsg
		} else {
			// 加密数据
			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = constant.SERVICE_TYPE

			myKey := util.MD52Bytes(head.DevId)
			if strToDevData, err = util.ECBEncrypt([]byte(appMsg), myKey); err == nil {
				toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
				toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, toDevHead)
				strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
			}
		}
		SendMQMsg2Device(head.DevId, strToDevData, strconv.Itoa(head.Cmd))
	}
	case constant.MQTT_PAD_PLATFORM: {// WiFi平板，MQTT通道
		esLog.ThirdPlatform = "王力IoT平台"
		var strToDevData string
		var err error
		if constant.PadDoor_RealVideo == head.Cmd { // TODO:JHHE 临时方案，平板锁开启视频不加密
			strToDevData = appMsg
		} else {
			// 加密数据
			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = constant.SERVICE_TYPE

			myKey := util.MD52Bytes(head.DevId)
			if strToDevData, err = util.ECBEncrypt([]byte(appMsg), myKey); err == nil {
				toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
				toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

				buf := new(bytes.Buffer)
				binary.Write(buf, binary.BigEndian, toDevHead)
				strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
			}
		}
		WlMqttPublishPad(head.DevId, strToDevData)
	}
	case constant.ALIIOT_PLATFORM: {	// 阿里云飞燕平台
		esLog.ThirdPlatform = "阿里云飞燕平台"
		bData, err_ := WlJson2BinMsg(appMsg, constant.GENERAL_PROTOCOL)
		if nil != err_ {
			log.Error("ProcAppMsg() WlJson2BinMsg, error: ", err_)
			return err_
		}

		respStatus, _ := httpgo.HttpSetAliPro(head.DevId, hex.EncodeToString(bData), strconv.Itoa(head.Cmd))
		if 200 != respStatus {
			var devAct entity.DeviceActive
			devAct.Cmd = constant.Upload_lock_active
			devAct.Ack = 0
			devAct.DevType = head.DevType
			devAct.DevId = head.DevId
			devAct.Vendor = head.Vendor
			devAct.SeqId = 0
			devAct.Signal = 0
			devAct.Time = 0

			//1. 回复APP，设备离线状态
			if toApp_str, err := json.Marshal(devAct); err == nil {
				//log.Info("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, ", string(toApp_str))
				//producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
				rabbitmq.Publish2app(toApp_str, devAct.DevId)
			} else {
				log.Error("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
			}

			//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
			go redis.SetActTimePool(devAct.DevId, int64(devAct.Time))
		}
	}
	case constant.MQTT_PLATFORM: {	// MQTT
		esLog.ThirdPlatform = "王力IoT平台"
		bData, err_ := WlJson2BinMsg(appMsg, constant.GENERAL_PROTOCOL)
		if nil != err_ {
			log.Error("ProcAppMsg() WlJson2BinMsg, error: ", err_)
			return err_
		}

		WlMqttPublish(head.DevId, bData)
	}
	case constant.FEIBEE_PLATFORM: {//飞比zigbee锁
		esLog.ThirdPlatform = "飞比云平台"
		var msgHead entity.ZigbeeLockHead
		if err := json.Unmarshal([]byte(appMsg), &msgHead); err != nil {
			log.Error("ProcAppMsg json.Unmarshal() error = ", err)
			return err
		}

		var i uint8 = 1
		var nCount uint8 = 1
		if constant.Set_Wifi == msgHead.Cmd { // zigbee常在线锁设置wifi拆分两条命令下行
			nCount = 2
		}
		for ; i <= nCount; i++ {
			appData, err_ := WlJson2BinMsgZigbee(appMsg, i)
			if nil != err_ {
				log.Error("ProcAppMsg() WlJson2BinMsg, error: ", err_)
				return err_
			}
			if "" == ret["uuid"] || "" == ret["uid"] {
				return errors.New("下发给飞比zigbee锁的uuid, uid为空")
			}
			// TODO:JHHE 包体前面增加长度
			bLen := IntToBytes(len(appData))
			strLen := hex.EncodeToString(bLen)
			go httpgo.Http2FeibeeZigbeeLock(strLen[len(strLen)-2:]+"00"+hex.EncodeToString(appData), msgHead.Bindid, msgHead.Bindstr, ret["uuid"], ret["uid"])

			// 连续两条命令间隔100毫秒
			if constant.Set_Wifi == msgHead.Cmd {
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
	default: {
			log.Error("ProcAppMsg::Unknow Platform from redis, please check the platform: ", ret)
		}
	}

	esLog.DeviceId = head.DevId
	esLog.Vendor = head.Vendor
	// esLog.ThirdPlatform = "王力RabbitMQ"
	esLog.RawData = appMsg
	esData, err_ := json.MarshalToString(esLog)
	if err_ != nil {
		log.Warningf("ProcessJsonMsg > json.Marshal > %s", err_)
		return err_
	}
	rabbitmq.SendGraylogByMQ("%s", esData)

	// time.Sleep(time.Second)
	return nil
}

func pushMsgForSceneTrigger(msg *entity.RangeHoodAlarm) {
	msg2pms := entity.Feibee2AutoSceneMsg{
		Header:      msg.Header,
		Time:        msg.Time,
		TriggerType: 0,
		AlarmType:   "rangeHoodGas",
		AlarmFlag:   1,
	}
	msg2pms.Cmd = constant.Scene_Trigger
	data2pms, err := json.Marshal(msg2pms)
	if err != nil {
		log.Warning("ProcAppMsg Wonly_LGuard_Msg json.Marshal() error = ", err)
		return
	}
	//作为场景触发通知
	rabbitmq.Publish2Scene(data2pms, "")
}

func pushMsgForSave(msg *entity.RangeHoodAlarm) {
    msg2pms := entity.Feibee2AlarmMsg{
		Header:     msg.Header,
		Time:       msg.Time,
		AlarmType:  "rangeHoodGas",
		AlarmValue: "油烟机燃气泄漏",
		AlarmFlag:  1,
	}
	msg2pms.Cmd = 0xfc
	data2pms, err := json.Marshal(msg2pms)
	if err != nil {
		log.Warning("ProcAppMsg Wonly_LGuard_Msg json.Marshal() error = ", err)
		return
	}
	rabbitmq.Publish2pms(data2pms, "")
}

//func sendMQTTDownLogMsg(devId, oriData string) {
//	var logMsg entity.SysLogMsg
//	currT := time.Now()
//	logMsg.Timestamp = currT.Unix()
//	logMsg.NanoTimestamp = currT.UnixNano()
//	logMsg.MsgType = 4
//	logMsg.MsgName = "下行设备数据"
//	logMsg.UUid = devId
//	logMsg.VendorName = "王力MQTT"
//
//	buf := bytebufferpool.Get()
//	defer bytebufferpool.Put(buf)
//
//	buf.WriteString("Json数据：")
//	buf.WriteString(oriData)
//
//	logMsg.RawData = buf.String()
//	data,err := json.Marshal(logMsg)
//	if err != nil {
//		log.Warningf("sendMQTTDownLogMsg > json.Marshal > %s", err)
//	} else {
//		rabbitmq.Publish2log(data, "")
//	}
//}

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
