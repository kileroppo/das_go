package log

import (
	"../file"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"os"
	"strings"
	"time"
)

const MaxFileCap = 1024 * 1024 * 35

var m_FileName string
var m_PathName string

var (
	ArgsInvaild      = errors.New("args can be vaild")
	ObtainFileFail   = errors.New("obtain file failed")
	OpenFileFail     = errors.New("open file failed")
	GetLineNumFail   = errors.New("get line num faild")
	WriteLogInfoFail = errors.New("write log msg failed")
	LogFileError     = errors.New("log file path invaild")
)
var log = logging.MustGetLogger("das_go")

var format = logging.MustStringFormatter(
	`%{color}%{time} %{pid} %{shortfile} %{longfunc} > %{level:.4s} %{color:reset} %{message}: `,
)

func NewLogger(pathDir string, level string) {
	log.Debug("NewLogger, pathDir=", pathDir, ", level=", level)

	//时间文件夹
	destFilePath := fmt.Sprintf("%s/%d%02d%02d", pathDir, time.Now().Year(), time.Now().Month(), time.Now().Day())
	flag, err := file.IsExist(destFilePath)
	if err != nil {
		fmt.Println(ArgsInvaild)
	}
	if !flag {
		os.MkdirAll(destFilePath, os.ModePerm)
	}
	m_PathName = destFilePath

	// 文件夹存在, 直接以创建的方式打开文件
	logFilePath := fmt.Sprintf("%s/%s_%02d%02d%02d%s", destFilePath, "das_go", time.Now().Hour(), time.Now().Minute(), time.Now().Second(), ".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(OpenFileFail, err.Error())
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

		//1. 文件大小大于35M，另起文件
		_, fileSize := file.GetFileByteSize(m_FileName)
		if int64(fileSize) > int64(MaxFileCap) {
			NewLogger(pathDir, level)
		}

		//2. 每天一个文件夹
		destFilePath := fmt.Sprintf("%s/%d%02d%02d", pathDir, time.Now().Year(), time.Now().Month(), time.Now().Day())
		if 0 != strings.Compare(destFilePath, m_PathName) {
			NewLogger(pathDir, level)
		}
	}
}

var (
	Info      = log.Info
	Infof     = log.Infof
	Notice    = log.Notice
	Noticef   = log.Noticef
	Debug     = log.Debug
	Debugf    = log.Debugf
	Warning   = log.Warning
	Warningf  = log.Warningf
	Error     = log.Error
	Errorf    = log.Errorf
	Critical  = log.Critical
	Criticalf = log.Criticalf
)

func CheckError(err error) {
	if err != nil {
		Errorf("Fatal error: %s", err.Error())
	}
}
