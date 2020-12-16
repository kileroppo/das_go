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

type LogHeader struct {
	MsgType       int32  `json:"msgType" bson:"msgType"`
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
	RawData       string `json:"rawData,omitempty" bson:"rawData,omitempty"`
	TimeStr       string `json:"timeStr,omitempty" bson:"timeStr,omitempty"`
	MsgName       string `json:"msgName,omitempty" bson:"msgName,omitempty"`
	NanoTimestamp int64  `json:"nanoTimestamp" bson:"nanoTimestamp"`
}

type SysLogMsg struct {
	LogHeader `json:",inline" bson:",inline"`

	UUid       string `json:"uuid,omitempty" bson:"uuid,omitempty"`
	DevType    string `json:"devType,omitempty" bson:"devType,omitempty"`
	URL        string `json:"url" bson:"url,omitempty"`
	UserName   string `json:"username" bson:"username,omitempty"`
	GtwId      string `json:"gatewayId,omitempty" bson:"gatewayId,omitempty"`
	Result     string `json:"result" bson:"result,omitempty"`
	FamilyId   string `json:"familyId,omitempty" bson:"familyId,omitempty"`
	SceneName  string `json:"sceneName,omitempty" bson:"sceneName,omitempty"`
	VendorName string `json:"vendorName,omitempty" bson:"vendorName,omitempty"`
	DevName    string `json:"devName,omitempty" bson:"devName,omitempty"`
}

type OtherVendorDevMsg struct {
	Header

	OriData string `json:"oriData,omitempty"`
}
