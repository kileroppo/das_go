package entity

type Header struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`
}

//1. 添加设备用户(APP->后台-->锁)
type MyDTM struct {
	Start int32 `json:"start"`
	End   int32 `json:"end"`
}

type MyDate struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type AddDevUser struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string 		 `json:"appUser"`	  // 远程用户，APP账号
	UserId   uint16      `json:"userId"`      // 设备用户ID
	UserNote string      `json:"userNote"`    // 设备用户别名
	UserType uint8       `json:"userType"`    // 用户类型（0-管理员，1-普通用户，2-临时用户）
	MainOpen uint8       `json:"mainOpen"`    // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen  uint8       `json:"subOpen"`     // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Passwd   string      `json:"passwd"`      // 如果是添加密码需要填写
	Total    uint16      `json:"total"`       // 开门次数，0xffff为无限次
	Count    uint16      `json:"count"`       // 开门次数，0xffff为无限次
	MyDate   MyDTM       `json:"date"`        // 开始有效时间
	MyTime   [3]MyDTM    `json:"time"`        // 时段
	TimeLen  interface{} `json:"time_length"` // 兼容捷博生产商，临时用户时长（单位：秒）
	Bindid   string      `json:"bindid,omitempty"`      // zigbee锁网关账号
	Bindstr  string      `json:"bindstr,omitempty"`     // zigbee锁网关密码
}

type _DelDevUser struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string 		 `json:"appUser"`	  // 远程用户，APP账号
	UserId   uint16      `json:"userId"`      // 设备用户ID
	UserNote string      `json:"userNote"`    // 设备用户别名
	UserType uint8       `json:"userType"`    // 用户类型（0-管理员，1-普通用户，2-临时用户）
	MainOpen uint8       `json:"mainOpen"`    // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen  uint8       `json:"subOpen"`     // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Passwd   string      `json:"passwd"`      // 如果是添加密码需要填写
	Total    uint16      `json:"total"`       // 开门次数，0xffff为无限次
	MyDate   MyDTM       `json:"date"`        // 开始有效时间
	MyTime   [3]MyDTM    `json:"time"`        // 时段
	TimeLen  interface{} `json:"time_length"` // 兼容捷博生产商，临时用户时长（单位：秒）
	Bindid   string      `json:"bindid,omitempty"`      // zigbee锁网关账号
	Bindstr  string      `json:"bindstr,omitempty"`     // zigbee锁网关密码
}

//2. 设置临时用户
type SetTmpDevUser struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserId uint16   `json:"userId"` // 设备用户ID
	Total  uint16   `json:"total"`  // 开门次数，0xffff为无限次
	Count  uint16   `json:"count"`  // 开门次数，0xffff为无限次
	MyDate MyDTM    `json:"date"`   // 开始有效时间
	MyTime [3]MyDTM `json:"time"`   // 时段
	TimeLen interface{} `json:"time_length"` // 捷博锁临时用户，时长（单位：秒）
}

type UserOperUpload struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserType int    `json:"userType"` 	// 用户类型
	UserId int  	`json:"userId"` 	// 用户1
	UserId2 int  	`json:"userId2"` 	// 用户2，单人模式用户2为0xffff
	OpType int 		`json:"opType"` 	// 操作，0-新增用户，1-修改用户，2-删除用户，3-删除普通组，4-删除临时组，5-设置参数
	OpUserPara int	`json:"opUserPara"` // 被操作用户/参数
	OpValue int 	`json:"opValue"` 	// 内容(1)，当操作不为5时有效，0-用户整体，1-密码，2-卡，3-指纹

	Time int64 		`json:"time"`		// 时间戳，整型（单位：秒）
}

//3. 新增用户步骤（锁-->后台-->APP，锁主动上报，指纹，卡）
type AddDevUserStep struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer   uint32 `json:"userVer"`   // 设备用户版本号
	UserId    uint16 `json:"userId"`    // 设备用户ID
	MainOpen  int    `json:"mainOpen"`  // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen   int    `json:"subOpen"`   // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Step      int    `json:"step"`      // 步骤序号（指纹：需要4步，1，2，3，4分别代表上下左右；刷卡：需要1步；密码：需要2步，分别是第一次输入密码和第二次输入密码）
	StepState int    `json:"stepState"` // 0表示成功，1表示失败
	Time      int32  `json:"time"`
}

//4. 删除设备用户（APP-->后台-->锁）
type DelDevUser struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string 	`json:"appUser"`  // 远程用户，APP账号
	UserId   uint16 `json:"userId"`   // 设备用户ID
	MainOpen uint8  `json:"mainOpen"` // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen  uint8  `json:"subOpen"`  // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Time     int32  `json:"time"`
	Bindid   string `json:"bindid,omitempty"`  // zigbee锁网关账号
	Bindstr  string `json:"bindstr,omitempty"` // zigbee锁网关密码
}

//5. 用户更新上报
type DevUserUpload struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	OpType    int      `json:"opType"`
	UserVer   uint32   `json:"userVer"`   // 设备用户版本号
	UserId    uint16   `json:"userId"`    // 设备用户ID
	UserNote  string   `json:"userNote"`  // 用户别名，设备用户别名，存到DB要做判断，如果不为空则更新别名
	UserType  int      `json:"userType"`  // 用户类型（0-管理员，1-普通用户，2-临时用户）
	Finger    int      `json:"finger"`    // 指纹数量
	Ffinger   int      `json:"ffinger"`   // 胁迫指纹数量
	Passwd    int      `json:"passwd"`    // 密码数量
	Card      int      `json:"card"`      // 卡数量
	Face      int      `json:"face"`      // 人脸数量（可视人脸锁带该字段）
	Bluetooth int      `json:"bluetooth"` // 蓝牙数量（蓝牙开锁方式）
	Total     int      `json:"total"`     // 开门次数，0为无限次
	Remainder int      `json:"remainder"` // 剩下的开门次数
	MyDate    MyDTM    `json:"date"`      // 开始有效时间
	MyTime    [3]MyDTM `json:"time"`      // 时段
}

//6. 同步设备用户列表
type SyncDevUserReq struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Num uint16 `json:"num"` // 每次请求10个

	Bindid  string `json:"bindid"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr"` // zigbee锁网关密码
}
type DevUser struct {
	UserId    uint16   `json:"userId"`    // 设备用户ID
	UserType  int      `json:"userType"`  // 用户类型（0-管理员，1-普通用户，2-临时用户）
	Finger    int      `json:"finger"`    // 指纹数量
	Ffinger   int      `json:"ffinger"`   // 胁迫指纹数量
	Passwd    int      `json:"passwd"`    // 密码数量
	Card      int      `json:"card"`      // 卡数量
	Face      int      `json:"face"`      // 人脸数量（可视人脸锁带该字段）
	Bluetooth int      `json:"bluetooth"` // 蓝牙数量（蓝牙开锁方式）
	Total     int      `json:"total"`     // 开门次数，0为无限次
	Remainder int      `json:"remainder"` // 剩下的开门次数
	MyDate    MyDTM    `json:"date"`      // 开始有效时间
	MyTime    [3]MyDTM `json:"time"`      // 时段
}
type SyncDevUserResp struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer  uint32    `json:"userVer"`   // 设备用户版本号
	Num      int       `json:"num"`       // 返回锁体内的10个设备用户
	UserList []DevUser `json:"user_list"` // 设备用户数组
}
type SyncDevUserRespEx struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer  uint32   `json:"userVer"`   // 设备用户版本号
	Num      int      `json:"num"`       // 返回锁体内的10个设备用户
	UserList []string `json:"user_list"` // 设备用户数组
}

