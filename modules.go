package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/alecthomas/go-logging"
	"gopkg.in/alecthomas/kingpin.v2"
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
	logFormatFlag string
	logFileFlag   *os.File
	logStderrFlag bool
	DebugFlag     bool
	daemonizeFlag bool
	pidFileFlag   *os.File
)

type bootstrapOptions struct {
	modules              ModuleFlags
	useSystemPIDFilePath bool   // Use $TEMPDIR/<appname>.pid by default.
	pidFile              string // Path to PID file. Overrides previous option.
	logToStderrByDefault bool   // Log to stderr.
	logFile              string // Log to this file.
	logFormat            string // Log format (from go-logging).
	logLevel             logging.Level
}

type Option func(*bootstrapOptions)

func WithModule(module ModuleFlags) func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.modules |= module }
}

func WithSystemPIDFile() func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.useSystemPIDFilePath = true }
}

func WithPIDFile(path string) func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.pidFile = path }
}

func WithStderrByDefault() func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.logToStderrByDefault = true }
}

func WithFile(path string) func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.logFile = path }
}

func WithFormat(fmt string) func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.logFormat = fmt }
}

func WithLevel(level logging.Level) func(*bootstrapOptions) {
	return func(o *bootstrapOptions) { o.logLevel = level }
}

// Bootstrap the application.
func Bootstrap(app *kingpin.Application, options ...Option) string {
	opts := &bootstrapOptions{}

	for _, opt := range options {
		opt(opts)
	}

	if opts.logFormat == "" {
		opts.logFormat = "%{time:2006-01-02 15:04:05} %{module} ▶ %{level:.1s} 0x%{id:x} %{message}"
	}

	if opts.modules&PIDFileModule != 0 {
		path := opts.pidFile
		if opts.useSystemPIDFilePath && path == "" {
			path = filepath.Join(os.TempDir(), app.Name+".pid")
		}
		app.Flag("pid-file", "Write PID file to PATH.").Short('p').Default(path).OpenFileVar(&pidFileFlag, os.O_CREATE|os.O_RDWR, 0600)
	}

	if opts.modules&DaemonizeModule != 0 {
		app.Flag("daemon", "Daemonize the process.").Short('d').BoolVar(&daemonizeFlag)
	}

	// Configure flags.
	if opts.modules&LoggingModule != 0 {
		LogLevelFlagParserVar(&logLevelFlag, app.Flag("log-level", "Set the default log level.").Default(opts.logLevel.String()))
		app.Flag("log-format", "Set log output format (see go-logging).").Default(opts.logFormat).PlaceHolder("FORMAT").StringVar(&logFormatFlag)
		flag := app.Flag("log-file", "Enable file logging to PATH.")
		if opts.logFile != "" {
			flag.Default(opts.logFile)
		} else {
			flag.PlaceHolder("PATH")
		}
		flag.FileVar(&logFileFlag)
		value := "false"
		if opts.logToStderrByDefault {
			value = "true"
		}
		app.Flag("log-stderr", "Log to stderr.").Default(value).BoolVar(&logStderrFlag)
	}

	if opts.modules&DebugModule != 0 {
		app.Flag("debug", "Enable debug mode.").BoolVar(&DebugFlag)
	}
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	// Initialise all the various modules.
	if opts.modules&PIDFileModule != 0 && pidFileFlag != nil {
		_, err := AcquireLock(pidFileFlag)
		if err != nil {
			kingpin.Fatalf("failed to acquire lock %s", pidFileFlag.Name())
		}
	}

	if opts.modules&DaemonizeModule != 0 && daemonizeFlag {
		p, err := Daemonize(false, DebugFlag)
		kingpin.FatalIfError(err, "failed to daemonize")
		// We are the parent, exit.
		if p != nil {
			os.Exit(0)
		}
	}

	if opts.modules&LoggingModule != 0 {
		ConfigureLogging(logLevelFlag, logStderrFlag, logFormatFlag, logFileFlag)
	}

	return command
}
