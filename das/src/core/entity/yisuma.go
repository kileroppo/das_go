package entity

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

