package consumer

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"das/core/constant"
	"das/core/entity"
	"das/core/httpgo"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/util"
	"das/rmq/producer"
)

/*
*	处理APP发送过来的命令消息
*
 */
func ProcAppMsg(appMsg string) error {
	log.Debug("ProcAppMsg process msg from app.")
	// 1、解析消息
	var head entity.Header
	if err := json.Unmarshal([]byte(appMsg), &head); err != nil {
		log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		return err
	}

	//2. 数据干预处理
	// 若为远程开锁流程且查询redis能查到random，则需要进行SM2加签
	switch head.Cmd {
	case constant.Remote_open:
		{
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

					decryptPwdFlag := false	// 密码解密
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
						decryptPwdFlag = true
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
						decryptPwdFlag = true
					}

					if decryptPwdFlag {
						jsonStr, err2 := json.Marshal(mROpenLockReq)
						if err2 != nil {
							log.Error("ProcAppMsg MRemoteOpenLockReq json.Marshal failed, err=", err2)
							return err2
						}
						appMsg = string(jsonStr)
					}
				} else { // 单人
					var sROpenLockReq entity.SRemoteOpenLockReq
					if err := json.Unmarshal([]byte(appMsg), &sROpenLockReq); err != nil {
						log.Error("ProcAppMsg json.Unmarshal SRemoteOpenLockReq error, err=", err)
					}

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
						jsonStr, err2 := json.Marshal(sROpenLockReq)
						if err2 != nil {
							log.Error("ProcAppMsg SRemoteOpenLockReq json.Marshal failed, err=", err2)
							return err2
						}
						appMsg = string(jsonStr)
					}
				}
			}
		}
	case constant.Add_dev_user:
		{
			// 添加锁用户，用户名
			var addDevUser entity.AddDevUser
			if err := json.Unmarshal([]byte(appMsg), &addDevUser); err != nil {
				log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
			}

			// if 0xFFFF == addDevUser.UserId { // TODO:jhhe remove
			userNoteTag := strconv.FormatInt(time.Now().Unix(), 16)
			if uNoteErr := redis.SetDevUserNotePool(addDevUser.DevId, userNoteTag, addDevUser.UserNote); uNoteErr != nil {
				log.Error("ProcAppMsg redis.SetDevUserNotePool error, err=", uNoteErr)
				return uNoteErr
			}
			addDevUser.UserNote = userNoteTag // 值跟KEY交换，下发到锁端
			// }

			if 1 == addDevUser.MainOpen { // 主开锁方式（1-密码，2-刷卡，3-指纹，5-人脸，12-蓝牙）
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
			log.Debug("ProcAppMsg , appMsg=", appMsg)
		}

	case constant.Wonly_Guard_Msg:
		//小卫士消息
		httpgo.Http2FeibeeWonlyLGuard(appMsg)

	}

	platform, errPlat := redis.GetDevicePlatformPool(head.DevId)
	if errPlat != nil {
		log.Error("Get Platform from redis failed, err=", errPlat)
		return errPlat
	}

	switch platform {
	case constant.ONENET_PLATFORM: // 移动OneNET平台
		{
			// 加密数据
			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = constant.SERVICE_TYPE

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
						log.Info("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, ", string(toApp_str))
						//producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
						rabbitmq.Publish2app(toApp_str, devAct.DevId)
					} else {
						log.Error("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
					}

					//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
					redis.SetActTimePool(devAct.DevId, int64(devAct.Time))
				}
			}
		}
	case constant.TELECOM_PLATFORM:
		{
		} // 电信平台
	case constant.ANDLINK_PLATFORM:
		{
		} // 移动AndLink平台
	case constant.WIFI_PLATFORM: // WiFi平板锁
		{
			// 加密数据
			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = constant.SERVICE_TYPE

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

			producer.SendMQMsg2Device(head.DevId, strToDevData, strconv.Itoa(head.Cmd))
		}
	case constant.ALIIOT_PLATFORM: // 阿里云飞燕平台
		{
			bData, err_ := WlJson2BinMsg(appMsg)
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
					log.Info("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, ", string(toApp_str))
					//producer.SendMQMsg2APP(devAct.DevId, string(toApp_str))
					rabbitmq.Publish2app(toApp_str, devAct.DevId)
				} else {
					log.Error("[", head.DevId, "] ProcAppMsg() device timeout, resp to APP, json.Marshal, err=", err)
				}

				//2. 锁响应超时唤醒，以此判断锁离线，将状态存入redis
				redis.SetActTimePool(devAct.DevId, int64(devAct.Time))
			}
		}
	default:
		{
			log.Error("ProcAppMsg::Unknow Platform from redis, please check the platform: ", platform)
		}
	}

	// time.Sleep(time.Second)
	return nil
}
