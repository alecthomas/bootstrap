package bootstrap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/go-logging"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	floatFormat = 'f'
	timeFormat  = "15:04:05"
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
	level, ok := logLevels[strings.ToLower(v)]
	if !ok {
		return fmt.Errorf("invalid log level '%s'", v)
	}
	*l = LogLevel(level)
	return nil
}

func (l *LogLevel) Level() logging.Level {
	return logging.Level(*l)
}

func ConfigureLogging(level logging.Level, stderr bool, format string, logFile io.Writer) {
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

func escapeString(s string) string {
	needQuotes := false
	e := bytes.Buffer{}
	e.WriteByte('"')
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			needQuotes = true
		}

		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteByte('\\')
			e.WriteByte('n')
		case '\r':
			e.WriteByte('\\')
			e.WriteByte('r')
		case '\t':
			e.WriteByte('\\')
			e.WriteByte('t')
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	start, stop := 0, e.Len()
	if !needQuotes {
		start, stop = 1, stop-1
	}
	return string(e.Bytes()[start:stop])
}

func formatLogfmtValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	value = formatShared(value)
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case string:
		return escapeString(v)
	default:
		return escapeString(fmt.Sprintf("%+v", value))
	}
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}
