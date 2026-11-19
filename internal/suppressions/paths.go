package suppressions

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (f *File) resolveWorkingDirectory(cwd string) error {
	if f.baseDir == "" {
		var err error
		f.baseDir, err = syscall.Getwd()
		if err != nil {
			return err
		}
	}
	physical := f.baseDir
	if runtime.GOOS == "windows" {
		// The native resolver follows junctions as well as directory symlinks.
		physical = filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(f.baseDir)))
	}
	f.physicalCwd = physical
	if cwd == physical {
		return nil
	}

	// Only links traversed to reach the working directory may be substituted
	// in a lint target. Remember intermediate links in their targets as well.
	links := make(map[string]string)
	var collect func(string, string) error
	collect = func(parent, path string) error {
		path = filepath.FromSlash(path)
		volume := filepath.VolumeName(path)
		path = path[len(volume):]
		if volume != "" || len(path) > 0 && os.IsPathSeparator(path[0]) {
			if volume == "" {
				volume = filepath.VolumeName(parent)
			}
			parent = volume + string(filepath.Separator)
		}
		for name := range strings.SplitSeq(path, string(filepath.Separator)) {
			if name == "" || name == "." {
				continue
			}
			part := filepath.Join(parent, name)
			if target, ok := links[part]; ok {
				parent = target
				continue
			}
			info, err := os.Lstat(part)
			if err != nil {
				return err
			}
			if info.IsDir() {
				parent = part
				continue
			}
			destination, err := os.Readlink(part)
			if err != nil {
				return err
			}
			target := filepath.FromSlash(osvfs.FS().Realpath(filepath.ToSlash(part)))
			if target == part { // Realpath returns its input if the link no longer resolves.
				return fmt.Errorf("cannot resolve working directory link %q", part)
			}
			links[part] = target
			if err := collect(parent, destination); err != nil {
				return err
			}
			f.cwdAliases = append(f.cwdAliases, utils.DirectoryIdentity{LexicalPath: part, CanonicalPath: target})
			parent = target
		}
		return nil
	}
	return collect("", cwd)
}

func (f *File) projectPath(path string) string {
project:
	for {
		for _, dir := range f.cwdAliases {
			suffix, err := relativeFilePath(dir.LexicalPath, path)
			if err == nil && filepath.IsLocal(suffix) {
				path = filepath.Join(dir.CanonicalPath, suffix)
				continue project
			}
		}
		return path
	}
}

func (f *File) fileKey(path string) string {
	path = filepath.FromSlash(path)
	target := f.projectPath(path)
	key, err := relativeFilePath(f.physicalCwd, target)
	// A Windows working directory may itself be a junction. Moving up from
	// its name can land in a different directory than moving up from its target.
	if f.baseDir != f.physicalCwd && !filepath.IsLocal(key) {
		fromBase := f.projectPath(filepath.Join(f.baseDir, key))
		matched, _ := relativeFilePath(fromBase, target)
		if err != nil || matched != "." {
			key, err = relativeFilePath(f.baseDir, path)
		}
	}
	if err != nil {
		key = path // A target on another Windows volume has no relative path.
	}
	return filepath.ToSlash(key)
}

// Volumes may equate case or Unicode spellings. Lstat compares the directory
// entries themselves, so independently named links keep their own paths.
func relativeFilePath(dir, path string) (string, error) {
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(path, dir) {
		return rel, err
	}
	sep := string(filepath.Separator)
	left := strings.TrimPrefix(dir[len(filepath.VolumeName(dir)):], sep)
	right := strings.TrimPrefix(path[len(filepath.VolumeName(path)):], sep)
	nonASCII := func(r rune) bool { return r > unicode.MaxASCII }
	checkedSpelling := false
	for left != "" && right != "" {
		leftName, nextLeft, _ := strings.Cut(left, sep)
		rightName, nextRight, _ := strings.Cut(right, sep)
		if leftName != rightName {
			if !strings.EqualFold(leftName, rightName) && !strings.ContainsFunc(leftName, nonASCII) && !strings.ContainsFunc(rightName, nonASCII) {
				break
			}
			checkedSpelling = true
			leftInfo, _ := os.Lstat(strings.TrimSuffix(dir, left) + leftName)
			rightInfo, _ := os.Lstat(strings.TrimSuffix(path, right) + rightName)
			if !os.SameFile(leftInfo, rightInfo) {
				break
			}
		}
		left, right = nextLeft, nextRight
	}
	if checkedSpelling {
		up, _ := filepath.Rel(dir, strings.TrimSuffix(dir, left))
		rel = filepath.Join(up, right)
	}
	return rel, nil
}

