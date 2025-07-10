package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Init инициализирует глобальный логгер в JSON формате
func Init() *logrus.Logger {
	logger := logrus.New()

	// Устанавливаем JSON формат
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})

	// Устанавливаем уровень логирования из переменной окружения
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Выводим в stdout
	logger.SetOutput(os.Stdout)

	return logger
}

// GetGinWriter возвращает writer для Gin в JSON формате
func GetGinWriter() *logrus.Logger {
	return Init()
}
