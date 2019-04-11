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
