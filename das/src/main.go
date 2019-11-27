package main

import (
	"github.com/dlintw/goconf"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"./core/log"
	"./core/rabbitmq"
	"./core/redis"
	"./dindingtask"
	"./feibee2srv"
	"./onenet2srv"
	"./rmq/consumer"
	"./wifi2srv"
	"./aliIoT2srv"
)

func main() {
	go func() {
		http.ListenAndServe(":14999", nil)
	}()

	conf := log.Conf

	//2. 初始化日志
	initLogger(conf)

	//3. 初始化Redis连接池
	redis.InitRedisPool(conf)

	//4. 初始化rabbitmq
	rabbitmq.Init(conf)

	//接收app消息
	go consumer.Run()

	//10. 初始化平板消费者交换器，消息队列的参数
	go wifi2srv.Run()

	//11. 启动定时器
	dindingtask.InitTimer_IsStart(conf)
	dindingtask.StartMyTimer()

	//12. 启动http/https服务
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)

	// 15. 启动http/https服务
	feibee2srv := feibee2srv.Feibee2HttpSrvStart(conf)

	// 启动ali IOT推送接收服务
	aliIOTsrv := aliIot2srv.NewAliIOT2Srv(conf)
	aliIOTsrv.Run()

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
	aliIOTsrv.Close()

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

	// 21. 停止定时器
	dindingtask.StopMyTimer()

	log.Info("das_go server quit......")
}

func initLogger(conf *goconf.ConfigFile) {
	logPath, err := conf.GetString("server", "log_path")
	logLevel, err := conf.GetString("server", "log_level")
	if err != nil {
		log.Errorf("日志文件配置有误, %s\n", err)
		os.Exit(1)
	}
	log.NewLogger(logPath, logLevel)
}
