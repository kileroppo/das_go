package feibee2srv

import (
	"das/core/redis"
	"errors"
	"fmt"
	"strconv"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	ErrFbProtocalStruct = errors.New("the feibee original data's structure was invalid")
	ErrFbProtocalLen    = errors.New("the feibee original data's lens was wrong")
	ErrFbProtocalType   = errors.New("the feibee original data's type was not supported")
)

type FbLockHandle struct {
	data     *entity.FeibeeData
	Protocal FbDevProtocal
}

func (fh *FbLockHandle) PushMsg() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("FbLockHandle.PushMsg > %s", err)
		}
	}()

	if err := fh.decodeValue(); err == nil {
		return
	}

	if err := fh.decodeOrg(); err != nil {
		log.Warningf("FbLockHandle.PushMsg > %s", err)
	}
}

func (fh *FbLockHandle) decodeValue() (err error) {
	msgType, ok := otherMsgTyp[getSpMsgKey(fh.data.Records[0].Cid, fh.data.Records[0].Aid)]
	if !ok {
		return fmt.Errorf("pushMsg > %w", ErrFbProtocalType)
	}

	switch msgType {
	case FbZigbeeLockActivation:
		fh.FbLockActivationDecode()
	case FbZigbeeLockEnable:
		fh.FbLockEnableDecode()
	default:
		return fmt.Errorf("pushMsg > %w", ErrFbProtocalType)
	}

	return nil
}

//todo: feibeelock handle
func (fh *FbLockHandle) decodeOrg() (err error) {
	err = fh.Protocal.Decode(fh.data.Records[0].Orgdata)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.decodeOrg > fh.Protocal.Decode > %w", err)
		return
	}

	switch fh.Protocal.Cluster {
	case Fb_Lock_Alarm:
		err = fh.FbLockAlarmDecode()
	case Fb_Lock_Batt:
		err = fh.FbLockBattDecode()
	case Fb_Lock_Onoff:
		err = fh.FbLockOnoffDecode()
	}

	if err != nil {
		err = fmt.Errorf("FbLockHandle.decodeOrg > %w", err)
	}

	return
}

func (fh *FbLockHandle) FbLockEnableDecode() {
	val, err := strconv.ParseInt(fh.data.Records[0].Value, 16, 32)
	if err != nil {
		log.Warningf("FbLockHandle.FbLockEnableDecode > strconv.ParseInt > %s", err)
		return
	}

	msg := entity.DeviceActive{
		Cmd:     0x46,
		Ack:     1,
		DevType: "WonlyFBlock",
		DevId:   fh.data.Records[0].Uuid,
		Vendor:  "feibee",
		SeqId:   0,
		Signal:  0,
		Time:    val,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("FbLockHandle.FbLockEnableDecode > json.Marshal > %s", err)
		return
	}

	rabbitmq.Publish2app(data, fh.data.Records[0].Uuid)
	rabbitmq.Publish2pms(data, "")
}

func (fh *FbLockHandle) FbLockActivationDecode() {

}

func (fh *FbLockHandle) FbLockAlarmDecode() (err error) {
	if fh.Protocal.Value[2] != 0x42 {
		err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > %w", ErrFbProtocalStruct)
		return
	}

	if dataLen := fh.Protocal.Value[3]; int(dataLen) != len(fh.Protocal.Value[4:]) {
		err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > %w", ErrFbProtocalLen)
		return
	}

	msg := entity.FeibeeLockAlarmMsg{
		Header: entity.Header{
			Cmd:     0,
			Ack:     1,
			DevType: "WonlyFBlock",
			DevId:   fh.data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   0,
		},
		Timestamp: fh.data.Records[0].Uptime,
	}

	switch fh.Protocal.Value[5] {
	case 0x04: //强拆报警
		msg.Cmd = 0x22
	case 0x05: //门未关报警
		msg.Cmd = 0x26
	case 0x06: //胁迫报警
	case 0x07: //假锁报警
		msg.Cmd = 0x24
	case 0x33: //非法操作报警
		msg.Cmd = 0x20
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > json.Marshal > %w", err)
		return
	}
	rabbitmq.Publish2pms(bs, "")
	rabbitmq.Publish2mns(bs, "")
	return nil
}

