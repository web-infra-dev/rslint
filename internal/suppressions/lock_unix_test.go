//go:build unix && !aix

package suppressions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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

func TestPublishWithLockFileContentionAndPublicReaders(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	target := filepath.Join(cwd, DefaultFilename)

	const writers = 16
	const readers = 4
	start := make(chan struct{})
	stopReaders := make(chan struct{})
	var firstWinners atomic.Int32
	var readerChecks atomic.Int64
	writerErrs := make([]error, writers)
	readerErrs := make([]error, readers)

	var readerWG sync.WaitGroup
	for r := range readers {
		readerWG.Add(1)
		go func(idx int) {
			defer readerWG.Done()
			<-start
			for {
				select {
				case <-stopReaders:
					return
				default:
				}
				// Default-location check must observe either absent (nil, nil)
				// or a valid non-empty baseline, never an empty or partial file.
				opened, err := Open(cwd, Options{})
				if err != nil {
					readerErrs[idx] = fmt.Errorf("default reader %d: %w", idx, err)
					return
				}
				if opened != nil && len(opened.entries) == 0 {
					readerErrs[idx] = fmt.Errorf("default reader %d saw empty baseline", idx)
					return
				}
				// Explicit-location check must return either "does not exist"
				// or a valid non-empty baseline, never a JSON parse error.
				explicit, err := Open(cwd, Options{Location: DefaultFilename})
				if err != nil {
					if !strings.Contains(err.Error(), "does not exist") {
						readerErrs[idx] = fmt.Errorf("explicit reader %d: %w", idx, err)
						return
					}
				} else if explicit == nil || len(explicit.entries) == 0 {
					readerErrs[idx] = fmt.Errorf("explicit reader %d saw empty baseline", idx)
					return
				}
				readerChecks.Add(1)
			}
		}(r)
	}

	var writerWG sync.WaitGroup
	for i := range writers {
		writerWG.Add(1)
		go func(idx int) {
			defer writerWG.Done()
			key := fmt.Sprintf("w-%02d.js", idx)
			tempName := filepath.Join(cwd, fmt.Sprintf(".stage-%02d.tmp", idx))
			payload, err := json.Marshal(entries{key: {"no-debugger": {Count: 1}}})
			if err != nil {
				writerErrs[idx] = err
				return
			}
			if err := os.WriteFile(tempName, payload, 0o666); err != nil {
				writerErrs[idx] = err
				return
			}
			<-start
			err = publishWithLockFile(t.Context(), tempName, target)
			if err == nil {
				firstWinners.Add(1)
				return
			}
			_ = os.Remove(tempName)
			if !errors.Is(err, fs.ErrExist) {
				writerErrs[idx] = fmt.Errorf("writer %d unexpected publish error: %w", idx, err)
				return
			}
			locked, lockErr := lockExistingBaseline(t.Context(), target)
			if lockErr != nil {
				writerErrs[idx] = fmt.Errorf("writer %d lock: %w", idx, lockErr)
				return
			}
			data, readErr := readEntriesFromFile(locked.file, target)
			if readErr != nil {
				locked.unlock()
				writerErrs[idx] = fmt.Errorf("writer %d read: %w", idx, readErr)
				return
			}
			data[key] = map[string]entry{"no-debugger": {Count: 1}}
			writeErr := writeAndPublishBaseline(t.Context(), target, data, locked)
			locked.unlock()
			if writeErr != nil {
				writerErrs[idx] = fmt.Errorf("writer %d merge: %w", idx, writeErr)
			}
		}(i)
	}

	close(start)
	writerWG.Wait()
	close(stopReaders)
	readerWG.Wait()

	for _, err := range writerErrs {
		assert.NilError(t, err)
	}
	for _, err := range readerErrs {
		assert.NilError(t, err)
	}
	assert.Equal(t, firstWinners.Load(), int32(1))
	assert.Assert(t, readerChecks.Load() > 0)

	reopened, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.Equal(t, len(reopened.entries), writers)
	for i := range writers {
		assert.Equal(t, reopened.entries[fmt.Sprintf("w-%02d.js", i)]["no-debugger"].Count, 1)
	}
	// No coordination lock file or staged temp file should remain on disk.
	dirEntries, err := os.ReadDir(cwd)
	assert.NilError(t, err)
	assert.Equal(t, len(dirEntries), 1)
	assert.Equal(t, dirEntries[0].Name(), DefaultFilename)
}

