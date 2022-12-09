package monitor

import "flag"

//option info
var Options struct {
	port int `rpc service port`
	//logPath string `log file path`
	//logPrefix string `log file prefix`
}

//parse flags
func parseFlags() {
	flag.IntVar(&Options.port, "port", 9100, "rpc service port")
	//flag.StringVar(&Options.logPath, "logPath", "logs", "log file path")
	//flag.StringVar(&Options.logPrefix, "logPrefix", "monitor", "log file prefix")
	flag.Parse()
}
