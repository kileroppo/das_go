package mqtt2srv

import "das/procLock"

type MqttJob struct {
	rawData []byte
}

func NewMqttJob(rawData []byte) MqttJob {
	return MqttJob{
		rawData: rawData,
	}
}

func (o MqttJob) Handle() {
	procLock.ParseData(o.rawData)
}
