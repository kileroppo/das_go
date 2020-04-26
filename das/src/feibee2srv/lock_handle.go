package feibee2srv

import "das/core/entity"

type LockAlarmHandle struct {
	data *entity.FeibeeData
}

func (ah *LockAlarmHandle) PushMsg() {

}

func (ah *LockAlarmHandle) createAlarmMsg2pms() (msg entity.FeibeeLockAlarmMsg){
    return
}

func (ah *LockAlarmHandle) parseAlarmType() (alarmType int) {
	//rawVal := ah.data.Records[0].Value



	return
}