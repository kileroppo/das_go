package smartHomeProtocol

import (
	"errors"
	"fmt"
)

var (
	ErrDataValid = errors.New("data was invalid")
)

func (s *SmartHomeProtocol) Decode(rawData []byte) error {
	data,err := s.decrypt(rawData, s.DevId)
	if err != nil {
		return fmt.Errorf("SmartHomeProtocol.Decode > decrypt > %w", err)
	}
	if !s.checkValid(data) {
		return fmt.Errorf("SmartHomeProtocol.Decode > checkValid > %w", ErrDataValid)
	}

	s.decodeHeader(data)
	s.decodeData(data)

	return nil
}

func (s *SmartHomeProtocol) checkValid(data []byte) bool {
	if len(data) < 11 {
		return false
	}

	if (len(data) - 11) % 4 != 0 {
		return false
	}

	return s.checkSum(data) == uint8(data[4]) && int(data[3]) == len(data)
}

func (s *SmartHomeProtocol) decodeHeader(rawData []byte) {
	s.Len = rawData[3]
	s.CheckSum = rawData[4]
	s.HighId = rawData[5]
	s.LowId = rawData[6]
	s.Count = rawData[7]
	s.DevType = rawData[8]
	s.MsgType = rawData[9]
}

func (s *SmartHomeProtocol) decodeData(rawData []byte) {
    data := rawData[10:len(rawData) - 1]
    len := len(data)/4
    s.Data = make([]SmartDevData, len)

    for i:=0;i<len;i++ {
        s.Data[i].Type = data[4*i]
        s.Data[i].Val0 = data[4*i + 1]
		s.Data[i].Val1 = data[4*i + 2]
		s.Data[i].Val2 = data[4*i + 3]
	}
}
