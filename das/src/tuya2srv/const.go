package tuya2srv

import (
	"github.com/tidwall/gjson"

	"das/core/constant"
)

//涂鸦设备状态code枚举
const (
	Ty_Status_Electricity_Left    = "electricity_left"
	Ty_Status_Clean_Record        = "clean_record"
	Ty_Status_Power               = "power"
	Ty_Status                     = "status"
	Ty_Status_Doorcontact_State   = "doorcontact_state"
	Ty_Status_Gas_Sensor_State    = "gas_sensor_state"
	Ty_Status_Smoke_Sensor_Status = "smoke_sensor_status"
	Ty_Status_Watersensor_State   = "watersensor_state"
	Ty_Status_Temper_Alarm        = "temper_alarm"
	Ty_Status_Battery_Percentage  = "battery_percentage"
	Ty_Status_SOS_State           = "sos_state"

	Ty_Status_Air_Temperature = "TMP"
	Ty_Status_Air_Humidity    = "HUM"
	Ty_Status_Air_PM25_Value  = "PM2_5"
	Ty_Status_Air_CO_Value    = "CO"
	TY_Status_Air_CO2_Value   = "CO2"
	Ty_Status_Air_VOC         = "VOC"
	Ty_Status_Air_CH2O        = "CH2O"

	Ty_Status_Va_Temperature = "va_temperature"
	Ty_Status_Va_Humidity    = "va_humidity"

	Ty_Status_Bright_Value = "bright_value"

	Ty_Status_Presence_State = "presence_state"
	Ty_Status_Pir            = "pir"
	Ty_Status_Scene_1        = "scene_1"
	Ty_Status_Scene_2        = "scene_2"
	Ty_Status_Scene_3        = "scene_3"
	Ty_Status_Scene_4        = "scene_4"

	Ty_Status_Switch         = "switch"
	Ty_Status_Switch_1       = "switch_1"
)

//涂鸦设备事件bizCode
const (
	Ty_Event_Online         = "online"
	Ty_Event_Offline        = "offline"
	Ty_Event_Name_Update    = "nameUpdate"
	Ty_Event_Dp_Name_Update = "dpNameUpdate"
	Ty_Event_Bind_User      = "bindUser"
	Ty_Event_Delete         = "delete"
	Ty_Event_Upgrade_Status = "upgradeStatus"
)

//涂鸦传感器异常报警值
const (
	Ty_AlarmVal_Gas         = "1"
	Ty_AlarmVal_Smoke       = "alarm"
	Ty_AlarmVal_Watersensor = "1"
	Ty_AlarmVal_Pir         = "pir"
	Ty_AlarmVal_Presence    = "presence"
	Ty_AlarmVal_Doorcontact = "true"
	Ty_AlarmVal_Temper      = "true"
	Ty_AlarmVal_SOS         = "true"
)

type TyStatusHandle func(devId string, rawJsonData gjson.Result)
type TyEventHandle func(devId, tyEvent string, rawJsonData gjson.Result)

