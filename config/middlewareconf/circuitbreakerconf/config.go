// Package circuitbreakerconf provides a circuit breaker middleware for Gin.
package circuitbreakerconf

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

// ApplyDefaults sets default values if missing.
func (cfg *PolicyConfig) ApplyDefaults() {
	if cfg.Type == "" {
		cfg.Type = "failure_rate"
	}
	if cfg.ConsecutiveFailures == 0 {
		cfg.ConsecutiveFailures = defaultConsecutiveFailures
	}
	if cfg.MinRequests == 0 {
		cfg.MinRequests = defaultMinRequests
	}
	if cfg.FailureRateThreshold <= 0 {
		cfg.FailureRateThreshold = defaultFailureRateThreshold
	}
}

const (
	defaultConsecutiveFailures  = 5   // Default number of consecutive failures to trip the circuit
	defaultMinRequests          = 10  // Default minimum number of requests before tripping the circuit
	defaultFailureRateThreshold = 0.5 // Default failure rate threshold
)
