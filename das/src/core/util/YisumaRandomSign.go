package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZZMarquis/gm/sm2"

	"das/core/entity"
	"das/core/log"
)

// 若能查询出随机数 说明为亿速码加密的数据
func AddYisumaRandomSign(head entity.Header, pri string, random string) (respBody string, err error) {
	//1. 解析json字符串
	var randomUnSign entity.YisumaRandomSign
	if err := json.Unmarshal([]byte(pri), &randomUnSign); err != nil {
		log.Error("ProcessAppMsg json.Unmarshal Header error, err=", err)
		return "", err
	}

	stringKey := strings.ToUpper(Md5(head.DevId))

	if len(randomUnSign.Password) > 6 {
		psw1, err_0 := hex.DecodeString(randomUnSign.Password)
		if err_0 != nil {
			log.Error("AddYisumaRandomSign DecodeString 0 failed, err=", err_0)
			return "", err_0
		}

		passwd1, err0 := ECBDecryptByte(psw1, []byte(stringKey))
		if err0 != nil {
			log.Error("AddYisumaRandomSign ECBDecryptByte 0 failed, err=", err0)
			return "",  err0
		}
		randomUnSign.Password = string(passwd1)
	}

	if len(randomUnSign.Password2) > 6 {
		psw2, err_1 := hex.DecodeString(randomUnSign.Password2)
		if err_1 != nil {
			log.Error("AddYisumaRandomSign DecodeString 1 failed, err=", err_1)
			return "",  err_1
		}
		passwd2, err1 := ECBDecryptByte(psw2, []byte(stringKey))
		if err1 != nil {
			log.Error("AddYisumaRandomSign ECBDecryptByte 1 failed, err=", err1)
			return "",  err1
		}
		randomUnSign.Password2 = string(passwd2)
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
