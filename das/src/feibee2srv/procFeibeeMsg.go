package feibee2srv

import (
	"encoding/json"
	"errors"

	"../core/entity"
	"../core/log"
	"../rmq/producer"
)

type FeibeeData struct {
	entity.FeibeeData
}

func ProcessFeibeeMsg(feibeeData FeibeeData) (err error) {

	//feibee数据合法性检查
	if !feibeeData.isDataValid() {
		err := errors.New("the feibee message's structure is invalid")
		log.Debug(err)
		return err
	}

	//feibee数据推送到mq
	if err := feibeeData.push2mq(); err != nil {
		log.Error("feibeeData.push2mq() error = ", err)
		return err
	}

	return nil
}

func NewFeibeeData(data []byte) (FeibeeData, error) {
	var feibeeData FeibeeData

	if err := json.Unmarshal(data, &feibeeData); err != nil {
		log.Error("json.Unmarshal() error = ", err)
		return feibeeData, err
	}

	return feibeeData, nil
}

func (f FeibeeData) isDataValid() bool {
	if f.Status != "" && f.Ver != "" {
		switch f.Code {
		case 3, 4, 5, 12:
			if len(f.Msg) > 0 {
				return true
			}
		case 2, 15, 32:
			return true

		default:
			return false
		}
	}
	return false
}

func (f FeibeeData) push2mq() error {

	switch f.Code {
	//设备入网数据推送到app和db
	case 3, 4, 5, 12:

		if err := f.push2mq2app(); err != nil {
			log.Error("f.push2mq2app() error = ", err)
			return err
		}

		if err := f.push2mq2db2(); err != nil {
			log.Error("f.push2mq2db() error = ", err)
			return err
		}

	//其他消息推送到db
	default:
		if err := f.push2mq2db2(); err != nil {
			log.Error("f.push2mq2db() error = ", err)
			return err
		}
	}
	return nil
}

func (f FeibeeData) push2mq2app() error {

	feibee2appMsg := dataFormat(f)

	data, err := json.Marshal(feibee2appMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return err
	}
	msg := string(data)
	producer.SendMQMsg2Db(msg)
	producer.SendMQMsg2APP(f.Msg[0].Bindid, msg)
	return nil
}

func (f FeibeeData) push2mq2db2() error {

	data, err := json.Marshal(f)

	if err != nil {
		log.Error("json.Marshal() error = ", err)
	}

	producer.SendMQMsg2Db2(string(data))
	return nil
}

func dataFormat(data FeibeeData) entity.Feibee2AppMsg {
	msg := data.Msg[0]
	res := entity.Feibee2AppMsg{
		Cmd:     0xfb,
		Ack:     0,
		DevType: msg.Devicetype,
		Devid:   msg.Uuid,
		Vendor:  "feibee",
		SeqId:   1,

		Note:      msg.Name,
		Deviceuid: msg.Deviceuid,
		Online:    msg.Online,
		Battery:   msg.Battery,
	}

	switch data.Code {
	case 3:
		res.OpType = "newdevice"
	case 4:
		res.OpType = "newonline"
		res.Battery = 0xff
	case 5:
		res.OpType = "devdelete"
		res.Battery = 0xff
	case 12:
		res.OpType = "devnewname"
		res.Battery = 0xff
	}

	return res

}
