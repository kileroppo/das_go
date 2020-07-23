package httpgo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"das/core/log"
	"das/core/redis"
)

var hClient *http.Client

func init() {
	transport := &http.Transport{
		TLSHandshakeTimeout: time.Second * 3,
	}
	hClient = &http.Client{
		Transport: transport,
	}
}

func DoHTTPReq(req *http.Request) (respData []byte, err error) {
	resp, err := hClient.Do(req)
	if err != nil {
		log.Error("DoHTTPReq() error = ", err)
		return
	}

	respData,err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return
}

func DoHTTPReqWithResp(req *http.Request) (resp *http.Response, err error) {
	return hClient.Do(req)
}

func DoHTTP(method, url string, data []byte) (respData []byte, err error) {
	body := bytes.NewReader(data)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	return DoHTTPReq(req)
}

func DoFeibeeControlReq(data []byte) (respData []byte, err error) {
	url := ""
	url,err = log.Conf.GetString("feibee2http", "url_control")
	if err != nil {
		err = fmt.Errorf("DoFeibeeControlReq > %w", err)
		return
	}

	if redis.IsFeibeeSpSrv(data) {
		url,err = log.Conf.GetString("feibee2http", "url_control_sp")
		if err != nil {
			err = fmt.Errorf("DoFeibeeControlReq > %w", err)
			return
		}
	}

	return DoHTTP("POST", url, data)
}