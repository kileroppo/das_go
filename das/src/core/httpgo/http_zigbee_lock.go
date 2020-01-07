package httpgo

import (
	"das/core/entity"
	"das/core/log"
)

func Http2FeibeeZigbeeLock(appData []byte, bindid, bindstr string) {
    var err error
	var reqMsg entity.ZigbeeLockMsg2Feibee
	var conf = log.Conf

	reqMsg.Act = "standardWriteAttribute"
	reqMsg.Code = "286"
	reqMsg.Bindid = bindid
	reqMsg.Bindstr = bindstr
	reqMsg.AccessId,err = conf.GetString("feibee2http", "accessid")
	if err != nil {
		log.Warning("Http2FeibeeZigbeeLock get accessId error = ", err)
		return
	}

	reqMsg.Key, err = conf.GetString("feibee2http", "key")
	if err != nil {
		log.Warning("Http2FeibeeZigbeeLock get key error = ", err)
		return
	}

	reqMsg.Command = string(appData)

	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		log.Warning("Http2FeibeeZigbeeLock json.Marshal() error = ", err)
		return
	}

	log.Debug("Send to Feibee: ", string(reqData))
	respData, err := DoHTTP("POST", "https://dev.fbeecloud.com/devcontrol/", reqData)
	if err != nil {
		log.Warning("Http2FeibeeZigbeeLock DoHTTP() error = ", err)
		return
	}

	var respMsg entity.RespFromFeibee
	err = json.Unmarshal(respData, &respMsg)
	if err != nil {
		log.Warning("Control ZigbeeLock failed")
		return
	}

	if respMsg.Code != 1 {
		log.Warning("Control ZigbeeLock failed")
	} else {
		log.Info("Control ZigbeeLock successfully")
	}
}