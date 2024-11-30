package store

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestInitRedis(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	_, err = InitRedis(RedisConf{
		Addr: s.Addr(),
	})
	assert.Nil(t, err)
}
