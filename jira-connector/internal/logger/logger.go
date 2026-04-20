package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Instance *logrus.Logger

func InitLogger() {
	Instance = logrus.New()

	Instance.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		ForceColors:     true,
	})

	Instance.SetOutput(os.Stdout)
	Instance.SetLevel(logrus.InfoLevel)
}
