package feibee2srv

import (
	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"das/filter"
	"fmt"
	"strconv"
	"time"
)

//PMHandle handle the feibee's PM device original data, which arming to parse and transfer data to app
type PMHandle struct {
	data     *entity.FeibeeData
	Protocal FbDevProtocal
}

func (pm *PMHandle) PushMsg() {
	err := pm.decodeHeader()
	if err != nil {
		log.Warning("PMHandle.PushMsg > %s", err)
		return
	}

	pm.pushMsg()
}

func (pm *PMHandle) decodeHeader() (err error) {
	err = pm.Protocal.Decode(pm.data.Records[0].Orgdata)
	if err != nil {
		err = fmt.Errorf("PMHandle.decode > pm.Protocal.Decode > %w", err)
		return
	}

	return
}

func (pm *PMHandle) pushMsg() {
	opType, opValue := "", ""
	opFlag := 0

	if len(pm.Protocal.Value) < 5 {
		return
	}

	opFlag = int(uint16(pm.Protocal.Value[3]) | (uint16(pm.Protocal.Value[4]) << 8))

	switch pm.Protocal.Cluster {
	case Fb_PM_PM25:
		opType = "PM2.5"
		opValue = strconv.Itoa(opFlag)
	case Fb_PM_VOC:
		opType = "VOC"
		opValue = fmt.Sprintf("%0.2f", float64(opFlag)/float64(100))
	case Fb_PM_Formaldehyde:
		opType = "formaldehyde"
		opValue = fmt.Sprintf("%0.2f", float64(opFlag)/float64(100))
	case Fb_PM_Temperature:
		opType = "temperature"
		opValue = fmt.Sprintf("%0.2f", float64(opFlag)/float64(100))
	case Fb_PM_Humidity:
		opType = "humidity"
		opValue = fmt.Sprintf("%0.2f", float64(opFlag)/float64(100))
	case Fb_PM_CO2:
		pm.decodeCO2()
		return
	default:
		return
	}
	pm.push2pms(opType, opValue, opFlag)
	pm.push2app(opType, opValue)
	return
}

func (pm *PMHandle) decodeCO2() {
	attrs := make([]uint16, pm.Protocal.DataNum)
	opType, opValue := "", ""
	opFlag := int(0)

	for i := 0; i < len(attrs); i++ {
		attrs[i] = uint16(pm.Protocal.Value[i*5]) | (uint16(pm.Protocal.Value[i*5+1]) << 8)
		switch attrs[i] {
		case 0x0000:
			opType = "CO2"
			opFlag = int(uint16(pm.Protocal.Value[i*5+3]) | (uint16(pm.Protocal.Value[i*5+1+3]) << 8))
			opValue = strconv.Itoa(opFlag)
		case 0x0001:
			continue
			//opType = "CO2level"
			//opFlag = int(pm.Protocal.Value[i*5+3])
			//opValue = fmt.Sprintf("%0.2f", float64(opFlag)/float64(100))
		}
		pm.push2pms(opType, opValue, opFlag)
		pm.push2app(opType, opValue)
	}
}

func (pm *PMHandle) push2pms(opType, opValue string, opFlag int) {
	_, triggerFlag := filter.SensorFilter(pm.data.Records[0].Uuid, opType, opValue, opFlag)

	msg2pms := entity.Feibee2AutoSceneMsg{
		Header: entity.Header{
			Cmd:     constant.Scene_Trigger,
			DevId:   pm.data.Records[0].Uuid,
			DevType: devTypeConv(pm.data.Records[0].Deviceid, pm.data.Records[0].Zonetype),
			Vendor:  "feibee",
		},
		Time:        int(time.Now().Unix()),
		TriggerType: 0,
		AlarmFlag:   opFlag,
		AlarmType:   opType,
		AlarmValue:  opValue,
		SceneId:     "",
		Zone:        "",
	}

	var data []byte
	var err error
	if triggerFlag {
		data, err = json.Marshal(msg2pms)
		if err != nil {
			log.Warningf("PMHandle.push2pms > Trigger > json.Marshal > %s", err)
			return
		}
		rabbitmq.Publish2Scene(data, "")
	}

	msg2pms.Cmd = constant.Device_Sensor_Msg
	data, err = json.Marshal(msg2pms)
	if err != nil {
		log.Warningf("PMHandle.push2pms > Alarm > json.Marshal > %s", err)
		return
	}

	rabbitmq.Publish2pms(data, "")
}

func (pm *PMHandle) push2app(opType, opValue string) {
	msg2app := entity.Feibee2DevMsg{
		Header:        entity.Header{
			DevId: pm.data.Records[0].Uuid,
			DevType:devTypeConv(pm.data.Records[0].Deviceid, pm.data.Records[0].Zonetype),
			Vendor:"feibee",
		},
		OpType:        opType,
		OpValue:       opValue,
		Time:          int(time.Now().Unix()),
	}

	data, err := json.Marshal(msg2app)
	if err != nil {
		log.Warningf("PMHandle.push2app > json.Marshal > %s", err)
		return
	}
	rabbitmq.Publish2app(data, msg2app.DevId)
}
