// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

//go:build !linux && !darwin && !windows

package sysfs

import (
	"net"

	"github.com/tetratelabs/wazero/experimental/sys"
	socketapi "github.com/tetratelabs/wazero/internal/sock"
)

func newTcpConn(tc *net.TCPConn) socketapi.TCPConn {
	return &unsupportedSockFile{}
}

// Recvfrom implements the same method as documented on socketapi.TCPConn
func (*unsupportedSockFile) Recvfrom([]byte, int) (int, sys.Errno) {
	return 0, sys.ENOSYS
}

// Shutdown implements the same method as documented on sys.Conn
func (*unsupportedSockFile) Shutdown(int) sys.Errno {
	return sys.ENOSYS
}
