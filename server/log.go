package server

import (
	"fmt"
	"time"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/util"
)

// LogRequest logs a request.
func LogRequest(l logger.Logger, statusCode int, duration time.Duration, ip, method, ua, req, resp string, err error) {
	msg := fmt.Sprintf("%3d | %10s | %-15s | %-30s | UA: %s",
		statusCode,
		util.FormatDuration(duration),
		ip,
		method,
		ua,
	)

	if req != "" || resp != "" {
		msg += fmt.Sprintf(" | req: %s", req)
		msg += fmt.Sprintf(" | resp: %s", resp)
	}
	if err != nil {
		msg += fmt.Sprintf(" | err: %v", err)
	}

	switch {
	case err != nil:
		l.Errorf(msg)
	case duration > time.Second:
		l.Infof("[SLOW] " + msg)
	default:
		l.Infof(msg)
	}
}