func TestPublishWithLockFilePausedAndInterrupted(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	target := filepath.Join(cwd, DefaultFilename)
	lockPath := publishLockPath(target)

	// Hold the coordination lock as a writer paused before renaming into target.
	held, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o666)
	assert.NilError(t, err)
	ofd, err := lockUnixFile(held)
	assert.NilError(t, err)
	_, err = held.WriteAt([]byte(publishLockMagic+"ABCDEFGHIJKLMNOPQRSTUVWXYZ\n"), 0)
	assert.NilError(t, err)

	// While the first writer is paused inside its lock, public readers must
	// observe the final baseline as absent rather than an empty file.
	defaultOpened, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, defaultOpened == nil)
	_, err = Open(cwd, Options{Location: DefaultFilename})
	assert.ErrorContains(t, err, "does not exist")

	// Start a second publisher; it must block while the first writer holds the lock.
	secondTemp := filepath.Join(cwd, ".stage-second.tmp")
	assert.NilError(t, os.WriteFile(secondTemp, []byte(`{"b.js":{"no-debugger":{"count":1}}}`), 0o666))
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- publishWithLockFile(t.Context(), secondTemp, target)
	}()

	select {
	case err := <-secondDone:
		t.Fatalf("second publisher did not wait on held coordination lock (err=%v)", err)
	case <-time.After(50 * time.Millisecond):
	}

	// Simulate the first writer terminating before os.Rename (kernel releases
	// its lock, leaving only the coordination lock file while target stays absent).
	unlockUnixFile(held, ofd)
	assert.NilError(t, held.Close())

	select {
	case err := <-secondDone:
		assert.NilError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("second publisher did not finish after interrupted first writer released lock")
	}

	reopened, err := Open(cwd, Options{})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.DeepEqual(t, reopened.entries, entries{
		"b.js": {"no-debugger": {Count: 1}},
	})
	// The recovery publish also cleans up the interrupted writer's lock file.
	dirEntries, err := os.ReadDir(cwd)
	assert.NilError(t, err)
	assert.Equal(t, len(dirEntries), 1)
	assert.Equal(t, dirEntries[0].Name(), DefaultFilename)
}

func TestPublishWithLockFilePreservesPreExistingUserFiles(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{name: "empty", content: ""},
		{name: "similar-prefix", content: "rslint-lock: personal build note"},
		{name: "ordinary-text", content: "keep this file untouched\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cwd := tempWorkingDirectory(t)
			target := filepath.Join(cwd, DefaultFilename)
			lockPath := publishLockPath(target)
			assert.NilError(t, os.WriteFile(lockPath, []byte(tc.content), 0o644))

			// Concurrent writers must still serialize on the pre-existing file
			// and preserve all updates without modifying or deleting it.
			const writers = 8
			start := make(chan struct{})
			var firstWinners atomic.Int32
			errs := make([]error, writers)
			var wg sync.WaitGroup
			for i := range writers {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					key := fmt.Sprintf("u-%02d.js", idx)
					tempName := filepath.Join(cwd, fmt.Sprintf(".stage-%02d.tmp", idx))
					payload, err := json.Marshal(entries{key: {"no-debugger": {Count: 1}}})
					if err != nil {
						errs[idx] = err
						return
					}
					if err := os.WriteFile(tempName, payload, 0o666); err != nil {
						errs[idx] = err
						return
					}
					<-start
					err = publishWithLockFile(t.Context(), tempName, target)
					if err == nil {
						firstWinners.Add(1)
						return
					}
					_ = os.Remove(tempName)
					if !errors.Is(err, fs.ErrExist) {
						errs[idx] = fmt.Errorf("writer %d: %w", idx, err)
						return
					}
					locked, lockErr := lockExistingBaseline(t.Context(), target)
					if lockErr != nil {
						errs[idx] = lockErr
						return
					}
					data, readErr := readEntriesFromFile(locked.file, target)
					if readErr != nil {
						locked.unlock()
						errs[idx] = readErr
						return
					}
					data[key] = map[string]entry{"no-debugger": {Count: 1}}
					writeErr := writeAndPublishBaseline(t.Context(), target, data, locked)
					locked.unlock()
					if writeErr != nil {
						errs[idx] = writeErr
					}
				}(i)
			}
			close(start)
			wg.Wait()
			for _, err := range errs {
				assert.NilError(t, err)
			}
			assert.Equal(t, firstWinners.Load(), int32(1))

			gotLock, err := os.ReadFile(lockPath)
			assert.NilError(t, err)
			assert.Equal(t, string(gotLock), tc.content)

			reopened, err := Open(cwd, Options{})
			assert.NilError(t, err)
			assert.Assert(t, reopened != nil)
			assert.Equal(t, len(reopened.entries), writers)
		})
	}

	t.Run("baseline-named-after-primary-lock", func(t *testing.T) {
		cwd := tempWorkingDirectory(t)
		targetNamedLock := publishLockPath(filepath.Join(cwd, DefaultFilename))
		tempName := filepath.Join(cwd, ".stage-named-lock.tmp")
		payload := `{"a.js":{"no-debugger":{"count":1}}}`
		assert.NilError(t, os.WriteFile(tempName, []byte(payload), 0o666))
		assert.NilError(t, publishWithLockFile(t.Context(), tempName, targetNamedLock))

		got, err := os.ReadFile(targetNamedLock)
		assert.NilError(t, err)
		assert.Equal(t, string(got), payload)
		_, err = os.Lstat(publishLockPath(targetNamedLock))
		assert.Assert(t, errors.Is(err, fs.ErrNotExist))
	})
}

