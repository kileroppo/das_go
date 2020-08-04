package log

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
)

func Version() string {
	return version
}

func SysType() VerType {
	return sysType
}

func PrintVersion()  {
	log.Infof("%s Ver:%s starting...\n", sysName, version)
}