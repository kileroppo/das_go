package smartHomeProtocol

import (
	"das/core/util"
)

func (s *SmartHomeProtocol) decrypt(cipherText []byte, key string) ([]byte, error) {
	return util.ECBDecryptByte(cipherText, util.MD52Bytes(key))
}

func (s *SmartHomeProtocol) encrypt(rawText []byte, key string) ([]byte, error) {
	return util.ECBEncryptByte(rawText, util.MD52Bytes(key))
}

func (s *SmartHomeProtocol) Encode() ([]byte,error) {
	s.getLen()

	sumData := make([]byte, 0, s.Len)
	header := s.encodeHeader()
	data := s.encodeData()

	sumData = append(sumData, s.BeginFlag)
	sumData = append(sumData, header...)
	sumData = append(sumData, data...)
	sumData = append(sumData, s.EndFlag)

	s.checkSum(sumData)
	return s.encrypt(sumData, s.DevId)
	//return sumData, nil
}

func (s *SmartHomeProtocol) encodeHeader() []byte {
	sumData := make([]byte, 9)
	sumData[0] = s.FeatureCode0
	sumData[1] = s.FeatureCode1
	sumData[2] = s.Len
	sumData[3] = s.CheckSum
	sumData[4] = s.HighId
	sumData[5] = s.LowId
	sumData[6] = s.Count
	sumData[7] = s.DevType
	sumData[8] = s.MsgType
	return sumData
}

func (s *SmartHomeProtocol) encodeData() []byte {
	sumData := make([]byte, 0, len(s.Data)*4)
	singleData := make([]byte, 4)
	for i := 0; i < len(s.Data); i++ {
		singleData[0] = s.Data[i].Type
		singleData[1] = s.Data[i].Val0
		singleData[2] = s.Data[i].Val1
		singleData[3] = s.Data[i].Val2
		sumData = append(sumData, singleData...)
	}
	return sumData
}

func (s *SmartHomeProtocol) getLen() {
	s.Len = uint8(len(s.Data)*4 + 11)
}

func (s *SmartHomeProtocol) checkSum(data []byte) uint8 {
	s.CheckSum = s.HighId
	for i := 6; i < len(data)-1; i++ {
		s.CheckSum = s.CheckSum ^ data[i]
	}
	data[4] = s.CheckSum
	return s.CheckSum
}