func (fh *FbLockHandle) FbLockOnoffDecode() (err error) {
	if len(fh.Protocal.Value) < 16 {
		err = fmt.Errorf("FbLockHandle.FbLockOnoffDecode > %w", ErrFbProtocalLen)
		return
	}

	if int(fh.Protocal.Value[3]) != len(fh.Protocal.Value[4:]) {
		err = fmt.Errorf("FbLockHandle.FbLockOnoffDecode > %w", ErrFbProtocalLen)
		return
	}

	unlockType := fh.Protocal.Value[5]
	if fh.Protocal.Value[6] != 2 {
		//非开锁消息
		return
	}

	userId := int(fh.Protocal.Value[7]) + int(fh.Protocal.Value[8])<<8

	switch unlockType {
	case 0x04:
		err = fh.remoteUnlock(userId)
	case 0x00, 0x01, 0x02, 03:
		err = fh.otherUnlock(userId)
	}

	return
}

func (fh *FbLockHandle) remoteUnlock(userId int) (err error) {
	stateFlag := fh.Protocal.Value[15]

	msg := entity.FeibeeLockRemoteOn{
		Header: entity.Header{
			Cmd:     0x52,
			Ack:     1,
			DevId:   fh.data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   1,
			DevType: "WonlyFBlock",
		},
		UserId:    userId,
		Timestamp: fh.data.Records[0].Uptime / 1000,
	}
	sendFlag := false
	if stateFlag&0b0001_0000 > 1 {
		redisKey := "FbRemoteUnlock_" + fh.data.Records[0].Uuid
		if prevUserId, err := redis.GetFbLockUserId(redisKey); err != nil {
			if err = redis.SetFbLockUserId(redisKey, userId); err != nil {
				log.Warningf("FbLockHandle.remoteUnlock > %s", err)
				sendFlag = true
			} else {
				return nil
			}
		} else {
			sendFlag = true
			msg.UserId2 = prevUserId
		}
	} else {
		sendFlag = true
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.remoteUnlock > json.Marshal > %w", err)
		return
	}

	if sendFlag {
		rabbitmq.Publish2pms(bs, "")
		rabbitmq.Publish2app(bs, msg.DevId)
		rabbitmq.Publish2mns(bs, "")
	}
	return nil
}

func (fh *FbLockHandle) otherUnlock(userId int) (err error) {
	msg := entity.UploadOpenLockLog{
		Cmd:     0x40,
		Ack:     1,
		DevType: "WonlyFBlock",
		DevId:   fh.data.Records[0].Uuid,
		Vendor:  "feibee",
		SeqId:   0,
		LogList: []entity.OpenLockLog{
			entity.OpenLockLog{
				UserId:    uint16(userId),
				MainOpen:  0,
				SubOpen:   0,
				SinMul:    1,
				Remainder: 0,
				Time:      int32(fh.data.Records[0].Uptime / 1000),
			},
		},
	}
	stateFlag := fh.Protocal.Value[15]
	if stateFlag&0b1000_0000 > 1 {
		msg.LogList[0].SubOpen = 1
	}

	if stateFlag&0b0001_0000 > 1 {
		msg.LogList[0].SinMul = 2
	}

	switch fh.Protocal.Value[5] {
	case 0x00, 0x01:
		msg.LogList[0].MainOpen = 1
	case 0x02:
		msg.LogList[0].MainOpen = 3
	case 0x03:
		msg.LogList[0].MainOpen = 2
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > json.Marshal > %w", err)
		return
	}
	rabbitmq.Publish2pms(bs, "")
	rabbitmq.Publish2mns(bs, "")
	rabbitmq.Publish2app(bs, fh.data.Records[0].Uuid)
	return nil
}

func (fh *FbLockHandle) FbLockStateDecode(data []byte) (err error) {
	return
}

func (fh *FbLockHandle) FbLockBattDecode() (err error) {
	if fh.Protocal.Value[2] != 0x1b || len(fh.Protocal.Value) != 9 || fh.Protocal.Value[7] != 1 {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > %w", ErrFbProtocalStruct)
		return
	}

	msg := entity.FeibeeLockBattMsg{
		Header: entity.Header{
			Cmd:     0x2a,
			Ack:     1,
			DevType: "WonlyFBlock",
			DevId:   fh.data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   0,
		},
		Timestamp: fh.data.Records[0].Uptime,
		Value:     int(fh.Protocal.Value[10]) / 2,
	}

	bs, err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > json.Marshal > %w", err)
		return
	}
	rabbitmq.Publish2pms(bs, "")
	rabbitmq.Publish2mns(bs, "")
	return nil
}

