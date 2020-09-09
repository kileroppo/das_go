package main

import (
	"das/feibee2srv"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"das/core/etcd"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/http2srv"
	"das/mqtt2srv"
	"das/onenet2srv"
	"das/procLock"
	"das/tuya2srv"
)

func main() {
	go func() {
		http.ListenAndServe(":14999", nil)
	}()

	//1. log初始
	conf := log.Init()

	//2. 初始化Redis连接池
	redis.InitRedis()
	etcd.Init()

	//3. 初始化rabbitmq
	rabbitmq.Init()

	//4. 接收app消息
	go procLock.Run()

	mqtt2srv.Init()

	//6. 启动ali IOT推送接收服务
	// aliSrv := aliIot2srv.NewAliIOT2Srv(conf)
	// aliSrv.Run()

	//7. 启动http/https服务
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)

	//8. 启动http/https服务
	//feibee2srv := feibee2srv.Feibee2HttpSrvStart(conf)

	//8. 启动雄迈告警消息接收
	http2srv.Http2SrvStart()

	go tuya2srv.Tuya2SrvStart()

	//10. Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	//11. 信号处理
	s := <-ch
	switch s {
	case syscall.SIGINT:
		log.Error("SIGINT: get signal: ", s)
	case syscall.SIGTERM:
		log.Error("SIGTERM: get signal: ", s)

	case syscall.SIGQUIT:
		log.Error("SIGSTOP: get signal: ", s)

	case syscall.SIGHUP:
		log.Error("SIGHUP: get signal: ", s)

	case syscall.SIGKILL:
		log.Error("SIGKILL: get signal: ", s)

	default:
		log.Error("default: get signal: ", s)
	}

	//12. 关闭阿里云IOT推送接收服务
	// aliSrv.Close()

	mqtt2srv.Close()

	//14. 停止接收app消息
	procLock.Close()

	//15. 停止rabbitmq连接
	rabbitmq.Close()

	tuya2srv.Close()

	http2srv.Close()

	feibee2srv.Close()

	//16. 停止OneNETHTTP服务器
	if err := oneNet2Srv.Shutdown(nil); err != nil {
		log.Error("oneNet2Srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	//20. 关闭redis
	redis.Close()
	etcd.CloseEtcdCli()

	log.Info("das_go server quit......")
}
