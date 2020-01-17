package wlprotocol

// 开始，结束标志
const (
	Started = 0xAA
	Ended = 0x55
	Version = 0x01
)

// 设备编号（为了兼容，设备编号带上设备编号的长度）
type DeviceId struct {
	Len uint8
	Uuid string
}

// 王力消息体
type WlMessage struct {
	Started uint8		// 开始标志
	Version uint16		// 协议版本号
	Length uint16		// 包体长度
	Check uint16		// 校验和
	SeqId uint16		// 包序列号
	Cmd byte			// 命令
	Ack byte			// 回应标志
	Type uint16			// 设备类型
	DevId DeviceId		// 设备编号
	PkBody interface{}	// 包体
	Ended uint8			// 结束标志
}

// 王力zigbee锁消息体
type WlZigbeeMsg struct {
	Started uint8		// 开始标志
	Version uint8		// 协议版本号
	Length uint8		// 包体长度
	Check uint16		// 校验和
	SeqId uint16		// 包序列号
	Cmd uint8			// 命令
	Ack uint8			// 回应标志
	Type uint8			// 设备类型
	Uuid string			// 设备编号 - 加解密使用
	PkBody interface{}	// 包体
	Ended uint8			// 结束标志
}

//1. 获取用户列表版本号(0x30)(服务器-->前板)
type GetDevUserVerReq struct {
	Time int32	// 时间戳
}

type GetDevUserVerResp struct {
	DevUserVer uint32 // 用户列表版本号(4)
}

//2. 请求同步用户列表(0x31)(服务器-->前板)
type SyncDevUser struct {
	Num uint16	// 请求数量
	Time int32	// 时间戳
}
type DevUserInfo struct {
	UserType uint8		// 用户类型(1)，用户类型:  0 - 管理员，1 - 普通用户，2 - 临时用户
	UserNo uint16		// 设备用户编号
	OpenBitMap int32	// 验证方式位图(4)
	PermitNum uint16 	// 允许开门次数
	Remainder uint16	// 剩余开门次数
	StartDate [3]byte	// 开始日期
	EndDate [3]byte		// 结束日期
	TimeSlot1 [4]byte	// 时段1
	TimeSlot2 [4]byte	// 时段2
	TimeSlot3 [4]byte	// 时段3
}
// 用户列表版本号(4)+用户数量(2)+(用户类型(1)+用户编号(2)+验证方式位图(4) +总次数(2)+剩余次数（2）+开始日期(3)+截止日期(3)+时段1(4)+时段2(4)+时段3(4))*N
type SyncDevUserResp struct {
	DevUserVer uint32 			// 用户列表版本号(4)
	DevUserNum uint16			// 用户数量(2)
	DevUserInfos []DevUserInfo	// 用户列表
}

//3. 删除用户(0x32)(服务器-->前板)
type DelDevUser struct {
	UserNo uint16	// 设备用户编号
	MainOpen uint8	// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen uint8	// 是否胁迫，是否胁迫：0-正常，1-胁迫
	Time int32		// 时间戳
}

//4. 新增用户(0x33)(服务器-->前板)
// 用户编号(2)+开锁方式(1)+是否胁迫(1)+用户类型(1)+密码(10)+时间(4)+用户随机数(4)+临时用户时效(20)
type AddDevUser struct {
	UserNo uint16		// 设备用户编号，指定操作的用户编号，如果是0XFFFF表示新添加一个用户
	MainOpen uint8		// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫
	UserType uint8		// 用户类型(1)，用户类型:  0 - 管理员，1 - 普通用户，2 - 临时用户
	Passwd [6]byte		// 密码(6)，密码开锁方式，目前是6个字节.如果添加的是其他验证方式,则为0xff.密码位数少于10位时,多余的填0xff
	UserNote int32		// 时间戳作为随机数
	PermitNum uint16 	// 允许开门次数
	StartDate [3]byte	// 开始日期
	EndDate [3]byte		// 结束日期
	TimeSlot1 [4]byte	// 时段1
	TimeSlot2 [4]byte	// 时段2
	TimeSlot3 [4]byte	// 时段3
}

