package utils

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

// TypeScriptProjectDiscovery searches parsed config roots in one filesystem
// generation. It never reads source files or constructs Programs. The caller
// owns config parsing and constructs the selected project through its existing
// Program loader. Calls are sequential.
type TypeScriptProjectDiscovery struct {
	fs          vfs.FS
	parseConfig func(string) (*tsoptions.ParsedCommandLine, error)
	configs     map[tspath.Path]*tsoptions.ParsedCommandLine
}

func NewTypeScriptProjectDiscovery(
	fs vfs.FS,
	parseConfig func(string) (*tsoptions.ParsedCommandLine, error),
) *TypeScriptProjectDiscovery {
	return &TypeScriptProjectDiscovery{
		fs: fs, parseConfig: parseConfig,
		configs: make(map[tspath.Path]*tsoptions.ParsedCommandLine),
	}
}

func (d *TypeScriptProjectDiscovery) path(name string) tspath.Path {
	return tspath.ToPath(name, "", d.fs.UseCaseSensitiveFileNames())
}

func (d *TypeScriptProjectDiscovery) config(name string) (*tsoptions.ParsedCommandLine, error) {
	key := d.path(name)
	if parsed, known := d.configs[key]; known {
		return parsed, nil
	}
	if !d.fs.FileExists(name) {
		d.configs[key] = nil
		return nil, nil //nolint:nilnil // No config at this location is a normal search miss.
	}
	parsed, err := d.parseConfig(name)
	if err != nil {
		return nil, fmt.Errorf("parse TypeScript project %q: %w", name, err)
	}
	if parsed == nil {
		return nil, fmt.Errorf("parse TypeScript project %q: no parsed config returned", name)
	}
	d.configs[key] = parsed
	return parsed, nil
}

type typeScriptProjectSearch struct {
	discovery *TypeScriptProjectDiscovery
	file      tspath.Path
	fallback  *tsoptions.ParsedCommandLine
	seen      map[tspath.Path]bool
}

// referencedRoot identifies source redirects from config metadata. A cycle
// back to the containing config is not a redirect away from that config.
func (q *typeScriptProjectSearch) referencedRoot(parsed *tsoptions.ParsedCommandLine, seen map[tspath.Path]bool) (bool, error) {
	for _, name := range parsed.ResolvedProjectReferencePaths() {
		key := q.discovery.path(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		child, err := q.discovery.config(name)
		if err != nil {
			return false, err
		}
		if child == nil {
			continue
		}
		if _, direct := child.FileNamesByPath()[q.file]; direct {
			return true, nil
		}
		if found, err := q.referencedRoot(child, seen); err != nil || found {
			return found, err
		}
	}
	return false, nil
}

func (q *typeScriptProjectSearch) direct(parsed *tsoptions.ParsedCommandLine) (*tsoptions.ParsedCommandLine, error) {
	if _, included := parsed.FileNamesByPath()[q.file]; !included {
		return nil, nil //nolint:nilnil // Imports do not make a lint target a configured root.
	}
	options := parsed.CompilerOptions()
	if options.DisableSourceOfProjectReferenceRedirect.IsTrue() {
		return parsed, nil
	}
	redirected, err := q.referencedRoot(parsed, map[tspath.Path]bool{q.discovery.path(parsed.ConfigName()): true})
	if err != nil {
		return nil, err
	}
	if !redirected {
		return parsed, nil
	}
	if q.fallback == nil {
		q.fallback = parsed
	}
	return nil, nil //nolint:nilnil // Prefer the referenced project's own compiler options.
}

// references checks immediate references before their subtrees. Disabling
// reference loading stops this discovery step regardless of earlier requests;
// discovery does not reproduce the TypeScript server's warm-project state.
func (q *typeScriptProjectSearch) references(parsed *tsoptions.ParsedCommandLine) (*tsoptions.ParsedCommandLine, error) {
	if parsed.CompilerOptions().DisableReferencedProjectLoad.IsTrue() {
		return nil, nil //nolint:nilnil // The config explicitly stops reference discovery.
	}
	var children []*tsoptions.ParsedCommandLine
	for _, name := range parsed.ResolvedProjectReferencePaths() {
		key := q.discovery.path(name)
		if q.seen[key] {
			continue
		}
		q.seen[key] = true
		child, err := q.discovery.config(name)
		if err != nil {
			return nil, err
		}
		if child == nil {
			continue
		}
		if selected, err := q.direct(child); err != nil || selected != nil {
			return selected, err
		}
		children = append(children, child)
	}
	for _, child := range children {
		if selected, err := q.references(child); err != nil || selected != nil {
			return selected, err
		}
	}
	return nil, nil //nolint:nilnil // No referenced config directly includes this target.
}

// Find returns a config whose files/include roots contain fileName, or nil
// when the target must use gap linting. rootDirectory bounds ancestor search;
// extends and explicit references can still point outside that directory.
func (d *TypeScriptProjectDiscovery) Find(fileName, rootDirectory string) (*tsoptions.ParsedCommandLine, error) {
	query := typeScriptProjectSearch{
		discovery: d, file: d.path(tspath.NormalizePath(fileName)), seen: make(map[tspath.Path]bool),
	}
	root := d.path(tspath.GetNormalizedAbsolutePath(rootDirectory, ""))
	for directory := tspath.GetDirectoryPath(fileName); ; directory = tspath.GetDirectoryPath(directory) {
		for _, name := range []string{"tsconfig.json", "jsconfig.json"} {
			parsed, err := d.config(tspath.ResolvePath(directory, name))
			if err != nil {
				return nil, err
			}
			if parsed == nil {
				continue
			}
			query.seen[d.path(parsed.ConfigName())] = true
			if selected, err := query.direct(parsed); err != nil || selected != nil {
				return selected, err
			}
			if selected, err := query.references(parsed); err != nil || selected != nil {
				return selected, err
			}
			if parsed.CompilerOptions().DisableSolutionSearching.IsTrue() {
				return query.fallback, nil
			}
		}
		if d.path(directory) == root || tspath.GetBaseFileName(string(d.path(directory))) == "node_modules" ||
			tspath.GetDirectoryPath(directory) == directory {
			return query.fallback, nil
		}
	}
}
