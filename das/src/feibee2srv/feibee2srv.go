package feibee2srv

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dlintw/goconf"

	"../core/log"
	"../core/jobque"
	"bytes"
)

type FeibeeJob struct {
	rawData []byte
}

func NewFeibeeJob(rawData []byte) FeibeeJob {

	return FeibeeJob{
		rawData: rawData,
	}

}

func (f FeibeeJob) Handle() {
	log.Debug("feibee2srv.Handle() get: ", bytes.NewBuffer(f.rawData).String())
	ProcessFeibeeMsg(f.rawData)
}

func Feibee2HttpSrvStart(conf *goconf.ConfigFile) *http.Server {
	isHttps, err := conf.GetBool("feibee2http", "is_https")

	if err != nil {
		log.Errorf("读取https配置失败, %s\n", err)
		os.Exit(1)
	}

	httpPort, _ := conf.GetInt("feibee2http", "feibee2http_port")

	srv := &http.Server{
		Addr: ":" + strconv.Itoa(httpPort),
	}

	http.HandleFunc("/feibee", FeibeeHandler)

	go func() {
		if isHttps {
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Httpserver: ListenAndServeTLS(): %s", err_https)
			}
		} else {
			if err_http := srv.ListenAndServe(); err_http != nil {
				log.Error("feibeeServer ListenAndServer() error=", err_http)
			}
		}
	}()

	return srv

}

func FeibeeHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method != "POST" {
		log.Error("feibee推送http接口http方法错误")
	} else {
		rawData, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Error("get feibee http Body failed")
		} else {
			jobque.JobQueue <- NewFeibeeJob(rawData)
		}
	}
}
