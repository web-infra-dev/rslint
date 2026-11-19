package suppressions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const DefaultFilename = "rslint-suppressions.json"

// Open returns nil for the common case of a plain check without a baseline.
func Open(cwd string, options Options) (*File, error) {
	cwd = filepath.FromSlash(cwd)
	f := &File{cwd: cwd, options: options}
	path := filepath.Join(cwd, DefaultFilename)
	if location := options.Location; location != "" {
		path = location
		if filepath.VolumeName(path) != "" || os.IsPathSeparator(path[0]) {
			// Let Windows resolve drive-relative and current-drive-rooted paths.
			var err error
			path, err = filepath.Abs(path)
			if err != nil {
				return nil, err
			}
		} else {
			path = filepath.Join(cwd, path)
		}
		info, err := os.Stat(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if os.IsPathSeparator(location[len(location)-1]) || (info != nil && info.IsDir()) {
			hashDirectory := cwd
			if runtime.GOOS != "windows" {
				// Node's process.cwd() uses the physical path on Unix. os.Getwd
				// can instead preserve a shell's symlink spelling from PWD.
				if err := f.resolveWorkingDirectory(); err != nil {
					return nil, err
				}
				hashDirectory = f.physicalCwd
			}
			path = filepath.Join(path, "rslint-suppressions_"+cwdHash(hashDirectory))
		}
	}
	f.path = path
	data, err := f.read(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 && options.Mode == Check {
		return nil, nil //nolint:nilnil // No active baseline means the CLI can skip collecting linted files.
	}
	if err := f.resolveWorkingDirectory(); err != nil {
		return nil, err
	}
	f.entries = data
	return f, nil
}

func (f *File) resolveWorkingDirectory() error {
	if f.physicalCwd != "" {
		return nil
	}
	physical, err := filepath.EvalSymlinks(f.cwd)
	if err != nil {
		return err
	}
	f.physicalCwd = physical
	for dir := f.cwd; dir != physical; {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		physicalParent, err := filepath.EvalSymlinks(parent)
		if err != nil {
			return err
		}
		// An ordinary component already has the same spelling under its
		// physical parent. Only a link needs its own prefix substitution.
		link := filepath.Join(physicalParent, filepath.Base(dir))
		if link != physical {
			f.cwdAliases = append(f.cwdAliases, utils.DirectoryIdentity{LexicalPath: link, CanonicalPath: physical})
		}
		dir, physical = parent, physicalParent
	}
	slices.Reverse(f.cwdAliases)
	return nil
}

func (f *File) fileKey(path string) string {
	path = filepath.FromSlash(path)
	// Apply shared links from outermost to innermost. An absolute target
	// may already have expanded any of them. Independent links stay lexical.
	for _, dir := range f.cwdAliases {
		if suffix, err := relativeFilePath(dir.LexicalPath, path); err == nil && filepath.IsLocal(suffix) {
			path = filepath.Join(dir.CanonicalPath, suffix)
		}
	}
	key, err := relativeFilePath(f.physicalCwd, path)
	if err != nil {
		// A target on another Windows volume has no relative path.
		key = path
	}
	return filepath.ToSlash(key)
}

func relativeFilePath(dir, path string) (string, error) {
	rel, err := filepath.Rel(dir, path)
	if runtime.GOOS == "windows" && err == nil {
		rel = verifyRelativePathCase(dir, filepath.Clean(path), rel)
	}
	return rel, err
}

// Windows Rel folds every component, even inside case-sensitive directories.
// Verify only spellings that Rel collapsed, including the identity of a link
// itself when two differently named links lead to the same directory.
func verifyRelativePathCase(dir, path, rel string) string {
	if strings.HasPrefix(path, dir) {
		return rel
	}
	for left, right := filepath.Join(dir, rel), path; left != right; {
		leftParent, rightParent := filepath.Dir(left), filepath.Dir(right)
		if leftParent == left || rightParent == right {
			break // filepath.Rel already compared the volume names.
		}
		if filepath.Base(left) != filepath.Base(right) {
			leftInfo, _ := os.Lstat(left)
			rightInfo, _ := os.Lstat(right)
			if !os.SameFile(leftInfo, rightInfo) {
				// Use the outermost distinct entry to preserve its entire subtree.
				up, _ := filepath.Rel(dir, leftParent)
				down, _ := filepath.Rel(rightParent, path)
				rel = filepath.Join(up, down)
			}
		}
		left, right = leftParent, rightParent
	}
	return rel
}

func (f *File) fileMissing(key string) (bool, error) {
	if strings.HasPrefix(key, "//") {
		return false, nil // Do not contact a network share to prune an unchecked file.
	}
	path := filepath.FromSlash(key)
	volume := filepath.VolumeName(path)
	if volume == "" && len(key) >= 3 && key[1:3] == ":/" && tspath.IsVolumeCharacter(key[0]) {
		return false, nil // A Unix host cannot check a Windows drive.
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(f.physicalCwd, path)
	}
	_, err := os.Stat(path)
	if err != nil && volume != "" {
		// A detached drive can report all of its files as missing.
		if root, _ := os.Stat(volume + string(filepath.Separator)); root == nil {
			return false, nil
		}
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		return true, nil
	}
	return false, err
}

func (f *File) read(path string) (entries, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if f.options.Location == "" || f.options.Mode == Record {
			return nil, nil //nolint:nilnil // A missing optional or newly requested baseline has no entries.
		}
		return nil, fmt.Errorf("suppressions file %q does not exist: use --suppress-all or --suppress-rule to create it", path)
	}
	if err != nil {
		return nil, err
	}
	var data entries
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("invalid suppressions file %q: %w", path, err)
	}
	if data == nil {
		return nil, fmt.Errorf("invalid suppressions file %q: expected an object", path)
	}
	for file, rules := range data {
		if file == "" || rules == nil {
			return nil, fmt.Errorf("invalid suppressions for file %q in %q", file, path)
		}
		for id, value := range rules {
			if id == "" || value.Count <= 0 {
				return nil, fmt.Errorf("invalid suppression for %q and rule %q in %q: expected a positive count", file, id, path)
			}
		}
	}
	return data, nil
}

// Stage in the destination directory so a reader always sees complete JSON.
// Exclusive creation serializes the read/modify/replace operation for writers.
func (f *File) update(ctx context.Context, files map[string]*checkedFile) error {
	path := f.path
	physical, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = physical
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	} else if _, linkErr := os.Lstat(path); linkErr == nil {
		return fmt.Errorf("cannot resolve suppressions destination %q: %w", path, err)
	}
	stagePath := path + ".tmp"
	stage, err := os.OpenFile(stagePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("suppressions update already staged at %q: retry after the writer finishes or remove the temporary file if it was interrupted", stagePath)
	}
	if err != nil {
		return fmt.Errorf("stage suppressions update: %w", err)
	}
	// Once renamed the temporary name may immediately belong to another writer.
	removeStage := true
	defer func() {
		if removeStage {
			_ = stage.Close()
			_ = os.Remove(stagePath)
		}
	}()
	data, err := f.read(path)
	if err != nil {
		return err
	}
	if data == nil {
		data = make(entries)
	}
	switch f.options.Mode {
	case Record:
		f.record(data, files)
	case Prune:
		if err := f.prune(data, files); err != nil {
			return err
		}
	}
	if info, err := os.Stat(path); err == nil {
		if err := stage.Chmod(info.Mode().Perm()); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	encoder := json.NewEncoder(stage)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return err
	}
	if err := stage.Close(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Rename(stagePath, path); err != nil {
		return err
	}
	removeStage = false
	f.entries = data
	return nil
}
