package main

import (
	"./core/log"
	"github.com/dlintw/goconf"
	"flag"
	"os"
	"net/http"
	"strconv"
	"./httpJob"
	"./core/redis"
	"./core/rabbitmq"
	"./mq/consumer"
	"./mq/producer"
			)

func main() {
	// upgrade.GetUpgradeFileInfo("866971031002111", "WonlyNBLock", 0)
	// upgrade.TransferFileData("866971031002111", "WonlyNBLock", 0, 1, "mcu-v1.0.43.bin")
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

	//9. 启动http/https服务
	httpServerStart(conf)

	log.Info("quit")
}

func httpServerStart(conf *goconf.ConfigFile) {
	// hanlder
	http.HandleFunc("/", httpJob.Entry)

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
			log.Debug(http.ListenAndServe(":"+strconv.Itoa(httpPort), nil))
		}
	}
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