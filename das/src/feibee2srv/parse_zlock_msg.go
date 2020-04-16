package feibee2srv

import (
	"fmt"
	"strconv"
	"time"
	"bytes"
	"encoding/hex"

	"das/procLock"
	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/core/wlprotocol"
)


/*
*	解包
*	1、先解包头；
*	2、根据包头来确定包体
*	3、组JSON包后转发APP，PMS模块
 */
func ParseZlockData(hexData, devType, uuid string) error {
	data, err := hex.DecodeString(hexData)
	if nil != err {
		log.Error("ParseZlockData hex.DecodeString, err=", err)
		return err
	}

	var wlMsg wlprotocol.WlZigbeeMsg
	bBody, err0 := wlMsg.PkDecode(data)
	if err0 != nil {
		log.Error("ParseZlockData wlMsg.PkDecode, err0=", err0)
		return err0
	}
	switch wlMsg.Cmd {
	case constant.Add_dev_user:			// 新增用户(0x33)(服务器-->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Add_dev_user")
		if wlMsg.Ack > 1 {
			addDevUser := entity.Header{
				Cmd: int(wlMsg.Cmd),
				Ack: int(wlMsg.Ack),
				DevType: devType,
				DevId: uuid,
				Vendor: "feibee",
				SeqId: int(wlMsg.SeqId),
			}
			if to_byte, err1 := json.Marshal(addDevUser); err1 == nil {
				rabbitmq.Publish2app(to_byte, uuid)
			} else {
				log.Error("[", uuid, "] constant.Add_dev_user to_byte json.Marshal, err=", err1)
				return err1
			}
		}
	case constant.Set_dev_user_temp:	// 设置临时用户时段(0x76)(服务器-->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Set_dev_user_temp")

		// 组包
		factoryReset := entity.Header{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(factoryReset); err1 == nil {
			//producer.SendMQMsg2APP(wlMsg.DevId.Uuid, string(to_byte))
			rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Set_dev_user_temp, err=", err1)
			return err1
		}
	case constant.Add_dev_user_step:	// 新增用户报告步骤(0x34)(前板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Add_dev_user_step")
		pdu := &wlprotocol.AddDevUserStep{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Add_dev_user_step pdu.Decode, err=", err)
			return err
		}

		addDevUserStep := entity.AddDevUserStep {
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			UserVer: pdu.DevUserVer,
			UserId: pdu.UserNo,
			MainOpen: int(pdu.MainOpen),
			SubOpen: int(pdu.SubOpen),
			Step: int(pdu.StepNo),
			StepState: int(pdu.StepState),
			Time: pdu.Time,
		}
		if to_byte, err1 := json.Marshal(addDevUserStep); err1 == nil {
			//producer.SendMQMsg2APP(wlMsg.DevId.Uuid, string(to_byte))
			rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Add_dev_user_step to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Del_dev_user:			// 删除用户(0x32)(服务器-->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Del_dev_user")
		if wlMsg.Ack > 1 {
			delDevUser := entity.Header{
				Cmd: int(wlMsg.Cmd),
				Ack: int(wlMsg.Ack),
				DevType: devType,
				DevId: uuid,
				Vendor: "feibee",
				SeqId: int(wlMsg.SeqId),
			}
			if to_byte, err1 := json.Marshal(delDevUser); err1 == nil {
				rabbitmq.Publish2app(to_byte, uuid)
			} else {
				log.Error("[", uuid, "] constant.Del_dev_user to_byte json.Marshal, err=", err1)
				return err1
			}
		}
	case constant.Update_dev_user:		// 用户更新上报(0x35)(前板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Update_dev_user")
		pdu := &wlprotocol.UserUpdateLoad{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Update_dev_user pdu.Decode, err=", err)
			return err
		}

		devUserUpload := entity.DevUserUpload{}
		devUserUpload.Cmd = int(wlMsg.Cmd)
		devUserUpload.Ack = int(wlMsg.Ack)
		devUserUpload.DevType = devType
		devUserUpload.DevId = uuid
		devUserUpload.Vendor = "feibee"
		devUserUpload.SeqId = int(wlMsg.SeqId)

		devUserUpload.OpType = int(pdu.OperType)
		devUserUpload.UserVer = pdu.DevUserVer
		devUserUpload.UserId = pdu.UserNo
		devUserUpload.UserNote = strconv.FormatInt(int64(pdu.Time), 16)
		devUserUpload.UserType = int(pdu.UserType)
		devUserUpload.Finger = int(pdu.OpenBitMap & 0x01)
		devUserUpload.Passwd = int(pdu.OpenBitMap & 0x01)
		devUserUpload.Card = int(pdu.OpenBitMap >> 1 & 0x01)
		devUserUpload.Finger = int((pdu.OpenBitMap >> 2 & 0x01) + (pdu.OpenBitMap >> 3 & 0x01) + (pdu.OpenBitMap >> 4 & 0x01))
		devUserUpload.Ffinger = int((pdu.OpenBitMap >> 5 & 0x01) + (pdu.OpenBitMap >> 6 & 0x01))
		devUserUpload.Face = int(pdu.OpenBitMap >> 7 & 0x01)
		devUserUpload.Bluetooth = 0
		devUserUpload.Count	= int(pdu.PermitNum)
		devUserUpload.Remainder	= int(pdu.Remainder)

		// 开始日期
		// 转10进制
		mSDate := (int32(pdu.StartDate[0]) * 10000) + (int32(pdu.StartDate[1]) * 100) + int32(pdu.StartDate[2])
		strSDate := strconv.FormatInt(int64(mSDate), 10) // 转10进制字符串
		nSDate, err2 := strconv.ParseInt(strSDate, 16, 32) // 转16进制值
		if nil != err2 {
			log.Error("ParseZlockData strconv.ParseInt, err2: ", err2)
		}
		devUserUpload.MyDate.Start = int32(nSDate)

		// 结束日期
		// 转10进制
		mEDate := (int32(pdu.EndDate[0]) * 10000) + (int32(pdu.EndDate[1]) * 100) + int32(pdu.EndDate[2])
		strEDate := strconv.FormatInt(int64(mEDate), 10) // 转10进制字符串
		nEDate, err3 := strconv.ParseInt(strEDate, 16, 32) // 转16进制值
		if nil != err3 {
			log.Error("ParseZlockData strconv.ParseInt, err3: ", err3)
		}
		devUserUpload.MyDate.End = int32(nEDate)

		// 时段1 - 开始
		mTimeSlot1_s := (int32(pdu.TimeSlot1[0]) * 100) + int32(pdu.TimeSlot1[1])
		strTimeSlot1_s := strconv.FormatInt(int64(mTimeSlot1_s), 10) // 转10进制字符串
		nTimeSlot1_s, err4 := strconv.ParseInt(strTimeSlot1_s, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[0].Start = int32(nTimeSlot1_s)

		// 时段1 - 结束
		mTimeSlot1_e := (int32(pdu.TimeSlot1[2]) * 100) + int32(pdu.TimeSlot1[3])
		strTimeSlot1_e := strconv.FormatInt(int64(mTimeSlot1_e), 10) // 转10进制字符串
		nTimeSlot1_e, err4 := strconv.ParseInt(strTimeSlot1_e, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[0].End = int32(nTimeSlot1_e)

		// 时段2 - 开始
		mTimeSlot2_s := (int32(pdu.TimeSlot2[0]) * 100) + int32(pdu.TimeSlot2[1])
		strTimeSlot2_s := strconv.FormatInt(int64(mTimeSlot2_s), 10) // 转10进制字符串
		nTimeSlot2_s, err4 := strconv.ParseInt(strTimeSlot2_s, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[1].Start = int32(nTimeSlot2_s)

		// 时段2 - 结束
		mTimeSlot2_e := (int32(pdu.TimeSlot2[2]) * 100) + int32(pdu.TimeSlot2[3])
		strTimeSlot2_e := strconv.FormatInt(int64(mTimeSlot2_e), 10) // 转10进制字符串
		nTimeSlot2_e, err4 := strconv.ParseInt(strTimeSlot2_e, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[1].End = int32(nTimeSlot2_e)

		// 时段3 - 开始
		mTimeSlot3_s := (int32(pdu.TimeSlot3[0]) * 100) + int32(pdu.TimeSlot3[1])
		strTimeSlot3_s := strconv.FormatInt(int64(mTimeSlot3_s), 10) // 转10进制字符串
		nTimeSlot3_s, err4 := strconv.ParseInt(strTimeSlot3_s, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[2].Start = int32(nTimeSlot3_s)

		// 时段3 - 结束
		mTimeSlot3_e := (int32(pdu.TimeSlot3[2]) * 100) + int32(pdu.TimeSlot3[3])
		strTimeSlot3_e := strconv.FormatInt(int64(mTimeSlot3_e), 10) // 转10进制字符串
		nTimeSlot3_e, err4 := strconv.ParseInt(strTimeSlot3_e, 16, 32) // 转16进制值
		if nil != err4 {
			log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
		}
		devUserUpload.MyTime[2].End = int32(nTimeSlot3_e)

		if toPms_byte, err1 := json.Marshal(devUserUpload); err1 == nil {
			// 需存入数据库
			rabbitmq.Publish2pms(toPms_byte, "")
		} else {
			log.Error("[", uuid, "] ParseZlockData constant.Update_dev_user toPms_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Sync_dev_user:		// 请求同步用户列表(0x31)(服务器-->前板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Sync_dev_user")
		pdu := &wlprotocol.SyncDevUserResp{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Sync_dev_user pdu.Decode, err=", err)
			return err
		}

		syncDevUser := entity.SyncDevUserResp{}
		syncDevUser.Cmd = int(wlMsg.Cmd)
		syncDevUser.Ack = int(wlMsg.Cmd)
		syncDevUser.DevType = devType
		syncDevUser.DevId = uuid
		syncDevUser.Vendor = "feibee"
		syncDevUser.SeqId = int(wlMsg.SeqId)
		syncDevUser.UserVer = pdu.DevUserVer
		syncDevUser.Num = int(pdu.DevUserNum)
		for i:=0;i<syncDevUser.Num;i++ {
			var devUser entity.DevUser
			devUser.UserId = pdu.DevUserInfos[i].UserNo
			devUser.UserType = int(pdu.DevUserInfos[i].UserType)
			devUser.Passwd = int(pdu.DevUserInfos[i].OpenBitMap & 0x01)
			devUser.Card = int(pdu.DevUserInfos[i].OpenBitMap >> 1 & 0x01)
			devUser.Finger = int((pdu.DevUserInfos[i].OpenBitMap >> 2 & 0x01) + (pdu.DevUserInfos[i].OpenBitMap >> 3 & 0x01) + (pdu.DevUserInfos[i].OpenBitMap >> 4 & 0x01))
			devUser.Ffinger = int((pdu.DevUserInfos[i].OpenBitMap >> 5 & 0x01) + (pdu.DevUserInfos[i].OpenBitMap >> 6 & 0x01))
			devUser.Face = int(pdu.DevUserInfos[i].OpenBitMap >> 7 & 0x01)
			devUser.Bluetooth = int(pdu.DevUserInfos[i].OpenBitMap >> 7 & 0x01)
			devUser.Count = int(pdu.DevUserInfos[i].PermitNum)
			devUser.Remainder = int(pdu.DevUserInfos[i].Remainder)

			// 开始日期
			// 转10进制
			mSDate := (int32(pdu.DevUserInfos[i].StartDate[0]) * 10000) + (int32(pdu.DevUserInfos[i].StartDate[1]) * 100) + int32(pdu.DevUserInfos[i].StartDate[2])
			strSDate := strconv.FormatInt(int64(mSDate), 10) // 转10进制字符串
			nSDate, err2 := strconv.ParseInt(strSDate, 16, 32) // 转16进制值
			if nil != err2 {
				log.Error("ParseZlockData strconv.ParseInt, err2: ", err2)
			}
			devUser.MyDate.Start = int32(nSDate)

			// 结束日期
			// 转10进制
			mEDate := (int32(pdu.DevUserInfos[i].EndDate[0]) * 10000) + (int32(pdu.DevUserInfos[i].EndDate[1]) * 100) + int32(pdu.DevUserInfos[i].EndDate[2])
			strEDate := strconv.FormatInt(int64(mEDate), 10) // 转10进制字符串
			nEDate, err3 := strconv.ParseInt(strEDate, 16, 32) // 转16进制值
			if nil != err3 {
				log.Error("ParseZlockData strconv.ParseInt, err3: ", err3)
			}
			devUser.MyDate.End = int32(nEDate)

			// 时段1 - 开始
			mTimeSlot1_s := (int32(pdu.DevUserInfos[i].TimeSlot1[0]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot1[1])
			strTimeSlot1_s := strconv.FormatInt(int64(mTimeSlot1_s), 10) // 转10进制字符串
			nTimeSlot1_s, err4 := strconv.ParseInt(strTimeSlot1_s, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[0].Start = int32(nTimeSlot1_s)

			// 时段1 - 结束
			mTimeSlot1_e := (int32(pdu.DevUserInfos[i].TimeSlot1[2]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot1[3])
			strTimeSlot1_e := strconv.FormatInt(int64(mTimeSlot1_e), 10) // 转10进制字符串
			nTimeSlot1_e, err4 := strconv.ParseInt(strTimeSlot1_e, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[0].End = int32(nTimeSlot1_e)

			// 时段2 - 开始
			mTimeSlot2_s := (int32(pdu.DevUserInfos[i].TimeSlot2[0]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot2[1])
			strTimeSlot2_s := strconv.FormatInt(int64(mTimeSlot2_s), 10) // 转10进制字符串
			nTimeSlot2_s, err4 := strconv.ParseInt(strTimeSlot2_s, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[1].Start = int32(nTimeSlot2_s)

			// 时段2 - 结束
			mTimeSlot2_e := (int32(pdu.DevUserInfos[i].TimeSlot2[2]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot2[3])
			strTimeSlot2_e := strconv.FormatInt(int64(mTimeSlot2_e), 10) // 转10进制字符串
			nTimeSlot2_e, err4 := strconv.ParseInt(strTimeSlot2_e, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[1].End = int32(nTimeSlot2_e)

			// 时段3 - 开始
			mTimeSlot3_s := (int32(pdu.DevUserInfos[i].TimeSlot3[0]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot3[1])
			strTimeSlot3_s := strconv.FormatInt(int64(mTimeSlot3_s), 10) // 转10进制字符串
			nTimeSlot3_s, err4 := strconv.ParseInt(strTimeSlot3_s, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[2].Start = int32(nTimeSlot3_s)

			// 时段3 - 结束
			mTimeSlot3_e := (int32(pdu.DevUserInfos[i].TimeSlot3[2]) * 100) + int32(pdu.DevUserInfos[i].TimeSlot3[3])
			strTimeSlot3_e := strconv.FormatInt(int64(mTimeSlot3_e), 10) // 转10进制字符串
			nTimeSlot3_e, err4 := strconv.ParseInt(strTimeSlot3_e, 16, 32) // 转16进制值
			if nil != err4 {
				log.Error("ParseZlockData strconv.ParseInt, err4: ", err4)
			}
			devUser.MyTime[2].End = int32(nTimeSlot3_e)

			syncDevUser.UserList = append(syncDevUser.UserList, devUser)
		}

		if toPms_byte, err1 := json.Marshal(syncDevUser); err1 == nil {
			rabbitmq.Publish2pms(toPms_byte, "")
		} else {
			log.Error("[", uuid, "] toPms_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Remote_open:			// 远程开锁命令(0x52)(服务器->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Remote_open")
		pdu := &wlprotocol.RemoteOpenLockResp{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Remote_open pdu.Decode, err=", err)
			return err
		}

		// 组包
		remoteOpenLockResp := entity.RemoteOpenLockResp{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			UserId: pdu.UserNo,
			UserId2: pdu.UserNo2,	// 用户id不存在时，用0xffff填写
			Time: pdu.Time,
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(remoteOpenLockResp); err1 == nil {
			if 0 != wlMsg.Ack {
				rabbitmq.Publish2app(to_byte, uuid)
			}

			if 1 == wlMsg.Ack { // 开门成功才记录远程开锁记录
				rabbitmq.Publish2pms(to_byte, "")
			}
		} else {
			log.Error("[", uuid, "] constant.Remote_open, err=", err1)
			return err1
		}
	case constant.Upload_dev_info:		// 发送设备信息(0x70)(前板，后板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Upload_dev_info")
		//1. 回复锁
		tPdu := &wlprotocol.UploadDevInfoResp{
			Time: int32(time.Now().Unix()),
		}
		wlMsg.Ack = 1
		bData, err_ := wlMsg.PkEncode(tPdu)
		if nil != err_ {
			log.Error("ParseZlockData() Upload_dev_info wlMsg.PkEncode, error: ", err_)
			return err_
		}
		go procLock.Cmd2Device(uuid, hex.EncodeToString(bData), "constant.Upload_dev_info resp")

		//2. 解包体
		pdu := &wlprotocol.UploadZigbeeDevInfo{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Upload_dev_info pdu.Decode, err=", err)
			return err
		}

		//3. 组json包
		uploadDevInfo := entity.UploadDevInfo{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}

		uploadDevInfo.UserVer =	pdu.DevUserVer	// 设备用户版本号，如果是0则不需要发起同步请求
		fVer := fmt.Sprintf("V%d.%d.%d", pdu.FMainVer, pdu.FSubVer, pdu.FModVer)
		uploadDevInfo.FVer = fVer		// 前板版本号
		uploadDevInfo.FType = strconv.Itoa(int(pdu.FType))		// 前板型号（Z201)
		uploadDevInfo.HasScr = pdu.IsHasScr		// 是否带屏幕（1-带屏幕，0-不带屏幕）
		uploadDevInfo.Battery =	pdu.Battery	// 电池电量
		uploadDevInfo.VolumeLevel =	pdu.Volume	// 音量等级(带屏幕的锁，可以设置为静音，1-3音量等级，3表示音量最大)
		uploadDevInfo.PasswdSwitch = pdu.PwdSwitch	// 密码开关（0：无法使用密码开锁，1：可以使用密码开锁）
		uploadDevInfo.SinMul = pdu.SinMul	// 开门模式（1：表示单人模式, 2：表示双人模式）
		bVer := fmt.Sprintf("V%d.%d.%d", pdu.BMainVer, pdu.BSubVer, pdu.BModVer)
		uploadDevInfo.BVer = bVer			// 后板版本号
		uploadDevInfo.NbVer = ""			// NB版本号
		uploadDevInfo.Sim  = ""			// SIM卡号
		uploadDevInfo.OpenMode = pdu.OpenMode		// 常开模式
		uploadDevInfo.RemoteSwitch = pdu.RemoteSwitch	// 远程开关（0：无法使用远程开锁，1：可以使用远程开锁）
		uploadDevInfo.ActiveMode = pdu.ActiveMode	// 远程开锁激活方式，0：门锁唤醒后立即激活，1：输入激活码激活
		uploadDevInfo.NolockSwitch = pdu.NolockSwitch	// 门未关开关，0：关闭，1：开启
		uploadDevInfo.FakelockSwitch = pdu.FakelockSwitch	// 假锁开关，0：关闭，1：开启
		uploadDevInfo.InfraSwitch =	pdu.InfraSwitch // 人体感应报警开关，0：关闭，1：唤醒，但不推送消息，2：唤醒并且推送消息
		uploadDevInfo.InfraTime = pdu.InfraTime		// 人体感应报警，红外持续监测到多少秒 就上报消息
		uploadDevInfo.AlarmSwitch =	pdu.AlarmSwitch // 报警类型开关，0：关闭，1：拍照+录像，2：拍照
		uploadDevInfo.BellSwitch = pdu.BellSwitch	// 门铃开关 0：关闭，1：开启
		uploadDevInfo.FBreakSwitch = pdu.FBreakSwitch	// 防拆报警开关：0关闭，1开启
		uploadDevInfo.Capability = pdu.Capability			// 能力集

		// 亿速码安全芯片相关参数
		uploadDevInfo.UId =	"" 			// 安全芯片id
		uploadDevInfo.ProjectNo = ""		// 项目编号
		uploadDevInfo.MerChantNo = ""		// 商户号
		uploadDevInfo.Random =	""		// 安全芯片随机数

		// 兼容字段，某些功能不支持的NB锁
		uploadDevInfo.Unsupport = 0 		// 0-所有功能支持，1-临时用户时段不支持

		//4. 发送到PMS模块
		if to_byte, err1 := json.Marshal(uploadDevInfo); err1 == nil {
			//producer.SendMQMsg2PMS(string(to_byte))
			rabbitmq.Publish2pms(to_byte, "")
		} else {
			log.Error("[", uuid, "] constant.Upload_dev_info, err=", err1)
			return err1
		}
	case constant.Set_dev_para:			// 设置参数(0x72)(服务器-->前板，后板)
		log.Info("[", uuid, "] ParseZlockData constant.Set_dev_para")
		pdu := &wlprotocol.ParamUpdate{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Set_dev_para pdu.Decode, err=", err)
			return err
		}

		// 组包
		lockParam := entity.LockParam{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			ParaNo: pdu.ParamNo,
			PaValue: pdu.ParamValue,
			PaValue2: pdu.ParamValue2,
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(lockParam); err1 == nil {
			// 回复到APP
			//rabbitmq.Publish2app(to_byte, uuid)
			if 1 == wlMsg.Ack { // 设置成功存入DB
				rabbitmq.Publish2pms(to_byte, "")
			} else {
				rabbitmq.Publish2app(to_byte, uuid)
			}
		} else {
			log.Error("[", uuid, "] constant.Set_dev_para, err=", err1)
			return err1
		}
	case constant.Update_dev_para:		// 参数更新(0x73)(前板,后板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Update_dev_para")
		pdu := &wlprotocol.ParamUpdate{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Update_dev_para pdu.Decode, err=", err)
			return err
		}

		// 组包
		lockParam := entity.LockParam{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			ParaNo: pdu.ParamNo,
			PaValue: pdu.ParamValue,
			PaValue2: pdu.ParamValue2,
		}

		if constant.IPC_SN_PNO == pdu.ParamNo || constant.WIFI_SSID_PNO == pdu.ParamNo || constant.PROJECT_No_PNO == pdu.ParamNo {
			var byteData []byte
			rbyf_pn := make([]byte, 32, 32)    //make语法声明 ，len为32，cap为32
			paramValue, ok := pdu.ParamValue.(string)
			if !ok {
				log.Error("ParseZlockData Update_dev_para pdu.ParamValue.(string), ok=", ok)
				return nil
			}
			for m:=0;m<len(paramValue);m++{
				if m >= 32 {
					break
				}
				byteData =  append(byteData, paramValue[m])
			}
			index := bytes.IndexByte(byteData, 0)
			if -1 == index {
				rbyf_pn = byteData[0:len(byteData)]
			} else {
				rbyf_pn = byteData[0:index]
			}
			lockParam.PaValue = string(rbyf_pn[:])
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(lockParam); err1 == nil {
			// 回复到APP
			//rabbitmq.Publish2app(to_byte, uuid)
			rabbitmq.Publish2pms(to_byte, "")

		} else {
			log.Error("[", uuid, "] constant.Update_dev_para, err=", err1)
			return err1
		}
	case constant.Soft_reset:			// 软件重启命令(0x74)(服务器-->前、后板)
		log.Info("[", uuid, "] ParseZlockData constant.Soft_reset")

		// 组包
		softReset := entity.Header{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(softReset); err1 == nil {
			rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Soft_reset, err=", err1)
			return err1
		}
	case constant.Factory_reset:		// 恢复出厂化(0xEA)( 服务器-->前、后板)
		log.Info("[", uuid, "] ParseZlockData constant.Factory_reset")

		// 组包
		factoryReset := entity.Header{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}

		//2. 发送到PMS模块
		if to_byte, err1 := json.Marshal(factoryReset); err1 == nil {
			rabbitmq.Publish2app(to_byte, uuid)

			// PMS初始化设备信息的参数
			rabbitmq.Publish2pms(to_byte, "")
		} else {
			log.Error("[", uuid, "] constant.Factory_reset, err=", err1)
			return err1
		}
	case constant.Upload_open_log:	// 用户开锁消息上报(0x40)(前板--->服务器)
		log.Info("[", uuid, "] ParseZlockData Upload_open_log")
		pdu := &wlprotocol.OpenLockMsg{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Upload_open_log pdu.Decode, err=", err)
			return err
		}

		//2. 发送到PMS模块
		openLogUpload := entity.UploadOpenLockLog{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			UserVer: pdu.DevUserVer,
			UserNum: pdu.UserNum,
			Battery: int(pdu.Battery),
		}
		var lockLog entity.OpenLockLog
		lockLog.UserId = pdu.UserNo 		// 设备用户ID
		lockLog.MainOpen = pdu.MainOpen 	// 主开锁方式（1-密码，2-刷卡，3-指纹）
		lockLog.SubOpen = pdu.SubOpen   	// 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
		lockLog.SinMul = pdu.SinMul			// 开门模式（1：表示单人模式, 2：表示双人模式）
		lockLog.Remainder = pdu.Remainder 	// 0表示成功，1表示失败
		lockLog.Time = pdu.Time
		openLogUpload.LogList = append(openLogUpload.LogList, lockLog)
		if 2 == pdu.SinMul { // 双人模式
			lockLog.UserId = pdu.UserNo2       	// 设备用户ID
			lockLog.MainOpen = pdu.MainOpen2   	// 主开锁方式（1-密码，2-刷卡，3-指纹）
			lockLog.SubOpen = pdu.SubOpen2     	// 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
			lockLog.SinMul = pdu.SinMul     	// 开门模式（1：表示单人模式, 2：表示双人模式）
			lockLog.Remainder = pdu.Remainder2 	// 0表示成功，1表示失败
			lockLog.Time = pdu.Time
			openLogUpload.LogList = append(openLogUpload.LogList, lockLog)
		}

		if to_byte, err1 := json.Marshal(openLogUpload); err1 == nil {
			rabbitmq.Publish2pms(to_byte, "")

			rabbitmq.Publish2mns(to_byte, "")
		} else {
			log.Error("[", uuid, "] constant.Upload_open_log, Uplocal_open_log, err=", err1)
			return err1
		}
	case constant.UpEnter_menu_log:	// 用户进入菜单上报(0x42)(前板--->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.UpEnter_menu_log")
		pdu := &wlprotocol.EnterMenuMsg{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData UpEnter_menu_log pdu.Decode, err=", err)
			return err
		}

		//2. 发送到PMS模块
		enterMenuUpload := entity.UploadEnterMenuLog{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			UserVer: pdu.DevUserVer,
			Battery: int(pdu.Battery),
		}
		var enterMenu entity.EnterMenu
		enterMenu.UserId = pdu.UserNo 		// 设备用户ID
		enterMenu.MainOpen = pdu.MainOpen 	// 主开锁方式（1-密码，2-刷卡，3-指纹）
		enterMenu.SubOpen = pdu.SubOpen   	// 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
		enterMenu.SinMul = pdu.SinMul			// 开门模式（1：表示单人模式, 2：表示双人模式）
		enterMenu.Time = pdu.Time
		enterMenuUpload.LogList = append(enterMenuUpload.LogList, enterMenu)
		if 2 == pdu.SinMul { // 双人模式
			enterMenu.UserId = pdu.UserNo2       	// 设备用户ID
			enterMenu.MainOpen = pdu.MainOpen2   	// 主开锁方式（1-密码，2-刷卡，3-指纹）
			enterMenu.SubOpen = pdu.SubOpen2     	// 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
			enterMenu.SinMul = pdu.SinMul     	// 开门模式（1：表示单人模式, 2：表示双人模式）
			enterMenu.Time = pdu.Time
			enterMenuUpload.LogList = append(enterMenuUpload.LogList, enterMenu)
		}

		if to_byte, err1 := json.Marshal(enterMenuUpload); err1 == nil {
			rabbitmq.Publish2pms(to_byte, "")
		} else {
			log.Error("[", uuid, "] constant.Uplocal_open_log, err=", err1)
			return err1
		}
	case constant.Infrared_alarm, constant.Noatmpt_alarm, constant.Forced_break_alarm, constant.Fakelock_alarm, constant.Nolock_alarm:
		// 人体感应报警(0x39)(前板-->服务器) // 非法操作报警(0x20)(前板--->服务器) // 强拆报警(0x22)(前板--->服务器) // 假锁报警(0x24)(前板--->服务器) // 门未关报警(0x26)(前板--->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Infrared_alarm, Noatmpt_alarm, Forced_break_alarm, Fakelock_alarm, Nolock_alarm")
		pdu := &wlprotocol.Alarms{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Infrared_alarm pdu.Decode, err=", err)
			return err
		}

		//2. 发送到PMS模块
		alarmMsg := entity.AlarmMsg{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			Time: pdu.Time,
		}
		if to_byte, err1 := json.Marshal(alarmMsg); err1 == nil {
			rabbitmq.Publish2pms(to_byte, "")

			rabbitmq.Publish2mns(to_byte, "") // MNS
		} else {
			log.Error("[", uuid, "] constant.Infrared_alarm, Noatmpt_alarm, Forced_break_alarm, Fakelock_alarm, Nolock_alarm to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Low_battery_alarm:	// 低压报警(0x2A)(前板--->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Low_battery_alarm")
		pdu := &wlprotocol.LowBattAlarm{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Low_battery_alarm pdu.Decode, err=", err)
			return err
		}

		//2. 发送到PMS模块
		doorBellCall := entity.AlarmMsgBatt{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			Value: int(pdu.Battery),
			Time: pdu.Time,
		}
		if to_byte, err1 := json.Marshal(doorBellCall); err1 == nil {
			rabbitmq.Publish2pms(to_byte, "")

			rabbitmq.Publish2mns(to_byte, "") // MNS
		} else {
			log.Error("[", uuid, "] constant.Low_battery_alarm, err=", err1)
			return err1
		}
	case constant.Lock_PIC_Upload:		// 图片上传(0x2F)(前板--->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Lock_PIC_Upload")
		pdu := &wlprotocol.PicUpload{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Lock_PIC_Upload pdu.Decode, err=", err)
			return err
		}

		//2. 发送到PMS模块
		picUpload := entity.PicUpload{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			CmdType: int(pdu.CmdType),
			TimeId: int(pdu.MsgId),
			PicName: pdu.PicPath,
		}
		if to_byte, err1 := json.Marshal(picUpload); err1 == nil {
			rabbitmq.Publish2pms(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Lock_PIC_Upload, err=", err1)
			return err1
		}
	case constant.Upload_lock_active:	// 在线离线(0x46)(后板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Upload_lock_active")
		pdu := &wlprotocol.OnOffLine{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Upload_lock_active pdu.Decode, err=", err)
			return err
		}

		// 组包
		deviceActive := entity.DeviceActive{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}
		if 1 == pdu.OnOff {
			deviceActive.Time = 1
		} else {
			deviceActive.Time = 0
		}

		//2. 锁唤醒，存入redis
		redis.SetActTimePool(uuid, int64(deviceActive.Time))

		//3. 发送
		if to_byte, err1 := json.Marshal(deviceActive); err1 == nil {
			// 回复到APP
			rabbitmq.Publish2app(to_byte, uuid)

			// 到PMS模块
			rabbitmq.Publish2pms(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Upload_lock_active, err=", err1)
			return err1
		}
	case constant.Real_Video:			// 实时视频(0x36)(服务器-->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Real_Video")
		pdu := &wlprotocol.RealVideo{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Real_Video pdu.Decode, err=", err)
			return err
		}

		realVideo := entity.RealVideoLock{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			Act: pdu.Act,
		}
		if to_byte, err1 := json.Marshal(realVideo); err1 == nil {
			rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Real_Video to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Set_Wifi:				// Wifi设置(0x37)(服务器-->前板)
		log.Info("[", uuid, "] ParseZlockData constant.Set_Wifi")
		pdu := &wlprotocol.WiFiSet{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Set_Wifi pdu.Decode, err=", err)
			return err
		}

		setLockWiFi := entity.SetLockWiFi{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			//WifiSsid: string(pdu.Ssid[:]),
			//WifiPwd: string(pdu.Passwd[:]),
		}

		var byteData []byte
		rbyf_pn := make([]byte, 32, 32)    //make语法声明 ，len为32，cap为32
		for m:=0;m<len(pdu.Ssid);m++{
			if m >= 32 {
				break
			}
			byteData =  append(byteData, pdu.Ssid[m])
		}
		index := bytes.IndexByte(byteData, 0)
		if -1 == index {
			rbyf_pn = byteData[0:len(byteData)]
		} else {
			rbyf_pn = byteData[0:index]
		}
		setLockWiFi.WifiSsid = string(rbyf_pn[:])

		byteData = byteData[0:0]
		for m:=0;m<len(pdu.Passwd);m++{
			if m >= 32 {
				break
			}
			byteData = append(byteData, pdu.Passwd[m])
		}
		index = bytes.IndexByte(byteData, 0)
		if -1 == index {
			rbyf_pn = byteData[0:len(byteData)]
		} else {
			rbyf_pn = byteData[0:index]
		}
		setLockWiFi.WifiPwd = string(rbyf_pn[:])

		if _, err1 := json.Marshal(setLockWiFi); err1 == nil {
			//rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Set_Wifi to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Notify_Set_Wifi:
		log.Info("[", uuid, "] ParseZlockData constant.Notify_Set_Wifi")
		notifyHead := entity.Header {
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),
		}
		if to_byte, err1 := json.Marshal(notifyHead); err1 == nil {
			rabbitmq.Publish2app(to_byte, uuid)
		} else {
			log.Error("[", uuid, "] constant.Notify_Set_Wifi to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Door_Call:			// 门铃呼叫(0x38)(前板-->服务器)
		log.Info("[", uuid, "] ParseZlockData constant.Door_Call")
		pdu := &wlprotocol.DoorbellCall{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Door_Call pdu.Decode, err=", err)
			return err
		}

		// 发送到PMS模块
		doorBellCall := entity.DoorBellCall{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			Time: pdu.Time,
		}
		if to_byte, err1 := json.Marshal(doorBellCall); err1 == nil {

			rabbitmq.Publish2pms(to_byte, "")

			rabbitmq.Publish2mns(to_byte, "") // MNS
		} else {
			log.Error("[", uuid, "] constant.Door_Call to_byte json.Marshal, err=", err1)
			return err1
		}
	case constant.Door_State: // 锁状态上报
		log.Info("[", uuid, "] ParseZlockData constant.Door_State")

		pdu := &wlprotocol.DoorStateUpload{}
		err = pdu.Decode(bBody, uuid)
		if nil != err {
			log.Error("ParseZlockData Door_State pdu.Decode, err=", err)
			return err
		}

		// 发送到PMS模块
		doorState := entity.DoorStateUpload{
			Cmd: int(wlMsg.Cmd),
			Ack: int(wlMsg.Ack),
			DevType: devType,
			DevId: uuid,
			Vendor: "feibee",
			SeqId: int(wlMsg.SeqId),

			State: pdu.State,
		}

		if to_byte, err1 := json.Marshal(doorState); err1 == nil {
			//2. 推到APP
			rabbitmq.Publish2app(to_byte, uuid)

			//3. 需要存到mongodb
			rabbitmq.Publish2pms(to_byte, "")
		} else {
			log.Error("[", uuid, "] constant.Door_State to_byte json.Marshal, err=", err1)
			return err1
		}
	default:
		log.Info("[", uuid, "] ParseZlockData Default, Cmd=", wlMsg.Cmd)
	}

	return nil
}
