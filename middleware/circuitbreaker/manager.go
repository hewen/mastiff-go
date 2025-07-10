package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// Manager is a circuit breaker manager.
type Manager struct {
	breakers sync.Map // map[string]*gobreaker.CircuitBreaker
	config   *Config
}

// NewManager creates a new circuit breaker manager.
func NewManager(cfg *Config) *Manager {
	return &Manager{config: cfg}
}

// Get returns a circuit breaker by name.
func (m *Manager) Get(name string) *gobreaker.CircuitBreaker {
	if cb, ok := m.breakers.Load(name); ok {
		return cb.(*gobreaker.CircuitBreaker)
	}

	st := gobreaker.Settings{
		Name:        name,
		MaxRequests: m.config.MaxRequests,
		Interval:    time.Duration(m.config.Interval) * time.Second,
		Timeout:     time.Duration(m.config.Timeout) * time.Second,
		ReadyToTrip: m.config.ReadyToTrip,
	}
	cb := gobreaker.NewCircuitBreaker(st)
	m.breakers.Store(name, cb)
	return cb
}

// Break breaks a circuit breaker.
func (m *Manager) Break(name string, times int) {
	cb := m.Get(name)
	for i := 0; i < times; i++ {
		_, _ = cb.Execute(func() (interface{}, error) {
			return nil, errors.New("test fail")
		})
	}
}
