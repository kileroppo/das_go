package httpgo

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"das/core/log"
)

func Http2WonlyUpgrade(devType string) (b []byte, err error) {
	mydata := "{\"PUS\":{\"header\":{\"api_version\":\"1.0\",\"message_type\":\"MSG_PRODUCT_UPGRADE_DOWN_REQ\",\"seq_id\":\"1\"},\"body\":{\"token\":\"xxxxxxxxxxxxxxx\",\"vendor_name\":\"general\",\"platform\":\"device\",\"endpoint_type\":\"" + devType + "\",\"current_version\":\"1.0.0\"}}}"

	req_body := bytes.NewBuffer([]byte(mydata))
	log.Debug(req_body)

	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(30 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*30)
				if err != nil {
					log.Error("Http2WonlyUpgrade net.DialTimeout，err=", err)
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	sUrl := "https://pus.wonlycloud.com:10400"
	log.Debug("Http2WonlyUpgrade() ", sUrl, ", req_body=", mydata)
	req, err0 := http.NewRequest("POST", sUrl, req_body)
	if err0 != nil {
		// handle error
		log.Error("Http2WonlyUpgrade http.NewRequest()，error=", err0)
		return nil, err0
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err1 := client.Do(req)
	if nil != err1 {
		// handle error
		log.Error("Http2WonlyUpgrade client.Do, error=", err1)
		return nil, err1
	}

	defer resp.Body.Close()

	if 200 == resp.StatusCode {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2WonlyUpgrade ioutil.ReadAll() 1，error=", err)
			return nil, err
		}

		log.Info("Http2WonlyUpgrade() ", string(body))
		return body, nil
	} else {
		log.Error("Http2WonlyUpgrade Post failed，resp.StatusCode=", resp.StatusCode, ", error=", err1)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			log.Error("Http2WonlyUpgrade ioutil.ReadAll() 2, error=", err)
			return nil, err
		}

		log.Info("Http2WonlyUpgrade() ", string(body))
		return body, nil
	}
}
