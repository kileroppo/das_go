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
	User_oper_upload  = 0x77 // 用户操作上报
	Add_dev_user_step = 0x34 // 新增用户步骤
	Del_dev_user      = 0x32 // 删除设备用户
	Update_dev_user   = 0x35 // 用户更新上报
	Sync_dev_user     = 0x31 // 同步设备用户列表
	Set_dev_user_para = 0x3B // 设置设备用户参数
	Remote_open       = 0x52 // 远程开锁
	Upload_dev_info   = 0x70 // 上传设备信息

	Set_dev_para    = 0x72 // 设置参数
	Update_dev_para = 0x73 // 设备参数更新上报
	Soft_reset      = 0x74 // 软件复位
	// Factory_reset = 0x75		// 恢复出厂设置
	Factory_reset    = 0xEA // 恢复出厂设置
	Upload_open_log  = 0x40 // 门锁开门日志上报
	UpEnter_menu_log = 0x42 // 用户进入菜单上报

	// 报警
	Noatmpt_alarm      = 0x20 // 非法操作报警
	Forced_break_alarm = 0x22 // 强拆报警
	Fakelock_alarm     = 0x24 // 假锁报警
	Nolock_alarm       = 0x26 // 门未关报警
	Gas_Alarm          = 0x27 // 燃气报警
	Low_battery_alarm  = 0x2A // 锁体的电池，低电量报警
	Infrared_alarm     = 0x39 // 人体感应报警（infra红外感应)
	Lock_PIC_Upload    = 0x2F // 视频锁图片上报

	// 锁激活
	Upload_lock_active = 0x46 // 锁激活状态上报

	// 视频设备
	Real_Video      = 0x36 // 实时视频
	Set_Wifi        = 0x37 // Wifi设置
	Notify_Set_Wifi = 0x3A // 锁端Wifi设置成功通知
	Door_Call       = 0x38 // 门铃呼叫

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

	Wonly_LGuard_Msg = 0xfb

	PadDoor_RealVideo     = 0x1001 // 平板锁实时视频
	PadDoor_Weather       = 0x1002 // 平板锁当前天气
	Set_AIPad_Reboot_Time = 0x1003 // 设置中控网关定时参数
	RangeHood_Control     = 0x1005 // 油烟机档位控制
	RangeHood_Ctrl_Query  = 0x1006 // 油烟机档位查询
	RangeHood_Lock_Query  = 0x1007 // 油烟机绑定门锁查询
	Body_Fat_Scale        = 0x1008 // 体脂称数据上报

	PadDoor_Num_Upload = 0x1100 // 平板锁人流检测上报
	PadDoor_Num_Reset  = 0x1101 // 平板门锁人流检测重置
	Other_Vendor_Msg   = 0x1200

	Feibee_Ori_Msg     = 0xfa   //飞比原始消息
	Device_Normal_Msg  = 0xfb   //设备状态消息
	Device_Sensor_Msg  = 0xfc   //设备传感器消息
	Scene_Trigger      = 0xf1   //爱岗场景触发（中控平板的闹钟作为触发条件）
	Feibee_Gtw_Info    = 0xf5   //飞比网关信息

)

