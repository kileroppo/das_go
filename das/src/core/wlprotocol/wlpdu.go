package wlprotocol

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"das/core/util"
	"das/core/log"
)

//2. 请求同步用户列表(0x31)(服务器-->前板)
// 同步设备用户-包体打包
func (pdu *SyncDevUser) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Num); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] SyncDevUser Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *SyncDevUser) Decode(bBody []byte, uuid string) error {
	return nil
}

// 请求同步用户列表(0x31)(前板-->服务器)
// 同步设备用户返回-包体解包
func (pdu *SyncDevUserResp) Decode(bBody []byte, uuid string) error {
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

	log.Debug("[ ", uuid, " ] SyncDevUserResp Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserNum); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	var i uint16
	for  i =0; i < pdu.DevUserNum; i++ {
		var devUserInfo DevUserInfo
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.UserType); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.UserNo); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.OpenBitMap); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.PermitNum); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.Remainder); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.StartDate); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.EndDate); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.TimeSlot1); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.TimeSlot2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &devUserInfo.TimeSlot3); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}

		pdu.DevUserInfos = append(pdu.DevUserInfos, devUserInfo)
	}

	return err
}

//3. 删除用户(0x32)(服务器-->前板)
func (pdu *DelDevUser) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.UserNo); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.MainOpen); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.SubOpen); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] DelDevUser Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *DelDevUser) Decode(bBody []byte, uuid string) error { // 包体解包，AES解密
	return nil
}

//4. 新增用户(0x33)(服务器-->前板)
// 用户编号(2)+开锁方式(1)+是否胁迫(1)+用户类型(1)+密码(10)+时间(4)+用户随机数(4)+临时用户时效(20)
func (pdu *AddDevUser) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.UserNo); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.MainOpen); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.SubOpen); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.UserType); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Passwd); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.UserNote); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.PermitNum); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.StartDate); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.EndDate); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot1); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot2); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot3); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.BlePin); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] AddDevUser Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *AddDevUser) Decode(bBody []byte, uuid string) error { // 包体解包，AES解密
	return nil
}

//5. 新增用户报告步骤(0x34)(前板-->服务器)
// 用户列表版本号(4)+用户编号(2)+开锁方式(1)+是否胁迫(1)+步骤序号(1)+步骤状态(1)+时间(4)
func (pdu *AddDevUserStep) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] AddDevUserStep Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.MainOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SubOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.StepNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.StepState); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//6. 用户更新上报(0x35)(前板-->服务器)
// 用户列表版本号(4)+操作类型(1)+ 用户编号(2)+用户类型(1)+用户随机数(4)+ 验证方式位图(4)+总次数(2)+剩余次数（2）+开始日期(3)+截止日期(3)+时段1(4)+时段2(4)+时段3(4)+时间（4）
func (pdu *UserUpdateLoad) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] UserUpdateLoad Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.OperType); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserType); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.OpenBitMap); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.PermitNum); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Remainder); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.StartDate); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.EndDate); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.TimeSlot1); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.TimeSlot2); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.TimeSlot3); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//7. 实时视频(0x36)(服务器-->前板)
func (pdu *RealVideo) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Act); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] RealVideo Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}

func (pdu *RealVideo) Decode(bBody []byte, uuid string) error {
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

	log.Debug("[ ", uuid, " ] RealVideo Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.Act); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//8. Wifi设置(0x37)(服务器-->前板)
// Ssid（32）+密码（16）
func (pdu *WiFiSet) Encode(uuid string) ([]byte, error) {
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
	log.Debug("[ ", uuid, " ] WiFiSet Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}

func (pdu *WiFiSet) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] WiFiSet Decode [ ", hex.EncodeToString(DValue), " ]")

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

//9. 门铃呼叫(0x38)(前板-->服务器)
func (pdu *DoorbellCall) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] DoorbellCall Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//10. 人体感应报警(0x39)(前板-->服务器)
func (pdu *Alarms) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] Alarms Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//15. 低压报警(0x2A)(前板--->服务器)
func (pdu *LowBattAlarm) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] LowBattAlarm Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.Battery); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//16. 图片上传(0x2F)(前板--->服务器)
// 消息类型(1)+消息id(4)+图片路径长度（1）+图片路径(n)
func (pdu *PicUpload) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] PicUpload Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.CmdType); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.MsgId); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.PicLen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	var bPicPath []byte
	bPicPath = buf.Next(int(pdu.PicLen))
	pdu.PicPath = string(bPicPath[:pdu.PicLen])

	return err
}

//17. 用户开锁消息上报(0x40)(前板--->服务器)
/*
用户列表版本号(4)+ 用户数量（1）+时间(4)+ 电量百分比（1）+ 单/双人模式（1）
+用户编号(2)+验证方式(1)+是否胁迫(1)+剩余次数（2）
+用户编号(2)+验证方式(1)+是否胁迫(1)+剩余次数（2）
*/
func (pdu *OpenLockMsg) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] OpenLockMsg Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNum); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Battery); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SinMul); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.MainOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SubOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Remainder); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if 2 == pdu.SinMul { // 双人开门模式
		if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &pdu.MainOpen2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &pdu.SubOpen2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &pdu.Remainder2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
	}

	return nil
}

