package filter

import (
	"das/core/constant"
	"math"
	"strconv"
)

var (
	AlarmDataFilterMap = map[string]struct{}{
		constant.Wonly_Status_Sensor_Gas:          {},
		constant.Wonly_Status_Sensor_Smoke:        {},
		constant.Wonly_Status_Sensor_Doorcontact:  {},
		constant.Wonly_Status_Sensor_Infrared:     {},
		constant.Wonly_Status_Sensor_Forced_Break: {},
		constant.Wonly_Status_Sensor_Flood:        {},
		constant.Wonly_Status_Audible_Alarm:       {},
		constant.Wonly_Status_Low_Power:           {},
	}

	EnvAlarmDataFilterMap = map[string] struct{} {
		constant.Wonly_Status_Sensor_PM25: {},
		constant.Wonly_Status_Sensor_CO2: {},
		constant.Wonly_Status_Sensor_Formaldehyde: {},
		constant.Wonly_Status_Sensor_VOC: {},
	}

	LimitPH2_5 = []float64{101, 151, 201, math.MaxFloat64}
	LimitCO2 = []float64{0.1, 0.2, 0.5, math.MaxFloat64}
	LimitCH2O = 0.1
	LimitVOC = 0.6

	GradeEnv = []string{"0", "A","B","C"}

	ReflectGrade = map[string] []float64 {
		constant.Wonly_Status_Sensor_PM25: LimitPH2_5,
		constant.Wonly_Status_Sensor_CO2: LimitCO2,
	}

	ReflectLimit = map[string]float64 {
		constant.Wonly_Status_Sensor_Formaldehyde: LimitCH2O,
		constant.Wonly_Status_Sensor_VOC: LimitVOC,
	}
)

func SensorFilter(devId, sensorType, sensorVal string, val interface{}) (notifyFlag bool, triggerFlag bool) {
	// pm2.5 三个过滤等级 0 - 100  100 - 200  val = 涂鸦返回的value
	notifyFlag, triggerFlag = true, false
	if !sensorMsgFilter(devId, sensorType, sensorVal, val) {
		if sensorType == constant.Wonly_Status_Sensor_Infrared {
			notifyFlag = false
			triggerFlag = true
		} else {
			if _,ok := EnvAlarmDataFilterMap[sensorType]; ok {
				notifyFlag = true
			} else {
				notifyFlag = false
			}
		}
	} else {
		notifyFlag, triggerFlag = true, true
	}
	return
}

func sensorMsgFilter(devId, code, sensorVal string, val interface{}) bool {
	key := devId
	// 如果 属于报警类型列表  devId_code
	if _,ok := AlarmDataFilterMap[code]; ok {
		key += "_" + code
	} else {
		// 否则  或是 等级设备等级
		level,ok := GetEnvSensorLevel(code, sensorVal)
		if ok {
			key += "_" + code
			val = level
		} else {
			return true
		}
	}
	 // key = dev_code   val = value of ty api return
	return AlarmMsgFilter(key, val, -1)
}

func GetEnvSensorLevel(sensorType, sensorVal string) (levelVal string, ok bool)  {
	_, ok = ReflectGrade[sensorType]
	if ok {
		levelVal = envSensorLevelTable(sensorType, sensorVal)
	} else {
		levelVal = envSensorVOCLevelTable(sensorType, sensorVal)
	}
	if len(levelVal) == 0 {
		ok = false
	} else {
		ok = true
	}
	return
}

func envSensorLevelTable(sensorType, sensorVal string) (levelVal string) {
	rangeLimit,ok := ReflectGrade[sensorType]
	if !ok {
		return
	}
	val,err := strconv.ParseFloat(sensorVal, 64)
	if err != nil {
		return
	}

	level := 0
	for i := 0; i < len(rangeLimit); i++ {
		if val < rangeLimit[i] {
			break
		} else {
			level++
		}
	}
	return GradeEnv[level]
}

func envSensorVOCLevelTable(sensorType, sensorVal string) (levelVal string) {
	limit,ok := ReflectLimit[sensorType]
	if !ok {
		return
	}
	val,err := strconv.ParseFloat(sensorVal, 64)
	if err != nil {
		return
	}
	if val < limit {
		return GradeEnv[0]
	} else {
		return GradeEnv[1]
	}
}

