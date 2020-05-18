package feibee2srv

import (
	"das/core/redis"
	"encoding/hex"
	"errors"
	"fmt"

	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	ErrFbLockStruct = errors.New("the FbLock's structure was invalid")
	ErrFbLockLen    = errors.New("the FbLock's lens was wrong")
)

type FbLockMsgTyp = uint16

const (
	FbLockOnoff FbLockMsgTyp = 0x0101
	FbLockBatt  FbLockMsgTyp = 0x0100
	FbLockState FbLockMsgTyp = 0x0000
	FbLockAlarm FbLockMsgTyp = 0x0900
)

type FbLockHeader struct {
	Typ      uint8
	Len      uint8
	DataNum  uint8
	DataTyp  uint16
	DataFlag uint16
}

type FbLockProtocal struct {
	FbLockHeader
	Data []byte
}

type FbLockHandle struct {
	data     *entity.FeibeeData
	Protocal FbLockProtocal
}

func (fh *FbLockHandle) PushMsg() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("FbLockHandle.PushMsg > %s", err)
		}
	}()

    if err := fh.decode(); err != nil {
    	log.Warningf("FbLockHandle.PushMsg > %s", err)
	}
}
//todo: feibeelock handle
func (fh *FbLockHandle) decode() (err error) {
	fh.Protocal, err = FbLockDecode(fh.data.Records[0].Orgdata)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.decode > %w", err)
		return
	}

	switch fh.Protocal.DataTyp {
	case FbLockAlarm:
		err = fh.FbLockAlarmDecode()
	case FbLockBatt:
		err = fh.FbLockBattDecode()
	case FbLockOnoff:
		err = fh.FbLockOnoffDecode()
	}

	if err != nil {
		err = fmt.Errorf("FbLockHandle.decode > %w", err)
	}

	return
}

func (fh *FbLockHandle) FbLockAlarmDecode() (err error){
    if fh.Protocal.Data[0] != 0x42 {
    	err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > %w", ErrFbLockStruct)
    	return
	}

    if dataLen := fh.Protocal.Data[1]; int(dataLen) != len(fh.Protocal.Data[2:]) {
		err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > %w", ErrFbLockLen)
		return
	}

    msg := entity.FeibeeLockAlarmMsg{
		Header:    entity.Header{
			Cmd:     0,
			Ack:     1,
			DevType: "WonlyFBlock",
			DevId:   fh.data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   0,
		},
		Timestamp: fh.data.Records[0].Uptime,
	}

    switch fh.Protocal.Data[3] {
	case 0x04: //强拆报警
		msg.Cmd = 0x22
	case 0x05://门未关报警
		msg.Cmd = 0x26
	case 0x06://胁迫报警
	case 0x07://假锁报警
		msg.Cmd = 0x24
	case 0x33://非法操作报警
	    msg.Cmd = 0x20
	}

    bs,err := json.Marshal(msg)
    if err != nil {
    	err = fmt.Errorf("FbLockHandle.FbLockAlarmDecode > json.Marshal > %w", err)
    	return
	}
    rabbitmq.Publish2pms(bs, "")
    rabbitmq.Publish2mns(bs, "")
    return nil
}

func (fh *FbLockHandle) FbLockOnoffDecode() (err error){
	if len(fh.Protocal.Data) < 14 {
		err = fmt.Errorf("FbLockHandle.FbLockOnoffDecode > %w", ErrFbLockLen)
		return
	}

	if int(fh.Protocal.Data[1]) != len(fh.Protocal.Data[2:]) {
		err = fmt.Errorf("FbLockHandle.FbLockOnoffDecode > %w", ErrFbLockLen)
		return
	}

	unlockType := fh.Protocal.Data[3]
	if fh.Protocal.Data[4] != 2 {
		//非开锁消息
		return
	}

	userId := int(fh.Protocal.Data[5]) + int(fh.Protocal.Data[6]) << 8

	switch unlockType {
	case 0x04:
        err = fh.remoteUnlock(userId)
	case 0x00,0x01,0x02,03:
		err = fh.otherUnlock(userId)
	}

	return
}

