// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sys

import (
	"net"
	"os"

	"github.com/tetratelabs/wazero/internal/fsapi"
	"github.com/tetratelabs/wazero/internal/sysfs"
)

// [WATER] Extended api.Module interface to support inserting *net.TCPConn, *net.TCPListener and *os.File.

// InsertTCPConn inserts a *net.TCPConn into the module's openedFiles and
// returns the key and a boolean indicating whether the insertion is successful.
//
// The key could be used as an opened file descriptor from within the WebAssembly
// instance.
func (c *Context) InsertTCPConn(conn *net.TCPConn) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      fsapi.Adapt(sysfs.NewTCPConnFile(conn)),
	})
}

// InsertTCPListener inserts a *net.TCPListener into the module's openedFiles and
// returns the key and a boolean indicating whether the insertion is successful.
//
// The key could be used as an opened file descriptor from within the WebAssembly
// instance.
func (c *Context) InsertTCPListener(listener *net.TCPListener) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      fsapi.Adapt(sysfs.NewTCPListenerFile(listener)),
	})
}

// InsertOSFile inserts a *os.File into the module's openedFiles and returns the
// key and a boolean indicating whether the insertion is successful.
//
// The key could be used as an opened file descriptor from within the WebAssembly
// instance.
func (c *Context) InsertOSFile(file *os.File) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      sysfs.NewOSFile(file.Name(), 0, 0, file), // TODO: fix flag and perm
	})
}
