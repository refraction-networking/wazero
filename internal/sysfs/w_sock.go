package sysfs

import (
	"net"

	socketapi "github.com/tetratelabs/wazero/internal/sock"
)

// NewTCPListenerFile creates a socketapi.TCPSock for a given *net.TCPListener.
func NewTCPConnFile(tc *net.TCPConn) socketapi.TCPConn {
	return newTcpConn(tc)
}
