package log

import "fmt"

type VerType int

const (
	Test   VerType = 0
	Beta   VerType = 1
	Stable VerType = 2
)

var (
	sysName = "das"
	version = "0.3.0"
	sysType = Stable

	SysName = getSysName()
)

func getSysName() string {
	bt := ""
	switch sysType {
	case Test:
		bt = "_test"
	case Beta:
		bt = "_beta"
	case Stable:
		bt = "_stable"
	}

	return fmt.Sprintf("%s%s", sysName, bt)
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
