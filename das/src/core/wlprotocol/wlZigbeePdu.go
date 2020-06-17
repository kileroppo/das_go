package wlprotocol

import (
	"bytes"
	"das/core/log"
	"das/core/util"
	"encoding/binary"
	"encoding/hex"
)

//8. Wifi设置(0x37)(服务器-->前板)
// Ssid（32）+密码（16）
func (pdu *WiFiSet_Zigbee) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Ssid); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Passwd); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] WiFiSet_Zigbee Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}

func (pdu *WiFiSet_Zigbee) Decode(bBody []byte, uuid string) error {
	var err error
	var DValue []byte

	//1. 生成密钥
	myKey := util.MD52Bytes(uuid)

	//2. 解密
	DValue, err = util.ECBDecryptByte(bBody, myKey)
	if nil != err {
		log.Error("ECBDecryptByte failed, err=", err)
		return err
	}
	log.Debug("[ ", uuid, " ] WiFiSet_Zigbee Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.Ssid); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Passwd); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//27. zigbee锁 发送设备信息(0x70)(前板，后板-->服务器)
/*
前板信息：
版本号(4)：例1.0.78，主版本号(1)：1；次版本号(1)：0；修订号(2)：78；
型号(2)：门锁设备型号；
用户列表版本号(4): 初始为0，每次更改用户信息后加1
音量(1)： 0静音，1小，2中，3大。
验证模式(1)：1单人，2双人。
是否带屏(1)：0无屏，1带屏。
密码开关(1)：0表示密码禁用，1表示密码使能
电量(1):电池电量1~100
门未关报警开关(1)：0关闭，1开启
假锁报警开关(1)：0关闭，1开启
人体感应报警开关(1): 0关闭，1开启
人体感应报警时间(1): 1字节（单位秒）
报警类型(1)：1拍照+录像，2拍照
门铃开关(1)：0关闭，1开启
激活模式(1)：0门锁唤醒后立即激活，1输入激活码激活
视频模组sn码(16)：视频模组序列号
Ssid(32):模组连接的路由器的ssid
后板信息：
版本号(4)：例1.0.78，主版本号(1)：1；次版本号(1)：0；修订号(2)：78；
常开模式：0常开关闭，1常开启用
远程开关：0关闭，1开启
产品序列号：12字节字符串，例：Z12345670001
*/
func (pdu *UploadZigbeeDevInfo) Decode(bBody []byte, uuid string) error {
	var err error
	var DValue []byte

	//1. 生成密钥
	myKey := util.MD52Bytes(uuid)

	//2. 解密
	DValue, err = util.ECBDecryptByte(bBody, myKey)
	if nil != err {
		log.Error("ECBDecryptByte failed, err=", err)
		return err
	}
	log.Debug("[ ", uuid, " ] UploadDevInfo Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体 FLen
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.FMainVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.FSubVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.FModVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.FType); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Volume); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SinMul); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.IsHasScr); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.PwdSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Battery); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.NolockSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.FakelockSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.InfraSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.InfraTime); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.AlarmSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.BellSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.ActiveMode); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.FBreakSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Capability); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if err = binary.Read(buf, binary.BigEndian, &pdu.BMainVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.BSubVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.BModVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.OpenMode); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.RemoteSwitch); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return nil
}

