package redis

import (
	"das/core/log"
	"testing"
	"time"
)

var (
	reqData = `
{"act":"standardWriteAttribute","code":"286","AccessID":"v258ejzqsuumi4fnel0whofa0","key":"2r9b9l66oa3seebj9740z7vi5","bindid":"51668978","bindstr":"175d","ver":"2.0","devs":[{"uuid":"00158d0005807903_01","value":"AB0024003B"}]}
`
)

func init() {
	log.Init()
	InitRedis()
}

func TestIsFeibeeSpSrv(t *testing.T) {
	IsFeibeeSpSrv([]byte(reqData))
	redisDevPool0.HMSetNX(4, "msgFilter", "test1", true, time.Minute)
}
