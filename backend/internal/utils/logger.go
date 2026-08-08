package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the application-wide structured logger.
// Initialized once at package import time via init().
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
	})
	Logger.SetLevel(logrus.InfoLevel)
}
