package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"das/core/etcd"
	"das/core/log"
	"das/core/mysql"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/feibee2srv"
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
	log.Init()
	redis.InitRedis()
	mysql.Init()
	etcd.Init()
	rabbitmq.Init()
	procLock.Run()
	feibee2srv.Init()
	//aliSrv := aliIot2srv.NewAliIOT2Srv(conf)
	//aliSrv.Run()
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart()
	http2srv.Init()
	mqtt2srv.Init()
	tuya2srv.Init()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

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

	//aliSrv.Close()
	procLock.Close()
	feibee2srv.Close()
	rabbitmq.Close()
	http2srv.Close()
	mqtt2srv.Close()
	tuya2srv.Close()
	oneNet2Srv.Close()
	redis.Close()
	mysql.Close()
	etcd.CloseEtcdCli()
	log.Info("das_go server quit......")
}
