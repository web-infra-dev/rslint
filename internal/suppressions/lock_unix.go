//go:build unix && !aix

package suppressions

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type lockedBaseline struct {
	file   *os.File
	mode   fs.FileMode
	unlock func()
}

func readBaselineFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func flockExclusive(f *os.File) error {
	for {
		err := unix.Flock(int(f.Fd()), unix.LOCK_EX)
		if !errors.Is(err, unix.EINTR) {
			return err
		}
	}
}

func lockExistingBaseline(ctx context.Context, path string) (*lockedBaseline, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Open O_RDWR so Linux NFS accepts exclusive locks and read-only
		// baselines fail with os.ErrPermission before staging a replacement.
		f, err := os.OpenFile(path, os.O_RDWR, 0)
		if err != nil {
			if errors.Is(err, syscall.EISDIR) {
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return nil, err
		}
		ofd, err := lockUnixFile(f)
		if err != nil {
			_ = f.Close()
			return nil, err
		}
		fdInfo, fdErr := f.Stat()
		pathInfo, pathErr := os.Stat(path)
		if fdErr == nil && pathErr == nil && os.SameFile(fdInfo, pathInfo) {
			if !fdInfo.Mode().IsRegular() {
				unlockUnixFile(f, ofd)
				_ = f.Close()
				return nil, fmt.Errorf("suppressions file %q must be a regular file", path)
			}
			return &lockedBaseline{
				file: f,
				mode: fdInfo.Mode().Perm(),
				unlock: func() {
					unlockUnixFile(f, ofd)
					_ = f.Close()
				},
			}, nil
		}
		unlockUnixFile(f, ofd)
		_ = f.Close()
		if fdErr != nil {
			return nil, fdErr
		}
		if pathErr != nil {
			if errors.Is(pathErr, fs.ErrNotExist) {
				return nil, fs.ErrNotExist
			}
			return nil, pathErr
		}
	}
}

func replaceLockedBaseline(_ context.Context, tempName, path string, _ *lockedBaseline) error {
	return os.Rename(tempName, path)
}

const (
	// Binary signature written only into coordination files created by rslint.
	// Leading NUL bytes and fixed RFC 4648 base32 token length avoid matching
	// pre-existing user files.
	publishLockMagic    = "\x00rslint-suppressions-publish-lock-v1\x00"
	publishLockTokenLen = 26 // crypto/rand.Text() produces 26 uppercase base32 chars.
	publishLockSize     = len(publishLockMagic) + publishLockTokenLen + 1
)

func publishLockPath(path string) string {
	dir, base := filepath.Dir(path), filepath.Base(path)
	name := ".rslint-suppressions-publish.lock"
	if strings.EqualFold(base, name) {
		name = ".rslint-suppressions-publish-alt.lock"
	}
	return filepath.Join(dir, name)
}

func isOwnedPublishLockBytes(b []byte) bool {
	if len(b) != publishLockSize {
		return false
	}
	if string(b[:len(publishLockMagic)]) != publishLockMagic {
		return false
	}
	if b[publishLockSize-1] != '\n' {
		return false
	}
	for _, ch := range b[len(publishLockMagic) : publishLockSize-1] {
		if (ch < 'A' || ch > 'Z') && (ch < '2' || ch > '7') {
			return false
		}
	}
	return true
}

func publishWithHardLink(tempName, path string) (bool, error) {
	err := os.Link(tempName, path)
	if err == nil {
		_ = os.Remove(tempName)
		return true, nil
	}
	if errors.Is(err, fs.ErrExist) {
		return true, fs.ErrExist
	}
	if errors.Is(err, syscall.ENOTSUP) || errors.Is(err, syscall.EOPNOTSUPP) ||
		errors.Is(err, syscall.ENOSYS) || errors.Is(err, syscall.EPERM) ||
		errors.Is(err, syscall.EXDEV) {
		return false, nil
	}
	return true, err
}

