package rabbitmq

import (
	"das/core/log"
	"das/core/redis"
	"fmt"
	"testing"
)

var (
	rawData = `count %d`
)

func init() {
	log.Init()
	redis.InitRedis()
	Init()
}

func TestPublish(t *testing.T) {
	for i:=0;i<100;i++ {
		data := []byte(fmt.Sprintf(rawData, i))
		Publish2pms(data, "")
		Publish2app(data, "test")
		Publish2dev(data, "test")
		Publish2mns(data, "")
		Publish2wonlyms(data, "")
	}
}