//5. 新增用户报告步骤(0x34)(前板-->服务器)
// 用户列表版本号(4)+用户编号(2)+开锁方式(1)+是否胁迫(1)+步骤序号(1)+步骤状态(1)+时间(4)
type AddDevUserStep struct {
	DevUserVer uint32 // 用户列表版本号(4)
	UserNo uint16		// 设备用户编号
	MainOpen uint8		// 主开锁方式，开锁方式：附表开锁方式
	SubOpen uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫
	StepNo uint8		// 步骤序号(1)
	StepState uint8		// 步骤状态(1)
	Time int32			// 时间戳
}

//6. 用户更新上报(0x35)(前板-->服务器)
// 用户列表版本号(4)+操作类型(1)+ 用户编号(2)+用户类型(1)+用户随机数(4)+ 验证方式位图(4)+总次数(2)+剩余次数（2）+开始日期(3)+截止日期(3)+时段1(4)+时段2(4)+时段3(4)+时间（4）
type UserUpdateLoad struct {
	DevUserVer uint32 	// 用户列表版本号(4)，当前锁保存的用户列表版本号
	OperType uint8		// 操作类型(1)，0-新增用户，1-更新用户，2-删除用户, 3-删除所有普通用户, 4-删除所有临时用户
	UserNo uint16		// 设备用户编号
	UserType uint8		// 用户类型(1)，用户类型:  0 - 管理员，1 - 普通用户，2 - 临时用户
	Time int32			// 时间戳作为随机数
	OpenBitMap int32	// 验证方式位图(4)

	// 临时用户有效
	PermitNum uint16 	// 允许开门次数
	Remainder uint16	// 剩余开门次数

	// 时间段均为BCD码格式。
	StartDate [3]byte	// 开始日期
	EndDate [3]byte		// 结束日期
	TimeSlot1 [4]byte	// 时段1
	TimeSlot2 [4]byte	// 时段2
	TimeSlot3 [4]byte	// 时段3
}

//7. 实时视频(0x36)(服务器-->前板)
type RealVideo struct {
	Act uint8	// 视频打开/关闭：1打开视频，0关闭视频
}

//8. Wifi设置(0x37)(服务器-->前板)
// Ssid（32）+密码（16）
type WiFiSet struct {
	Ssid [32]byte		// Ssid（32）
	Passwd [16]byte	// 密码（16）
}

//9. 门铃呼叫(0x38)(前板-->服务器)
type DoorbellCall struct {
	Time int32	// 时间戳
}

//10. 人体感应报警(0x39)(前板-->服务器)
type Alarms struct {
	Time int32	// 时间戳
}

//15. 低压报警(0x2A)(前板--->服务器)
type LowBattAlarm struct {
	Battery uint8 	// 电量百分比
	Time int32		// 时间戳
}

//16. 图片上传(0x2F)(前板--->服务器)
// 消息类型(1)+消息id(4)+图片路径长度（1）+图片路径(n)
type PicUpload struct {
	CmdType byte	// 命令类型
	MsgId int32		// 消息id
	PicLen uint8	// 图片路径长度（1）
	PicPath string	// 图片路径(N)
}

//17. 用户开锁消息上报(0x40)(前板--->服务器)
/*
用户列表版本号(4)+ 用户数量（1）+时间(4)+ 电量百分比（1）+ 单/双人模式（1）
+用户编号(2)+验证方式(1)+是否胁迫(1)+剩余次数（2）
+用户编号(2)+验证方式(1)+是否胁迫(1)+剩余次数（2）
*/
type OpenLockMsg struct {
	DevUserVer uint32 	// 用户列表版本号(4)，当前锁保存的用户列表版本号
	UserNum uint8		// 用户数量(1)，当前锁保存的用户总数量
	Time int32			// 时间戳
	Battery uint8		// 电量百分比：0-100十进制数
	SinMul uint8		// 单/双人模式:1-单人（只有用户1），2-双人

	// 用户1
	UserNo uint16		// 设备用户编号
	MainOpen uint8		// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫
	Remainder uint16	// 剩余开门次数

	// 用户2
	UserNo2 uint16		// 设备用户编号
	MainOpen2 uint8		// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen2 uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫
	Remainder2 uint16	// 剩余开门次数
}

