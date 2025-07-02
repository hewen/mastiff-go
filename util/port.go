// Package util provides utility functions for common tasks such as
// time formatting, network helpers, and other reusable components.
package util

import "net"

// GetFreePort returns an available TCP port from the local system.
//
// It works by asking the OS to assign a free port by binding to "localhost:0".
// The OS will choose an available port automatically. The listener is then closed
// immediately, and the selected port is returned.
//
// Note: There is a small chance of a race condition if the port is used by another
// process before your program binds to it again.
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	a, err = net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return
	}

	var l *net.TCPListener
	l, err = net.ListenTCP("tcp", a)
	if err != nil {
		return
	}

	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}
