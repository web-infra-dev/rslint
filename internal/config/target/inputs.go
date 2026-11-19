package target

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
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

// InputMatch connects an operand's filesystem spelling to the paths selected
// from it. For a glob, Operand and Path name the static root, and Files is the
// existing expansion. A literal directory is expanded later by target planning.
type InputMatch struct {
	Operand   string
	Path      string
	Directory bool
	Files     []string
}

type ownedInput struct{ owner, path string }

// InputRoots optionally records which independent scans reached an exact path
// under its config owner. Callers that only need the target set leave it nil.
type InputRoots map[ownedInput][]string

func (roots InputRoots) Add(owner, path, root string) {
	if roots == nil {
		return
	}
	key := ownedInput{owner, path}
	if !slices.Contains(roots[key], root) {
		roots[key] = append(roots[key], root)
	}
}

func (roots InputRoots) For(owner, path string) []string {
	return roots[ownedInput{owner, path}]
}

// SelectInputs interprets raw file, directory, and glob inputs against fsys.
// Exact-file APIs bypass this operation. Unmatched inputs remain separate so
// each entrypoint keeps its existing reporting policy without admitting overlay
// dependencies as fallback targets.
func SelectInputs(ctx context.Context, fsys vfs.FS, cwd string, inputs []string) (*InputSelection, error) {
	return SelectInputsWithMatches(ctx, fsys, cwd, inputs, nil)
}

// SelectInputsWithMatches also reports how each successful operand matched.
func SelectInputsWithMatches(ctx context.Context, fsys vfs.FS, cwd string, inputs []string, onMatch func(InputMatch)) (*InputSelection, error) {
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
		absolute, err := resolveInputPath(cwd, raw)
		if err != nil {
			return nil, err
		}
		classified[index].path = absolute
		if fsys.FileExists(absolute) {
			classified[index].file = true
			continue
		}
		if fsys.DirectoryExists(absolute) {
			classified[index].directory = true
			continue
		}
		pattern := filepath.ToSlash(raw)
		decoded, literal := literalPatternPath(pattern)
		if raw != "" && (!literal || decoded != pattern) {
			if literal {
				directory, err := resolveInputPath(cwd, decoded)
				if err != nil {
					return nil, err
				}
				if fsys.DirectoryExists(directory) {
					pattern = strings.TrimRight(pattern, "/") + "/**/*"
				}
			}
			classified[index].glob = true
			patterns = append(patterns, pattern)
		}
	}
	expansions, err := expandGlobs(ctx, fsys, cwd, patterns)
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
			if onMatch != nil {
				onMatch(InputMatch{Operand: inputs[index], Path: input.path, Directory: true})
			}
			if !directories[input.path] {
				directories[input.path] = true
				selection.Directories = append(selection.Directories, input.path)
			}
			continue
		}
		if !input.glob {
			if input.file {
				if onMatch != nil {
					onMatch(InputMatch{Operand: inputs[index], Path: input.path})
				}
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
		if onMatch != nil {
			operand, _, _ := globDirectory(patterns[globIndex-1])
			onMatch(InputMatch{Operand: operand, Path: expansion.Root, Files: expansion.Files})
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

// resolveInputPath resolves filesystem paths, not glob syntax. Windows drive-
// relative inputs use the request's cwd on that drive; other drives retain the
// operating system's per-drive working directory. Native cleaning protects UNC
// share roots from parent segments without changing the process cwd.
func resolveInputPath(cwd, name string) (string, error) {
	if runtime.GOOS != "windows" {
		return tspath.ResolvePath(cwd, name), nil
	}
	if !filepath.IsAbs(name) {
		volume := filepath.VolumeName(name)
		switch {
		case volume != "":
			if strings.EqualFold(volume, filepath.VolumeName(cwd)) {
				name = filepath.Join(cwd, name[len(volume):])
			} else {
				absolute, err := filepath.Abs(name)
				if err != nil {
					return "", err
				}
				name = absolute
			}
		case len(name) > 0 && os.IsPathSeparator(name[0]):
			name = filepath.VolumeName(cwd) + name
		default:
			name = filepath.Join(cwd, name)
		}
	}
	return filepath.ToSlash(filepath.Clean(name)), nil
}
