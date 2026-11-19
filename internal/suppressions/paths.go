package suppressions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (f *File) resolveWorkingDirectory(cwd string) error {
	if f.baseDir == "" {
		var err error
		f.baseDir, err = syscall.Getwd()
		if err != nil {
			return err
		}
	}
	physical := f.baseDir
	if runtime.GOOS == "windows" {
		// The native resolver follows junctions as well as directory symlinks.
		physical = filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(f.baseDir)))
	}
	f.physicalCwd = physical
	if cwd == physical {
		return nil
	}

	// Only links traversed to reach the working directory may be substituted
	// in a lint target. Remember intermediate links in their targets as well.
	links := make(map[string]string)
	var collect func(string, string) error
	collect = func(parent, path string) error {
		path = filepath.FromSlash(path)
		volume := filepath.VolumeName(path)
		path = path[len(volume):]
		if volume != "" || len(path) > 0 && os.IsPathSeparator(path[0]) {
			if volume == "" {
				volume = filepath.VolumeName(parent)
			}
			parent = volume + string(filepath.Separator)
		}
		for name := range strings.SplitSeq(path, string(filepath.Separator)) {
			if name == "" || name == "." {
				continue
			}
			part := filepath.Join(parent, name)
			if target, ok := links[part]; ok {
				parent = target
				continue
			}
			info, err := os.Lstat(part)
			if err != nil {
				return err
			}
			if info.IsDir() {
				parent = part
				continue
			}
			destination, err := os.Readlink(part)
			if err != nil {
				return err
			}
			target := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(part)))
			if target == part { // Realpath returns its input if the link no longer resolves.
				return fmt.Errorf("cannot resolve working directory link %q", part)
			}
			links[part] = target
			if err := collect(parent, destination); err != nil {
				return err
			}
			f.cwdAliases = append(f.cwdAliases, utils.DirectoryIdentity{LexicalPath: part, CanonicalPath: target})
			parent = target
		}
		return nil
	}
	return collect("", cwd)
}

func (f *File) projectPath(path string) string {
project:
	for {
		for _, dir := range f.cwdAliases {
			suffix, err := relativeFilePath(dir.LexicalPath, path)
			if err == nil && filepath.IsLocal(suffix) {
				path = filepath.Join(dir.CanonicalPath, suffix)
				continue project
			}
		}
		return path
	}
}

func (f *File) fileKey(path string) string {
	path = filepath.FromSlash(path)
	target := f.projectPath(path)
	key, err := relativeFilePath(f.physicalCwd, target)
	// A Windows working directory may itself be a junction. Moving up from
	// its name can land in a different directory than moving up from its target.
	if f.baseDir != f.physicalCwd && !filepath.IsLocal(key) {
		fromBase := f.projectPath(filepath.Join(f.baseDir, key))
		matched, _ := relativeFilePath(fromBase, target)
		if err != nil || matched != "." {
			key, err = relativeFilePath(f.baseDir, path)
		}
	}
	if err != nil {
		key = path // A target on another Windows volume has no relative path.
	}
	return filepath.ToSlash(key)
}

// Volumes may equate case or Unicode spellings. Lstat compares the directory
// entries themselves, so independently named links keep their own paths.
func relativeFilePath(dir, path string) (string, error) {
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(path, dir) {
		return rel, err
	}
	sep := string(filepath.Separator)
	left := strings.TrimPrefix(dir[len(filepath.VolumeName(dir)):], sep)
	right := strings.TrimPrefix(path[len(filepath.VolumeName(path)):], sep)
	nonASCII := func(r rune) bool { return r > unicode.MaxASCII }
	checkedSpelling := false
	for left != "" && right != "" {
		leftName, nextLeft, _ := strings.Cut(left, sep)
		rightName, nextRight, _ := strings.Cut(right, sep)
		if leftName != rightName {
			if !strings.EqualFold(leftName, rightName) && !strings.ContainsFunc(leftName, nonASCII) && !strings.ContainsFunc(rightName, nonASCII) {
				break
			}
			checkedSpelling = true
			leftInfo, _ := os.Lstat(strings.TrimSuffix(dir, left) + leftName)
			rightInfo, _ := os.Lstat(strings.TrimSuffix(path, right) + rightName)
			if !os.SameFile(leftInfo, rightInfo) {
				break
			}
		}
		left, right = nextLeft, nextRight
	}
	if checkedSpelling {
		up, _ := filepath.Rel(dir, strings.TrimSuffix(dir, left))
		rel = filepath.Join(up, right)
	}
	return rel, nil
}

// ESLint records the actual input spelling relative to process.cwd(). Keep an
// existing entry under that exact name. Newly recorded files use a stable key
// that can also be checked from the physical working directory.
func (f *File) entryKey(data entries, path, inputPath string) string {
	if len(data) > 0 {
		observed := filepath.FromSlash(path)
		if inputPath != "" && inputPath != path && f.fileKey(inputPath) == f.fileKey(path) {
			observed = filepath.FromSlash(inputPath)
		}
		rel := filepath.Rel
		if runtime.GOOS == "windows" {
			// A case-sensitive Windows directory can contain both spellings.
			rel = relativeFilePath
		}
		key, err := rel(f.baseDir, observed)
		if err != nil {
			key = observed
		}
		if key = filepath.ToSlash(key); data[key] != nil {
			return key
		}
	}
	return f.fileKey(path)
}
