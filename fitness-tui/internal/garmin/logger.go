package garmin

import "fmt"

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