func (fh *FbLockHandle) remoteUnlock(userId int) (err error){
	stateFlag := fh.Protocal.Data[13]
	
	msg := entity.FeibeeLockRemoteOn{
		Header:    entity.Header{
			Cmd:0x52,
			Ack:1,
			DevId:fh.data.Records[0].Uuid,
			Vendor:"feibee",
			SeqId:1,
			DevType:"WonlyFBlock",
		},
		UserId:    userId,
		Timestamp: fh.data.Records[0].Uptime/1000,
	}
    sendFlag := false
	if stateFlag & 0b0001_0000 > 1 {
		redisKey := "FbRemoteUnlock_" + fh.data.Records[0].Uuid
		if prevUserId,err := redis.GetFbLockUserId(redisKey); err != nil {
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

    bs,err := json.Marshal(msg)
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

func (fh *FbLockHandle) otherUnlock(userId int) (err error){
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
				Time:      int32(fh.data.Records[0].Uptime/1000),
			},
		},
	}
	stateFlag := fh.Protocal.Data[13]
	if stateFlag & 0b1000_0000 > 1 {
		msg.LogList[0].SubOpen = 1
	}

	if stateFlag & 0b0001_0000 > 1 {
		msg.LogList[0].SinMul = 2
	}

	switch fh.Protocal.Data[3] {
	case 0x00,0x01:
		msg.LogList[0].MainOpen = 1
	case 0x02:
		msg.LogList[0].MainOpen = 3
	case 0x03:
		msg.LogList[0].MainOpen = 2
	}

	bs,err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > json.Marshal > %w", err)
		return
	}
	rabbitmq.Publish2pms(bs, "")
	rabbitmq.Publish2mns(bs, "")
	rabbitmq.Publish2app(bs, fh.data.Records[0].Uuid)
	return nil
}

func (fh *FbLockHandle) FbLockStateDecode(data []byte) (err error){
    return
}

func (fh *FbLockHandle) FbLockBattDecode() (err error){
	if fh.Protocal.Data[0] != 0x1b || len(fh.Protocal.Data) != 9 || fh.Protocal.Data[5] != 1 {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > %w", ErrFbLockStruct)
		return
	}

	msg := entity.FeibeeLockBattMsg{
		Header:    entity.Header{
			Cmd:     0x2a,
			Ack:     1,
			DevType: "WonlyFBlock",
			DevId:   fh.data.Records[0].Uuid,
			Vendor:  "feibee",
			SeqId:   0,
		},
		Timestamp: fh.data.Records[0].Uptime,
		Value: int(fh.Protocal.Data[8]) / 2,
	}

	bs,err := json.Marshal(msg)
	if err != nil {
		err = fmt.Errorf("FbLockHandle.FbLockBattDecode > json.Marshal > %w", err)
		return
	}
	rabbitmq.Publish2pms(bs, "")
	rabbitmq.Publish2mns(bs, "")
	return nil
}

func FbLockDecode(rawData string) (protocal FbLockProtocal, err error) {
	hexData, err := hex.DecodeString(rawData)
	if err != nil {
		err = fmt.Errorf("FbLockDecode > hex.DecodeString > %w", err)
		return
	}

	if len(hexData) <= 1 || hexData[0] != 0x70 {
		err = ErrFbLockStruct
		return
	}

	protocal.Typ = hexData[0]
	protocal.Len = hexData[1]

	if int(protocal.Len) != len(hexData[2:]) {
		err = ErrFbLockLen
		return
	}

	protocal.DataTyp = uint16(hexData[5])<<8 + uint16(hexData[6])
	protocal.DataNum = hexData[7]
	protocal.DataFlag = uint16(hexData[8])<<8 + uint16(hexData[9])
	protocal.Data = hexData[10:]

	return
}