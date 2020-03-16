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
    floorHeatDevStatus = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":6,"aid":0,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatlockStatus = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":0,"aid":18,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatMaxTemp = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":6,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`
	floorHeatMinTemp = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233586","deviceuid":85390,"cid":513,"aid":5,"value":"0100","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","deviceid":769,"zonetype":1,"uptime":1581527097545}]}`


	infrared = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":13,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	door     = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":21,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	smoke    = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":40,"cid":1280,"aid":128,"value":"1000","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	flood    = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":42,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	gas      = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":43,"cid":1280,"aid":128,"value":"0100","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	sos      = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":1026,"zonetype":44,"cid":1280,"aid":128,"value":"0200","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"00158d0003e8b2e3_01","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	temp = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":770,"zonetype":1,"cid":1026,"aid":0,"value":"2817","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"test0","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
	illumi = `{"code":2,"status":"report","ver":"2","records":[{"deviceid":262,"zonetype":1,"cid":1024,"aid":0,"value":"28","bindid":"5233586","orgdata":"700f8e4d010000010c414205ab000b41e7","pushstring":"","uuid":"test0","snid":"FZD56","devicetype":"0x30b0001","uptime":1581527097545}]}`
    gsss = `{"code":2,"status":"report","ver":"2","records":[{"bindid":"5233517","deviceuid":88232,"cid":1280,"aid":128,"value":"1100","orgdata":"720ba858010005018000211100","pushstring":"气体传感器 燃气泄漏","uuid":"00158d000400efe2_01","snid":"FNB54-GAS07ML0.8","devicetype":"0x402002b","deviceid":1026,"zonetype":43,"uptime":1584346269975}]}`
)

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
	{"floorHeatDevStatus", floorHeatDevStatus},
	{"floorHeatlockStatus", floorHeatlockStatus},
	{"floorHeatMaxTemp", floorHeatMaxTemp},
	{"floorHeatMinTemp", floorHeatMinTemp},

	{"infrared", infrared},
	{"door", door},
	{"smoke", smoke},
	{"flood", flood},
	{"gas", gas},
	{"sos", sos},
	{"temp", temp},
	{"huminity", illumi},
}

func TestAlarmHandle(t *testing.T) {
	for _, ts := range tests {
		if ProcessFeibeeMsg([]byte(ts.msgValue)) != nil {
			t.Errorf("Process %s error", ts.msgName)
		}
	}
}

func TestOneAlarm(t *testing.T) {
	if ProcessFeibeeMsg([]byte(gsss)) != nil {
		t.Errorf("Process feibee alarm error")
	}
}


func BenchmarkProcessFeibeeMsg(b *testing.B) {
	for i:=0;i<b.N;i++ {
		for _, ts := range tests {
			ProcessFeibeeMsg([]byte(ts.msgValue))
		}
	}
}