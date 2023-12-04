package sysfs

import (
	"io/fs"
	"os"

	experimentalsys "github.com/tetratelabs/wazero/experimental/sys"
	"github.com/tetratelabs/wazero/internal/fsapi"
)

func NewOSFile(path string, flag experimentalsys.Oflag, perm fs.FileMode, f *os.File) fsapi.File {
	return newOsFile(path, flag, perm, f)
}
