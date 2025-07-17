package decorators

import (
	"log"
	"sync"
)

// LogLevel defines logging level
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelNormal
	LogLevelVerbose
)

// Logger controla o logging do framework
type Logger struct {
	level LogLevel
	mu    sync.RWMutex
}

var globalLogger = &Logger{level: LogLevelNormal}

// SetLogLevel defines logging level globally
func SetLogLevel(level LogLevel) {
	globalLogger.mu.Lock()
	defer globalLogger.mu.Unlock()
	globalLogger.level = level
}

// GetLogLevel returns current logging level
func GetLogLevel() LogLevel {
	globalLogger.mu.RLock()
	defer globalLogger.mu.RUnlock()
	return globalLogger.level
}

// SetVerbose ativa/desativa logs verbose
func SetVerbose(verbose bool) {
	if verbose {
		SetLogLevel(LogLevelVerbose)
	} else {
		SetLogLevel(LogLevelNormal)
	}
}

// LogVerbose imprime log apenas em modo verbose
func LogVerbose(format string, args ...interface{}) {
	if GetLogLevel() >= LogLevelVerbose {
		log.Printf(format, args...)
	}
}

// LogNormal imprime log em modo normal e verbose
func LogNormal(format string, args ...interface{}) {
	if GetLogLevel() >= LogLevelNormal {
		log.Printf(format, args...)
	}
}

// LogSilent always prints log (used for important errors)
func LogSilent(format string, args ...interface{}) {
	log.Printf(format, args...)
}
