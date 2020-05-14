package logger

import (
	"github.com/Sirupsen/logrus"
	"sync"
)

var defaultLogger *logrus.Logger
var once sync.Once

type Logger interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

func Get() *logrus.Logger {
	once.Do(func() {
		defaultLogger = logrus.New()
	})

	return defaultLogger
}
