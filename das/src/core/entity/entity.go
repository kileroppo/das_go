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

type RespOneNET struct {
	RespErrno int    `json:"errno"`
	RespError string `json:"error"`
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

//发送给PMS消息体
type Feibee2PMS struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	DevId   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	FeibeeData `json:"fb_data"`
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
	Bindid     string `json:"bindid"`
	Deviceuid  int    `json:"deviceuid"`
	Uuid       string `json:"uuid"`
	Devicetype string `json:"devicetype"`
	Zonetype   int    `json:"zonetype"`
	Deviceid   int    `json:"deviceid"`
	Cid        int    `json:"cid"`
	Aid        int    `json:"aid"`
	Value      string `json:"value"`
	Orgdata    string `json:"orgdata"`
	Uptime     int    `json:"uptime"`
	Pushstring string `json:"pushstring"`
}

//feibee设备消息通知(推送给app)
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
	OpValue   string `json:"opValue"`
	Time      int    `json:"time"`
}

//feibee设备消息通知(推送给db)
type Feibee2DBMsg struct {
	Feibee2AppMsg
	Bindid string `json:"bindid"`
}

//feibee设备报警消息通知(推送给app)
type FeibeeAlarm2AppMsg struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	Devid   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`
	Time    int    `json:"time"`

	AlarmType  string `json:"alarmType"`
	AlarmValue string `json:"alarmValue"`

	AlarmFlag int    `json:"alarmFlag,omitempty"`
	Bindid    string `json:"bindid"`
}

//feibee传感器报警消息 作为自动场景触发消息(推送给pms)
type FeibeeAutoScene2pmsMsg struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	Devid   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`
	Time    int    `json:"time"`

	TriggerType int    `json:"triggerT"`
	AlarmValue  string `json:"alarmValue"`
	AlarmType   string `json:"alarmType"`
	SceneId     string `json:"sceneId"`
	Zone        string `json:"zone"`
}

type AliRawData struct {
	RawData []byte
	Topic   string
}
