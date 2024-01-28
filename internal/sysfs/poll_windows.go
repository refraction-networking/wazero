package sysfs

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/tetratelabs/wazero/experimental/sys"
)

var (
	procWSAPoll          = modws2_32.NewProc("WSAPoll")
	procGetNamedPipeInfo = kernel32.NewProc("GetNamedPipeInfo")
)

const (
	// _POLLRDNORM subscribes to normal data for read.
	_POLLRDNORM = 0x0100
	// _POLLRDBAND subscribes to priority band (out-of-band) data for read.
	_POLLRDBAND = 0x0200
	// _POLLWRNORM subscribes to normal data for write.
	_POLLWRNORM = 0x0010 // [WATER] added _POLLWRNORM to support subscription to FdWrite events
	// _POLLIN subscribes a notification when any readable data is available.
	_POLLIN = (_POLLRDNORM | _POLLRDBAND)
	// _POLLOUT subscribes a notification when any writeable data can be written.
	_POLLOUT = _POLLWRNORM // [WATER] added _POLLOUT to support subscription to FdWrite events
)

// pollFd is the struct to query for file descriptor events using poll.
type pollFd struct {
	// fd is the file descriptor.
	fd uintptr
	// events is a bitmap containing the requested events.
	events int16
	// revents is a bitmap containing the returned events.
	revents int16
}

// newPollFd is a constructor for pollFd that abstracts the platform-specific type of file descriptors.
func newPollFd(fd uintptr, events, revents int16) pollFd {
	return pollFd{fd: fd, events: events, revents: revents}
}

// pollInterval is the interval between each calls to peekNamedPipe in selectAllHandles
const pollInterval = 100 * time.Millisecond

// _poll implements poll on Windows, for a subset of cases.
//
// fds may contain any number of file handles, but regular files and pipes are only processed for _POLLIN.
// Stdin is a pipe, thus it is checked for readiness when present. Pipes are checked using PeekNamedPipe.
// Regular files always immediately reported as ready, regardless their actual state and timeouts.
//
// If n==0 it will wait for the given timeout duration, but it will return sys.ENOSYS if timeout is nil,
// i.e. it won't block indefinitely. The given ctx is used to allow for cancellation,
// and it is currently used only in tests.
//
// The implementation actually polls every 100 milliseconds (pollInterval) until it reaches the
// given timeout (in millis).
//
// The duration may be negative, in which case it will wait indefinitely. The given ctx is
// used to allow for cancellation, and it is currently used only in tests.
func _poll(fds []pollFd, timeoutMillis int32) (n int, errno sys.Errno) {
	if fds == nil {
		return -1, sys.ENOSYS
	}

	regular, pipes, sockets, errno := partionByFtype(fds)
	nregular := len(regular)
	if errno != 0 {
		return -1, errno
	}

	// [WATER SECTION BEGIN]

	// automatically merge the partitions back to the original slice to update revents
	defer mergePartitions(fds, regular, pipes, sockets)

	// phony poll regular files: always writable and readable
	for _, fd := range regular {
		if fd.events == _POLLIN || fd.events == _POLLOUT {
			fd.revents = fd.events
		}
	}

	// First do a one-shot check for any ready-to-go pipes or sockets.
	npipes, nsockets, errno := peekAll(pipes, sockets)
	if errno != 0 {
		return -1, errno
	}
	count := nregular + npipes + nsockets
	if count > 0 {
		return count, 0
	}

	// Now we learned:
	//  - no regular files in the list (otherwise already returned)
	//  - none of the pipes or sockets are ready

	// We can invoke wsaPoll with the given timeout instead of busy-looping if
	// only sockets are present.
	if len(pipes) == 0 {
		return wsaPoll(sockets, int(timeoutMillis))
	}

	// Otherwise, we need to check both pipes and sockets, and cannot use wsaPoll.
	// We use a ticker to trigger a check periodically, and a timer to expire after
	// the given timeout.

	// [WATER SECTION END]

	// Ticker that emits at every pollInterval.
	tick := time.NewTicker(pollInterval)
	tickCh := tick.C
	defer tick.Stop()

	// Timer that expires after the given duration.
	// Initialize afterCh as nil: the select below will wait forever.
	var afterCh <-chan time.Time
	if timeoutMillis >= 0 {
		// If duration is not nil, instantiate the timer.
		after := time.NewTimer(time.Duration(timeoutMillis) * time.Millisecond)
		defer after.Stop()
		afterCh = after.C
	}

	for {
		select {
		case <-afterCh:
			return 0, 0
		case <-tickCh:
			npipes, nsockets, errno := peekAll(pipes, sockets)
			if errno != 0 {
				return -1, errno
			}
			count = nregular + npipes + nsockets
			if count > 0 {
				return count, 0
			}
		}
	}
}

