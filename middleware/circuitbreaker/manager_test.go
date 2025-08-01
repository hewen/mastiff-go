package circuitbreaker

import (
	"testing"

	"github.com/hewen/mastiff-go/config/middlewareconf/circuitbreakerconf"
	"github.com/stretchr/testify/assert"
)

func TestNewManagerAndGet(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 3,
		Interval:    1,
		Timeout:     1,
	}

	mgr := NewManager(cfg)
	cb := mgr.Get("test-path")

	assert.NotNil(t, cb)
	assert.Equal(t, "test-path", cb.Name())
}

func TestGetReuseBreaker(t *testing.T) {
	cfg := &circuitbreakerconf.Config{
		MaxRequests: 3,
		Interval:    1,
		Timeout:     1,
	}

	mgr := NewManager(cfg)
	cb1 := mgr.Get("same")
	cb2 := mgr.Get("same")

	assert.Same(t, cb1, cb2)
}
