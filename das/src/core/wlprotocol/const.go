package wlprotocol

const (
	Sleepace_Msg_HeartBeat   = 0x0000
	Sleepace_Msg_Notify      = 0x1000
	Sleepace_Msg_Verify      = 0x2000
	Sleepace_Msg_Time_Check  = 0x2001
	Sleepace_Msg_Dev_Info    = 0x0202
	Sleepace_Msg_RestOn_Info = 0x0203
	Sleepace_Msg_Upgrade     = 0x0204

	Sleepace_Msg_Alarm_Set                   = 0x0300
	Sleepace_Msg_Alarm_Query                 = 0x0301
	Sleepace_Msg_SleepAid_Conf_Set           = 0x0302
	Sleepace_Msg_SleepAid_Conf_Query         = 0x0303
	Sleepace_Msg_RestOn_Threshold_Conf_Set   = 0x0304
	Sleepace_Msg_RestOn_Threshold_Conf_Query = 0x0305
	Sleepace_Msg_RestOn_Bind_Info_Set        = 0x0306
	Sleepace_Msg_RestOn_Bind_Info_Query      = 0x0307

	Sleepace_Msg_Alarm_Del          = 0x3020
	Sleepace_Msg_RestOn_Bind_Del    = 0x3021
	Sleepace_Msg_Auto_Collect_Set   = 0x3022
	Sleepace_Msg_Auto_Collect_Query = 0x3023

	Sleepace_Msg_Music_List_Query        = 0x0400
	Sleepace_Msg_RestOn_Online_Query     = 0x0401
	Sleepace_Msg_RestOn_Collect_Query    = 0x0402
	Sleepace_Msg_RestOn_Battery_Query    = 0x0403
	Sleepace_Msg_Server_Addr_Query       = 0x0404
	Sleepace_Msg_SleepAid_Status_Query   = 0x0405
	Sleepace_Msg_Blue_List_Query         = 0x0406
	Sleepace_Msg_Dev_Work_Status_Query   = 0x0407
	Sleepace_Msg_Nox_WIFI_Firmware_Query = 0x0408
	Sleepace_Msg_Alarm_Run_Status_Query  = 0x0409
	Sleepace_Msg_History_Data_Query      = 0x0410
	Sleepace_Msg_Music_Info_Query        = 0x0411

	Sleepace_Msg_Realtime_Data         = 0x0500
	Sleepace_Msg_History_Data          = 0x0502
	Sleepace_Msg_Music_Data            = 0x0503
	Sleepace_Msg_Firmware_Outline_Data = 0x0504
	Sleepace_Msg_Firmware_Details_Date = 0x0505
)

const (
	Sleepace_Notice_Online                    = 0x0000
	Sleepace_Notice_Offline                   = 0x0001
	Sleepace_Notice_TurnOn_Light              = 0x0002
	Sleepace_Notice_TurnOff_Light             = 0x0003
	Sleepace_Notice_TurnOn_Alarm              = 0x0004
	Sleepace_Notice_TurnOff_Alarm             = 0x0005
	Sleepace_Notice_Snooze_Notity             = 0x0006
	Sleepace_Notice_Start_Realtime_Data_Query = 0x0007
	Sleepace_Notice_End_Realtime_Data_Query   = 0x0008
	Sleepace_Notice_Start_Collet              = 0x0009
	Sleepace_Notice_End_Collect               = 0x000a
	Sleepace_Notice_RestOn_Low_Battery        = 0x000b
	Sleepace_Notice_Clear_RestOn_Low_Battery  = 0x000c
	Sleepace_Notice_RestOn_Oncharge           = 0x000d

	Sleepace_Notice_Firmware_Upgrade         = 0x0020
	Sleepace_Notice_Firmware_Upgrade_Status  = 0x0021
	Sleepace_Notice_History_Data_Upload_Done = 0x0022

	Sleepace_Notice_Light_Brightness_Set = 0x0030
	Sleepace_Notice_Volume_Set           = 0x0031
	Sleepace_Notice_Start_Play_Music     = 0x0032
	Sleepace_Notice_End_Play_Music       = 0x0033
	Sleepace_Notice_End_Light_Brightness = 0x0034
	Sleepace_Notice_Enable_Alarm_Conf    = 0x0035
	Sleepace_Notice_Disable_Alarm_Conf   = 0x0036
	Sleepace_Notice_SleepAid_Op          = 0x0037
	Sleepace_Notice_Start_Config         = 0x0038
	Sleepace_Notice_End_Config           = 0x0039

	Sleepace_Notice_Music_Download        = 0x0044
	Sleepace_Notice_Music_Download_Status = 0x0045
	Sleepace_Notice_Scene_Op              = 0x0051
	Sleepace_Notice_Light_Op              = 0x0052
	Sleepace_Notice_Alarm_Op              = 0x0053
	Sleepace_Notice_Music_Op              = 0x0054
	Sleepace_Notice_Aroma_Op              = 0x0055
)
