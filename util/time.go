package util

import (
	"time"
)

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
