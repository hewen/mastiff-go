package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// TestNewGatewayHandler_Success tests that the gateway handler
// correctly registers a route and responds with HTTP 200.
func TestNewGatewayHandler_Success(t *testing.T) {
	// Mock a grpc-gateway registration function that registers a GET /test endpoint.
	mockRegister := func(_ context.Context, mux *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
		_ = mux.HandlePath("GET", "/test", func(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		return nil
	}

	r := gin.New()
	handlerFunc := NewGatewayHandler("localhost:1234", mockRegister)
	handlerFunc(r)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/test")
	assert.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestNewGatewayHandler_RegisterErrorPanic tests that the gateway handler
// panics if the registration function returns an error.
func TestNewGatewayHandler_RegisterErrorPanic(t *testing.T) {
	// Mock a grpc-gateway registration function that returns an error.
	mockRegister := func(_ context.Context, _ *runtime.ServeMux, _ string, _ []grpc.DialOption) error {
		return context.DeadlineExceeded
	}

	r := gin.New()
	handlerFunc := NewGatewayHandler("localhost:1234", mockRegister)

	// Assert that the handler function panics on registration error.
	assert.Panics(t, func() {
		handlerFunc(r)
	})
}
