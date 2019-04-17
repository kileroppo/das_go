package log

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"time"
	"../file"
)
const MaxFileCap = 1024*1024*35
var m_FileName string

var(
	ArgsInvaild = errors.New("args can be vaild")
	ObtainFileFail = errors.New("obtain file failed")
	OpenFileFail = errors.New("open file failed")
	GetLineNumFail = errors.New("get line num faild")
	WriteLogInfoFail = errors.New("write log msg failed")
	LogFileError = errors.New("log file path invaild")
)
var log = logging.MustGetLogger("das_go")

var format = logging.MustStringFormatter(
	`%{color}%{time} %{shortfunc} > %{level:.4s} %{pid}%{color:reset} %{message}`,
)

func NewLogger(pathDir string, level string) {
	log.Debug("NewLogger, pathDir=", pathDir, ", level=", level)

	//时间文件夹
	destFilePath := fmt.Sprintf("%s/%d%02d%02d", pathDir, time.Now().Year(), time.Now().Month(), time.Now().Day())
	flag, err := file.IsExist(destFilePath)
	if err != nil{
		fmt.Println(ArgsInvaild)
	}
	if !flag {
		os.MkdirAll(destFilePath, os.ModePerm)
	}

	// 文件夹存在, 直接以创建的方式打开文件
	logFilePath := fmt.Sprintf("%s/%s_%02d%02d%02d%s", destFilePath, "das_go", time.Now().Hour(), time.Now().Minute(), time.Now().Second(),".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE | os.O_APPEND | os.O_RDWR, 0755)
	if err != nil{
		fmt.Println( OpenFileFail,err.Error())
		return
	}
	m_FileName = logFilePath
	/*dirExist, _ := file.PathExists(path.Dir(pathName))
	if !dirExist {
		os.MkdirAll(path.Dir(pathName), 0777)
	}
	os.Create(pathName)
	logFile, err := os.OpenFile(pathName, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}*/
	fileLog := logging.NewLogBackend(logFile, "", 0)
	stdLog := logging.NewLogBackend(os.Stderr, "", 0)
	stdFormatter := logging.NewBackendFormatter(fileLog, format)
	fileLeveled := logging.AddModuleLevel(stdLog)
	lLevel, _ := logging.LogLevel(level)
	fileLeveled.SetLevel(lLevel, "")
	logging.SetBackend(fileLeveled, stdFormatter)

	// 负责检测文件大小，超过35M则分文件
	go ReNewLogger(pathDir, level)
}

func ReNewLogger(pathDir string, level string) {
	for {
		time.Sleep(10000) // 10秒检测一次

		_, fileSize := file.GetFileByteSize(m_FileName)
		if int64(fileSize) > int64(MaxFileCap) { // 文件大小大于35M，另起文件
			NewLogger(pathDir, level)
		}
	}
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Notice(args ...interface{}) {
	log.Notice(args...)
}

func Noticef(format string, args ...interface{}) {
	log.Noticef(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Warning(args ...interface{}) {
	log.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Critical(args ...interface{}) {
	log.Critical(args...)
}

func Criticalf(format string, args ...interface{}) {
	log.Criticalf(format, args...)
}

func CheckError(err error) {
	if err != nil {
		Errorf("Fatal error: %s", err.Error())
	}
}
