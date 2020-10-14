package procLock

import (
	"errors"

	"github.com/modern-go/reflect2"

	"das/core/log"
)

var (
	ErrMqttCliNil = errors.New("Mqtt Client is nil")
)

func WlMqttPublish(uuid string, data []byte) error {
	if reflect2.IsNil(mqttCli) {
		return ErrMqttCliNil
	}

	if token := mqttCli.Publish(topic2Dev+uuid, 0, false, data); token.Wait() && token.Error() != nil {
		log.Error("WlMqttPublish failed, err: ", token.Error())
		return token.Error()
	}

	return nil
}

func WlMqttPublishPad(uuid string, data string) error {
	if reflect2.IsNil(mqttCli) {
		return ErrMqttCliNil
	}

	if token := mqttCli.Publish(topic2Pad+uuid, 0, false, data); token.Wait() && token.Error() != nil {
		log.Error("WlMqttPublishPad failed, err: ", token.Error())
		return token.Error()
	}

	return nil
}