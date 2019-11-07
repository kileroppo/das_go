package main

import (
	"./andlink2srv"
	"./core/log"
	"./core/rabbitmq"
	"./core/redis"
	"./dindingtask"
	"./feibee2srv"
	"./onenet2srv"
	"./rmq/consumer"
	"./rmq/producer"
	"./telecom2srv"
	"./wifi2srv"
	"./aliIoT2srv"
	"flag"
	"github.com/dlintw/goconf"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//1. 加载配置文件
	conf := loadConfig()

	//2. 初始化日志
	initLogger(conf)

	//3. 初始化Redis连接池
	redis.InitRedisPool(conf)

	//4. 初始化生产者rabbitmq_uri
	rabbitmq.InitProducerMqConnection(conf)
	rabbitmq.InitProducerMqConnection2Db(conf)
	rabbitmq.InitProducerMqConnection2Device(conf)

	//5. 初始化消费者rabbitmq_uri
	rabbitmq.InitConsumerMqConnection(conf)

	//6. 初始化到APP生成者交换器，消息队列的参数
	producer.InitRmq_Ex_Que_Name(conf)

	//7. 初始化到Mongodb生成者交换器，消息队列的参数
	producer.InitRmq_Ex_Que_Name_mongo(conf)

	//8. 初始化发送到平板设备的生成者交换器，消息队列的参数
	producer.InitRmq_Ex_Que_Name_Device(conf)

	//9. 初始化消费者交换器，消息队列的参数
	consumer.InitRmq_Ex_Que_Name(conf)
	go consumer.ReceiveMQMsgFromAPP()

	//10. 初始化平板消费者交换器，消息队列的参数
	wifi2srv.InitRmq_Ex_Que_Name(conf)
	go wifi2srv.ReceiveMQMsgFromDevice()

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

SERVER_EXIT:
	// 处理信号
	for {
		s := <-ch
		switch s {
		case syscall.SIGINT:
			log.Error("SIGINT: get signal: ", s)
			break SERVER_EXIT
		case syscall.SIGTERM:
			log.Error("SIGTERM: get signal: ", s)
			break SERVER_EXIT
		case syscall.SIGQUIT:
			log.Error("SIGSTOP: get signal: ", s)
			break SERVER_EXIT
		case syscall.SIGHUP:
			log.Error("SIGHUP: get signal: ", s)
			break SERVER_EXIT
		case syscall.SIGKILL:
			log.Error("SIGKILL: get signal: ", s)
			break SERVER_EXIT
		default:
			log.Error("default: get signal: ", s)
			break SERVER_EXIT
		}
	}
    //停止ali消息接收
	aliIOTsrv.Shutdown()

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

	time.Sleep(1*time.Second)

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
