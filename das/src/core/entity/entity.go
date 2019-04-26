package entity

/*数据包格式
* +——----——+——-----——+——----——+——----——+——----——+
* | 版本号  |  模块号 |    长度 |  校验和 |   数据 |
* +——----——+——-----——+——----——+——----——+——----——+
*/
// 2 + 2 + 2 + 2
type MyHeader struct {
	ApiVersion uint16
	ServiceType uint16
	MsgLen uint16
	CheckSum uint16
}

type OneMsg struct {
	At int64			`json:"at"`
	Msgtype int			`json:"type"`				// 数据点消息(type=1)，设备上下线消息(type=2)
	Value string		`json:"value"`
	Imei string			`json:"imei"`
	Dev_id int			`json:"dev_id"`
	Ds_id string		`json:"ds_id"`
	Status int			`json:"status"`				// 设备上下线标识：0-下线, 1-上线
	Login_type int		`json:"login_type"`
}

type OneNETData struct {
	Msg_signature string	`json:"msg_signature"`
	Nonce string			`json:"nonce"`
	Msg OneMsg 				`json:"msg"`
}

type Header struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`
}

type DeviceActive struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`

	Time int64			`json:"time"`
}

type SetDeviceTime struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`

	ParaNo int			`json:"paraNo"`
	PaValue int64		`json:"paValue"`
	Time int64			`json:"time"`
}

type UpgradeQuery struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`

	Part int			`json:"part"`
}

type UpgradeReq struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`

	Part int			`json:"part"`
	Offset int64		`json:"offset"`
	FileName string		`json:"fileName"`
}

type RespOneNET struct {
	RespErrno int		`json:"errno"`
	RespError string	`json:"error"`
}


type AddDevUserStep struct {
	Cmd int				`json:"cmd"`
	Ack int      		`json:"ack"`
	DevType string 		`json:"devType"`
	DevId string 		`json:"devId"`
	Vendor string		`json:"vendor"`
	SeqId int			`json:"seqId"`

	UserVer string		`json:"userVer"` 		// 设备用户版本号
	UserId int 			`json:"userId"`			// 设备用户ID
	MainOpen int		`json:"mainOpen"` 		// 主开锁方式（1-密码，2-刷卡，3-指纹）
	SubOpen int 		`json:"subOpen"` 		// 次开锁方式 (0-正常指纹，1-胁迫指纹, 0:正常密码，1:胁迫密码，2:时间段密码，3:远程密码）
	Step int 			`json:"step"` 			// 步骤序号（指纹：需要4步，1，2，3，4分别代表上下左右；刷卡：需要1步；密码：需要2步，分别是第一次输入密码和第二次输入密码）
	StepState int 		`json:"stepState"` 		// 0表示成功，1表示失败
	Time int			`json:"time"`
}