//18. 用户进入菜单上报(0x42)(前板--->服务器)
/*
用户列表版本号(4)+ 时间(4)+ 电量百分比（1）+ 单/双人模式（1）
+用户编号(2)+验证方式(1)+是否胁迫(1)
+用户编号(2)+验证方式(1)+是否胁迫(1)
*/
func (pdu *EnterMenuMsg) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] EnterMenuMsg Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.DevUserVer); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Battery); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SinMul); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.MainOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.SubOpen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if 2 == pdu.SinMul { // 双人开门模式
		if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &pdu.MainOpen2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		if err = binary.Read(buf, binary.BigEndian, &pdu.SubOpen2); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
	}

	return nil
}

//19. 在线离线(0x46)(后板-->服务器)
func (pdu *OnOffLine) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] OnOffLine Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.OnOff); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	/*TODO:JHHE
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}*/

	return err
}

//20. 远程开锁命令(0x52)(服务器->前板)
// 密码1（6）+密码2（6）+随机数（4）+md5（16）
func (pdu *RemoteOpenLock) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Passwd); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Passwd2); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] RemoteOpenLock Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *RemoteOpenLock) Decode(bBody []byte, uuid string) error {
	return nil
}

func (pdu *RemoteOpenLockResp) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] RemoteOpenLockResp Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.UserNo2); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Time); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return err
}

//22. 设置参数(0x72)(服务器-->前板，后板)
func (pdu *SetLockParamReq) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.ParamNo); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.ParamValue); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	if 0xFF != pdu.ParamValue2 { // 如果为0xFF则不填充
		if err = binary.Write(buf, binary.BigEndian, pdu.ParamValue2); err != nil {
			log.Error("binary.Write failed:", err)
			return nil, err
		}
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] SetLockParamReq Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *SetLockParamReq) Decode(bBody []byte, uuid string) error {
	return nil
}

//23. 参数更新(0x73)(前板,后板-->服务器)
func (pdu *ParamUpdate) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] ParamUpdate Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.ParamNo); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if 0x0d == pdu.ParamNo {		// 视频模组sn	0x0d	16字节(仅上报,不能修改)
		pdu.ParamValue = string(buf.Next(16))
	} else if 0x0f == pdu.ParamNo {	// WIFI_SSID	0x0f	32个字节
		pdu.ParamValue = string(buf.Next(32))
	} else {
		var paramValue uint8
		if err = binary.Read(buf, binary.BigEndian, &paramValue); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		pdu.ParamValue = paramValue

		if 0x0b == pdu.ParamNo {
			if err = binary.Read(buf, binary.BigEndian, &pdu.ParamValue2); err != nil {
				log.Error("特殊处理，binary.Read failed:", err)
				return err // 特殊处理，当不含参数值2时则直接返回空
			}
		}
	}

	return nil
}

//24. 软件重启命令(0x74)(服务器-->前、后板)
func (pdu *RebootLock) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] RebootLock Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}

//26. 设置临时用户时段(0x76)(服务器-->前板)
// 临时用户编号(2)+次数(2)+开始日期(3)+截止日期(3)+时段1(4)+时段2(4)+时段3(4)
func (pdu *SetTmpDevUser) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.UserNo); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.PermitNum); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.StartDate); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.EndDate); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot1); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot2); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, pdu.TimeSlot3); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] SetTmpDevUser Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *SetTmpDevUser) Decode(bBody []byte, uuid string) error {
	return nil
}

//27. 发送设备信息(0x70)(前板，后板-->服务器)
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
func (pdu *UploadDevInfo) Decode(bBody []byte, uuid string) error {
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
	if err = binary.Read(buf, binary.BigEndian, &pdu.FLen); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
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
	if err = binary.Read(buf, binary.BigEndian, &pdu.IpcSn); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Ssid); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	if err = binary.Read(buf, binary.BigEndian, &pdu.Capability); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	if err = binary.Read(buf, binary.BigEndian, &pdu.BLen); err != nil {
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
	if err = binary.Read(buf, binary.BigEndian, &pdu.ProductId); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return nil
}
func (pdu *UploadDevInfoResp) Encode(uuid string) ([]byte, error) {
	buf := new(bytes.Buffer) // 定义一个buffer，给了打包数据使用

	// 组body
	var err error
	if err = binary.Write(buf, binary.BigEndian, pdu.Time); err != nil {
		log.Error("binary.Write failed:", err)
		return nil, err
	}

	toDevice_byte := buf.Bytes()
	log.Debug("[ ", uuid, " ] UploadDevInfoResp Encode [ ", hex.EncodeToString(toDevice_byte), " ]")

	var toDevData []byte
	myKey := util.MD52Bytes(uuid)
	if toDevData, err = util.ECBEncryptByte(toDevice_byte, myKey); err != nil {
		log.Error("ECBEncryptByte failed, err=", err)
		return nil, err
	}

	return toDevData, nil
}
func (pdu *UploadDevInfoResp) Decode(bBody []byte, uuid string) error {
	return nil
}

// 锁状态上报(0x55)(后板->服务器)
func (pdu *DoorStateUpload) Decode(bBody []byte, uuid string) error {
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
	log.Debug("[ ", uuid, " ] DoorStateUpload Decode [ ", hex.EncodeToString(DValue), " ]")

	//3. 解包体 FLen
	buf := bytes.NewBuffer(DValue)
	if err = binary.Read(buf, binary.BigEndian, &pdu.State); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}

	return nil
}