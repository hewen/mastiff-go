package dnsresolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdResolver_Lookup(t *testing.T) {
	mockLookup := func(_ string) ([]string, error) {
		return []string{"1.2.3.4"}, nil
	}

	resolver := NewStdResolver(mockLookup)

	ips, err := resolver.Lookup("example.com")
	assert.Nil(t, err)
	assert.Equal(t, len(ips), 1)
	assert.Equal(t, ips[0], "1.2.3.4")
}

func TestNewStdResolver(_ *testing.T) {
	_ = NewStdResolver(nil)
}
