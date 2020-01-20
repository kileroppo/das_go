package wlprotocol

/*
*	定义两个接口
*	不同的包体解包
 */
type IPdu interface {
	Encode(uuid string) ([]byte, error)			// 打包包体，返回的是AES加密后的数据
	Decode(bBody []byte, uuid string) error		// 包体解包，AES解密
}

/*
*	定义两个接口
*	不同的协议解包
 */
type IPKG interface {
	PkEncode(pdu IPdu) ([]byte, error)			// 打包包体，返回的是AES加密后的数据
	PkDecode(pkg []byte) ([]byte, error)		// 包体解包，AES解密
}
