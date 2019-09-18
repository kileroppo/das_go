package httpgo

import (
	"github.com/dlintw/goconf"
)

var oneNET_Url string
var oneNET_Apikey string

func InitOneNETConfig(conf *goconf.ConfigFile) (err error) {
	var errs error
	oneNET_Url, errs = conf.GetString("onenet2http", "onenet_url")
	if nil != errs {
		return errs
	}

	oneNET_Apikey, errs = conf.GetString("onenet2http", "onenet_apikey")
	if nil != errs {
		return errs
	}

	return errs
}
