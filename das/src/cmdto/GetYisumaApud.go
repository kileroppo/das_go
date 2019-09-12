package cmdto

import (
	"../core/httpgo"
	"../core/entity"
)

func GetYisumaApud(reqBody entity.YisumaHttpsReq) (respBody string, err error) {
	return httpgo.Http2YisumaActive(reqBody)
}