package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Instance = logrus.New()

func InitLogger() {
	Instance.SetFormatter(&logrus.TextFormatter{ // перед продакшеном нужно будет установить на json
		FullTimestamp: true,
		ForceColors:   true,
	})

	Instance.SetOutput(os.Stdout)
	Instance.SetLevel(logrus.InfoLevel)
}
