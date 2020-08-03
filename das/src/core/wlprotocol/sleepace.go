package wlprotocol

import (
	"encoding/binary"
	"hash/crc32"

	"github.com/valyala/bytebufferpool"
)

var (
	Seperate = []byte{byte(0x24), byte(0x5f), byte(0x40), byte(0x2d)}
)

type SleepaceProtocalHandle struct {
	Protocal SleepaceProtocal
	BinaryData []byte
}

func (self *SleepaceProtocalHandle) Decode() {

}

func (self *SleepaceProtocalHandle) Encode() {
	oriData := self.encode()
	checkSum := self.checkSum(oriData)

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	buf.Write(oriData)
	buf.Write(checkSum)
	buf.Write(Seperate)

	self.send2dev(buf.Bytes())
}

func (self *SleepaceProtocalHandle) send2dev([]byte) {

}

func (self *SleepaceProtocalHandle) send2app() {

}

func (self *SleepaceProtocalHandle) checkSum(oriData []byte) []byte {
    sum := crc32.ChecksumIEEE(oriData)
    by := make([]byte, 4, 4)
    binary.BigEndian.PutUint32(by, sum)
    return by
}

func (self *SleepaceProtocalHandle) encode() ([]byte) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	buf.WriteByte(byte(self.Protocal.Ver))
	buf.WriteByte(byte(self.Protocal.Source))

	byte4Buf := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(byte4Buf, self.Protocal.ChannelNo)
	buf.Write(byte4Buf)

	buf.WriteByte(byte(self.Protocal.FrameType))

	byte2Buf := make([]byte, 2, 2)
	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.FrameNum)
	buf.Write(byte2Buf)

	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.FrameSerial)
	buf.Write(byte2Buf)

	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.FrameNo)
	buf.Write(byte2Buf)

	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.SpecialFlag)
	buf.Write(byte2Buf)

	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.DevType)
	buf.Write(byte2Buf)

	binary.BigEndian.PutUint16(byte2Buf, self.Protocal.MsgType)
	buf.Write(byte2Buf)

	buf.Write(self.Protocal.Data)

	return buf.Bytes()
}

type AromaLamps struct {
	SleepaceProtocalHandle
}

func check() {
	return
}