var WlLockAlarmMsg = map[int]string{
	Noatmpt_alarm : "非法操作报警",
	Forced_break_alarm : "强拆报警",
	Fakelock_alarm : "假锁报警",
	Nolock_alarm : "门未关报警",
	Gas_Alarm : "燃气报警",
	Low_battery_alarm : "锁体的电池，低电量报警",
	Infrared_alarm : "人体感应报警（infra红外感应)",
}
const (
	ONENET_PLATFORM     = "onenet"
	TELECOM_PLATFORM    = "telecom"
	ANDLINK_PLATFORM    = "andlink"
	PAD_DEVICE_PLATFORM = "paddevice"
	ALIIOT_PLATFORM     = "aliIoT"
	FEIBEE_PLATFORM     = "feibee"
	MQTT_PLATFORM       = "mqtt"
	MQTT_PAD_PLATFORM   = "mqttpad"

	Device_Type_ZKGtw = "WonlyZKgateway"
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

// 协议选择
const (
	GENERAL_PROTOCOL = 1
	ZIGBEE_PROTOCOL  = 2
)

//	设备参数设置编号，需要特殊处理的字符串
const (
	IPC_SN_PNO     = 0x0d // 视频模组sn
	WIFI_SSID_PNO  = 0x0f // WIFI_SSID
	PROJECT_No_PNO = 0x10 // 产品序列号
	IPC_SN_PNO_lencens = 0x1b // 揽胜视频模组sn

)

// 开锁方式 1-密码，2-刷卡，3-指纹，5-人脸，12-蓝牙
const (
	OPEN_PWD    = 1  // 1-密码
	OPEN_CARD   = 2  // 2-刷卡
	OPEN_FINGER = 3  // 3-指纹
	OPEN_FACE   = 5  // 5-人脸
	OPEN_BLE    = 12 // 12-蓝牙
)

const (
	VENDOR_WONLY  = "general"
	VENDOR_FEIBEE = "feibee"
	Vendor_Lancens = "lancens"
)

//王力内部传感器报警字符类型枚举
const (
	Wonly_Status_Sensor_Illuminance  = "illuminance"
	Wonly_Status_Sensor_Humidity     = "humidity"
	Wonly_Status_Sensor_Temperature  = "temperature"
	Wonly_Status_Sensor_PM25         = "PM2.5"
	Wonly_Status_Sensor_VOC          = "VOC"
	Wonly_Status_Sensor_Formaldehyde = "formaldehyde"
	Wonly_Status_Sensor_CO2          = "CO2"
	Wonly_Status_Sensor_Gas          = "gas"
	Wonly_Status_Sensor_Smoke        = "smoke"
	Wonly_Status_Sensor_Infrared     = "infrared"
	Wonly_Status_Sensor_Doorcontact  = "doorContact"
	Wonly_Status_Sensor_Flood        = "flood"
	Wonly_Status_Sensor_SOSButton    = "sosButton"
	Wonly_Status_Pad_People          = "peopleDetection"
	Wonly_Status_Sensor_Forced_Break  = "forcedBreak"

	Wonly_Status_Low_Voltage   = "lowVoltage"
	Wonly_Status_Low_Power     = "lowPower"
	Wonly_Status_FbLock_Status = "lockStatus"
	Wonly_Status_FbDev_Status  = "devStatus"
	Wonly_Status_Forced_Break  = "forcedBreak"

	Wonly_Status_Airer_Illumination      = "illumination"
	Wonly_Status_Airer_Disinfection      = "disinfection"
	Wonly_Status_Airer_Disinfection_Time = "disinfectionTime"
	Wonly_Status_Airer_MotorOperation    = "motorOperation"
	Wonly_Status_Airer_Drying            = "drying"
	Wonly_Status_Airer_Air_Drying        = "airDrying"
	Wonly_Status_Airer_Drying_Time       = "dryingTime"
	Wonly_Status_Airer_Air_Drying_Time   = "airDryingTime"

	Wonly_Status_Aircondition_Mode              = "mode"
	Wonly_Status_Aircondition_State             = "state"
	Wonly_Status_Aircondition_Windspeed         = "windspeed"
	Wonly_Status_Aircondition_Curr_Temperature  = "currentTemperature"
	Wonly_Status_Aircondition_Local_Temperature = "localTemperature"
	Wonly_Status_Aircondition_Max_Temperature   = "maxTemperature"
	Wonly_Status_Aircondition_Min_Temperature   = "minTemperature"

	Wonly_Status_Sleep_Stage = "sleepStatus"
)

var (
	Wonly_Sensor_Vals_Gas          = []string{"检测正常", "燃气浓度已超标，正在报警"}
	Wonly_Sensor_Vals_Smoke        = []string{"检测正常", "烟雾浓度已超标，正在报警"}
	Wonly_Sensor_Vals_Flood        = []string{"检测正常", "水浸位已超标，正在报警"}
	Wonly_Sensor_Vals_Infrared     = []string{"无人经过", "有人经过"}
	Wonly_Sensor_Vals_Doorcontact  = []string{"门磁已关闭", "门磁已打开"}
	Wonly_Sensor_Vals_SOSButton    = []string{"检测正常", "发生紧急呼叫"}
	Wonly_Sensor_Vals_Forced_Break = []string{"传感器未被强拆", "传感器被强拆"}

	Wonly_FbAirer_Vals_Illumination    = []string{"关闭", "开启"}
	Wonly_FbAirer_Vals_Disinfection    = []string{"关闭", "开启"}
	Wonly_FbAirer_Vals_Motor_Operation = []string{"正常", "上限位", "下限位"}
	Wonly_FbAirer_Vals_Drying          = []string{"关闭", "开启"}
	Wonly_FbAirer_Vals_Air_Drying      = []string{"关闭", "开启"}
	Wonly_FbAirer_Vals_Mode            = []string{"关闭", "", "", "制冷", "制热", "打开"}
	Wonly_FbAirer_Vals_Windspeed       = []string{"关闭", "低速", "中速", "高速", "", "自动"}

	Wonly_FbDev_Vals_Status  = []string{"关闭", "开启"}
	Wonly_FbLock_Vals_Status = []string{"锁定", "解锁"}
)

var (
	SensorVal2Str = map[string]([]string){
		Wonly_Status_Sensor_Gas:         Wonly_Sensor_Vals_Gas,
		Wonly_Status_Sensor_Smoke:       Wonly_Sensor_Vals_Smoke,
		Wonly_Status_Sensor_Flood:       Wonly_Sensor_Vals_Flood,
		Wonly_Status_Sensor_Infrared:    Wonly_Sensor_Vals_Infrared,
		Wonly_Status_Sensor_Doorcontact: Wonly_Sensor_Vals_Doorcontact,
		Wonly_Status_Sensor_Forced_Break: Wonly_Sensor_Vals_Forced_Break,
		Wonly_Status_Sensor_SOSButton:    Wonly_Sensor_Vals_SOSButton,
	}
)
