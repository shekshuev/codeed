package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
)

// Logger wraps a zap.Logger instance with a singleton pattern.
// It ensures that only one logger is created and reused across the application.
type Logger struct {
	Log *zap.Logger
}

var (
	instance *Logger
	once     sync.Once
)

// NewLogger returns a singleton Logger instance.
// It initializes the logger only once, setting the log level to "info" by default.
func NewLogger() *Logger {
	once.Do(func() {
		instance = &Logger{Log: zap.NewNop()}
		if err := instance.initialize("info"); err != nil {
			log.Fatalf("Error initializing zap logger: %v", err)
		}
	})
	return instance
}

// initialize sets up the zap.Logger configuration based on the provided log level.
func (l *Logger) initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	l.Log = zl
	return nil
}
