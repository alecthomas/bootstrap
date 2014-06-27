package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/op/go-logging"
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

func LogLevelFlagParserVar(target *logging.Level, settings kingpin.Settings) {
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

func ConfigureLogging(level logging.Level, format string, stderr bool, logFile *os.File) {
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
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(level, "")
}
