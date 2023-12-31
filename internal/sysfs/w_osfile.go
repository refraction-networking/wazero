// Copyright 2023 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sysfs

import (
	"io/fs"
	"os"

	experimentalsys "github.com/tetratelabs/wazero/experimental/sys"
	"github.com/tetratelabs/wazero/internal/fsapi"
)

// NewOSFile wraps an *os.File as a fsapi.File, which can then be used as a
// preopened file through the WASI API.
func NewOSFile(path string, flag experimentalsys.Oflag, perm fs.FileMode, f *os.File) fsapi.File {
	return newOsFile(path, flag, perm, f)
}
