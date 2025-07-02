package util

import (
	"time"
)

// FormatDuration returns a human-readable string representation of a time.Duration
// with adaptive precision based on its length:
//
//   - If duration > 10 minutes: rounded to the nearest second
//   - If duration > 100 seconds: rounded to the nearest millisecond
//   - If duration > 10 seconds: rounded to 10-millisecond precision
//   - If duration > 1 millisecond: rounded to the nearest microsecond
//
// The purpose is to avoid excessive precision for large durations and provide
// more detail for smaller ones.
func FormatDuration(d time.Duration) string {
	switch {
	case d > time.Minute*10:
		d = d.Round(time.Second)
	case d > time.Second*100:
		d = d.Round(time.Millisecond)
	case d > time.Second*10:
		d = d.Round(time.Millisecond / 100)
	case d > time.Millisecond:
		d = d.Round(time.Microsecond)
	}
	return d.String()
}
