// Package handler provides a context interface for HTTP handlers.
package handler

import (
	"net/http"

	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
)

// WrapHandlerFunc is the function signature for HTTP handlers.
type WrapHandlerFunc[T any, R any] func(ctx unicontext.UniversalContext, req T) (R, error)

// WrapHandler wraps a handler function into a handler function that takes a Context.
func WrapHandler[T any, R any](handle WrapHandlerFunc[T, R]) func(ctx unicontext.UniversalContext) error {
	return func(ctx unicontext.UniversalContext) error {
		var req T
		l := logger.NewLoggerWithContext(ctx.Request().Context())

		if err := ctx.BindJSON(&req); err != nil {
			l.Fields(map[string]any{"err": err}).Errorf("invalid request")
			return ctx.JSON(http.StatusBadRequest, BaseResp{
				Code:  http.StatusBadRequest,
				Trace: l.GetTraceID(),
			})
		}
		ctx.Set("req", req)

		resp, err := handle(ctx, req)
		if err != nil {
			l.Fields(map[string]any{"err": err}).Errorf("handler error")
			return ctx.JSON(http.StatusInternalServerError, BaseResp{
				Code:  http.StatusInternalServerError,
				Trace: l.GetTraceID(),
			})
		}

		ctx.Set("resp", resp)
		return ctx.JSON(http.StatusOK, RespWithData[R]{
			BaseResp: BaseResp{
				Code:  http.StatusOK,
				Trace: l.GetTraceID(),
			},
			Data: resp,
		})
	}
}
