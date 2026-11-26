//go:build (!unix || aix) && !windows

package suppressions

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

var fallbackMu sync.Mutex

type lockedBaseline struct {
	file   *os.File
	mode   fs.FileMode
	unlock func()
}

func readBaselineFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func lockExistingBaseline(ctx context.Context, path string) (*lockedBaseline, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fallbackMu.Lock()
	f, err := os.Open(path)
	if err != nil {
		fallbackMu.Unlock()
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		fallbackMu.Unlock()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		fallbackMu.Unlock()
		return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
	}
	return &lockedBaseline{
		file: f,
		mode: info.Mode().Perm(),
		unlock: func() {
			_ = f.Close()
			fallbackMu.Unlock()
		},
	}, nil
}

func publishNewBaseline(_ context.Context, tempName, path string) error {
	if err := os.Link(tempName, path); err == nil {
		_ = os.Remove(tempName)
		return nil
	} else if errors.Is(err, fs.ErrExist) {
		return fs.ErrExist
	}

	fallbackMu.Lock()
	defer fallbackMu.Unlock()
	if _, err := os.Lstat(path); err == nil {
		return fs.ErrExist
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.Rename(tempName, path)
}

func replaceLockedBaseline(_ context.Context, tempName, path string, _ *lockedBaseline) error {
	return os.Rename(tempName, path)
}
