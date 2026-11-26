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
	publishLockPrefix        = "rslint-lock:"
	publishLockMaxCandidates = 128
)

func publishLockName(idx int) string {
	if idx == 0 {
		return ".rslint-suppressions-publish.lock"
	}
	return fmt.Sprintf(".rslint-suppressions-publish-%d.lock", idx)
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

// publishWithLockFile serializes first-time publication on filesystems that
// support neither exclusive rename nor hard links (such as macOS exFAT).
// It keeps path strictly absent until os.Rename publishes the complete JSON
// temp file, so concurrent readers or an interrupted first save never leave an
// empty file at path.
func publishWithLockFile(ctx context.Context, tempName, path string) error {
	dir, base := filepath.Dir(path), filepath.Base(path)
	candidateIdx := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := os.Lstat(path); err == nil {
			return fs.ErrExist
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}

		var lockPath string
		for {
			if candidateIdx >= publishLockMaxCandidates {
				return fmt.Errorf("cannot allocate suppressions publish lock in %q", dir)
			}
			name := publishLockName(candidateIdx)
			if !strings.EqualFold(base, name) {
				lockPath = filepath.Join(dir, name)
				break
			}
			candidateIdx++
		}

		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR|unix.O_NOFOLLOW, 0o600)
		if err != nil {
			if errors.Is(err, syscall.EISDIR) || errors.Is(err, syscall.ELOOP) || errors.Is(err, syscall.EMLINK) {
				candidateIdx++
				continue
			}
			if errors.Is(err, os.ErrPermission) {
				if _, statErr := os.Lstat(lockPath); statErr == nil {
					candidateIdx++
					continue
				}
			}
			return err
		}

		ofd, err := lockUnixFile(f)
		if err != nil {
			_ = f.Close()
			return err
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
			candidateIdx++
			continue
		}

		// Never modify or remove an unrelated non-empty user file at lockPath.
		var header [len(publishLockPrefix)]byte
		n, readErr := f.ReadAt(header[:], 0)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return readErr
		}
		if fdInfo.Size() > 0 && (n < len(publishLockPrefix) || string(header[:n]) != publishLockPrefix) {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			candidateIdx++
			continue
		}

		// Write a unique per-acquisition token and read it back by path so
		// verification remains exact even on filesystems (like exFAT) that may
		// reuse inode numbers after an unlink.
		token := publishLockPrefix + rand.Text()
		if fdInfo.Size() != 0 && fdInfo.Size() != int64(len(token)) {
			if err := f.Truncate(0); err != nil {
				unlockUnixFile(f, ofd)
				_ = f.Close()
				return err
			}
		}
		if _, err := f.WriteAt([]byte(token), 0); err != nil {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return err
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
		var verifyBuf [len(publishLockPrefix) + 64]byte
		verifyN, verifyReadErr := io.ReadFull(verifyFile, verifyBuf[:len(token)+1])
		_ = verifyFile.Close()
		if verifyStatErr != nil {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return verifyStatErr
		}
		if verifyReadErr != nil && !errors.Is(verifyReadErr, io.EOF) && !errors.Is(verifyReadErr, io.ErrUnexpectedEOF) {
			unlockUnixFile(f, ofd)
			_ = f.Close()
			return verifyReadErr
		}
		if !verifyInfo.Mode().IsRegular() || !os.SameFile(fdInfo, verifyInfo) ||
			verifyN != len(token) || string(verifyBuf[:verifyN]) != token {
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

		_ = os.Remove(lockPath)
		unlockUnixFile(f, ofd)
		_ = f.Close()
		return publishErr
	}
}
