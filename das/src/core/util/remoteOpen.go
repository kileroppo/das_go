package util

import (
	"../entity"
	"../log"
	"../redis"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ZZMarquis/gm/sm2"
	"strings"
)

func RemoteOpen(head entity.Header, pri string) (respBody string, err error) {
	//若为远程开锁流程且能查到random，则需要加密
	//查询redis
	random, err0 := redis.GetDeviceYisumaRandomfromPool(head.DevId)
	if err0 != nil {
		log.Error("Get YisumaRandom from redis failed, err=", err0)
		return "", err0
	}
	//若能查询出随机数 说明为亿速码加密的数据
	if random != "" {
		str := "607EC530749978DD8D32123B3F2FDF423D1632E6281EB83D083B6375109BB740"
		data, err := hex.DecodeString(str)
		if err != nil {
			return "", err
		}
		privateKey, e := sm2.RawBytesToPrivateKey(data)
		if e != nil {
			fmt.Println(*privateKey)
		}
		//用私钥加密msg
		msg := "541DD0B0CA2F780B9DFB0C3527632789"
		r, s, err := sm2.SignToRS(privateKey, nil, []byte(msg))
		signatureR := hex.EncodeToString(r.Bytes())
		signatureS := hex.EncodeToString(s.Bytes())
		signature := strings.ToUpper(signatureR + signatureS)
		randomSign := entity.YisumaRandomSign{head.Cmd, head.Ack, head.DevType, head.DevId, head.Vendor, head.SeqId, random, signature}
		randomSignStr, err := json.Marshal(randomSign)
		if err != nil {
			log.Error("Get YisumaRandom json.Marshal failed, err=", err)
			return "", err
		}
		return string(randomSignStr), err
	}
	return pri, err
}
