// logger/logger.go
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

// InitLogger initializes the global logger with settings
func InitLogger() *logrus.Logger {
	// Create logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			log.Fatalf("Failed to create logs directory: %v", err)
		}
	}

	log = logrus.New()

	// Set the format to JSON or text as per preference
	log.SetFormatter(&logrus.JSONFormatter{})

	// Set the log level, can be configured via environment or as required
	log.SetLevel(logrus.InfoLevel)

	// Configure rotating file with lumberjack
	log.SetOutput(&lumberjack.Logger{
		Filename:   "logs/server.log",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     30,   // days
		Compress:   true, // compress old log files
	})

	return log
}

// GetLogger returns the configured logger
func GetLogger() *logrus.Logger {
	if log == nil {
		return InitLogger()
	}
	return log
}
