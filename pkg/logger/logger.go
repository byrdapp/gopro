package logger

import (
	"github.com/sirupsen/logrus"
)

// NewLogger -
func NewLogger() *logrus.Logger {
	logger := logrus.StandardLogger()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:    true,
		FullTimestamp:  true,
		DisableSorting: true,
	})
	return logger
}
