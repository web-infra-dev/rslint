// Package projectservice selects configured TypeScript projects for linted
// files. It owns only project discovery and membership; callers own the
// filesystem generation, config parsing, and complete Program construction.
package projectservice

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// Host supplies one immutable filesystem generation. CreateProgram must retain
// the parsed config's complete roots, enable source project references, and
// allow non-TS extensions after config parsing, as TypeScript's service does.
// ParseConfig is called only for existing configs and must return a parsed
// config or an error; an unreadable config is not an ordinary search miss.
type Host struct {
	FS            vfs.FS
	ParseConfig   func(configPath string) (*tsoptions.ParsedCommandLine, error)
	CreateProgram func(configPath string, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error)
	// LoadedProgram optionally exposes a previously loaded configured project
	// from the caller's store, updated for this filesystem generation. It must
	// return nil for a project that is not loaded and must not create that project.
	// Only searches with reference loading disabled consult this callback.
	LoadedProgram func(configPath string) (*compiler.Program, error)
}

// Selection is the configured project that supplies this file's type context.
// A zero Selection means no configured project owns the file. Callers decide
// how to lint that file without a configured type context.
type Selection struct {
	Program    *compiler.Program
	ConfigPath string
}

type candidate struct {
	path    string
	parsed  *tsoptions.ParsedCommandLine
	program *compiler.Program
}

// Selector retains parsed configs and Programs for one caller-owned source
// generation. Select calls are sequential; it is not shared between requests.
type Selector struct {
	host       Host
	candidates map[tspath.Path]*candidate
}

func New(host Host) *Selector {
	return &Selector{host: host, candidates: make(map[tspath.Path]*candidate)}
}

func (s *Selector) path(fileName string) tspath.Path {
	return tspath.ToPath(fileName, "", s.host.FS.UseCaseSensitiveFileNames())
}

func (s *Selector) candidate(configPath string, allowLoad bool) (*candidate, error) {
	configPath = tspath.NormalizePath(configPath)
	key := s.path(configPath)
	entry, exists := s.candidates[key]
	if exists && (allowLoad || entry == nil || entry.program != nil) {
		return entry, nil
	}
	if !allowLoad {
		if s.host.LoadedProgram == nil {
			return entry, nil
		}
		program, err := s.host.LoadedProgram(configPath)
		if err != nil {
			return nil, fmt.Errorf("read loaded TypeScript project %q: %w", configPath, err)
		}
		if program == nil {
			return entry, nil
		}
		if entry == nil {
			entry = &candidate{path: configPath}
			s.candidates[key] = entry
		}
		entry.parsed = program.CommandLine()
		s.setProgram(entry, program)
		return entry, nil
	}
	if !s.host.FS.FileExists(configPath) {
		s.candidates[key] = nil
		return nil, nil //nolint:nilnil // No configured project is a normal search miss.
	}
	parsed, err := s.host.ParseConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("parse TypeScript project %q: %w", configPath, err)
	}
	if parsed == nil {
		return nil, fmt.Errorf("parse TypeScript project %q: no parsed config returned", configPath)
	}
	entry = &candidate{path: configPath, parsed: parsed}
	s.candidates[key] = entry
	return entry, nil
}

func (s *Selector) build(entry *candidate) error {
	if entry.program != nil {
		return nil
	}
	program, err := s.host.CreateProgram(entry.path, entry.parsed)
	if err != nil {
		return fmt.Errorf("create TypeScript project %q: %w", entry.path, err)
	}
	if program == nil {
		return fmt.Errorf("create TypeScript project %q: no Program returned", entry.path)
	}
	s.setProgram(entry, program)
	return nil
}

func (s *Selector) setProgram(entry *candidate, program *compiler.Program) {
	entry.program = program
	// The compiler may parse referenced configs while constructing this Program.
	// Those configs are available to a later disabled-reference-load search, but
	// parsing one does not mean its own configured Program has been loaded.
	program.RangeResolvedProjectReference(func(path tspath.Path, parsed, _ *tsoptions.ParsedCommandLine, _ int) bool {
		if parsed != nil {
			if _, exists := s.candidates[path]; !exists {
				s.candidates[path] = &candidate{path: tspath.NormalizePath(parsed.ConfigName()), parsed: parsed}
			}
		}
		return true
	})
}

type search struct {
	selector *Selector
	fileName string
	filePath tspath.Path
	fallback Selection
	// A config first reached through a disabled reference may be revisited by
	// a later branch that permits loading it.
	seenReferences map[tspath.Path]bool
}

