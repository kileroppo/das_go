package httpJob

import (
	"runtime"
	"fmt"
	"net/http"
	"io/ioutil"
	"bytes"
	"../core/log"
)

var (
	// Max_Num = os.Getenv("MAX_NUM")
	MaxWorker = runtime.NumCPU()/2
	MaxQueue  = 1000
)


type Job struct {
	serload Serload
}

var JobQueue chan Job

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	Quit chan bool
}

func NewWorker(workPool chan chan Job) Worker {
	return Worker {
		WorkerPool:workPool,
		JobChannel:make(chan Job),
		Quit:make(chan bool),
	}
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				// excute job
				// fmt.Println(job.serload.pri)
				job.serload.ProcessJob();
			case <-w.Quit:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}

type Dispatcher struct {
	MaxWorkers int
	WorkerPool chan chan Job
	Quit chan bool
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{MaxWorkers: maxWorkers, WorkerPool: pool, Quit: make(chan bool)}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.MaxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.Dispatch()
}

func (d *Dispatcher) Stop() {
	go func() {
		d.Quit <- true
	}()
}

func (d *Dispatcher) Dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				jobChannel := <-d.WorkerPool
				jobChannel <- job
			}(job)

		case <-d.Quit:
			return
		}
	}
}

func Entry(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //解析参数，默认是不会解析的
	if ("GET" == req.Method) { // 基本配置：oneNET校验第三方接口
		log.Debug("httpJob.init MaxWorker: ", MaxWorker, ", MaxQueue: ", MaxQueue)
		msg := req.Form.Get("msg")
		// signature := req.Form.Get("signature")
		// nonce := req.Form.Get("nonce")
		if("" != msg) { // 存在则返回msg
			fmt.Fprintf(res, msg)
			log.Info("return msg to OneNET, ", msg)
		}
	} else if ("POST" == req.Method) { // 接收OneNET推送过来的数据
		result, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Error("get req.Body failed")
		} else {
			// fetch job
			work := Job { serload: Serload { pri : bytes.NewBuffer(result).String() }}
			JobQueue <- work
		}
	}
}

func init() {
	log.Debug("httpJob.init MaxWorker: ", MaxWorker, ", MaxQueue: ", MaxQueue)
	runtime.GOMAXPROCS(MaxWorker)
	JobQueue = make(chan Job, MaxQueue)
	dispatcher := NewDispatcher(MaxWorker)
	dispatcher.Run()
}

func checkToken(msg string, nonce string, signature string, token string) bool {


	return true
}