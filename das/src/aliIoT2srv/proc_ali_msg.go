package aliIot2srv

import (
	"../core/log"
)

func ProcessAliMsg(data []byte, topic string) {
	log.Debugf("ali-topic: %s -> \n %s", topic , string(data))
}