func (q *search) membership(entry *candidate, referenced, allowLoad bool) (Selection, error) {
	_, direct := entry.parsed.FileNamesByPath()[q.filePath]
	if !direct && (referenced || entry.parsed.CompilerOptions().Composite.IsTrue()) {
		return Selection{}, nil
	}
	// allowJs filters config globs and ordinary imports, but a service Program
	// can still own JavaScript through a triple-slash reference. Let the actual
	// Program decide membership instead of rejecting it from parsed options.
	if len(entry.parsed.FileNames()) == 0 {
		return Selection{}, nil
	}
	if entry.program == nil {
		if !allowLoad {
			return Selection{}, nil
		}
		if err := q.selector.build(entry); err != nil {
			return Selection{}, err
		}
	}
	program := entry.program
	file := program.GetSourceFile(q.fileName)
	if file == nil {
		// With source redirects disabled, a root may be represented only by the
		// referenced declaration output. Selecting another project would hide the
		// parser's inability to lint this file in the selected TypeScript context.
		if direct && program.GetProjectReferenceFromSource(q.filePath) != nil &&
			!program.IsSourceFromProjectReference(q.filePath) {
			return Selection{}, fmt.Errorf("TypeScript project %q contains %q through a project reference but does not provide its source file", entry.path, q.fileName)
		}
		return Selection{}, nil
	}
	selection := Selection{Program: program, ConfigPath: entry.path}
	if !program.IsSourceFromProjectReference(file.Path()) {
		return selection, nil
	}
	if q.fallback.Program == nil {
		q.fallback = selection
	}
	return Selection{}, nil
}

// references follows TypeScript's sibling-then-subtree order: examine every
// immediate reference before recursively searching each reference's subtree.
// A referenced candidate needs direct config membership even when a nearby
// non-composite project could admit the file through an import.
func (q *search) references(parent *candidate, allowLoad bool) (Selection, error) {
	allowLoad = allowLoad && !parent.parsed.CompilerOptions().DisableReferencedProjectLoad.IsTrue()
	var children []*candidate
	for _, configPath := range parent.parsed.ResolvedProjectReferencePaths() {
		key := q.selector.path(configPath)
		if previous, seen := q.seenReferences[key]; seen && (previous || !allowLoad) {
			continue
		}
		child, err := q.selector.candidate(configPath, allowLoad)
		if err != nil {
			return Selection{}, err
		}
		if child == nil {
			continue
		}
		q.seenReferences[key] = allowLoad
		selection, err := q.membership(child, true, allowLoad)
		if err != nil || selection.Program != nil {
			return selection, err
		}
		children = append(children, child)
	}
	for _, child := range children {
		selection, err := q.references(child, allowLoad)
		if err != nil || selection.Program != nil {
			return selection, err
		}
	}
	return Selection{}, nil
}

// Select searches from the caller-visible file path. rootDirectory bounds
// ancestor discovery only when the search actually reaches that directory;
// config extends and references remain free to point outside that boundary.
// Ordinary search misses return a zero Selection and no error; host failures
// and selected projects that cannot provide the source still return errors.
func (s *Selector) Select(fileName, rootDirectory string) (Selection, error) {
	fileName = tspath.NormalizePath(fileName)
	query := search{
		selector:       s,
		fileName:       fileName,
		filePath:       s.path(fileName),
		seenReferences: make(map[tspath.Path]bool),
	}
	rootPath := s.path(tspath.GetNormalizedAbsolutePath(rootDirectory, ""))
	for directory := tspath.GetDirectoryPath(fileName); ; directory = tspath.GetDirectoryPath(directory) {
		for _, name := range []string{"tsconfig.json", "jsconfig.json"} {
			entry, err := s.candidate(tspath.ResolvePath(directory, name), true)
			if err != nil {
				return Selection{}, err
			}
			if entry == nil {
				continue
			}
			selection, err := query.membership(entry, false, true)
			if err != nil || selection.Program != nil {
				return selection, err
			}
			selection, err = query.references(entry, true)
			if err != nil || selection.Program != nil {
				return selection, err
			}
			if entry.parsed.CompilerOptions().DisableSolutionSearching.IsTrue() {
				return query.fallback, nil
			}
		}
		if s.path(directory) == rootPath || tspath.GetBaseFileName(string(s.path(directory))) == "node_modules" ||
			tspath.GetDirectoryPath(directory) == directory {
			return query.fallback, nil
		}
	}
}
