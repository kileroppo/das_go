package consumer

import (
	"runtime"
	)

var (
	// Max_Num = os.Getenv("MAX_NUM")
	MaxWorker = runtime.NumCPU()/2
	MaxQueue  = 1000
)

type Job struct {
	appMsg AppMsg
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
				job.appMsg.ProcessAppMsg();
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

/*func Entry(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //解析参数，默认是不会解析的
	if ("GET" == req.Method) { // 基本配置：oneNET校验第三方接口
		msg, err_1 := req.Form["msg"]
		// signature, err_2 := req.Form["signature"]
		// nonce, err_3 := req.Form["nonce"]
		if(!err_1) { // 存在则返回msg
			fmt.Fprintf(res, msg[0])
		}
	} else if ("POST" == req.Method) { // 接收OneNET推送过来的数据
		// fetch job
		work := Job{serload: Serload{pri:"Just do it"}}
		JobQueue <- work
		fmt.Fprintf(res, "Hello World ...post")
	}
}*/

func init() {
	runtime.GOMAXPROCS(MaxWorker)
	JobQueue = make(chan Job, MaxQueue)
	dispatcher := NewDispatcher(MaxWorker)
	dispatcher.Run()
}
