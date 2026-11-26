//go:build (!unix || aix) && !windows

package suppressions

import (
	"context"
	"fmt"
	"io/fs"
	"os"
)

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
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
	}
	return &lockedBaseline{
		file:   f,
		mode:   info.Mode().Perm(),
		unlock: func() { _ = f.Close() },
	}, nil
}

func publishNewBaseline(_ context.Context, tempName, path string) error {
	return os.Rename(tempName, path)
}

func replaceLockedBaseline(_ context.Context, tempName, path string, _ *lockedBaseline) error {
	return os.Rename(tempName, path)
}
