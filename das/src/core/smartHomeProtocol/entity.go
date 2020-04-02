package smartHomeProtocol

func NewSmartHomeProtocol() SmartHomeProtocol {
	return SmartHomeProtocol{
		DevId: "12343r345345",
		BeginFlag: '{',
		Header: Header{
			FeatureCode0: 'W',
			FeatureCode1: 'L',
		},
		Data:    []SmartDevData{},
		EndFlag: '}',
	}
}

type SmartHomeProtocol struct {
	BeginFlag byte
	Header
	Data    []SmartDevData
	EndFlag byte

	DevId string
}

type Header struct {
	FeatureCode0 byte
	FeatureCode1 byte
	Len      uint8
	CheckSum uint8
	HighId   uint8
	LowId    uint8
	Count    uint8
	DevType  uint8
	MsgType  uint8
}

type SmartDevData struct {
	Type uint8
	DataVal
}

type DataVal struct {
	Val0 uint8
	Val1 uint8
	Val2 uint8
}
