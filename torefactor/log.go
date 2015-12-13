package torefactor

import (
	"github.com/fatih/color"
	//logpkg "log"
)

type logLevel string

const (
	Trace logLevel = "trace"
	Info           = "info"
	Warn           = "warn"
	Error          = "error"
)

var (
	L        *log
	LogLevel logLevel = Trace
)

type log struct {
	//logpkg.Logger
}

func (l *log) Trace(f string, a ...interface{}) {
	color.Cyan(f, a...)
}

func (l *log) Info(f string, a ...interface{}) {
	color.Green(f, a...)
}

func (l *log) Warn(f string, a ...interface{}) {
	color.Yellow(f, a...)
}

func (l *log) Error(f string, a ...interface{}) {
	color.Red(f, a...)
}
