package feibee2srv

import (
	"testing"
)

var (
	airerLightStatus      = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":10,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerDisinfection     = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":11,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerDisinfectionTime = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":12,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerWorkStatus       = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":13,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerDryStatus        = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":514,"aid":2,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerAirDryStatus     = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":514,"aid":3,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerDryTime          = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":514,"aid":4,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`
	airerAirDryTime       = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":514,"aid":5,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":516,"zonetype":1,"uptime":1581527097545}]}`

	floorHeatMode      = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":28,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatLocalTemp = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":0,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatCurrTemp  = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":17,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatWindspeed = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":514,"aid":0,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`

	infrared = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":13,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	door     = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":21,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	smoke    = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":40,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	flood    = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":42,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	gas      = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":43,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	sos      = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":44,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
)

//func init() {
//	conf := log.Init()
//	rabbitmq.Init(conf)
//}

func TestAlarmHandle(t *testing.T) {
	var tests = []struct {
		msgName  string
		msgValue string
	}{
		{"airerLightStatus", airerLightStatus},
		{"airerDisinfection", airerDisinfection},
		{"airerDisinfectionTime", airerDisinfectionTime},
		{"airerDryStatus", airerDryStatus},
		{"airerAirDryStatus", airerAirDryStatus},
		{"airerDryTime", airerDryTime},
		{"airerAirDryTime", airerAirDryTime},
		{"airerWorkStatus", airerWorkStatus},
		{"floorHeatMode", floorHeatMode},
		{"floorHeatLocalTemp", floorHeatLocalTemp},
		{"floorHeatCurrTemp", floorHeatCurrTemp},
		{"floorHeatWindspeed", floorHeatWindspeed},
		{"infrared", infrared},
		{"door", door},
		{"smoke", smoke},
		{"flood", flood},
		{"gas", gas},
		{"sos", sos},
	}

	for _, ts := range tests {
		if ProcessFeibeeMsg([]byte(ts.msgValue)) != nil {
			t.Errorf("Process %s error", ts.msgName)
		}
	}
}
