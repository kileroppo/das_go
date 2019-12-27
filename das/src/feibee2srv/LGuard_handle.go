package feibee2srv

import (
	"strconv"
	"errors"

	"das/core/entity"
	)

var ErrParseRawData = errors.New("WonlyLGuard parse rawData error")

type LGuardMsgParse struct {
	msg entity.FeibeeData
	funcCode int32
	funcData string
}

func (l *LGuardMsgParse) parseRawData() error {
	rawData := l.msg.Records[0].Value
	lenStr := rawData[2:4]
	funcStr := rawData[4:6]

	lens,err := strconv.ParseInt(lenStr, 16, 64)
	if err != nil || int64(len(rawData[6:])/2) < lens+2 {
		return ErrParseRawData
	}
	funcCode, err := strconv.ParseInt(funcStr, 16, 64)
	if err != nil {
		return ErrParseRawData
	}

	l.funcCode = int32(funcCode)
	l.funcData = rawData[6:6+2*lens]
	return nil
}

func (l *LGuardMsgParse) handleByType() {

}