//18. 用户进入菜单上报(0x42)(前板--->服务器)
/*
用户列表版本号(4)+ 时间(4)+ 电量百分比（1）+ 单/双人模式（1）
+用户编号(2)+验证方式(1)+是否胁迫(1)
+用户编号(2)+验证方式(1)+是否胁迫(1)
*/
type EnterMenuMsg struct {
	DevUserVer uint32 	// 用户列表版本号(4)，当前锁保存的用户列表版本号
	Time int32			// 时间戳
	Battery uint8		// 电量百分比：0-100十进制数
	SinMul uint8		// 单/双人模式:1-单人（只有用户1），2-双人

	// 用户1
	UserNo uint16		// 设备用户编号
	MainOpen uint8		// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫

	// 用户2
	UserNo2 uint16		// 设备用户编号
	MainOpen2 uint8		// 主开锁方式，开锁方式：附表开锁方式，如果该字段是0，表示删除该用户
	SubOpen2 uint8		// 是否胁迫，是否胁迫：0-正常，1-胁迫
}

//19. 在线离线(0x46)(后板-->服务器)
type OnOffLine struct {
	OnOff uint8		// 标志:0:离线;1在线
	// Time int32		// 时间戳 TODO:JHHE 去掉时间
}

//20. 远程开锁命令(0x52)(服务器->前板)
// 密码1（6）+密码2（6）+随机数（4）+md5（16）
type RemoteOpenLock struct {
	Passwd [6]byte	// 密码1（6）
	Passwd2 [6]byte	// 密码2（6）
	Time int32		// 随机数（4）
}

// 用户id1（2）+用户id2（2）+时间（4）
type RemoteOpenLockResp struct {
	UserNo uint16	// 用户id1（2）
	UserNo2 uint16	// 用户id2（2）
	Time int32		// 随机数（4）
}

//21. 获取参数(0x71)(服务器-->前板，后板)
// 参数编号(1)+时间(4)
type GetLockParamReq struct {
	ParamNo uint8	// 参数编号(1)
	Time int32		// 时间(4)
}

type GetLockParamResp struct {
	ParamNo uint8		// 参数编号(1)
	ParamValue uint8	// 参数值(1)
	ParamValue2 uint8	// 参数值2(1)
	Time int32			// 时间(4)
}

//22. 设置参数(0x72)(服务器-->前板，后板)
type SetLockParamReq struct {
	ParamNo uint8		// 参数编号(1)
	ParamValue uint8	// 参数值(1)
	ParamValue2 uint8	// 参数值2(1)
	Time int32			// 时间(4)
}

//23. 参数更新(0x73)(前板,后板-->服务器)
type ParamUpdate struct {
	ParamNo uint8			// 参数编号(1)
	ParamValue interface{}	// 参数值(1)
	ParamValue2 uint8		// 参数值2(1)
}

type ParamUpdateResp struct {
	ParamNo uint8		// 参数编号(1)
	ParamValue uint8	// 参数值(1)
	ParamValue2 uint8	// 参数值2(1)
	Time int32			// 时间(4)
}

//24. 软件重启命令(0x74)(服务器-->前、后板)
type RebootLock struct {
	Time int32			// 时间(4)
}

//25. 恢复出厂化(0xEA)( 服务器-->前、后板)
type RestLock struct {
}

//26. 设置临时用户时段(0x76)(服务器-->前板)
// 临时用户编号(2)+次数(2)+开始日期(3)+截止日期(3)+时段1(4)+时段2(4)+时段3(4)
type SetTmpDevUser struct {
	UserNo uint16		// 设备用户编号
	PermitNum uint16 	// 允许开门次数
	StartDate [3]byte	// 开始日期
	EndDate [3]byte		// 结束日期
	TimeSlot1 [4]byte	// 时段1
	TimeSlot2 [4]byte	// 时段2
	TimeSlot3 [4]byte	// 时段3
}

