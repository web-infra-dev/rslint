//go:build darwin

package suppressions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"gotest.tools/v3/assert"
)

func TestPublishWithExclusiveOpenContention(t *testing.T) {
	cwd := tempWorkingDirectory(t)
	target := filepath.Join(cwd, "exclusive-open-fallback.json")

	const writers = 16
	start := make(chan struct{})
	var firstWinners atomic.Int32
	errs := make([]error, writers)
	var wg sync.WaitGroup

	for i := range writers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("w-%02d.js", idx)
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
			err = publishWithExclusiveOpen(tempName, target)
			if err == nil {
				firstWinners.Add(1)
				return
			}
			_ = os.Remove(tempName)
			if !errors.Is(err, fs.ErrExist) {
				errs[idx] = fmt.Errorf("writer %d unexpected publish error: %w", idx, err)
				return
			}
			// Contenders merge via the normal locked path and must never observe
			// an empty placeholder file.
			locked, lockErr := lockExistingBaseline(t.Context(), target)
			if lockErr != nil {
				errs[idx] = fmt.Errorf("writer %d lock: %w", idx, lockErr)
				return
			}
			data, readErr := readEntriesFromFile(locked.file, target)
			if readErr != nil {
				locked.unlock()
				errs[idx] = fmt.Errorf("writer %d read: %w", idx, readErr)
				return
			}
			data[key] = map[string]entry{"no-debugger": {Count: 1}}
			writeErr := writeAndPublishBaseline(t.Context(), target, data, locked)
			locked.unlock()
			if writeErr != nil {
				errs[idx] = fmt.Errorf("writer %d merge: %w", idx, writeErr)
			}
		}(i)
	}
	close(start)
	wg.Wait()

	for _, err := range errs {
		assert.NilError(t, err)
	}
	assert.Equal(t, firstWinners.Load(), int32(1))

	reopened, err := Open(cwd, Options{Location: target})
	assert.NilError(t, err)
	assert.Assert(t, reopened != nil)
	assert.Equal(t, len(reopened.entries), writers)
	for i := range writers {
		assert.Equal(t, reopened.entries[fmt.Sprintf("w-%02d.js", i)]["no-debugger"].Count, 1)
	}
}
