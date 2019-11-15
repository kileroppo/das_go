package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dlintw/goconf"

	"./aliIoT2srv"
	"./andlink2srv"
	"./core/log"
	"./core/rabbitmq"
	"./core/redis"
	"./dindingtask"
	"./feibee2srv"
	"./onenet2srv"
	"./rmq/consumer"
	"./telecom2srv"
	"./wifi2srv"
	"runtime/pprof"
)

func loadCpuProfile() *os.File {
	cpuProfile := flag.String("cpuprofile", "./cpu", "record the cpu profile to file")
	if *cpuProfile == "" {
		panic("cpu profile created error")
	}

	f, err := os.Create(*cpuProfile)
	if err != nil {
		panic(err)
	}

	return f
}

func main() {
    f := loadCpuProfile()
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

	//1. 加载配置文件
	conf := loadConfig()

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

	//13. 启动http/https服务
	telecom2srv := telecom2srv.Telecom2HttpSrvStart(conf)

	//14. 启动http/https服务
	andlink2srv := andlink2srv.Andlink2HttpSrvStart(conf)

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
	//停止ali消息接收
	aliIOTsrv.Shutdown()

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

	// 18. 停止HTTP服务器
	if err := telecom2srv.Shutdown(nil); err != nil {
		log.Error("telecom2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	// 19. 停止HTTP服务器
	if err := andlink2srv.Shutdown(nil); err != nil {
		log.Error("andlink2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	// 20. 停止HTTP服务器
	if err := feibee2srv.Shutdown(nil); err != nil {
		log.Error("feibee2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	// 21. 停止定时器
	dindingtask.StopMyTimer()

	time.Sleep(1 * time.Second)

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

func loadConfig() *goconf.ConfigFile {
	conf_file := flag.String("config", "./das.ini", "设置配置文件.")
	flag.Parse()
	conf, err := goconf.ReadConfigFile(*conf_file)
	if err != nil {
		log.Errorf("加载配置文件失败，无法打开%q，%s\n", conf_file, err)
		os.Exit(1)
	}
	return conf
}
