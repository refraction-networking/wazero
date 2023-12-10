package main

import (
	"testing"

	"github.com/tetratelabs/wazero/internal/testing/maintester"
	"github.com/tetratelabs/wazero/internal/testing/require"
)

func Test_main(t *testing.T) {
	stdout, _ := maintester.TestMain(t, main, "")
	require.Equal(t, `read 12 bytes: hello world
read 12 bytes: hello world
`, stdout)
}
