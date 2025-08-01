package metrics

import (
	"net/http"
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
	r.Get("/ping", func(c unicontext.UniversalContext) error {
		return c.String(http.StatusOK, "pong")
	})

	req, _ := http.NewRequest("GET", "/ping", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