func peekAll(pipes, sockets []pollFd) (npipes, nsockets int, errno sys.Errno) {
	npipes, errno = peekPipes(pipes)
	if errno != 0 {
		return
	}

	// Invoke wsaPoll with a 0-timeout to avoid blocking.
	// Timeouts are handled in pollWithContext instead.
	nsockets, errno = wsaPoll(sockets, 0)
	if errno != 0 {
		return
	}

	count := npipes + nsockets
	if count > 0 {
		return
	}

	return
}

func peekPipes(fds []pollFd) (n int, errno sys.Errno) {
	for _, fd := range fds {
		bytes, errno := peekNamedPipe(syscall.Handle(fd.fd))
		if errno != 0 {
			return -1, sys.UnwrapOSError(errno)
		}
		if bytes > 0 {
			n++
		}
	}
	return
}

// wsaPoll is the WSAPoll function from winsock2.
//
// See https://learn.microsoft.com/en-us/windows/win32/api/winsock2/nf-winsock2-wsapoll
func wsaPoll(fds []pollFd, timeout int) (n int, errno sys.Errno) {
	if len(fds) > 0 {
		sockptr := &fds[0]
		ns, _, e := syscall.SyscallN(
			procWSAPoll.Addr(),
			uintptr(unsafe.Pointer(sockptr)),
			uintptr(len(fds)),
			uintptr(timeout))
		if e != 0 {
			return -1, sys.UnwrapOSError(e)
		}
		n = int(ns)
	}
	return
}

// ftype is a type of file that can be handled by poll.
type ftype uint8

const (
	ftype_regular ftype = iota
	ftype_pipe
	ftype_socket
)

// partionByFtype checks the type of each fd in fds and returns 3 distinct partitions
// for regular files, named pipes and sockets.
func partionByFtype(fds []pollFd) (regular, pipe, socket []pollFd, errno sys.Errno) {
	for _, pfd := range fds {
		t, errno := ftypeOf(pfd.fd)
		if errno != 0 {
			return nil, nil, nil, errno
		}
		switch t {
		case ftype_regular:
			regular = append(regular, pfd)
		case ftype_pipe:
			pipe = append(pipe, pfd)
		case ftype_socket:
			socket = append(socket, pfd)
		}
	}
	return
}

// mergePartitions merges the given partitions back to the original slice and
// updates the revents field of each pollFd.
func mergePartitions(dst []pollFd, partitions ...[]pollFd) {
	for _, p := range partitions {
	LOOP_EACH_FD_IN_PARTITION:
		for _, pfd := range p {
			for i, fd := range dst {
				if fd.fd == pfd.fd && fd.events == pfd.events {
					dst[i].revents = pfd.revents

					// special case: POLLIN combines POLLRDNORM and POLLRDBAND and
					// when one of them is set, we need to set the other to maintain
					// consistency with the Linux implementation.
					if pfd.revents&_POLLIN != 0 {
						dst[i].revents |= _POLLIN
					}

					continue LOOP_EACH_FD_IN_PARTITION // go to next fd in partition, we assume dst is unique
				}
			}
		}
	}
}

// ftypeOf checks the type of fd and return the corresponding ftype.
func ftypeOf(fd uintptr) (ftype, sys.Errno) {
	h := syscall.Handle(fd)
	t, err := syscall.GetFileType(h)
	if err != nil {
		return 0, sys.UnwrapOSError(err)
	}
	switch t {
	case syscall.FILE_TYPE_CHAR, syscall.FILE_TYPE_DISK:
		return ftype_regular, 0
	case syscall.FILE_TYPE_PIPE:
		if isSocket(h) {
			return ftype_socket, 0
		} else {
			return ftype_pipe, 0
		}
	default:
		return ftype_regular, 0
	}
}

// isSocket returns true if the given file handle
// is a pipe.
func isSocket(fd syscall.Handle) bool {
	r, _, errno := syscall.SyscallN(
		procGetNamedPipeInfo.Addr(),
		uintptr(fd),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)),
		uintptr(unsafe.Pointer(nil)))
	return r == 0 || errno != 0
}
