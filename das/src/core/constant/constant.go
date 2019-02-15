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

const (
	Add_dev_user = 0x33			// 添加设备用户
	Set_dev_user_temp = 0x76	// 设置临时用户
	Add_dev_user_step = 0x34	// 新增用户步骤
	Del_dev_user = 0x32			// 删除设备用户
	Update_dev_user = 0x35		// 用户更新上报
	Sync_dev_user = 0x31		// 同步设备用户列表
	Remote_open = 0x52			// 远程开锁
	Upload_dev_info = 0x70		// 上传设备信息

	Set_dev_para = 0x72			// 设置参数
	Update_dev_para = 0x73		// 设备参数更新上报
	Soft_reset = 0x74			// 软件复位
	// Factory_reset = 0x75		// 恢复出厂设置
	Factory_reset = 0xEA		// 恢复出厂设置
	Upload_open_log = 0x40		// 门锁开门日志上报

	// 报警
	Noatmpt_alarm = 0x20		// 非法操作报警
	Forced_break_alarm = 0x22	// 强拆报警
	Fakelock_alarm = 0x24		// 假锁报警
	Nolock_alarm = 0x26			// 门未关报警
	Low_battery_alarm = 0x2A	// 锁体的电池，低电量报警

	// 锁激活
	Upload_lock_active = 0x46	// 锁激活状态上报

	// 视频设备
	Real_Video	= 0x36			// 实时视频
	Set_Wifi	= 0x37			// Wifi设置
	Door_Call 	= 0x38			// 门铃呼叫
)