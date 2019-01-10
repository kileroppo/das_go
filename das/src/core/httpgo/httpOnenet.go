package httpgo

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"../log"
	"bytes"
	)

func Http2OneNET_exe(imei string,  sBody string) {
	mydata := "{\"args\":\"" + sBody + "\"}"
	req_body := bytes.NewBuffer([]byte(mydata))
	log.Error(req_body)

	client := &http.Client{}
	sUrl := "http://api.zj.cmcconenet.com/nbiot/execute?imei=" + imei + "&obj_id=3201&obj_inst_id=0&res_id=5750&timeout=30"
	req, err := http.NewRequest("POST", sUrl, req_body)
	if err != nil {
		// handle error
		log.Error("Http2OneNET_exe http请求下发命令到OneNET失败")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", "HH=A=y1D9vuArz1JTcpvReUf5Uc=")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2OneNET_exe ReadAll Body 1 failed，err=", err)
		}

		fmt.Println(string(body))
	} else {
		log.Error("Http2OneNET_exe Post failed，resp.StatusCode=", resp.StatusCode, ", err=", err)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2OneNET_exe ReadAll Body 2 failed，err=", err)
		}

		fmt.Println(string(body))
	}
}

func Http2OneNET_write(imei string,  sBody string) {
	mydata := "{\"data\":[{\"res_id\":5750,\"val\":'" + sBody + "'}]}"

	req_body := bytes.NewBuffer([]byte(mydata))
	fmt.Println(req_body)

	client := &http.Client{}
	sUrl := "http://api.zj.cmcconenet.com/nbiot?imei=" + imei + "&obj_id=3201&obj_inst_id=0&mode=1"
	fmt.Println(sUrl)
	req, err := http.NewRequest("POST", sUrl, req_body)
	if err != nil {
		// handle error
		fmt.Println("Http2OneNET_write http请求下发命令到OneNET失败")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", "HH=A=y1D9vuArz1JTcpvReUf5Uc=")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			fmt.Println("Http2OneNET_write ReadAll Body 1 failed，err=", err)
		}

		fmt.Println(string(body))
	} else {
		fmt.Println("Http2OneNET_write Post failed，resp.StatusCode=", resp.StatusCode, ", err=", err)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			fmt.Println("Http2OneNET_write ReadAll Body 2 failed，err=", err)
		}

		fmt.Println(string(body))
	}
}