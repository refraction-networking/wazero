// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package wazerotest

import (
	"net"
	"os"
)

// TODO: actually implement these functions defined below and add corresponding tests to this package.

// Implements api.WATERExtendedModule
func (m *Module) InsertTCPConn(*net.TCPConn) (key int32, ok bool) {
	return 0, false
}

// Implements api.WATERExtendedModule
func (m *Module) InsertTCPListener(*net.TCPListener) (key int32, ok bool) {
	return 0, false
}

// Implements api.WATERExtendedModule
func (m *Module) InsertOSFile(*os.File) (key int32, ok bool) {
	return 0, false
}