// ESLint keys are relative to the native process cwd and retain the spelling
// of the linted input.
func (f *File) operandKey(path string) string {
	path = filepath.FromSlash(path)
	rel := filepath.Rel
	if runtime.GOOS == "windows" {
		rel = relativeFilePath // A case-sensitive directory can contain both spellings.
	}
	key, err := rel(f.baseDir, path)
	if err != nil {
		key = path
	}
	return filepath.ToSlash(key)
}

func (f *File) matchFiles(selected []FileResult) map[string]*checkedFile {
	files := make(map[string]*checkedFile)
	byPortableKey := make(map[string][]*checkedFile)
	exactMatches := make(map[string]bool)
	if f.options.Mode == Check {
		// An exact entry can belong to a group whose portable key is absent.
		// Retain that group, including peers with no baseline entry of their own.
		hasEntry := func(path string) bool { return f.entries[f.operandKey(path)] != nil }
		for _, result := range selected {
			if hasEntry(result.Path) || slices.ContainsFunc(result.InputPaths, hasEntry) {
				if portable := f.fileKey(result.Path); f.entries[portable] == nil {
					exactMatches[portable] = true
				}
			}
		}
	}
	for _, result := range selected {
		portable := f.fileKey(result.Path)
		if f.options.Mode == Check && f.entries[portable] == nil && !exactMatches[portable] {
			continue
		}
		literal := f.operandKey(result.Path)
		keys := make([]string, 0, max(1, len(result.InputPaths)))
		for _, input := range result.InputPaths {
			key := literal
			// From a symlinked cwd, .. can name a different physical file.
			if input != result.Path && f.fileKey(input) == portable {
				key = f.operandKey(input)
			}
			keys = append(keys, key)
		}
		if len(keys) == 0 {
			keys = append(keys, literal)
		}
		slices.Sort(keys)
		file := &checkedFile{
			keys:         slices.Compact(keys),
			literal:      literal,
			portable:     portable,
			skippedRules: result.SkippedRules,
		}
		byPortableKey[portable] = append(byPortableKey[portable], file)
		if !result.Unchecked {
			files[result.Path] = file
		}
	}
	for _, group := range byPortableKey {
		if len(group) > 1 {
			separateFilenames(group)
		}
	}
	return files
}

// Rules can inspect the filename, so separately selected paths that share a
// portable key cannot use each other's counts. Keep a disputed operand on the
// name passed to the rule. Unchecked aliases reserve names by the same logic.
func separateFilenames(group []*checkedFile) {
	owners := make(map[string]string)
	claim := func(key, literal string) {
		if owner, ok := owners[key]; ok && owner != literal {
			owners[key] = ""
		} else {
			owners[key] = literal
		}
	}
	for _, file := range group {
		claim(file.literal, file.literal)
		for _, key := range file.keys {
			claim(key, file.literal)
		}
	}
	for _, file := range group {
		file.portable = ""
		for i, key := range file.keys {
			if owners[key] != file.literal {
				file.keys[i] = file.literal
			}
		}
		slices.Sort(file.keys)
		file.keys = slices.Compact(file.keys)
	}
}

func selectExistingKeys(data entries, files map[string]*checkedFile) {
	for _, file := range files {
		// A missing key has no allowance when several spellings were requested.
		if len(file.keys) == 1 && file.portable != "" && data[file.keys[0]] == nil {
			file.keys[0] = file.portable
		}
	}
}
