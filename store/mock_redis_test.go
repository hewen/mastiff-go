package store

import (
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitMockRedis(t *testing.T) {
	db := InitMockRedis()
	assert.NotNil(t, db)
	db.Set("test", 1, time.Hour)
	db.Get("test")
}

func TestInitMockRedis_MiniredisFail(t *testing.T) {
	original := miniredisRun
	defer func() { miniredisRun = original }()

	miniredisRun = func() (*miniredis.Miniredis, error) {
		return nil, errors.New("mock miniredis failure")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("InitMockRedis should panic when miniredis fails")
		}
	}()

	_ = InitMockRedis()
}
