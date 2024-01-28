//go:build (linux || darwin || windows) && !tinygo

// Copyright 2024 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"

	"github.com/tetratelabs/wazero/experimental/sys"
	"github.com/tetratelabs/wazero/internal/fsapi"
)

type PollFd struct {
	Fd      uintptr
	Events  fsapi.Pflag
	Revents fsapi.Pflag // set only if events triggered
}

func Poll(fds []PollFd, timeoutMillis int32) (int, sys.Errno) {
	var pollFds []pollFd
	for _, fd := range fds {
		if fsapi.Pflag(fd.Events) == fsapi.POLLIN {
			pollFds = append(pollFds, newPollFd(fd.Fd, _POLLIN, 0))
		} else if fsapi.Pflag(fd.Events) == fsapi.POLLOUT {
			pollFds = append(pollFds, newPollFd(fd.Fd, _POLLOUT, 0))
		} else {
			return 0, sys.ENOTSUP
		}
	}

	n, err := _poll(pollFds, timeoutMillis)
	if !errors.Is(err, sys.Errno(0)) {
		return n, err
	}

	// check the pollFds for errors
	if n > 0 {
		for i, pfd := range pollFds {
			if pfd.revents&pfd.events != 0 {
				fds[i].Revents = fds[i].Events
			} else if pfd.revents != 0 {
				fds[i].Revents = fsapi.POLLUNKNOWN // TODO: Need more in-depth checking of the returned event.
			}
		}
	}
	return n, 0
}
