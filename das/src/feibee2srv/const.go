package feibee2srv

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
		InfraredSensor:     "infrared",
		DoorMagneticSensor: "doorContact",
		SmokeSensor:        "smoke",
		FloodSensor:        "flood",
		GasSensor:          "gas",
		SosBtnSensor:       "sosButton",
	}

	varAlarmName = map[int]string{
		voltage:            "lowVoltage",
		battery:            "lowPower",
		temperature:        "temperature",
		humidity:           "humidity",
		illuminance:        "illuminance",
		illumination:       "illumination",
		disinfection:       "disinfection",
		disinfectionTime:   "disinfectionTime",
		motorOperation:     "motorOperation",
		drying:             "drying",
		airDrying:          "airDrying",
		dryingTime:         "dryingTime",
		airDryingTime:      "airDryingTime",
		mode:               "mode",
		windspeed:          "windspeed",
		localTemperature:   "localTemperature",
		currentTemperature: "currentTemperature",
		lockStatus:         "lockStatus",
		maxTemperature:     "maxTemperature",
		minTemperature:     "minTemperature",
		devStatus:          "devStatus",
	}

	alarmValueMapByTyp = map[MsgType]([]string){
		InfraredSensor:     []string{"无人", "有人"},
		DoorMagneticSensor: []string{"关闭", "开启"},
		SmokeSensor:        []string{"无烟", "有烟"},
		FloodSensor:        []string{"无水", "有水"},
		GasSensor:          []string{"无气体", "有气体"},
		SosBtnSensor:       []string{"正常", "报警"},
	}

	alarmValueMapByCid = map[int]([]string){
		illumination:   []string{"关闭", "开启"},
		disinfection:   []string{"关闭", "开启"},
		motorOperation: []string{"正常", "上限位", "下限位"},
		drying:         []string{"关闭", "开启"},
		airDrying:      []string{"关闭", "开启"},
		mode:           []string{"关闭", "", "", "制冷", "制热", "打开"},
		windspeed:      []string{"关闭", "低速", "中速", "高速", "", "自动"},
		devStatus:      []string{"关闭", "开启"},
		lockStatus:     []string{"锁定", "解锁"},
	}
)
