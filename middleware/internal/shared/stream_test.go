package shared

import (
	"context"
	"testing"

	"github.com/hewen/mastiff-go/pkg/contextkeys"
	"github.com/stretchr/testify/assert"
)

func TestGrpcServerStream_Context(t *testing.T) {
	expectedCtx := contextkeys.SetValue(context.Background(), "key", "value")
	stream := &GrpcServerStream{
		Ctx: expectedCtx,
	}
	actualCtx := stream.Context()
	assert.Equal(t, expectedCtx, actualCtx)
	value, ok := contextkeys.GetValue[string](actualCtx, "key")
	assert.Equal(t, true, ok)
	assert.Equal(t, "value", value)
}
