package feibee2srv

import (
	"encoding/hex"
	"fmt"
)

type FbLockMsgTyp = uint16

const (
	Fb_Lock_Onoff      FbLockMsgTyp = 0x0101 //开锁
	Fb_Lock_Batt       FbLockMsgTyp = 0x0001 //电量
	Fb_Lock_State      FbLockMsgTyp = 0x0000 //锁状态
	Fb_Lock_Alarm      FbLockMsgTyp = 0x0009 //锁报警
	Fb_PM_Formaldehyde FbLockMsgTyp = 0x0417 //PM 甲醛
	Fb_PM_CO2          FbLockMsgTyp = 0x0416 //PM CO2
	Fb_PM_VOC          FbLockMsgTyp = 0x0418 //PM VOC
	Fb_PM_Batt         FbLockMsgTyp = 0x0402 //PM 电量
	Fb_PM_Temperature  FbLockMsgTyp = 0x0402 //PM 温度
	Fb_PM_Humidity     FbLockMsgTyp = 0x0405 //PM 湿度
)

type FbDevHeader struct {
	Head      uint8  //帧头
	Len       uint8  //后续数据包长度
	Addr      uint16 //设备短地址
	Cluster   uint16 //协议簇
	DataNum   uint8  //数据包数量
	Attribute uint16 //属性
}

type FbDevProtocal struct {
	FbDevHeader
	Value []byte
}

func (protocal *FbDevProtocal) Decode(rawData string) (err error) {
	hexData, err := hex.DecodeString(rawData)
	if err != nil {
		err = fmt.Errorf("FbDevProtocal > hex.DecodeString > %w", err)
		return
	}

	if len(hexData) <= 1 || hexData[0] != 0x70 {
		err = ErrFbProtocalStruct
		return
	}

	protocal.Head = hexData[0]
	protocal.Len = hexData[1]

	if int(protocal.Len) != len(hexData[2:]) {
		err = ErrFbProtocalLen
		return
	}
	protocal.Addr = uint16(hexData[2]) | (uint16(hexData[3]) << 8)
	protocal.Cluster = uint16(hexData[5]) | (uint16(hexData[6]) << 8)
	protocal.DataNum = hexData[7]
	protocal.Value = hexData[8:]

	return
}
