package logger

import (
	"reflect"
	"time"

	masker "github.com/ggwhite/go-masker/v2"
)

var (
	marshaler = masker.NewMaskerMarshaler()

	// enableMasking is a flag to enable masking of sensitive fields in log entries.
	enableMasking bool
)

// LogRequest logs an HTTP or gRPC request in structured format (JSON-style if backend supports).
func LogRequest(l Logger, statusCode int, duration time.Duration, ip, method, ua string, req, resp any, err error) {
	// Type assertion to see if logger backend supports structured logging (e.g. zerologLogger).
	fields := map[string]any{
		"status":   statusCode,
		"duration": duration.String(),
		"ip":       ip,
		"method":   method,
		"ua":       ua,
		"req":      MaskValue(req),
		"resp":     MaskValue(resp),
	}
	if err != nil {
		fields["err"] = err.Error()
	}

	entry := l.Fields(fields)
	switch {
	case err != nil:
		entry.Errorf("req")
	case duration > time.Second:
		entry.Infof("slow req")
	default:
		entry.Infof("req")
	}
}

// SetLogMasking sets the global log masking flag.
func SetLogMasking(enable bool) {
	enableMasking = enable
}

// MaskValue masks sensitive fields in a struct.
func MaskValue(v any) any {
	if !enableMasking || v == nil {
		return v
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Struct, reflect.Ptr:
		m, err := marshaler.Struct(v)
		if err != nil {
			return v
		}
		return m
	default:
		return v
	}
}
