package entity

/*数据包格式
* +——----——+——-----——+——----——+——----——+——----——+
* | 版本号  |  模块号 |    长度 |  校验和 |   数据 |
* +——----——+——-----——+——----——+——----——+——----——+
 */
// 2 + 2 + 2 + 2
type MyHeader struct {
	ApiVersion  uint16
	ServiceType uint16
	MsgLen      uint16
	CheckSum    uint16
}

type OneMsg struct {
	At         int64  `json:"at"`
	Msgtype    int    `json:"type"` // 数据点消息(type=1)，设备上下线消息(type=2)
	Value      string `json:"value"`
	Imei       string `json:"imei"`
	Dev_id     int    `json:"dev_id"`
	Ds_id      string `json:"ds_id"`
	Status     int    `json:"status"` // 设备上下线标识：0-下线, 1-上线
	Login_type int    `json:"login_type"`
}

type OneNETData struct {
	Msg_signature string `json:"msg_signature"`
	Nonce         string `json:"nonce"`
	Msg           OneMsg `json:"msg"`
}

type Header struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`
}

type DeviceActive struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Time int64 `json:"time"`
}

type SetDeviceTime struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	ParaNo  int   `json:"paraNo"`
	PaValue int64 `json:"paValue"`
	Time    int64 `json:"time"`
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

type RespOneNET struct {
	RespErrno int    `json:"errno"`
	RespError string `json:"error"`
}

type AddDevUserStep struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UserVer   int `json:"userVer"`   // 设备用户版本号
	UserId    int `json:"userId"`    // 设备用户ID
	MainOpen  int `json:"mainOpen"`  // 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen   int `json:"subOpen"`   // 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Step      int `json:"step"`      // 步骤序号（指纹：需要4步，1，2，3，4分别代表上下左右；刷卡：需要1步；密码：需要2步，分别是第一次输入密码和第二次输入密码）
	StepState int `json:"stepState"` // 0表示成功，1表示失败
	Time      int `json:"time"`
}

// 电信平台设备数据变化，推送的JSON
type TelecomDeviceServiceData struct {
	ServiceId   string `json:"serviceId"`
	ServiceType string `json:"serviceType"`
	Data        string `json:"data"`
	EventTime   string `json:"eventTime"`
}

type TelecomDeviceDataChanged struct {
	NotifyType string `json:"notifyType"`
	RequestId  string `json:"requestId"`
	DeviceId   string `json:"deviceId"`
	GatewayId  string `json:"gatewayId"`
	Service    TelecomDeviceServiceData
}

//亿速码芯片激活
type YisumaActiveSE struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	UId        string `json:"uId"`        // 安全芯片编号
	ProjectNo  string `json:"projectNo"`  // 项目编号
	MerchantNo string `json:"merchantNo"` // 商户号
	Random     string `json:"random"`     // 随机数
}

//亿速码加签数据
type YisumaSign struct {
	UId           string `json:"uId"`
	ProjectNo     string `json:"projectNo"`
	MerchantNo    string `json:"merchantNo"`
	CardChanllege string `json:"cardChanllege"`
}

//亿速码请求apdu
type YisumaHttpsReq struct {
	Body      YisumaSign `json:"body"`
	Signature string     `json:"signature"`
}

//亿速码返回apdu
type YisumaHttpsRes struct {
	ResultCode string `json:"resultCode"`
	ResultMsg  string `json:"resultMsg"`
	Apdu       string `json:"apdu"`
}

//亿速码激活指令下发
type YisumaActiveApdu struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Apdu string `json:"apdu"`
}

//亿速码加签指令与锁端交互
type YisumaRandomSign struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Password  string `json:"passwd"`
	Password2 string `json:"passwd2"`
	Random    string `json:"random"`
	Signature string `json:"signature"`
}

//锁端上报亿速码随机数
type YisumaStateRandom struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Random string `json:"random"`
}

//feibee推送的消息体
type FeibeeData struct {
	Code    int                `josn:"code"`
	Status  string             `json:"status"`
	Ver     string             `json:"ver"`
	Msg     []FeibeeDevMsg     `json:"msg"`
	Gateway []FeibeeGatewayMsg `json:"gateway"`
	Records []FeibeeRecordsMsg `json:"records"`
}

//feibee推送的设备信息
type FeibeeDevMsg struct {
	Name       string `json:"name"`
	Bindid     string `json:"bindid"`
	Uuid       string `json:"uuid"`
	Devicetype string `json:"devicetype"`
	Deviceuid  int    `json:"deviceuid"`
	Snid       string `json:"snid"`
	Profileid  int    `json:"profileed"`
	Deviceid   int    `json:"deviceid"`
	Onoff      int    `json:"onoff"`
	Online     int    `json:"online"`
	Zonetype   int    `json:"zonetype"`
	Battery    int    `json:"battery"`
	Lastvalue  int    `json:"lastvalue"`
	IEEE       string `json:"IEEE"`
	Pushstring string `json:"pushstring"`
}

//feibee推送的网关信息
type FeibeeGatewayMsg struct {
	Bindid   string `json:"bindid"`
	Bindstr  string `json:"bindstr"`
	Version  string `json:"version"`
	Snid     string `json:"snid"`
	Devnum   int    `json:"devnum"`
	Areanum  int    `json:"areanum"`
	Timernum int    `json:"timernum"`
	Scenenum int    `json:"scenenum"`
	Tasknum  int    `json:"tasknum"`
	Uptime   int    `json:"uptime"`
	Online   int    `json:"online"`
}

//feibee推送的操作设备状态
type FeibeeRecordsMsg struct {
	Bindid     string
	Deviceuid  int
	Uuid       string
	Devicetype string
	Zonetype   int
	Deviceid   int
	Cid        int
	Aid        int
	Value      string
	Orgdata    string
	Uptime     int
	Pushstring string
}

//设备入网通知(推送给app)
type Feibee2AppMsg struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	Devid   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Note      string `json:"note"` //设备别名
	Deviceuid int    `json:"deviceuid"`
	Online    int    `json:"online"`
	Battery   int    `json:"battery"`
	OpType    string `json:"opType"`
}
