package httpJob

import (
	"../core/constant"
	"../core/entity"
	"../core/httpgo"
	"../core/log"
	"../core/redis"
	"../core/util"
	"../mq/producer"
	"../upgrade"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Serload struct {
	pri string
}


// 转换8进制utf-8字符串到中文
// eg: `\346\200\241` -> 怡
func convertOctonaryUtf8(in string) string {
	s := []byte(in)
	reg := regexp.MustCompile(`\\[0-7]{3}`)

	out := reg.ReplaceAllFunc(s,
		func(b []byte) []byte {
			i, _ := strconv.ParseInt(string(b[1:]), 8, 0)
			return []byte{byte(i)}
		})
	return string(out)
}
/*
*	处理OneNET推送过来的消息
*
*/
func (p *Serload) ProcessJob() error {
	// 处理OneNET推送过来的消息
	log.Info("process msg from onenet before: ", p.pri)

	// 1、解析OneNET消息
	var data entity.OneNETData
	if err := json.Unmarshal([]byte(p.pri), &data); err != nil {
		log.Error("OneNETData json.Unmarshal, err=", err)
		return nil
	}

	switch data.Msg.Msgtype {
	case 2:	// 设备上下线消息(type=2)
		{
			log.Info("OneNET Upload_lock_active, imei=", data.Msg.Imei, ", time=", data.Msg.At/1000)

			var nTime int64
			nTime = 0
			if 1 == data.Msg.Status {			// 设备上线
				nTime = data.Msg.At/1000
			} else if 0 == data.Msg.Status {	// 设备离线
				nTime = 0
			}

			//1. 锁状态，存入redis
			redis.SetData(data.Msg.Imei, nTime)

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
	case 1:	// 数据点消息(type=1)，
		{
			// httpgo.Http2OneNET_write(data.Msg.Imei, "Hei, man, what are you doing?")

			/*ret, _ := hex.DecodeString(data.Msg.Value)
			log.Debugf("中文：%s", ret)
			log.Debug("中文：", ret)*/
			myKey := util.MD52Bytes(data.Msg.Imei)

			// 增加二进制包头，以及加密的包体
			// 1、 获取包头部分 8个字节
			var myHead entity.MyHeader
			if !strings.ContainsAny(data.Msg.Value, "{ & }") { // 判断数据中是否包含{ }，不存在，则是加密数据
				log.Debug("[", data.Msg.Imei, "] get aes data: ", data.Msg.Value)
				lens := strings.Count(data.Msg.Value,"") - 1
				if lens < 16 {
					log.Error("[", data.Msg.Imei, "] ProcessJob() error msg : ", data.Msg.Value, ", len: ", lens)
					return errors.New("error msg.")
				}

				var strHead string
				strHead = data.Msg.Value[0:16]
				byteHead, _ := hex.DecodeString(strHead)

				myHead.ApiVersion = util.BytesToInt16(byteHead[0:2])
				myHead.ServiceType = util.BytesToInt16(byteHead[2:4])
				myHead.MsgLen = util.BytesToInt16(byteHead[4:6])
				myHead.CheckSum = util.BytesToInt16(byteHead[6:8])
				log.Info("[", data.Msg.Imei, "] ApiVersion: ", myHead.ApiVersion, ", ServiceType: ", myHead.ServiceType, ", MsgLen: ", myHead.MsgLen, ", CheckSum: ", myHead.CheckSum)

				var checkSum uint16
				var strData string
				strData = data.Msg.Value[16:]
				checkSum = util.CheckSum([]byte(strData))
				if checkSum != myHead.CheckSum {
					log.Error("[", data.Msg.Imei, "] ProcessJob() CheckSum failed, src:", myHead.CheckSum, ", dst: ", checkSum)
					return errors.New("CheckSum failed.")
				}

				if constant.SERVICE_TYPE_UNENCRY == myHead.ServiceType { // 不加密
					data.Msg.Value = strData
				} else {
					var err_aes error
					data.Msg.Value, err_aes = util.ECBDecrypt(strData, myKey)
					if nil != err_aes {
						log.Error("[", data.Msg.Imei, "] util.ECBDecrypt failed, strData:", strData, ", key: ", myKey, ", error: ", err_aes)
						return err_aes
					}
					log.Info("[", data.Msg.Imei, "] After ECBDecrypt, data.Msg.Value: ", data.Msg.Value)
				}
			}

			data.Msg.Value = strings.Replace(data.Msg.Value, "#", ",", -1)
			log.Debug("[", data.Msg.Imei, "] ProcessJob() data.Msg.Value after: ", data.Msg.Value)

			// 2、解析王力的消息
			//json str 转struct(部份字段)
			var head entity.Header
			if err := json.Unmarshal([]byte(data.Msg.Value), &head); err != nil {
				log.Error("[", head.DevId, "] Header json.Unmarshal, err=", err)
				break
			}

			var toDevHead entity.MyHeader
			toDevHead.ApiVersion = constant.API_VERSION
			toDevHead.ServiceType = myHead.ServiceType

			// 3、根据命令，分别做业务处理
			switch head.Cmd {
			case constant.Add_dev_user: // 添加设备用户
				{
					log.Info("[", head.DevId, "] constant.Add_dev_user")

					//1. 回复到APP
					if 1 < head.Ack { // 错误码返回给APP
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}
				}
			case constant.Set_dev_user_temp: // 设置临时用户
				{
					log.Info("[", head.DevId, "] constant.Set_dev_user_temp")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Add_dev_user_step: // 新增用户步骤
				{
					log.Info("[", head.DevId, "] constant.Add_dev_user_step")

					//1. 判断是否失败，失败则通知APP
					var addUserStep entity.AddDevUserStep
					if err_step := json.Unmarshal([]byte(data.Msg.Value), &addUserStep); err_step != nil {
						log.Error("[", head.DevId, "] entity.AddDevUserStep json.Unmarshal, err_step=", err_step)
						break
					}

					if 1 == addUserStep.StepState {
						// 回复到APP
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}
				}
			case constant.Del_dev_user: // 删除设备用户
				{
					log.Info("[", head.DevId, "] constant.Del_dev_user")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Update_dev_user: // 用户更新上报
				{
					log.Info("[", head.DevId, "] constant.Update_dev_user")
					//1. 更新设备用户操作需要存到mongodb
					if 0 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}

					//2. 回复设备
					head.Ack = 1
					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Update_dev_user, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}
				}
			case constant.Sync_dev_user: // 同步设备用户列表
				{
					//1. 设备用户同步
					log.Info("[", head.DevId, "] constant.Sync_dev_user")
					if 1 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Remote_open: // 远程开锁
				{
					log.Info("[", head.DevId, "] constant.Remote_open")
					//1. 回复到APP
					if 0 != head.Ack {
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}

					//2. 远程开门操作需要存到mongodb
					if 0 != head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Upload_dev_info: // 上传设备信息
				{
					log.Info("constant.Upload_dev_info")
					//1. 回复设备
					head.Ack = 1
					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
						// go httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
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
					toDev.Time = t.Unix()
					if toDevice_byte, err := json.Marshal(toDev); err == nil {
						log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, constant.Set_dev_para to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//3. 上传设备信息，需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Set_dev_para: // 设置设备参数
				{
					log.Info("[", head.DevId, "] constant.Set_dev_para")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//2. 需要存到mongodb
					if 1 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Update_dev_para: // 设备参数更新上报
				{
					log.Info("[", head.DevId, "] constant.Update_dev_para")
					//1. 回复设备
					head.Ack = 1

					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Update_dev_para, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//2. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//3. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Soft_reset: // 软件复位
				{
					log.Info("[", head.DevId, "] constant.Soft_reset")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Factory_reset: // 恢复出厂设置
				{
					log.Info("[", head.DevId, "] constant.Factory_reset")
					//1. 回复设备
					head.Ack = 1
					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Factory_reset, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//2. 重置设备用户列表mongodb
					producer.SendMQMsg2Db(data.Msg.Value)

					//3. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Upload_open_log: // 门锁开门日志上报
				{
					log.Info("[", head.DevId, "] constant.Upload_open_log")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Noatmpt_alarm: // 非法操作报警
				{
					log.Info("[", head.DevId, "] constant.Noatmpt_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Forced_break_alarm: // 强拆报警
				{
					log.Info("[", head.DevId, "] constant.Forced_break_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Fakelock_alarm: // 假锁报警
				{
					log.Info("[", head.DevId, "] constant.Fakelock_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Nolock_alarm: // 门未关报警
				{
					log.Info("[", head.DevId, "] constant.Nolock_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Low_battery_alarm: // 锁体的电池，低电量报警
				{
					log.Info("[", head.DevId, "] constant.Low_battery_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Infrared_alarm:	// 人体感应报警（infra红外感应)
				{
					log.Info("[", head.DevId, "] constant.Infrared_alarm")

					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Lock_PIC_Upload:	// 视频锁图片上报
				{
					log.Info("[", head.DevId, "] constant.Lock_PIC_Upload")

					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Upload_lock_active: // 锁激活状态上报
				{
					log.Info("[", head.DevId, "] constant.Upload_lock_active")

					//1. 解析锁激活上报包
					var lockActive entity.DeviceActive
					if err_lockActive := json.Unmarshal([]byte(data.Msg.Value), &lockActive); err_lockActive != nil {
						log.Error("[", head.DevId, "] entity.Upload_lock_active json.Unmarshal, err_lockActive=", err_lockActive)
						break
					}

					var lockTime int64
					lockTime = lockActive.Time

					//2. 回复设备
					lockActive.Ack = 1
					t := time.Now()
					lockActive.Time = t.Unix()
					if toDevice_byte, err := json.Marshal(lockActive); err == nil {
						log.Info("[", head.DevId, "] constant.Upload_lock_active, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//3. 锁唤醒，存入redis
					redis.SetData(lockActive.DevId, lockTime)

					//4. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Real_Video:	// 实时视频
				{
					log.Info("[", head.DevId, "] constant.Upload_lock_active")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Set_Wifi:	// Wifi设置
				{
					log.Info("[", head.DevId, "] constant.Set_Wifi")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//2. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Door_Call:	// 门铃呼叫
				{
					log.Info("[", head.DevId, "] constant.Door_Call")
					//1. 回复设备
					head.Ack = 1
					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Upload_dev_info, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//2. 推到APP
					// producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//3. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Door_State:	// 锁状态上报
				{
					log.Info("[", head.DevId, "] constant.Door_State")
					//1. 回复设备
					head.Ack = 1
					if toDevice_byte, err := json.Marshal(head); err == nil {
						log.Info("[", head.DevId, "] constant.Door_State, resp to device, ", string(toDevice_byte))
						var strToDevData string
						if strToDevData, err = util.ECBEncrypt(toDevice_byte, myKey); err == nil {
							toDevHead.CheckSum = util.CheckSum([]byte(strToDevData))
							toDevHead.MsgLen =  (uint16)(strings.Count(strToDevData,"") - 1)

							buf := new(bytes.Buffer)
							binary.Write(buf, binary.BigEndian, toDevHead)
							strToDevData = hex.EncodeToString(buf.Bytes()) + strToDevData
						}

						go httpgo.Http2OneNET_write(head.DevId, strToDevData)
					} else {
						log.Error("[", head.DevId, "] toDevice_str json.Marshal, err=", err)
					}

					//2. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//3. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Notify_F_Upgrade:	// 通知前板升级（APP—后台—>锁）
				{
					log.Info("[", head.DevId, "] constant.Notify_F_Upgrade")

					//1. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Notify_B_Upgrade: // 通知后板升级（APP—后台—>锁）
				{
					log.Info("[", head.DevId, "] constant.Notify_B_Upgrade")

					//1. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Get_Upgrade_FileInfo: // 锁查询升级固件包信息
				{
					log.Info("[", head.DevId, "] constant.Get_Upgrade_FileInfo")

					var upQuery entity.UpgradeQuery
					if err := json.Unmarshal([]byte(p.pri), &upQuery); err != nil {
						log.Error("UpgradeQuery json.Unmarshal, err=", err)
						return err
					}

					// 获取升级包信息
					upgrade.GetUpgradeFileInfo(head.DevId, head.DevType, head.SeqId, upQuery.Part)
				}
			case constant.Download_Upgrade_File: // 锁下载固件升级包（锁—>后台，分包传输）
				{
					log.Info("[", head.DevId, "] constant.Download_Upgrade_File")
					var upReq entity.UpgradeReq
					if err := json.Unmarshal([]byte(data.Msg.Value), &upReq); err != nil {
						log.Error("[", head.DevId, "] UpgradeReq json.Unmarshal, err=", err)
						return err
					}

					// 获取文件传输给设备
					log.Info("[", head.DevId, "] constant.Get_Upgrade_FileInfo, TransferFileData")
					upgrade.TransferFileData(head.DevId, head.DevType, head.SeqId, upReq.Offset, upReq.FileName, upReq.Part)
				}
			case constant.Upload_F_Upgrade_State: // 前板上传升级状态
				{
					log.Info("[", head.DevId, "] constant.Upload_F_Upgrade_State")

					//1. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Upload_B_Upgrade_State: // 后板上传升级状态
				{
					log.Info("[", head.DevId, "] constant.Upload_B_Upgrade_State")

					//1. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			default:
				log.Info("[", head.DevId, "] Default, Cmd=", head.Cmd)
			}
		}
	}
	return nil
}
