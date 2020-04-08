package xm2srv

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/dlintw/goconf"

	"das/core/entity"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
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
	http.HandleFunc("/yk", XKMsgHandler)

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
		log.Infof("Get XM alarm msg: %s", rawData)
	}
}

func XKMsgHandler(res http.ResponseWriter, req *http.Request) {
	rawData, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		log.Error("get yk http Body failed")
	} else {
		jobque.JobQueue <- NewYKJob(rawData)
	}
}

type YKJob struct {
	rawData []byte
}

func (y YKJob) Handle() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
    //todo(zh): 遥看红外宝在线状态推送处理
	log.Debugf("yk2srv.Handle() get: %s", y.rawData)
	ProcessYKMsg(y.rawData)
}

func ProcessYKMsg(rawData []byte)  {
	msg2app := entity.Feibee2DevMsg{
		Header:        entity.Header{
			Cmd:0xfb,
			DevType:"WonlyYKInfrared",
			Vendor:"yk",
		},
		Note:          "",
		Deviceuid:     0,
		Online:        0,
		Battery:       0,
		OpType:        "newOnline",
		OpValue:       "",
		Time:          0,
		Bindid:        "",
		Snid:          "",
		SceneMessages: nil,
	}

	msg := entity.YKInfraredStatus{}
	if err := json.Unmarshal(rawData, &msg); err != nil {
		log.Warningf("ProcessYKMsg > json.Unmarshal > %s", err)
		return
	}
	msg2app.Online = msg.Online
	msg2app.DevId = msg.Devid
	msg2app.OpValue = strconv.Itoa(msg.Online)
	msg2app.Time = msg.Timestamp

	data2app,err := json.Marshal(msg2app)
	if err != nil {
		log.Warningf("ProcessYKMsg > json.Marshal > %s", err)
		return
	}
	rabbitmq.Publish2app(data2app, msg2app.DevId)
	rabbitmq.Publish2mns(data2app, "")
}

func NewYKJob(rawData []byte) YKJob {
    return YKJob{rawData:rawData}
}