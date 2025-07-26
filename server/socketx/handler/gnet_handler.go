// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/hewen/mastiff-go/config/serverconf"
	"github.com/hewen/mastiff-go/logger"
	"github.com/hewen/mastiff-go/server/socketx/codec"
	"github.com/panjf2000/gnet/v2"
)

var (
	// ErrDeviceNotFound is the error returned when the device is not found.
	ErrDeviceNotFound = fmt.Errorf("device not found")
	// ErrConnMetaNotFound is the error returned when the connection metadata is not found.
	ErrConnMetaNotFound = fmt.Errorf("conn meta not found")
	// ErrIncompletePacket is the error returned when a packet is incomplete.
	ErrIncompletePacket = fmt.Errorf("incomplete packet error")
)

// GnetHandler is a handler that provides a unified socket abstraction over gnet.
type GnetHandler struct {
	logger      logger.Logger
	conf        *serverconf.SocketConfig
	event       GnetEventHandler
	engine      gnet.Engine
	connManager *ConnManager
}

// OnBoot is called once when the engine starts.
func (h *GnetHandler) OnBoot(e gnet.Engine) gnet.Action {
	h.engine = e
	return h.event.OnBoot(e)
}

// OnShutdown is called when the engine is shutting down.
func (h *GnetHandler) OnShutdown(e gnet.Engine) {
	if h.event == nil {
		return
	}

	h.connManager.deviceConns.Range(func(_, val any) bool {
		conn := val.(*ConnMeta).Conn
		_ = conn.Close()
		return true
	})

	h.event.OnShutdown(e)
}

// OnOpen is called when a new connection is opened.
func (h *GnetHandler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	h.logger.Infof("OPEN: %v", c.RemoteAddr())
	if h.event == nil {
		return nil, gnet.None
	}

	return h.event.OnOpen(c)
}

// OnClose is called when a connection is closed.
func (h *GnetHandler) OnClose(c gnet.Conn, err error) gnet.Action {
	h.logger.Infof("CLOSE: %v", c.RemoteAddr())
	if h.event == nil {
		return gnet.None
	}

	if deviceID, ok := GetContextValue[string](&GnetConn{c}, string(ContextKeyDeviceID)); ok {
		h.connManager.UnbindDevice(deviceID)
	}

	return h.event.OnClose(c, err)
}

// OnTick is called periodically by the event loop.
func (h *GnetHandler) OnTick() (time.Duration, gnet.Action) {
	h.logger.Infof("CleanupInactive start")
	h.connManager.CleanupInactive(int64(h.conf.MaxIdleTime.Seconds()))
	h.logger.Infof("CleanupInactive end")

	if h.event != nil {
		return h.event.OnTick()
	}
	return h.conf.TickInterval, gnet.None
}

// OnTraffic is triggered when a complete message is received.
func (h *GnetHandler) OnTraffic(c gnet.Conn) gnet.Action {
	conn := &GnetConn{c}

	// Read all available data
	data, _ := c.Next(-1)
	if len(data) == 0 {
		return gnet.None
	}

	deviceID, ok := GetContextValue[string](conn, string(ContextKeyDeviceID))
	if ok {
		h.connManager.UpdateActivity(deviceID)
	}

	meta, found := h.connManager.GetConnMeta(deviceID)
	if found {
		meta.Buf = append(meta.Buf, data...)
		return h.handleMessage(conn, deviceID, meta)
	}

	// no session yet, treat as handshake
	return h.handleHandshake(conn, data)
}

func (h *GnetHandler) handleHandshake(conn *GnetConn, data []byte) gnet.Action {
	protocol, matched := h.event.MatchHandshakePrefix(data)
	if !matched {
		h.logger.Warnf("unmatched handshake prefix from %s", conn.RemoteAddr().String())
		return gnet.Close
	}

	codec, err := h.event.NewCodec(protocol)
	if err != nil {
		h.logger.Errorf("failed to init codec: %v", err)
		return gnet.Close
	}

	msg, key, deviceID, err := codec.Handshake(data)
	if err != nil {
		h.logger.Errorf("handshake failed: %v", err)
		return gnet.Close
	}

	SetContextValue(conn, string(ContextKeyDeviceID), deviceID)
	h.connManager.BindDevice(deviceID, &ConnMeta{
		Conn:        conn,
		Codec:       codec,
		SessionKey:  key,
		HandshakeOK: true,
		LastActive:  time.Now().Unix(),
	})

	resp := h.event.OnHandshakeMessage(conn.Conn, msg)
	if resp == nil {
		return gnet.None
	}

	out, err := codec.Encode(resp, key)
	if err != nil {
		h.logger.Errorf("handshake response encode failed: %v", err)
		h.connManager.UnbindDevice(deviceID)
		return gnet.Close
	}

	_, _ = conn.Write(out)
	return gnet.None
}

func (h *GnetHandler) handleMessage(conn *GnetConn, deviceID string, meta *ConnMeta) gnet.Action {
	for {
		msg, consumed, err := meta.Codec.Decode(meta.Buf, meta.SessionKey)
		if err != nil {
			if err == ErrIncompletePacket {
				break
			}
			h.logger.Errorf("decode failed: %v", err)
			h.connManager.UnbindDevice(deviceID)
			return gnet.Close
		}

		if consumed == 0 {
			break
		}

		meta.Buf = meta.Buf[consumed:]

		resp := h.event.OnMessage(conn.Conn, msg)
		if resp == nil {
			continue
		}

		out, err := meta.Codec.Encode(resp, meta.SessionKey)
		if err != nil {
			h.logger.Errorf("encode response failed: %v", err)
			h.connManager.UnbindDevice(deviceID)
			return gnet.Close
		}

		_, _ = conn.Write(out)
	}
	return gnet.None
}

// PushTo sends a message to a specific device by ID, with an optional async callback.
func (h *GnetHandler) PushTo(deviceID string, msg codec.Message, callback AsyncCallback) error {
	val, ok := h.connManager.deviceConns.Load(deviceID)
	if !ok {
		return ErrDeviceNotFound
	}

	meta, ok := val.(*ConnMeta)
	if !ok {
		return ErrConnMetaNotFound
	}

	if meta.Codec == nil {
		return nil
	}

	data, err := meta.Codec.Encode(msg, meta.SessionKey)
	if err != nil {
		return err
	}

	return meta.Conn.AsyncWrite(data, callback)
}

// UnbindDevice unbinds a device by ID.
func (h *GnetHandler) UnbindDevice(deviceID string) {
	h.connManager.UnbindDevice(deviceID)
}

// Name returns the name of the GnetHandler.
func (h *GnetHandler) Name() string {
	return "gnet"
}

// Start starts the GnetHandler.
func (h *GnetHandler) Start() error {
	return gnet.Run(h, h.conf.Addr, gnet.WithOptions(h.conf.GnetOptions))
}

// Stop stops the GnetHandler.
func (h *GnetHandler) Stop() error {
	return h.engine.Stop(context.TODO())
}

// NewGnetHandler creates a new GnetHandler.
func NewGnetHandler(conf *serverconf.SocketConfig, event GnetEventHandler) (*GnetHandler, error) {
	if conf == nil {
		return nil, ErrEmptySocketConf
	}

	conf.SetDefault()

	l := logger.NewLogger()
	conf.GnetOptions.Logger = l

	return &GnetHandler{
		logger:      l,
		connManager: NewConnManager(),
		conf:        conf,
		event:       event,
	}, nil
}
