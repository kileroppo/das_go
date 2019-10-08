package feibee2srv

import (
	"encoding/json"
	"errors"

	"../core/entity"
	"../core/log"
	"../rmq/producer"
)

//type FeibeeData struct {
//	entity.FeibeeData
//}

type FeibeeData entity.FeibeeData

func ProcessFeibeeMsg(rawData []byte) (err error) {

	feibeeData, err := NewFeibeeData(rawData)
	if err != nil {
		return err
	}

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

		f.push2mq2app()

		//f.push2mq2db()

		f.push2pms()

	//其他消息推送到db
	default:
		f.push2pms()
	}
	return nil
}

func (f FeibeeData) push2mq2app() {

	sendOneMsg := func(index int) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("sendOneMsg() error = ", err)
			}
		}()

		if isAlarmMsg(f, index) {
			devAlarm := NewDevAlarm(f, index)
			if devAlarm == nil {
				log.Error("该报警设备类型未支持")
				return
			}

			datas, err := devAlarm.GetMsg2app(index)
			if err != nil {
				log.Error("alarmMsg2app error = ", err)
				return
			}

			if len(datas) <= 0 {
				return
			}

			for _, data := range datas {
				if len(data) > 0 {
					producer.SendMQMsg2APP(f.Records[index].Bindid, string(data))
					producer.SendMQMsg2Db(string(data))
				}
			}
			return
		}
		feibee2appMsg, bindid := msg2appDataFormat(f, index)

		data2app, err := json.Marshal(feibee2appMsg)
		if err != nil {
			log.Error("One Msg push2mq2app() error = ", err)
		} else {
			producer.SendMQMsg2APP(bindid, string(data2app))
		}

		data2db, err := json.Marshal(entity.Feibee2DBMsg{
			feibee2appMsg,
			bindid,
		})
		if err != nil {
			log.Error("One Msg push2mq2app() error = ", err)
		} else {
			producer.SendMQMsg2Db(string(data2db))
		}

	}
	msgNums := 0
	switch f.Code {
	case 2:
		msgNums = len(f.Records)
	default:
		msgNums = len(f.Msg)
	}

	for i := 0; i < msgNums; i++ {
		go sendOneMsg(i)
	}

}

func (f FeibeeData) push2pms() {

	sendOneMsg := func(index int) {
		msg := msg2pmsDataFormat(f, index)

		data, err := json.Marshal(msg)
		if err != nil {
			log.Error("One Msg push2pms() error = ", err)
		}
		producer.SendMQMsg2PMS(string(data))
	}
	msgNums := 1
	switch f.Code {
	case 2:
		msgNums = len(f.Records)
	case 3, 4, 5, 12:
		msgNums = len(f.Msg)
	}

	for i := 0; i < msgNums; i++ {
		go sendOneMsg(i)
	}

}

func msg2appDataFormat(data FeibeeData, index int) (res entity.Feibee2AppMsg, bindid string) {

	switch data.Code {
	case 2:
		res = entity.Feibee2AppMsg{
			Cmd:     0xfb,
			Ack:     0,
			DevType: data.Records[index].Devicetype,
			Devid:   data.Records[index].Uuid,
			Vendor:  "feibee",
			SeqId:   1,

			Note:      "",
			Deviceuid: data.Records[index].Deviceuid,
			Online:    1,
			Battery:   0xff,
			Time:      data.Records[index].Uptime,
		}
		bindid = data.Records[index].Bindid
	default:
		res = entity.Feibee2AppMsg{
			Cmd:     0xfb,
			Ack:     0,
			DevType: data.Msg[index].Devicetype,
			Devid:   data.Msg[index].Uuid,
			Vendor:  "feibee",
			SeqId:   1,

			Note:      data.Msg[index].Name,
			Deviceuid: data.Msg[index].Deviceuid,
			Online:    data.Msg[index].Online,
			Battery:   data.Msg[index].Battery,
			Time:      -1,
		}
		bindid = data.Msg[index].Bindid
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

	return
}

func msg2pmsDataFormat(data FeibeeData, index int) (res entity.Feibee2PMS) {
	res.Cmd = 0xfa
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1

	res.FeibeeData = entity.FeibeeData(data)

	switch data.Code {
	case 2:
		res.DevType = data.Records[index].Devicetype
		res.DevId = data.Records[index].Uuid
		res.Records = []entity.FeibeeRecordsMsg{data.Records[index]}
	case 3, 4, 5, 12:
		res.DevType = data.Msg[index].Devicetype
		res.DevId = data.Msg[index].Uuid
		res.Msg = []entity.FeibeeDevMsg{data.Msg[index]}
	default:
	}

	return
}

func isAlarmMsg(data FeibeeData, index int) bool {

	if data.Code == 2 && len(data.Records) > 0 {
		rawData := data.Records[index].Orgdata
		if rawData[0:4] == "700B" || rawData[0:4] == "7015" {
			return true
		}
	}

	return false
}
