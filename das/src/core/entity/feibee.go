package entity

//发送给PMS消息体
type Feibee2PMS struct {
	Header

	FeibeeData `json:"fb_data"`
}

//feibee推送的消息体
type FeibeeData struct {
	Code          int                `josn:"code"`
	Status        string             `json:"status"`
	Ver           string             `json:"ver"`
	Msg           []FeibeeDevMsg     `json:"msg,omitempty"`
	Gateway       []FeibeeGatewayMsg `json:"gateway,omitempty"`
	Records       []FeibeeRecordsMsg `json:"records,omitempty"`
	SceneMessages []FeibeeSceneMsg   `json:"sceneMessages,omitempty"`
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
	Deviceid   int    `json:"deviceid"`
	Onoff      int    `json:"onoff"`
	Online     int    `json:"online"`
	Zonetype   int    `json:"zonetype"`
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
	Cid        int    `json:"cid"`
	Aid        int    `json:"aid"`
	Value      string `json:"value,omitempty"`
	Orgdata    string `json:"orgdata,omitempty"`
	Uptime     int    `json:"uptime,omitempty"`
	Pushstring string `json:"pushstring,omitempty"`
	Snid       string `json:"snid,omitempty"`
}

//feibee设备消息通知(推送给app)
type Feibee2DevMsg struct {
	Header

	Note      string `json:"note,omitempty"` //设备别名
	Deviceuid int    `json:"deviceuid,omitempty"`
	Online    int    `json:"online"`
	Battery   int    `json:"battery,omitempty"`
	OpType    string `json:"opType,omitempty"`
	OpValue   string `json:"opValue,omitempty"`
	Time      int    `json:"time,omitempty"`
	Bindid    string `json:"bindid,omitempty"`
	Snid      string `json:"snid,omitempty"`

	SceneMessages []FeibeeSceneMsg `json:"sceneMessages,omitempty"`
}

//feibee设备报警消息通知(推送给app)
type Feibee2AlarmMsg struct {
	Header

	Time int `json:"time"`

	AlarmType  string `json:"alarmType,omitempty"`
	AlarmValue string `json:"alarmValue,omitempty"`

	AlarmFlag int    `json:"alarmFlag,omitempty"`
	Bindid    string `json:"bindid,omitempty"`
}

//feibee传感器报警消息 作为自动场景触发消息(推送给pms)
type Feibee2AutoSceneMsg struct {
	Header

	Time int `json:"time"`

	TriggerType int    `json:"triggerT"`
	AlarmFlag   int    `json:"alarmFlag"`
	AlarmType   string `json:"alarmType"`
	AlarmValue  string `json:"alarmValue"`
	SceneId     string `json:"sceneId"`
	Zone        string `json:"zone"`
}

type AliRawData struct {
	RawData []byte
	Topic   string
}

//app透传至das的WonlyGuard消息
type WonlyGuardMsgFromApp struct {
	Header

	Bindid  string `json:"bindid"`
	Bindstr string `json:"bindstr"`
	Value   string `json:"value"`
}

type ReqFeibeeHead struct {
	Act      string `json:"act"`
	Code     string `json:"code"`
	AccessId string `json:"AccessID"`
	Key      string `json:"key"`
	Bindid   string `json:"bindid"`
	Bindstr  string `json:"bindstr"`
	Ver      string `json:"ver"`
}

type Req2Feibee struct {
	ReqFeibeeHead

	Devs []ReqDevInfo2Feibee `json:"devs"`
}

type ZigbeeLockMsg2Feibee struct {
	ReqFeibeeHead

	Uuid    string `json:"uuid"`
	Uid     int    `json:"deviceuid"`
	Command string `json:"command"`
}

type ReqDevInfo2Feibee struct {
	Uuid  string `json:"uuid"`
	Value string `json:"value"`
}

type RespFromFeibee struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Ver    string `json:"ver"`
}

type InfraredTreasureControlResult struct {
	Uuid           string `json:"uuid"`
	DevType        string `json:"deviceType"`
	FirmVer        string `json:"firmwareVer"`
	ControlDevType int64  `json:"controlDevType"`
	FunctionKey    int64  `json:"functionKey"`
}

type FeibeeSceneMsg struct {
	Bindid string            `json:"bindid,omitempty"`
	Scenes []FeibeeSceneInfo `json:"scenes,omitempty"`
}

type FeibeeSceneInfo struct {
	SceneName    string              `json:"sceneName,omitempty"`
	SceneID      int                 `json:"sceneID"`
	SceneMembers []FeibeeSceneMember `json:"sceneMembers,omitempty"`
}

type FeibeeSceneMember struct {
	Deviceuid       int    `json:"deviceuid"`
	DeviceID        int    `json:"deviceID"`
	Data1           int    `json:"data1"`
	Data2           int    `json:"data2"`
	Data3           int    `json:"data3"`
	Data4           int    `json:"data4"`
	IRID            int    `json:"IRID"`
	Delaytime       int    `json:"delaytime"`
	SceneFunctionID int    `json:"sceneFunctionID"`
	Uuid            string `json:"uuid,omitempty"`
}

type YKInfraredStatus struct {
	Devid     string `json:"mac"`
	Online    int    `json:"state"`
	Timestamp int    `json:"timestamp"`
}

type FeibeeLockAlarmMsg struct {
	Header
	Timestamp int `json:"time"`
}

type FeibeeLockBattMsg struct {
	Header
	Value     int `json:"value"`
	Timestamp int `json:"time"`
}

type FeibeeLockRemoteOn struct {
	Header
	UserId    int `json:"userId"`
	UserId2   int `json:"userId2"`
	Timestamp int `json:"time"`
}

type FeibeeLockOpen struct {
	Uuid         string `json:"uuid"`
	VendorName   string `json:"vendor_name"`
	Timestamp    int    `json:"timestamp"`
	DeviceUserId int    `json:"device_user_id"`
	UnlockMode   string `json:"unlock_mode"`
	AuthMode     string `json:"auth_mode"`
	StressStatus string `json:"stress_status"`
	OpType       string `json:"op_type"`
	MsgType      string `json:"unlock_message_type"`
}
