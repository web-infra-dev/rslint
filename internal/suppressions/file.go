package suppressions

import (
	"bytes"
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

// Open reads the baseline for cwd. A check without a default baseline returns nil.
func Open(cwd string, options Options) (*File, error) {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	cwd = filepath.Clean(filepath.FromSlash(cwd))
	physicalCwd := resolvePhysicalDirectory(cwd)

	f := &File{
		cwd:        physicalCwd,
		logicalCwd: cwd,
		path:       filepath.Join(cwd, DefaultFilename),
		options:    options,
	}
	location := options.Location
	if location != "" {
		f.path = resolveLocationPath(cwd, location)
	}
	info, err := os.Stat(f.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if info != nil && info.IsDir() || location != "" && os.IsPathSeparator(location[len(location)-1]) {
		// Match ESLint's directory-based filename hash against Node's process.cwd().
		hashCwd := physicalCwd
		if runtime.GOOS == "windows" {
			hashCwd = cwd
		}
		f.path = filepath.Join(f.path, "rslint-suppressions_"+cwdHash(hashCwd))
		info, err = os.Stat(f.path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	if err == nil {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("suppressions file %q must be a regular file", f.path)
		}
		f.entries, err = readEntries(f.path)
	}
	if errors.Is(err, fs.ErrNotExist) {
		if options.Mode != Check {
			if link, lstatErr := os.Lstat(f.path); lstatErr == nil && link.Mode()&fs.ModeSymlink != 0 {
				return nil, fmt.Errorf("cannot resolve suppressions destination %q", f.path)
			}
		}
		if err = f.missing(f.path); err != nil {
			return nil, err
		}
		if options.Mode != Check {
			if parentErr := validateParentDirectory(filepath.Dir(f.path)); parentErr != nil {
				return nil, parentErr
			}
		}
	}
	if err != nil || f.entries == nil && options.Mode == Check {
		return nil, err
	}
	return f, nil
}

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

func (f *File) missing(path string) error {
	if f.options.Location != "" && f.options.Mode != Record {
		return fmt.Errorf("suppressions file %q does not exist: use --suppress-all or --suppress-rule to create it", path)
	}
	return nil
}

func readEntries(path string) (entries, error) {
	content, err := readBaselineFile(path)
	if err != nil {
		return nil, err
	}
	return decodeEntries(content, path)
}

func readEntriesFromFile(file *os.File, path string) (entries, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return decodeEntries(content, path)
}

func decodeEntries(content []byte, path string) (entries, error) {
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
				if prior, ok := existing[id]; ok {
					if prior.Count != value.Count {
						return fmt.Errorf("conflicting counts for rule %q in %q and %q", id, file, key)
					}
					if len(value.Extra) == 0 {
						continue
					}
					for name, field := range prior.Extra {
						if other, ok := value.Extra[name]; ok && !bytes.Equal(field, other) {
							return fmt.Errorf("conflicting field %q for rule %q in %q and %q", name, id, file, key)
						}
						value.Extra[name] = field
					}
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

// update checks whether the baseline needs modification, then serializes
// writes on the baseline file itself, re-reads the latest contents, applies
// change, and replaces the baseline atomically via a temporary file.
func (f *File) update(ctx context.Context, change func(entries) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := f.path
	if resolved, err := filepath.EvalSymlinks(f.path); err == nil {
		path = resolved
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	} else if info, lstatErr := os.Lstat(f.path); lstatErr == nil && info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("cannot resolve suppressions destination %q", f.path)
	}

	// Avoid taking a write lock when an existing baseline already matches the
	// desired counts, so no-op runs succeed even inside a read-only directory.
	if info, err := os.Stat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("suppressions file %q must be a regular file", path)
		}
		snapshot, readErr := readEntries(path)
		if readErr != nil {
			if !errors.Is(readErr, fs.ErrNotExist) {
				return readErr
			}
		} else if !change(snapshot) {
			if err := ctx.Err(); err != nil {
				return err
			}
			f.entries = snapshot
			return nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if f.options.Mode == Record {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	if realDir, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
		path = filepath.Join(realDir, filepath.Base(path))
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		locked, err := lockExistingBaseline(ctx, path)
		switch {
		case err == nil:
			data, readErr := readEntriesFromFile(locked.file, path)
			if readErr != nil {
				locked.unlock()
				return readErr
			}
			if change(data) {
				if writeErr := writeAndPublishBaseline(ctx, path, data, locked); writeErr != nil {
					locked.unlock()
					return writeErr
				}
			}
			locked.unlock()
			if err := ctx.Err(); err != nil {
				return err
			}
			f.entries = data
			return nil
		case errors.Is(err, fs.ErrNotExist):
			if link, lstatErr := os.Lstat(path); lstatErr == nil && link.Mode()&fs.ModeSymlink != 0 {
				return fmt.Errorf("cannot resolve suppressions destination %q", path)
			}
			if err := f.missing(path); err != nil {
				return err
			}
			data := make(entries)
			change(data)
			err = writeAndPublishBaseline(ctx, path, data, nil)
			if errors.Is(err, fs.ErrExist) {
				// Another writer created the baseline first; lock and merge into it.
				continue
			}
			if err != nil {
				return err
			}
			f.entries = data
			return nil
		default:
			return err
		}
	}
}

func writeAndPublishBaseline(ctx context.Context, path string, data entries, locked *lockedBaseline) error {
	preserveMode := locked != nil
	if !preserveMode {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	tempName := filepath.Join(filepath.Dir(path), ".rslint-suppressions-"+rand.Text()+".tmp")
	perm := fs.FileMode(0o666)
	if preserveMode {
		perm = 0o600
	}
	stage, err := os.OpenFile(tempName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	published := false
	defer func() {
		if !published {
			_ = stage.Close()
			_ = os.Remove(tempName)
		}
	}()
	if preserveMode {
		if err := stage.Chmod(locked.mode); err != nil {
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
	if preserveMode {
		if err := replaceLockedBaseline(ctx, tempName, path, locked); err != nil {
			return err
		}
	} else {
		if err := publishNewBaseline(ctx, tempName, path); err != nil {
			return err
		}
	}
	published = true
	return nil
}
