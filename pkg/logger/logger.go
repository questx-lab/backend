package logger

import (
	"log"
)

const (
	DEBUG int = iota
	INFO
	WARNING
	ERROR
	SILENCE
)

type Logger interface {
	Debugf(msg string, a ...any)
	Infof(msg string, a ...any)
	Warnf(msg string, a ...any)
	Errorf(msg string, a ...any)
}

type defaultLogger struct {
	level int
}

func NewLogger(level int) *defaultLogger {
	return &defaultLogger{level: level}
}

func (l *defaultLogger) Debugf(msg string, a ...any) {
	if l.level <= DEBUG {
		log.Printf(msg+"\n", a...)
	}
}

func (l *defaultLogger) Infof(msg string, a ...any) {
	if l.level <= INFO {
		log.Printf(msg+"\n", a...)
	}
}

func (l *defaultLogger) Warnf(msg string, a ...any) {
	if l.level <= WARNING {
		log.Printf(msg+"\n", a...)
	}
}

func (l *defaultLogger) Errorf(msg string, a ...any) {
	if l.level <= ERROR {
		log.Printf(msg+"\n", a...)
	}
}
