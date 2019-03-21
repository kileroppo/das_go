package entity

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

	Offset int64		`json:"offset"`
	FileName string		`json:"fileName"`
}

type RespOneNET struct {
	RespErrno int		`json:"errno"`
	RespError string	`json:"error"`
}