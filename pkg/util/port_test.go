package util

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTCPHelper struct {
	mock.Mock
}

func (m *MockTCPHelper) ResolveTCPAddr(network, address string) (*net.TCPAddr, error) {
	args := m.Called(network, address)
	return args.Get(0).(*net.TCPAddr), args.Error(1)
}

func (m *MockTCPHelper) ListenTCP(network string, laddr *net.TCPAddr) (*net.TCPListener, error) {
	args := m.Called(network, laddr)
	return args.Get(0).(*net.TCPListener), args.Error(1)
}

func TestGetFreePort(t *testing.T) {
	_, err := GetFreePort()
	assert.Nil(t, err)
}

func TestGetFreePortWithHelper_Success(t *testing.T) {
	mockHelper := new(MockTCPHelper)

	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}
	listener, _ := net.ListenTCP("tcp", addr)
	defer func() {
		_ = listener.Close()
	}()
	mockHelper.On("ResolveTCPAddr", "tcp", "localhost:0").Return(addr, nil)
	mockHelper.On("ListenTCP", "tcp", addr).Return(listener, nil)

	port, err := GetFreePortWithHelper(mockHelper)
	assert.Nil(t, err)
	assert.Equal(t, listener.Addr().(*net.TCPAddr).Port, port)
}

func TestGetFreePortWithHelper_ResolveFail(t *testing.T) {
	mockHelper := new(MockTCPHelper)
	mockHelper.On("ResolveTCPAddr", "tcp", "localhost:0").Return(&net.TCPAddr{}, errors.New("resolve failed"))

	port, err := GetFreePortWithHelper(mockHelper)
	assert.NotNil(t, err)
	assert.Equal(t, 0, port)
}

func TestGetFreePortWithHelper_ListenFail(t *testing.T) {
	mockHelper := new(MockTCPHelper)
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0}

	mockHelper.On("ResolveTCPAddr", "tcp", "localhost:0").Return(addr, nil)
	mockHelper.On("ListenTCP", "tcp", addr).Return(&net.TCPListener{}, errors.New("listen failed"))

	port, err := GetFreePortWithHelper(mockHelper)
	assert.NotNil(t, err)
	assert.Equal(t, 0, port)
}
