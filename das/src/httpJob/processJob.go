package httpJob

import (
	"fmt"
	"encoding/json"
	"../core/constant"
	"../mq/producer"
	"../core/redis"
	"../core/log"

)

type Serload struct {
	pri string
}

type Header struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Time int64			`json:"time"`
}

/*
*	处理OneNET推送过来的消息
*
*/
func (p *Serload) ProcessJob() error {
	// 处理OneNET推送过来的消息
	log.Debug("process msg from onenet: ", p.pri)

	// 1、解析消息
	//json str 转struct(部份字段)
	var head Header
	if err := json.Unmarshal([]byte(p.pri), &head); err == nil {
		fmt.Println("================json str 转struct==")
		fmt.Println(head)
		fmt.Println(head.Cmd)
	}

	// 2、根据命令，分别做业务处理
	switch head.Cmd {
	case constant.Add_dev_user:	// 添加设备用户
		{
			fmt.Printf("constant.Add_dev_user")

			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 添加设备用户操作需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Set_dev_user_temp: // 设置临时用户
		{
			fmt.Printf("constant.Set_dev_user_temp")

			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	case constant.Add_dev_user_step: // 新增用户步骤
		{
			fmt.Printf("constant.Add_dev_user_step")
		}
	case constant.Del_dev_user: // 删除设备用户
		{
			fmt.Printf("constant.Del_dev_user")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	case constant.Update_dev_user: // 用户更新上报
		{
			fmt.Printf("constant.Update_dev_user")
			//1. 更新设备用户操作需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Sync_dev_user: // 同步设备用户列表
		{
			//1. 设备用户同步
			fmt.Printf("constant.Sync_dev_user")
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Remote_open: // 远程开锁
		{
			fmt.Printf("constant.Remote_open")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 远程开门操作需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Upload_dev_info: // 上传设备信息
		{
			fmt.Printf("constant.Upload_dev_info")
			//2. 上传设备信息，需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Volume_level: // 音量等级
		{
			fmt.Printf("constant.Volume_level")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Open_mode: // 常开模式
		{
			fmt.Printf("constant.Open_mode")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Passwd_switch: // 密码开关
		{
			fmt.Printf("constant.Passwd_switch")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Remote_switch: // 远程开关
		{
			fmt.Printf("constant.Remote_switch")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Sin_mul: // 开门模式
		{
			fmt.Printf("constant.Sin_mul")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)

			//2. 需要存到mongodb
			if 1 == head.Ack {
				producer.SendMQMsg2Db(p.pri)
			}
		}
	case constant.Set_time: // 设置时间
		{
			fmt.Printf("constant.Set_time")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	case constant.Update_dev_para: // 设备参数更新上报
		{
			fmt.Printf("constant.Update_dev_para")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Soft_reset: // 软件复位
		{
			fmt.Printf("constant.Soft_reset")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	case constant.Factory_reset: // 恢复出厂设置
		{
			fmt.Printf("constant.Factory_reset")
			//1. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	case constant.Upload_open_log: // 门锁开门日志上报
		{
			fmt.Printf("constant.Upload_open_log")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Noatmpt_alarm: // 非法操作报警
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Forced_break_alarm: // 强拆报警
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Fakelock_alarm: // 假锁报警
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Nolock_alarm: // 门未关报警
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Low_battery_alarm: // 锁体的电池，低电量报警
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 需要存到mongodb
			producer.SendMQMsg2Db(p.pri)
		}
	case constant.Upload_lock_active: // 锁激活状态上报
		{
			fmt.Printf("constant.Noatmpt_alarm")
			//1. 锁唤醒，存入redis
			// TODO:JHHE
			redis.SetData(head.DevId, head.Time)

			//2. 回复到APP
			producer.SendMQMsg2APP(head.DevId, p.pri)
		}
	default:
		fmt.Printf("Default, Cmd=", head.Cmd)
	}

	return nil
}
