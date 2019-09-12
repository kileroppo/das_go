package feibee2srv

import (
	"flag"
	"os"

	"github.com/dlintw/goconf"

	"../core/log"
	"../core/rabbitmq"
	"../mq/consumer"
	"../mq/producer"
)

func RunFeibeeSrv() {
	conf := loadConfig()

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

	Feibee2HttpSrvStart(conf)
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

func initLogger(conf *goconf.ConfigFile) {
	logPath, err := conf.GetString("server", "log_path")
	logLevel, err := conf.GetString("server", "log_level")
	if err != nil {
		log.Errorf("日志文件配置有误, %s\n", err)
		os.Exit(1)
	}
	log.NewLogger(logPath, logLevel)
}
