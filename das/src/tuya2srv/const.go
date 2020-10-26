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
	Ty_Status_Gas_Sensor_Status   = "gas_sensor_status"
	Ty_Status_Smoke_Sensor_Status = "smoke_sensor_status"
	Ty_Status_Temperature         = "TMP"
	Ty_Status_Humidity            = "HUM"
	Ty_Status_Watersensor_State   = "watersensor_state"
	Ty_Status_PM25_Value          = "PM2_5"
	Ty_Status_CO_Value            = "CO"
	TY_Status_CO2_Value           = "CO2"
	Ty_Status_VOC                 = "VOC"
	Ty_Status_CH2O                = "CH2O"
	Ty_Status_Presence_State      = "presence_state"
	Ty_Status_Scene_1             = "scene_1"
	Ty_Status_Scene_2             = "scene_2"
	Ty_Status_Scene_3             = "scene_3"
	Ty_Status_Scene_4             = "scene_4"
	Ty_Status_Temper_Alarm        = "temper_alarm"
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

type TyStatusHandle func(devId string, rawJsonData gjson.Result)
type TyEventHandle func(devId, tyEvent string, rawJsonData gjson.Result)

//涂鸦处理分类
var (
	TyDevStatusHandlers = map[string]TyStatusHandle{
		Ty_Status_Electricity_Left: TyStatusBattHandle,
		Ty_Status_Power:            TyStatusPowerHandle,
		Ty_Status:                  TyStatusNormalHandle,

		Ty_Status_Clean_Record: TyStatusCleanRecordHandle,

		Ty_Status_Gas_Sensor_Status:   TyStatusAlarmSensorHandle,
		Ty_Status_Smoke_Sensor_Status: TyStatusAlarmSensorHandle,
		Ty_Status_Watersensor_State:   TyStatusAlarmSensorHandle,
		Ty_Status_Presence_State:      TyStatusAlarmSensorHandle,
		Ty_Status_Doorcontact_State:   TyStatusAlarmSensorHandle,
		Ty_Status_Temper_Alarm:        TyStatusAlarmSensorHandle,

		Ty_Status_Temperature: TyStatusEnvSensorHandle,
		Ty_Status_Humidity:    TyStatusEnvSensorHandle,
		Ty_Status_PM25_Value:  TyStatusEnvSensorHandle,
		TY_Status_CO2_Value:   TyStatusEnvSensorHandle,
		Ty_Status_VOC:         TyStatusEnvSensorHandle,
		Ty_Status_CH2O:        TyStatusEnvSensorHandle,

		Ty_Status_Scene_1: TyStatusSceneHandle,
		Ty_Status_Scene_2: TyStatusSceneHandle,
		Ty_Status_Scene_3: TyStatusSceneHandle,
		Ty_Status_Scene_4: TyStatusSceneHandle,
	}

	TyDevEventHandlers = map[string]TyEventHandle{
		Ty_Event_Online: TyEventOnOffHandle,
		Ty_Event_Offline: TyEventOnOffHandle,
	}

	TySensor2WonlySensor = map[string]string{
		Ty_Status_Gas_Sensor_Status:   constant.Wonly_Status_Sensor_Gas,
		Ty_Status_Smoke_Sensor_Status: constant.Wonly_Status_Sensor_Smoke,
		Ty_Status_Doorcontact_State:   constant.Wonly_Status_Sensor_Doorcontact,
		Ty_Status_Temperature:         constant.Wonly_Status_Sensor_Temperature,
		Ty_Status_Humidity:            constant.Wonly_Status_Sensor_Humidity,
		Ty_Status_Watersensor_State:   constant.Wonly_Status_Sensor_Flood,
		Ty_Status_PM25_Value:          constant.Wonly_Status_Sensor_PM25,
		TY_Status_CO2_Value:           constant.Wonly_Status_Sensor_CO2,
		Ty_Status_Presence_State:      constant.Wonly_Status_Sensor_Infrared,
		Ty_Status_CH2O:                constant.Wonly_Status_Sensor_Formaldehyde,
		Ty_Status_Temper_Alarm:        constant.Wonly_Status_Sensor_Forced_Break,
	}

	TyEnvSensorUnitTrans = map[string]int{
		Ty_Status_Temperature: 10,
		TY_Status_CO2_Value:   100,
		Ty_Status_CH2O:        100,
	}
)
