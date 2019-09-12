package feibee2srv

import (
	"encoding/json"

	"../core/entity"
	"../core/log"
)

type FeibeeData struct {
	entity.FeibeeData
}

func NewFeibeeData(data []byte) (FeibeeData, error) {
	var feibeeData FeibeeData

	if err := json.Unmarshal(data, &feibeeData); err != nil {
		log.Warning("json.Unmarshal() warning = feibee数据解析失败")
		return feibeeData, err
	}

	return feibeeData, nil
}

func ProcessFeibeeMsg(pushData []byte) (err error) {

	var feibeeData FeibeeData
	if feibeeData, err = NewFeibeeData(pushData); err != nil {
		log.Error("NewFeibeeData() error=", err)
		return
	}

	//feibee数据合法性检查
	if !feibeeData.isDataValid() {
		log.Warning("the feibee message'type is invalid")
		return nil
	}

	//feibee数据格式化为本地协议
	if err := feibeeData.dataFormat(); err != nil {
		return err
	}

	//feibee数据推送到mq
	if err := feibeeData.push2mq(); err != nil {
		return err
	}

	return nil
}

func (f FeibeeData) isDataValid() bool {
	if f.Status != "" && f.Ver != "" {
		if f.Msg != nil || f.Gateway != nil {
			switch f.Code {
			case 3, 4, 5, 12, 15, 32:
				return true
			default:
				return false
			}
		}
	}
	return false
}

func (f FeibeeData) dataFormat() error {

	return nil
}

func (f FeibeeData) push2mq() error {

	return nil
}
