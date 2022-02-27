package logs

import (
	"fmt"
	"log"
)

type Level = int

const (
	Debug Level = iota
	Info
	Warning
	Error
)

type Logger struct {
	Level Level
}

func (l Logger) Debug(format string, a ...interface{}) {
	if l.Level <= Debug {
		log.Printf("[DEBUG] %s", fmt.Sprintf(format, a...))
	}
}

func (l Logger) Error(format string, a ...interface{}) {
	log.Printf("[ERROR] %s", fmt.Sprintf(format, a...))
}
