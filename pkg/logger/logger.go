package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	cfg       *LoggerConfig
	clsLogger *log.Logger
	// fileLogger *log.Logger
}

type LoggerConfig struct {
	Name  string
	Level LogLevel
}
type LogLevel uint

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

const LOG_LEVEL_TRACE = "TRACE"
const LOG_LEVEL_DEBUG = "DEBUG"
const LOG_LEVEL_INFO = "INFO"
const LOG_LEVEL_WARN = "WARNING"
const LOG_LEVEL_ERROR = "ERROR"
const LOG_LEVEL_FATAL = "FATAL"

var (
	logNameColor = "\x1b[35m\x1b[1m%s\x1b[0m"
	resetColor   = "\x1b[0m"
	flagColor    = "\x1b[234m\x1b[1m"
	debugColor   = "\x1b[43m\x1b[4m\x1b[1m"
	infoColor    = "\x1b[32m\x1b[1m"
	warnColor    = "\x1b[33m\x1b[1m"
	errorColor   = "\x1b[31m\x1b[1m"
	fatalColor   = "\x1b[41m\x1b[1m"
	unknownColor = "\x1b[37m%s\x1b[0m"
)

func GetLogger(cfg *LoggerConfig) *Logger {
	return &Logger{
		cfg:       cfg,
		clsLogger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

func getLogLevelStr(level LogLevel) string {
	switch level {
	case TRACE:
		return LOG_LEVEL_TRACE
	case DEBUG:
		return fmt.Sprintf("%s[%s]%s", debugColor, LOG_LEVEL_DEBUG, resetColor)
	case INFO:
		return fmt.Sprintf("%s[%s]%s", infoColor, LOG_LEVEL_INFO, resetColor)
	case WARN:
		return fmt.Sprintf("%s[%s]%s", warnColor, LOG_LEVEL_WARN, resetColor)
	case ERROR:
		return fmt.Sprintf("%s[%s]%s", errorColor, LOG_LEVEL_ERROR, resetColor)
	case FATAL:
		return fmt.Sprintf("%s[%s]%s", fatalColor, LOG_LEVEL_FATAL, resetColor)
	default:
		return fmt.Sprintf("%s[%s]%s", unknownColor, "unknown", resetColor)
	}
}

func (l *Logger) executeLog(logger *log.Logger, msg string, level LogLevel, color string) error {
	if l.cfg.Level >= level {
		fmt.Println(l.cfg.Level, level)
		return nil
	}
	logger.SetPrefix(flagColor)
	logNameColor := fmt.Sprintf(logNameColor, "["+l.cfg.Name+"]")
	logger.Printf("%s%s%s: %s", resetColor, logNameColor, getLogLevelStr(level), msg)

	return nil
}

func (l *Logger) Trace(msg string) error {
	return l.executeLog(l.clsLogger, msg, TRACE, unknownColor)
}

func (l *Logger) Debug(msg string) error {
	return l.executeLog(l.clsLogger, msg, DEBUG, debugColor)
}

func (l *Logger) Info(msg string) error {
	return l.executeLog(l.clsLogger, msg, INFO, infoColor)
}

func (l *Logger) Warn(msg string) error {
	return l.executeLog(l.clsLogger, msg, WARN, warnColor)
}

func (l *Logger) Error(msg string) error {
	return l.executeLog(l.clsLogger, msg, ERROR, errorColor)
}

func (l *Logger) Fatal(msg string) error {
	return l.executeLog(l.clsLogger, msg, FATAL, fatalColor)
}
