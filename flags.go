package util

import (
	"github.com/alecthomas/go-logging"
	"github.com/alecthomas/kingpin"

	"os"
)

var (
	logLevelFlag  *logging.Level
	logFileFlag   **os.File
	logStderrFlag *bool
	DebugFlag     bool
)

// ConfigureFlags initialises flags for the utility library.
func ConfigureFlags(app *kingpin.Application) *kingpin.Application {
	logLevelFlag = LogLevelFlagParser(app.Flag("log-level", "Set the default log level.").Default("info"))
	logFileFlag = app.Flag("log-file", "Enable file logging to PATH.").PlaceHolder("PATH").OpenFile(os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	logStderrFlag = app.Flag("log-stderr", "Log to stderr (defaults to true).").Default("true").Bool()
	app.Flag("debug", "Enable debug mode.").BoolVar(&DebugFlag)
	return app
}