//7. 远程开锁
// 单人
type SRemoteOpenLockReq struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string 	`json:"appUser"`	// 远程用户，APP账号
	Passwd string `json:"passwd"`
	Time   interface{}  `json:"time"`

	Bindid  string `json:"bindid,omitempty"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr,omitempty"` // zigbee锁网关密码
}

// 双人
type MRemoteOpenLockReq struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string `json:"appUser"`	// 远程用户，APP账号
	Passwd  string `json:"passwd"`
	Passwd2 string `json:"passwd2"`
	Time    interface{}  `json:"time"`

	Bindid  string `json:"bindid,omitempty"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr,omitempty"` // zigbee锁网关密码
}

// 设置锁用户参数
type SetDevUserParam struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserId  uint16 `json:"userId"`
	ParamType  uint8 `json:"paramType"`
	ParamValue  uint8 `json:"paramValue"`
}

type RemoteOpenLockResp struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserId  uint16 `json:"userId"`
	UserId2 uint16 `json:"userId2"`
	Time    interface{}  `json:"time"`
}

//8. 上传设备信息
type UploadDevInfo struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer        uint32 `json:"userVer"`         // 设备用户版本号，如果是0则不需要发起同步请求
	FVer           string `json:"fVer"`            // 前板版本号
	FType          string `json:"fType"`           // 前板型号（Z201)
	HasScr         uint8  `json:"hasScr"`          // 是否带屏幕（1-带屏幕，0-不带屏幕）
	Battery        uint8  `json:"battery"`         // 电池电量
	VolumeLevel    uint8  `json:"volume_level"`    // 音量等级(带屏幕的锁，可以设置为静音，1-3音量等级，3表示音量最大)
	PasswdSwitch   uint8  `json:"passwd_switch"`   // 密码开关（0：无法使用密码开锁，1：可以使用密码开锁）
	SinMul         uint8  `json:"sin_mul"`         // 开门模式（1：表示单人模式, 2：表示双人模式）
	BVer           string `json:"bVer"`            // 后板版本号
	NbVer          string `json:"nbVer"`           // NB版本号
	Sim            string `json:"sim"`             // SIM卡号
	OpenMode       uint8  `json:"open_mode"`       // 常开模式
	RemoteSwitch   uint8  `json:"remote_switch"`   // 远程开关（0：无法使用远程开锁，1：可以使用远程开锁）
	ActiveMode     uint8  `json:"active_mode"`     // 远程开锁激活方式，0：门锁唤醒后立即激活，1：输入激活码激活
	NolockSwitch   uint8  `json:"nolock_switch"`   // 门未关开关，0：关闭，1：开启
	FakelockSwitch uint8  `json:"fakelock_switch"` // 假锁开关，0：关闭，1：开启
	InfraSwitch    uint8  `json:"infra_switch"`    // 人体感应报警开关，0：关闭，1：唤醒，但不推送消息，2：唤醒并且推送消息
	InfraTime      uint8  `json:"infra_time"`      // 人体感应报警，红外持续监测到多少秒 就上报消息
	AlarmSwitch    uint8  `json:"alarm_switch"`    // 报警类型开关，0：关闭，1：拍照+录像，2：拍照
	WifiSsid       string `json:"wifi_ssid"`       // wifi的ssid
	BellSwitch     uint8  `json:"bell_switch"`     // 门铃开关 0：关闭，1：开启
	FBreakSwitch   uint8  `json:"fbreak_switch"`    // 防拆报警开关：0关闭，1开启
	ProductID      string `json:"productID"`       // 产品序列号
	Capability     uint32 `json:"capability"`      // 能力集

	// 说明：NB锁包含两个版本：1、基础NB版本，2、视频（IPC）的版本，含以下字段
	IpcSn string `json:"ipc_sn"` // 视频设备（IPC）序列号

	// 亿速码安全芯片相关参数
	UId        string `json:"uId"`        // 安全芯片id
	ProjectNo  string `json:"projectNo"`  // 项目编号
	MerChantNo string `json:"merChantNo"` // 商户号
	Random     string `json:"random"`     // 安全芯片随机数

	// 兼容字段，某些功能不支持的NB锁
	Unsupport int `json:"unsupport"` // 0-所有功能支持，1-临时用户时段不支持
}

