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
	case 2, 3, 4, 5, 12:

		if err := f.push2mq2app(); err != nil {
			log.Error("f.push2mq2app() error = ", err)
			return err
		}

		if err := f.push2mq2db(); err != nil {
			log.Error("f.push2mq2db() error = ", err)
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

	feibee2appMsg, bindid := dataFormat(f)
	data, err := json.Marshal(feibee2appMsg)
	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return err
	}
	producer.SendMQMsg2APP(bindid, string(data))

	return nil
}

func (f FeibeeData) push2mq2db() error {
	feibee2appMsg, bindid := dataFormat(f)

	feibee2dbMsg := entity.Feibee2DBMsg{
		feibee2appMsg,
		bindid,
	}

	data, err := json.Marshal(feibee2dbMsg)

	if err != nil {
		log.Error("json.Marshal() error = ", err)
		return err
	}

	producer.SendMQMsg2Db(string(data))

	return nil
}

func (f FeibeeData) push2mq2db2() error {

	data, err := json.Marshal(f)

	if err != nil {
		log.Error("json.Marshal() error = ", err)
	}

	producer.SendMQMsg2Db(string(data))
	return nil
}

func dataFormat(data FeibeeData) (res entity.Feibee2AppMsg, bindid string) {

	switch data.Code {
	case 2:
		res = entity.Feibee2AppMsg{
			Cmd:     0xfb,
			Ack:     0,
			DevType: data.Records[0].Devicetype,
			Devid:   data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   1,

			Note:      "",
			Deviceuid: data.Records[0].Deviceuid,
			Online:    1,
			Battery:   0xff,
			Time:      data.Records[0].Uptime,
		}
		bindid = data.Records[0].Bindid
	default:
		res = entity.Feibee2AppMsg{
			Cmd:     0xfb,
			Ack:     0,
			DevType: data.Msg[0].Devicetype,
			Devid:   data.Msg[0].Uuid,
			Vendor:  "feibee",
			SeqId:   1,

			Note:      data.Msg[0].Name,
			Deviceuid: data.Msg[0].Deviceuid,
			Online:    data.Msg[0].Online,
			Battery:   data.Msg[0].Battery,
			Time:      -1,
		}
		bindid = data.Msg[0].Bindid
	}

	switch data.Code {
	case 2:
		if data.Records[0].Value == "00" {
			res.OpType = "switchclose"
		} else {
			res.OpType = "switchopen"
		}
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

	return

}
