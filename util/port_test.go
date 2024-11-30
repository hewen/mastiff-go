package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFreePort(t *testing.T) {
	_, err := GetFreePort()
	assert.Nil(t, err)
}
