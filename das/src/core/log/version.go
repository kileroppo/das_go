package log

import "fmt"

type VerType int

const (
	alpha    VerType = 0
	Beta     VerType = 1
	Official VerType = 2
)

var (
	sysName = "DAS"
	version = "0.2.1"
	sysType = Beta

	SysName = getSysName()
)

func getSysName() string {
	bt := ""
	switch sysType {
	case Beta:
		bt = "_beta"
	case Official:
	}

	return fmt.Sprintf("%s%s %s", sysName, bt, version)
}

func Version() string {
	return version
}

func SysType() VerType {
	return sysType
}

func PrintVersion()  {
	log.Infof("%s Ver:%s starting...\n", sysName, version)
}