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
			log.Errorf("Http2FeibeeWonlyLGuard > %s", err)
			return
		}
	}()

	var msg entity.WonlyGuardMsgFromApp
	if err := json.Unmarshal([]byte(appData), &msg); err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > json.Unmarshal > %s", err)
		return
	}
	log.Infof("Send WonlyGuard '%s' control to feibee", msg.DevId)
	var err error
	var reqMsg entity.Req2Feibee
	var conf = log.Conf

	reqMsg.Act = "standardWriteAttribute"
	reqMsg.Code = "286"
	reqMsg.Bindid = msg.Bindid
	reqMsg.AccessId,err = conf.GetString("feibee2http", "accessid")
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > get accessId > %s", err)
		return
	}

	reqMsg.Key, err = conf.GetString("feibee2http", "key")
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > get key > %s", err)
		return
	}

	key := "W" + msg.DevId + "only"

	reqMsg.Bindstr, err = WonlyAESDecrypt(msg.Bindstr, key)
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > WonlyAESDecrypt > %s", err)
		return
	}
	reqMsg.Ver = "2.0"
	reqMsg.Devs = append(reqMsg.Devs, entity.ReqDevInfo2Feibee{
		Uuid:  msg.DevId,
		Value: msg.Value,
	})

	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > json.Marshal > %s", err)
		return
	}

	log.Debugf("Send to Feibee: %s", reqData)
	respData, err := DoFeibeeControlReq(reqData)
	if err != nil {
		log.Warningf("Http2FeibeeWonlyLGuard > %s", err)
		return
	}

	//log.Infof("Http2FeibeeWonlyLGuard > resp: %s", respData)
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

func WonlyAESDecrypt(cipher, key string) (res string ,err error) {
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