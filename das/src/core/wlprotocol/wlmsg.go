package wlprotocol

import (
	"bytes"
	"encoding/binary"
	"../log"
	"../util"
)

/*
*	定义两个接口
*	不同的包体解包
*/
type IPdu interface {
	Encode(uuid string) ([]byte, error)			// 打包包体，返回的是AES加密后的数据
	Decode(bBody []byte, uuid string) error		// 包体解包，AES解密
}

// 打包
func (msg *WlMessage) PkEncode(pdu IPdu) ([]byte, error)  {
	var err error
	var bBody []byte
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	//1. 先组包体
	if nil != pdu {
		bBody, err = pdu.Encode(msg.DevId.Uuid)
		if err != nil {
			log.Error("pdu.Encode err:", err)
		}
		msg.Length = uint16(len(bBody))  // 包体长度
		msg.Check = util.CheckSum(bBody) // 包体校验和
	}

	//2. 打包
	if err = binary.Write(buf, binary.BigEndian, msg.Started); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Version); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Length); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Check); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.SeqId); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Cmd); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Ack); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.Type); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, msg.DevId.Len); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, []byte(msg.DevId.Uuid)); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	if nil != pdu { // 判断是否包含包体
		if err = binary.Write(buf, binary.BigEndian, bBody); err != nil {
			log.Error("binary.Write failed:", err)
			return nil, err
		}
	}

	if err = binary.Write(buf, binary.BigEndian, msg.Ended); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

// 解包
func (msg *WlMessage) PkDecode(pkg []byte) ([]byte, error) {
	var err error
	buf := bytes.NewBuffer(pkg)

	//1. 先解包头
	if err = binary.Read(buf, binary.BigEndian, &msg.Started); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Version); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Length); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Check); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.SeqId); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Cmd); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Ack); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.Type); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	if err = binary.Read(buf, binary.BigEndian, &msg.DevId.Len); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}
	var bDevId []byte
	bDevId = buf.Next(int(msg.DevId.Len))
	msg.DevId.Uuid = string(bDevId[:msg.DevId.Len])

	//2. 获取包体
	var bBody []byte
	bBody = buf.Next(int(msg.Length))

	if err = binary.Read(buf, binary.BigEndian, &msg.Ended); err != nil {
		log.Error("binary.Read failed:", err)
		return nil, err
	}

	return bBody, nil
}
