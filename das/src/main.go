package main

import (
	"flag"
	"os"

	"github.com/dlintw/goconf"

	"./core/log"
	"./onenet2srv"
	"os/signal"
	"syscall"
)

func main() {
	//1. 加载配置文件
	conf := loadConfig()

	//2. 初始化日志
	initLogger(conf)

	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)

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

	// 17. 停止HTTP服务器
	if err := oneNet2Srv.Shutdown(nil); err != nil {
		log.Error("oneNet2Srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
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
