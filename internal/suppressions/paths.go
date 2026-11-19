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

// SamePathSpelling accepts native case and Unicode variants of the same
// sequence of directory entries. Two separately named symlinks stay distinct.
func SamePathSpelling(left, right string) bool {
	if left == right {
		return true
	}
	rel, err := relativeFilePath(filepath.FromSlash(left), filepath.FromSlash(right))
	return err == nil && rel == "."
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

// Compare after substituting only the links traversed to reach cwd. A link in
// the selected path keeps its own name, even if another link has the same target.
func (f *File) sameProjectedPath(left, right string) bool {
	return SamePathSpelling(f.projectPath(filepath.FromSlash(left)), f.projectPath(filepath.FromSlash(right)))
}

func (f *File) matchFiles(selected []FileResult) map[string]*checkedFile {
	files := make(map[string]*checkedFile)
	groups := make(map[string][]*checkedFile)
	// With operand aliases, even a filename absent from the baseline can own
	// an exact key that another result would otherwise use.
	hasOrigins := slices.ContainsFunc(selected, func(result FileResult) bool {
		return len(result.InputPaths) != 0 || result.CoalescedInto != ""
	})
	resolveNames := hasOrigins
	for _, result := range selected {
		portable, literal := f.fileKey(result.Path), f.operandKey(result.Path)
		if f.options.Mode == Check && !hasOrigins && f.entries[portable] == nil && f.entries[literal] == nil {
			continue
		}
		keys := make([]string, 0, max(1, len(result.InputPaths)))
		for _, input := range result.InputPaths {
			key := literal
			// From a symlinked cwd, .. can name a different physical file.
			if input != result.Path && f.sameProjectedPath(input, result.Path) {
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
			unchecked:    result.Unchecked,
		}
		group := portable
		if result.CoalescedInto != "" && f.sameProjectedPath(result.Path, result.CoalescedInto) {
			group = f.fileKey(result.CoalescedInto)
		}
		resolveNames = resolveNames || literal != portable || len(groups[group]) != 0
		groups[group] = append(groups[group], file)
		if !result.Unchecked {
			files[result.Path] = file
		}
	}
	if resolveNames {
		separateFilenames(groups)
	}
	return files
}

// A single execution requested through a single exact key can also use its
// portable entry. Repeating the same key through a skipped alias changes nothing.
func hasSingleInput(group []*checkedFile) bool {
	key := group[0].keys[0]
	checked := 0
	for _, file := range group {
		if len(file.keys) != 1 || file.keys[0] != key {
			return false
		}
		if !file.unchecked {
			checked++
		}
	}
	return checked == 1
}

// Exact keys share one JSON namespace across all files. Independent executions
// may need distinct names even when their portable paths differ in spelling.
func separateFilenames(groups map[string][]*checkedFile) {
	observed := make(map[string]*checkedFile)
	reserved := make(map[string]bool)
	claim := func(key string, file *checkedFile) {
		if owner, ok := observed[key]; ok && owner != file {
			observed[key] = nil // Multiple executions can have different counts.
		} else {
			observed[key] = file
		}
	}
	for _, group := range groups {
		for _, file := range group {
			if file.unchecked {
				if slices.Contains(file.keys, file.literal) {
					reserved[file.literal] = true
				}
				continue
			}
			claim(file.literal, file)
			for _, key := range file.keys {
				claim(key, file)
			}
		}
	}
	for _, group := range groups {
		single := hasSingleInput(group)
		for _, file := range group {
			if file.unchecked {
				continue
			}
			owner, claimed := observed[file.portable]
			if !single || claimed && owner != file || reserved[file.portable] && file.keys[0] != file.portable {
				file.portable = ""
			}
			hasOwnName := slices.Contains(file.keys, file.literal)
			keys := file.keys[:0]
			for _, key := range file.keys {
				switch {
				case observed[key] != file:
					key = file.literal // Give each actual lint result its own key.
				case hasOwnName && key != file.literal && reserved[key]:
					continue // This separately requested filename was never linted.
				}
				keys = append(keys, key)
			}
			slices.Sort(keys)
			file.keys = slices.Compact(keys)
		}
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
