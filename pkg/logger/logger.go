// Package logger 提供分级日志记录功能
// Package logger provides hierarchical logging functionality
package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger 日志记录器结构体
// Logger structure for logging
type Logger struct {
	*logrus.Logger
}

// LogLevel 日志级别
// LogLevel enumeration
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
)

var (
	// DefaultLogger 默认日志记录器
	DefaultLogger *Logger
)

func init() {
	DefaultLogger = NewLogger()
}

// NewLogger 创建新的日志记录器
// NewLogger creates a new logger instance
func NewLogger() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	return &Logger{Logger: log}
}

// SetLevel 设置日志级别
// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		l.Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.Logger.SetLevel(logrus.ErrorLevel)
	}
}

// SetOutput 设置日志输出目标
// SetOutput sets the log output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.Logger.SetOutput(output)
}

// Debug 记录调试级别日志
// Debug logs a debug level message
func Debug(args ...interface{}) {
	DefaultLogger.Logger.Debug(args...)
}

// Debugf 记录格式化调试级别日志
// Debugf logs a formatted debug level message
func Debugf(format string, args ...interface{}) {
	DefaultLogger.Logger.Debugf(format, args...)
}

// Info 记录信息级别日志
// Info logs an info level message
func Info(args ...interface{}) {
	DefaultLogger.Logger.Info(args...)
}

// Infof 记录格式化信息级别日志
// Infof logs a formatted info level message
func Infof(format string, args ...interface{}) {
	DefaultLogger.Logger.Infof(format, args...)
}

// Warn 记录警告级别日志
// Warn logs a warning level message
func Warn(args ...interface{}) {
	DefaultLogger.Logger.Warn(args...)
}

// Warnf 记录格式化警告级别日志
// Warnf logs a formatted warning level message
func Warnf(format string, args ...interface{}) {
	DefaultLogger.Logger.Warnf(format, args...)
}

// Error 记录错误级别日志
// Error logs an error level message
func Error(args ...interface{}) {
	DefaultLogger.Logger.Error(args...)
}

// Errorf 记录格式化错误级别日志
// Errorf logs a formatted error level message
func Errorf(format string, args ...interface{}) {
	DefaultLogger.Logger.Errorf(format, args...)
}

// Fatal 记录致命错误并退出程序
// Fatal logs a fatal error and exits the program
func Fatal(args ...interface{}) {
	DefaultLogger.Logger.Fatal(args...)
}

// Fatalf 记录格式化致命错误并退出程序
// Fatalf logs a formatted fatal error and exits the program
func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Logger.Fatalf(format, args...)
}
