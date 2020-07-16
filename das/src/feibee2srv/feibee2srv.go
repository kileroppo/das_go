package feibee2srv

import (
	"context"
	"errors"
	"github.com/etcd-io/etcd/clientv3"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dlintw/goconf"
	"github.com/tidwall/gjson"

	"das/core/entity"
	"das/core/etcd"
	"das/core/jobque"
	"das/core/log"
	"das/core/rabbitmq"
)

var (
	ErrMsgInvalid = errors.New("msg was invalid")
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
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	err := ProcessFeibeeMsg(f.rawData)
	if err != nil {
		log.Warningf("FeibeeJob.Handle > %s", err)
	}
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
			log.Info("Feibee2HttpSrvStart ListenAndServeTLS() start...")
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Feibee2HttpSrvStart ListenAndServeTLS() error = ", err_https)
			}
		} else {
			log.Info("Feibee2HttpSrvStart ListenAndServer() start...")
			if err_http := srv.ListenAndServe(); err_http != nil {
				log.Error("Feibee2HttpSrvStart ListenAndServer() error = ", err_http)
			}
		}
	}()

	return srv

}

func FeibeeHandler(res http.ResponseWriter, req *http.Request) {
	rawData, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		log.Error("get feibee http Body failed")
	} else {
		jobque.JobQueue <- NewFeibeeJob(rawData)
	}
}

type FeibeeData struct {
	data entity.FeibeeData
}

func ProcessFeibeeMsg(rawData []byte) (err error) {
	feibeeData, err := NewFeibeeData(rawData)
	if err != nil {
		return err
	}

	go sendFeibeeLogMsg(rawData)

	seqId := gjson.GetBytes(rawData, "seqId").Int()
	if seqId > 0 {
		go setSceneResultCache(rawData)
	}

	//feibee数据合法性检查
	if !feibeeData.isDataValid() {
		log.Warningf("ProcessFeibeeMsg > %s > msg: %s", ErrMsgInvalid, rawData)
		return err
	}

	//feibee数据推送到MQ
	feibeeData.push2MQ()

	return nil
}

func setSceneResultCache(rawData []byte) {
	seq := gjson.GetBytes(rawData, "seqId").String()
	bindid, val := "", ""

	val = "1"
	bindid = gjson.GetBytes(rawData, "bindid").String()

	etcdClt := etcd.GetEtcdClient()
	if etcdClt == nil {
		log.Error("setSceneResultCache > etcd.GetEtcdClient > get etcd failed")
		return
	}
	key := bindid+"_"+seq
	resp,err := etcdClt.Get(context.Background(), key)
	if err != nil {
		return
	}
	if len(resp.Kvs) <= 0 {
		return
	}
	rawVal := resp.Kvs[0].Value
	vals := strings.Split(string(rawVal), "_")
	if len(vals) > 1 {
		leaseId, err := strconv.ParseInt(vals[1], 10, 64)
		val += "_" + vals[1]
		if err == nil {
			log.Infof("Set etcd[%s] %s", key, val)
			etcdClt.Put(context.Background(), key, val, clientv3.WithLease(clientv3.LeaseID(leaseId)))
		}
	}
}

func NewFeibeeData(data []byte) (FeibeeData, error) {
	var feibeeData FeibeeData

	if err := json.Unmarshal(data, &feibeeData.data); err != nil {
		log.Errorf("NewFeibeeData > json.Unmarshal error > %s", data)
		return feibeeData, err
	}

	return feibeeData, nil
}

func (f *FeibeeData) isDataValid() bool {
	if f.data.Status != "" && f.data.Ver != "" {
		switch f.data.Code {
		case 3, 4, 5, 7, 10, 12:
			if len(f.data.Msg) > 0 {
				return true
			}
		case 2:
			if len(f.data.Records) > 0 {
				return true
			}
		case 15, 32:
			if len(f.data.Gateway) > 0 {
				return true
			}
		case 21,22,23:
			if len(f.data.SceneMessages) > 0 {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func (f *FeibeeData) push2MQ() {
	//飞比推送数据条数 分条处理
	datas := splitFeibeeMsg(&f.data)

	for _, data := range datas {
		msgHandle := msgHandleFactory(&data)
		if msgHandle == nil {
			return
		}
		msgHandle.PushMsg()
	}

}

func splitFeibeeMsg(data *entity.FeibeeData) (datas []entity.FeibeeData) {

	switch data.Code {
	case 3, 4, 5, 7, 12, 10:
		datas = make([]entity.FeibeeData, len(data.Msg))
		for i := 0; i < len(data.Msg); i++ {
			datas[i].Msg = []entity.FeibeeDevMsg{
				data.Msg[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	case 2:
		datas = make([]entity.FeibeeData, len(data.Records))
		for i := 0; i < len(data.Records); i++ {
			datas[i].Records = []entity.FeibeeRecordsMsg{
				data.Records[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	case 15,32:
		datas = make([]entity.FeibeeData, len(data.Gateway))
		for i := 0; i < len(data.Gateway); i++ {
			datas[i].Gateway = []entity.FeibeeGatewayMsg{
				data.Gateway[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	case 21,22,23:
		datas = make([]entity.FeibeeData, len(data.SceneMessages))
		for i:=0; i<len(data.SceneMessages); i++ {
			datas[i].SceneMessages = []entity.FeibeeSceneMsg{
				data.SceneMessages[i],
			}
			datas[i].Code = data.Code
			datas[i].Ver = data.Ver
			datas[i].Status = data.Status
		}
	}

	return
}

func sendFeibeeLogMsg(rawData []byte) {
    var logMsg entity.SysLogMsg

    logMsg.Timestamp = time.Now().Unix()
    logMsg.MsgType = 1
    logMsg.RawData = string(rawData)

    data,err := json.Marshal(logMsg)
    if err != nil {
    	log.Warningf("sendFeibeeLogMsg > json.Marshal > %s", err)
	} else {
		rabbitmq.Publish2log(data, "")
	}
}
