package smartHomeProtocol

import (
	"testing"
)

var (
	p = SmartHomeProtocol{
		DevId:     "12343r345345",
		BeginFlag: '{',
		Header: Header{
			FeatureCode0: 'W',
			FeatureCode1: 'L',
		},
		Data: []SmartDevData{
			{
				Type: 1,
				DataVal: DataVal{
					0xff,
					0xff,
					0xff,
				},
			},
			{
				Type: 2,
				DataVal: DataVal{
					0x32,
					0x32,
					0x32,
				},
			},
		},
		EndFlag: '}',
	}
)

func TestSmartHomeProtocol_EncodeDirect(t *testing.T) {
	cipher, _ := p.Encode()

	n := NewSmartHomeProtocol()
	n.Decode(cipher)
}

func BenchmarkNewSmartHomeProtocol_EncodeDirect(b *testing.B) {
	//res := []byte{}
	for i := 0; i < b.N; i++ {
		p.Encode()
	}
	//fmt.Printf("%s", res)
}
