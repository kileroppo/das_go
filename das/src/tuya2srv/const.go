package tuya2srv

import "github.com/tidwall/gjson"

//王力传感器报警字符类型枚举
const (
	Wonly_Status_Sensor_Illuminance  = "illuminance"
	Wonly_Status_Sensor_Humidity     = "humidity"
	Wonly_Status_Sensor_Temperature  = "temperature"
	Wonly_Status_Sensor_PM25         = "PM2.5"
	Wonly_Status_Sensor_VOC          = "VOC"
	Wonly_Status_Sensor_Formaldehyde = "formaldehyde"
	Wonly_Status_Sensor_CO2          = "CO2"
	Wonly_Status_Sensor_Gas          = "gas"
	Wonly_Status_Sensor_Smoke        = "smoke"
	Wonly_Status_Sensor_Infrared     = "infrared"
	Wonly_Status_Sensor_Doorcontact  = "doorContact"
	Wonly_Status_Sensor_Flood        = "flood"
	Wonly_Status_Sensor_SOSButton    = "sosButton"
	Wonly_Status_Pad_People          = "peopleDetection"
)

//涂鸦设备状态code枚举
const (
	Ty_Status_Electricity_Left    = "electricity_left"
	Ty_Status_Clean_Record        = "clean_record"
	Ty_Status_Power               = "power"
	Ty_Status                     = "status"
	Ty_Status_Doorcontact_State   = "doorcontact_state"
	Ty_Status_Gas_Sensor_Status   = "gas_sensor_status"
	Ty_Status_Smoke_Sensor_Status = "smoke_sensor_status"
	Ty_Status_Va_Temperature      = "va_temperature"
	Ty_Status_Va_Humidity         = "va_humidity"
	Ty_Status_Watersensor_State   = "watersensor_state"
	Ty_Status_PM25_Value          = "pm25_value"
	Ty_Status_CO_Value            = "co_value"
	TY_Status_CO2_Value           = "co2_value"
	Ty_Status_Presence_State      = "presence_state"
	Ty_Status_Scene_1             = "scene_1"
	Ty_Status_Scene_2             = "scene_2"
	Ty_Status_Scene_3             = "scene_3"
	Ty_Status_Scene_4             = "scene_4"
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

type TyHandle func(devId string, rawJsonData gjson.Result)

//涂鸦处理分类
var (
	TyHandleMap = map[string]TyHandle{
		Ty_Status_Electricity_Left: tyDevBattHandle,
		Ty_Status_Power:            tyDevOnlineHandle,
		Ty_Status:                  tyDevStatusHandle,

		Ty_Status_Gas_Sensor_Status:   tyAlarmSensorHandle,
		Ty_Status_Smoke_Sensor_Status: tyAlarmSensorHandle,
		Ty_Status_Watersensor_State:   tyAlarmSensorHandle,
		Ty_Status_Presence_State:      tyAlarmSensorHandle,
		Ty_Status_Doorcontact_State:   tyAlarmSensorHandle,

		Ty_Status_Va_Temperature: tyEnvSensorHandle,
		Ty_Status_Va_Humidity:    tyEnvSensorHandle,
		Ty_Status_PM25_Value:     tyEnvSensorHandle,
		TY_Status_CO2_Value:      tyEnvSensorHandle,

		Ty_Status_Scene_1: tyDevSceneHandle,
		Ty_Status_Scene_2: tyDevSceneHandle,
		Ty_Status_Scene_3: tyDevSceneHandle,
		Ty_Status_Scene_4: tyDevSceneHandle,
	}

	TySensor2WonlySensor = map[string]string{
		Ty_Status_Gas_Sensor_Status:   Wonly_Status_Sensor_Gas,
		Ty_Status_Smoke_Sensor_Status: Wonly_Status_Sensor_Smoke,
		Ty_Status_Doorcontact_State:   Wonly_Status_Sensor_Doorcontact,
		Ty_Status_Va_Temperature:      Wonly_Status_Sensor_Temperature,
		Ty_Status_Va_Humidity:         Wonly_Status_Sensor_Humidity,
		Ty_Status_Watersensor_State:   Wonly_Status_Sensor_Flood,
		Ty_Status_PM25_Value:          Wonly_Status_Sensor_PM25,
		TY_Status_CO2_Value:           Wonly_Status_Sensor_CO2,
		Ty_Status_Presence_State:      Wonly_Status_Sensor_Infrared,
	}

	SensorVal2Str = map[string]([]string){
		Wonly_Status_Sensor_Gas:         {"检测正常", "燃气浓度已超标，正在报警"},
		Wonly_Status_Sensor_Smoke:       {"检测正常", "烟雾浓度已超标，正在报警"},
		Wonly_Status_Sensor_Flood:       {"检测正常", "水浸位已超标，正在报警"},
		Wonly_Status_Sensor_Infrared:    {"无人经过", "有人经过"},
		Wonly_Status_Sensor_Doorcontact: {"门磁已关闭", "门磁已打开"},
	}
)
