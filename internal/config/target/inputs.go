package target

import (
	"context"
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils/minimatch3"
)

// InputSelection preserves literal directories and each glob's independent scan
// entry. It does not choose config owners or apply user ignores; discovery and
// target planning own those decisions.
type InputSelection struct {
	Files           []string
	Unmatched       []string
	Directories     []string
	GlobRoots       []string
	ScanDirectories map[string][]string
}

// SelectInputs interprets raw file, directory, and glob inputs against fsys.
// Exact-file APIs bypass this operation. Unmatched inputs remain separate so
// each entrypoint keeps its existing reporting policy without admitting overlay
// dependencies as fallback targets.
func SelectInputs(ctx context.Context, fsys vfs.FS, cwd string, inputs []string) (*InputSelection, error) {
	selection := &InputSelection{Files: []string{}, ScanDirectories: make(map[string][]string)}
	type input struct {
		path      string
		directory bool
		glob      bool
		file      bool
	}
	classified := make([]input, len(inputs))
	var patterns []string
	for index, raw := range inputs {
		absolute := tspath.ResolvePath(cwd, raw)
		classified[index].path = absolute
		if fsys.FileExists(absolute) {
			classified[index].file = true
			continue
		}
		if fsys.DirectoryExists(absolute) {
			classified[index].directory = true
			continue
		}
		if _, literal := minimatch3.New(filepath.ToSlash(raw), inputPatternOptions).Literal(); !literal && raw != "" {
			classified[index].glob = true
			patterns = append(patterns, raw)
		}
	}
	expansions, err := ExpandGlobs(ctx, fsys, cwd, patterns)
	if err != nil {
		return nil, err
	}
	literals := make(map[string]bool)
	files := make(map[string]bool)
	directories := make(map[string]bool)
	roots := make(map[string]bool)
	unmatched := make(map[string]bool)
	type scanEntry struct{ file, root string }
	scans := make(map[scanEntry]bool)
	addFile := func(path string) {
		if !files[path] {
			files[path] = true
			selection.Files = append(selection.Files, path)
		}
	}
	globIndex := 0
	for index, input := range classified {
		if inputs[index] == "" {
			continue
		}
		if input.directory {
			if !directories[input.path] {
				directories[input.path] = true
				selection.Directories = append(selection.Directories, input.path)
			}
			continue
		}
		if !input.glob {
			if input.file {
				addFile(input.path)
				literals[input.path] = true
			} else if !unmatched[input.path] {
				unmatched[input.path] = true
				selection.Unmatched = append(selection.Unmatched, input.path)
			}
			continue
		}
		expansion := expansions[globIndex]
		globIndex++
		if len(expansion.Files) == 0 {
			if !unmatched[input.path] {
				unmatched[input.path] = true
				selection.Unmatched = append(selection.Unmatched, input.path)
			}
			continue
		}
		if !roots[expansion.Root] {
			roots[expansion.Root] = true
			selection.GlobRoots = append(selection.GlobRoots, expansion.Root)
		}
		for _, file := range expansion.Files {
			addFile(file)
			entry := scanEntry{file, expansion.Root}
			if !scans[entry] {
				scans[entry] = true
				selection.ScanDirectories[file] = append(selection.ScanDirectories[file], expansion.Root)
			}
		}
	}
	for file := range literals {
		delete(selection.ScanDirectories, file)
	}
	return selection, nil
}
