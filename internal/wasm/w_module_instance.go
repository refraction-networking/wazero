// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package wasm

import (
	"net"
	"os"
)

// Implements api.WATERExtendedModule
func (m *ModuleInstance) InsertTCPConn(conn *net.TCPConn) (key int32, ok bool) {
	return m.Sys.InsertTCPConn(conn)
}

// Implements api.WATERExtendedModule
func (m *ModuleInstance) InsertTCPListener(lis *net.TCPListener) (key int32, ok bool) {
	return m.Sys.InsertTCPListener(lis)
}

// Implements api.WATERExtendedModule
func (m *ModuleInstance) InsertOSFile(f *os.File) (key int32, ok bool) {
	return m.Sys.InsertOSFile(f)
}
