package httpgo

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"das/core/log"
)

var ErrBadResp = errors.New("HTTP Response error")

var feibeeHTTPClient *http.Client

func init() {
	transport := &http.Transport{
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: time.Second * 2,
	}
	feibeeHTTPClient = &http.Client{
		Transport: transport,
	}
}

func DoHTTPReq(req *http.Request) (respData []byte, err error) {
	resp, err := feibeeHTTPClient.Do(req)
	if err != nil {
		log.Warning("DoHTTPReq() error = ", err)
		return
	}

	if resp.StatusCode != 200 {
	    err = ErrBadResp
	    return
	}

	respData,err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	return
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

func GetHTTP(url string) (respData []byte, err error){
	return DoHTTP("GET", url, []byte{})
}