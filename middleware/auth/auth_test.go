package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWhiteListed(t *testing.T) {
	assert.True(t, isWhiteListed("/public", []string{"/public"}))

	assert.True(t, isWhiteListed("/api/v1/user", []string{"/api/"}))

	assert.False(t, isWhiteListed("/user", []string{"/admin"}))

	assert.True(t, isWhiteListed("/healthz", []string{"/metrics", "/healthz", "/debug/"}))

	assert.False(t, isWhiteListed("/foo", []string{}))
}
