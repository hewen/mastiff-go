package contextkeys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	res := LoggerTraceIDKey.String()
	assert.Equal(t, "context key: "+string(LoggerTraceIDKey), res)
}
