// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

import "github.com/sony/gobreaker"

// Config defines breaker settings.
type Config struct {
	Name        string
	MaxRequests uint32
	Interval    int64 // Seconds
	Timeout     int64 // Seconds
	ReadyToTrip func(counts gobreaker.Counts) bool
}
