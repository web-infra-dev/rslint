//go:build (!unix || aix) && !windows

package suppressions

import (
	"errors"
	"os"
)

func openBaselineFile(path string, flags int) (*os.File, error) { return os.OpenFile(path, flags, 0) }
func openReplacementStage(string, *os.File) (*os.File, error)   { return nil, errors.ErrUnsupported }

// The browser worker uses the service API, which does not write CLI baselines.
func lockBaseline(file *os.File) (*os.File, error) {
	return nil, &os.PathError{Op: "lock", Path: file.Name(), Err: errors.ErrUnsupported}
}

func renameNew(_, _ string) error       { return errors.ErrUnsupported }
func nativeUpdateError(err error) error { return err }
