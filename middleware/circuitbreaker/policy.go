package circuitbreaker

import (
	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/sony/gobreaker"
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

// NewPolicyFromConfig creates a Policy from config with defaults.
func NewPolicyFromConfig(cfg *circuitbreakerconf.PolicyConfig) Policy {
	if cfg == nil {
		cfg = &circuitbreakerconf.PolicyConfig{}
	}
	cfg.ApplyDefaults()

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
