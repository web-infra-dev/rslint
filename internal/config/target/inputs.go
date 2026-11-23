package target

import (
	"cmp"
	"context"
	"os"
	"path/filepath"
	"runtime"
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

// SelectInputs interprets raw file, directory, and glob inputs against fsys.
// Exact-file APIs bypass this operation. Unmatched inputs remain separate so
// each entrypoint keeps its existing reporting policy without admitting overlay
// dependencies as fallback targets.
func SelectInputs(ctx context.Context, fsys vfs.FS, cwd string, inputs []string) (*InputSelection, error) {
	return (InputRoot{Directory: cwd}).Select(ctx, fsys, inputs)
}

// InputRoot anchors ordinary paths at Directory. The file CLI also supplies
// the native Windows working directory for current-drive and drive-relative
// names, which can be on another volume when Directory resolves a junction.
type InputRoot struct {
	Directory              string
	WindowsVolumeDirectory string
}

// Select expands the request's literal files, directories, and glob patterns.
func (base InputRoot) Select(ctx context.Context, fsys vfs.FS, inputs []string) (*InputSelection, error) {
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
		absolute, err := base.ResolvePath(raw)
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
				directory, err := base.ResolvePath(decoded)
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
	expansions, err := base.expandGlobs(ctx, fsys, patterns)
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

// ResolvePath resolves filesystem paths, not glob syntax. A drive-relative
// path on another Windows volume uses that drive's OS working directory.
func (base InputRoot) ResolvePath(name string) (string, error) {
	if runtime.GOOS != "windows" {
		return tspath.ResolvePath(base.Directory, name), nil
	}
	if !filepath.IsAbs(name) {
		volume := filepath.VolumeName(name)
		volumeDir := cmp.Or(base.WindowsVolumeDirectory, base.Directory)
		switch {
		case volume != "":
			if strings.EqualFold(volume, filepath.VolumeName(volumeDir)) {
				name = filepath.Join(volumeDir, name[len(volume):])
			} else {
				absolute, err := filepath.Abs(name)
				if err != nil {
					return "", err
				}
				name = absolute
			}
		case len(name) > 0 && os.IsPathSeparator(name[0]):
			name = filepath.VolumeName(volumeDir) + name
		default:
			name = filepath.Join(base.Directory, name)
		}
	}
	return filepath.ToSlash(filepath.Clean(name)), nil
}
