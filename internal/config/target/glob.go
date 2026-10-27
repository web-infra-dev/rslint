package target

import (
	"context"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

type GlobExpansion struct {
	Root  string
	Files []string
}

// ExpandGlobs returns one expansion per input, in the same order. Patterns with
// the same lexical root share a walk; other roots retain independent entry points.
// Config ownership and ignores are evaluated later using each result's root.
func ExpandGlobs(ctx context.Context, fsys vfs.FS, cwd string, patterns []string) ([]GlobExpansion, error) {
	type compiledGlob struct {
		inputs    []int
		files     *minimatch3.Matcher
		directory *minimatch3.Matcher
	}
	type globGroup struct {
		root     string
		patterns map[string]*compiledGlob
		matchers []*compiledGlob
	}
	expansions := make([]GlobExpansion, len(patterns))
	groupsByRoot := make(map[string]*globGroup)
	var groups []*globGroup
	// Full and partial matching share a grammar so pruning preserves selected files.
	options := minimatch3.Options{Dot: true, NoNegate: true, NoComment: true, NoExt: true, PreserveWhitespace: true}
	partialOptions := options
	partialOptions.Partial = true
	for index, pattern := range patterns {
		base, pattern := doublestar.SplitPattern(filepath.ToSlash(filepath.Clean(pattern)))
		root := tspath.ResolvePath(cwd, base)
		expansions[index].Root = root
		if !doublestar.ValidatePattern(pattern) || !fsys.DirectoryExists(root) {
			continue
		}
		group := groupsByRoot[root]
		if group == nil {
			group = &globGroup{root: root, patterns: make(map[string]*compiledGlob)}
			groupsByRoot[root] = group
			groups = append(groups, group)
		}
		matcher := group.patterns[pattern]
		if matcher == nil {
			matcher = &compiledGlob{
				files:     minimatch3.New(pattern, options),
				directory: minimatch3.New(pattern, partialOptions),
			}
			group.patterns[pattern] = matcher
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
