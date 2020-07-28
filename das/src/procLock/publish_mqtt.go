package procLock

import "das/core/log"

func WlMqttPublish(uuid string, data []byte) error {
	if token := mqttCli.Publish(topic2Dev+uuid, 0, false, data); token.Wait() && token.Error() != nil {
		log.Error("WlMqttPublish failed, err: ", token.Error())
		return token.Error()
	}

	return nil
}

func WlMqttPublishPad(uuid string, data string) error {
	if token := mqttCli.Publish(topic2Pad+uuid, 0, false, data); token.Wait() && token.Error() != nil {
		log.Error("WlMqttPublishPad failed, err: ", token.Error())
		return token.Error()
	}

	return nil
}