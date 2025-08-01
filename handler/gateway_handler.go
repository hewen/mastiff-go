// Package handler provides a handler for the gRPC gateway.
package handler

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GatewayRegisterFunc is a function that registers a gRPC service with the gateway.
type GatewayRegisterFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

// NewGatewayHandlerGin returns a new handler for the gRPC gateway.
func NewGatewayHandlerGin(grpcTarget string, register GatewayRegisterFunc) func(*gin.Engine) {
	return func(r *gin.Engine) {
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		if err := register(context.Background(), mux, grpcTarget, opts); err != nil {
			panic(fmt.Sprintf("failed to register gateway handler: %v", err))
		}

		r.Any("/*any", gin.WrapH(mux))
	}
}

// NewGatewayHandlerFiber returns a new handler for the gRPC gateway.
func NewGatewayHandlerFiber(grpcTarget string, register GatewayRegisterFunc) func(app *fiber.App) {
	return func(app *fiber.App) {
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		if err := register(context.Background(), mux, grpcTarget, opts); err != nil {
			panic(fmt.Sprintf("failed to register gateway handler: %v", err))
		}

		app.All("/*", adaptor.HTTPHandler(mux))
	}
}
