package suppressions

import (
	"cmp"
	"context"
	"encoding/json"
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
)

const DefaultFilename = "rslint-suppressions.json"

// Open reads the file CLI's baseline. A check without a baseline returns nil.
func Open(options Options) (*File, error) {
	f := &File{options: options, path: DefaultFilename}
	var cwd string
	if location := options.Location; location != "" {
		path, err := filepath.Abs(location)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		if os.IsPathSeparator(location[len(location)-1]) || (info != nil && info.IsDir()) {
			// ESLint hashes Node's native process.cwd() for directory locations.
			cwd, err = syscall.Getwd()
			if err != nil {
				return nil, err
			}
			path = filepath.Join(path, "rslint-suppressions_"+cwdHash(cwd))
		}
		f.path = path
	}
	data, err := f.read(f.path)
	if err != nil {
		return nil, err
	}
	if data == nil && options.Mode == Check {
		return nil, nil //nolint:nilnil // No active baseline means the CLI can skip collecting linted files.
	}
	if cwd == "" {
		cwd, err = syscall.Getwd()
		if err != nil {
			return nil, err
		}
	}
	if runtime.GOOS == "windows" {
		// Resolve directory junctions with the same native filesystem used by target selection.
		cwd = filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(cwd)))
	}
	if options.Location == "" {
		f.path = filepath.Join(cwd, DefaultFilename)
	}
	f.cwd = cwd
	f.entries = data
	return f, nil
}

// Keep the linted spelling: different paths to one file can activate different rules.
func (f *File) fileKey(path string) string {
	if runtime.GOOS == "windows" {
		from, to := filepath.VolumeName(f.cwd), filepath.VolumeName(path)
		if !strings.EqualFold(from, to) {
			return path // Files on another drive or share have no relative path.
		}
		// Windows volumes are case-insensitive. Inside a volume, preserve the
		// file and directory spellings that the rule actually received.
		return tspath.GetRelativePathFromDirectory(
			cmp.Or(filepath.ToSlash(f.cwd[len(from):]), "/"), cmp.Or(path[len(to):], "/"),
			tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true},
		)
	}
	relative, err := filepath.Rel(f.cwd, path)
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
	// A missing child under a dangling directory link or a detached drive
	// may become reachable again. Find the nearest directory entry that exists.
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

// The exclusive staging file is also the writer lock. Read the latest counts
// after creating it, then replace the destination with the completed JSON.
func (f *File) update(ctx context.Context, files map[string]*checkedFile) error {
	path, err := filepath.EvalSymlinks(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		path = f.path // A new baseline has no symlink target to resolve.
	} else if err != nil {
		return err
	}
	stagePath := path + ".tmp"
	stage, err := os.OpenFile(stagePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("stage suppressions: %w (retry after the writer finishes or remove the temporary file if it was interrupted)", err)
	}
	if err != nil {
		return err
	}
	published := false
	defer func() {
		if !published { // After a successful rename the name can belong to a new writer.
			_ = stage.Close()
			_ = os.Remove(stagePath)
		}
	}()

	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if info != nil {
		// A dangling symlink also reports ErrNotExist during resolution.
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("cannot resolve suppressions destination %q", path)
		}
		// A rename can replace a read-only file on Unix. Recording and pruning
		// require write access even when the counts are unchanged.
		existing, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		if err := existing.Close(); err != nil {
			return err
		}
	}
	data, err := f.read(path)
	if err != nil {
		return err
	}
	newFile := data == nil
	if newFile {
		data = make(entries)
	}
	var changed bool
	if f.options.Mode == Record {
		changed = f.record(data, files)
	} else {
		changed = f.prune(data, files)
	}
	if !newFile && !changed {
		f.entries = data
		return ctx.Err()
	}
	if info != nil {
		if err := stage.Chmod(info.Mode().Perm()); err != nil {
			return err
		}
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
	published = true
	f.entries = data
	return nil
}
