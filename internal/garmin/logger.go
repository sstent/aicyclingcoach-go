package garmin

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Logger defines the interface for logging in Garmin operations
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// CLILogger implements Logger for CLI output
type CLILogger struct{}

func (l *CLILogger) Debugf(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *CLILogger) Infof(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *CLILogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func (l *CLILogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// NoopLogger implements Logger that does nothing
type NoopLogger struct{}

func (l *NoopLogger) Debugf(format string, args ...interface{}) {}
func (l *NoopLogger) Infof(format string, args ...interface{})  {}
func (l *NoopLogger) Warnf(format string, args ...interface{})  {}
func (l *NoopLogger) Errorf(format string, args ...interface{}) {}

// FileLogger implements Logger that writes to a file
type FileLogger struct {
	logger *log.Logger
	file   *os.File
}

func NewFileLogger(logPath string) (*FileLogger, error) {
	// Create log directory if it doesn't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(file, "", log.LstdFlags)
	return &FileLogger{
		logger: logger,
		file:   file,
	}, nil
}

func (l *FileLogger) Debugf(format string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+format, args...)
}

func (l *FileLogger) Infof(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

func (l *FileLogger) Warnf(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

func (l *FileLogger) Errorf(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

func (l *FileLogger) Close() error {
	return l.file.Close()
}
