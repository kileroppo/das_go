package tuya2srv

import (
	"github.com/tidwall/gjson"

	"das/core/constant"
)

//涂鸦设备状态code枚举
const (
	Ty_Status_Electricity_Left = "electricity_left"
	Ty_Status_Clean_Record     = "clean_record"
	Ty_Status_Power            = "power"
	Ty_Status                  = "status"
	Ty_Status_Mode             = "mode"

	//传感器
	Ty_Status_Doorcontact_State   = "doorcontact_state"
	Ty_Status_Gas_Sensor_State    = "gas_sensor_state"
	Ty_Status_Smoke_Sensor_Status = "smoke_sensor_status"
	Ty_Status_Watersensor_State   = "watersensor_state"
	Ty_Status_Temper_Alarm        = "temper_alarm"
	Ty_Status_Battery_Percentage  = "battery_percentage"
	Ty_Status_SOS_State           = "sos_state"
	Ty_Status_Presence_State      = "presence_state"
	Ty_Status_Pir                 = "pir"

	Ty_Status_Air_Temperature = "TMP"
	Ty_Status_Air_Humidity    = "HUM"
	Ty_Status_Air_PM25_Value  = "PM2_5"
	Ty_Status_Air_CO_Value    = "CO"
	TY_Status_Air_CO2_Value   = "CO2"
	Ty_Status_Air_VOC         = "VOC"
	Ty_Status_Air_CH2O        = "CH2O"

	Ty_Status_Va_Temperature = "va_temperature"
	Ty_Status_Va_Humidity    = "va_humidity"

	//灯
	Ty_Status_Switch_Led      = "switch_led"
	Ty_Status_Bright_Value    = "bright_value"
	Ty_Status_Bright_Value_V2 = "bright_value_v2"
	Ty_Status_Colour_Data     = "colour_data"
	Ty_Status_Colour_Data_V2  = "colour_data_v2"
	Ty_Status_Work_Mode       = "work_mode"
	Ty_Status_Scene_1         = "scene_1"
	Ty_Status_Scene_2         = "scene_2"
	Ty_Status_Scene_3         = "scene_3"
	Ty_Status_Scene_4         = "scene_4"

	Ty_Status_Switch   = "switch"
	Ty_Status_Switch_1 = "switch_1"
	Ty_Status_Switch_2 = "switch_2"
	Ty_Status_Switch_3 = "switch_3"
	Ty_Status_Switch_4 = "switch_4"

	Ty_Status_Switch_Val  = "switch_value"
	Ty_Status_Switch1_Val = "switch1_value"
	Ty_Status_Switch2_Val = "switch2_value"
	Ty_Status_Switch3_Val = "switch3_value"
	Ty_Status_Switch4_Val = "switch4_value"

	//窗帘电机
	Ty_Status_Percent_Control   = "percent_control"
	Ty_Status_Percent_Control_2 = "percent_control_2"
	Ty_Status_Work_State        = "work_state"

	//睡眠带
	Ty_Status_Sleep_Stage      = "sleep_stage"
	Ty_Status_Off_Bed          = "off_bed"
	Ty_Status_Wakeup           = "wakeup"
	Ty_Status_Heart_Rate       = "heart_rate"
	Ty_Status_Respiratory_Rate = "respiratory_rate"
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

//涂鸦扫地机状态
const (
	Ty_Cleaner_Standby     = "standby"
	Ty_Cleaner_Smart       = "smart"
	Ty_Cleaner_Spiral      = "spiral"
	Ty_Cleaner_Single      = "single"
	Ty_Cleaner_Chargego    = "chargego"
	Ty_Cleaner_Wall_Follow = "wall_follow"
	Ty_Cleaner_Power_Go    = "power_go"
	Ty_Cleaner_Cleaning    = "cleaning"
	Ty_Cleaner_Goto_Charge = "goto_charge"
	Ty_Cleaner_Stop        = "stop"
	Ty_Cleaner_Paused      = "paused"
	Ty_Cleaner_Charging    = "charging"
	Ty_Cleaner_Charge_Done = "charge_done"
	Ty_Cleaner_Sleep       = "sleep"
)

//涂鸦睡眠袋睡眠状态
const (
	Ty_Sleep_Stage_Awake = "awake"
	Ty_Sleep_Stage_Sleep = "sleep"
)

//涂鸦窗帘状态
const (
	Ty_Cmd_Work_State_Val_Open  = "opening"
	Ty_Cmd_Work_State_Val_Close = "closing"
)

type TyStatusHandle func(devId string, rawJsonData gjson.Result)
type TyEventHandle func(devId, tyEvent string, rawJsonData gjson.Result)

//涂鸦处理分类
var (
	TyDevStatusHandlers = map[string]TyStatusHandle{
		Ty_Status_Electricity_Left:   TyStatusRobotCleanerBattHandle,
		Ty_Status_Power:              TyStatusPowerHandle,
		Ty_Status:                    TyStatusNormalHandle,
		Ty_Status_Mode:               TyStatusNormalHandle,
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
		Ty_Status_Bright_Value:   TyStatusEnvSensorHandle,

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

		Ty_Status_Sleep_Stage: TyStatusSleepStage,
		Ty_Status_Off_Bed:     TyStatusOffBed,
		Ty_Status_Wakeup:      TyStatusWakeup,

		Ty_Status_Work_State: TyStatusCurtainHandle,
	}

	TyDevEventHandlers = map[string]TyEventHandle{
		Ty_Event_Online:  TyEventOnlineHandle,
		Ty_Event_Offline: TyEventOnlineHandle,
		Ty_Event_Delete:  TyEventDeleteHandle,
	}

	TyDevEventOperZh = map[string]string{
		Ty_Event_Online:         "设备上线",
		Ty_Event_Offline:        "设备离线",
		Ty_Event_Name_Update:    "修改设备名称",
		Ty_Event_Dp_Name_Update: "修改设备功能点名称",
		Ty_Event_Bind_User:      "设备绑定用户",
		Ty_Event_Delete:         "删除设备",
		Ty_Event_Upgrade_Status: "设备升级状态",
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

	TyEnvSensorValTransfer = map[string]float64{
		Ty_Status_Air_Temperature: 0.1,
		TY_Status_Air_CO2_Value:   0.01,
		Ty_Status_Air_CH2O:        0.01,
		Ty_Status_Va_Humidity:     0.01,
		Ty_Status_Va_Temperature:  0.01,
		Ty_Status_Bright_Value:    10.483,
	}

	TyCleanerStatusNote = map[string]string{
		Ty_Cleaner_Standby:     "待机",
		Ty_Cleaner_Chargego:    "回充中",
		Ty_Cleaner_Single:      "清扫中",
		Ty_Cleaner_Smart:       "清扫中",
		Ty_Cleaner_Spiral:      "清扫中",
		Ty_Cleaner_Wall_Follow: "清扫中",
		Ty_Cleaner_Cleaning:    "清扫中",
		Ty_Cleaner_Goto_Charge: "回充中",
		Ty_Cleaner_Paused:      "暂停",
		Ty_Cleaner_Stop:        "暂停",
		Ty_Cleaner_Charging:    "充电中",
		Ty_Cleaner_Charge_Done: "充电完成",
		Ty_Cleaner_Sleep:       "休眠",
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

	TyStatusDataFilterMap = map[string]struct{}{
		Ty_Status_Power:              {},
		Ty_Status:                    {},
		Ty_Status_Mode:               {},
		Ty_Status_Battery_Percentage: {},

		//灯类
		Ty_Status_Switch_Led:      {},
		Ty_Status_Bright_Value:    {},
		Ty_Status_Bright_Value_V2: {},
		Ty_Status_Colour_Data:     {},
		Ty_Status_Colour_Data_V2:  {},
		Ty_Status_Work_Mode:       {},

		//开关类
		Ty_Status_Scene_1:     {},
		Ty_Status_Scene_2:     {},
		Ty_Status_Scene_3:     {},
		Ty_Status_Scene_4:     {},
		Ty_Status_Switch:      {},
		Ty_Status_Switch_1:    {},
		Ty_Status_Switch_2:    {},
		Ty_Status_Switch_3:    {},
		Ty_Status_Switch_4:    {},
		Ty_Status_Switch1_Val: {},
		Ty_Status_Switch2_Val: {},
		Ty_Status_Switch3_Val: {},
		Ty_Status_Switch4_Val: {},

		//窗帘类
		Ty_Status_Percent_Control:   {},
		Ty_Status_Percent_Control_2: {},
		Ty_Status_Work_State:        {},

		//传感器类
		Ty_Status_Doorcontact_State:   {},
		Ty_Status_Gas_Sensor_State:    {},
		Ty_Status_Smoke_Sensor_Status: {},
		Ty_Status_Watersensor_State:   {},
		Ty_Status_Temper_Alarm:        {},
		Ty_Status_SOS_State:           {},
		Ty_Status_Presence_State:      {},
		Ty_Status_Pir:                 {},
		Ty_Status_Air_Temperature:     {},
		Ty_Status_Air_Humidity:        {},
		Ty_Status_Air_PM25_Value:      {},
		Ty_Status_Air_CO_Value:        {},
		TY_Status_Air_CO2_Value:       {},
		Ty_Status_Air_VOC:             {},
		Ty_Status_Air_CH2O:            {},
		Ty_Status_Va_Temperature:      {},
		Ty_Status_Va_Humidity:         {},

		//睡眠带
		Ty_Status_Wakeup:           {},
		Ty_Status_Sleep_Stage:      {},
		Ty_Status_Off_Bed:          {},
		Ty_Status_Heart_Rate:       {},
		Ty_Status_Respiratory_Rate: {},
	}

	tyAlarmDataFilterMap = map[string]struct{}{
		Ty_Status_Gas_Sensor_State:    {},
		Ty_Status_Smoke_Sensor_Status: {},
		Ty_Status_Doorcontact_State:   {},
		Ty_Status_Pir:                 {},
		Ty_Status_Temper_Alarm:        {},
		Ty_Status_Watersensor_State:   {},
	}
)
