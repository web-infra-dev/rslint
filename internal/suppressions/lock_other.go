//go:build (!unix || aix) && !windows

package suppressions

import (
	"errors"
	"os"
)

// The browser worker uses the service API, which does not write CLI baselines.
func lockBaseline(file *os.File) (*os.File, error) {
	return nil, &os.PathError{Op: "lock", Path: file.Name(), Err: errors.ErrUnsupported}
}

func renameNew(_, _ string) error       { return errors.ErrUnsupported }
func nativeUpdateError(err error) error { return err }
