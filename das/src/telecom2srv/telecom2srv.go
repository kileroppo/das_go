package telecom2srv

import (
	"bytes"
	"fmt"
	"github.com/dlintw/goconf"
	"../core/log"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"../httpJob"
)

func Telecom2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	// 判断是否为https协议
	var httpPort int

	// 判断是否为https协议
	isHttps, err := conf.GetBool("telecom2http", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	}

	httpPort, _ = conf.GetInt("telecom2http", "telecom2https_port")

	srv := &http.Server{Addr: ":"+strconv.Itoa(httpPort)}

	http.HandleFunc("/telecom", Entry)

	go func() {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Httpserver: ListenAndServeTLS(): %s", err_https)
			}
		} else {
			log.Debug("httpServerStart http.ListenAndServe()......")
			if err_http := srv.ListenAndServe(); err_http != nil {
				// cannot panic, because this probably is an intentional close
				log.Error("Httpserver: ListenAndServe(): %s", err_http)
			}
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func Entry(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //解析参数，默认是不会解析的
	if ("GET" == req.Method) { // 基本配置：oneNET校验第三方接口
		log.Debug("httpJob.init MaxWorker: ", httpJob.MaxWorker, ", MaxQueue: ", httpJob.MaxQueue)
		msg := req.Form.Get("msg")
		// signature := req.Form.Get("signature")
		// nonce := req.Form.Get("nonce")
		if("" != msg) { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to telecom, ", msg)
		}
	} else if ("POST" == req.Method) { // 接收OneNET推送过来的数据
		result, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Error("get req.Body failed")
		} else {
			log.Debug("telecom.Entry() get: ", bytes.NewBuffer(result).String())

			// fetch job
			work := httpJob.Job { Serload: httpJob.Serload { DValue : bytes.NewBuffer(result).String(), Imei:"111"}}
			httpJob.JobQueue <- work
		}
	}
}