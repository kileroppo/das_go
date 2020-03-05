package xm2srv

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dlintw/goconf"

	"das/core/log"
)

func XM2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	isHttps, err := conf.GetBool("xm2http", "is_https")

	if err != nil {
		log.Errorf("读取https配置失败, %s\n", err)
		os.Exit(1)
	}

	httpPort, _ := conf.GetInt("xm2http", "xm2http_port")

	srv := &http.Server{
		Addr: ":" + strconv.Itoa(httpPort),
	}

	http.HandleFunc("/xm", XMAlarmMsgHandler)

	go func() {
		if isHttps {
			log.Info("XM2HttpSrvStart ListenAndServeTLS() start...")
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("XM2HttpSrvStart ListenAndServeTLS() error = ", err_https)
			}
		} else {
			log.Info("XM2HttpSrvStart ListenAndServer() start...")
			if err_http := srv.ListenAndServe(); err_http != nil {
				log.Error("XM2HttpSrvStart ListenAndServer() error = ", err_http)
			}
		}
	}()

	return srv

}

func XMAlarmMsgHandler(res http.ResponseWriter, req *http.Request) {
	rawData, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		log.Error("Get XM alarm msg Body failed")
	} else {
		log.Info("Get XM alarm msg: ", string(rawData))
	}
}
