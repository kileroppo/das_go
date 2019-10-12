package feibee2srv

import (
	"errors"
	"strconv"

	"encoding/json"

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
		log.Error("NewFeibeeData() unmarshal error = ", err)
		return feibeeData, err
	}

	return feibeeData, nil
}

func (f FeibeeData) isDataValid() bool {
	if f.Status != "" && f.Ver != "" {
		switch f.Code {
		case 3, 4, 5, 7, 12:
			if len(f.Msg) > 0 {
				return true
			}
		case 2:
			if len(f.Records) > 0 {
				return true
			}
		case 15, 32:
			if len(f.Gateway) > 0 {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func (f FeibeeData) push2mq() error {

	switch f.Code {
	case 2:
		f.code2MsgPush2app()
		f.push2pms()

	case 3, 4, 5, 12:
		f.push2app()
		f.push2pms()

	//其他消息推送到db
	default:
		f.push2pms()
	}
	return nil
}

func (f FeibeeData) push2app() {

	sendOneMsg := func(index int) {

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

	for i := 0; i < len(f.Msg); i++ {
		sendOneMsg(i)
	}

}

func (f FeibeeData) code2MsgPush2app() {

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

		if isManualOpMsg(f, index) {
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
	}

	for i := 0; i < len(f.Records); i++ {
		sendOneMsg(i)
	}

}

func (f FeibeeData) push2pms() {

	sendOneMsg := func(index int) {

		if f.Code == 2 {
			if !isAlarmMsg(f, index) && !isManualOpMsg(f, index) {
				return
			}
		}

		var data []byte
		var err error

		if isAlarmMsg(f, index) {
			data, err = json.Marshal(autoScene2pmsDataFormat(f, index))
		} else {
			data, err = json.Marshal(msg2pmsDataFormat(f, index))
		}

		if err != nil {
			log.Error("One Msg push2pms() error = ", err)
			return
		}

		producer.SendMQMsg2PMS(string(data))
	}

	msgNums := 1
	switch f.Code {
	case 2:
		msgNums = len(f.Records)
	case 3, 4, 5, 7, 12:
		msgNums = len(f.Msg)
	case 15, 32:
		msgNums = len(f.Gateway)
	}

	for i := 0; i < msgNums; i++ {
		sendOneMsg(i)
	}

}

func msg2appDataFormat(data FeibeeData, index int) (res entity.Feibee2AppMsg, bindid string) {

	switch data.Code {
	case 2:
		res = entity.Feibee2AppMsg{
			Cmd:     0xfb,
			Ack:     0,
			DevType: devTypeConv(data.Records[index].Deviceid, data.Records[index].Zonetype),
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
			DevType: devTypeConv(data.Msg[index].Deviceid, data.Msg[index].Zonetype),
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
	case 2:
		if data.Records[index].Value == "00" {
			res.OpType = "devOff"
		} else if data.Records[index].Value == "01" {
			res.OpType = "devOn"
		} else if data.Records[index].Value == "02" {
			res.OpType = "devStop"
		}
	case 3:
		res.OpType = "newDevice"
	case 4:
		res.OpType = "newOnline"
		res.Battery = 0xff
	case 5:
		res.OpType = "devDelete"
		res.Battery = 0xff
	case 7:
		if data.Msg[index].Onoff == 1 {
			res.OpType = "devRemoteOn"
		} else if data.Msg[index].Onoff == 0 {
			res.OpType = "devRemoteOff"
		} else if data.Msg[index].Onoff == 2 {
			res.OpType = "devRemoteStop"
		}

	case 12:
		res.OpType = "devNewName"
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
		res.DevType = devTypeConv(data.Records[index].Deviceid, data.Records[index].Zonetype)
		res.DevId = data.Records[index].Uuid
		res.Records = []entity.FeibeeRecordsMsg{data.Records[index]}
		res.Records[index].Devicetype = res.DevType
	case 3, 4, 5, 7, 12:
		res.DevType = devTypeConv(data.Msg[index].Deviceid, data.Msg[index].Zonetype)
		res.DevId = data.Msg[index].Uuid
		res.Msg = []entity.FeibeeDevMsg{data.Msg[index]}
		res.Msg[index].Devicetype = res.DevType
	case 15, 32:
		res.Gateway = []entity.FeibeeGatewayMsg{data.Gateway[index]}
	}
	return
}

func autoScene2pmsDataFormat(data FeibeeData, index int) (res entity.FeibeeAutoScene2pmsMsg) {
	res.Cmd = 0xf1
	res.Ack = 0
	res.Vendor = "feibee"
	res.SeqId = 1
	res.DevType = devTypeConv(data.Records[index].Deviceid, data.Records[index].Zonetype)
	res.Devid = data.Records[index].Uuid
	res.TriggerType = 0
	res.Zone = "hz"

	return
}

func isAlarmMsg(data FeibeeData, index int) bool {

	if data.Code == 2 && len(data.Records) > 0 {
		cid := data.Records[index].Cid
		aid := data.Records[index].Aid
		if cid == 1280 && aid == 128 {
			return true
		}

		if (cid == 1 && aid == 33) || (cid == 1 && aid == 53) {
			return true
		}

		if (cid == 1 && aid == 32) || (cid == 1 && aid == 62) {
			return true
		}
	}

	return false
}

func isManualOpMsg(data FeibeeData, index int) bool {

	if data.Code == 2 && len(data.Records) > 0 {

		if data.Records[index].Cid == 6 && data.Records[index].Aid == 0 {
			return true
		}
	}

	return false
}

func devTypeConv(devId, zoneType int) string {
	pre := strconv.FormatInt(int64(devId), 16)
	tail := strconv.FormatInt(int64(zoneType), 16)
	lenPre := len(pre)
	lenTail := len(tail)

	if lenPre < 4 {
		for i := 0; i < 4-lenPre; i++ {
			pre = "0" + pre
		}
	}

	if lenTail < 4 {
		for i := 0; i < 4-lenTail; i++ {
			tail = "0" + tail
		}
	}
	return "0x" + pre + tail
}
