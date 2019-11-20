package httpgo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"

	"github.com/json-iterator/go"

	"../entity"
	"../log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Http2FeibeeWonlyGuard(appData string) {
	var msg entity.WonlyGuardMsgFromApp
	if err := json.Unmarshal([]byte(appData), &msg); err != nil {
		log.Warning("Http2FeibeeWonlyGuard json.Unmarshal() error = ", err)
		return
	}

	var reqMsg entity.Req2Feibee

	reqMsg.Act = "controlstate"
	reqMsg.Code = "220"
	reqMsg.Bindid = msg.Bindid

	md5Ctx := md5.New()
	md5Ctx.Write([]byte("W" + msg.Devid + "only"))
	key := md5Ctx.Sum(nil)

	reqMsg.Bindstr = AESCBCDecrypt(msg.Bindstr, key)
	reqMsg.Ver = "2.0"
	reqMsg.Devs = append(reqMsg.Devs, entity.ReqDevInfo2Feibee{
		Uuid:  msg.Devid,
		Value: msg.Value,
	})

	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		log.Warning("Http2FeibeeWonlyGuard json.Marshal() error = ", err)
		return
	}

	respData, err := doHttpReq("POST", "https://dev.fbeecloud.com/devcontrol/", reqData)
	var respMsg entity.RespFromFeibee
	err = json.Unmarshal(respData, &respMsg)
	if err != nil {
		log.Warning("Control WonlyGuard failed")
		return
	}

	if respMsg.Code != 1 {
		log.Warning("Control WonlyGuard failed")
	} else {
		log.Info("Control WonlyGuard successfully")
	}
}

func AESCBCEncrypt(originData string, key []byte) string {
	originByte := []byte(originData)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()
	originByte = PKCS7Padding(originByte, blockSize)
	// ciphertext := make([]byte, blockSize+len(originByte))
	//设置初始化向量
	iv := key[:blockSize]
	blockMode := cipher.NewCBCEncrypter(block, iv)

	res := make([]byte, len(originByte))
	blockMode.CryptBlocks(res, originByte)

	return hex.EncodeToString(res)
}

func AESCBCDecrypt(ciphertext string, key []byte) string {
	cipherByte, _ := hex.DecodeString(ciphertext)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()
	iv := key[:blockSize]

	blockMode := cipher.NewCBCDecrypter(block, iv)

	res := make([]byte, len(cipherByte))
	blockMode.CryptBlocks(res, cipherByte)

	res = PKCS7UnPadding(res)
	return string(res)
}

//填充
func PKCS7Padding(ciphertext []byte, blocksize int) []byte {
	padding := blocksize - len(ciphertext)%blocksize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//取消填充
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func doHttpReq(method, url string, data []byte) (respData []byte, err error) {
	body := bytes.NewReader(data)
	httpClient := http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return respData, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}

	respData, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return
}
