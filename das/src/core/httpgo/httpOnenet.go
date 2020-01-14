package httpgo

import (
	"bytes"
	"das/core/log"
	"fmt"
	"io"
	"io/ioutil"
	glog "log"
	"net/http"
)

func checkError(err error) {
	if err != nil {
		glog.Fatalln(err)
	}
}

func Http2OneNET_exe(imei string, sBody string) {
	mydata := "{\"args\":\"" + sBody + "\"}"
	req_body := bytes.NewBuffer([]byte(mydata))
	log.Error(req_body)

	//client := &http.Client{}
	sUrl := "http://api.heclouds.com/nbiot/execute?imei=" + imei + "&obj_id=3201&obj_inst_id=0&res_id=5750&timeout=30" // api.zj.cmcconenet.com,
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("Http2OneNET_exe http请求下发命令到OneNET失败，error=", err0)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", "6kjzYeG=oSVVPCi2n9FdnKBMehs=")

	resp, err1 := DoHTTPReqWithResp(req)
	// 关闭 resp.Body 的正确姿势
	if resp != nil {
		defer resp.Body.Close()

		// If reusing the http connection is important for your application you might need to add something like this at the end of your response processing logic.
		defer io.Copy(ioutil.Discard, resp.Body) // 手动丢弃读取完毕的数据
	}

	checkError(err1)

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2OneNET_exe ReadAll Body 1 failed，err=", err)
		}

		fmt.Println(string(body))
	} else {
		log.Error("Http2OneNET_exe Post failed，resp.StatusCode=", resp.StatusCode, ", err=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2OneNET_exe ReadAll Body 2 failed，err=", err)
		}

		fmt.Println(string(body))
	}
}

func Http2OneNET_write(imei string, sBody string, cmd string) (respBody string, err error) {
	log.Info("[", imei, "]Http2OneNET_write start_write......")
	mydata := "{\"data\":[{\"res_id\":5750,\"val\":'" + sBody + "'}]}"
	// mydata := "{\"data\":[{\"res_id\":5750,\"val\":\"" + sBody + "\"}]}"

	req_body := bytes.NewBuffer([]byte(mydata))
	log.Debug(req_body)

	//client := &http.Client{
	//	Transport: &http.Transport{
	//		Dial: func(netw, addr string) (net.Conn, error) {
	//			deadline := time.Now().Add(30 * time.Second)
	//			c, err := net.DialTimeout(netw, addr, time.Second*30)
	//			if err != nil {
	//				log.Error("Http2OneNET_write net.DialTimeout，err=", err)
	//				return nil, err
	//			}
	//			c.SetDeadline(deadline)
	//			return c, nil
	//		},
	//	},
	//}


	// sUrl := "http://api.zj.cmcconenet.com/nbiot?imei=" + imei + "&obj_id=3200&obj_inst_id=0&mode=1" // api.zj.cmcconenet.com, api.heclouds.com
	// sUrl := "http://api.heclouds.com/nbiot?imei=" + imei + "&obj_id=3200&obj_inst_id=0&mode=1"		// api.zj.cmcconenet.com, api.heclouds.com
	sUrl := oneNET_Url + imei + "&obj_id=3200&obj_inst_id=0&mode=1" // api.zj.cmcconenet.com, api.heclouds.com

	log.Debug("[", imei, "] "+cmd+" Http2OneNET_write() ", sUrl, ", sBody=", sBody)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("[", imei, "] "+cmd+" Http2OneNET_write http.NewRequest()，error=", err0)
		return "", err0
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", oneNET_Apikey) // 重庆：6kjzYeG=oSVVPCi2n9FdnKBMehs=, 浙江：HH=A=y1D9vuArz1JTcpvReUf5Uc=
	// req.Header.Set("api-key", "6kjzYeG=oSVVPCi2n9FdnKBMehs=") // 重庆：6kjzYeG=oSVVPCi2n9FdnKBMehs=, 浙江：HH=A=y1D9vuArz1JTcpvReUf5Uc=

	resp, err1 := DoHTTPReqWithResp(req)
	if nil != resp {
		defer resp.Body.Close()

		// If reusing the http connection is important for your application you might need to add something like this at the end of your response processing logic.
		defer io.Copy(ioutil.Discard, resp.Body) // 手动丢弃读取完毕的数据
	}

	if nil != err1 {
		// handle error
		log.Error("[", imei, "] "+cmd+" Http2OneNET_write client.Do, error=", err1)
		return "", err1
	}

	if 200 == resp.StatusCode {
		// body, err := ioutil.ReadAll(resp.Body)
		bodyBuf := new(bytes.Buffer)
		nLen, err := bodyBuf.ReadFrom(resp.Body)
		if err != nil {
			// handle error
			log.Error("[", imei, "] "+cmd+" Http2OneNET_write bodyBuf.ReadFrom() 1，error=", err)
			return "", err
		}

		log.Debug("[", imei, "] "+cmd+" Http2OneNET_write() ", bodyBuf.String(), nLen)
		return bodyBuf.String(), nil
	} else {
		log.Error("[", imei, "] "+cmd+" Http2OneNET_write Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		// body, err := ioutil.ReadAll(resp.Body)
		bodyBuf := new(bytes.Buffer)
		nLen, err := bodyBuf.ReadFrom(resp.Body)
		if err != nil {
			// handle error
			log.Error("[", imei, "] "+cmd+" Http2OneNET_write bodyBuf.ReadFrom() 2, error=", err)
			return "", err
		}

		log.Debug("[", imei, "] "+cmd+" Http2OneNET_write() ", bodyBuf.String(), nLen)
		return "", err1
	}
}
