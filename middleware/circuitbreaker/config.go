// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

// Config defines circuit breaker configuration.
type Config struct {
	Policy      *PolicyConfig // Policy configuration
	Interval    int64         // Interval in seconds
	Timeout     int64         // Timeout in seconds
	MaxRequests uint32        // Maximum number of requests
}

// PolicyConfig defines policy configuration loaded from YAML or code.
type PolicyConfig struct {
	Type                 string  // Type: "consecutive_failures" | "failure_rate"
	ConsecutiveFailures  uint32  // Continuous failure threshold (for consecutive_failures)
	MinRequests          uint32  // Minimum number of requests (for failure_rate)
	FailureRateThreshold float64 // Failure rate threshold (for failure_rate)
}
