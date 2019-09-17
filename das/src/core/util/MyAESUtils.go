/**
 * AES加密，包含填充 支持CBC加密 ECB加密
 *
 * @author jhhe66
 *
 * Copyright(c) 2019
 *
 */
package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"../aes/ecb"
)

/*	CBC加密 按照golang标准库的例子代码
	不过里面没有填充的部分,所以补上

	ECB加密
*/

//使用PKCS7进行填充，IOS也是7
func pKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil
	}
	return origData[:(length - unpadding)]
}

//aes加密，填充秘钥key的16位，24,32分别对应AES-128, AES-192, or AES-256.
func aesCBCEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	//填充原文
	blockSize := block.BlockSize()
	rawData = pKCS7Padding(rawData, blockSize)
	//初始向量IV必须是唯一，但不需要保密
	cipherText := make([]byte, blockSize+len(rawData))
	//block大小 16
	iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	//block大小和初始向量大小一定要一致
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[blockSize:], rawData)

	return cipherText, nil
}

//aes加密，填充秘钥key的16位，24,32分别对应AES-128, AES-192, or AES-256.
func aesECBEncrypt(rawData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	//填充原文
	blockSize := block.BlockSize()
	rawData = pKCS7Padding(rawData, blockSize)

	//初始向量IV必须是唯一，但不需要保密
	cipherText := make([]byte, len(rawData))
	//block大小 16
	/*iv := cipherText[:blockSize]
	if _, err := io.ReadFull(rand.Reader,iv); err != nil {
		panic(err)
	}*/

	//block大小和初始向量大小一定要一致
	// mode := cipher.NewCBCEncrypter(block,iv)
	mode := ecb.NewECBEncrypter(block)

	mode.CryptBlocks(cipherText, rawData)

	return cipherText, nil
}

func aesCBCDecrypt(encryptData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()

	if len(encryptData) < blockSize {
		panic("ciphertext too short")
	}
	iv := encryptData[:blockSize]
	encryptData = encryptData[blockSize:]

	// CBC mode always works in whole blocks.
	if len(encryptData)%blockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encryptData, encryptData)
	//解填充
	encryptData = pKCS7UnPadding(encryptData)
	return encryptData, nil
}

func aesECBDecrypt(encryptData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()

	if len(encryptData) < blockSize {
		panic("ciphertext too short")
	}
	// iv := encryptData[:blockSize]
	// encryptData = encryptData[blockSize:]

	// CBC mode always works in whole blocks.
	if len(encryptData)%blockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := ecb.NewECBDecrypter(block)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(encryptData, encryptData)

	//解填充
	encryptData = pKCS7UnPadding(encryptData)
	if nil == encryptData {
		return nil, errors.New("pKCS7UnPadding error.")
	}

	return encryptData, nil
}

func CBCEncrypt(rawData, key []byte) (string, error) {
	data, err := aesCBCEncrypt(rawData, key)
	fmt.Printf("%02x\n", data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
	// return hex.EncodeToString(data), nil
}

func ECBEncrypt(rawData, key []byte) (string, error) {
	data, err := aesECBEncrypt(rawData, key)
	// fmt.Printf("%02x\n", data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func CBCDecrypt(rawData string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(rawData)
	// data, err := hex.DecodeString(rawData)
	if err != nil {
		return "", err
	}
	dnData, err := aesCBCDecrypt(data, key)
	if err != nil {
		return "", err
	}
	return string(dnData), nil
}

func ECBDecrypt(rawData string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(rawData)
	// data, err := hex.DecodeString(rawData)
	if err != nil {
		return "", err
	}

	dnData, err := aesECBDecrypt(data, key)
	if err != nil {
		return "", err
	}
	return string(dnData), nil
}
