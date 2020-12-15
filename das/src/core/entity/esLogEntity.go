package entity

type GrayLog struct {
	Version  string `json:"version"`
	Host     string `json:"host"`
	Facility string `json:"facility"`
	Message  string `json:"short_message"`
	Timestamp int64 `json:"timestamp"`
	LogType  string `json:"logType"`
}

type EsLogEntiy struct {
	Vendor string `json:"vendor"`
	DeviceName string `json:"deviceName"`
	DeviceId string `json:"deviceId"`
	FamilyId string `json:"familyId"`
	User string `json:"user"`
	Operation string `json:"operation"`
	ThirdPlatform string `json:"thirdPlatform"`
	RetMsg string `json:"retMsg"`
	RawData string `json:"rawData"`
}
