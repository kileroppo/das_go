package main

import (
	aliIot2srv "das/aliIoT2srv"
	"das/paddoor2srv"
	xm2srv2 "das/xm2srv"
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
)

func main() {
	go func() {
		http.ListenAndServe(":14999", nil)
	}()

	//1. log初始
	conf := log.Init()

	//2. 初始化Redis连接池
	redis.InitRedisPool(conf)

	//3. 初始化rabbitmq
	rabbitmq.Init(conf)

	//4. 接收app消息
	go consumer.Run()

	//5. 初始化平板消费者交换器，消息队列的参数
	go paddoor2srv.Run()

	//6. 启动ali IOT推送接收服务
	aliSrv := aliIot2srv.NewAliIOT2Srv(conf)
	aliSrv.Run()

	//7. 启动http/https服务
	oneNet2Srv := onenet2srv.OneNET2HttpSrvStart(conf)

	//8. 启动http/https服务
	feibee2srv := feibee2srv.Feibee2HttpSrvStart(conf)

	//8. 启动雄迈告警消息接收
	xm2srv := xm2srv2.XM2HttpSrvStart(conf)

	//9. 启动MQTT
	// mqtt2srv.MqttInit(conf)	// 订阅接收端
	// mqtt.MqttInit(conf)		// 发布端

	//10. Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	//11. 信号处理
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

	//12. 关闭阿里云IOT推送接收服务
	aliSrv.Close()

	//13. 停止接收平板消息
	paddoor2srv.Close()

	//14. 停止接收app消息
	consumer.Close()

	//15. 停止rabbitmq连接
	rabbitmq.Close()

	//16. 停止OneNETHTTP服务器
	if err := oneNet2Srv.Shutdown(nil); err != nil {
		log.Error("oneNet2Srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	//17. 停止飞比HTTP服务器
	if err := feibee2srv.Shutdown(nil); err != nil {
		log.Error("feibee2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	//18. 停止雄迈HTTP服务器
	if err := xm2srv.Shutdown(nil); err != nil {
		log.Error("xm2srv.Shutdown failed, err=", err)
		// panic(err) // failure/timeout shutting down the server gracefully
	}

	//19. 断开MQTT连接
	// mqtt2srv.MqttRelease()
	// mqtt.MqttRelease()

	//20. 关闭redis
	redis.CloseRedisCli()

	log.Info("das_go server quit......")
}
