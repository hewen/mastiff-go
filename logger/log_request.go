package logger

import (
	"time"
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
		"req":      req,
		"resp":     resp,
	}
	if err != nil {
		fields["err"] = err.Error()
	}

	l.Fields(fields)
	switch {
	case err != nil:
		l.Errorf("req")
	case duration > time.Second:
		l.Infof("slow req")
	default:
		l.Infof("req")
	}
}
