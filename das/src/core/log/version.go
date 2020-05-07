package log

var (
	sysName = "DAS"
	version = "0.0.1"
	verType = "beta"
)

func Version() string {
	return version
}

func VerType() string {
	return verType
}

func PrintVersion()  {
	log.Infof("%s Ver:%s%s starting...\n", sysName, verType, version)
}