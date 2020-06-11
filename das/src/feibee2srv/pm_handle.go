package feibee2srv

import (
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"fmt"
	"strconv"
	"time"
)

//PMHandle handle the feibee's PM device original data, which arming to parse and transfer data to app
type PMHandle struct {
	data     *entity.FeibeeData
	Protocal FbDevProtocal
	msg2app  entity.Feibee2DevMsg
	msg2pms  entity.Feibee2AlarmMsg
}

func (pm *PMHandle) PushMsg() {
	err := pm.decodeHeader()
	if err != nil {
		log.Warning("PMHandle.PushMsg > %s", err)
		return
	}

	pm.pushMsg()
}

func (pm *PMHandle) initMsg() {
	pm.msg2app.Cmd = 0xfb
	pm.msg2app.DevId = pm.data.Records[0].Uuid
	pm.msg2app.Vendor = "feibee"
	pm.msg2app.DevType = devTypeConv(pm.data.Records[0].Deviceid, pm.data.Records[0].Zonetype)
	pm.msg2app.Time = int(time.Now().Unix())

	pm.msg2pms.Cmd = 0xfc
	pm.msg2pms.DevId = pm.msg2app.DevId
	pm.msg2pms.Vendor = pm.msg2app.Vendor
	pm.msg2pms.Time = pm.msg2app.Time
	pm.msg2pms.DevType = pm.msg2app.DevType
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
	if pm.Protocal.Value[2] != 0x21 {
		return
	}

	pm.initMsg()
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
	pm.push2pmsForSave(opType, opValue, opFlag)
	pm.push2app(opType, opValue)
	return
}

func (pm *PMHandle) decodeCO2() {
	attrs := make([]uint16, pm.Protocal.DataNum)
    opType, opValue := "", ""
    opFlag := int(0)

	for i:=0;i<len(attrs);i++ {
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
		pm.push2pmsForSave(opType, opValue, opFlag)
		pm.push2app(opType, opValue)
	}
}

func (pm *PMHandle) push2pmsForSave(opType, opValue string, opFlag int) {
	pm.msg2pms.AlarmType = opType
	pm.msg2pms.AlarmValue = opValue
	pm.msg2pms.AlarmFlag = opFlag

	data, err := json.Marshal(pm.msg2pms)
	if err != nil {
		log.Warningf("PMHandle.push2pmsForSave > json.Marshal > %s", err)
		return
	}

	rabbitmq.Publish2pms(data, "")
}

func (pm *PMHandle) push2app(opType, opValue string) {
	pm.msg2app.OpType = opType
	pm.msg2app.OpValue = opValue

	data, err := json.Marshal(pm.msg2app)
	if err != nil {
		log.Warningf("PMHandle.push2app > json.Marshal > %s", err)
		return
	}
	rabbitmq.Publish2app(data, pm.msg2app.DevId)
}