func TestPublishWithLockFileRejectsInaccessibleLockInsteadOfBypassing(t *testing.T) {
	t.Run("directory", func(t *testing.T) {
		cwd := tempWorkingDirectory(t)
		target := filepath.Join(cwd, DefaultFilename)
		lockPath := publishLockPath(target)
		assert.NilError(t, os.Mkdir(lockPath, 0o755))

		tempName := filepath.Join(cwd, ".stage-dir.tmp")
		assert.NilError(t, os.WriteFile(tempName, []byte(`{"a.js":{"no-debugger":{"count":1}}}`), 0o666))
		err := publishWithLockFile(t.Context(), tempName, target)
		assert.Assert(t, err != nil)
		_, statErr := os.Lstat(target)
		assert.Assert(t, errors.Is(statErr, fs.ErrNotExist))
		info, lockStatErr := os.Lstat(lockPath)
		assert.NilError(t, lockStatErr)
		assert.Assert(t, info.IsDir())
	})

	t.Run("symlink", func(t *testing.T) {
		cwd := tempWorkingDirectory(t)
		target := filepath.Join(cwd, DefaultFilename)
		lockPath := publishLockPath(target)
		victim := filepath.Join(cwd, "victim.txt")
		assert.NilError(t, os.WriteFile(victim, []byte("keep"), 0o644))
		assert.NilError(t, os.Symlink(victim, lockPath))

		tempName := filepath.Join(cwd, ".stage-link.tmp")
		assert.NilError(t, os.WriteFile(tempName, []byte(`{"a.js":{"no-debugger":{"count":1}}}`), 0o666))
		err := publishWithLockFile(t.Context(), tempName, target)
		assert.Assert(t, err != nil)
		_, statErr := os.Lstat(target)
		assert.Assert(t, errors.Is(statErr, fs.ErrNotExist))
		gotVictim, readErr := os.ReadFile(victim)
		assert.NilError(t, readErr)
		assert.Equal(t, string(gotVictim), "keep")
	})

	t.Run("read-only-file", func(t *testing.T) {
		cwd := tempWorkingDirectory(t)
		target := filepath.Join(cwd, DefaultFilename)
		lockPath := publishLockPath(target)
		assert.NilError(t, os.WriteFile(lockPath, []byte("keep"), 0o444))
		t.Cleanup(func() { _ = os.Chmod(lockPath, 0o644) })
		if f, err := os.OpenFile(lockPath, os.O_RDWR, 0); err == nil {
			_ = f.Close()
			t.Skip("current user can open 0444 files read-write")
		}

		tempName := filepath.Join(cwd, ".stage-ro.tmp")
		assert.NilError(t, os.WriteFile(tempName, []byte(`{"a.js":{"no-debugger":{"count":1}}}`), 0o666))
		err := publishWithLockFile(t.Context(), tempName, target)
		assert.Assert(t, errors.Is(err, os.ErrPermission), "expected permission error, got %v", err)
		_, statErr := os.Lstat(target)
		assert.Assert(t, errors.Is(statErr, fs.ErrNotExist))
		got, readErr := os.ReadFile(lockPath)
		assert.NilError(t, readErr)
		assert.Equal(t, string(got), "keep")
	})

	t.Run("respects-umask-for-shared-writers", func(t *testing.T) {
		cwd := tempWorkingDirectory(t)
		lockPath := publishLockPath(filepath.Join(cwd, DefaultFilename))
		probe := filepath.Join(cwd, ".umask-probe")
		pf, err := os.OpenFile(probe, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o666)
		assert.NilError(t, err)
		probeInfo, err := pf.Stat()
		assert.NilError(t, err)
		assert.NilError(t, pf.Close())

		f, createdByUs, alreadyLocked, err := openOrCreatePublishLock(lockPath)
		assert.NilError(t, err)
		defer func() {
			_ = os.Remove(lockPath)
			if alreadyLocked {
				unlockUnixFile(f, false)
			}
			_ = f.Close()
		}()
		assert.Assert(t, createdByUs)
		lockInfo, err := f.Stat()
		assert.NilError(t, err)
		assert.Equal(t, lockInfo.Mode().Perm(), probeInfo.Mode().Perm())
	})
}

func TestPublishWithLockFileInNonListableDir(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	dir := filepath.Join(cwd, "write-only-publish")
	assert.NilError(t, os.Mkdir(dir, 0o333))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if _, err := os.ReadDir(dir); err == nil {
		t.Skip("current user can list a 0333 directory")
	}

	target := filepath.Join(dir, DefaultFilename)
	tempName := filepath.Join(dir, ".stage-non-listable.tmp")
	payload := `{"a.js":{"no-debugger":{"count":1}}}`
	assert.NilError(t, os.WriteFile(tempName, []byte(payload), 0o666))
	assert.NilError(t, publishWithLockFile(t.Context(), tempName, target))

	got, err := os.ReadFile(target)
	assert.NilError(t, err)
	assert.Equal(t, string(got), payload)
	_, err = os.Lstat(publishLockPath(target))
	assert.Assert(t, errors.Is(err, fs.ErrNotExist))
}
