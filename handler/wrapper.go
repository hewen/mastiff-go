package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hewen/mastiff-go/logger"
)

// WithRequest is a handler with request and response type.
type WithRequest[T any, R any] func(c *gin.Context, req T) (R, error)

// WrapHandler wraps a handler with request and response type.
func WrapHandler[T any, R any](handle WithRequest[T, R]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T
		l := logger.NewLoggerWithContext(c.Request.Context())

		if err := c.ShouldBindJSON(&req); err != nil {
			l.Fields(map[string]any{"err": err}).Errorf("invalid request")
			c.JSON(http.StatusBadRequest, BaseResp{
				Code:  http.StatusBadRequest,
				Trace: l.GetTraceID(),
			})
			return
		}
		c.Set("req", req)

		resp, err := handle(c, req)
		if err != nil {
			l.Fields(map[string]any{"err": err}).Errorf("handler error")
			c.JSON(http.StatusInternalServerError, BaseResp{
				Code:  http.StatusInternalServerError,
				Trace: l.GetTraceID(),
			})
			return
		}

		c.Set("resp", resp)
		c.JSON(http.StatusOK, RespWithData[R]{
			BaseResp: BaseResp{
				Code:  http.StatusOK,
				Trace: l.GetTraceID(),
			},
			Data: resp,
		})
	}
}
