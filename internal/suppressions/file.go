package suppressions

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/web-infra-dev/rslint/internal/utils"
)

const DefaultFilename = "rslint-suppressions.json"

// Open reads and validates the baseline for cwd without creating or modifying
// files on disk. A check without a default baseline returns (nil, nil).
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
	}
	if err != nil || f.entries == nil && options.Mode == Check {
		return nil, err
	}
	return f, nil
}

// ValidateWritable checks that a not-yet-created baseline's nearest existing
// ancestor directory is writable by creating and removing a temporary probe
// file. Callers can invoke this before mutating lint targets (such as --fix)
// to fail fast on invalid or read-only destinations. Existing baselines are
// left untouched so no-op runs in read-only directories still succeed.
func (f *File) ValidateWritable() error {
	if f.options.Mode == Check || f.entries != nil {
		return nil
	}
	return validateParentDirectory(filepath.Dir(f.path))
}

func (f *File) missing(path string) error {
	if f.options.Location != "" && f.options.Mode != Record {
		return fmt.Errorf("suppressions file %q does not exist: use --suppress-all or --suppress-rule to create it", path)
	}
	return nil
}

func readEntries(path string) (entries, error) {
	content, err := os.ReadFile(path)
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

// update re-reads the latest baseline on disk, applies change, and atomically
// replaces the baseline via a temporary file in the same directory when needed.
// Matching ESLint, concurrent processes writing to the same baseline are not
// synchronized with cross-process locks.
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

	var data entries
	var existingMode fs.FileMode
	exists := false
	info, err := os.Stat(path)
	switch {
	case err == nil:
		if !info.Mode().IsRegular() {
			return fmt.Errorf("suppressions file %q must be a regular file", path)
		}
		data, err = readEntries(path)
		if err != nil {
			return err
		}
		exists = true
		existingMode = info.Mode().Perm()
	case errors.Is(err, fs.ErrNotExist):
		if err := f.missing(path); err != nil {
			return err
		}
		data = make(entries)
	default:
		return err
	}

	changed := change(data)
	if exists && !changed {
		if err := ctx.Err(); err != nil {
			return err
		}
		f.entries = data
		return nil
	}

	if exists {
		// Verify write permission on the existing baseline before replacing it
		// via rename, which otherwise only requires directory write access on Unix.
		probe, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		_ = probe.Close()
	} else if f.options.Mode == Record {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
	if realDir, err := filepath.EvalSymlinks(filepath.Dir(path)); err == nil {
		path = filepath.Join(realDir, filepath.Base(path))
	}

	if err := writeBaselineAtomically(ctx, path, data, exists, existingMode); err != nil {
		return err
	}
	f.entries = data
	return nil
}

func writeBaselineAtomically(ctx context.Context, path string, data entries, preserveMode bool, mode fs.FileMode) error {
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
		if err := stage.Chmod(mode); err != nil {
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
	if err := os.Rename(tempName, path); err != nil {
		return err
	}
	published = true
	return nil
}
