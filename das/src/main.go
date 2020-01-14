package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"das/core/log"
	"das/core/rabbitmq"
	"das/core/redis"
	"das/feibee2srv"
	"das/onenet2srv"
	"das/rmq/consumer"
	"das/wifi2srv"
)

func main() {
	go func() {
		http.ListenAndServe(":14999", nil)
	}()

	conf := log.Conf

	//3. 初始化Redis连接池
	redis.InitRedisPool(conf)

	//4. 初始化rabbitmq
	rabbitmq.Init(conf)

	//接收app消息
	go consumer.Run()

	//10. 初始化平板消费者交换器，消息队列的参数
	go wifi2srv.Run()

	//11. 启动ali IOT推送接收服务
	// aliSrv := aliIot2srv.NewAliIOT2Srv(conf)
	// aliSrv.Run()

	//12. 启动http/https服务
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)

	// 15. 启动http/https服务
	feibee2srv := feibee2srv.Feibee2HttpSrvStart(conf)

	//16. Handle SIGINT and SIGTERM.
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
	// 关闭阿里云IOT推送接收服务
	// aliSrv.Close()

	//停止接收平板消息
	wifi2srv.Close()

	//停止接收app消息
	consumer.Close()

	//停止rabbitmq连接
	rabbitmq.Close()

	// 17. 停止HTTP服务器
	if err := oneNet2Srv.Shutdown(nil); err != nil {
		log.Error("oneNet2Srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	// 20. 停止HTTP服务器
	if err := feibee2srv.Shutdown(nil); err != nil {
		log.Error("feibee2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	log.Info("das_go server quit......")
}
