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
)

const DefaultFilename = "rslint-suppressions.json"

// Open reads the file CLI's baseline. A check without a baseline returns nil.
func Open(options Options) (*File, error) {
	f := &File{options: options, path: DefaultFilename}
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
		f.path = filepath.Join(cwd, DefaultFilename)
	}
	if runtime.GOOS == "windows" {
		// Windows Getwd can retain the name of a junction.
		cwd, err = filepath.EvalSymlinks(cwd)
		if err != nil {
			return nil, err
		}
	}
	f.cwd = cwd
	f.entries = data
	return f, nil
}

// Keep the linted spelling: different paths to one file can activate different rules.
func (f *File) fileKey(path string) string {
	relative, err := filepath.Rel(f.cwd, path)
	if err != nil {
		return path // A file on another Windows drive or share has no relative path.
	}
	if runtime.GOOS == "windows" {
		return windowsRelativeKey(f.cwd, path, relative)
	}
	return filepath.ToSlash(relative)
}

// Rel folds every Windows directory name, but a case-sensitive directory can contain
// both pkg and Pkg. Verify only spellings that Rel actually folded.
func windowsRelativeKey(cwd, target, relative string) string {
	cwd, target = filepath.Clean(cwd), filepath.Clean(target)
	resolved := filepath.Join(cwd, relative)
	baseVolume, targetVolume := filepath.VolumeName(cwd), filepath.VolumeName(target)
	if resolved[len(baseVolume):] == target[len(targetVolume):] {
		return filepath.ToSlash(relative)
	}
	sep := string(filepath.Separator)
	from := strings.Split(strings.TrimPrefix(cwd[len(baseVolume):], sep), sep)
	to := strings.Split(strings.TrimPrefix(target[len(targetVolume):], sep), sep)
	for i := range min(len(from), len(to)) {
		if from[i] == to[i] {
			continue
		}
		if !strings.EqualFold(from[i], to[i]) {
			break
		}
		// Compare the entries themselves to distinguish junctions to the same target.
		left, leftErr := os.Lstat(baseVolume + sep + strings.Join(from[:i+1], sep))
		right, rightErr := os.Lstat(targetVolume + sep + strings.Join(to[:i+1], sep))
		if leftErr != nil || rightErr != nil || !os.SameFile(left, right) {
			return strings.Repeat("../", len(from)-i) + strings.Join(to[i:], "/")
		}
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
	// A dangling symlink also reports ErrNotExist during resolution.
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
