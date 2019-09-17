package httpgo

import (
	"github.com/dlintw/goconf"
)

var OneNET_Url string
var OneNET_Apikey string

func InitOneNETConfig(conf *goconf.ConfigFile) (err error) {
	var errs error
	OneNET_Url, errs = conf.GetString("onenet2http", "onenet_url")
	if nil != errs {
		return errs
	}

	OneNET_Apikey, errs = conf.GetString("onenet2http", "onenet_apikey")
	if nil != errs {
		return errs
	}

	return errs
}
