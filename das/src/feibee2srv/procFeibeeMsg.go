package feibee2srv

import (
	"encoding/json"
	"errors"

	"../core/entity"
	"../core/log"
)

type FeibeeData struct {
	data entity.FeibeeData
}

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

	//feibee数据推送到MQ
	feibeeData.push2MQ()

	return nil
}

func NewFeibeeData(data []byte) (FeibeeData, error) {
	var feibeeData FeibeeData

	if err := json.Unmarshal(data, &feibeeData.data); err != nil {
		log.Error("NewFeibeeData() unmarshal error = ", err)
		return feibeeData, err
	}

	return feibeeData, nil
}

func (f FeibeeData) isDataValid() bool {
	if f.data.Status != "" && f.data.Ver != "" {
		switch f.data.Code {
		case 3, 4, 5, 7, 12:
			if len(f.data.Msg) > 0 {
				return true
			}
		case 2:
			if len(f.data.Records) > 0 {
				return true
			}
		case 15, 32:
			if len(f.data.Gateway) > 0 {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func (f FeibeeData) push2MQ() {
	//飞比推送数据条数 分条处理
	datas := splitFeibeeMsg(f.data)

	for _, data := range datas {
		msgHandle := MsgHandleFactory(data)
		if msgHandle == nil {
			return
		}
		msgHandle.PushMsg()
	}

}

func splitFeibeeMsg(data entity.FeibeeData) (datas []entity.FeibeeData) {

	switch data.Code {
	case 3, 4, 5, 7, 12:
		datas = make([]entity.FeibeeData, len(data.Msg))
		for i := 0; i < len(data.Msg); i++ {
			datas[i].Msg = []entity.FeibeeDevMsg{
				data.Msg[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	case 2:
		datas = make([]entity.FeibeeData, len(data.Records))
		for i := 0; i < len(data.Records); i++ {
			datas[i].Records = []entity.FeibeeRecordsMsg{
				data.Records[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	case 32:
		datas = make([]entity.FeibeeData, len(data.Gateway))
		for i := 0; i < len(data.Gateway); i++ {
			datas[i].Gateway = []entity.FeibeeGatewayMsg{
				data.Gateway[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	}

	return
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
