// Copyright 2024 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"

	"github.com/tetratelabs/wazero/experimental/sys"
	socketapi "github.com/tetratelabs/wazero/internal/sock"
)

const (
	POLLIN = _POLLIN // export the value
)

func PollTCPConns(conns []socketapi.TCPConn, events []int16) (int, error) {
	var pollFds []pollFd
	for i, conn := range conns {
		syscallConnControl(conn.(*tcpConnFile).tc, func(fd uintptr) (int, sys.Errno) {
			if events[i]&_POLLIN == 0 {
				return 0, sys.EINVAL
			}
			pollFds = append(pollFds, newPollFd(fd, _POLLIN, 0))
			return 0, 0
		})
	}

	for {
		ready, err := _poll(pollFds, -1)
		if ready == 0 {
			if errors.Is(err, sys.EINTR) || errors.Is(err, sys.Errno(0)) {
				continue
			}
		}

		return ready, err
	}
}
