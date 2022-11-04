package logger

import (
	"io"
	"wizapp/sdk/app"

	"github.com/sirupsen/logrus"
)

type (
	// Logger is the fundamental interface for all logger operations.
	// We abstract over logrus.FieldLogger to be agnostic to the underlying logging library.
	Logger interface {
		logrus.FieldLogger
	}

	// Config is the configuration for the logger.
	Config struct {
		// The default logging level.
		Level string `mapstructure:"level"`
		// The formatter to use (default: text, options: text, json, stackdriver).
		Formatter string `mapstructure:"formatter"`
		// The output to use (default: stdout, options: stdout, stderr, file).
		Output io.Writer `mapstructure:"output"`
	}
)

// logger is a global variable that provides central logging capabilities.
var logger Logger
var level string

// init initializes the global logger if AUTO_INIT is set to true or no logger instance is already present.
// if no configuration is found the default configuration is used.
func init() {

	var cfg *Config
	defaultLogger := false
	_ = app.Config().UnmarshalKey("logger", &cfg)
	if cfg == nil { // use logrus as default logger
		defaultLogger = true
		cfg = &Config{
			Level: "info",
		}
	}

	logger = setupLogger(cfg)
	level = cfg.Level

	if defaultLogger {
		logger.Debug("No logger configuration found. Using default logger.")
	}
	logger.WithFields(map[string]interface{}{"formatter": cfg.Formatter, "level": cfg.Level}).Info("Logger configured")
}

// setupLogger creates a new logrus.Logger with the given configuration
func setupLogger(cfg *Config) Logger {
	l := logrus.New()
	if lvl, err := logrus.ParseLevel(cfg.Level); err == nil {
		l.SetLevel(lvl)
	}

	l.SetReportCaller(false)

	if cfg.Output != nil {
		l.SetOutput(cfg.Output)
	}

	switch cfg.Formatter {
	case "stackdriver":
		l.Formatter = stackdriver()
	case "json":
		l.Formatter = &logrus.JSONFormatter{}
	default:
	}

	return l
}

// ConfigureLogger is a convenience function to configure the logger.
// Calling this function overrides any previous configuration.
func ConfigureLogger(cfg *Config) {
	if logger != nil {
		logrus.Warn("Overriding previous logger configuration.")
	}
	logger = setupLogger(cfg)
}

// Log is a getter for the logger.
// You can use it in your app or library by simply calling it.
// Example:
//
//	import "gitlab.com/akordacorp/go-commons/logger"
//	logger.Log().Info("Hello World")"
func Log() Logger {
	return logger
}

// Level the current logging level
//
// Values: panic, fatal, error, warn, info, debug, trace
func Level() string {
	return level
}

// LoggingCategory represent a logging category.
type LoggingCategory int

const (
	Service LoggingCategory = iota
	Util
	Controller
	Domain
)

func (s LoggingCategory) String() string {
	switch s {
	case Service:
		return "service"
	case Util:
		return "util"
	case Controller:
		return "controller"
	case Domain:
		return "domain"
	}
	return "unknown"
}

// ServiceLogger returns a logger for the given service.
func ServiceLogger() Logger {
	return CategoryLogger(Service)
}

// DomainLogger is a getter for the domain logger.
func DomainLogger() Logger {
	return CategoryLogger(Domain)
}

// UtilLogger is a utility logger.
func UtilLogger() Logger {
	return CategoryLogger(Util)
}

// ControllerLogger is a convenience function to get a logger for the controller.
func ControllerLogger() Logger {
	return CategoryLogger(Controller)
}

// CategoryLogger creates a new logger with key-value pairs for the given category.
func CategoryLogger(category LoggingCategory) Logger {
	return Log().WithField("category", category.String())
}

// PackageLogger creates a new logger with key-value pairs for the given package.
func PackageLogger(packageName string) Logger {
	return Log().WithField("package", packageName)
}
