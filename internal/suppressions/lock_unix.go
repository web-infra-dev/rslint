//go:build unix && !aix

package suppressions

import (
	"os"

	"golang.org/x/sys/unix"
)

func openBaselineFile(path string, flags int) (*os.File, error) {
	// A named pipe must reach the regular-file check without waiting for a peer.
	return os.OpenFile(path, flags|unix.O_NONBLOCK, 0)
}

func openReplacementStage(name string, existing *os.File) (*os.File, error) {
	// Start with no mode permissions and apply the original access policy before writing.
	stage, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return nil, err
	}
	if err := copyBaselineAccess(existing, stage); err != nil {
		_ = stage.Close()
		_ = os.Remove(name)
		return nil, err
	}
	return stage, nil
}

func copyBaselineAccess(from, to *os.File) error {
	var old, stage unix.Stat_t
	if err := unix.Fstat(int(from.Fd()), &old); err != nil {
		return err
	}
	if err := unix.Fstat(int(to.Fd()), &stage); err != nil {
		return err
	}
	if old.Uid != stage.Uid || old.Gid != stage.Gid {
		if err := to.Chown(int(old.Uid), int(old.Gid)); err != nil {
			return err
		}
	}
	// Install the ACL before granting group or other mode bits on the new inode.
	if err := copyAccessACL(from, to); err != nil {
		return err
	}
	return to.Chmod(os.FileMode(old.Mode).Perm())
}

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
