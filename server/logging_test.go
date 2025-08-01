package server

import (
	"testing"

	"github.com/hewen/mastiff-go/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogggingServers(t *testing.T) {
	servers := NewServers(nil)
	ms := &MockServers{}
	ls := &LoggingServer{
		Inner:  ms,
		Logger: logger.NewLogger(),
	}
	assert.Equal(t, ls.Name(), ms.Name())
	ls.WithLogger(logger.NewLogger())
	servers.Add(ls)
	servers.Start()
	servers.Stop()
}
