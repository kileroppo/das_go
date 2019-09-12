package httpJob

import (
	"../core/constant"
	"../core/log"
	"regexp"
	"strconv"
	"../feibee2srv"
)

type Serload struct {
	DValue 	string
	Imei   	string
	MsgFrom string	// 消息来自哪个平台
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
*	处理OneNET，Andlink, Telecom，NB-wifi，feibee推送过来的消息
*
*/
func (p *Serload) ProcessJob() error {
	//
	switch p.MsgFrom {
	case constant.FEIBEE_MSG:	// feibee
		{
			log.Debug("ProcessJob() from FEIBEE_MSG.")
			return feibee2srv.ProcessFeibeeMsg(p.DValue, p.Imei)
		}
	case constant.NBIOT_MSG: 	// OneNET，Andlink, Telecom，NB-wifi
		{
			log.Debug("ProcessJob() from NBIOT_MSG.")
			//return ProcessNbMsg(p.DValue, p.Imei)
			return nil
		}
	}

	return nil
}
