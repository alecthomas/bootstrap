package util

import (
	"fmt"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2-unstable"
	"github.com/alecthomas/log15"
)

var (
	logLevels = map[string]log15.Lvl{
		"debug":    log15.LvlDebug,
		"info":     log15.LvlInfo,
		"warning":  log15.LvlWarn,
		"error":    log15.LvlError,
		"critical": log15.LvlCrit,
	}
)

type LogLevel log15.Lvl

func LogLevelFlagParser(settings kingpin.Settings) (target *log15.Lvl) {
	target = new(log15.Lvl)
	settings.SetValue((*LogLevel)(target))
	return
}

func LogLevelFlagParserVar(target *log15.Lvl, settings kingpin.Settings) {
	settings.SetValue((*LogLevel)(target))
	return
}

func (l *LogLevel) String() string {
	return strings.ToLower(log15.Lvl(*l).String())
}

func (l *LogLevel) Set(v string) error {
	level, ok := logLevels[v]
	if !ok {
		return fmt.Errorf("invalid log level '%s'", v)
	}
	*l = LogLevel(level)
	return nil
}

func (l *LogLevel) Level() log15.Lvl {
	return log15.Lvl(*l)
}

func ConfigureLogging(log log15.Logger, level log15.Lvl, stderr bool, logFile string) {
	backends := []log15.Handler{}

	if stderr {
		backends = append(backends, log15.StderrHandler)
	}

	if logFile != "" {
		backends = append(backends, log15.Must.FileHandler(logFile, log15.LogfmtFormat()))
	}

	log.SetHandler(log15.LvlFilterHandler(level, log15.MultiHandler(backends...)))
}
