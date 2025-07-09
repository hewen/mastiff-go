package recovery

import (
	"context"
	"runtime/debug"
	"strings"

	"github.com/hewen/mastiff-go/logger"
	"google.golang.org/grpc"
)

// UnaryRecoveryInterceptor executes the handler and recovers from any panic, logging the error if it occurs.
func UnaryRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		defer func() {
			if r := recover(); r != nil {
				l := logger.NewLoggerWithContext(ctx)
				l.Errorf("panic: %v $%s", r, strings.ReplaceAll(string(debug.Stack()), "\n", "$"))
			}
		}()
		return handler(ctx, req)
	}
}