//27. 发送设备信息(0x70)(前板，后板-->服务器)
// 前板信息长度(1)+前板信息+后板信息长度(1)+后板信息
/*
前板信息：主版本号(1)+次版本号(1)+修订版本号(1)+型号(2)+保留(2)+用户列表版本号(4)+音量(1)+验证模式(1)+是否带屏(1)+ 密码开关(1)+电量(1)+门未关(1)+假锁(1)+ 人体感应报警(1)+报警类型(1)+模组sn(16)+ssid(32)
版本号：V1.1.120，表示主版本号：0x01，次版本号：0x01，修订版本号：0x78；
型号：门锁设备型号2字节信息；
用户列表版本号: 4个字节, 初始为0
音量：1字节；0静音，1小，2中，3大。
验证模式：1单人，2双人。
是否带屏：0无屏，1带屏。
密码开关：0表示密码禁用，1表示密码使能
电量:电池电量1~100
门未关报警开关：0关闭，1开启
假锁报警开关：0关闭，1开启
人体感应报警开关: 0关闭，1开启
报警类型： 1拍照+录像，2拍照
视频模组sn码：16字节
Ssid:模组连接的路由器的ssid 32字节

后板信息：主版本号(1)+次版本号(1)+修订号(1)
版本号：V1.1.120，表示主版本号：0x01，次版本号：0x01，修订号：0x78；
串码：12-18字节IMEI码
常开模式：0常开关闭，1常开启用
*/
type UploadDevInfo struct {
	FLen uint8				// 前板信息长度
	FMainVer uint8			// 版本号：0x01
	FSubVer uint8			// 次版本号：0x01
	FModVer uint16			// 修订版本号：0x78
	FType uint16			// 门锁设备型号2字节信息
	DevUserVer uint32 		// 用户列表版本号(4)，当前锁保存的用户列表版本号
	Volume uint8			// 1字节；0静音，1小，2中，3大。
	SinMul uint8			// 验证模式：1单人，2双人
	IsHasScr uint8			// 是否带屏：0无屏，1带屏
	PwdSwitch uint8			// 密码开关：0表示密码禁用，1表示密码使能
	Battery uint8			// 电量:电池电量1~100
	NolockSwitch uint8		// 门未关报警开关：0关闭，1开启
	FakelockSwitch uint8	// 假锁报警开关：0关闭，1开启
	InfraSwitch uint8		// 人体感应报警开关: 0关闭，1开启
	InfraTime uint8			// 人体感应拍照时间: 1字节（单位秒）
	AlarmSwitch uint8		// 报警类型： 1拍照+录像，2拍照
	BellSwitch uint8		// 门铃开关 0：关闭，1：开启
	ActiveMode uint8		// 0门锁唤醒后立即激活，1输入激活码激活
	IpcSn [16]byte			// 视频模组sn码：16字节
	Ssid [32]byte			// Ssid:模组连接的路由器的ssid 32字节
	Capability uint32		// 能力集：无符号4字节

	BLen uint8				// 后板信息长度
	BMainVer uint8			// 后板主版本号：0x01
	BSubVer uint8			// 后板次版本号：0x01
	BModVer uint16			// 后板修订号：0x78
	OpenMode uint8			// 常开模式：0常开关闭，1常开启用
	RemoteSwitch uint8		// 远程开关（0：无法使用远程开锁，1：可以使用远程开锁）
	ProductId [12]byte		// 12字节字符串，例：Z12345670001
}

type UploadZigbeeDevInfo struct {
	FMainVer uint8			// 版本号：0x01
	FSubVer uint8			// 次版本号：0x01
	FModVer uint16			// 修订版本号：0x78
	FType uint16			// 门锁设备型号2字节信息
	DevUserVer uint32 		// 用户列表版本号(4)，当前锁保存的用户列表版本号
	Volume uint8			// 1字节；0静音，1小，2中，3大。
	SinMul uint8			// 验证模式：1单人，2双人
	IsHasScr uint8			// 是否带屏：0无屏，1带屏
	PwdSwitch uint8			// 密码开关：0表示密码禁用，1表示密码使能
	Battery uint8			// 电量:电池电量1~100
	NolockSwitch uint8		// 门未关报警开关：0关闭，1开启
	FakelockSwitch uint8	// 假锁报警开关：0关闭，1开启
	InfraSwitch uint8		// 人体感应报警开关: 0关闭，1开启
	InfraTime uint8			// 人体感应拍照时间: 1字节（单位秒）
	AlarmSwitch uint8		// 报警类型： 1拍照+录像，2拍照
	BellSwitch uint8		// 门铃开关 0：关闭，1：开启
	ActiveMode uint8		// 0门锁唤醒后立即激活，1输入激活码激活
	Capability uint32		// 能力集：无符号4字节

	BMainVer uint8			// 后板主版本号：0x01
	BSubVer uint8			// 后板次版本号：0x01
	BModVer uint16			// 后板修订号：0x78
	OpenMode uint8			// 常开模式：0常开关闭，1常开启用
	RemoteSwitch uint8		// 远程开关（0：无法使用远程开锁，1：可以使用远程开锁）
}

type UploadDevInfoResp struct {
	Time int32
}

//28. 锁状态上报(0x55)(后板->服务器)
type DoorStateUpload struct {
	State uint8
}
