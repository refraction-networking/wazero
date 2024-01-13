// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sysfs

import (
	"net"

	socketapi "github.com/tetratelabs/wazero/internal/sock"
)

// NewTCPConnFile creates a socketapi.TCPSock for a given *net.TCPConn, which
// can then be used as a preopened file through the WASI API.
func NewTCPConnFile(tc *net.TCPConn) socketapi.TCPConn {
	return newTcpConn(tc)
}
