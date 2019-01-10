package consumer

import (
	"fmt"
	"../../core/httpgo"
	"encoding/json"
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
	fmt.Println("ProcessAppMsg process msg from app: ", p.pri)

	// 1、解析消息
	//json str 转struct(部份字段)
	var head Header
	if err := json.Unmarshal([]byte(p.pri), &head); err == nil {
		fmt.Println("ProcessAppMsg================json str 转struct==")
		fmt.Println(head)
		fmt.Println(head.Cmd)
	}

	// 将命令发到OneNET
	imei := head.DevId

	httpgo.Http2OneNET_write(imei, p.pri)

	// time.Sleep(time.Second)
	return nil
}
