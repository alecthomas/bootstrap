package util

import (
	"os"
	"path/filepath"

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

type Options struct {
	UseSystemPIDFilePath bool             // Use $TEMPDIR/<appname>.pid by default.
	PIDFile              string           // Path to PID file. Overrides previous option.
	Logger               **logging.Logger // If logging is enabled, assign to this logger.
	LogToStderrByDefault bool             // Log to stderr.
	LogFile              string           // Log to this file.
}

func Bootstrap(app *kingpin.Application, flags ModuleFlags, options *Options) string {
	if options == nil {
		options = &Options{}
	}

	// Configure flags.
	if flags&LoggingModule != 0 {
		if options.Logger == nil {
			panic("options.Logger must be provided for LoggingModule to be used")
		}
		LogLevelFlagParserVar(&logLevelFlag, app.Flag("log-level", "Set the default log level.").Default("info"))
		flag := app.Flag("log-file", "Enable file logging to PATH.")
		if options.LogFile != "" {
			flag.Default(options.LogFile)
		} else {
			flag.PlaceHolder("PATH")
		}
		flag.OpenFileVar(&logFileFlag, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		value := "false"
		if options.LogToStderrByDefault {
			value = "true"
		}
		app.Flag("log-stderr", "Log to stderr.").Default(value).BoolVar(&logStderrFlag)
	}

	if flags&DebugModule != 0 {
		app.Flag("debug", "Enable debug mode.").BoolVar(&DebugFlag)
	}

	if flags&PIDFileModule != 0 {
		path := options.PIDFile
		if options.UseSystemPIDFilePath && path == "" {
			path = filepath.Join(os.TempDir(), app.Name+".pid")
		}
		app.Flag("pid-file", "Write PID file to PATH.").Short('p').Default(path).OpenFileVar(&pidFileFlag, os.O_CREATE|os.O_RDWR, 0600)
	}
	if flags&DaemonizeModule != 0 {
		app.Flag("daemonize", "Daemonize the process.").Short('d').BoolVar(&daemonizeFlag)
	}

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// Initialise all the various modules.
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
		*options.Logger = ConfigureLogging(app.Name, logLevelFlag, logStderrFlag, logFileFlag)
	}

	return command
}
