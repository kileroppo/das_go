package aliIot2srv

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dlintw/goconf"

	"../core/entity"
	"../core/h2client"
	"../core/httpgo"
	"../core/jobque"
	"../core/log"
	"../core/rabbitmq"
)

var (
	ctx, cancel = context.WithCancel(context.Background())
)

func Run() {
	conf := log.Conf
	appKey, err := conf.GetString("aliIoT2http", "appKey")
	if err != nil {
		log.Error("get-aliIoT2http-appKey error = ", err)
		return
	}

	appSecret, err := conf.GetString("aliIoT2http", "appSecret")
	if err != nil {
		log.Error("get-aliIoT2http-appSecret error = ", err)
		return
	}

	httpgo.InitAliIoTConfig(appKey, appSecret)

	msgs, err := rabbitmq.Consumer2aliMQ.Consumer()
	if err != nil {
		log.Error("Consumer2aliMQ.Consumer() error = ", err)
		if err = rabbitmq.Consumer2aliMQ.ReConn(); err != nil {
			log.Warningf("Consumer2aliMQ Reconnection Failed")
			return
		}
		log.Debug("Consumer2aliMQ Reconnection Successful")
		msgs, err = rabbitmq.Consumer2aliMQ.Consumer()
	}

	for msg := range msgs {
		//log.Info("ReceiveMQMsgFromAli: ", string(msg.Body))
		jobque.JobQueue <- NewAliJob(msg.Body)
	}

	select {
	case <-ctx.Done():
		log.Info("ReceiveMQMsgFromAli Close")
		return
	default:
		go Run()
		return
	}
}

func Close() {
	cancel()
}

type AliIOTSrv struct {
	rawUrl string
	topic  string

	appKey    string
	appSecret string

	ctx       context.Context
	cancel    context.CancelFunc
	cancelCli context.CancelFunc

	reConnNum     int32
	currReConnNum int32
}

func (a *AliIOTSrv) Run() {

	go func() {
		for {
			select {
			default:
				err := a.conn()
				if err != nil {
					if atomic.LoadInt32(&a.currReConnNum) < a.reConnNum {
						log.Warningf("AliIOTSrv第%d重连中...", a.currReConnNum+1)
						atomic.AddInt32(&a.currReConnNum, 1)
						a.cancelCli()
						time.Sleep(time.Second * 3)
						continue
					} else {
						log.Warningf("AliIOTSrv重连失败")
						//a.Close()
						return
					}
				}
			}
		}
	}()

	go func() {
		for rawData := range h2client.AliDataCh {
			jobque.JobQueue <- NewAliIOTJob(rawData)
		}
		log.Debug("AliDataCh close")
	}()

}

func (a *AliIOTSrv) conn() error {
	ctxCli, cancelCli := context.WithCancel(a.ctx)
	a.cancelCli = cancelCli
	h2 := h2client.Newh2Client(ctxCli)
	h2.SetAliHeader(a.appKey, a.appSecret)
	return h2.Get(a.rawUrl + a.topic)
}

func (a *AliIOTSrv) Close() {
	close(h2client.AliDataCh)
	a.cancel()
	log.Info("AliIOTSrv closed")
}

func NewAliIOT2Srv(conf *goconf.ConfigFile) *AliIOTSrv {
	ctx, cancel := context.WithCancel(context.Background())

	rawUrl, err := conf.GetString("aliIoT2http2", "ali_url")
	if err != nil {
		log.Error("get-aliIoT2http2-endPoint error = ", err)
		return nil
	}

	topic, err := conf.GetString("aliIoT2http2", "topic")
	if err != nil {
		log.Error("get-aliIoT2http2-topic error = ", err)
		return nil
	}

	appKeyH2, err := conf.GetString("aliIoT2http2", "appKey")
	if err != nil {
		log.Error("get-aliIoT2http2-topic error = ", err)
		return nil
	}

	appSecretH2, err := conf.GetString("aliIoT2http2", "appSecret")
	if err != nil {
		log.Error("get-aliIoT2http2-topic error = ", err)
		return nil
	}

	appKey, err := conf.GetString("aliIoT2http", "appKey")
	if err != nil {
		log.Error("get-aliIoT2http-appKey error = ", err)
		return nil
	}

	appSecret, err := conf.GetString("aliIoT2http", "appSecret")
	if err != nil {
		log.Error("get-aliIoT2http-appSecret error = ", err)
		return nil
	}

	httpgo.InitAliIoTConfig(appKey, appSecret)

	return &AliIOTSrv{
		ctx:    ctx,
		cancel: cancel,

		rawUrl: rawUrl,
		topic:  topic,

		appKey:    appKeyH2,
		appSecret: appSecretH2,

		reConnNum:     10,
		currReConnNum: 0,
	}
}

type AliIOTJob struct {
	rawData []byte
	topic   string
}

func (a AliIOTJob) Handle() {
	ProcessAliMsg(a.rawData, a.topic)
}

func NewAliIOTJob(aliRawData entity.AliRawData) AliIOTJob {
	return AliIOTJob{
		rawData: aliRawData.RawData,
		topic:   aliRawData.Topic,
	}
}

func NewAliJob(rawData []byte) (job AliIOTJob) {
	dataSli := strings.Split(string(rawData), "#")
	if len(dataSli) != 2 {
		return
	} else {
		job.topic = dataSli[0]
		job.rawData = []byte(dataSli[1])
	}

	return
}
