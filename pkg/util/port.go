// Package util provides utility functions for common tasks such as
// time formatting, network helpers, and other reusable components.
package util

import "net"

// TCPHelper defines the interface for TCP address resolution and listening.
type TCPHelper interface {
	ResolveTCPAddr(network, address string) (*net.TCPAddr, error)
	ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
}

// realTCPHelper is the default implementation of TCPHelper that uses the net package.
type realTCPHelper struct{}

// ResolveTCPAddr resolves a TCP address using the net package.
func (r *realTCPHelper) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr(network, address)
}

// ListenTCP listens for TCP connections on the specified address using the net package.
func (r *realTCPHelper) ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error) {
	return net.ListenTCP(network, laddr)
}

// GetFreePortWithHelper returns an available TCP port from the local system.
func GetFreePortWithHelper(helper TCPHelper) (int, error) {
	addr, err := helper.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := helper.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetFreePort returns an available TCP port from the local system.
//
// It works by asking the OS to assign a free port by binding to "localhost:0".
// The OS will choose an available port automatically. The listener is then closed
// immediately, and the selected port is returned.
//
// Note: There is a small chance of a race condition if the port is used by another
// process before your program binds to it again.
func GetFreePort() (int, error) {
	return GetFreePortWithHelper(&realTCPHelper{})
}
