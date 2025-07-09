package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/middleware/logging"
)

func TestWrapHandler(_ *testing.T) {
	type Test struct {
		Test int `json:"test"`
	}

	wrapHandler := WrapHandler(func(_ *gin.Context, req Test) (resp Test, err error) {
		return req, nil
	})
	handlerLog := logging.GinLoggingHandler()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/test", bytes.NewReader([]byte(``)))
	ctx.Request.Header.Add("Content-Type", "application/json")

	wrapHandler(ctx)
	handlerLog(ctx)

	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/test", bytes.NewReader([]byte(`{"test":1}`)))
	ctx.Request.Header.Add("Content-Type", "application/json")

	wrapHandler(ctx)
	handlerLog(ctx)
}

func TestWrapHandlerError(_ *testing.T) {
	type Test struct {
		Test int `json:"test"`
	}

	wrapHandler := WrapHandler(func(_ *gin.Context, req Test) (resp Test, err error) {
		return req, fmt.Errorf("wapp error")
	})
	handlerLog := logging.GinLoggingHandler()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/test", bytes.NewReader([]byte(`{"test":1}`)))
	ctx.Request.Header.Add("Content-Type", "application/json")

	wrapHandler(ctx)
	handlerLog(ctx)
}
