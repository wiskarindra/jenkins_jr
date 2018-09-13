// Package log is used to write logs to stdout.
package log

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

// Field holds key-value to be written to log.
type Field struct {
	Key   string
	Value interface{}
}

// NewField returns Field with given key and value.
func NewField(key string, value interface{}) Field {
	return Field{key, value}
}

// RequestInfo writes log with severity = info.
// It will only write log if mandatory fields are given.
// Otherwise, it will return an error.
func RequestInfo(message string, fields ...Field) error {
	if zapFields, proceed := convertAndCheckFields(fields...); proceed && message != "" {
		logger.Info(message, zapFields...)
		return nil
	}
	return errors.New("fields don't contain all mandatory fields")
}

// RequestError writes log with severity = error.
// It will only write log if mandatory fields are given.
// Otherwise, it will return an error.
func RequestError(message string, fields ...Field) error {
	if zapFields, proceed := convertAndCheckFields(fields...); proceed && message != "" {
		logger.Error(message, zapFields...)
		return nil
	}
	return errors.New("fields don't contain all mandatory fields")
}

func convertAndCheckFields(fields ...Field) ([]zapcore.Field, bool) {
	var flag uint
	var zapFields []zapcore.Field

	for _, field := range fields {
		if field.Key == "request_id" {
			flag = flag | 1
		} else if field.Key == "tags" {
			flag = flag | 2
		} else if field.Key == "duration" {
			flag = flag | 4
		}

		zapFields = append(zapFields, zap.Any(field.Key, field.Value))

	}
	return zapFields, flag == 7
}
