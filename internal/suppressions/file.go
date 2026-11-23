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
	"github.com/web-infra-dev/rslint/internal/utils"
)

const DefaultFilename = "rslint-suppressions.json"

// Open reads the CLI baseline from the native process working directory.
// workingDirectory is the spelling used to select lint targets. A plain
// check without a baseline returns nil.
func Open(options Options, workingDirectory string, storedDirectory func(string) string) (*File, error) {
	cwd, err := syscall.Getwd()
	if err != nil {
		return nil, err
	}
	f := &File{options: options}
	path := filepath.Join(cwd, DefaultFilename)
	if location := options.Location; location != "" {
		path = location
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
	}
	f.path = path
	data, err := f.read(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 && options.Mode == Check {
		return nil, nil //nolint:nilnil // No active baseline means the CLI can skip collecting linted files.
	}
	cwd = filepath.ToSlash(cwd)
	if runtime.GOOS == "windows" {
		// GetCurrentDirectory can retain a junction in the working directory.
		cwd = osvfs.FS().Realpath(cwd)
	}
	// A shared mount can retain the spelling used to enter the directory.
	// File keys use its stored name even when two invocations enter it differently.
	if storedDirectory != nil {
		cwd = storedDirectory(cwd)
	}
	f.cwd = tspath.EnsureTrailingDirectorySeparator(cwd)
	if selected := filepath.ToSlash(workingDirectory); selected != cwd {
		f.cwdAlias = tspath.EnsureTrailingDirectorySeparator(utils.NormalizeAbsoluteDrive(selected))
	}
	f.entries = data
	return f, nil
}

// Lint paths under cwd already contain their slash-separated JSON key.
func (f *File) fileKey(path string) string {
	// A target under the cwd used for selection has the same relative name
	// when the caller enters through a symlink. Keep links below that cwd.
	if f.cwdAlias != "" {
		if relative, ok := strings.CutPrefix(utils.NormalizeAbsoluteDrive(path), f.cwdAlias); ok {
			return relative
		}
	}
	dir := f.cwd
	if relative, ok := strings.CutPrefix(path, dir); ok {
		return relative
	}
	if runtime.GOOS == "windows" {
		baseVolume, fileVolume := filepath.VolumeName(dir), filepath.VolumeName(path)
		// Target identity preserves UNC server and share spelling.
		if strings.HasPrefix(path, "//") && baseVolume != fileVolume {
			return path
		}
		if !strings.EqualFold(baseVolume, fileVolume) {
			return strings.ToUpper(fileVolume) + path[len(fileVolume):]
		}
		return tspath.GetRelativePathFromDirectory(dir, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
	}
	relative, _ := filepath.Rel(dir, path) // Both absolute paths have the same Unix root.
	return relative
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
		path = filepath.Join(f.cwd, path)
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

func resolveUpdatePath(path string) (string, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		// EvalSymlinks treats an intermediate junction as a non-directory.
		// The native resolver follows junctions and returns its input on failure.
		absolute, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		physical := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(absolute)))
		if physical != absolute || info.Mode()&fs.ModeSymlink == 0 {
			return physical, nil
		}
	}
	// Preserve a final symlink, including a dangling one, when replacing the file.
	physical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve suppressions destination %q: %w", path, err)
	}
	return physical, nil
}

type stagedUpdate struct {
	destination string
	file        *os.File
	entries     entries
	publish     bool
}

// BeginUpdate reserves the baseline and validates its latest counts.
func (f *File) BeginUpdate() error {
	if f.pending != nil {
		return nil
	}
	path, err := resolveUpdatePath(f.path)
	if err != nil {
		return err
	}
	stagePath := path + ".tmp"
	stage, err := os.OpenFile(stagePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("suppressions update already staged at %q: retry after the writer finishes or remove the temporary file if it was interrupted", stagePath)
	}
	if err != nil {
		return fmt.Errorf("stage suppressions update: %w", err)
	}
	f.pending = &stagedUpdate{destination: path, file: stage}
	data, err := f.read(path)
	if err == nil {
		err = f.checkTargetAliases(data)
	}
	if err != nil {
		f.DiscardUpdate()
		return err
	}
	f.pending.publish = data == nil
	if data == nil {
		data = make(entries)
	}
	f.pending.entries = data
	return nil
}

// DiscardUpdate releases an unpublished update, including its writer lock.
func (f *File) DiscardUpdate() {
	if pending := f.pending; pending != nil {
		f.pending = nil
		_ = pending.file.Close()
		_ = os.Remove(pending.file.Name())
	}
}

func (f *File) prepareUpdate(ctx context.Context, files map[string]*checkedFile) error {
	if err := f.BeginUpdate(); err != nil {
		return err
	}
	pending := f.pending
	var changed bool
	var err error
	if f.options.Mode == Record {
		changed, err = f.record(pending.entries, files)
	} else {
		changed, err = f.prune(pending.entries, files)
	}
	if err != nil {
		return err
	}
	pending.publish = pending.publish || changed
	if pending.publish {
		if err := pending.write(); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// CommitUpdate publishes the prepared counts and releases their writer lock.
// A check has no pending update.
func (f *File) CommitUpdate() error {
	pending := f.pending
	if pending == nil {
		return nil
	}
	if pending.publish {
		if err := os.Rename(pending.file.Name(), pending.destination); err != nil {
			return err
		}
		f.pending = nil // After rename, a new writer may already own the temporary name.
	} else {
		f.DiscardUpdate()
	}
	f.entries = pending.entries
	return nil
}

func (pending *stagedUpdate) write() error {
	if info, err := os.Stat(pending.destination); err == nil {
		if err := pending.file.Chmod(info.Mode().Perm()); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	encoder := json.NewEncoder(pending.file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(pending.entries); err != nil {
		return err
	}
	return pending.file.Close()
}
