package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/server/httpx"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

func TestHttpxMiddleware(t *testing.T) {
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)
	r.Use(HttpxMiddleware())

	r.Get("/log", func(ctx unicontext.UniversalContext) error {
		ctx.Set("req", "req")
		ctx.Set("resp", "resp")
		return nil
	})

	req := httptest.NewRequest("GET", "/log", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
