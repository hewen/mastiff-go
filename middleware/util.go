// Package middleware provides common middleware for HTTP and gRPC.
package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// GrpcServerStream is a mock implementation of grpc.ServerStream for testing.
type GrpcServerStream struct {
	grpc.ServerStream
	Ctx context.Context
}

// Context returns the context associated with the stream.
func (s *GrpcServerStream) Context() context.Context {
	return s.Ctx
}
