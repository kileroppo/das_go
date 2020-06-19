package mqtt2srv

import "das/procLock"

type SmartPadData struct {
	rawData string
	devID   string
}

func NewSmartPadJob(rawData string, devID string) SmartPadData {
	return SmartPadData{
		rawData: rawData,
		devID:   devID,
	}
}

func (d SmartPadData) Handle() {
	procLock.ProcessJsonMsg(d.rawData, d.devID)
}
