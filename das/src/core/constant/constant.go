package constant

const Debug = true

//session key name
const GohSessionName = "goh_session_account" //session键值
const GohUserRedisDb = 1                     //redis用户库

//response code
const Success = "0000"             //成功
const SystemForbidden = "403"      //禁止访问
const SystemError = "500"          //系统错误
const LoginPasswordError = "10001" //账号或密码错误
const LoginAccountLock = "10002"   //账号锁定

//lock
const LockTimes = 5                 //输入错误次数锁定账号
const LockExpireTime = 24 * 60 * 60 //锁定时间
const LockAccountPrefix = "lock_"   //锁定账号的redis键值前缀

//sms code
const SmsCodeExpireTime = 5 * 60 //短信验证码超时时间，单位:s
const SmsCodePrefix = "sms_"     //短信验证码的redis键值前缀

//http request
const HttpTimeOut = 30 //http请求超时时间

const Base_Dianxin_Url = "180.101.147.89:8743"

const (
	API_VERSION          = 108  // 协议版本号
	SERVICE_TYPE         = 0x06 // 服务类型（DAS (0x06)）
	SERVICE_TYPE_UNENCRY = 0x16 // 服务类型，设备升级（DAS_UPGRADE (0x16)）

	Add_dev_user      = 0x33 // 添加设备用户
	Set_dev_user_temp = 0x76 // 设置临时用户
	Add_dev_user_step = 0x34 // 新增用户步骤
	Del_dev_user      = 0x32 // 删除设备用户
	Update_dev_user   = 0x35 // 用户更新上报
	Sync_dev_user     = 0x31 // 同步设备用户列表
	Remote_open       = 0x52 // 远程开锁
	Upload_dev_info   = 0x70 // 上传设备信息

	Set_dev_para    = 0x72 // 设置参数
	Update_dev_para = 0x73 // 设备参数更新上报
	Soft_reset      = 0x74 // 软件复位
	// Factory_reset = 0x75		// 恢复出厂设置
	Factory_reset   = 0xEA // 恢复出厂设置
	Upload_open_log = 0x40 // 门锁开门日志上报

	// 报警
	Noatmpt_alarm      = 0x20 // 非法操作报警
	Forced_break_alarm = 0x22 // 强拆报警
	Fakelock_alarm     = 0x24 // 假锁报警
	Nolock_alarm       = 0x26 // 门未关报警
	Low_battery_alarm  = 0x2A // 锁体的电池，低电量报警
	Infrared_alarm     = 0x39 // 人体感应报警（infra红外感应)
	Lock_PIC_Upload    = 0x2F // 视频锁图片上报

	// 锁激活
	Upload_lock_active = 0x46 // 锁激活状态上报

	// 视频设备
	Real_Video = 0x36 // 实时视频
	Set_Wifi   = 0x37 // Wifi设置
	Door_Call  = 0x38 // 门铃呼叫

	// 锁状态
	Door_State = 0x55 // 锁状态上报

	//NB锁升级
	Notify_F_Upgrade = 0xE0 // 通知前板升级（APP—后台—>锁）
	Notify_B_Upgrade = 0xE1 // 通知后板升级（APP—后台—>锁）

	Get_Upgrade_FileInfo   = 0xC0 // 锁查询升级固件包信息
	Download_Upgrade_File  = 0xC1 // 锁下载固件升级包（锁—>后台，分包传输）
	Upload_F_Upgrade_State = 0xC2 // 前板上传升级状态
	Upload_B_Upgrade_State = 0xD0 // 后板上传升级状态

	Device_Resp_TimeOut = 0x99 // 设备响应超时

	FILE_MAX = 256 // 升级每次发送的文件大小
	//亿速码
	Active_Yisuma_SE    = 0x68 //激活亿速码安全芯片
	Random_Yisuma_State = 0x66 //亿速码随机数上报
)

const (
	ONENET_PLATFORM 	= "onenet"
	TELECOM_PLATFORM 	= "telecom"
	ANDLINK_PLATFORM	= "andlink"
	WIFI_PLATFORM		= "wifi"
)

// 电信平台订阅消息类型
const (
	/*
	 * service Notify Type
	 * serviceInfoChanged|deviceInfoChanged|LocationChanged|deviceDataChanged|deviceDatasChanged
	 * deviceAdded|deviceDeleted|messageConfirm|commandRsp|deviceEvent|ruleEvent|deviceModelAdded
	 * deviceModelDeleted|deviceDesiredPropertiesModifyStatusChanged
	 */
	DEVICE_ADDED          = "deviceAdded"
	DEVICE_INFO_CHANGED   = "deviceInfoChanged"
	DEVICE_DATA_CHANGED   = "deviceDataChanged"
	DEVICE_DELETED        = "deviceDeleted"
	MESSAGE_CONFIRM       = "messageConfirm"
	SERVICE_INFO_CHANGED  = "serviceInfoChanged"
	COMMAND_RSP           = "commandRsp"
	DEVICE_EVENT          = "deviceEvent"
	RULE_EVENT            = "ruleEvent"
	DEVICE_DATAS_CHANGED  = "deviceDatasChanged"
	DEVICE_DESIRED_MODIFY = "deviceDesiredPropertiesModifyStatusChanged"

	/*
	 * management Notify Type
	 * swUpgradeStateChangeNotify|swUpgradeResultNotify|fwUpgradeStateChangeNotify|fwUpgradeResultNotify
	 */
	SW_UPGRADE_STATE_CHANGED = "swUpgradeStateChangeNotify"
	SW_UPGRADE_RESULT        = "swUpgradeResultNotify"
	FW_UPGRADE_STATE_CHANGED = "fwUpgradeStateChangeNotify"
	FW_UPGRADE_RESULT        = "fwUpgradeResultNotify"
)
