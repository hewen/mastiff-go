// Package handler provides a unified socket abstraction over gnet.
package handler

import (
	"sync"
	"time"

	"github.com/hewen/mastiff-go/server/socketx/codec"
)

// ConnMeta represents the metadata associated with a connection.
type ConnMeta struct {
	// Conn is the underlying connection.
	Conn Conn
	// Codec is the codec used for encoding/decoding messages.
	Codec codec.SecureCodec
	// SessionKey is the session key used for encoding/decoding messages.
	SessionKey []byte
	// Buf is the buffer used for storing incomplete messages.
	Buf []byte
	// LastActive is the last active time of the connection.  Unix timestamp (second)
	LastActive int64
	// HandshakeOK indicates whether the handshake is complete.
	HandshakeOK bool
}

// ConnManager is a manager that manages connections and their metadata.
type ConnManager struct {
	// deviceID -> ConnMeta
	deviceConns sync.Map
}

// NewConnManager creates a new ConnManager.
func NewConnManager() *ConnManager {
	return &ConnManager{}
}

// UpdateActivity updates the last active time of a connection.
func (cm *ConnManager) UpdateActivity(deviceID string) {
	val, ok := cm.deviceConns.Load(deviceID)
	if !ok {
		return
	}
	meta := val.(*ConnMeta)
	meta.LastActive = time.Now().Unix()
	cm.deviceConns.Store(deviceID, meta)
}

// UnbindDevice unbinds a device by ID.
func (cm *ConnManager) UnbindDevice(deviceID string) {
	val, ok := cm.deviceConns.LoadAndDelete(deviceID)
	if !ok {
		return
	}

	meta, ok := val.(*ConnMeta)
	if ok && meta.Conn != nil {
		_ = meta.Conn.Close()
	}
}

// BindDevice binds a device by ID.
func (cm *ConnManager) BindDevice(deviceID string, meta *ConnMeta) {
	meta.LastActive = time.Now().Unix()
	cm.deviceConns.Store(deviceID, meta)
}

// GetConnMeta gets the connection metadata for a device by ID.
func (cm *ConnManager) GetConnMeta(deviceID string) (*ConnMeta, bool) {
	if deviceID == "" {
		return nil, false
	}

	val, ok := cm.deviceConns.Load(deviceID)
	if !ok {
		return nil, false
	}
	meta, ok := val.(*ConnMeta)
	if !ok {
		return nil, false
	}
	return meta, true
}

// CleanupInactive cleans up inactive connections.
func (cm *ConnManager) CleanupInactive(maxIdleSeconds int64) {
	now := time.Now().Unix()
	cm.deviceConns.Range(func(key, val any) bool {
		deviceID := key.(string)
		meta := val.(*ConnMeta)
		if now-meta.LastActive > maxIdleSeconds {
			_ = meta.Conn.Close()
			cm.deviceConns.Delete(deviceID)
		}
		return true
	})
}
