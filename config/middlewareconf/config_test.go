package middlewareconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDefaults(t *testing.T) {
	var conf Config
	conf.SetDefaults()
	assert.Equal(t, true, *conf.EnableRecovery)
	assert.Equal(t, 30, *conf.TimeoutSeconds)
}