//9. 设置参数
//10. 参数更新
type SetDeviceTime struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	ParaNo  int    `json:"paraNo"`
	PaValue int64  `json:"paValue"`
	Time    string `json:"time"`
}
type SetLockParamReq struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	AppUser string 	`json:"appUser"`  // 远程用户，APP账号
	ParaNo   uint8  `json:"paraNo"`   // 参数编号
	PaValue  uint8  `json:"paValue"`  // 参数值1
	PaValue2 uint8  `json:"paValue2"` // 参数值2，当参数编号为0x0b（人体感应报警开关）且”参数值1”为2时候，此字段有效
	Time     interface{} `json:"time"` // 时间戳

	Bindid  string `json:"bindid,omitempty"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr,omitempty"` // zigbee锁网关密码
}
type LockParam struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	ParaNo   uint8       `json:"paraNo"`   // 参数编号
	PaValue  interface{} `json:"paValue"`  // 参数值1
	PaValue2 uint8       `json:"paValue2"` // 参数值2，当参数编号为0x0b（人体感应报警开关）且”参数值1”为2时候，此字段有效
}

//11. 主动上报门锁开门消息
type OpenLockLog struct {
	UserId    uint16 `json:"userId"`    // 设备用户ID
	MainOpen  uint8  `json:"mainOpen"`  // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen   uint8  `json:"subOpen"`   // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	SinMul    uint8  `json:"sin_mul"`   // 开门模式（1：表示单人模式, 2：表示双人模式）
	Remainder uint16 `json:"remainder"` // 临时用户剩下能开门的次数,其他用户为0xffff
	Time      int32  `json:"time"`
}
type UploadOpenLockLog struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer uint32        `json:"userVer"`  // 设备用户版本号
	UserNum uint8         `json:"userNum"`  // 设备用户总数
	Battery int           `json:"battery"`  // 电池电量
	LogList []OpenLockLog `json:"log_list"` // 开锁日志
}

