package aliIot2srv

import (
	"context"
	"time"

	"github.com/dlintw/goconf"

	"../core/entity"
	"../core/h2client"
	"../core/jobque"
	"../core/log"
)

var (
	URL    = "https://ilop.iot-as-http2.cn-shanghai.aliyuncs.com"
	Topic  = "/message/ack"
	ctx    context.Context
	cancel context.CancelFunc
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}

func AliIOT2SrvStart(conf *goconf.ConfigFile) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(h2client.AliDataCh)
				return
			default:
				err := conn(URL, Topic)
				if err != nil {
					log.Warning("重连中...")
					time.Sleep(time.Second * 3)
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

func conn(addr, topic string) error {
	h2 := h2client.Newh2Client(ctx)
	h2.SetAliHeader()
	return h2.Get(addr + topic)
}

func Shutdown() {
	cancel()
	log.Info("AliIOT2Srv closed")
}

type AliIOTJob struct {
	rawData []byte
	topic   string
}

func (a AliIOTJob) Handle() {
	log.Debug("aliIOT2srv.Handle() get: ")
	ProcessAliMsg(a.rawData, a.topic)
}

func NewAliIOTJob(aliRawData entity.AliRawData) AliIOTJob {
	return AliIOTJob{
		rawData: aliRawData.RawData,
		topic:   aliRawData.Topic,
	}
}
