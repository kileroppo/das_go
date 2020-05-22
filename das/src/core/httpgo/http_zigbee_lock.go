package httpgo

import (
	"das/core/entity"
	"das/core/log"
	"strconv"
)

func Http2FeibeeZigbeeLock(appData, bindid, bindstr, uuid, uid string) {
    var err error
	var reqMsg entity.ZigbeeLockMsg2Feibee
	var conf = log.Conf

	reqMsg.Act = "setcommand"// "standardWriteAttribute"
	reqMsg.Code = "295" // "286"
	reqMsg.Bindid = bindid
	reqMsg.Ver = "2.0"
	reqMsg.Bindstr,err = WonlyAESDecrypt(bindstr, "W" + uuid + "only")
	if err != nil {
		log.Errorf("Http2FeibeeZigbeeLock.WonlyAESDecrypt > %s", err)
		return
	}
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

	reqMsg.Uuid = uuid
	reqMsg.Uid, _ = strconv.Atoi(uid)
	reqMsg.Command = appData //转为16进制字符串

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
		log.Warning("Control ZigbeeLock failed", string(respData))
	} else {
		log.Info("Control ZigbeeLock successfully", string(respData))
	}
}