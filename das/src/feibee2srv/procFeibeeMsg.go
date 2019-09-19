package feibee2srv

import (
	"encoding/json"
	"errors"

	"../core/entity"
	"../core/log"
	"../mq"
)

type FeibeeData struct {
	entity.FeibeeData
}

var MQPool mq.MQChannelPool

func init() {

	MQPool = mq.NewMQChannelPool()
	MQPool.Init("amqp://wonly:Wl2016822@139.196.221.163:5672/")

}

func ProcessFeibeeMsg(pushData []byte) (err error) {

	var feibeeData FeibeeData
	if feibeeData, err = NewFeibeeData(pushData); err != nil {
		log.Error("NewFeibeeData() error=", err)
		return
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
	//设备入网数据推送到app和db
	case 3:

		if err := f.push2mq2app(); err != nil {
			log.Error("f.push2mq2app() error = ", err)
			return err
		}

		if err := f.push2mq2db(); err != nil {
			log.Error("f.push2mq2db() error = ", err)
			return err
		}

	//其他消息推送到db
	default:
		if err := f.push2mq2db(); err != nil {
			log.Error("f.push2mq2db() error = ", err)
			return err
		}
	}
	return nil
}

func (f FeibeeData) push2mq2app() error {

	//feibee2appMsg := dataFormat(f.Msg[0])
	//
	//data, err := json.Marshal(feibee2appMsg)
	//if err != nil {
	//	log.Error("json.Marshal() error = ", err)
	//	return err
	//}

	//producer.SendMQMsg2APP(f.Msg[0].Bindid, string(data))
	return nil
}

func (f FeibeeData) push2mq2db() error {

	data, err := json.Marshal(f)

	if err != nil {
		log.Error("json.Marshal() error = ", err)
	}

	sendMQMsg2Db(data)
	return nil
}

func dataFormat(msg entity.FeibeeDevMsg) entity.Feibee2AppMsg {

	return entity.Feibee2AppMsg{
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

}

func sendMQMsg2Db(data []byte) {
	conf := mq.MQConfig{
		Exchange:     "Device2Db_ex2",
		ExchangeType: "fanout",
		RoutingKey:   "Device2Db_queue2",
	}
	MQPool.Product(data, conf)
}
