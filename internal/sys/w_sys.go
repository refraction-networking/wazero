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

func (c *Context) InsertTCPConn(conn *net.TCPConn) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      fsapi.Adapt(sysfs.NewTCPConnFile(conn)),
	})
}

func (c *Context) InsertTCPListener(listener *net.TCPListener) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      fsapi.Adapt(sysfs.NewTCPListenerFile(listener)),
	})
}

func (c *Context) InsertOSFile(file *os.File) (key int32, ok bool) {
	return c.fsc.openedFiles.Insert(&FileEntry{
		IsPreopen: true,
		File:      sysfs.NewOSFile(file.Name(), 0, 0, file), // TODO: fix flag and perm
	})
}
