package main

import (
	"flag"
	"os"

	"github.com/dlintw/goconf"

	"./onenet2srv"
	"./core/log"
	"syscall"
	"os/signal"
)

func main()  {

	conf := loadConfig()

	//2. 初始化日志
	initLogger(conf)
	oneNetSrv := onenet2srv.OneNET2HttpSrvStart(conf)

	ch := make(chan os.Signal)

	signal.Notify(ch, os.Interrupt)

	for {
		switch <-ch {
		case syscall.SIGQUIT:
		    break
		default:
			break
		}
	}

	oneNetSrv.Shutdown(nil)


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