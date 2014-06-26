package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/go-logging"
	"github.com/alecthomas/kingpin"
)

var (
	logLevels = map[string]logging.Level{
		"debug":    logging.DEBUG,
		"info":     logging.INFO,
		"notice":   logging.NOTICE,
		"warning":  logging.WARNING,
		"error":    logging.ERROR,
		"critical": logging.CRITICAL,
	}
)

type LogLevel logging.Level

func LogLevelFlagParser(settings kingpin.Settings) (target *logging.Level) {
	target = new(logging.Level)
	settings.SetValue((*LogLevel)(target))
	return
}

func (l *LogLevel) String() string {
	return strings.ToLower(logging.Level(*l).String())
}

func (l *LogLevel) Set(v string) error {
	level, ok := logLevels[v]
	if !ok {
		return fmt.Errorf("invalid log level '%s'", v)
	}
	*l = LogLevel(level)
	return nil
}

func (l *LogLevel) Level() logging.Level {
	return logging.Level(*l)
}

func ConfigureLogging(module string, level logging.Level, stderr bool, logFile *os.File) *logging.Logger {
	log := logging.MustGetLogger("sbusd")

	backends := []logging.Backend{}

	if stderr {
		logBackend := logging.NewLogBackend(os.Stderr, "", 0)
		logBackend.Color = true
		backends = append(backends, logBackend)
	}

	if logFile != nil {
		fileLogBackend := logging.NewLogBackend(logFile, "", 0)
		backends = append(backends, fileLogBackend)
	}

	logging.SetBackend(backends...)
	logging.SetFormatter(logging.MustStringFormatter("%{time:2006-01-02 15:04:05} %{shortfile} ▶ %{level:.1s} 0x%{id:x} %{message}"))
	logging.SetLevel(level, module)
	return log
}

func ConfigureLoggingFromFlags(module string) *logging.Logger {
	return ConfigureLogging(module, *logLevelFlag, *logStderrFlag, *logFileFlag)
}
