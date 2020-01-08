package httpgo

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"das/core/constant"
	"das/core/log"
)

func HttpLoginTelecom() (respBody string, err error) {
	req_body := bytes.NewBuffer([]byte("appId=uVWENSQL_XSY_yPE1nImoVVbUj4a&secret=F9ufOj41Z9GIeawCbhffNG2MLCIa"))
	log.Debug(req_body)

	ca_1Path := "cert/ca_1.pem"
	ca_2Path := "cert/ca_2.pem"
	ca_3Path := "cert/ca_3.pem"

	pool := x509.NewCertPool()
	// 3. ca_1.pem（测试根证书）、ca_2.pem（商用根证书）、ca_3.pem（商用根证书）、PEM格式证书，无密码，三个根证书都需要添加到证书信任
	caCrt_1, err := ioutil.ReadFile(ca_1Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_1)

	caCrt_2, err := ioutil.ReadFile(ca_2Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_2)

	caCrt_3, err := ioutil.ReadFile(ca_3Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_3)

	// 4. server.crt与server.key        CERT格式证书 server.crt是证书，server.key是私钥 ，私钥密码IoM@1234
	cliCrt, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            pool,
				Certificates:       []tls.Certificate{cliCrt},
				InsecureSkipVerify: true},
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

	sUrl := "https://" + constant.Base_Dianxin_Url + "/iocm/app/sec/v1.1.0/login"
	log.Debug("HttpLoginDianxinIoT() ", sUrl)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("HttpLoginDianxinIoT http.NewRequest()，error=", err0)
		return "", err0
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("HttpLoginDianxinIoT client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpLoginDianxinIoT ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Debug("HttpLoginDianxinIoT() ", string(body))
		return string(body), nil
	} else {
		log.Error("HttpLoginDianxinIoT Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpLoginDianxinIoT ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Debug("HttpLoginDianxinIoT() ", string(body))
		return "", err1
	}
	// {"accessToken":"5ea043fd73a452e6054c10e4aaeb9d","tokenType":"bearer","refreshToken":"3517f31feac0323923c51698b5db3a75","expiresIn":3600,"scope":"default"}
}

func HttpSubNotifyTelecom() (respBody string, err error) {
	respLoginBody, loginErr := HttpLoginTelecom()
	if loginErr != nil {
		return "login failed", loginErr
	}
	var respMap map[string]interface{}
	if errRespMap := json.Unmarshal([]byte(respLoginBody), &respMap); errRespMap != nil {
		log.Error("HttpSubNotifyDianxinIoT json.Unmarshal respLoginBody failed，errRespMap=", errRespMap)
		return "json.Unmarshal string2map failed.", errRespMap
	}

	req_body := bytes.NewBuffer([]byte(
		"{\"appId\":\"uVWENSQL_XSY_yPE1nImoVVbUj4a\"," +
			"\"notifyType\":\"" + constant.DEVICE_DATA_CHANGED + "\"," +
			"\"callbackUrl\":\"http://139.196.221.163:10702\"}"))
	log.Debug(req_body)

	ca_1Path := "cert/ca_1.pem"
	ca_2Path := "cert/ca_2.pem"
	ca_3Path := "cert/ca_3.pem"

	pool := x509.NewCertPool()
	// 3. ca_1.pem（测试根证书）、ca_2.pem（商用根证书）、ca_3.pem（商用根证书）、PEM格式证书，无密码，三个根证书都需要添加到证书信任
	caCrt_1, err := ioutil.ReadFile(ca_1Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_1)

	caCrt_2, err := ioutil.ReadFile(ca_2Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_2)

	caCrt_3, err := ioutil.ReadFile(ca_3Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_3)

	// 4. server.crt与server.key        CERT格式证书 server.crt是证书，server.key是私钥 ，私钥密码IoM@1234
	cliCrt, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            pool,
				Certificates:       []tls.Certificate{cliCrt},
				InsecureSkipVerify: true},
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

	sUrl := "https://" + constant.Base_Dianxin_Url + "/iocm/app/sub/v1.2.0/subscriptions?ownerFlag=true"
	log.Debug("HttpSubNotifyDianxinIoT() ", sUrl)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("HttpSubNotifyDianxinIoT http.NewRequest()，error=", err0)
		return "", err0
	}

	req.Header.Set("app_key", "uVWENSQL_XSY_yPE1nImoVVbUj4a")
	req.Header.Set("Authorization", "Bearer "+respMap["accessToken"].(string))
	req.Header.Set("Content-Type", "application/json")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("HttpSubNotifyDianxinIoT client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 201 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpSubNotifyDianxinIoT ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Debug("HttpSubNotifyDianxinIoT() ", string(body))
		return string(body), nil
	} else {
		log.Error("HttpSubNotifyDianxinIoT Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpSubNotifyDianxinIoT ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Debug("HttpSubNotifyDianxinIoT() ", string(body))
		return "", err1
	}
}

