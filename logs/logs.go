package logs

import (
	"fmt"
	"os"

	logs "github.com/withmandala/go-log"
)

type Level = int

const (
	Debug Level = iota
	Info
	Warning
	Error
)

type Logger struct {
	Level  Level
	logger *logs.Logger
}

func New(level Level) *Logger {
	logger := logs.New(os.Stderr).WithDebug().WithColor()
	return &Logger{Level: level, logger: logger}
}

var Default = New(Info)

func (l Logger) Debug(format string, a ...interface{}) {
	if l.Level <= Debug && l.logger != nil {
		l.logger.Debugf("[DEBUG] %s", fmt.Sprintf(format, a...))
	}
}

func (l Logger) Info(format string, a ...interface{}) {
	if l.Level <= Info && l.logger != nil {
		l.logger.Infof("[INFO] %s", fmt.Sprintf(format, a...))
	}
}

func (l Logger) Warning(format string, a ...interface{}) {
	if l.Level <= Warning && l.logger != nil {
		l.logger.Warnf("[WARN] %s", fmt.Sprintf(format, a...))
	}
}

func (l Logger) Error(format string, a ...interface{}) {
	l.logger.Errorf("[ERROR] %s", fmt.Sprintf(format, a...))
}
