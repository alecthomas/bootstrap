package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/log15"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	floatFormat = 'f'
	timeFormat  = "15:04:05"
)

type LogLevel log15.Lvl

type ConsoleLogHandler struct {
	w colour.Printer
}

func NewConsoleLogHandler(w io.Writer) log15.Handler {
	return &ConsoleLogHandler{colour.TTY(w)}
}

func (c *ConsoleLogHandler) Log(r *log15.Record) error {
	cf := ""
	switch r.Lvl {
	case log15.LvlCrit:
		cf = "^B^1"
	case log15.LvlError:
		cf = "^1"
	case log15.LvlWarn:
		cf = "^3"
	case log15.LvlInfo:
		cf = "^7"
	case log15.LvlDebug:
		cf = "^5"
	case log15.LvlFine:
		cf = "^D^7"
	case log15.LvlFiner:
		cf = "^D^5"
	case log15.LvlFinest:
		cf = "^D^4"
	}
	c.w.Printf(cf+"%s %s ▶^R %s", r.Time.Format(timeFormat), strings.ToUpper(r.Lvl.String()), r.Msg)
	if len(r.Ctx) > 0 {
		c.w.Printf("   ^D^7(^R")
		for i := 0; i < len(r.Ctx); i += 2 {
			if i > 0 {
				c.w.Printf(" ")
			}
			k, ok := r.Ctx[i].(string)
			if !ok {
				c.w.Printf("^1invalidKey^R=%v", r.Ctx[i])
				continue
			}
			c.w.Printf(cf+"%s^R=%s", k, formatLogfmtValue(r.Ctx[i+1]))
		}
		c.w.Printf("^D^7)^R")
	}
	c.w.Printf("\n")
	return nil
}

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
	level, err := log15.LvlFromString(v)
	if err != nil {
		return fmt.Errorf("invalid log level '%s'", v)
	}
	*l = LogLevel(level)
	return nil
}

func (l *LogLevel) Level() log15.Lvl {
	return log15.Lvl(*l)
}

func ConfigureLogging(level log15.Lvl, stderr bool, logFile string) {
	backends := []log15.Handler{}

	if stderr {
		backends = append(backends, NewConsoleLogHandler(os.Stderr))
	}

	if logFile != "" {
		backends = append(backends, log15.Must.FileHandler(logFile, log15.LogfmtFormat()))
	}

	log15.Root().SetHandler(log15.LvlFilterHandler(level, log15.MultiHandler(backends...)))
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
