package procLock

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ZZMarquis/gm/sm2"

	"das/core/constant"
	"das/core/entity"
	"das/core/httpgo"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/util"
)

/*
 *	处理锁上行的json数据（nb锁、平板门）
 *
 */

func ProcessJsonMsg(DValue string, devID string) error {
	// 处理OneNET推送过来的消息
	//log.Info("[", devID, "] ProcessJsonMsg msg from before: ", DValue)
	myKey := util.MD52Bytes(devID)

	// 增加二进制包头，以及加密的包体
	// 1、 获取包头部分 8个字节
	var myHead entity.MyHeader
	if !strings.ContainsAny(DValue, "{ & }") { // 判断数据中是否包含{ }，不存在，则是加密数据
		lens := strings.Count(DValue, "") - 1
		if lens < 16 {
			log.Error("[", devID, "] ProcessJsonMsg() error msg : ", DValue, ", len: ", lens)
			return errors.New("error msg.")
		}

		var strHead string
		strHead = DValue[0:16]
		byteHead, _ := hex.DecodeString(strHead)

		myHead.ApiVersion = util.BytesToInt16(byteHead[0:2])
		myHead.ServiceType = util.BytesToInt16(byteHead[2:4])
		myHead.MsgLen = util.BytesToInt16(byteHead[4:6])
		myHead.CheckSum = util.BytesToInt16(byteHead[6:8])
		//log.Info("[", devID, "] ApiVersion: ", myHead.ApiVersion, ", ServiceType: ", myHead.ServiceType, ", MsgLen: ", myHead.MsgLen, ", CheckSum: ", myHead.CheckSum)

		var checkSum uint16
		var strData string
		strData = DValue[16:]
		checkSum = util.CheckSum([]byte(strData))
		if checkSum != myHead.CheckSum {
			log.Error("[", devID, "] ProcessJsonMsg() CheckSum failed, src:", myHead.CheckSum, ", dst: ", checkSum)
			return errors.New("CheckSum failed.")
		}

		if constant.SERVICE_TYPE_UNENCRY == myHead.ServiceType { // 不加密
			DValue = strData
		} else {
			var err_aes error
			DValue, err_aes = util.ECBDecrypt(strData, myKey)
			if nil != err_aes {
				log.Error("[", devID, "] util.ECBDecrypt failed, strData:", strData, ", key: ", myKey, ", error: ", err_aes)
				return err_aes
			}
			//log.Info("[", devID, "] After ECBDecrypt, data.Msg.Value: ", DValue)
		}
	}
	//sendPadDoorUpLogMsg(devID, DValue, "上行设备数据")
	DValue = strings.Replace(DValue, "#", ",", -1)
	//log.Debug("[", devID, "] ProcessJsonMsg() DValue after: ", DValue)

	var esLog entity.EsLogEntiy // 记录日志
	esLog.Operation = "device-mq->DAS"
	// rabbitmq.SendGraylogByMQ("上行数据(device-mq->DAS): dev[%s]; %s", devID, DValue)
	//rabbitmq.SendGraylogByMQ("[%s] ProcessJsonMsg DValue after: %s", devID, DValue)
	if !strings.ContainsAny(DValue, "{ & }") { // 判断数据中是否正确的json，不存在，则是错误数据.
		log.Error("[", devID, "] ProcessJsonMsg() error msg : ", DValue)
		return errors.New("error msg.")
	}

	// 2、解析王力的消息
	//json str 转struct(部份字段)
	var head entity.Header
	if err := json.Unmarshal([]byte(DValue), &head); err != nil {
		log.Error("[", devID, "] Header json.Unmarshal, err=", err)
		return err // break
	}

	var toDevHead entity.MyHeader
	toDevHead.ApiVersion = constant.API_VERSION
	toDevHead.ServiceType = myHead.ServiceType

	// 3、根据命令，分别做业务处理
	switch head.Cmd {
	case constant.Add_dev_user: // 添加设备用户
		{
			//log.Info("[", head.DevId, "] constant.Add_dev_user")
			esLog.Operation += "添加设备用户"

			//1. 回复到APP
			if 1 < head.Ack { // 错误码返回给APP
				//producer.SendMQMsg2APP(head.DevId, DValue)
				rabbitmq.Publish2app([]byte(DValue), head.DevId)
			}
		}
	case constant.Set_dev_user_temp: // 设置临时用户
		{
			//log.Info("[", head.DevId, "] constant.Set_dev_user_temp")
			esLog.Operation += "设置临时用户"

			//1. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Add_dev_user_step: // 新增用户步骤
		{
			//log.Info("[", head.DevId, "] constant.Add_dev_user_step")
			esLog.Operation += "新增用户步骤"

			//1. 判断是否失败，失败则通知APP
			var addUserStep entity.AddDevUserStep
			if err_step := json.Unmarshal([]byte(DValue), &addUserStep); err_step != nil {
				log.Error("[", head.DevId, "] entity.AddDevUserStep json.Unmarshal, err_step=", err_step)
				break
			}

			//if 1 == addUserStep.StepState {
			// 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
			//}
		}
	case constant.Del_dev_user: // 删除设备用户
		{
			//log.Info("[", head.DevId, "] constant.Del_dev_user")
			esLog.Operation += "删除设备用户"

			//1. 回复到APP
			if head.Ack > 1 { // 失败消息直接返回给APP
				//producer.SendMQMsg2APP(head.DevId, DValue)
				rabbitmq.Publish2app([]byte(DValue), head.DevId)
			}
		}
	case constant.Update_dev_user: // 用户更新上报
		{
			//log.Info("[", head.DevId, "] constant.Update_dev_user")
			esLog.Operation += "用户更新上报"

			//1. 更新设备用户操作需要存到mongodb
			if 0 == head.Ack {
				//producer.SendMQMsg2Db(DValue)
				//rabbitmq.Publish2mns([]byte(DValue), "")
				rabbitmq.Publish2pms([]byte(DValue), "")
			}

			//2. 回复设备
			head.Ack = 1
			if toDevice_byte, err := json.Marshal(head); err == nil {
				//log.Info("[", head.DevId, "] constant.Update_dev_user, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Update_dev_user")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}
		}
	case constant.Set_dev_user_para: { // 用户参数设置（0x3B）
		esLog.Operation += "用户参数设置"

		//1. 回复到APP
		if 1 <= head.Ack { // 错误码返回给APP
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	}
	case constant.Sync_dev_user: {// 同步设备用户列表
		//1. 设备用户同步
		//log.Info("[", head.DevId, "] constant.Sync_dev_user")
		esLog.Operation += "同步设备用户列表"

		if 1 == head.Ack {
			//1. 解析Json串
			var syncDevUserRespEx entity.SyncDevUserRespEx
			if err := json.Unmarshal([]byte(DValue), &syncDevUserRespEx); err != nil {
				log.Error("[", head.DevId, "] Header json.Unmarshal, err=", err)
				return err // break
			}

			//2.解析User_List中的值
			syncDevUserResp := entity.SyncDevUserResp {
				Cmd: syncDevUserRespEx.Cmd,
				Ack: syncDevUserRespEx.Ack,
				DevType: syncDevUserRespEx.DevType,
				DevId: syncDevUserRespEx.DevId,
				Vendor: syncDevUserRespEx.Vendor,
				SeqId: syncDevUserRespEx.SeqId,

				UserVer: syncDevUserRespEx.UserVer,
				Num: syncDevUserRespEx.Num,
			}

			for i := 0; i < len(syncDevUserRespEx.UserList); i++ {
				if 10 <= i { // 同步锁用户，一般不超过10个
					break
				}
				var devUser entity.DevUser
				devUser.ParseUser(syncDevUserRespEx.UserList[i])
				syncDevUserResp.UserList = append(syncDevUserResp.UserList, devUser)
			}

			if toPms_byte, err1 := json.Marshal(syncDevUserResp); err1 == nil {
				rabbitmq.Publish2pms(toPms_byte, "")
			} else {
				log.Error("[", head.DevId, "] toPms_byte json.Marshal, err=", err1)
				return err1
			}
		}
	}
	case constant.Remote_open: {// 远程开锁
		//log.Info("[", head.DevId, "] constant.Remote_open")
		esLog.Operation += "远程开锁"

		//1. 回复到APP
		if 0 != head.Ack {
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}

		//2. 远程开门操作需要存到mongodb，开门成功才记录开门记录
		if 1 == head.Ack {
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			//远程开锁作为场景触发条件
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockOpen", 0)
		}
	}
	case constant.Upload_dev_info: // 上传设备信息
		{
			//log.Info("constant.Upload_dev_info")
			esLog.Operation += "上传设备信息"
			//1. 回复设备
			head.Ack = 1
			if toDevice_byte, err := json.Marshal(head); err == nil {
				log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Upload_dev_info")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//2. 设置设备时间
			t := time.Now()
			var toDev entity.SetDeviceTime
			toDev.Cmd = constant.Set_dev_para
			toDev.Ack = 0
			toDev.DevType = head.DevType
			toDev.DevId = head.DevId
			toDev.Vendor = head.Vendor
			toDev.SeqId = 0
			toDev.ParaNo = 7
			toDev.PaValue = t.Unix()
			toDev.Time = strconv.Itoa(int(t.Unix()))
			if toDevice_byte, err := json.Marshal(toDev); err == nil {
				//log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, constant.Set_dev_para to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Upload_dev_info")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//3. 上传设备信息，需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2pms([]byte(DValue), "")
		}
	case constant.Set_dev_para: // 设置设备参数
		{
			//log.Info("[", head.DevId, "] constant.Set_dev_para")
			esLog.Operation += "设置设备参数"

			if 1 == head.Ack {
				rabbitmq.Publish2pms([]byte(DValue), "")
			} else {
				rabbitmq.Publish2app([]byte(DValue), head.DevId)
			}
		}
	case constant.Update_dev_para: // 设备参数更新上报
		{
			//log.Info("[", head.DevId, "] constant.Update_dev_para")
			esLog.Operation += "设备参数更新上报"
			//1. 回复设备
			head.Ack = 1

			if toDevice_byte, err := json.Marshal(head); err == nil {
				//log.Info("[", head.DevId, "] constant.Update_dev_para, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Update_dev_para")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}
			//3. 需要存到mongodb
			rabbitmq.Publish2pms([]byte(DValue), "")

		}
	case constant.Active_Yisuma_SE: // 亿速码安全芯片激活(锁-后台-亿速码-后台->锁)
		{
			//log.Info("[", head.DevId, "] constant.Active_yisuma_SE")
			esLog.Operation += "亿速码安全芯片激活"

			if toDevice_byte, err := json.Marshal(head); err == nil {
				log.Info("[", head.DevId, "] constant.Active_yisuma_SE, resp to device, ", string(toDevice_byte))
				if _, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					//1. 获取参数
					var yisumaActiveSE entity.YisumaActiveSE
					if err_step := json.Unmarshal([]byte(DValue), &yisumaActiveSE); err_step != nil {
						log.Error("[", head.DevId, "] entity.yisumaActiveSE json.Unmarshal, err_step=", err_step)
						break
					}
					//2. SM2算法加密数据
					//2.1 获取privateKey
					privateStr := "607EC530749978DD8D32123B3F2FDF423D1632E6281EB83D083B6375109BB740"
					data, err := hex.DecodeString(privateStr)
					if err != nil {
						log.Error("[", head.DevId, "] privateStr hex.DecodeString, err_step=", err)
						break
					}
					privateKey, err := sm2.RawBytesToPrivateKey(data)
					if err != nil {
						log.Error("[", head.DevId, "] privateKey sm2.RawBytesToPrivateKey, err_step=", err)
						break
					}
					//2.2 封装业务数据
					sign := entity.YisumaSign{UId: yisumaActiveSE.UId, ProjectNo: yisumaActiveSE.ProjectNo, MerchantNo: yisumaActiveSE.MerchantNo, CardChanllege: yisumaActiveSE.Random}
					b, err := json.Marshal(sign)
					if err != nil {
						log.Error("[", head.DevId, "] entity.YisumaSign json.Marshal, err_step=", err)
						break
					}
					//2.3 用私钥加密 得到signature
					r, s, err := sm2.SignToRS(privateKey, nil, b)
					if err != nil {
						log.Error("[", head.DevId, "] r, s, err sm2.SignToRS, err_step=", err)
						break
					}
					signature := strings.ToUpper(hex.EncodeToString(r.Bytes()) + hex.EncodeToString(s.Bytes()))
					//3 模拟https请求
					//3.1 封装业务数据
					httpsParm := entity.YisumaHttpsReq{Body: sign, Signature: signature}
					//3,2 发送https请求
					respBody, err := httpgo.Http2YisumaActive(httpsParm)
					if err != nil {
						log.Error("[", head.DevId, "] cmdto cmdto.Cmd2Yisuma, err_step=", err)
						break
					}
					var jsonRes entity.YisumaHttpsRes
					// 3.3 获取apdu指令
					err1 := json.Unmarshal([]byte(respBody), &jsonRes)
					if err1 != nil {
						log.Error("[", head.DevId, "] YisumaHttpsRes json.Unmarshal, err_step=", err)
						break
					}
					if jsonRes.ResultCode != "0000" {
						log.Error("[", head.DevId, "] resultCode jsonRes.ResultCode, err_step=", err)
						break
					}
					apdu := jsonRes.Apdu
					//4 将命令发到OneNET
					//4.1 将数据组装成json字符串
					apduBody := entity.YisumaActiveApdu{head.Cmd, 1, head.DevType, head.DevId, head.Vendor, head.SeqId, apdu}
					//4.2 通过onenet平台透传
					if toDevice_byte, err := json.Marshal(apduBody); err == nil {
						//log.Info("[", head.DevId, "] constant.Active_yisuma_SE, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go Cmd2Device(head.DevId, strToDevData, "constant.Active_yisuma_SE")
					} else {
						log.Error("toDevice_Data json.Marshal, err=", err)
					}

				} else {
					log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
				}
			}
		}
	case constant.Random_Yisuma_State: //上报随机数
		{
			//log.Info("constant.Random_Yisuma_State")
			esLog.Operation += "上报随机数"
			//1. 获取参数
			var yisumaStateRandom entity.YisumaStateRandom
			if err_step := json.Unmarshal([]byte(DValue), &yisumaStateRandom); err_step != nil {
				log.Error("[", head.DevId, "] entity.YisumaStateRandom json.Unmarshal, err_step=", err_step)
				break
			}
			//2. 存入redis
			go redis.SetDeviceYisumaRandomfromPool(head.DevId, yisumaStateRandom.Random)
			random, _ := redis.GetDeviceYisumaRandomfromPool(head.DevId)
			log.Info("redis.SetDeviceYisumaRandomfromPool=============" + random)

		}
	case constant.Soft_reset: // 软件复位
		{
			//log.Info("[", head.DevId, "] constant.Soft_reset")
			esLog.Operation += "软件复位"
			//1. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Factory_reset: // 恢复出厂设置
		{
			//log.Info("[", head.DevId, "] constant.Factory_reset")
			esLog.Operation += "恢复出厂设置"
			//1. 回复设备
			head.Ack = 1
			if toDevice_byte, err := json.Marshal(head); err == nil {
				//log.Info("[", head.DevId, "] constant.Factory_reset, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Factory_reset")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//2. 重置设备用户列表mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2pms([]byte(DValue), "")

			//3. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Upload_open_log: // 门锁开门日志上报
		{
			//log.Info("[", head.DevId, "] constant.Upload_open_log")
			esLog.Operation += "门锁开门日志上报"

			//1. 需要存到mongodb
			rabbitmq.Publish2pms([]byte(DValue), "")

			//mns推送到app
			rabbitmq.Publish2mns([]byte(DValue), "")

			//2. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)

			//可视锁开门触发场景
			ParseOpenLog(DValue)
		}
	case constant.Noatmpt_alarm: // 非法操作报警
		{
			//log.Info("[", head.DevId, "] constant.Noatmpt_alarm")
			esLog.Operation += "非法操作报警"
			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType,"lockAlarm", 1)
		}
	case constant.Forced_break_alarm: // 强拆报警
		{
			//log.Info("[", head.DevId, "] constant.Forced_break_alarm")
			esLog.Operation += "强拆报警"
			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockAlarm", 1)
		}
	case constant.Fakelock_alarm: // 假锁报警
		{
			//log.Info("[", head.DevId, "] constant.Fakelock_alarm")
			esLog.Operation += "假锁报警"
			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockAlarm", 1)
		}
	case constant.Nolock_alarm: // 门未关报警
		{
			//log.Info("[", head.DevId, "] constant.Nolock_alarm")
			esLog.Operation += "门未关报警"
			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockAlarm", 1)
		}
	case constant.Gas_Alarm: //	燃气报警
		{
			esLog.Operation += "燃气报警"
			msg := entity.RangeHoodAlarm{}
			if err := json.Unmarshal([]byte(DValue), &msg); err != nil {
				log.Warning("ProcessJsonMsg RangeHoodAlarm json.Unmarshal() error = ", err)
				return err
			}
			pushMsgForSceneTrigger(&msg)

			// 燃气告警通知APP
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
			rabbitmq.Publish2mns([]byte(DValue), "")

			// 燃气告警，存redis缓存
			go redis.SetDevGasAlarmState(head.DevId, 1)
		}
	case constant.Low_battery_alarm: // 锁体的电池，低电量报警
		{
			//log.Info("[", head.DevId, "] constant.Low_battery_alarm")
			esLog.Operation += "锁体的电池，低电量报警"
			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockAlarm", 1)
		}
	case constant.Infrared_alarm: // 人体感应报警（infra红外感应)
		{
			//log.Info("[", head.DevId, "] constant.Infrared_alarm")
			esLog.Operation += "人体感应报警"

			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
			SendLockMsgForSceneTrigger(head.DevId, head.DevType, "lockAlarm", 1)
		}
	case constant.Lock_PIC_Upload: // 视频锁图片上报
		{
			//log.Info("[", head.DevId, "] constant.Lock_PIC_Upload")
			esLog.Operation += "视频锁图片上报"

			//1. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2pms([]byte(DValue), "")
		}
	case constant.Upload_lock_active: // 锁激活状态上报
		{
			//log.Info("[", head.DevId, "] constant.Upload_lock_active")
			esLog.Operation += "锁激活状态上报"
			//1. 解析锁激活上报包
			var lockActive entity.DeviceActiveResp
			if err_lockActive := json.Unmarshal([]byte(DValue), &lockActive); err_lockActive != nil {
				log.Error("[", head.DevId, "] entity.Upload_lock_active json.Unmarshal, err_lockActive=", err_lockActive)
				break
			}

			//2. 回复设备
			lockActive.Ack = 1
			/*t := time.Now()
			lockActive.Time = t.Unix()*/
			if toDevice_byte, err := json.Marshal(lockActive); err == nil {
				//log.Info("[", head.DevId, "] constant.Upload_lock_active, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Upload_lock_active")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//3. 锁唤醒，存入redis
			go redis.SetActTimePool(lockActive.DevId, int64(1))

			//4. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)

			//5. 通知深圳中控，设备在线状态
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")

			if lockActive.Time > 0 {
				esLog.RetMsg = "设备上线"
			} else {
				esLog.RetMsg = "设备离线"
			}
		}
	case constant.Real_Video: // 实时视频
		{
			//log.Info("[", head.DevId, "] constant.Upload_lock_active")
			esLog.Operation += "实时视频"

			//1. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Set_Wifi: // Wifi设置
		{
			//log.Info("[", head.DevId, "] constant.Set_Wifi")
			esLog.Operation += "Wifi设置"
			//1. 回复到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			//rabbitmq.Publish2app([]byte(DValue), head.DevId)

			//2. 需要存到mongodb
			// producer.SendMQMsg2Db(DValue)
		}
	case constant.Door_Call: // 门铃呼叫
		{
			//log.Info("[", head.DevId, "] constant.Door_Call")
			esLog.Operation += "门铃呼叫"
			//1. 回复设备
			head.Ack = 1
			if toDevice_byte, err := json.Marshal(head); err == nil {
				//log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Upload_lock_active")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//2. 推到APP
			// producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

			//3. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2mns([]byte(DValue), "")
			rabbitmq.Publish2pms([]byte(DValue), "")
		}
	case constant.Door_State: // 锁状态上报
		{
			//log.Info("[", head.DevId, "] constant.Door_State")
			esLog.Operation += "锁状态上报"
			//1. 回复设备
			head.Ack = 1
			if toDevice_byte, err := json.Marshal(head); err == nil {
				//log.Info("[", head.DevId, "] constant.Door_State, resp to device, ", string(toDevice_byte))
				var strToDevData string
				if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
					toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
					toDevHead.MsgLen = (uint16)(strings.Count(strToDevData, "") - 1)

					buf := new(bytes.Buffer)
					binary.Write(buf, binary.BigEndian, toDevHead)
					strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
				}

				go Cmd2Device(head.DevId, strToDevData, "constant.Door_State")
			} else {
				log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
			}

			//2. 推到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)

			//3. 需要存到mongodb
			//producer.SendMQMsg2Db(DValue)
			rabbitmq.Publish2pms([]byte(DValue), "")
		}
	case constant.Notify_F_Upgrade: // 通知前板升级（APP—后台—>锁）
		{
			//log.Info("[", head.DevId, "] constant.Notify_F_Upgrade")
			esLog.Operation += "通知前板升级"

			//1. 推到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Notify_B_Upgrade: // 通知后板升级（APP—后台—>锁）
		{
			//log.Info("[", head.DevId, "] constant.Notify_B_Upgrade")
			esLog.Operation += "通知后板升级"

			//1. 推到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Get_Upgrade_FileInfo: // 锁查询升级固件包信息
		{
			//log.Info("[", head.DevId, "] constant.Get_Upgrade_FileInfo")
			esLog.Operation += "锁查询升级固件包信息"

			var upQuery entity.UpgradeQuery
			if err := json.Unmarshal([]byte(DValue), &upQuery); err != nil {
				log.Error("UpgradeQuery json.Unmarshal, err=", err)
				return err
			}

			// 获取升级包信息
			go GetUpgradeFileInfo(head.DevId, head.DevType, head.SeqId, upQuery.Part)
		}
	case constant.Download_Upgrade_File: // 锁下载固件升级包（锁—>后台，分包传输）
		{
			//log.Info("[", head.DevId, "] constant.Download_Upgrade_File")
			esLog.Operation += "锁下载固件升级包"
			var upReq entity.UpgradeReq
			if err := json.Unmarshal([]byte(DValue), &upReq); err != nil {
				log.Error("[", head.DevId, "] UpgradeReq json.Unmarshal, err=", err)
				return err
			}

			// 获取文件传输给设备
			log.Info("[", head.DevId, "] constant.Get_Upgrade_FileInfo, TransferFileData")
			TransferFileData(head.DevId, head.DevType, head.SeqId, upReq.Offset, upReq.FileName, upReq.Part)
		}
	case constant.Upload_F_Upgrade_State: // 前板上传升级状态
		{
			//log.Info("[", head.DevId, "] constant.Upload_F_Upgrade_State")
			esLog.Operation += "前板上传升级状态"

			//1. 推到APP
			//producer.SendMQMsg2APP(head.DevId, DValue)
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.Upload_B_Upgrade_State: // 后板上传升级状态
		{
			//log.Info("[", head.DevId, "] constant.Upload_B_Upgrade_State")
			esLog.Operation += "后板上传升级状态"

			//1. 推到APP
			rabbitmq.Publish2app([]byte(DValue), head.DevId)
		}
	case constant.PadDoor_Weather: // 平板锁天气信息透传至mns
		{
			//log.Info("[", head.DevId, "] constant.Door_Pad_Weather")
			esLog.Operation += "平板锁天气信息透传至mns"

			//推送到mns
			rabbitmq.Publish2mns([]byte(DValue), "")
		}
	case constant.Set_AIPad_Reboot_Time: // 设置中控网关定时参数
		{
			//log.Info("[", head.DevId, "] constant.Set_AIPad_Reboot_Time")
			esLog.Operation += "设置中控网关定时参数"

			//1. 重启时间存储到mysql
			rabbitmq.Publish2wonlyms([]byte(DValue), "")
		}
	case constant.PadDoor_Num_Upload, constant.PadDoor_Num_Reset: // 平板锁人流检测上报
		{
			//log.Info("[", head.DevId, "] constant.PadDoor_Num")
			esLog.Operation += "平板锁人流检测上报"

			// TODO:JHHE 2020-05-19
			//1. 场景触发
			rabbitmq.Publish2pms([]byte(DValue), "")
		}
	case constant.RangeHood_Control: // 油烟机档位控制
		{
			//log.Info("[", head.DevId, "] constant.Range_Hood_Control")
			esLog.Operation += "油烟机档位控制回复"

			//1. 推到APP
			if 1 == head.Ack { // 油烟机回应包
				rabbitmq.Publish2app([]byte(DValue), head.DevId)
			}
		}
	case constant.RangeHood_Ctrl_Query: // 油烟机档位查询
		{
			//log.Info("[", head.DevId, "] constant.RangeHood_Ctrl_Query")
			esLog.Operation += "油烟机档位查询回复"

			if 1 == head.Ack { // 油烟机回应包
				rabbitmq.Publish2app([]byte(DValue), head.DevId)
			}
		}
	case constant.RangeHood_Lock_Query: {
		//log.Info("[", head.DevId, "] constant.RangeHood_Lock_Query")
		esLog.Operation += "油烟机绑定锁列表查询回复"
		rabbitmq.Publish2mns(util.Str2Bytes(DValue), "")
	}
	case constant.Scene_Trigger: //中控闹钟触发爱岗场景
	    {
			//log.Info("[", head.DevId, "] constant.Scene_Trigger")
			esLog.Operation += "中控闹钟触发爱岗场景"
	    	rabbitmq.Publish2pms([]byte(DValue), "")
	    }
	default:
		log.Info("[", head.DevId, "] Default, Cmd=", head.Cmd)
	}

	esLog.DeviceId = devID
	esLog.Vendor = "general"
	esLog.ThirdPlatform = "王力RabbitMQ"
	esLog.RawData = string(DValue)
	esData, err_ := json.Marshal(esLog)
	if err_ != nil {
		log.Warningf("ProcessJsonMsg > json.Marshal > %s", err_)
		return err_
	}
	rabbitmq.SendGraylogByMQ("%s", esData)

	return nil
}

func SendLockMsgForSceneTrigger(devId, devType , alarmType string, alarmFlag int) {
	var msg entity.Feibee2AutoSceneMsg

	msg.Cmd = constant.Scene_Trigger
	msg.Ack = 0
	msg.DevType = devType
	msg.DevId = devId

	msg.TriggerType = 0
	msg.Time = int(time.Now().Unix())
	msg.AlarmFlag = alarmFlag
	msg.AlarmType = alarmType

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warning("createMsg2pmsForSceneTrigger json.Marshal() error = ", err)
		return
	}

	//producer.SendMQMsg2PMS(string(data))
	rabbitmq.Publish2Scene(data, "")
}

func ParseOpenLog(rawData string) {
	var openLog entity.UploadOpenLockLog
	if err := json.Unmarshal([]byte(rawData), &openLog); err != nil {
		log.Warningf("ParseOpenLog > json.Unmarshal > %s", err)
		return
	}

	HandleOpenLog(&openLog)
}

func HandleOpenLog(openLog *entity.UploadOpenLockLog) {
	for i,_ := range openLog.LogList {
		if openLog.LogList[i].MainOpen == 13 { //手动开门-室内开门
			SendLockMsgForSceneTrigger(openLog.DevId, openLog.DevType, "lockOpen", 0)
		} else {
			SendLockMsgForSceneTrigger(openLog.DevId, openLog.DevType, "lockOpen", 1)
		}
	}
}

