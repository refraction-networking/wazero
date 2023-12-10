// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package api

import (
	"net"
	"os"
)

type WATERModuleExtension interface {
	InsertTCPConn(*net.TCPConn) (key int32, ok bool)
	InsertTCPListener(*net.TCPListener) (key int32, ok bool)
	InsertOSFile(*os.File) (key int32, ok bool)
}
