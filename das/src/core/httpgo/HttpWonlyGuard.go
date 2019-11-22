package httpgo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"encoding/hex"

	"github.com/json-iterator/go"

	"../entity"
	"../log"
	"../util"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Http2FeibeeWonlyGuard(appData string) {
	defer func() {
		if err:=recover();err != nil {
			log.Error("Http2FeibeeWonlyGuard error = ", err)
			return
		}
	}()

	var msg entity.WonlyGuardMsgFromApp
	if err := json.Unmarshal([]byte(appData), &msg); err != nil {
		log.Warning("Http2FeibeeWonlyGuard json.Unmarshal() error = ", err)
		return
	}
	log.Infof("Send WonlyGuard '%s' control to feibee", msg.Devid)

	var reqMsg entity.Req2Feibee

	reqMsg.Act = "standardWriteAttribute"
	reqMsg.Code = "286"
	reqMsg.Bindid = msg.Bindid
	
	key := "W" + msg.Devid + "only"
	var err error
	reqMsg.Bindstr, err = WonlyGuardAESDecrypt(msg.Bindstr, key)
	if err != nil {
		log.Warningf("Http2FeibeeWonlyGuard WonlyGuardAESDecrypt() error = ", err)
		return
	}
	reqMsg.Ver = "2.0"
	reqMsg.Devs = append(reqMsg.Devs, entity.ReqDevInfo2Feibee{
		Uuid:  msg.Devid,
		Value: msg.Value,
	})

	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		log.Warning("Http2FeibeeWonlyGuard json.Marshal() error = ", err)
		return
	}

	log.Debug("Send to Feibee: ", string(reqData))
	respData, err := doHttpReq("POST", "https://dev.fbeecloud.com/devcontrol/", reqData)
	var respMsg entity.RespFromFeibee
	err = json.Unmarshal(respData, &respMsg)
	if err != nil {
		log.Warning("Control WonlyGuard failed")
		return
	}

	if respMsg.Code != 1 {
		log.Warning("Control WonlyGuard failed")
	} else {
		log.Info("Control WonlyGuard successfully")
	}
}

func doHttpReq(method, url string, data []byte) (respData []byte, err error) {
	body := bytes.NewReader(data)
	httpClient := http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return respData, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}

	respData, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
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