//涂鸦处理分类
var (
	TyDevStatusHandlers = map[string]TyStatusHandle{
		Ty_Status_Electricity_Left:   TyStatusRobotCleanerBattHandle,
		Ty_Status_Power:              TyStatusPowerHandle,
		Ty_Status:                    TyStatusNormalHandle,
		Ty_Status_Battery_Percentage: TyStatusDevBatt,
		Ty_Status_Clean_Record:       TyStatus2PMSHandle,
		Ty_Status_Switch:             TyStatus2PMSHandle,
		Ty_Status_Switch_1:           TyStatus2PMSHandle,

		Ty_Status_Gas_Sensor_State:    TyStatusAlarmSensorHandle,
		Ty_Status_Smoke_Sensor_Status: TyStatusAlarmSensorHandle,
		Ty_Status_Watersensor_State:   TyStatusAlarmSensorHandle,
		Ty_Status_Presence_State:      TyStatusAlarmSensorHandle,
		Ty_Status_Doorcontact_State:   TyStatusAlarmSensorHandle,
		Ty_Status_Temper_Alarm:        TyStatusAlarmSensorHandle,
		Ty_Status_Pir:                 TyStatusAlarmSensorHandle,
		Ty_Status_SOS_State:           TyStatusAlarmSensorHandle,

		Ty_Status_Va_Temperature: TyStatusEnvSensorHandle,
		Ty_Status_Va_Humidity:    TyStatusEnvSensorHandle,
		Ty_Status_Bright_Value:  TyStatusEnvSensorHandle,

		Ty_Status_Air_Temperature: TyStatusEnvSensorHandle,
		Ty_Status_Air_Humidity:    TyStatusEnvSensorHandle,
		Ty_Status_Air_PM25_Value:  TyStatusEnvSensorHandle,
		TY_Status_Air_CO2_Value:   TyStatusEnvSensorHandle,
		Ty_Status_Air_VOC:         TyStatusEnvSensorHandle,
		Ty_Status_Air_CH2O:        TyStatusEnvSensorHandle,

		Ty_Status_Scene_1: TyStatusSceneHandle,
		Ty_Status_Scene_2: TyStatusSceneHandle,
		Ty_Status_Scene_3: TyStatusSceneHandle,
		Ty_Status_Scene_4: TyStatusSceneHandle,
	}

	TyDevEventHandlers = map[string]TyEventHandle{
		Ty_Event_Online:  TyEventOnOffHandle,
		Ty_Event_Offline: TyEventOnOffHandle,
		Ty_Event_Delete:  TyEventDeleteHandle,
	}

	TySensor2WonlySensor = map[string]string{
		Ty_Status_Gas_Sensor_State:    constant.Wonly_Status_Sensor_Gas,
		Ty_Status_Smoke_Sensor_Status: constant.Wonly_Status_Sensor_Smoke,
		Ty_Status_Doorcontact_State:   constant.Wonly_Status_Sensor_Doorcontact,
		Ty_Status_Pir:                 constant.Wonly_Status_Sensor_Infrared,
		Ty_Status_Temper_Alarm:        constant.Wonly_Status_Sensor_Forced_Break,
		Ty_Status_Presence_State:      constant.Wonly_Status_Sensor_Infrared,
		Ty_Status_Watersensor_State:   constant.Wonly_Status_Sensor_Flood,
		Ty_Status_SOS_State:           constant.Wonly_Status_Sensor_SOSButton,

		Ty_Status_Air_Temperature: constant.Wonly_Status_Sensor_Temperature,
		Ty_Status_Air_Humidity:    constant.Wonly_Status_Sensor_Humidity,
		Ty_Status_Air_PM25_Value:  constant.Wonly_Status_Sensor_PM25,
		TY_Status_Air_CO2_Value:   constant.Wonly_Status_Sensor_CO2,
		Ty_Status_Air_CH2O:        constant.Wonly_Status_Sensor_Formaldehyde,

		Ty_Status_Va_Temperature: constant.Wonly_Status_Sensor_Temperature,
		Ty_Status_Va_Humidity:    constant.Wonly_Status_Sensor_Humidity,
		Ty_Status_Bright_Value:   constant.Wonly_Status_Sensor_Illuminance,
	}

	TyEnvSensorValDivisor = map[string]int{
		Ty_Status_Air_Temperature: 10,
		TY_Status_Air_CO2_Value:   100,
		Ty_Status_Air_CH2O:        100,
		Ty_Status_Va_Humidity:     100,
		Ty_Status_Va_Temperature:  100,
	}

	TySensorAlarmReflect = map[string]string{
		Ty_Status_Gas_Sensor_State:    Ty_AlarmVal_Gas,
		Ty_Status_Smoke_Sensor_Status: Ty_AlarmVal_Smoke,
		Ty_Status_Doorcontact_State:   Ty_AlarmVal_Doorcontact,
		Ty_Status_Pir:                 Ty_AlarmVal_Pir,
		Ty_Status_Temper_Alarm:        Ty_AlarmVal_Temper,
		Ty_Status_Watersensor_State:   Ty_AlarmVal_Watersensor,
		Ty_Status_Presence_State:      Ty_AlarmVal_Presence,
		Ty_Status_SOS_State:           Ty_AlarmVal_SOS,
	}
)
