package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// This is a no-op test to ensure that the main function runs without errors.
	// Actual testing of the CLI commands is done in their respective packages.
	assert.NotPanics(t, func() { main() })
}
