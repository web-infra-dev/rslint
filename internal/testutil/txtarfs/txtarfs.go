// Package txtarfs safely materializes text-only filesystem fixtures stored in
// golang.org/x/tools/txtar archives.
//
// Archives are deliberately limited to portable relative paths and regular
// files. Tests that exercise permissions, symbolic links, or other operating
// system behavior should continue to construct those details imperatively.
package txtarfs

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/tools/txtar"
)

// Archive is a validated, immutable txtar archive.
type Archive struct {
	source string
	files  map[string][]byte
	names  []string
}

// Parse parses data and validates every archived file's portable path
// semantics. source is used only to make errors actionable and may be empty
// for in-memory archives.
func Parse(source string, data []byte) (*Archive, error) {
	parsed := txtar.Parse(data)
	archive := &Archive{
		source: source,
		files:  make(map[string][]byte, len(parsed.Files)),
		names:  make([]string, 0, len(parsed.Files)),
	}
	foldedNames := make(map[string]string, len(parsed.Files))

	for _, file := range parsed.Files {
		if err := validatePath(file.Name); err != nil {
			return nil, fmt.Errorf("%s: invalid archived file %q: %w", archive.label(), file.Name, err)
		}
		if _, exists := archive.files[file.Name]; exists {
			return nil, fmt.Errorf("%s: duplicate archived file %q", archive.label(), file.Name)
		}
		folded := strings.ToLower(file.Name)
		if previous, exists := foldedNames[folded]; exists {
			return nil, fmt.Errorf(
				"%s: archived files %q and %q collide on case-insensitive filesystems",
				archive.label(), previous, file.Name,
			)
		}

		archive.files[file.Name] = append([]byte(nil), file.Data...)
		archive.names = append(archive.names, file.Name)
		foldedNames[folded] = file.Name
	}

	for _, name := range archive.names {
		for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
			if _, exists := archive.files[parent]; exists {
				return nil, fmt.Errorf(
					"%s: archived file %q is also a parent directory of %q",
					archive.label(), parent, name,
				)
			}
		}
	}

	sort.Strings(archive.names)
	return archive, nil
}

// ParseFile reads and parses a txtar archive from fileName.
func ParseFile(fileName string) (*Archive, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read txtar archive %q: %w", fileName, err)
	}
	return Parse(fileName, data)
}

// MustParseFile reads a txtar archive or fails the calling test.
func MustParseFile(t testing.TB, fileName string) *Archive {
	t.Helper()
	archive, err := ParseFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	return archive
}

// ReadFile returns a copy of the named archived file.
func (a *Archive) ReadFile(name string) ([]byte, error) {
	if a == nil {
		return nil, fmt.Errorf("read %q from nil txtar archive", name)
	}
	if err := validatePath(name); err != nil {
		return nil, fmt.Errorf("invalid archived file %q: %w", name, err)
	}
	data, exists := a.files[name]
	if !exists {
		return nil, fmt.Errorf("%s: archived file %q does not exist", a.label(), name)
	}
	return append([]byte(nil), data...), nil
}

// FileNames returns the sorted file names below prefix with the prefix
// stripped. An empty prefix lists the entire archive. A prefix that selects no
// files is an error.
func (a *Archive) FileNames(prefix string) ([]string, error) {
	selected, err := a.selectFiles(prefix)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(selected))
	for index, file := range selected {
		names[index] = file.name
	}
	return names, nil
}

// extract writes the regular files below prefix into root, stripping prefix
// from their names. Existing files are never overwritten. It remains private
// so callers use Materialize and always receive a fresh test-owned directory.
func (a *Archive) extract(root string, prefix string) error {
	selected, err := a.selectFiles(prefix)
	if err != nil {
		return err
	}

	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("stat extraction root %q: %w", root, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("extraction root %q is not a directory", root)
	}

	rootFS, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("open extraction root %q: %w", root, err)
	}
	defer rootFS.Close()

	for _, file := range selected {
		parent := path.Dir(file.name)
		if parent != "." {
			if err := rootFS.MkdirAll(parent, 0o755); err != nil {
				return fmt.Errorf("create directory for archived file %q: %w", file.name, err)
			}
		}
		handle, err := rootFS.OpenFile(file.name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return fmt.Errorf("create archived file %q: %w", file.name, err)
		}
		_, writeErr := handle.Write(file.data)
		closeErr := handle.Close()
		if writeErr != nil {
			return fmt.Errorf("write archived file %q: %w", file.name, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close archived file %q: %w", file.name, closeErr)
		}
	}
	return nil
}

// Materialize extracts prefix into a new test-owned temporary directory.
func (a *Archive) Materialize(t testing.TB, prefix string) string {
	t.Helper()
	directory := t.TempDir()
	if err := a.extract(directory, prefix); err != nil {
		t.Fatal(err)
	}
	return directory
}

func (a *Archive) label() string {
	if a.source == "" {
		return "txtar archive"
	}
	return fmt.Sprintf("txtar archive %q", a.source)
}

type selectedFile struct {
	name string
	data []byte
}

func (a *Archive) selectFiles(prefix string) ([]selectedFile, error) {
	if a == nil {
		return nil, errors.New("select files from nil txtar archive")
	}
	if prefix != "" {
		if err := validatePath(prefix); err != nil {
			return nil, fmt.Errorf("invalid archive prefix %q: %w", prefix, err)
		}
	}

	selected := make([]selectedFile, 0, len(a.names))
	prefixWithSlash := prefix
	if prefixWithSlash != "" {
		prefixWithSlash += "/"
	}
	for _, name := range a.names {
		relativeName := name
		if prefixWithSlash != "" {
			if !strings.HasPrefix(name, prefixWithSlash) {
				continue
			}
			relativeName = strings.TrimPrefix(name, prefixWithSlash)
		}
		selected = append(selected, selectedFile{name: relativeName, data: a.files[name]})
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("%s: archive prefix %q selects no files", a.label(), prefix)
	}
	return selected, nil
}

func validatePath(name string) error {
	if name == "" || name == "." || !fs.ValidPath(name) {
		return errors.New("must be a non-empty, slash-separated relative path")
	}
	if strings.ContainsRune(name, '\\') {
		return errors.New("must use forward slashes")
	}
	if strings.ContainsRune(name, ':') || filepath.VolumeName(name) != "" {
		return errors.New("must not contain a filesystem volume")
	}
	if strings.ContainsAny(name, "<>\"|?*") {
		return errors.New("must not contain characters forbidden in Windows file names")
	}
	for _, character := range name {
		if unicode.IsControl(character) {
			return errors.New("must not contain control characters")
		}
	}
	for _, element := range strings.Split(name, "/") {
		if strings.HasSuffix(element, ".") || strings.HasSuffix(element, " ") {
			return fmt.Errorf("path component %q has a non-portable suffix", element)
		}
		base := strings.ToUpper(strings.SplitN(element, ".", 2)[0])
		if isWindowsDeviceName(base) {
			return fmt.Errorf("path component %q is a reserved Windows device name", element)
		}
	}
	return nil
}

func isWindowsDeviceName(base string) bool {
	if base == "CON" || base == "PRN" || base == "AUX" || base == "NUL" {
		return true
	}
	if len(base) == 4 && (strings.HasPrefix(base, "COM") || strings.HasPrefix(base, "LPT")) {
		return base[3] >= '1' && base[3] <= '9'
	}
	return false
}
