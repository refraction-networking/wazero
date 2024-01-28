// Copyright 2024 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sysfs

import (
	experimentalsys "github.com/tetratelabs/wazero/experimental/sys"
)

// Fd implements the same method as documented on fsapi.File
func (f *tcpListenerFile) Fd() uintptr {
	var fd uintptr

	syscallConnControl(f.tl, func(_fd uintptr) (int, experimentalsys.Errno) {
		fd = _fd
		return 0, 0
	})

	return fd
}

// Fd implements the same method as documented on fsapi.File
func (f *tcpConnFile) Fd() uintptr {
	var fd uintptr

	syscallConnControl(f.tc, func(_fd uintptr) (int, experimentalsys.Errno) {
		fd = _fd
		return 0, 0
	})

	return fd
}
