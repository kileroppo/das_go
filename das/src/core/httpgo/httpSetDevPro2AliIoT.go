package httpgo

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"das/core/redis"
	"das/core/go-aliyun-sign"
	"das/core/log"
)
//{"code":200,"data":{"isolationId":"a103HWXIkjOlEuzM","expireIn":7200000,"cloudToken":"2b796e0a78a444f084c5512b29dbc37e"},"id":"1509086454181"}
type aliData struct {
	IsolationId string	`json:"isolationId"`
	ExpireIn int32		`json:"expireIn"`
	CloudToken string	`json:"cloudToken"`
}
type aliIoTResp struct {
	Code int			`json:"code"`
	AliData aliData		`json:"data"`
	Id string			`json:"id"`
}

type aliIoTResp2 struct {
	Code int			`json:"code"`
	AliData interface{}	`json:"data"`
	Id string			`json:"id"`
}

// 获取cloud/token
func getAliIoTCloudToken() (token string, err error) {
	log.Info("getAliIoTCloudToken start.")

	mydata := "{\"id\":\"1509086454180\",\"version\":\"1.0\",\"request\":{\"apiVer\":\"1.0.0\"},\"params\":{\"grantType\":\"project\",\"res\":\"a124mCmKp3GjHD9y\"}}"
	req_body := bytes.NewBuffer([]byte(mydata))
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(30 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*30)
				if err != nil {
					log.Error("getAliIoTCloudToken net.DialTimeout，err=", err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	sUrl := "https://api.link.aliyun.com/cloud/token"
	log.Debug("getAliIoTCloudToken() ", sUrl, ", sBody=", mydata)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("getAliIoTCloudToken http.NewRequest()，error=", err0)
		return "", err0
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Sign the request
	if err_ := sign.Sign(req, aliAppkey, aliAppSecret); err_ != nil {
		// Handle error
		log.Error("getAliIoTCloudToken sign.Sign()，error=", err_)
		return "", err_
	}

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error(" getAliIoTCloudToken client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("getAliIoTCloudToken ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Debug("getAliIoTCloudToken() ", string(body))
		return string(body), nil
	} else {
		log.Error("getAliIoTCloudToken Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("getAliIoTCloudToken ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Debug("getAliIoTCloudToken() ", string(body))
		err2 := errors.New("get aliIoT Token failed.")
		return "", err2
	}

	return "", nil
}

// 设置设备属性
func setAliThingPro(token, deviceName, data, cmd string) (respBody string, err error) {
	log.Info("[", deviceName, "]HttpSetAliThingPro start.")

	mydata := "{\"id\":\"1509086454180\",\"version\":\"1.0\",\"request\":{\"apiVer\":\"1.0.2\",\"cloudToken\":\"" + token + "\"},\"params\":{\"productKey\":\"a1xhJ6eIxsn\",\"deviceName\":\"" + deviceName + "\",\"items\":{\"UserData\":\"" + data +"\"}}}"
	req_body := bytes.NewBuffer([]byte(mydata))
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(30 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*30)
				if err != nil {
					log.Error("Http2OneNET_write net.DialTimeout，err=", err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	sUrl := "https://api.link.aliyun.com/cloud/thing/properties/set"
	log.Debug("[", deviceName, "] "+cmd+" HttpSetAliThingPro() ", sUrl, ", sBody=", mydata)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("[", deviceName, "] "+cmd+" HttpSetAliThingPro http.NewRequest()，error=", err0)
		return "", err0
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Sign the request
	if err_ := sign.Sign(req, aliAppkey, aliAppSecret); err != nil {
		// Handle error
		log.Error("HttpSetAliThingPro sign.Sign()，error=", err_)
		return "", err_
	}

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("[", deviceName, "] "+cmd+" HttpSetAliThingPro client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("[", deviceName, "] "+cmd+" HttpSetAliThingPro ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Debug("[", deviceName, "] "+cmd+" HttpSetAliThingPro() ", string(body))
		return string(body), nil
	} else {
		log.Error("[", deviceName, "] "+cmd+" HttpSetAliThingPro Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("[", deviceName, "] "+cmd+" HttpSetAliThingPro ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Debug("[", deviceName, "] "+cmd+" HttpSetAliThingPro() ", string(body))
		return string(body), err1
	}

	return "", nil
}

func HttpSetAliPro(deviceName, data, cmd string) (int, error) {
	//1. 从缓存中获取token
	token, err := redis.GetAliIoTtoken()
	if nil != err || "" == token {
		log.Error("[", deviceName, "] HttpSetAliPro() redis.GetAliIoTtoken, token=", token, ", error=", err)
		tokenStr, err1 := getAliIoTCloudToken()
		if nil != err1 {
			log.Error("[", deviceName, "] HttpSetAliPro() getAliIoTCloudToken error, ", err1)
			return 0, err1
		}

		var tokenResp aliIoTResp
		if err = json.Unmarshal([]byte(tokenStr), &tokenResp); err != nil {
			log.Error("[", deviceName, "] HttpSetAliPro() json.Unmarshal 1, err=", err)
			return 0, err // break
		}

		token = tokenResp.AliData.CloudToken

		// 存储token
		go redis.SetAliIoTtoken(tokenResp.AliData.CloudToken, strconv.Itoa(int(tokenResp.AliData.ExpireIn/1000)))
	}

	//2. 下发数据给设备
	resp, err2 := setAliThingPro(token, deviceName, data, cmd)
	if nil != err2 {
		log.Error("[", deviceName, "] HttpSetAliPro() setAliThingPro 1, err=", err2)
	}

	var aliResp aliIoTResp2
	if err = json.Unmarshal([]byte(resp), &aliResp); err != nil {
		log.Error("[", deviceName, "] HttpSetAliPro() json.Unmarshal 2, err=", err)
		return 0, err // break
	}
	if 200 != aliResp.Code { // 接口请求失败，重新获取token，然后下行数据到设备
		tokenStr, err1 := getAliIoTCloudToken()
		if nil != err1 {
			log.Error("[", deviceName, "] HttpSetAliPro() getAliIoTCloudToken error, ", err1)
			return 0, err1
		}

		var tokenResp aliIoTResp
		if err = json.Unmarshal([]byte(tokenStr), &tokenResp); err != nil {
			log.Error("[", deviceName, "] HttpSetAliPro() json.Unmarshal 1, err=", err)
			return 0, err // break
		}

		token = tokenResp.AliData.CloudToken

		// 存储token
		go redis.SetAliIoTtoken(tokenResp.AliData.CloudToken, strconv.Itoa(int(tokenResp.AliData.ExpireIn/1000)))

		resp2, err2 := setAliThingPro(token, deviceName, data, cmd)
		if nil != err2 {
			log.Error("[", deviceName, "] HttpSetAliPro() setAliThingPro 2, err=", err2, ", resp=", resp2)
			return 0, err2
		}
		log.Debug("[", deviceName, "] HttpSetAliPro() setAliThingPro 2, resp=", resp2)

		var aliResp2 aliIoTResp2
		if err = json.Unmarshal([]byte(resp2), &aliResp2); err != nil {
			log.Error("[", deviceName, "] HttpSetAliPro() json.Unmarshal 2, err=", err)
			return 0, err // break
		}
		return aliResp2.Code, nil
	}

	return aliResp.Code, nil
}