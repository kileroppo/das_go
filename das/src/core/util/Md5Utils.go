package util

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func MD52Bytes(str string) []byte {
	h := md5.New()
	h.Write([]byte(str))
	return h.Sum(nil)
}
