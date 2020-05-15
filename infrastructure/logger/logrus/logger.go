package logrus

import (
	"github.com/Sirupsen/logrus"
	"sync"
)

var defaultLogger *logrus.Logger
var once sync.Once

func Get() *logrus.Logger {
	once.Do(func() {
		defaultLogger = logrus.New()
		defaultLogger.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	})

	return defaultLogger
}
