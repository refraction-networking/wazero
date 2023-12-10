//go:build !linux && !darwin && !windows

package sysfs

import (
	"net"

	"github.com/tetratelabs/wazero/experimental/sys"
	socketapi "github.com/tetratelabs/wazero/internal/sock"
)

// MSG_PEEK is a filler value.
const MSG_PEEK = 0x2

func newTCPListenerFile(tl *net.TCPListener) socketapi.TCPSock {
	return &unsupportedSockFile{}
}

type unsupportedSockFile struct {
	baseSockFile
}

// Accept implements the same method as documented on socketapi.TCPSock
func (f *unsupportedSockFile) Accept() (socketapi.TCPConn, sys.Errno) {
	return nil, sys.ENOSYS
}

func newTcpConn(tc *net.TCPConn) socketapi.TCPConn {
	return &unsupportedConnFile{}
}

type unsupportedConnFile struct {
	baseSockFile
}

func (f *unsupportedConnFile) Recvfrom(p []byte, flags int) (n int, errno sys.Errno) {
	return 0, sys.ENOSYS
}

func (f *unsupportedConnFile) Shutdown(how int) sys.Errno {
	return sys.ENOSYS
}
