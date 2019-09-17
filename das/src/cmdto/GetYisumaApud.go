package cmdto

import (
	"../core/entity"
	"../core/httpgo"
)

func GetYisumaApud(reqBody entity.YisumaHttpsReq) (respBody string, err error) {
	return httpgo.Http2YisumaActive(reqBody)
}
