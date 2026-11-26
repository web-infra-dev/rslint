package suppressions

import (
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func resolveLocationPath(cwd, location string) string {
	if filepath.IsAbs(location) {
		return filepath.Clean(location)
	}
	if runtime.GOOS == "windows" {
		if vol := filepath.VolumeName(location); vol != "" {
			if strings.EqualFold(vol, filepath.VolumeName(cwd)) {
				return filepath.Join(cwd, location[len(vol):])
			}
			if abs, err := filepath.Abs(location); err == nil {
				return abs
			}
		} else if len(location) > 0 && os.IsPathSeparator(location[0]) {
			if vol := filepath.VolumeName(cwd); vol != "" {
				return filepath.Clean(vol + location)
			}
		}
	}
	return filepath.Join(cwd, location)
}

func resolvePhysicalDirectory(cwd string) string {
	if runtime.GOOS == "windows" {
		if resolved := osvfs.FS().Realpath(filepath.ToSlash(cwd)); resolved != "" {
			return filepath.FromSlash(resolved)
		}
		return cwd
	}
	// When cwd is the process working directory, syscall.Getwd returns the
	// kernel's physical path (including stored casing on case-insensitive
	// volumes), matching Node's process.cwd().
	if wd, err := syscall.Getwd(); err == nil {
		if cwdInfo, err1 := os.Stat(cwd); err1 == nil {
			if wdInfo, err2 := os.Stat(wd); err2 == nil && os.SameFile(cwdInfo, wdInfo) {
				return wd
			}
		}
	}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		return resolved
	}
	return cwd
}

func validateParentDirectory(dir string) error {
	for cur := dir; ; {
		info, err := os.Stat(cur)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("suppressions parent path %q is not a directory", cur)
			}
			return probeDirectoryWritable(cur)
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if link, lstatErr := os.Lstat(cur); lstatErr == nil && link.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("cannot resolve suppressions parent directory %q", cur)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return err
		}
		cur = parent
	}
}

func probeDirectoryWritable(dir string) error {
	probePath := filepath.Join(dir, ".rslint-suppressions-"+rand.Text()+".tmp")
	probe, err := os.OpenFile(probePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_ = probe.Close()
	_ = os.Remove(probePath)
	return nil
}

// Keep the linted spelling within the working directory so symlinks to the
// checkout root produce portable keys, while intra-repo aliases stay distinct.
func (f *File) fileKey(path string) string {
	if f.logicalCwd != "" && f.logicalCwd != f.cwd {
		if rel, ok := relativeWithinBase(f.logicalCwd, path); ok {
			return rel
		}
	}
	return relativeToBase(f.cwd, path)
}

func relativeWithinBase(base, path string) (string, bool) {
	rel := relativeToBase(base, path)
	if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") || strings.HasPrefix(rel, "/") || filepath.IsAbs(rel) {
		return "", false
	}
	if runtime.GOOS == "windows" && filepath.VolumeName(rel) != "" {
		return "", false
	}
	return rel, true
}

func relativeToBase(base, path string) string {
	if runtime.GOOS == "windows" {
		from, to := filepath.ToSlash(filepath.VolumeName(base)), filepath.ToSlash(filepath.VolumeName(path))
		if !strings.EqualFold(from, to) {
			return utils.NormalizeAbsoluteDrive(path) // Another drive or share has no relative path.
		}
		// Windows volumes are case-insensitive. Inside a volume, preserve the
		// file and directory spellings that the rule actually received.
		return tspath.GetRelativePathFromDirectory(
			cmp.Or(filepath.ToSlash(base[len(from):]), "/"), cmp.Or(path[len(to):], "/"),
			tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true},
		)
	}
	relative, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return relative
}

func (f *File) fileMissing(key string) bool {
	if strings.HasPrefix(key, "//") {
		return false // Do not contact a network share to prune an unchecked file.
	}
	path := filepath.FromSlash(key)
	if filepath.VolumeName(path) == "" && len(key) >= 3 && key[1:3] == ":/" && tspath.IsVolumeCharacter(key[0]) {
		return false // A Unix host cannot check a Windows drive.
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(f.cwd, path)
	}
	// A missing child under a dangling directory link or detached drive may
	// become reachable again. Only treat a path as missing when its nearest
	// existing ancestor resolves cleanly.
	for current := path; ; {
		_, err := os.Lstat(current)
		if err == nil {
			if current == path {
				return false
			}
			_, err = os.Stat(current) // Follow directory links and Windows junctions.
			return err == nil
		}
		parent := filepath.Dir(current)
		if parent == current || (!errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ENOTDIR)) {
			return false
		}
		current = parent
	}
}
