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
	"strings"
	"syscall"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

const DefaultFilename = "rslint-suppressions.json"

// Open reads the CLI baseline. workingDirectory is the spelling of the process
// working directory used to select lint targets. A check without a baseline returns nil.
func Open(workingDirectory string, options Options) (*File, error) {
	selected := filepath.ToSlash(workingDirectory)
	f := &File{options: options, path: filepath.Join(filepath.FromSlash(selected), DefaultFilename)}
	var cwd string
	var err error
	if location := options.Location; location != "" {
		// Match Node's process.cwd(), including the stored case on Unix and
		// the native drive and junction spelling on Windows.
		cwd, err = syscall.Getwd()
		if err != nil {
			return nil, err
		}
		path := location
		if filepath.VolumeName(path) != "" || os.IsPathSeparator(path[0]) {
			// Let Windows resolve drive-relative and current-drive-rooted paths.
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
			path = filepath.Join(path, "rslint-suppressions_"+cwdHash(cwd))
		}
		f.path = path
	}
	data, err := f.read(f.path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 && options.Mode == Check {
		return nil, nil //nolint:nilnil // No active baseline means the CLI can skip collecting linted files.
	}
	if cwd == "" {
		cwd, err = syscall.Getwd()
		if err != nil {
			return nil, err
		}
	}
	physical := filepath.ToSlash(cwd)
	if runtime.GOOS == "windows" {
		physical = osvfs.FS().Realpath(physical)
	}
	f.cwd = tspath.EnsureTrailingDirectorySeparator(physical)
	if selected != physical {
		f.cwdAlias = selected
	}
	f.entries = data
	return f, nil
}

// The working directory itself may have a caller-visible symlink spelling.
// Keep the suffix as lint reported it, including symlinks below that directory.
func (f *File) fileKey(path string) string {
	if f.cwdAlias != "" {
		if relative, err := filepath.Rel(f.cwdAlias, path); err == nil && filepath.IsLocal(relative) {
			return filepath.ToSlash(relative)
		}
	}
	if relative, ok := strings.CutPrefix(path, f.cwd); ok {
		return relative
	}
	relative, err := filepath.Rel(f.cwd, path)
	if err != nil {
		return path // A file on another Windows drive or share has no relative path.
	}
	return filepath.ToSlash(relative)
}

func (f *File) fileMissing(key string) bool {
	if strings.HasPrefix(key, "//") {
		return false // Do not contact a network share to prune an unchecked file.
	}
	path := filepath.FromSlash(key)
	volume := filepath.VolumeName(path)
	if volume == "" && len(key) >= 3 && key[1:3] == ":/" && tspath.IsVolumeCharacter(key[0]) {
		return false // A Unix host cannot check a Windows drive.
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(f.cwd, path)
	}
	_, err := os.Stat(path)
	if !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ENOTDIR) {
		return false
	}
	if volume != "" {
		// A detached drive can report all of its files as missing.
		_, err := os.Stat(volume + string(filepath.Separator))
		return err == nil
	}
	return true
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
	path := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(f.path)))
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
	// Realpath returns its input on failure. Never replace a dangling symlink.
	if info != nil && info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("cannot resolve suppressions destination %q", path)
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
