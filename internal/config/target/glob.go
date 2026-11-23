package target

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

var inputPatternOptions = minimatch3.Options{Dot: true, NoComment: true, PreserveWhitespace: true}

// literalPatternPath decodes pattern syntax outside the native volume. Windows
// drive and UNC namespaces are fixed filesystem roots, not matching components.
func literalPatternPath(pattern string) (string, bool) {
	volume := filepath.ToSlash(filepath.VolumeName(pattern))
	if volume != "" && len(volume) == len(pattern) {
		return volume, true
	}
	options := inputPatternOptions
	options.NoNegate = volume != ""
	literal, ok := minimatch3.New(pattern[len(volume):], options).Literal()
	return volume + literal, ok
}

// globDirectory returns the decoded static directory, its authored prefix, and
// the untouched suffix. The caller normalizes dot segments in authored syntax
// after fixing the scan entry. Native volume boundaries keep the search finite.
func globDirectory(pattern string) (directory, prefix, relative string) {
	volumeLength := len(filepath.VolumeName(pattern))
	end := len(pattern)
	for {
		index := strings.LastIndexByte(pattern[volumeLength:end], '/')
		if index < 0 {
			return filepath.ToSlash(filepath.Join(pattern[:volumeLength], ".")), pattern[:volumeLength], pattern[volumeLength:]
		}
		end = volumeLength + index
		directoryEnd := end
		if end == volumeLength {
			directoryEnd++
		}
		if literal, ok := literalPatternPath(pattern[:directoryEnd]); ok {
			return literal, pattern[:directoryEnd], pattern[end+1:]
		}
	}
}

type globExpansion struct {
	Root  string
	Files []string
}

// expandGlobs returns one expansion per input, in the same order. Patterns with
// the same lexical root share a walk; other roots retain independent entry points.
// Config ownership and ignores are evaluated later using each result's root.
func expandGlobs(ctx context.Context, fsys vfs.FS, cwd string, patterns []string) ([]globExpansion, error) {
	type compiledGlob struct {
		inputs    []int
		files     *minimatch3.Matcher
		directory *minimatch3.Matcher
	}
	type patternKey struct {
		pattern  string
		noNegate bool
	}
	type globGroup struct {
		root     string
		patterns map[patternKey]*compiledGlob
		matchers []*compiledGlob
	}
	expansions := make([]globExpansion, len(patterns))
	groupsByRoot := make(map[string]*globGroup)
	var groups []*globGroup
	for index, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)
		// Only an authored leading bang can negate the whole input. Removing a
		// static prefix must not reinterpret a later bang or negative extglob.
		options := inputPatternOptions
		options.NoNegate = !strings.HasPrefix(pattern, "!")
		directory, prefix, relative := globDirectory(pattern)
		root, err := resolveInputPath(cwd, directory)
		if err != nil {
			return nil, err
		}
		expansions[index].Root = root
		if !fsys.DirectoryExists(root) {
			continue
		}
		// Normalize both sides in authored syntax, so escapes in a returning
		// suffix cannot be confused with the decoded filesystem root.
		if runtime.GOOS == "windows" {
			prefix, err = resolveInputPath(cwd, prefix)
			if err != nil {
				return nil, err
			}
		} else if !filepath.IsAbs(prefix) {
			prefix = filepath.Join(cwd, prefix)
		}
		pattern, err = filepath.Rel(prefix, filepath.Join(prefix, relative))
		if err != nil {
			return nil, err
		}
		pattern = filepath.ToSlash(pattern)
		group := groupsByRoot[root]
		if group == nil {
			group = &globGroup{root: root, patterns: make(map[patternKey]*compiledGlob)}
			groupsByRoot[root] = group
			groups = append(groups, group)
		}
		key := patternKey{pattern: pattern, noNegate: options.NoNegate}
		matcher := group.patterns[key]
		if matcher == nil {
			// Full and partial matching share the original input's grammar.
			partialOptions := options
			partialOptions.Partial = true
			matcher = &compiledGlob{
				files:     minimatch3.New(pattern, options),
				directory: minimatch3.New(pattern, partialOptions),
			}
			group.patterns[key] = matcher
			group.matchers = append(group.matchers, matcher)
		}
		matcher.inputs = append(matcher.inputs, index)
	}
	for _, group := range groups {
		err := fsys.WalkDir(group.root, func(path string, entry vfs.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if walkErr != nil {
				return vfs.SkipDir
			}
			path = tspath.NormalizePath(path)
			relative := tspath.GetRelativePathFromDirectory(group.root, path, tspath.ComparePathsOptions{UseCaseSensitiveFileNames: true})
			if entry.IsDir() {
				if rslintconfig.IsDefaultExcludedPath(path, group.root, fsys.UseCaseSensitiveFileNames()) {
					return vfs.SkipDir
				}
				if path == group.root {
					return nil
				}
				for _, matcher := range group.matchers {
					if matcher.directory.Match(relative) {
						return nil
					}
				}
				return vfs.SkipDir
			}
			exists := false
			for _, matcher := range group.matchers {
				if !matcher.files.Match(relative) {
					continue
				}
				if !exists {
					if !fsys.FileExists(path) {
						return nil
					}
					exists = true
				}
				for _, index := range matcher.inputs {
					expansions[index].Files = append(expansions[index].Files, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return expansions, nil
}
