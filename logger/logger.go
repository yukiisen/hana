// Package logger is a minimal leveled logger built on the standard library.
package logger

import (
	"log"
	"os"
)

// Level controls which log lines are emitted.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger is a leveled logger. Lines are tagged like `[Info] message` so they
// match the existing log output.
type Logger struct {
	level Level
	scope string
	out   *log.Logger
}

// New returns a logger that writes to stderr at the given minimum level.
func New(level Level) *Logger {
	return &Logger{
		level: level,
		out:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

// WithScope returns a logger whose messages are prefixed with name.
func (l *Logger) WithScope(name string) *Logger {
	return &Logger{level: l.level, scope: name, out: l.out}
}

func (l *Logger) logf(level Level, tag, format string, args ...any) {
	if level < l.level {
		return
	}
	prefix := "[" + tag + "]"
	// if l.scope != "" {
	// 	prefix += "[" + l.scope + "]"
	// }
	l.out.Printf(prefix+" "+format+"\n", args...)
}

func (l *Logger) Debugf(format string, args ...any) { l.logf(LevelDebug, "Debug", format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.logf(LevelInfo, "Info", format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.logf(LevelWarn, "Warn", format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.logf(LevelError, "Error", format, args...) }

// Default logger used by the package-level helpers.
var defaultLogger = New(LevelInfo)

// SetLevel changes the level of the default logger.
func SetLevel(level Level) { defaultLogger.level = level }

func Debugf(format string, args ...any) { defaultLogger.Debugf(format, args...) }
func Infof(format string, args ...any)  { defaultLogger.Infof(format, args...) }
func Warnf(format string, args ...any)  { defaultLogger.Warnf(format, args...) }
func Errorf(format string, args ...any) { defaultLogger.Errorf(format, args...) }
