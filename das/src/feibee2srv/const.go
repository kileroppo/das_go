package feibee2srv

import "das/core/constant"

//飞比传感器类型(deviceId<<16 + zoneType)
const (
	voltage          = 0x00010020
	battery          = 0x00010021
	lowVoltage       = 0x0001003e
	lowPower         = 0x00010035
	sensorAlarm      = 0x05000080
	temperature      = 0x04020000
	humidity         = 0x04050000
	illuminance      = 0x04000000
	illumination     = 0x0201000a
	disinfection     = 0x0201000b
	disinfectionTime = 0x0201000c
	motorOperation   = 0x0201000d
	drying           = 0x02020002
	airDrying        = 0x02020003
	dryingTime       = 0x02020004
	airDryingTime    = 0x02020005

	mode               = 0x0201001c
	windspeed          = 0x02020000
	localTemperature   = 0x02010000
	currentTemperature = 0x02010011
	lockStatus         = 0x00000012
	maxTemperature     = 0x02010006
	minTemperature     = 0x02010005
	devStatus          = 0x00060000
)

var (
	alarmMsgTyp = map[int]parseFunc{
		voltage:            parseVoltageVal,
		battery:            parseBatteryVal,
		sensorAlarm:        parseSensorVal,
		temperature:        parseTempAndHuminityVal,
		humidity:           parseTempAndHuminityVal,
		localTemperature:   parseTempAndHuminityVal,
		currentTemperature: parseTempAndHuminityVal,
		maxTemperature:     parseTempAndHuminityVal,
		minTemperature:     parseTempAndHuminityVal,
		illumination:       parseFixVal,
		disinfection:       parseFixVal,
		motorOperation:     parseFixVal,
		drying:             parseFixVal,
		airDrying:          parseFixVal,
		mode:               parseFixVal,
		windspeed:          parseFixVal,
		devStatus:          parseFixVal,
		lockStatus:         parseFixVal,
		illuminance:        parseContinuousVal,
		disinfectionTime:   parseContinuousVal,
		dryingTime:         parseContinuousVal,
		airDryingTime:      parseContinuousVal,
	}

	fixAlarmName = map[MsgType]string{
		InfraredSensor:     constant.Wonly_Status_Sensor_Infrared,
		DoorMagneticSensor: constant.Wonly_Status_Sensor_Doorcontact,
		SmokeSensor:        constant.Wonly_Status_Sensor_Smoke,
		FloodSensor:        constant.Wonly_Status_Sensor_Flood,
		GasSensor:          constant.Wonly_Status_Sensor_Gas,
		SosBtnSensor:       constant.Wonly_Status_Sensor_SOSButton,
	}

	varAlarmName = map[int]string{
		voltage:     constant.Wonly_Status_Low_Voltage,
		battery:     constant.Wonly_Status_Low_Power,
		temperature: constant.Wonly_Status_Sensor_Temperature,
		humidity:    constant.Wonly_Status_Sensor_Humidity,
		illuminance: constant.Wonly_Status_Sensor_Illuminance,

		illumination:     constant.Wonly_Status_Airer_Illumination,
		disinfection:     constant.Wonly_Status_Airer_Disinfection,
		disinfectionTime: constant.Wonly_Status_Airer_Disinfection_Time,
		motorOperation:   constant.Wonly_Status_Airer_MotorOperation,
		drying:           constant.Wonly_Status_Airer_Drying,
		airDrying:        constant.Wonly_Status_Airer_Air_Drying,
		dryingTime:       constant.Wonly_Status_Airer_Drying_Time,
		airDryingTime:    constant.Wonly_Status_Airer_Air_Drying_Time,

		mode:               constant.Wonly_Status_Aircondition_Mode,
		windspeed:          constant.Wonly_Status_Aircondition_Windspeed,
		localTemperature:   constant.Wonly_Status_Aircondition_Local_Temperature,
		currentTemperature: constant.Wonly_Status_Aircondition_Curr_Temperature,

		lockStatus:     constant.Wonly_Status_FbLock_Status,
		maxTemperature: constant.Wonly_Status_Aircondition_Max_Temperature,
		minTemperature: constant.Wonly_Status_Aircondition_Min_Temperature,
		devStatus:      constant.Wonly_Status_FbDev_Status,
	}

	alarmValueMapByTyp = map[MsgType]([]string){
		InfraredSensor:     constant.Wonly_Sensor_Vals_Infrared,
		DoorMagneticSensor: constant.Wonly_Sensor_Vals_Doorcontact,
		SmokeSensor:        constant.Wonly_Sensor_Vals_Smoke,
		FloodSensor:        constant.Wonly_Sensor_Vals_Flood,
		GasSensor:          constant.Wonly_Sensor_Vals_Gas,
		SosBtnSensor:       constant.Wonly_Sensor_Vals_SOSButton,
	}

	alarmValueMapByCid = map[int]([]string){
		illumination:   constant.Wonly_FbAirer_Vals_Illumination,
		disinfection:   constant.Wonly_FbAirer_Vals_Disinfection,
		motorOperation: constant.Wonly_FbAirer_Vals_Motor_Operation,
		drying:         constant.Wonly_FbAirer_Vals_Drying,
		airDrying:      constant.Wonly_FbAirer_Vals_Air_Drying,
		mode:           constant.Wonly_FbAirer_Vals_Mode,
		windspeed:      constant.Wonly_FbAirer_Vals_Windspeed,
		devStatus:      constant.Wonly_FbDev_Vals_Status,
		lockStatus:     constant.Wonly_FbLock_Vals_Status,
	}
)
