// Copyright 2024 The WATER Authors. All rights reserved.
// Use of this source code is governed by Apache 2 license
// that can be found in the LICENSE file.

package wasi_snapshot_preview1

import (
	"context"
	"time"

	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental/sys"
	"github.com/tetratelabs/wazero/internal/fsapi"
	internalsys "github.com/tetratelabs/wazero/internal/sys"
	internalsysfs "github.com/tetratelabs/wazero/internal/sysfs"
	"github.com/tetratelabs/wazero/internal/wasip1"
	"github.com/tetratelabs/wazero/internal/wasm"
)

// use the init function to override the default pollOneoffFn
func init() {
	// override the default pollOneoff
	pollOneoff = newHostFunc(
		wasip1.PollOneoffName, alternativePollOneoffFn,
		[]api.ValueType{i32, i32, i32, i32},
		"in", "out", "nsubscriptions", "result.nevents",
	)
}

// alternativePollOneoffFn is a modified version of pollOneoffFn that
// tries to be more syscall-aligned. It should block and return only when
// there is at least one event triggered.
func alternativePollOneoffFn(_ context.Context, mod api.Module, params []uint64) sys.Errno {
	in := uint32(params[0])
	out := uint32(params[1])
	nsubscriptions := uint32(params[2])
	resultNevents := uint32(params[3])

	if nsubscriptions == 0 {
		return sys.EINVAL // early returning on empty subscriptions list
	}

	mem := mod.Memory()

	// Ensure capacity prior to the read loop to reduce error handling.
	inBuf, ok := mem.Read(in, nsubscriptions*48)
	if !ok {
		return sys.EFAULT
	}
	outBuf, ok := mem.Read(out, nsubscriptions*32)
	// zero-out all buffer before writing
	for i := range outBuf {
		outBuf[i] = 0
	}

	if !ok {
		return sys.EFAULT
	}

	// start by writing 0 to resultNevents
	if !mod.Memory().WriteUint32Le(resultNevents, 0) {
		return sys.EFAULT
	}

	// Extract FS context, used in the body of the for loop for FS access.
	fsc := mod.(*wasm.ModuleInstance).Sys.FS()
	// Slice of events that are processed out of the loop (blocking stdin subscribers).
	var blockingStdinSubs []*event
	// The timeout is initialized at max Duration, the loop will find the minimum.
	var timeout time.Duration = 1<<63 - 1
	// Count of all the subscriptions that have been already written back to outBuf.
	// nevents*32 returns at all times the offset where the next event should be written:
	// this way we ensure that there are no gaps between records.
	var nevents uint32

	// Slice of all I/O events that will be written if triggered
	var ioEvents []*event

	// Slice of hostPollSub that will be used for polling
	var hostPollSubs []internalsysfs.PollFd

	// The clock event with the minimum timeout, if any.
	var clkevent *event

	// Layout is subscription_u: Union
	// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#subscription_u
	for i := uint32(0); i < nsubscriptions; i++ {
		inOffset := i * 48
		outOffset := nevents * 32

		eventType := inBuf[inOffset+8] // +8 past userdata
		// +8 past userdata +8 contents_offset
		argBuf := inBuf[inOffset+8+8:]
		userData := inBuf[inOffset : inOffset+8]

		evt := &event{
			eventType: eventType,
			userData:  userData,
			errno:     wasip1.ErrnoSuccess,
		}

		switch eventType {
		case wasip1.EventTypeClock: // handle later
			newTimeout, err := processClockEvent(argBuf)
			if err != 0 {
				return err
			}
			// Min timeout.
			if newTimeout < timeout {
				timeout = newTimeout
				// overwrite the clock event
				clkevent = evt
			}
		case wasip1.EventTypeFdRead:
			guestFd := int32(le.Uint32(argBuf))
			if guestFd < 0 {
				return sys.EBADF
			}

			if file, ok := fsc.LookupFile(guestFd); !ok {
				evt.errno = wasip1.ErrnoBadf
				writeEvent(outBuf[outOffset:], evt)
				nevents++
			} else if guestFd == internalsys.FdStdin { // stdin is always checked with Poll function later.
				if file.File.IsNonblock() { // non-blocking stdin is always ready to read
					writeEvent(outBuf[outOffset:], evt)
					nevents++
				} else {
					// if the fd is Stdin, and it is in blocking mode,
					// do not ack yet, append to a slice for delayed evaluation.
					blockingStdinSubs = append(blockingStdinSubs, evt)
				}
			} else if hostFd := file.File.Fd(); hostFd == 0 {
				evt.errno = wasip1.ErrnoNotsup
				writeEvent(outBuf[outOffset:], evt)
				nevents++
			} else {
				ioEvents = append(ioEvents, evt)
				hostPollSubs = append(hostPollSubs, internalsysfs.PollFd{
					Fd:     hostFd,
					Events: fsapi.POLLIN,
				})
			}
		case wasip1.EventTypeFdWrite:
			guestFd := int32(le.Uint32(argBuf))
			if guestFd < 0 {
				return sys.EBADF
			}

			if file, ok := fsc.LookupFile(guestFd); !ok {
				evt.errno = wasip1.ErrnoBadf
				writeEvent(outBuf[outOffset:], evt)
				nevents++
			} else if guestFd == internalsys.FdStdout || guestFd == internalsys.FdStderr { // stdout and stderr are always ready to write
				writeEvent(outBuf[outOffset:], evt)
				nevents++
			} else if hostFd := file.File.Fd(); hostFd == 0 {
				evt.errno = wasip1.ErrnoNotsup
				writeEvent(outBuf[outOffset:], evt)
				nevents++
			} else {
				ioEvents = append(ioEvents, evt)
				hostPollSubs = append(hostPollSubs, internalsysfs.PollFd{
					Fd:     hostFd,
					Events: fsapi.POLLOUT,
				})
			}
		default:
			return sys.EINVAL
		}
	}

	// We have scanned all the subscriptions, and there are several cases:
	// - Clock subscriptions-only: we block until the timeout expires.
	// - At least one I/O subscription: we call poll on the I/O fds. Then we check the poll results
	//   and write back the corresponding events ONLY if the revent in pollFd is properly set.
	//       - If no clock subscription, we block with max timeout.
	//       - If there are clock subscriptions, we block with the minimum timeout.

	// If there are no I/O subscriptions, we can block until the timeout expires.
	sysCtx := mod.(*wasm.ModuleInstance).Sys
	if len(ioEvents) == 0 {
		if timeout > 0 && clkevent != nil { // there is a clock subscription with a timeout
			sysCtx.Nanosleep(int64(timeout))
		}
		// Ack the clock event if there is one
		if clkevent != nil {
			writeEvent(outBuf[nevents*32:], clkevent)
			nevents++
		}
	}

	// If there are I/O subscriptions, we call poll on the I/O fds with the updated timeout.
	if len(hostPollSubs) > 0 {
		pollNevents, err := internalsysfs.Poll(hostPollSubs, int32(timeout.Milliseconds()))
		if err != 0 {
			return err
		}

		if pollNevents > 0 { // if there are events triggered
			// iterate over hostPollSubs and if the revent is set, write back
			// the event
			for i, pollFd := range hostPollSubs {
				if pollFd.Revents&pollFd.Events != 0 {
					// write back the event
					writeEvent(outBuf[nevents*32:], ioEvents[i])
					nevents++
				} else if pollFd.Revents != 0 {
					// write back the event
					writeEvent(outBuf[nevents*32:], ioEvents[i])
					nevents++
				}
			}
		} else { // otherwise it means that the timeout expired
			// Ack the clock event if there is one (it can also be a default max timeout)
			if clkevent != nil {
				writeEvent(outBuf[nevents*32:], clkevent)
				nevents++
			}
		}
	}

	// If there are blocking stdin subscribers, check for data with given timeout.
	if len(blockingStdinSubs) > 0 {
		stdin, ok := fsc.LookupFile(internalsys.FdStdin)
		if !ok {
			return sys.EBADF
		}

		// Wait for the timeout to expire, or for some data to become available on Stdin.
		if stdinReady, errno := stdin.File.Poll(fsapi.POLLIN, int32(timeout.Milliseconds())); errno != 0 {
			return errno
		} else if stdinReady {
			// stdin has data ready to for reading, write back all the events
			for i := range blockingStdinSubs {
				evt := blockingStdinSubs[i]
				evt.errno = 0
				writeEvent(outBuf[nevents*32:], evt)
				nevents++
			}
		}
	}

	// write nevents to resultNevents
	if !mem.WriteUint32Le(resultNevents, nevents) {
		return sys.EFAULT
	}

	return 0
}
