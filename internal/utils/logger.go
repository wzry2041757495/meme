package utils

import (
	"fmt"
	"os"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var (
	// CurrentLogLevel 当前日志级别，可通过环境变量 LOG_LEVEL 设置
	CurrentLogLevel = LogLevelInfo
)

func init() {
	// 从环境变量读取日志级别
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug", "DEBUG":
		CurrentLogLevel = LogLevelDebug
	case "info", "INFO":
		CurrentLogLevel = LogLevelInfo
	case "warn", "WARN":
		CurrentLogLevel = LogLevelWarn
	case "error", "ERROR":
		CurrentLogLevel = LogLevelError
	}
}

func log(level LogLevel, prefix string, format string, args ...interface{}) {
	if level < CurrentLogLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s] %s %s\n", timestamp, prefix, msg)
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	log(LogLevelDebug, "🔍 DEBUG", format, args...)
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	log(LogLevelInfo, "ℹ️  INFO", format, args...)
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	log(LogLevelWarn, "⚠️  WARN", format, args...)
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	log(LogLevelError, "❌ ERROR", format, args...)
}

// Request 请求日志 (特殊格式)
func Request(method, url string) {
	log(LogLevelDebug, "🌐 REQ", "%s %s", method, url)
}

// Response 响应日志 (特殊格式)
func Response(status int, duration time.Duration, size int) {
	log(LogLevelDebug, "📥 RES", "status=%d duration=%v size=%d bytes", status, duration, size)
}
