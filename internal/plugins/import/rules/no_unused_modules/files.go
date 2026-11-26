package no_unused_modules

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils/isglob"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

// filePattern is one `src` or `ignoreExports` entry, resolved against the
// process working directory the way ESLint's FileEnumerator resolves CLI
// patterns: a glob selects files by the pattern alone, an existing directory
// selects the files below it with an analyzed extension, and anything else
// names one file.
type filePattern struct {
	path    string // absolute, normalized; for a glob, its static parent
	glob    *minimatch3.Matcher
	recurse bool // a glob whose pattern part can match below its parent
}

type fileSelection struct {
	patterns   []filePattern
	extensions []string
	fs         vfs.FS
}

func newFileSelection(fs vfs.FS, cwd string, patterns []string, extensions []string) *fileSelection {
	selection := &fileSelection{extensions: extensions, fs: fs}
	for _, pattern := range patterns {
		pattern = tspath.NormalizeSlashes(pattern)
		if !isglob.Is(pattern) {
			selection.patterns = append(selection.patterns, filePattern{path: tspath.ResolvePath(cwd, pattern)})
			continue
		}
		absolute := pattern
		if !tspath.IsRootedDiskPath(pattern) {
			absolute = tspath.CombinePaths(cwd, pattern)
		}
		parent, rest := globParent(absolute)
		selection.patterns = append(selection.patterns, filePattern{
			path: tspath.NormalizePath(parent),
			// ESLint matches with `dot: true` and leaves dotfiles to its
			// default ignore patterns, which every walk below applies.
			glob:    minimatch3.New(absolute, minimatch3.Options{Dot: true}),
			recurse: strings.Contains(rest, "**") || strings.Contains(rest, "/"),
		})
	}
	return selection
}

// globParent returns the longest leading run of path segments without glob
// syntax, and the remaining pattern.
func globParent(pattern string) (string, string) {
	segments := strings.Split(pattern, "/")
	for index, segment := range segments {
		if isglob.Is(segment) {
			parent := strings.Join(segments[:index], "/")
			if parent == "" {
				parent = "/"
			}
			return parent, strings.Join(segments[index:], "/")
		}
	}
	return pattern, ""
}

// matches reports whether the selection contains fileName, an absolute
// normalized path. It agrees with enumerate without walking the filesystem.
func (selection *fileSelection) matches(fileName string) bool {
	for _, pattern := range selection.patterns {
		if pattern.matches(selection, fileName) {
			return true
		}
	}
	return false
}

func (pattern filePattern) matches(selection *fileSelection, fileName string) bool {
	if pattern.glob != nil {
		relative, ok := relativeBelow(pattern.path, fileName)
		return ok && !defaultIgnored(relative) && (pattern.recurse || !strings.Contains(relative, "/")) && pattern.glob.Match(fileName)
	}
	if fileName == pattern.path {
		return selection.fs.FileExists(fileName)
	}
	relative, ok := relativeBelow(pattern.path, fileName)
	return ok && !defaultIgnored(relative) && selection.hasExtension(fileName) && selection.fs.DirectoryExists(pattern.path)
}

// enumerate lists the selected files in a stable order.
func (selection *fileSelection) enumerate() []string {
	seen := make(map[string]struct{})
	var files []string
	add := func(fileName string) {
		if _, ok := seen[fileName]; !ok {
			seen[fileName] = struct{}{}
			files = append(files, fileName)
		}
	}
	for _, pattern := range selection.patterns {
		switch {
		case pattern.glob != nil:
			selection.walk(pattern.path, pattern.recurse, func(fileName string) {
				if pattern.glob.Match(fileName) {
					add(fileName)
				}
			})
		case selection.fs.DirectoryExists(pattern.path):
			selection.walk(pattern.path, true, func(fileName string) {
				if selection.hasExtension(fileName) {
					add(fileName)
				}
			})
		case selection.fs.FileExists(pattern.path):
			add(pattern.path)
		}
	}
	return files
}

// walk visits the files below directory, skipping what ESLint ignores by
// default: dotfiles, dot directories and node_modules. A directory reached
// again through a symbolic link is not walked twice.
func (selection *fileSelection) walk(directory string, recurse bool, visit func(string)) {
	visited := make(map[string]struct{})
	var walk func(string)
	walk = func(directory string) {
		realDirectory := selection.fs.Realpath(directory)
		if _, ok := visited[realDirectory]; ok {
			return
		}
		visited[realDirectory] = struct{}{}
		entries := selection.fs.GetAccessibleEntries(directory)
		files := slices.Clone(entries.Files)
		slices.Sort(files)
		for _, name := range files {
			if !strings.HasPrefix(name, ".") {
				visit(tspath.CombinePaths(directory, name))
			}
		}
		if !recurse {
			return
		}
		directories := slices.Clone(entries.Directories)
		slices.Sort(directories)
		for _, name := range directories {
			if !strings.HasPrefix(name, ".") && name != "node_modules" {
				walk(tspath.CombinePaths(directory, name))
			}
		}
	}
	walk(directory)
}

func (selection *fileSelection) hasExtension(fileName string) bool {
	return slices.ContainsFunc(selection.extensions, func(extension string) bool {
		return strings.HasSuffix(fileName, extension)
	})
}

// relativeBelow returns fileName relative to directory when it lies below it.
func relativeBelow(directory, fileName string) (string, bool) {
	prefix := tspath.EnsureTrailingDirectorySeparator(directory)
	if !strings.HasPrefix(fileName, prefix) {
		return "", false
	}
	return fileName[len(prefix):], true
}

func defaultIgnored(relative string) bool {
	for segment := range strings.SplitSeq(relative, "/") {
		if strings.HasPrefix(segment, ".") || segment == "node_modules" {
			return true
		}
	}
	return false
}
