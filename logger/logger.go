// logger/logger.go
package logger

import (
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

type Config struct {
	LogLevel     string
	LogFile      string
	MaxSize      int // megabytes
	MaxBackups   int
	MaxAge       int // days
	Compress     bool
	ReportCaller bool
	JSONFormat   bool
}

func InitLogger(config Config) *logrus.Logger {
	once.Do(func() {
		instance = logrus.New()

		// Set log level
		level, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			level = logrus.InfoLevel
		}
		instance.SetLevel(level)

		// Configure output writers
		writers := []io.Writer{os.Stdout} // Always write to console

		if config.LogFile != "" {
			// Create log directory if it doesn't exist
			logDir := path.Dir(config.LogFile)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				panic(err)
			}

			// Configure log rotation
			fileWriter := &lumberjack.Logger{
				Filename:   config.LogFile,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			}
			writers = append(writers, fileWriter)
		}

		// Set multi-writer
		instance.SetOutput(io.MultiWriter(writers...))

		// Configure formatter
		if config.JSONFormat {
			instance.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
			})
		} else {
			instance.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: time.RFC3339,
			})
		}

		instance.SetReportCaller(config.ReportCaller)
	})

	return instance
}

func GetLogger() *logrus.Logger {
	if instance == nil {
		// Default configuration if not initialized
		InitLogger(Config{
			LogLevel:   "info",
			LogFile:    "logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		})
	}
	return instance
}

// Helper functions for structured logging
func NewPackageLogger(key string) *logrus.Entry {
	return GetLogger().WithField("package", key)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}
