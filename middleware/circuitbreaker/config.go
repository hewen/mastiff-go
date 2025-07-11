// Package circuitbreaker provides a circuit breaker middleware for Gin.
package circuitbreaker

// Config defines circuit breaker configuration.
type Config struct {
	// Policy configuration
	Policy *PolicyConfig
	// Interval in seconds
	Interval int64
	// Timeout in seconds
	Timeout int64
	// Maximum number of requests
	MaxRequests uint32
}

// PolicyConfig defines policy configuration loaded from YAML or code.
type PolicyConfig struct {
	// Type: "consecutive_failures" | "failure_rate"
	Type string
	// Continuous failure threshold (for consecutive_failures)
	ConsecutiveFailures uint32
	// Minimum number of requests (for failure_rate)
	MinRequests uint32
	// Failure rate threshold (for failure_rate)
	FailureRateThreshold float64
}
