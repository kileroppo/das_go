package httpgo

import (
	"encoding/hex"
	"strings"

	"github.com/json-iterator/go"

	"das/core/entity"
	"das/core/log"
	"das/core/util"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Http2FeibeeWonlyLGuard(appData string) {
	defer func() {
		if err:=recover();err != nil {
			log.Error("Http2FeibeeWonlyLGuard() error = ", err)
			return
		}
	}()

	var msg entity.WonlyGuardMsgFromApp
	if err := json.Unmarshal([]byte(appData), &msg); err != nil {
		log.Warning("Http2FeibeeWonlyLGuard json.Unmarshal() error = ", err)
		return
	}
	log.Infof("Send WonlyGuard '%s' control to feibee", msg.Devid)
	var err error
	var reqMsg entity.Req2Feibee
	var conf = log.Conf

	reqMsg.Act = "standardWriteAttribute"
	reqMsg.Code = "286"
	reqMsg.Bindid = msg.Bindid
	reqMsg.AccessId,err = conf.GetString("feibee2http", "accessid")
	if err != nil {
		log.Warning("Http2FeibeeWonlyLGuard get accessId error = ", err)
		return
	}

	reqMsg.Key, err = conf.GetString("feibee2http", "key")
	if err != nil {
		log.Warning("Http2FeibeeWonlyLGuard get key error = ", err)
		return
	}

	key := "W" + msg.Devid + "only"

	reqMsg.Bindstr, err = WonlyGuardAESDecrypt(msg.Bindstr, key)
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard WonlyGuardAESDecrypt() error = ", err)
		return
	}
	reqMsg.Ver = "2.0"
	reqMsg.Devs = append(reqMsg.Devs, entity.ReqDevInfo2Feibee{
		Uuid:  msg.Devid,
		Value: msg.Value,
	})

	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		log.Warning("Http2FeibeeWonlyLGuard json.Marshal() error = ", err)
		return
	}

	log.Debug("Send to Feibee: ", string(reqData))
	respData, err := DoHTTP("POST", "https://dev.fbeecloud.com/devcontrol/", reqData)
	if err != nil {
		log.Warning("Http2FeibeeWonlyLGuard DoHTTP() error = ", err)
		return
	}

	var respMsg entity.RespFromFeibee
	err = json.Unmarshal(respData, &respMsg)
	if err != nil {
		log.Warning("Control WonlyLGuard failed")
		return
	}

	if respMsg.Code != 1 {
		log.Warning("Control WonlyLGuard failed")
	} else {
		log.Info("Control WonlyLGuard successfully")
	}
}

func WonlyGuardAESDecrypt(cipher, key string) (res string ,err error) {
	cipherByte,err := hex.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	keyByte := []byte(strings.ToUpper(util.Md5(key)))

	data,err := util.ECBDecryptByte(cipherByte, keyByte)
	if err != nil {
		return "", err
	}

	return string(data), nil
}