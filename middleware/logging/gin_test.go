package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(GinMiddleware())

	router.GET("/log", func(ctx *gin.Context) {
		ctx.Set("req", "req")
		ctx.Set("resp", "resp")
	})

	req := httptest.NewRequest("GET", "/log", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
