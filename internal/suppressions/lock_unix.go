//go:build unix && !aix

package suppressions

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockBaseline(file *os.File) (*os.File, error) {
	for {
		switch err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err {
		case nil:
			return file, nil
		case unix.EINTR:
			continue
		case unix.EWOULDBLOCK:
			return nil, os.ErrExist
		default:
			return nil, &os.PathError{Op: "lock", Path: file.Name(), Err: err}
		}
	}
}

// A hard link also installs the complete file without overwriting an existing name.
func linkNew(from, to string) error {
	if err := os.Link(from, to); err != nil {
		return err
	}
	_ = os.Remove(from)
	return nil
}

func nativeUpdateError(err error) error { return err }
