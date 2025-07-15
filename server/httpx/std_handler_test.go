package httpx

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestStdHandlerBuilder(t *testing.T) {
	port, err := util.GetFreePort()
	assert.Nil(t, err)

	conf := &serverconf.HTTPConfig{
		Addr: fmt.Sprintf("localhost:%d", port),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	builder := &StdHTTPHandlerBuilder{
		Handler: mux,
		Conf:    conf,
	}

	s, err := NewHTTPServer(builder)
	assert.Nil(t, err)

	go func() {
		defer s.Stop()
		s.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/test", port))
	assert.Nil(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), `"message":"ok"`)
}

func TestStdHandlerBuilder_EmptyConfig(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	builder := &StdHTTPHandlerBuilder{
		Handler: mux,
		Conf:    nil,
	}

	_, err := NewHTTPServer(builder)
	assert.NotNil(t, err)
}
