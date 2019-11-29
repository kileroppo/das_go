package aliIot2srv

import (
	"context"

	"github.com/dlintw/goconf"

	"../core/entity"
	"../core/h2client"
	"../core/jobque"
	"../core/log"
	"../core/httpgo"
)

type AliIOTSrv struct {
	rawUrl string
	topic  string

	appKey    string
	appSecret string

	ctx    context.Context
	cancel context.CancelFunc
}

func (a *AliIOTSrv) Run() {

	go func() {
		for {
			select {
			case <-a.ctx.Done():
				close(h2client.AliDataCh)
				return
			default:
				err := a.conn()
				if err != nil {
					log.Warning("AliIOTSrv重连中...")
					//time.Sleep(time.Second * 3)
					continue
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

func (a AliIOTSrv) conn() error {
	h2 := h2client.Newh2Client(a.ctx)
	h2.SetAliHeader(a.appKey, a.appSecret)
	return h2.Get(a.rawUrl + a.topic)
}

func (a *AliIOTSrv) Close() {
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
