package logger

import (
	"log"
)

type Logger interface {
	Debugf(msg string, a ...any)
	Infof(msg string, a ...any)
	Warnf(msg string, a ...any)
	Errorf(msg string, a ...any)
}

type defaultLogger struct{}

func NewLogger() *defaultLogger {
	return &defaultLogger{}
}

func (l *defaultLogger) Debugf(msg string, a ...any) {
	log.Printf(msg+"\n", a...)
}

func (l *defaultLogger) Infof(msg string, a ...any) {
	log.Printf(msg+"\n", a...)
}

func (l *defaultLogger) Warnf(msg string, a ...any) {
	log.Printf(msg+"\n", a...)
}

func (l *defaultLogger) Errorf(msg string, a ...any) {
	log.Printf(msg+"\n", a...)
}
