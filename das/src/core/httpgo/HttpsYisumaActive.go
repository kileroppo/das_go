package httpgo

import (
	"../entity"
	"../log"
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

func Http2YisumaActive(reqBody entity.YisumaHttpsReq) (respBody string, err error) {
	log.Info("Http2YisumaActive()......")
	bytesData, err := json.Marshal(reqBody)
	req_body := bytes.NewBuffer([]byte(bytesData))

	log.Debug(req_body)
	//忽略证书请求https
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	url := "https://api.cityunion.org.cn:8443/SCAPI_TEST/v1/issue" + "/projects/" + reqBody.Body.ProjectNo + "/chips/" + reqBody.Body.UId + "/activese"
	req, err0 := http.NewRequest("POST", url, req_body)
	if err0 != nil {
		log.Error("Http2YisumaActive http.NewRequest()，error=", err0)
		return "", err0
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("Http2YisumaActive client.Do, error=", err1)
		return "", err1
	}
	defer resp.Body.Close()
	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2YisumaActive ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Info("Http2YisumaActive() ", string(body))
		return string(body), nil
	} else {
		log.Error("Http2YisumaActive Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2YisumaActive ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Info("Http2YisumaActive() ", string(body))
		return "", err1
	}
}
