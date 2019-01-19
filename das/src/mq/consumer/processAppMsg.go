package consumer

import (
		"../../core/httpgo"
	"encoding/json"
	"../../core/log"
)

type Header struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
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
		log.Error("ProcessAppMsg json.Unmarshal error, err=", err)
		return err
	}

	// 将命令发到OneNET
	imei := head.DevId

	httpgo.Http2OneNET_write(imei, p.pri)

	// time.Sleep(time.Second)
	return nil
}
