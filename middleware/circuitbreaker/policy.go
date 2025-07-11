package circuitbreaker

import (
	"github.com/sony/gobreaker"
)

const (
	defaultConsecutiveFailures  = 5   // Default number of consecutive failures to trip the circuit
	defaultMinRequests          = 10  // Default minimum number of requests before tripping the circuit
	defaultFailureRateThreshold = 0.5 // Default failure rate threshold
)

// Policy defines a circuit breaker policy interface.
type Policy interface {
	ShouldTrip(counts gobreaker.Counts) bool
}

// ConsecutiveFailuresPolicy checks if the number of consecutive failures exceeds the threshold.
type ConsecutiveFailuresPolicy struct {
	ConsecutiveFailures uint32
}

// ShouldTrip checks if the number of consecutive failures exceeds the threshold.
func (p *ConsecutiveFailuresPolicy) ShouldTrip(counts gobreaker.Counts) bool {
	return counts.ConsecutiveFailures >= p.ConsecutiveFailures
}

// FailureRatePolicy checks if the failure rate exceeds the threshold.
type FailureRatePolicy struct {
	MinRequests          uint32
	FailureRateThreshold float64
}

// ShouldTrip checks if the failure rate exceeds the threshold.
func (p *FailureRatePolicy) ShouldTrip(counts gobreaker.Counts) bool {
	if counts.Requests < p.MinRequests {
		return false
	}
	return float64(counts.TotalFailures)/float64(counts.Requests) >= p.FailureRateThreshold
}

// applyDefaults sets default values if missing.
func (cfg *PolicyConfig) applyDefaults() {
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

// NewPolicyFromConfig creates a Policy from config with defaults.
func NewPolicyFromConfig(cfg *PolicyConfig) Policy {
	if cfg == nil {
		cfg = &PolicyConfig{}
	}
	cfg.applyDefaults()

	switch cfg.Type {
	case "consecutive_failures":
		return &ConsecutiveFailuresPolicy{ConsecutiveFailures: cfg.ConsecutiveFailures}
	case "failure_rate":
		return &FailureRatePolicy{MinRequests: cfg.MinRequests, FailureRateThreshold: cfg.FailureRateThreshold}
	default:
		return &FailureRatePolicy{
			MinRequests:          cfg.MinRequests,
			FailureRateThreshold: cfg.FailureRateThreshold,
		}
	}
}
