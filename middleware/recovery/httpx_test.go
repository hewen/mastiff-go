package recovery

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

	r.Get("/panic", func(_ unicontext.UniversalContext) error {
		panic("simulated panic for test")
	})

	req := httptest.NewRequest("GET", "/panic", nil)

	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
