package util

import (
	"../entity"
	"../log"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ZZMarquis/gm/sm2"
	"strings"
)

// 若能查询出随机数 说明为亿速码加密的数据
func AddYisumaRandomSign(head entity.Header, pri string, random string) (respBody string, err error) {
	//1. 解析json字符串
	var randomUnSign entity.YisumaRandomSign
	if err := json.Unmarshal([]byte(pri), &randomUnSign); err != nil {
		log.Error("ProcessAppMsg json.Unmarshal Header error, err=", err)
		return "", err
	}

	//2. 将HEX私钥字符串转为byte
	str := "607EC530749978DD8D32123B3F2FDF423D1632E6281EB83D083B6375109BB740"
	data, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}
	privateKey, e := sm2.RawBytesToPrivateKey(data)
	if e != nil {
		fmt.Println(*privateKey)
	}

	//3. 用私钥加密msg
	msg := "541DD0B0CA2F780B9DFB0C3527632789"
	r, s, err := sm2.SignToRS(privateKey, nil, []byte(msg))
	signatureR := hex.EncodeToString(r.Bytes())
	signatureS := hex.EncodeToString(s.Bytes())
	signature := strings.ToUpper(signatureR + signatureS)
	randomSign := entity.YisumaRandomSign{head.Cmd, head.Ack, head.DevType, head.DevId, head.Vendor, head.SeqId, randomUnSign.Password, randomUnSign.Password2, random, signature}
	randomSignStr, err := json.Marshal(randomSign)
	if err != nil {
		log.Error("Get YisumaRandom json.Marshal failed, err=", err)
		return "", err
	}
	return string(randomSignStr), err
}
