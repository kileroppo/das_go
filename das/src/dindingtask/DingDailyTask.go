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

/*func (t *timerHandler) myinit() {
	timerCtl =  timer.New(1, 2)
}

type timerHandler struct {
}

func AddTimerTask(dueInterval int, taskId string) {
	timerCtl.AddFuncWithId(time.Duration(dueInterval)*time.Second, taskId, func() {
		log.Debug("AddFuncWithId() taskid is ", taskId, ", time Duration is ", dueInterval )
	})
}

func (t *timerHandler) StartLoop() {
	timerCtl.StartTimerLoop(timer.MIN_TIMER) // 扫描的间隔时间 eq cpu hz/tick
}*/

func StartMyTimer()  {
	if 1 == timer_is_start {
		log.Debug("StartMyTimer()......" )
		cronJob = cron.New()
		specDaily := "0 0 14 * * 1-5" 	// 定义执行时间点 参照上面的说明可知 执行时间为 周一至周五每天14:00:00执行
		specWeek := "0 0 14 * * 1-5" 	// 定义执行时间点 参照上面的说明可知 执行时间为 周一至周五每天14:00:00执行
		specMonth := "0 0 14 * * 1-5" 	// 定义执行时间点 参照上面的说明可知 执行时间为 周一至周五每天14:00:00执行
		specYear := "0 0 14 * * 1-5" 	// 定义执行时间点 参照上面的说明可知 执行时间为 周一至周五每天14:00:00执行
		cronJob.AddFunc(specDaily, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() timer is doing......")
			httpgo.Http2DingDaily()
		}, "DingDailyTask")

		cronJob.AddFunc(specWeek, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() timer is doing......")
			httpgo.Http2DingDaily()
		}, "DingWeekTask")

		cronJob.AddFunc(specMonth, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() timer is doing......")
			httpgo.Http2DingDaily()
		}, "DingMonthTask")

		cronJob.AddFunc(specYear, func() {
			t := time.Now()
			t3 := t.Format("2006-01-02 15:04:05")
			log.Debug(t3, ", StartMyTimer() timer is doing......")
			httpgo.Http2DingDaily()
		}, "DingYearTask")

		cronJob.Start()
	}

	/*if 1 == timer_is_start {
		log.Debug("StartMyTimer()......" )
		timerEntry := timerHandler{}
		timerEntry.myinit()
		timerEntry.StartLoop()

		interval := 1000 * time.Millisecond
		taskId := strconv.Itoa(0)
		timerCtl.AddFuncWithId(2 * interval, taskId, func() {
			log.Debug("StartMyTimer() timer is doing...... taskid is ", taskId, ", time Duration is ", interval )
			t := time.Now()
			time1 := time.Date(t.Year(), t.Month(), t.Day(), 14, 0, 0, 0, time.Local)
			time2 := time.Date(t.Year(), t.Month(), t.Day(), 14, 0, 6, 0, time.Local)
			if t.Unix() >= time1.Unix() && t.Unix() <= time2.Unix() { // 执行时间段：每天的下午2点0秒到6秒之间执行一次
				httpgo.Http2DingDaily()
			}
		})
	}*/
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