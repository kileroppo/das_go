package http2srv

import (
	"crypto/tls"

	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber"
	"github.com/tidwall/gjson"

	"das/core/entity"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/util"
	"das/feibee2srv"
)

var (
	app *fiber.App
)

func Init() {
	go Http2SrvStart()
}

func Http2SrvStart() {
    app = fiber.New()
    app.Post("/xm", XMAlarmMsgHandler)
    app.Post("/yk", YKMsgHandler)
    app.Post("/rg", RGMsgHandler)
    app.All("/feibee", feibee2srv.FeibeeHandler)

    isTLS, _ := log.Conf.GetBool("https", "is_https")
	httpPort, _ := log.Conf.GetInt("https", "https_port")
    if isTLS {
		certFile, _ := log.Conf.GetString("https", "https_server_crt")
		keyFile, _ := log.Conf.GetString("https", "https_server_key")
		cert,err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(err)
		}

		tlsCfg := &tls.Config{
			Certificates:                []tls.Certificate{cert},
		}

		app.Listen(httpPort, tlsCfg)
	} else {
		app.Listen(httpPort)
	}
}

func Close() {
    app.Shutdown()
}

//func OtherVendorHttp2SrvStart(conf *goconf.ConfigFile) *http.Server {
//	isHttps, err := conf.GetBool("xm2http", "is_https")
//
//	if err != nil {
//		log.Errorf("读取https配置失败, %s\n", err)
//		os.Exit(1)
//	}
//
//	httpPort, _ := conf.GetInt("xm2http", "xm2http_port")
//
//	srv := &http.Server{
//		Addr: ":" + strconv.Itoa(httpPort),
//	}
//
//	http.HandleFunc("/xm", XMAlarmMsgHandler)
//	http.HandleFunc("/yk", YKMsgHandler)
//	http.HandleFunc("/rg", RGMsgHandler)
//
//	go func() {
//		if isHttps {
//			log.Info("OtherVendorHttp2SrvStart ListenAndServeTLS() start...")
//			serverCrt, _ := conf.GetString("https", "https_server_crt")
//			serverKey, _ := conf.GetString("https", "https_server_key")
//			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
//				log.Error("OtherVendorHttp2SrvStart ListenAndServeTLS() error = ", err_https)
//			}
//		} else {
//			log.Info("OtherVendorHttp2SrvStart ListenAndServer() start...")
//			if err_http := srv.ListenAndServe(); err_http != nil {
//				log.Error("OtherVendorHttp2SrvStart ListenAndServer() error = ", err_http)
//			}
//		}
//	}()
//
//	return srv
//}

func XMAlarmMsgHandler(c *fiber.Ctx) {
	log.Infof("XMAlarmMsgHandler recv: %s", c.Body())
}

func YKMsgHandler(c *fiber.Ctx) {
	rabbitmq.SendGraylogByMQ("遥看Server-http->DAS: %s", c.Body())
	jobque.JobQueue <- NewYKJob(util.Str2Bytes(c.Body()))
}

func RGMsgHandler(c *fiber.Ctx) {
	rabbitmq.SendGraylogByMQ("锐吉Server-http->DAS: %s", c.Body())
	jobque.JobQueue <- RGJob{rawData: c.Body()}
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
	ProcessYKMsg(y.rawData)
}

func ProcessYKMsg(rawData []byte) {
	header := entity.Header{
		Cmd:     0xfb,
		DevType: "WonlyYKInfrared",
		Vendor:  "yk",
	}
	msg2app := entity.Feibee2DevMsg{
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
	header.DevId = msg.Devid
	msg2app.Online = msg.Online
	msg2app.OpValue = strconv.Itoa(msg.Online)
	msg2app.Time = msg.Timestamp
	msg2app.Header = header

	data, err := json.Marshal(msg2app)
	if err != nil {
		log.Warningf("ProcessYKMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2app(data, msg2app.DevId)
		rabbitmq.Publish2mns(data, "")
	}

	header.Cmd = 0x1200
	msg2pms := entity.OtherVendorDevMsg{
		Header:  header,
		OriData: string(rawData),
	}
	data,err = json.Marshal(msg2pms)
	if err != nil {
		log.Warningf("ProcessYKMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}

func NewYKJob(rawData []byte) YKJob {
	return YKJob{rawData: rawData}
}

type RGJob struct {
	rawData string
}

func (r RGJob) Handle() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	devId := gjson.Get(r.rawData, "mid").String()
	if len(devId) == 0 {
		return
	}

	msg := entity.OtherVendorDevMsg{
		Header: entity.Header{
			Cmd:     0x1200,
			DevId:   devId,
			Vendor:  "rg",
			DevType: "",
		},
		OriData: r.rawData,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Warningf("RGJob.Handle > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2pms(data, "")
	}
}
