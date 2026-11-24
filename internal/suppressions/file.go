package suppressions

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
			return utils.NormalizeAbsoluteDrive(path) // Another drive or share has no relative path.
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

func (f *File) missing(path string) error {
	if f.options.Location != "" && f.options.Mode != Record {
		return fmt.Errorf("suppressions file %q does not exist: use --suppress-all or --suppress-rule to create it", path)
	}
	return nil
}

func (f *File) read(path string) (entries, error) {
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, f.missing(path)
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return decodeEntries(file, path)
}

func decodeEntries(file *os.File, path string) (entries, error) {
	content, err := io.ReadAll(file)
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
	if runtime.GOOS == "windows" {
		if err := normalizeDriveKeys(data); err != nil {
			return nil, fmt.Errorf("invalid suppressions file %q: %w", path, err)
		}
	}
	return data, nil
}

func normalizeDriveKeys(data entries) error {
	for file, rules := range data {
		key := utils.NormalizeAbsoluteDrive(file)
		if key == file {
			continue
		}
		if existing := data[key]; existing != nil {
			for id, value := range rules {
				if prior, ok := existing[id]; ok && prior != value {
					return fmt.Errorf("conflicting counts for rule %q in %q and %q", id, file, key)
				}
				existing[id] = value
			}
		} else {
			data[key] = rules
		}
		delete(data, file)
	}
	return nil
}

// Reread and lock an existing baseline, or publish a new one without replacement.
func (f *File) update(ctx context.Context, change func(entries) bool) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := filepath.EvalSymlinks(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		path = f.path
	} else if err != nil {
		return err
	}
	defer func() {
		if err = nativeUpdateError(err); errors.Is(err, fs.ErrExist) {
			err = fmt.Errorf("cannot update suppressions %q: %w (if another writer is active, retry after it finishes)", path, err)
		}
	}()
	var info fs.FileInfo
	var data entries
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	switch {
	case err == nil:
		guard, err := lockBaseline(file)
		if err != nil {
			_ = file.Close()
			return err
		}
		defer guard.Close()
		data, info, err = readLockedBaseline(file, path)
		if guard != file {
			// Windows keeps a separate guard so the data handle can close before Rename.
			_ = file.Close()
		}
		if err != nil {
			return err
		}
	case errors.Is(err, fs.ErrNotExist):
		if err := f.missing(path); err != nil {
			return err
		}
		data = make(entries)
	default:
		return err
	}
	if changed := change(data); changed || info == nil {
		if err := writeBaseline(ctx, path, info, data); err != nil {
			return err
		}
	} else if err := ctx.Err(); err != nil {
		return err
	}
	f.entries = data
	return nil
}

func readLockedBaseline(file *os.File, path string) (entries, fs.FileInfo, error) {
	held, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	// A previous writer may have replaced the open file before we acquired its lock.
	named, err := os.Lstat(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, err
	}
	if !os.SameFile(held, named) {
		return nil, nil, fs.ErrExist
	}
	data, err := decodeEntries(file, path)
	return data, held, err
}

func writeBaseline(ctx context.Context, path string, info fs.FileInfo, data entries) error {
	// O_EXCL and the creation mode preserve the caller's umask for a new baseline.
	name := filepath.Join(filepath.Dir(path), ".rslint-suppressions-"+rand.Text()+".tmp")
	stage, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	published := false
	defer func() {
		if !published {
			_ = stage.Close()
			_ = os.Remove(name)
		}
	}()
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
	if info == nil {
		err = renameNew(name, path)
		if errors.Is(err, fs.ErrExist) {
			if link, statErr := os.Lstat(path); statErr == nil && link.Mode()&fs.ModeSymlink != 0 {
				return fmt.Errorf("cannot resolve suppressions destination %q", path)
			}
		}
	} else {
		err = os.Rename(name, path)
	}
	if err != nil {
		return err
	}
	published = true
	return nil
}
