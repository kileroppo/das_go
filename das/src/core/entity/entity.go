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
	Name       string `json:"name,omitempty"`
	Bindid     string `json:"bindid,omitempty"`
	Uuid       string `json:"uuid,omitempty"`
	Devicetype string `json:"devicetype,omitempty"`
	Deviceuid  int    `json:"deviceuid,omitempty"`
	Snid       string `json:"snid,omitempty"`
	Profileid  int    `json:"profileed,omitempty"`
	Deviceid   int    `json:"deviceid,omitempty"`
	Onoff      int    `json:"onoff,omitempty"`
	Online     int    `json:"online,omitempty"`
	Zonetype   int    `json:"zonetype,omitempty"`
	Battery    int    `json:"battery,omitempty"`
	Lastvalue  int    `json:"lastvalue,omitempty"`
	IEEE       string `json:"IEEE,omitempty"`
	Pushstring string `json:"pushstring,omitempty"`
	DevDegree  int    `json:"brightness,omitempty"`
}

//feibee推送的网关信息
type FeibeeGatewayMsg struct {
	Bindid   string `json:"bindid,omitempty"`
	Bindstr  string `json:"bindstr,omitempty"`
	Version  string `json:"version,omitempty"`
	Snid     string `json:"snid,omitempty"`
	Devnum   int    `json:"devnum,omitempty"`
	Areanum  int    `json:"areanum,omitempty"`
	Timernum int    `json:"timernum,omitempty"`
	Scenenum int    `json:"scenenum,omitempty"`
	Tasknum  int    `json:"tasknum,omitempty"`
	Uptime   int    `json:"uptime,omitempty"`
	Online   int    `json:"online,omitempty"`
}

//feibee推送的操作设备状态
type FeibeeRecordsMsg struct {
	Bindid     string `json:"bindid,omitempty"`
	Deviceuid  int    `json:"deviceuid,omitempty"`
	Uuid       string `json:"uuid,omitempty"`
	Devicetype string `json:"devicetype,omitempty"`
	Zonetype   int    `json:"zonetype,omitempty"`
	Deviceid   int    `json:"deviceid,omitempty"`
	Cid        int    `json:"cid,omitempty"`
	Aid        int    `json:"aid,omitempty"`
	Value      string `json:"value,omitempty"`
	Orgdata    string `json:"orgdata,omitempty"`
	Uptime     int    `json:"uptime,omitempty"`
	Pushstring string `json:"pushstring,omitempty"`
}

//feibee设备消息通知(推送给app)
type Feibee2AppMsg struct {
	Cmd     int    `json:"cmd"`
	Ack     int    `json:"ack"`
	DevType string `json:"devType"`
	Devid   string `json:"devId"`
	Vendor  string `json:"vendor"`
	SeqId   int    `json:"seqId"`

	Note      string `json:"note,omitempty"` //设备别名
	Deviceuid int    `json:"deviceuid,omitempty"`
	Online    int    `json:"online,omitempty"`
	Battery   int    `json:"battery,omitempty"`
	OpType    string `json:"opType,omitempty"`
	OpValue   string `json:"opValue,omitempty"`
	Time      int    `json:"time,omitempty"`
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

	AlarmType  string `json:"alarmType,omitempty"`
	AlarmValue string `json:"alarmValue,omitempty"`

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
