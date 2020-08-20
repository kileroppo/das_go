package log

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dlintw/goconf"
	"github.com/json-iterator/go"
	"github.com/op/go-logging"
	gelf "github.com/robertkowalski/graylog-golang"

	"das/core/file"
)

const MaxFileCap = 1024 * 1024 * 35

var m_FileName string
var m_PathName string

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	grayCli *gelf.Gelf

	ErrArgsInvaild  = errors.New("args can be vaild")
	ErrOpenFileFail = errors.New("open file failed")

	logPath = "./logs"
	logLevel = "DEBUG"
	logSaveDay = 7

	Conf *goconf.ConfigFile
	log = logging.MustGetLogger("das_go")

	format = logging.MustStringFormatter(
		`%{color}%{time} %{pid} %{shortfile} %{longfunc} > %{level:.4s} %{color:reset} %{message}`,
	)
)

var _ sort.Interface = LogFileInfo{}

type LogFileInfo struct {
	infos []os.FileInfo
}

func (l LogFileInfo) Len() int {
    return len(l.infos)
}

func (l LogFileInfo) Less(i, j int) bool {
	return l.infos[i].ModTime().Unix() < l.infos[j].ModTime().Unix()
}

func (l LogFileInfo) Swap(i, j int) {
	l.infos[i], l.infos[j] = l.infos[j], l.infos[i]
}

func Init() *goconf.ConfigFile{
	initLogger()
	Conf = loadConfig()
	initGraylogConn()
	return Conf
}

func initGraylogConn() {
	url,err := Conf.GetString("graylog", "url")
	if err != nil {
		panic(err)
	}
	port,err := Conf.GetInt("graylog", "port")
	if err != nil {
		panic(err)
	}

	cfg := gelf.Config{
		GraylogPort:     port,
		GraylogHostname: url,
	}
	grayCli = gelf.New(cfg)
}

func initLogger() {
	go autoClearLogFiles(logPath)
	NewLogger(logPath, logLevel)
}

func NewLogger(pathDir string, level string) {
	log.Debug("NewLogger, pathDir=", pathDir, ", level=", level)

	//时间文件夹
	destFilePath := fmt.Sprintf("%s/%d%02d%02d", pathDir, time.Now().Year(), time.Now().Month(), time.Now().Day())
	flag, err := file.IsExist(destFilePath)
	if err != nil {
		fmt.Println(ErrArgsInvaild)
	}
	if !flag {
		os.MkdirAll(destFilePath, os.ModePerm)
	}
	m_PathName = destFilePath

	// 文件夹存在, 直接以创建的方式打开文件
	logFilePath := fmt.Sprintf("%s/%s_%02d%02d%02d%s", destFilePath, "das_go", time.Now().Hour(), time.Now().Minute(), time.Now().Second(), ".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(ErrOpenFileFail, err.Error())
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
	go reNewLogger(pathDir, level)
}

func reNewLogger(pathDir string, level string) {
	for {
		time.Sleep(time.Hour * 2) // 2小时检测一次

		//每天一个文件夹
		destFilePath := fmt.Sprintf("%s/%d%02d%02d", pathDir, time.Now().Year(), time.Now().Month(), time.Now().Day())
		if 0 != strings.Compare(destFilePath, m_PathName) {
			NewLogger(pathDir, level)
			return
		}

		//文件大小大于35M，另起文件
		_, fileSize := file.GetFileByteSize(m_FileName)
		if int64(fileSize) > int64(MaxFileCap) {
			NewLogger(pathDir, level)
			return
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

//autoClearLogFiles: Automatic clean-up of logs
//Everytime the system start, delete the most previous 5 days' logfiles
func autoClearLogFiles(logsDirPath string) {
	for {
		if flag, err := file.IsExist(logsDirPath); err != nil || !flag {
			time.Sleep(time.Hour * 24)
			continue
		}

		logsSubFileList, err := ioutil.ReadDir(logsDirPath)
		if err != nil {
			time.Sleep(time.Hour * 24)
			continue
		}

		var dirLogs LogFileInfo
		var sdkLogs LogFileInfo
		for i, _ := range logsSubFileList {
			if logsSubFileList[i].IsDir() {
				dirLogs.infos = append(dirLogs.infos, logsSubFileList[i])
			} else {
				sdkLogs.infos = append(sdkLogs.infos, logsSubFileList[i])
			}
		}

		var fileDirName string
		if dirLogs.Len() > logSaveDay {
			sort.Sort(dirLogs)
			for i:=0;i<dirLogs.Len() - logSaveDay;i++ {
				fileDirName = fmt.Sprintf("%s/%s", logsDirPath, dirLogs.infos[i].Name())
				os.RemoveAll(fileDirName)
			}
		}

		if sdkLogs.Len() > logSaveDay {
			sort.Sort(sdkLogs)
			for i:=0;i<sdkLogs.Len() - logSaveDay;i++ {
				fileDirName = fmt.Sprintf("%s/%s", logsDirPath, sdkLogs.infos[i].Name())
				os.Remove(fileDirName)
			}
		}

		time.Sleep(time.Hour * 24)
	}
}

func loadConfig() *goconf.ConfigFile {
	PrintVersion()
	conf_file := flag.String("config", "./das.ini", "设置配置文件.")
	flag.Parse()

	//var conf_file *string
	//var tmp = "./das.ini"
	//conf_file = &tmp

	conf, err := goconf.ReadConfigFile(*conf_file)
	if err != nil {
		Errorf("Load config file:%s failed，error = %s", *conf_file, err)
		os.Exit(1)
	}

	confPath, err := filepath.Abs(*conf_file)
	if err != nil {
		Error("loadConfig() error = ", err)
		os.Exit(1)
	}

	Info("Load config file:", confPath)
	return conf
}

func SendGraylog(format string, args ...interface{}) {
	lmsg := ""
	if len(format) == 0 {
		lmsg = fmt.Sprint(args...)
	} else {
		lmsg = fmt.Sprintf(format, args...)
	}

	msg := GrayLog{
		Version:  "2.1",
		Host:     "das",
		Facility: SysName,
		Message:  lmsg,
	}
	b, err := json.Marshal(msg)
	if err == nil {
		grayCli.Log(Bytes2Str(b))
	}
}