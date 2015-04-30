package util

import (
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2-unstable"
	"gopkg.in/inconshreveable/log15.v2"
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
	logLevelFlag  log15.Lvl
	logFormatFlag string
	logFileFlag   string
	logStderrFlag bool
	DebugFlag     bool
	daemonizeFlag bool
	pidFileFlag   *os.File
)

type Options struct {
	UseSystemPIDFilePath bool   // Use $TEMPDIR/<appname>.pid by default.
	PIDFile              string // Path to PID file. Overrides previous option.
	LogToStderrByDefault bool   // Log to stderr.
	LogFile              string // Log to this file.
}

func Bootstrap(app *kingpin.Application, flags ModuleFlags, options *Options) string {
	if options == nil {
		options = &Options{}
	}

	if flags&PIDFileModule != 0 {
		path := options.PIDFile
		if options.UseSystemPIDFilePath && path == "" {
			path = filepath.Join(os.TempDir(), app.Name+".pid")
		}
		app.Flag("pid-file", "Write PID file to PATH.").Short('p').Default(path).OpenFileVar(&pidFileFlag, os.O_CREATE|os.O_RDWR, 0600)
	}

	if flags&DaemonizeModule != 0 {
		app.Flag("daemon", "Daemonize the process.").Short('d').BoolVar(&daemonizeFlag)
	}

	// Configure flags.
	if flags&LoggingModule != 0 {
		LogLevelFlagParserVar(&logLevelFlag, app.Flag("log-level", "Set the default log level.").Default("info"))
		flag := app.Flag("log-file", "Enable file logging to PATH.")
		if options.LogFile != "" {
			flag.Default(options.LogFile)
		} else {
			flag.PlaceHolder("PATH")
		}
		flag.StringVar(&logFileFlag)
		value := "false"
		if options.LogToStderrByDefault {
			value = "true"
		}
		app.Flag("log-stderr", "Log to stderr.").Default(value).BoolVar(&logStderrFlag)
	}

	if flags&DebugModule != 0 {
		app.Flag("debug", "Enable debug mode.").BoolVar(&DebugFlag)
	}

	args, err := kingpin.ExpandArgsFromFiles(os.Args[1:])
	kingpin.FatalIfError(err, "failed to expand flags from files")
	command := kingpin.MustParse(app.Parse(args))

	// Initialise all the various modules.
	if flags&PIDFileModule != 0 && pidFileFlag != nil {
		_, err := AcquireLock(pidFileFlag)
		if err != nil {
			kingpin.Fatalf("failed to acquire lock %s", pidFileFlag.Name())
		}
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
		ConfigureLogging(logLevelFlag, app.Name, logStderrFlag, logFileFlag)
	}

	return command
}
