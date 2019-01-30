package httpJob

import (
		"encoding/json"
	"../core/constant"
	"../mq/producer"
	"../core/redis"
	"../core/log"
	"../core/httpgo"
		"regexp"
	"strconv"
	_ "time"
)

type Serload struct {
	pri string
}

type OneMsg struct {
	At int64			`json:"at"`
	Msgtype int			`json:"type"`				// 数据点消息(type=1)，设备上下线消息(type=2)
	Value string		`json:"value"`
	Imei string			`json:"imei"`
	Dev_id int			`json:"dev_id"`
	Ds_id string		`json:"ds_id"`
	Status int			`json:"status"`				// 设备上下线标识：0-下线, 1-上线
	Login_type int		`json:"login_type"`
}
type OneNETData struct {
	Msg_signature string	`json:"msg_signature"`
	Nonce string			`json:"nonce"`
	Msg OneMsg 				`json:"msg"`
}

type Header struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	SeqId int			`json:"seqId"`
}

type DeviceActive struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	SeqId int			`json:"seqId"`

	Time int64			`json:"time"`
}

type SetDeviceTime struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	SeqId int			`json:"seqId"`

	paraNo int			`json:"paraNo"`
	value int64			`json:"value"`
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
	log.Info("process msg from onenet: ", p.pri)

	// 1、解析OneNET消息
	var data OneNETData
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
			var toApp DeviceActive
			toApp.Cmd = constant.Upload_lock_active
			toApp.Ack = 0
			toApp.DevType = ""
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
			/*ret, _ := hex.DecodeString(data.Msg.Value)
			log.Debugf("中文：%s", ret)
			log.Debug("中文：", ret)*/

			// 2、解析王力的消息
			//json str 转struct(部份字段)
			var head Header
			if err := json.Unmarshal([]byte(data.Msg.Value), &head); err != nil {
				log.Error("Header json.Unmarshal, err=", err)
				break
			}

			// 3、根据命令，分别做业务处理
			switch head.Cmd {
			case constant.Add_dev_user: // 添加设备用户
				{
					log.Info("constant.Add_dev_user")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Set_dev_user_temp: // 设置临时用户
				{
					log.Info("constant.Set_dev_user_temp")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Add_dev_user_step: // 新增用户步骤
				{
					log.Info("constant.Add_dev_user_step")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Del_dev_user: // 删除设备用户
				{
					log.Info("constant.Del_dev_user")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Update_dev_user: // 用户更新上报
				{
					log.Info("constant.Update_dev_user")
					//1. 回复设备
					head.Ack = 1
					if toDevice_str, err := json.Marshal(head); err == nil {
						log.Info("constant.Del_dev_user, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}

					//2. 更新设备用户操作需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Sync_dev_user: // 同步设备用户列表
				{
					//1. 设备用户同步
					log.Info("constant.Sync_dev_user")
					if 1 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Remote_open: // 远程开锁
				{
					log.Info("constant.Remote_open")
					//1. 回复到APP
					if 1 == head.Ack {
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}

					//2. 远程开门操作需要存到mongodb
					if 1 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Upload_dev_info: // 上传设备信息
				{
					log.Info("constant.Upload_dev_info")
					//1. 回复设备
					head.Ack = 1
					if toDevice_str, err := json.Marshal(head); err == nil {
						log.Info("constant.Upload_dev_info, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}

					//2. 设置设备时间
					/*
					t := time.Now()
					var toDev SetDeviceTime
					toDev.Cmd = constant.Set_dev_para
					toDev.Ack = 0
					toDev.DevType = head.DevType
					toDev.DevId = head.DevId
					toDev.SeqId = head.SeqId
					toDev.paraNo = 7
					toDev.value = t.Unix()
					if toDevice_str, err := json.Marshal(toDev); err == nil {
						log.Info("constant.Upload_dev_info, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}
					*/

					//3. 上传设备信息，需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Set_dev_para: // 设置设备参数
				{
					log.Info("constant.Set_dev_para")
					//1. 回复到APP
					if 1 == head.Ack {
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}

					//2. 需要存到mongodb
					if 1 == head.Ack {
						producer.SendMQMsg2Db(data.Msg.Value)
					}
				}
			case constant.Update_dev_para: // 设备参数更新上报
				{
					log.Info("constant.Update_dev_para")
					//1. 回复设备
					head.Ack = 1

					if toDevice_str, err := json.Marshal(head); err == nil {
						log.Info("constant.Update_dev_para, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}

					//2. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//3. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Soft_reset: // 软件复位
				{
					log.Info("constant.Soft_reset")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Factory_reset: // 恢复出厂设置
				{
					log.Info("constant.Factory_reset")
					//1. 重置设备用户列表mongodb
					producer.SendMQMsg2Db(data.Msg.Value)

					//2. 回复到APP
					if 1 == head.Ack {
						producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
					}
				}
			case constant.Upload_open_log: // 门锁开门日志上报
				{
					log.Info("constant.Upload_open_log")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Noatmpt_alarm: // 非法操作报警
				{
					log.Info("constant.Noatmpt_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Forced_break_alarm: // 强拆报警
				{
					log.Info("constant.Forced_break_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Fakelock_alarm: // 假锁报警
				{
					log.Info("constant.Fakelock_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Nolock_alarm: // 门未关报警
				{
					log.Info("constant.Nolock_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Low_battery_alarm: // 锁体的电池，低电量报警
				{
					log.Info("constant.Low_battery_alarm")
					//1. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Upload_lock_active: // 锁激活状态上报
				{
					log.Info("constant.Upload_lock_active")
					//1. 回复设备
					head.Ack = 1
					if toDevice_str, err := json.Marshal(head); err == nil {
						log.Info("constant.Upload_lock_active, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}

					//2. 锁唤醒，存入redis
					//json str 转struct(部份字段)
					var devAct DeviceActive
					if err := json.Unmarshal([]byte(data.Msg.Value), &devAct); err != nil {
						log.Error("Header json.Unmarshal, err=", err)
					}
					redis.SetData(devAct.DevId, devAct.Time)

					//3. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Real_Video:
				{
					log.Info("constant.Upload_lock_active")

					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)
				}
			case constant.Set_Wifi:
				{
					log.Info("constant.Set_Wifi")
					//1. 回复到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//2. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			case constant.Door_Call:
				{
					log.Info("constant.Door_Call")
					//1. 回复设备
					head.Ack = 1
					if toDevice_str, err := json.Marshal(head); err == nil {
						log.Info("constant.Door_Call, resp to device, ", string(toDevice_str))
						httpgo.Http2OneNET_write(head.DevId, string(toDevice_str))
					} else {
						log.Error("toDevice_str json.Marshal, err=", err)
					}

					//2. 推到APP
					producer.SendMQMsg2APP(head.DevId, data.Msg.Value)

					//3. 需要存到mongodb
					producer.SendMQMsg2Db(data.Msg.Value)
				}
			default:
				log.Info("Default, Cmd=", head.Cmd)
			}
		}
	}
	return nil
}