func openOrCreatePublishLock(lockPath string) (f *os.File, createdByUs bool, alreadyLocked bool, err error) {
	// Use mode 0o666 (umask-adjusted, matching the staged baseline file) so
	// collaborating accounts with shared write access contend on the same lock.
	createFlags := os.O_CREATE | os.O_EXCL | os.O_RDWR | publishCreateExtraFlags
	f, err = os.OpenFile(lockPath, createFlags, 0o666)
	if err == nil {
		return f, true, publishCreateExtraFlags != 0, nil
	}
	if publishCreateExtraFlags != 0 &&
		(errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EOPNOTSUPP) || errors.Is(err, unix.EINVAL)) {
		f, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o666)
		if err == nil {
			return f, true, false, nil
		}
	}
	if !errors.Is(err, fs.ErrExist) {
		return nil, false, false, err
	}
	f, err = os.OpenFile(lockPath, os.O_RDWR|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, false, false, err
	}
	return f, false, false, nil
}

// publishWithLockFile serializes first-time publication on filesystems that
// support neither exclusive rename nor hard links (such as macOS exFAT).
// All writers targeting path coordinate on one deterministic lock path so
// concurrent accounts either contend on the same inode or fail before renaming.
// Pre-existing files without an rslint ownership signature are never modified
// or removed, and path stays strictly absent until os.Rename publishes the
// complete JSON temp file.
func publishWithLockFile(ctx context.Context, tempName, path string) error {
	lockPath := publishLockPath(path)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := os.Lstat(path); err == nil {
			return fs.ErrExist
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		f, createdByUs, alreadyLocked, err := openOrCreatePublishLock(lockPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			// If another writer already finished publishing path while we failed
			// to open the transient lock file, let the caller merge into path.
			if _, statErr := os.Lstat(path); statErr == nil {
				return fs.ErrExist
			}
			return fmt.Errorf("cannot open suppressions publish lock %q: %w", lockPath, err)
		}

		var ofd bool
		if !alreadyLocked {
			ofd, err = lockUnixFile(f)
			if err != nil {
				_ = f.Close()
				return err
			}
		}

		fdInfo, fdErr := f.Stat()
		if fdErr != nil {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return fdErr
		}
		if !fdInfo.Mode().IsRegular() {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return fmt.Errorf("suppressions publish lock %q must be a regular file", lockPath)
		}

		owned := createdByUs
		if !owned && fdInfo.Size() == int64(publishLockSize) {
			var existing [publishLockSize + 1]byte
			n, readErr := f.ReadAt(existing[:], 0)
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				unlockUnixFile(f, ofd)
				_ = f.Close()
				return readErr
			}
			owned = isOwnedPublishLockBytes(existing[:n])
		}

		var token string
		if owned {
			// Write a unique per-acquisition token so verification remains exact
			// even on filesystems (like exFAT) that may reuse inode numbers
			// after an unlink.
			token = publishLockMagic + rand.Text() + "\n"
			if _, err := f.WriteAt([]byte(token), 0); err != nil {
				if createdByUs {
					_ = os.Remove(lockPath)
				}
				unlockUnixFile(f, ofd)
				_ = f.Close()
				return err
			}
		}

		verifyFile, verifyErr := os.OpenFile(lockPath, os.O_RDONLY|unix.O_NOFOLLOW, 0)
		if verifyErr != nil {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			if errors.Is(verifyErr, fs.ErrNotExist) || errors.Is(verifyErr, syscall.EISDIR) ||
				errors.Is(verifyErr, syscall.ELOOP) || errors.Is(verifyErr, syscall.EMLINK) {
				continue
			}
			return verifyErr
		}
		verifyInfo, verifyStatErr := verifyFile.Stat()
		var verifyBuf [publishLockSize + 1]byte
		var verifyN int
		var verifyReadErr error
		if owned {
			verifyN, verifyReadErr = io.ReadFull(verifyFile, verifyBuf[:])
		}
		_ = verifyFile.Close()
		if verifyStatErr != nil {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return verifyStatErr
		}
		if owned && verifyReadErr != nil && !errors.Is(verifyReadErr, io.EOF) && !errors.Is(verifyReadErr, io.ErrUnexpectedEOF) {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return verifyReadErr
		}
		if !verifyInfo.Mode().IsRegular() || !os.SameFile(fdInfo, verifyInfo) ||
			(owned && (verifyN != len(token) || string(verifyBuf[:verifyN]) != token)) {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			continue
		}

		var publishErr error
		if _, statErr := os.Lstat(path); statErr == nil {
			publishErr = fs.ErrExist
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			publishErr = statErr
		} else {
			publishErr = os.Rename(tempName, path)
		}

		if owned {
			_ = os.Remove(lockPath)
		}
		unlockUnixFile(f, ofd)
		_ = f.Close()
		return publishErr
	}
}
