package main

import (
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// colorize wraps s with ANSI escape codes for the given color code.
func colorize(code int, s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, s)
}

// greenTimeEncoder encodes a timestamp in ISO8601 format with green ANSI color,
// matching the Java operator's %green(%d{ISO8601}) Logback pattern.
func greenTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorize(32, t.Format("2006-01-02T15:04:05.000Z0700")))
}

// yellowNameEncoder encodes the logger name in yellow ANSI color,
// matching the Java operator's %yellow(%C{1}) Logback pattern.
func yellowNameEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(colorize(33, name))
}

// coloredConsoleEncoder returns a zap.Opts that configures the console encoder
// with colored output matching the Java operator's Logback configuration.
func coloredConsoleEncoder() zap.Opts {
	return zap.ConsoleEncoder(func(c *zapcore.EncoderConfig) {
		c.EncodeLevel = zapcore.CapitalColorLevelEncoder
		c.EncodeTime = greenTimeEncoder
		c.EncodeName = yellowNameEncoder
	})
}
