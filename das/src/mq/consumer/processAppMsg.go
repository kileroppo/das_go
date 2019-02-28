package consumer

import (
	"../../core/httpgo"
	"encoding/json"
	"../../core/log"
	"../../core/constant"
	"../producer"
		)

type Header struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
}

type RespOneNET struct {
	RespErrno int		`json:"errno"`
	RespError string	`json:"error"`
}

type AppMsg struct {
	pri string
}

/*
*	处理APP发送过来的命令消息
*
*/
func (p *AppMsg) ProcessAppMsg() error {
	log.Info("ProcessAppMsg process msg from app: ", p.pri)

	// 1、解析消息
	//json str 转struct(部份字段)
	var head Header
	if err := json.Unmarshal([]byte(p.pri), &head); err != nil {
		log.Error("ProcessAppMsg json.Unmarshal Header error, err=", err)
		return err
	}

	// 将命令发到OneNET
	imei := head.DevId

	respStr, err := httpgo.Http2OneNET_write(imei, p.pri)
	if "" != respStr && nil == err {
		var respOneNET RespOneNET
		if err := json.Unmarshal([]byte(respStr), &respOneNET); err != nil {
			log.Error("ProcessAppMsg json.Unmarshal RespOneNET error, err=", err)
			return err
		}

		if 0 != respOneNET.RespErrno {
			head.Ack = constant.Device_Resp_TimeOut

			if toApp_str, err := json.Marshal(head); err == nil {
				log.Info("[", head.DevId, "] ProcessAppMsg, device timeout, resp to APP, ", string(toApp_str))

				// 推到APP
				producer.SendMQMsg2APP(head.DevId, string(toApp_str))
			}
		}
	}

	// time.Sleep(time.Second)
	return nil
}