//12. 进入菜单消息
type EnterMenu struct {
	UserId   uint16 `json:"userId"`   // 设备用户ID
	MainOpen uint8  `json:"mainOpen"` // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen  uint8  `json:"subOpen"`  // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	SinMul   uint8  `json:"sin_mul"`  // 开门模式（1：表示单人模式, 2：表示双人模式）
	Time     int32  `json:"time"`
}
type UploadEnterMenuLog struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer uint32      `json:"userVer"`  // 设备用户版本号
	Battery int         `json:"battery"`  // 电池电量
	LogList []EnterMenu `json:"log_list"` // 日志
}

//13. 报警消息
type AlarmMsg struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Time int32 `json:"time"`
}

// 低电压告警
type AlarmMsgBatt struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Value int   `json:"value"` // 电量百分比 低压报警带有电池电压告警
	Time  int32 `json:"time"`
}

//14. 锁激活状态上报
type DeviceActive struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Signal int   `json:"signal"` // NB锁信号强度
	Time   int64 `json:"time"`
	Timestamp int64 `json:"timestamp"`
}
type DeviceActiveResp struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Time int64 `json:"time"`
}

//15. 实时视频（APP->锁）
type RealVideoLock struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Act uint8 `json:"act"` // 视频打开/关闭：1打开视频，0关闭视频

	Bindid  string `json:"bindid"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr"` // zigbee锁网关密码
}

//16. Wifi设置（APP->锁，DB更新设备信息表）
type SetLockWiFi struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	WifiSsid string `json:"wifi_ssid"` // WIFI的ssid
	WifiPwd  string `json:"wifi_pwd"`  // WIFI密码

	Bindid  string `json:"bindid"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr"` // zigbee锁网关密码
}

//17. 门铃呼叫（锁—后台—>APP）
type DoorBellCall struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Time int32 `json:"time"`

	Bindid  string `json:"bindid"`  // zigbee锁网关账号
	Bindstr string `json:"bindstr"` // zigbee锁网关密码
}

//18. 视频锁图片上报（锁—后台—>APP）
type PicUpload struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	CmdType int    `json:"cmdType"` // 哪个命令产生的图片：开门，所有的报警等
	TimeId  int    `json:"time_id"` // 开锁消息时间ID，根据开锁消息的time来检索更新图片。
	PicName string `json:"picName"` // 抓拍图片名称，视频锁抓拍图片
}

//19. 锁状态上报(0x55)(后板->服务器)
type DoorStateUpload struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	State uint8 `json:"state"`
}
type UpgradeQuery struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Part int `json:"part"`
}

type UpgradeReq struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Part     int    `json:"part"`
	Offset   int64  `json:"offset"`
	FileName string `json:"fileName"`
}

type ZigbeeLockHead struct {
	Header

	Bindid  string `json:"bindid"`
	Bindstr string `json:"bindstr"`
}

type RangeHoodAlarm struct {
	Header

	Time int `json:"time"`
}
