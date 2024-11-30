package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitMockRedis(t *testing.T) {
	db := InitMockRedis()
	assert.NotNil(t, db)
	db.Set("test", 1, time.Hour)
	db.Get("test")
}
