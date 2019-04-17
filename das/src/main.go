package main

import (
	"./core/log"
	"./core/rabbitmq"
	"./core/redis"
	"./dindingtask"
	"./httpJob"
	"./mq/consumer"
	"./mq/producer"
	"flag"
	"github.com/dlintw/goconf"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	//1. 加载配置文件
	conf := loadConfig()

	//2. 初始化日志
	initLogger(conf)

	//3. 初始化Redis
	redis.InitRedisSingle(conf)

	//4. 初始化生产者rabbitmq_uri
	rabbitmq.InitProducerMqConnection(conf)
	rabbitmq.InitProducerMqConnection2Db(conf)

	//5. 初始化消费者rabbitmq_uri
	rabbitmq.InitConsumerMqConnection(conf)

	//6. 初始化到APP生成者交换器，消息队列的参数
	producer.InitRmq_Ex_Que_Name(conf)

	//7. 初始化到Mongodb生成者交换器，消息队列的参数
	producer.InitRmq_Ex_Que_Name_mongo(conf)

	//8. 初始化消费者交换器，消息队列的参数
	consumer.InitRmq_Ex_Que_Name(conf)
	go consumer.ReceiveMQMsgFromAPP()

	// 9. 启动定时器
	dindingtask.InitTimer_IsStart(conf)
	dindingtask.StartMyTimer()

	//10. 启动http/https服务
	srv := httpServerStart(conf)

	//11. Handle SIGINT and SIGTERM.
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

	// 12. 停止HTTP服务器
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	// 13. 停止定时器
	dindingtask.StopMyTimer()

	log.Info("das_go server quit......")
}

func httpServerStart(conf *goconf.ConfigFile) *http.Server {
	// hanlder
	/*http.HandleFunc("/", httpJob.Entry)

	//判断是否为https协议
	isHttps, err := conf.GetBool("https", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	} else {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			httpsPort, _ := conf.GetInt("https", "https_port")
			log.Debug(http.ListenAndServeTLS(":"+strconv.Itoa(httpsPort), serverCrt, serverKey, nil))
		} else {
			httpPort, _ := conf.GetInt("http", "http_port")
			log.Debug("httpServerStart http.ListenAndServe()......")
			http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
		}
	}*/
	// 判断是否为https协议
	var httpPort int

	// 判断是否为https协议
	isHttps, err := conf.GetBool("https", "is_https")
	if err != nil {
		log.Errorf("读取https配置失败，%s\n", err)
		os.Exit(1)
	}
	if isHttps {
		httpPort, _ = conf.GetInt("https", "https_port")
	} else {
		httpPort, _ = conf.GetInt("http", "http_port")
	}

	srv := &http.Server{Addr: ":"+strconv.Itoa(httpPort)}

	http.HandleFunc("/", httpJob.Entry)

	go func() {
		if isHttps { //如果为https协议需要配置server.crt和server.key
			serverCrt, _ := conf.GetString("https", "https_server_crt")
			serverKey, _ := conf.GetString("https", "https_server_key")
			if err_https := srv.ListenAndServeTLS(serverCrt, serverKey); err_https != nil {
				log.Error("Httpserver: ListenAndServeTLS(): %s", err_https)
			}
		} else {
			log.Debug("httpServerStart http.ListenAndServe()......")
			if err_http := srv.ListenAndServe(); err_http != nil {
				// cannot panic, because this probably is an intentional close
				log.Error("Httpserver: ListenAndServe(): %s", err_http)
			}
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
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