func HttpCmd2DeviceTelecom(imei string, cmdBody string) (respBody string, err error) {
	respLoginBody, loginErr := HttpLoginTelecom()
	if loginErr != nil {
		return "HttpCmd2DeviceTelecom login failed", loginErr
	}
	var respMap map[string]interface{}
	if errRespMap := json.Unmarshal([]byte(respLoginBody), &respMap); errRespMap != nil {
		log.Error("HttpCmd2DeviceTelecom json.Unmarshal respLoginBody failed，errRespMap=", errRespMap)
		return "json.Unmarshal string2map failed.", errRespMap
	}

	req_body := bytes.NewBuffer([]byte(
		"{\"appId\":\"uVWENSQL_XSY_yPE1nImoVVbUj4a\"," +
			"\"deviceId\":\"" + imei + "\"," +
			"\"command\":{\"serviceId\": \"DoorLock\",\"method\": \"CHANGE_STATUS\",\"paras\":{\"cmd\":\"" + cmdBody + "\"}}," +
			"\"callbackUrl\":\"http://139.196.221.163:10702/telecom\", " +
			"\"maxRetransmit\":3" +
			"}"))
	log.Debug(req_body)

	ca_1Path := "cert/ca_1.pem"
	ca_2Path := "cert/ca_2.pem"
	ca_3Path := "cert/ca_3.pem"

	pool := x509.NewCertPool()
	// 3. ca_1.pem（测试根证书）、ca_2.pem（商用根证书）、ca_3.pem（商用根证书）、PEM格式证书，无密码，三个根证书都需要添加到证书信任
	caCrt_1, err := ioutil.ReadFile(ca_1Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_1)

	caCrt_2, err := ioutil.ReadFile(ca_2Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_2)

	caCrt_3, err := ioutil.ReadFile(ca_3Path)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}
	pool.AppendCertsFromPEM(caCrt_3)

	// 4. server.crt与server.key        CERT格式证书 server.crt是证书，server.key是私钥 ，私钥密码IoM@1234
	cliCrt, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            pool,
				Certificates:       []tls.Certificate{cliCrt},
				InsecureSkipVerify: true},
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(30 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*30)
				if err != nil {
					log.Error("HttpCmd2DeviceTelecom net.DialTimeout，err=", err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	sUrl := "https://" + constant.Base_Dianxin_Url + "/iocm/app/cmd/v1.4.0/deviceCommands"
	log.Debug("HttpCmd2DeviceTelecom() ", sUrl)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("HttpCmd2DeviceTelecom http.NewRequest()，error=", err0)
		return "", err0
	}

	req.Header.Set("app_key", "uVWENSQL_XSY_yPE1nImoVVbUj4a")
	req.Header.Set("Authorization", "Bearer "+respMap["accessToken"].(string))
	req.Header.Set("Content-Type", "application/json")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("HttpCmd2DeviceTelecom client.Do, error=", err1)
		return "", err1
	}

	defer resp.Body.Close()

	if 201 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpCmd2DeviceTelecom ioutil.ReadAll() 1，error=", err)
			return "", err
		}

		log.Debug("HttpCmd2DeviceTelecom() ", string(body))
		return string(body), nil
	} else {
		log.Error("HttpCmd2DeviceTelecom Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("HttpCmd2DeviceTelecom ioutil.ReadAll() 2, error=", err)
			return "", err
		}

		log.Debug("HttpCmd2DeviceTelecom() ", string(body))
		return "", err1
	}
}
