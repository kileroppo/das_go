package httpgo

import (
	"das/core/log"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func Http2DingDaily(reqBody string) (respBody string, err error) {
	log.Info("Http2DingDaily()......")
	mydata := "{\"msgtype\":\"text\",\"text\":{\"content\":\"" + reqBody + "\"},\"at\":{\"atMobiles\":[],\"isAtAll\":true}}"

	req_body := bytes.NewBuffer([]byte(mydata))
	log.Debug(req_body)

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

	sUrl := "https://oapi.dingtalk.com/robot/send?access_token=b6bafcaba9fab1c97f4c2c9fe2303750b05ea3ce0ce9b25c6c4802fcd54a7bf3"
	log.Debug("Http2DingDaily() ", sUrl, ", mydata=", mydata)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("Http2DingDaily http.NewRequest()，error=", err0)
		return "", err0
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("Http2DingDaily client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2DingDaily ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Info("Http2DingDaily() ", string(body))
		return string(body), nil
	} else {
		log.Error("Http2DingDaily Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2DingDaily ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Info("Http2DingDaily() ", string(body))
		return "", err1
	}
}
