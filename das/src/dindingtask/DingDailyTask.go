package dindingtask

import (
	"../core/timer/cron"
	"github.com/dlintw/goconf"
	"../core/httpgo"
	"../core/log"
		"time"
)

/*var (
	timerCtl *timer.TimerHeapHandler
)*/


var timer_is_start int
var cronJob *cron.Cron

//初始化
func InitTimer_IsStart(conf *goconf.ConfigFile) {
	timer_is_start, _ = conf.GetInt("timer", "is_start")
}

func StartMyTimer()  {
	if 1 == timer_is_start {
		log.Debug("StartMyTimer()......" )

		cronJob = cron.New()
		specDaily 	:= "0 0 17 * * 1-5" 		// 定义执行时间点 参照上面的说明可知 执行时间为：周一至周五每天17:00:00执行
		specWeek 	:= "0 30 16 * * 5" 			// 定义执行时间点 参照上面的说明可知 执行时间为：每周的周五16:30:00执行
		specMonth 	:= "0 0 16 23-28 * *" 		// 定义执行时间点 参照上面的说明可知 执行时间为：每个月的23日至28日16:00:00执行
		specYear 	:= "0 30 16 15-20 12 *" 	// 定义执行时间点 参照上面的说明可知 执行时间为：12月15-20日16:30:00执行

		reqDaily := "我是机器人小艾, 提醒大家：没有发日计划与总结的请及时发日计划与总结，否则罚款20元/次；日计划与总结上交时间为当日下班至次日上班前。"
		reqWeek := "我是机器人小艾, 提醒大家：又到周五了，大家的心情是不是又很嗨呢，别忘发周计划与总结，请及时发周计划与总结，否则罚款20元/次；周计划与总结上交时间为周五下班至周一上班前。"
		reqMonth := "我是机器人小艾, 提醒大家：月底了，没有发月计划与总结的请及时发月计划与总结，否则罚款20元/次；月计划时间为每月28日前，月总结为每月4日前。"
		reqYear := "我是机器人小艾, 提醒大家：没有发年计划与总结的请及时发年计划与总结，否则罚款20元/次；年计划上交时间每年12月20日前，年总结每年1月10日前。"

		cronJob.AddFunc(specDaily, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() DingDailyTask timer is doing......")
			httpgo.Http2DingDaily(reqDaily)
		}, "DingDailyTask")

		cronJob.AddFunc(specWeek, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() DingWeekTask timer is doing......")
			httpgo.Http2DingDaily(reqWeek)
		}, "DingWeekTask")

		cronJob.AddFunc(specMonth, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() DingMonthTask timer is doing......")
			httpgo.Http2DingDaily(reqMonth)
		}, "DingMonthTask")

		cronJob.AddFunc(specYear, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() DingYearTask timer is doing......")
			httpgo.Http2DingDaily(reqYear)
		}, "DingYearTask")

		cronJob.Start()
	}
}

func StopMyTimer()  {
	if 1 == timer_is_start {
		log.Debug("StopMyTimer()......" )

		// Remove an entry from the cron by name.
		log.Debug("StopMyTimer() RemoveJob()......" )
		cronJob.RemoveJob("DingDailyTask")
		cronJob.RemoveJob("DingWeekTask")
		cronJob.RemoveJob("DingMonthTask")
		cronJob.RemoveJob("DingYearTask")

		cronJob.Stop() // Stop the scheduler (does not stop any jobs already running).
	}
}