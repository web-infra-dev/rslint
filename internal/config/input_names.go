package config

import (
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/gitignore"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// InputPathNames freezes the stored entries of requested files and directories
// before Git and config discovery. Realpath on a shared mount can preserve the
// caller's spelling. Listings are shared within one request.
type InputPathNames struct {
	fs          vfs.FS
	directories map[string]string
	listings    map[inputListing][]string
}

type inputListing struct {
	parent    string
	directory bool
}

func NewInputPathNames(fsys vfs.FS) *InputPathNames {
	return &InputPathNames{fs: fsys, directories: make(map[string]string), listings: make(map[inputListing][]string)}
}

func (names *InputPathNames) storedName(parent, actualParent, base string, directory bool) (name string, hardLinkNameUnknown bool) {
	if names.listings == nil {
		// A live document can read one entry without listing its siblings.
		if native, ok := names.fs.(gitignore.StoredEntryNamer); ok {
			return nativeEntryName(native, parent, base)
		}
	}
	key := inputListing{actualParent, directory}
	entries, ok := names.listings[key]
	if !ok {
		entries = InputEntryNames(names.fs, actualParent, directory)
		if names.listings != nil {
			names.listings[key] = entries
		}
	}
	stored := StoredEntryName(entries, base)
	if stored == base {
		return base, false
	}
	if native, ok := names.fs.(gitignore.StoredEntryNamer); ok {
		// A searchable directory may hide its listing. Query the entry itself
		// without substituting another symlink or hard link to the same file.
		return nativeEntryName(native, parent, base)
	}
	if stored != "" && SamePhysicalPath(names.fs, tspath.CombinePaths(parent, base), tspath.CombinePaths(actualParent, stored)) {
		return stored, false
	}
	return base, false
}

func nativeEntryName(native gitignore.StoredEntryNamer, parent, base string) (string, bool) {
	name, err := native.StoredEntryName(tspath.CombinePaths(parent, base))
	if name != "" {
		return name, false
	}
	return base, errors.Is(err, fs.ErrPermission)
}

// Directory returns the stored spelling of a frozen physical directory.
func (names *InputPathNames) Directory(path string) string {
	if names.fs == nil {
		return path
	}
	if actual, ok := names.directories[path]; ok {
		return actual
	}
	actual := path
	if parent := FilesystemParentDirectory(path); parent != "" {
		actualParent := names.Directory(parent)
		base := tspath.GetBaseFileName(path)
		stored, _ := names.storedName(parent, actualParent, base, true)
		if actualParent != parent || stored != base {
			actual = tspath.CombinePaths(actualParent, stored)
		}
	}
	names.directories[path] = actual
	return actual
}

// File preserves the invocation path and freezes its physical target and Git entry.
func (names *InputPathNames) File(file PathIdentity, enumerated bool) PathIdentity {
	return names.file(file, enumerated, names)
}

// LiveFile shares directory observations but reads the document's leaf entry
// and any symlink target's leaf again, including after an unannounced rename.
func (names *InputPathNames) LiveFile(file PathIdentity, fsys vfs.FS) PathIdentity {
	return names.file(file, false, &InputPathNames{fs: fsys})
}

func (names *InputPathNames) file(file PathIdentity, enumerated bool, leaves *InputPathNames) PathIdentity {
	if names.fs == nil || leaves.fs == nil {
		return file
	}
	parent := file.CanonicalParentPath
	if parent == "" {
		return file
	}
	actualParent := names.Directory(parent)
	if enumerated && actualParent == parent && file.CanonicalPath == file.Path {
		return file
	}
	base := tspath.GetBaseFileName(file.Path)
	actual := base
	if !enumerated {
		var unknown bool
		actual, unknown = leaves.storedName(parent, actualParent, base, false)
		if unknown {
			file.entryNameStatus = entryNameUnlisted
			if native, ok := leaves.fs.(interface {
				ASCIILookupAnchors(path string) (preserved, exactCase bool)
			}); ok {
				if preserved, exact := native.ASCIILookupAnchors(file.Path); exact {
					file.entryNameStatus = entryNameUnlistedExactASCII
				} else if preserved {
					file.entryNameStatus = entryNameUnlistedFoldedASCII
				}
			}
		}
		if actual != base {
			file.EntryName = actual
		}
	}
	if actualParent == parent && actual == base && file.CanonicalPath == file.Path {
		return file
	}
	file.CanonicalParentPath = actualParent
	if file.CanonicalPath == tspath.CombinePaths(parent, base) {
		file.CanonicalPath = tspath.CombinePaths(actualParent, actual)
	} else if file.CanonicalPath != "" {
		// A leaf symlink has a separate physical target and Git entry.
		original := tspath.GetDirectoryPath(file.CanonicalPath)
		targetParent := names.Directory(original)
		targetName, _ := leaves.storedName(original, targetParent, tspath.GetBaseFileName(file.CanonicalPath), false)
		file.CanonicalPath = tspath.CombinePaths(targetParent, targetName)
	}
	return file
}

// KnownDirectory returns an earlier observation without touching the filesystem.
func (names *InputPathNames) KnownDirectory(path string) (string, bool) {
	if names == nil {
		return "", false
	}
	physical, ok := names.directories[path]
	return physical, ok
}

// InputEntryNames lists the files or directories in one parent. A shared mount
// may have cached an empty listing for an alternate spelling. In that case,
// WalkDir reads this one parent afresh.
func InputEntryNames(fsys vfs.FS, parent string, directory bool) []string {
	entries := fsys.GetAccessibleEntries(parent)
	names := entries.Files
	if directory {
		names = entries.Directories
	}
	if len(names) == 0 {
		_ = fsys.WalkDir(parent, func(path string, entry vfs.DirEntry, err error) error {
			if err != nil || path == parent {
				return err
			}
			if entry.IsDir() == directory || entry.Type()&fs.ModeSymlink != 0 {
				names = append(names, entry.Name())
			}
			if entry.IsDir() {
				return vfs.SkipDir
			}
			return nil
		})
	}
	return names
}

// PathSpellingKey groups canonically equivalent spellings for a lookup. The
// filesystem determines which names in the group refer to the same entry.
func PathSpellingKey(path string) string {
	for i := range len(path) {
		if path[i] >= utf8.RuneSelf {
			return cases.Fold().String(norm.NFD.String(path))
		}
	}
	return strings.ToLower(path)
}

// StoredEntryName finds an exact or unambiguous native spelling among the entries.
func StoredEntryName(names []string, requested string) string {
	if _, exact := slices.BinarySearch(names, requested); exact || slices.Contains(names, requested) {
		return requested
	}
	// A case-sensitive volume may preserve both Café and café while treating
	// the composed and decomposed spelling of each as the same entry.
	for _, key := range []func(string) string{norm.NFD.String, PathSpellingKey} {
		want, match := key(requested), ""
		for _, name := range names {
			if key(name) == want {
				if match != "" {
					return ""
				}
				match = name
			}
		}
		if match != "" {
			return match
		}
	}
	return ""
}

// SamePhysicalPath compares the filesystem objects reached by two paths.
func SamePhysicalPath(fsys vfs.FS, left, right string) bool {
	if ExactPathID(left) == ExactPathID(right) {
		return true
	}
	leftInfo, rightInfo := fsys.Stat(left), fsys.Stat(right)
	return leftInfo != nil && rightInfo != nil && os.SameFile(leftInfo, rightInfo)
}
