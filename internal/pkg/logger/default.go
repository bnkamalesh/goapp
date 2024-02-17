package logger

import "context"

var defaultLogger = New("", "", 0, nil)

// Info is for logging items with severity 'info'
func Info(ctx context.Context, payload ...interface{}) error {
	return defaultLogger.log(LogTypeInfo, payload...)
}

// Warn is for logging items with severity 'Warn'
func Warn(ctx context.Context, payload ...interface{}) error {
	return defaultLogger.log(LogTypeWarn, payload...)
}

// Error is for logging items with severity 'Error'
func Error(ctx context.Context, payload ...interface{}) error {
	return defaultLogger.log(LogTypeError, payload...)
}

// Fatal is for logging items with severity 'Fatal'
func Fatal(ctx context.Context, payload ...interface{}) error {
	return defaultLogger.log(LogTypeFatal, payload...)
}

// UpdateDefaultLogger resets the default logger
func UpdateDefaultLogger(lh *LogHandler) {
	defaultLogger = lh
}
