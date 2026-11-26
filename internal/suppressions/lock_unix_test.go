//go:build unix && !aix

package suppressions

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
	"gotest.tools/v3/assert"
)

func TestUpdateInNonListableWritableDirectory(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	dir := filepath.Join(cwd, "write-only")
	assert.NilError(t, os.Mkdir(dir, 0o333))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if _, err := os.ReadDir(dir); err == nil {
		t.Skip("current user can list a 0333 directory")
	} else if !errors.Is(err, os.ErrPermission) {
		t.Fatal(err)
	}

	const location = "write-only/base.json"
	first, err := Open(cwd, Options{Mode: Record, Location: location})
	assert.NilError(t, err)
	_, _, err = reportError(t.Context(), first, filepath.Join(cwd, "a.js"), "no-debugger")
	assert.NilError(t, err)

	second, err := Open(cwd, Options{Mode: Record, Location: location})
	assert.NilError(t, err)
	_, _, err = reportError(t.Context(), second, filepath.Join(cwd, "b.js"), "no-debugger")
	assert.NilError(t, err)

	reopened, err := Open(cwd, Options{Location: location})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.DeepEqual(t, reopened.entries, entries{
		"a.js": {"no-debugger": {Count: 1}},
		"b.js": {"no-debugger": {Count: 1}},
	})
}

func TestUpdateDoesNotBlockOnOuterDirectoryFlock(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	dir, err := os.Open(cwd)
	assert.NilError(t, err)
	defer func() { _ = dir.Close() }()
	assert.NilError(t, unix.Flock(int(dir.Fd()), unix.LOCK_EX))
	defer func() { _ = unix.Flock(int(dir.Fd()), unix.LOCK_UN) }()

	done := make(chan error, 1)
	go func() {
		first, err := Open(cwd, Options{Mode: Record})
		if err != nil {
			done <- err
			return
		}
		if _, _, err := reportError(t.Context(), first, filepath.Join(cwd, "a.js"), "no-debugger"); err != nil {
			done <- err
			return
		}
		second, err := Open(cwd, Options{Mode: Record})
		if err != nil {
			done <- err
			return
		}
		_, _, err = reportError(t.Context(), second, filepath.Join(cwd, "b.js"), "no-debugger")
		done <- err
	}()

	select {
	case err := <-done:
		assert.NilError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("baseline update deadlocked while outer process held directory flock")
	}

	reopened, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.DeepEqual(t, reopened.entries, entries{
		"a.js": {"no-debugger": {Count: 1}},
		"b.js": {"no-debugger": {Count: 1}},
	})
}
