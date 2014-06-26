package util

import (
	"os"

	"github.com/alecthomas/go-logging"
	"github.com/alecthomas/kingpin"
)

type ModuleFlags int

// Available functionality modules.
const (
	LoggingModule ModuleFlags = 1 << iota
	DebugModule
	PIDFileModule
	DaemonizeModule

	AllModules = -1
)

var (
	logLevelFlag  logging.Level
	logFileFlag   *os.File
	logStderrFlag bool
	DebugFlag     bool

	daemonizeFlag bool
	pidFileFlag   *os.File
)

func Bootstrap(app *kingpin.Application, flags ModuleFlags, log **logging.Logger) {
	if flags&LoggingModule != 0 {
		LogLevelFlagParserVar(&logLevelFlag, app.Flag("log-level", "Set the default log level.").Default("info"))
		app.Flag("log-file", "Enable file logging to PATH.").PlaceHolder("PATH").OpenFileVar(&logFileFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		app.Flag("log-stderr", "Log to stderr (defaults to true).").Default("true").BoolVar(&logStderrFlag)
	}
	if flags&DebugModule != 0 {
		app.Flag("debug", "Enable debug mode.").BoolVar(&DebugFlag)
	}
	if flags&PIDFileModule != 0 {
		app.Flag("pid-file", "Write PID file to PATH.").Short('p').PlaceHolder("PATH").OpenFileVar(&pidFileFlag, os.O_CREATE|os.O_RDWR, 0600)
	}
	if flags&DaemonizeModule != 0 {
		app.Flag("daemonize", "Daemonize the process.").Short('d').BoolVar(&daemonizeFlag)
	}

	kingpin.MustParse(app.Parse(os.Args[1:]))

	if flags&PIDFileModule != 0 && pidFileFlag != nil {
		_, err := AcquireLock(pidFileFlag)
		kingpin.FatalIfError(err, "failed to acquire lock")
	}

	if flags&DaemonizeModule != 0 && daemonizeFlag {
		p, err := Daemonize(false, DebugFlag)
		kingpin.FatalIfError(err, "failed to daemonize")
		// We are the parent, exit.
		if p != nil {
			os.Exit(0)
		}
	}

	if flags&LoggingModule != 0 {
		*log = ConfigureLogging(app.Name, logLevelFlag, logStderrFlag, logFileFlag)
	}
}
