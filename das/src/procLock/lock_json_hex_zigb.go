package procLock

import (
	"das/core/util"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"das/core/constant"
	"das/core/entity"
	"das/core/log"
	"das/core/wlprotocol"
)

func WlJson2BinMsgZigbee(jsonMsg string, wifiData uint8) ([]byte, error) {
	// 1、解析消息
	var head entity.Header
	if err := json.Unmarshal([]byte(jsonMsg), &head); err != nil {
		log.Error("ProcAppMsg json.Unmarshal Header error, err=", err)
		return nil, err
	}
	sendMQTTDownLogMsg(head.DevId, jsonMsg)

	//1、飞比设备编号去掉下划线
	retUuid := make([]string, 4)
	if "" != head.DevId { // uuid不能为空
		retUuid = strings.FieldsFunc(head.DevId, util.Split) // 去掉下划线后边，如：_01
	} else {
		return nil, errors.New("feibee device uuid is nil.")
	}

	wlMsg := &wlprotocol.WlZigbeeMsg{
		Started: wlprotocol.ZbStarted, // 开始标志
		Version: wlprotocol.Version, // 协议版本号
		SeqId:   uint16(head.SeqId), // 包序列号
		Cmd:     uint8(head.Cmd),    // 命令
		Ack:     uint8(head.Ack),    // 回应标志
		Type:    1,                  // 设备类型
		Uuid:	retUuid[0],			// 飞比zigbee设备编号
		Ended: wlprotocol.ZbEnded, // 结束标志
	}

	fmt.Println(wlMsg)

	switch head.Cmd {
	case constant.Add_dev_user: // 添加设备用户
		//log.Info("[", head.DevId, "] constant.Add_dev_user")

		var addDevUser entity.AddDevUser
		if err := json.Unmarshal([]byte(jsonMsg), &addDevUser); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		nRandom, err1 := strconv.ParseUint(addDevUser.UserNote, 16, 32)
		if err1 != nil {
			log.Error("WlJson2BinMsg() strconv.ParseUint: ", addDevUser.UserNote, ", error: ", err1)
		}
		// APP用户映射的字段
		appUser, err2 := strconv.ParseUint(addDevUser.AppUser, 16, 32)
		if err2 != nil {
			log.Error("WlJson2BinMsg() strconv.ParseUint: ", addDevUser.AppUser, ", error: ", err2)
		}

		pdu := &wlprotocol.AddDevUserZigbee {
			UserNo:   addDevUser.UserId,   // 设备用户编号，指定操作的用户编号，如果是0XFFFF表示新添加一个用户
			MainOpen: addDevUser.MainOpen, // 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
			SubOpen:  addDevUser.SubOpen,  // 是否胁迫，是否胁迫：0-正常，1-胁迫
			UserType: addDevUser.UserType, // 用户类型(1)，用户类型:  0 - 管理员，1 - 普通用户，2 - 临时用户
			// Passwd: addDevUser.Passwd,		// 密码(6)，密码开锁方式，目前是6个字节.如果添加的是其他验证方式,则为0xff.密码位数少于10位时,多余的填0xff
			UserNote:  int32(nRandom),   // 用户别名-时间戳存在redis中key-value对应 时间戳的16进制作为随机数
			PermitNum: addDevUser.Count, // 允许开门次数
			AppUser: int32(appUser),	// APP用户账号-时间戳存在redis中key-value对应 时间戳的16进制作为随机数
		}

		// 密码
		pwd := []byte(addDevUser.Passwd)
		for i := 0; i < len(pwd); i++ {
			if i < 6 {
				pdu.Passwd[i] = pwd[i]
			}
		}

		addDevUser.MyDate.Start = convertHexDateTime(addDevUser.MyDate.Start)
		pdu.StartDate[0] = uint8(addDevUser.MyDate.Start / 10000) // 开始日期
		pdu.StartDate[1] = uint8(addDevUser.MyDate.Start / 100 % 100)
		pdu.StartDate[2] = uint8(addDevUser.MyDate.Start % 100)

		addDevUser.MyDate.End = convertHexDateTime(addDevUser.MyDate.End)
		pdu.EndDate[0] = uint8(addDevUser.MyDate.End / 10000) // 结束日期
		pdu.EndDate[1] = uint8(addDevUser.MyDate.End / 100 % 100)
		pdu.EndDate[2] = uint8(addDevUser.MyDate.End % 100)

		addDevUser.MyTime[0].Start = convertHexDateTime(addDevUser.MyTime[0].Start)
		addDevUser.MyTime[0].End = convertHexDateTime(addDevUser.MyTime[0].End)
		pdu.TimeSlot1[0] = uint8(addDevUser.MyTime[0].Start / 100) // 小时
		pdu.TimeSlot1[1] = uint8(addDevUser.MyTime[0].Start % 100) // 分钟
		pdu.TimeSlot1[2] = uint8(addDevUser.MyTime[0].End / 100)   // 小时
		pdu.TimeSlot1[3] = uint8(addDevUser.MyTime[0].End % 100)   // 分钟

		addDevUser.MyTime[1].Start = convertHexDateTime(addDevUser.MyTime[1].Start)
		addDevUser.MyTime[1].End = convertHexDateTime(addDevUser.MyTime[1].End)
		pdu.TimeSlot2[0] = uint8(addDevUser.MyTime[1].Start / 100) // 小时
		pdu.TimeSlot2[1] = uint8(addDevUser.MyTime[1].Start % 100) // 分钟
		pdu.TimeSlot2[2] = uint8(addDevUser.MyTime[1].End / 100)   // 小时
		pdu.TimeSlot2[3] = uint8(addDevUser.MyTime[1].End % 100)   // 分钟

		addDevUser.MyTime[2].Start = convertHexDateTime(addDevUser.MyTime[2].Start)
		addDevUser.MyTime[2].End = convertHexDateTime(addDevUser.MyTime[2].End)
		pdu.TimeSlot3[0] = uint8(addDevUser.MyTime[2].Start / 100) // 小时
		pdu.TimeSlot3[1] = uint8(addDevUser.MyTime[2].Start % 100) // 分钟
		pdu.TimeSlot3[2] = uint8(addDevUser.MyTime[2].End / 100)   // 小时
		pdu.TimeSlot3[3] = uint8(addDevUser.MyTime[2].End % 100)   // 分钟

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Add_dev_user wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Set_dev_user_temp: // 设置临时用户
		//log.Info("[", head.DevId, "] constant.Set_dev_user_temp")
		var setTmpDevUser entity.SetTmpDevUser
		if err := json.Unmarshal([]byte(jsonMsg), &setTmpDevUser); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		pdu := &wlprotocol.SetTmpDevUser{
			UserNo:    setTmpDevUser.UserId, // 设备用户编号，指定操作的用户编号，如果是0XFFFF表示新添加一个用户
			PermitNum: setTmpDevUser.Count,  // 允许开门次数
		}

		setTmpDevUser.MyDate.Start = convertHexDateTime(setTmpDevUser.MyDate.Start)
		pdu.StartDate[0] = uint8(setTmpDevUser.MyDate.Start / 10000) // 开始日期
		pdu.StartDate[1] = uint8(setTmpDevUser.MyDate.Start / 100 % 100)
		pdu.StartDate[2] = uint8(setTmpDevUser.MyDate.Start % 100)

		setTmpDevUser.MyDate.End = convertHexDateTime(setTmpDevUser.MyDate.End)
		pdu.EndDate[0] = uint8(setTmpDevUser.MyDate.End / 10000) // 结束日期
		pdu.EndDate[1] = uint8(setTmpDevUser.MyDate.End / 100 % 100)
		pdu.EndDate[2] = uint8(setTmpDevUser.MyDate.End % 100)

		setTmpDevUser.MyTime[0].Start = convertHexDateTime(setTmpDevUser.MyTime[0].Start)
		setTmpDevUser.MyTime[0].End = convertHexDateTime(setTmpDevUser.MyTime[0].End)
		pdu.TimeSlot1[0] = uint8(setTmpDevUser.MyTime[0].Start / 100) // 小时
		pdu.TimeSlot1[1] = uint8(setTmpDevUser.MyTime[0].Start % 100) // 分钟
		pdu.TimeSlot1[2] = uint8(setTmpDevUser.MyTime[0].End / 100)   // 小时
		pdu.TimeSlot1[3] = uint8(setTmpDevUser.MyTime[0].End % 100)   // 分钟

		setTmpDevUser.MyTime[1].Start = convertHexDateTime(setTmpDevUser.MyTime[1].Start)
		setTmpDevUser.MyTime[1].End = convertHexDateTime(setTmpDevUser.MyTime[1].End)
		pdu.TimeSlot2[0] = uint8(setTmpDevUser.MyTime[1].Start / 100) // 小时
		pdu.TimeSlot2[1] = uint8(setTmpDevUser.MyTime[1].Start % 100) // 分钟
		pdu.TimeSlot2[2] = uint8(setTmpDevUser.MyTime[1].End / 100)   // 小时
		pdu.TimeSlot2[3] = uint8(setTmpDevUser.MyTime[1].End % 100)   // 分钟

		setTmpDevUser.MyTime[2].Start = convertHexDateTime(setTmpDevUser.MyTime[2].Start)
		setTmpDevUser.MyTime[2].End = convertHexDateTime(setTmpDevUser.MyTime[2].End)
		pdu.TimeSlot3[0] = uint8(setTmpDevUser.MyTime[2].Start / 100) // 小时
		pdu.TimeSlot3[1] = uint8(setTmpDevUser.MyTime[2].Start % 100) // 分钟
		pdu.TimeSlot3[2] = uint8(setTmpDevUser.MyTime[2].End / 100)   // 小时
		pdu.TimeSlot3[3] = uint8(setTmpDevUser.MyTime[2].End % 100)   // 分钟

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Set_dev_user_temp wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Del_dev_user: // 删除设备用户
		//log.Info("[", head.DevId, "] constant.Del_dev_user")

		var delDevUser entity.DelDevUser
		if err := json.Unmarshal([]byte(jsonMsg), &delDevUser); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		// APP用户映射的字段
		appUser, err2 := strconv.ParseUint(delDevUser.AppUser, 16, 32)
		if err2 != nil {
			log.Error("WlJson2BinMsg() strconv.ParseUint: ", delDevUser.AppUser, ", error: ", err2)
		}

		pdu := &wlprotocol.DelDevUser{
			UserNo:   delDevUser.UserId,   // 设备用户编号，指定操作的用户编号，如果是0XFFFF表示新添加一个用户
			MainOpen: delDevUser.MainOpen, // 允许开门次数
			SubOpen:  delDevUser.SubOpen,  // 是否胁迫，是否胁迫：0-正常，1-胁迫
			Time:     delDevUser.Time,     // 时间戳
			AppUser: int32(appUser),
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Set_dev_user_temp wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Sync_dev_user: // 同步设备用户列表
		//1. 设备用户同步
		//log.Info("[", head.DevId, "] constant.Sync_dev_user")
		var syncDevUser entity.SyncDevUserReq
		if err := json.Unmarshal([]byte(jsonMsg), &syncDevUser); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		pdu := &wlprotocol.SyncDevUser{
			Num: syncDevUser.Num,
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Sync_dev_user wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Remote_open: // 远程开锁
		//log.Info("[", head.DevId, "] constant.Remote_open")

		var remoteOpen entity.MRemoteOpenLockReq
		if err := json.Unmarshal([]byte(jsonMsg), &remoteOpen); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		// APP用户映射的字段
		appUser, err2 := strconv.ParseUint(remoteOpen.AppUser, 16, 32)
		if err2 != nil {
			log.Error("WlJson2BinMsg() strconv.ParseUint: ", remoteOpen.AppUser, ", error: ", err2)
		}

		pdu := &wlprotocol.RemoteOpenLock{
			/*Passwd: (remoteOpen.Passwd),	// 密码1（6）
			Passwd2: remoteOpen.Passwd2,	// 密码2（6）*/
			AppUser: int32(appUser),
		}

		nTime, ok := remoteOpen.Time.(float64) // 随机数（4）
		if ok {
			pdu.Time = int32(nTime)
		} else {
			log.Error("WlJson2BinMsg remoteOpen.Time.(float64) error, ok=", ok)
			pdu.Time = int32(time.Now().Unix())
		}
		pwd := []byte(remoteOpen.Passwd)
		for i := 0; i < len(pwd); i++ {
			if i < 6 {
				pdu.Passwd[i] = pwd[i]
			}
		}

		if "" == remoteOpen.Passwd2 { // 单人模式
			pdu.Passwd2[0] = 0xFF
		} else {
			pwd2 := []byte(remoteOpen.Passwd2)
			for i := 0; i < len(pwd2); i++ {
				if i < 6 {
					pdu.Passwd2[i] = pwd2[i]
				}
			}
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Remote_open wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Set_dev_para: // 设置设备参数
		//log.Info("[", head.DevId, "] constant.Set_dev_para")

		var setParam entity.SetLockParamReq
		if err := json.Unmarshal([]byte(jsonMsg), &setParam); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		// APP用户映射的字段
		appUser, err2 := strconv.ParseUint(setParam.AppUser, 16, 32)
		if err2 != nil {
			log.Error("WlJson2BinMsg() strconv.ParseUint: ", setParam.AppUser, ", error: ", err2)
		}

		if 0x0b != setParam.ParaNo { // 红外感应设置多久时间后拍照，其他的参数设置均为0xFF
			setParam.PaValue2 = 0xFF
		}
		pdu := &wlprotocol.SetLockParamReq{
			ParamNo:     setParam.ParaNo,   // 参数编号(1)
			ParamValue:  setParam.PaValue,  // 参数值(1)
			ParamValue2: setParam.PaValue2, // 参数值2(1)
			// Time: setParam.Time,			// 时间(4)
			AppUser: int32(appUser),
		}
		switch setParam.Time.(type) {
		case string:
			strTimeV, ok := setParam.Time.(string)
			if ok {
				if nTime, err_0 := strconv.Atoi(strTimeV); err_0 == nil {
					pdu.Time = int32(nTime)
				}
			}
		default:
			nTimeV, ok := setParam.Time.(float64)
			if ok {
				pdu.Time = int32(nTimeV)
			}
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Set_dev_para wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Soft_reset: // 软件复位
		//log.Info("[", head.DevId, "] constant.Soft_reset")

		bData, err_ := wlMsg.PkEncode(nil)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Soft_reset wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Factory_reset: // 恢复出厂设置
		//log.Info("[", head.DevId, "] constant.Factory_reset")

		bData, err_ := wlMsg.PkEncode(nil)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Factory_reset wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Real_Video: // 实时视频
		//log.Info("[", head.DevId, "] constant.Real_Video")

		var realVideo entity.RealVideoLock
		if err := json.Unmarshal([]byte(jsonMsg), &realVideo); err != nil {
			log.Error("WlJson2BinMsg json.Unmarshal Header error, err=", err)
			return nil, err
		}

		pdu := &wlprotocol.RealVideo{
			Act: realVideo.Act,
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Factory_reset wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Set_Wifi: // Wifi设置
		//log.Info("[", head.DevId, "] constant.Set_Wifi")

		var setWifi entity.SetLockWiFi
		if err := json.Unmarshal([]byte(jsonMsg), &setWifi); err != nil {
			log.Error("WlJson2BinMsgZigbee json.Unmarshal Header error, err=", err)
			return nil, err
		}

		// 类型（1）：1：ssid；2：密码
		var pdu wlprotocol.IPdu
		if 1 == wifiData {
			var ssid [32]byte
			wifiSsid := []byte(setWifi.WifiSsid)
			for i := 0; i < len(wifiSsid); i++ {
				if i < 32 {
					ssid[i] = wifiSsid[i]
				}
			}

			pdu = &wlprotocol.WiFiSetZigbeeSsid {
				DType: wifiData,
				Ssid: ssid,	// Ssid（32）
			}
		} else if 2 == wifiData {
			var pwd [32]byte
			wifiPwd := []byte(setWifi.WifiPwd)
			for i := 0; i < len(wifiPwd); i++ {
				if i < 32 {
					pwd[i] = wifiPwd[i]
				}
			}

			pdu = &wlprotocol.WiFiSetZigbeePwd{
				DType: wifiData,
				Passwd: pwd, // Passwd（32）
			}
		}

		bData, err_ := wlMsg.PkEncode(pdu)
		if nil != err_ {
			log.Error("WlJson2BinMsg() Set_Wifi wlMsg.PkEncode, error: ", err_)
			return nil, err_
		}
		return bData, nil
	case constant.Notify_F_Upgrade: // 通知前板升级（APP—后台—>锁）
		{
			//log.Info("[", head.DevId, "] constant.Notify_F_Upgrade")
		}
	case constant.Notify_B_Upgrade: // 通知后板升级（APP—后台—>锁）
		{
			//log.Info("[", head.DevId, "] constant.Notify_B_Upgrade")
		}
	default:
		log.Info("[", head.DevId, "] Default, Cmd=", head.Cmd)
	}

	return nil, nil
}

