//go:build linux

// cspell:ignore SETLK

package suppressions

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
)

func TestLinuxLockUsesSharedByteRange(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	target := filepath.Join(cwd, DefaultFilename)
	assert.NilError(t, os.WriteFile(target, []byte(`{"a.js":{"no-debugger":{"count":1}}}`), 0o644))

	locked, err := lockExistingBaseline(t.Context(), target)
	assert.NilError(t, err)
	defer locked.unlock()

	// A separate open file description must still be able to read [0, EOF)
	// while the beyond-EOF write lock is held.
	reopened, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.Equal(t, reopened.entries["a.js"]["no-debugger"].Count, 1)

	peer, err := os.OpenFile(target, os.O_RDWR, 0)
	assert.NilError(t, err)
	defer func() { _ = peer.Close() }()

	// Contending on the shared Linux/Windows byte range must fail immediately.
	sameRange := unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: io.SeekStart,
		Start:  baselineLockOffset,
		Len:    baselineLockLength,
	}
	err = unix.FcntlFlock(peer.Fd(), unix.F_OFD_SETLK, &sameRange)
	if errors.Is(err, unix.EINVAL) || errors.Is(err, unix.ENOSYS) {
		t.Skip("kernel does not support OFD locks")
	}
	assert.Assert(t, errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EACCES), "expected contention error, got %v", err)

	// Adjacent bytes outside the shared 1-byte lock range remain unlocked.
	adjacentRange := unix.Flock_t{
		Type:   unix.F_WRLCK,
		Whence: io.SeekStart,
		Start:  baselineLockOffset + baselineLockLength,
		Len:    1,
	}
	assert.NilError(t, unix.FcntlFlock(peer.Fd(), unix.F_OFD_SETLK, &adjacentRange))
	adjacentRange.Type = unix.F_UNLCK
	assert.NilError(t, unix.FcntlFlock(peer.Fd(), unix.F_OFD_SETLK, &adjacentRange))
}
