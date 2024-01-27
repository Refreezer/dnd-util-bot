package logging

import (
	"github.com/op/go-logging"
	"os"
)

const (
	module = "cherry-check"
)

// Logger is the logger instance for the cherry-check module.
var Logger = logging.MustGetLogger(module)

// InitGlobalLogger initializes the logger with the specified log level.
// If debug is true, the log level is set to DEBUG; otherwise, it is set to INFO.
func InitGlobalLogger(debug bool) {
	InitLogger(debug, Logger)
}

func InitLogger(debug bool, logger *logging.Logger) {
	var logLevel logging.Level
	if debug {
		logLevel = logging.DEBUG
	} else {
		logLevel = logging.INFO
	}

	backend := logging.NewLogBackend(os.Stdout, "", 0)
	var format = logging.MustStringFormatter(
		`%{time:15:04:05.000} ▶ %{level} %{message}`,
	)
	formatter := logging.NewBackendFormatter(backend, format)
	leveled := logging.AddModuleLevel(formatter)
	leveled.SetLevel(logLevel, module)
	logger.SetBackend(leveled)
}
