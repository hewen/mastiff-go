package timeout

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestTimeoutInterceptor(t *testing.T) {
	fn := UnaryTimeoutInterceptor(time.Second)

	ctx := context.TODO()
	_, err := fn(ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: "test",
	}, func(_ context.Context, _ any) (any, error) {
		return nil, nil
	})
	assert.Nil(t, err)
}
