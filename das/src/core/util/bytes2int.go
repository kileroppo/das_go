package util

import (
	"bytes"
	"encoding/binary"
)

func BytesToInt16(b []byte) uint16 {
	buf := bytes.NewBuffer(b)
	var tmp uint16
	binary.Read(buf, binary.BigEndian, &tmp)
	return tmp
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}