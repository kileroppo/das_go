package entity

import (
	"bytes"
	"encoding/binary"
	"../log"
	"encoding/hex"
	"strconv"
)

func (devUser *DevUser) ParseUser(DValue string) error {
	var userType uint8      // 用户类型（0-管理员，1-普通用户，2-临时用户）
	var userId uint16       // 设备用户ID
	var finger uint16		// 指纹数量
	var ffinger uint16		// 胁迫指纹数量
	var passwd uint8		// 密码数量
	var card uint8			// 卡数量
	var face uint8          // 人脸数量
	var bluetooth uint8     // 蓝牙数量
	var count uint16			// 开门次数，0为无限次
	var remainder uint16    	// 剩余开门次数
	var date_start [3]byte   	// 开始有效时间 : 年月日,BCD码，截止有效时间 : 年月日,BCD码
	var date_end [3]byte
	var time1_start [2]byte  	// 时段1，开始时间：时分BCD码，例10点5分—0x10 05，结束时间：时分BCD码，例12点5分—0x12 05
	var time1_end [2]byte
	var time2_start [2]byte  	// 时段2，开始时间：时分BCD码，例10点5分—0x10 05，结束时间：时分BCD码，例12点5分—0x12 05
	var time2_end [2]byte
	var time3_start [2]byte  	// 时段3，开始时间：时分BCD码，例10点5分—0x10 05，结束时间：时分BCD码，例12点5分—0x12 05
	var time3_end [2]byte

	byteData, err_0 := hex.DecodeString(DValue)
	if err_0 != nil {
		log.Error("parseUserList() hex.DecodeString failed, err=", err_0, ", Data=", DValue)
		return err_0
	}
	//3. 解包体
	var err error
	buf := bytes.NewBuffer(byteData)
	if err = binary.Read(buf, binary.BigEndian, &userType); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.UserType = int(userType)

	if err = binary.Read(buf, binary.BigEndian, &userId); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.UserId = uint16(userId)

	if err = binary.Read(buf, binary.BigEndian, &finger); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Finger = int(finger)

	if err = binary.Read(buf, binary.BigEndian, &ffinger); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Ffinger = int(ffinger)

	if err = binary.Read(buf, binary.BigEndian, &passwd); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Passwd = int(passwd)

	if err = binary.Read(buf, binary.BigEndian, &card); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Card = int(card)

	if 32 <= len(byteData) {	// 大于等于的时候包含人脸开锁
		if err = binary.Read(buf, binary.BigEndian, &face); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		devUser.Face = int(face)
	}

	if 33 <= len(byteData) { // 大于等于的时候包含蓝牙开锁
		if err = binary.Read(buf, binary.BigEndian, &bluetooth); err != nil {
			log.Error("binary.Read failed:", err)
			return err
		}
		devUser.Bluetooth = int(bluetooth)
	}
	if err = binary.Read(buf, binary.BigEndian, &count); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Count = int(count)

	if err = binary.Read(buf, binary.BigEndian, &remainder); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	devUser.Remainder = int(remainder)

	if err = binary.Read(buf, binary.BigEndian, &date_start); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	mSDate := (int32(date_start[0]) * 10000) + (int32(date_start[1]) * 100) + int32(date_start[2])
	strSDate := strconv.FormatInt(int64(mSDate), 10) 			// 转10进制字符串
	nSDate, err2 := strconv.ParseInt(strSDate, 16, 32) // 转16进制值
	if nil != err2 {
		log.Error("ParseUser strconv.ParseInt, err2: ", err2)
	}
	devUser.MyDate.Start = int32(nSDate)

	if err = binary.Read(buf, binary.BigEndian, &date_end); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 结束日期
	// 转10进制
	mEDate := (int32(date_end[0]) * 10000) + (int32(date_end[1]) * 100) + int32(date_end[2])
	strEDate := strconv.FormatInt(int64(mEDate), 10) // 转10进制字符串
	nEDate, err3 := strconv.ParseInt(strEDate, 16, 32) // 转16进制值
	if nil != err3 {
		log.Error("ParseUser strconv.ParseInt, err3: ", err3)
	}
	devUser.MyDate.End = int32(nEDate)

	if err = binary.Read(buf, binary.BigEndian, &time1_start); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段1 - 开始
	mTimeSlot1_s := (int32(time1_start[0]) * 100) + int32(time1_start[1])
	strTimeSlot1_s := strconv.FormatInt(int64(mTimeSlot1_s), 10) // 转10进制字符串
	nTimeSlot1_s, err4 := strconv.ParseInt(strTimeSlot1_s, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[0].Start = int32(nTimeSlot1_s)

	if err = binary.Read(buf, binary.BigEndian, &time1_end); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段1 - 结束
	mTimeSlot1_e := (int32(time1_end[0]) * 100) + int32(time1_end[1])
	strTimeSlot1_e := strconv.FormatInt(int64(mTimeSlot1_e), 10) // 转10进制字符串
	nTimeSlot1_e, err4 := strconv.ParseInt(strTimeSlot1_e, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[0].End = int32(nTimeSlot1_e)

	if err = binary.Read(buf, binary.BigEndian, &time2_start); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段2 - 开始
	mTimeSlot2_s := (int32(time2_start[0]) * 100) + int32(time2_start[1])
	strTimeSlot2_s := strconv.FormatInt(int64(mTimeSlot2_s), 10) // 转10进制字符串
	nTimeSlot2_s, err4 := strconv.ParseInt(strTimeSlot2_s, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[1].Start = int32(nTimeSlot2_s)

	if err = binary.Read(buf, binary.BigEndian, &time2_end); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段2 - 结束
	mTimeSlot2_e := (int32(time2_end[0]) * 100) + int32(time2_end[1])
	strTimeSlot2_e := strconv.FormatInt(int64(mTimeSlot2_e), 10) // 转10进制字符串
	nTimeSlot2_e, err4 := strconv.ParseInt(strTimeSlot2_e, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[1].End = int32(nTimeSlot2_e)

	if err = binary.Read(buf, binary.BigEndian, &time3_start); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段3 - 开始
	mTimeSlot3_s := (int32(time3_start[0]) * 100) + int32(time3_start[1])
	strTimeSlot3_s := strconv.FormatInt(int64(mTimeSlot3_s), 10) // 转10进制字符串
	nTimeSlot3_s, err4 := strconv.ParseInt(strTimeSlot3_s, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[2].Start = int32(nTimeSlot3_s)

	if err = binary.Read(buf, binary.BigEndian, &time3_end); err != nil {
		log.Error("binary.Read failed:", err)
		return err
	}
	// 时段3 - 结束
	mTimeSlot3_e := (int32(time3_end[0]) * 100) + int32(time3_end[1])
	strTimeSlot3_e := strconv.FormatInt(int64(mTimeSlot3_e), 10) // 转10进制字符串
	nTimeSlot3_e, err4 := strconv.ParseInt(strTimeSlot3_e, 16, 32) // 转16进制值
	if nil != err4 {
		log.Error("ParseUser strconv.ParseInt, err4: ", err4)
	}
	devUser.MyTime[2].End = int32(nTimeSlot3_e)

	return nil
}