package httpgo

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"

	"../entity"
	"../log"
	"../util"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Http2FeibeeWonlyGuard(appData string) {
	var msg entity.WonlyGuardMsgFromApp
	if err := json.Unmarshal([]byte(appData), &msg); err != nil {
		log.Warning("Http2FeibeeWonlyGuard json.Unmarshal() error = ", err)
		return
	}

	var reqMsg entity.Req2Feibee

	reqMsg.Act = "standardWriteAttribute"
	reqMsg.Code = "286"
	reqMsg.Bindid = msg.Bindid
	
	key := util.Md5("W" + msg.Devid + "only")
	var err error
	reqMsg.Bindstr, err = util.ECBDecrypt(msg.Bindstr, []byte(key))
	if err != nil {
		log.Warningf("Http2FeibeeWonlyGuard ECBDecrypt() error = ", err)
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
