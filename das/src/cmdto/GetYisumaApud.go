package cmdto

import (
	"das/core/entity"
	"das/core/httpgo"
)

func GetYisumaApud(reqBody entity.YisumaHttpsReq) (respBody string, err error) {
	return httpgo.Http2YisumaActive(reqBody)
}
