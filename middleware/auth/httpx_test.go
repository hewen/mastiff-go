package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/middlewareconf/authconf"
	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/internal/contextkeys"
	"github.com/hewen/mastiff-go/server/httpx"
	"github.com/hewen/mastiff-go/server/httpx/unicontext"
	"github.com/stretchr/testify/assert"
)

func TestHttpxWhiteList(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}

	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)

	r.Use(HttpxMiddleware(conf))
	r.Get("/public", func(c unicontext.UniversalContext) error {
		return c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHttpxMissingToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)
	r.Use(HttpxMiddleware(conf))
	r.Get("/public", func(c unicontext.UniversalContext) error {
		return c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
func TestHttpxInvalidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)
	r.Use(HttpxMiddleware(conf))
	r.Get("/secure", func(c unicontext.UniversalContext) error {
		return c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	req.Header.Set("Authorization", "invalid-token")
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHttpxValidToken(t *testing.T) {
	conf := &authconf.Config{
		HeaderKey:     "Authorization",
		TokenPrefixes: []string{"Bearer"},
		JWTSecret:     "secret",
		WhiteList:     []string{"/public"},
	}
	r, err := httpx.NewHTTPServer(&serverconf.HTTPConfig{
		FrameworkType: serverconf.FrameworkFiber,
	})
	assert.Nil(t, err)
	r.Use(HttpxMiddleware(conf))
	r.Get("/secure", func(c unicontext.UniversalContext) error {
		ctx := unicontext.ContextFrom(c)
		ai, exists := contextkeys.GetAuthInfo(ctx)
		if !exists || ai.UserID != "123" {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "auth info missing"})
		}
		return c.String(http.StatusOK, "hello "+ai.UserID)
	})

	req, _ := http.NewRequest("GET", "/secure", nil)
	tk, _ := GenerateJWTToken(map[string]any{"user_id": "123"}, conf.JWTSecret, time.Minute)
	req.Header.Set("Authorization", "Bearer "+tk)
	resp, err := r.Test(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
