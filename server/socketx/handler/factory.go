// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"fmt"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/panjf2000/gnet/v2"
)

var (
	// ErrEmptySocketConf is returned when the socket config is empty.
	ErrEmptySocketConf = fmt.Errorf("empty socket conf")
)

// BuildParams contains the parameters needed to build a socket handler.
type BuildParams struct {
	GnetHandler gnet.EventHandler
}

// NewHandler creates a handler from registered builders for different socket frameworks.
func NewHandler(conf *serverconf.SocketConfig, params BuildParams) (SocketHandler, error) {
	if conf == nil {
		return nil, ErrEmptySocketConf
	}

	switch conf.FrameworkType {
	case serverconf.FrameworkGnet:
		if params.GnetHandler == nil {
			return nil, fmt.Errorf("gnet: handler is nil")
		}

		return NewGnetHandler(conf, params.GnetHandler)
	default:
		return nil, fmt.Errorf("unsupported socket type: %s", conf.FrameworkType)
	}

}
