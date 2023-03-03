package log

import (
	"fmt"

	"github.com/charmbracelet/log"
)

func Info(msg interface{}, keyvals ...interface{}) {
	log.Info(msg, keyvals...)
}

func Infof(format string, args ...any) {
	log.Info(fmt.Sprintf(format, args...))
}

func Error(msg interface{}, keyvals ...interface{}) {
	log.Error(msg, keyvals...)
}

func Errorf(format string, args ...any) {
	log.Error(fmt.Sprintf(format, args...))
}

func Warn(msg interface{}, keyvals ...interface{}) {
	log.Warn(msg, keyvals...)
}

func Warnf(format string, args ...any) {
	log.Warn(fmt.Sprintf(format, args...))
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	log.Fatal(msg, keyvals...)
}

func Fatalf(format string, args ...any) {
	log.Fatal(fmt.Sprintf(format, args...))
}
