package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	aliIot2srv "das/aliIoT2srv"
	"das/core/etcd"
	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/feibee2srv"
	"das/http2srv"
	"das/onenet2srv"
	"das/procLock"
)

func main() {
	go func() {
		http.ListenAndServe(":14999", nil)
	}()
	conf := log.Init()

	redis.InitRedis()
	etcd.Init()
	rabbitmq.Init()
	procLock.Run()
	feibee2srv.Init()
	aliSrv := aliIot2srv.NewAliIOT2Srv(conf)
	aliSrv.Run()
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)
	http2srv.Init()
	//mqtt2srv.Init()
	//tuya2srv.Init()

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

	aliSrv.Close()
	procLock.Close()
	feibee2srv.Close()
	rabbitmq.Close()
	http2srv.Close()
	//mqtt2srv.Close()
	//tuya2srv.Close()
	oneNet2Srv.Close()
	redis.Close()
	etcd.CloseEtcdCli()
	log.Info("das_go server quit......")
}
