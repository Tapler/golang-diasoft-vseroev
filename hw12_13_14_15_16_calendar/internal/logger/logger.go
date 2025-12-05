package logger

import (
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

func New(level string) *Logger {
	logLevel := parseLevel(level)
	return &Logger{
		level:  logLevel,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func parseLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l *Logger) Debug(msg string) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] %s", msg)
	}
}

func (l *Logger) Info(msg string) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] %s", msg)
	}
}

func (l *Logger) Warn(msg string) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] %s", msg)
	}
}

func (l *Logger) Error(msg string) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] %s", msg)
	}
}
