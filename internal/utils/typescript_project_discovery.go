package utils

import (
	"fmt"
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
)

var typeScriptConfigFileNames = [2]string{"tsconfig.json", "jsconfig.json"}

type searchBoundaryKey struct {
	directory     string
	rootDirectory string
}

type dirEntry struct {
	dir           string
	exactID       string
	parentDir     string
	parent        *dirEntry
	isNodeModules bool
	atRoot        bool
	slots         [2]*configEntry
}

type configFlags uint8

const (
	configDisableRedirect configFlags = 1 << iota
	configDisableRefLoad
	configDisableSolution
)

const (
	projectRefChunkSize         = 32
	maxDiscoveryRedirectEntries = 4096
)

// missingConfigEntry is a read-only sentinel stored in cached reference and
// directory slots once a path has been probed and found absent.
var missingConfigEntry configEntry

type projectRef struct {
	path  tspath.Path
	entry *configEntry
}

type redirectFrame struct {
	entry *configEntry
	index int
}

// redirectSearch tracks an incremental DFS over configs reachable from one
// containing config (excluding itself) for source-redirect checks.
type redirectSearch struct {
	stack    []redirectFrame
	seen     map[tspath.Path]bool
	visited  []*configEntry
	complete bool
}

// configRefState holds project-reference and redirect-search state only for
// configs that declare references, keeping leaf configEntry values compact.
type configRefState struct {
	lookupPath       tspath.Path
	rawRefs          []string
	firstRef         projectRef
	refChunks        [][]projectRef
	refCount         int
	redirect         *redirectSearch
	redirectQueried  bool
	redirectDisabled bool
}

type configEntry struct {
	parsed     *tsoptions.ParsedCommandLine
	fileNames  map[tspath.Path]string
	configPath tspath.Path
	refs       *configRefState
	flags      configFlags
}

func (e *configEntry) rawRefPaths() []string {
	if e.refs == nil {
		return nil
	}
	if e.refs.rawRefs == nil {
		e.refs.rawRefs = e.parsed.ResolvedProjectReferencePaths()
	}
	return e.refs.rawRefs
}

func (e *configEntry) refAt(d *TypeScriptProjectDiscovery, index int) *projectRef {
	r := e.refs
	if index < r.refCount {
		if index == 0 {
			return &r.firstRef
		}
		rem := index - 1
		return &r.refChunks[rem/projectRefChunkSize][rem%projectRefChunkSize]
	}
	rawRefs := e.rawRefPaths()
	for r.refCount <= index {
		if r.refCount == 0 {
			r.firstRef = projectRef{path: d.path(rawRefs[0])}
			r.refCount = 1
			continue
		}
		rem := r.refCount - 1
		chunkIdx := rem / projectRefChunkSize
		slotIdx := rem % projectRefChunkSize
		if slotIdx == 0 {
			chunkLen := min(projectRefChunkSize, len(rawRefs)-r.refCount)
			r.refChunks = append(r.refChunks, make([]projectRef, chunkLen))
		}
		r.refChunks[chunkIdx][slotIdx] = projectRef{
			path: d.path(rawRefs[r.refCount]),
		}
		r.refCount++
	}
	if index == 0 {
		return &r.firstRef
	}
	rem := index - 1
	return &r.refChunks[rem/projectRefChunkSize][rem%projectRefChunkSize]
}

// TypeScriptProjectDiscovery searches parsed config roots in one filesystem
// generation. It never reads source files or constructs Programs. The caller
// owns config parsing and constructs the selected project through its existing
// Program loader. Calls are sequential.
type TypeScriptProjectDiscovery struct {
	fs               vfs.FS
	useCaseSensitive bool
	parseConfig      func(string) (*tsoptions.ParsedCommandLine, error)
	configs          map[tspath.Path]*configEntry
	lastDir          *dirEntry
	dirs             map[string]*dirEntry
	hasLastBoundary  bool
	lastBoundaryKey  searchBoundaryKey
	lastBoundary     string
	boundaries       map[searchBoundaryKey]string
	seenGen          map[tspath.Path]uint32
	searchGen        uint32
	redirectSeenGen  map[tspath.Path]uint32
	redirectGen      uint32
	redirectStack    []redirectFrame
	redirectEntries  int
}

func NewTypeScriptProjectDiscovery(
	fs vfs.FS,
	parseConfig func(string) (*tsoptions.ParsedCommandLine, error),
) *TypeScriptProjectDiscovery {
	return &TypeScriptProjectDiscovery{
		fs:               fs,
		useCaseSensitive: fs != nil && fs.UseCaseSensitiveFileNames(),
		parseConfig:      parseConfig,
		configs:          make(map[tspath.Path]*configEntry),
	}
}

func (d *TypeScriptProjectDiscovery) path(name string) tspath.Path {
	return tspath.ToPath(name, "", d.useCaseSensitive)
}

func (d *TypeScriptProjectDiscovery) config(name string) (*configEntry, error) {
	return d.configByPath(name, d.path(name))
}

func (d *TypeScriptProjectDiscovery) configByPath(name string, key tspath.Path) (*configEntry, error) {
	if entry, known := d.configs[key]; known {
		return entry, nil
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
	options := parsed.CompilerOptions()
	configPath := key
	if configName := parsed.ConfigName(); configName != "" && configName != name {
		configPath = d.path(configName)
	}
	var flags configFlags
	if options != nil {
		if options.DisableSourceOfProjectReferenceRedirect.IsTrue() {
			flags |= configDisableRedirect
		}
		if options.DisableReferencedProjectLoad.IsTrue() {
			flags |= configDisableRefLoad
		}
		if options.DisableSolutionSearching.IsTrue() {
			flags |= configDisableSolution
		}
	}
	var refs *configRefState
	if len(parsed.ProjectReferences()) > 0 {
		refs = &configRefState{lookupPath: key}
	}
	entry := &configEntry{
		parsed:     parsed,
		fileNames:  parsed.FileNamesByPath(),
		configPath: configPath,
		refs:       refs,
		flags:      flags,
	}
	d.configs[key] = entry
	return entry, nil
}

func (d *TypeScriptProjectDiscovery) loadRef(name string, ref *projectRef) (*configEntry, error) {
	if ref.entry != nil {
		if ref.entry == &missingConfigEntry {
			return nil, nil //nolint:nilnil // Cached missing reference.
		}
		return ref.entry, nil
	}
	entry, err := d.configByPath(name, ref.path)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		ref.entry = &missingConfigEntry
		return nil, nil //nolint:nilnil // Missing reference is a normal search miss.
	}
	ref.entry = entry
	return entry, nil
}

func (d *TypeScriptProjectDiscovery) dirEntry(dir string) *dirEntry {
	if d.lastDir != nil && d.lastDir.dir == dir {
		return d.lastDir
	}
	if entry, ok := d.dirs[dir]; ok {
		d.lastDir = entry
		return entry
	}
	parentDir := tspath.GetDirectoryPath(dir)
	entry := &dirEntry{
		dir:           dir,
		exactID:       ExactPathID(dir),
		parentDir:     parentDir,
		isNodeModules: tspath.GetBaseFileName(string(d.path(dir))) == "node_modules",
		atRoot:        parentDir == dir,
	}
	if d.lastDir != nil {
		if d.dirs == nil {
			d.dirs = make(map[string]*dirEntry, 4)
			d.dirs[d.lastDir.dir] = d.lastDir
		}
		d.dirs[dir] = entry
	}
	d.lastDir = entry
	return entry
}

func (d *TypeScriptProjectDiscovery) parentDir(entry *dirEntry) *dirEntry {
	if entry.parent == nil && !entry.atRoot {
		entry.parent = d.dirEntry(entry.parentDir)
	}
	return entry.parent
}

func (d *TypeScriptProjectDiscovery) dirSlotConfig(dir *dirEntry, index int) (*configEntry, error) {
	if slot := dir.slots[index]; slot != nil {
		if slot == &missingConfigEntry {
			return nil, nil //nolint:nilnil // Cached absent config slot.
		}
		return slot, nil
	}
	entry, err := d.config(tspath.ResolvePath(dir.dir, typeScriptConfigFileNames[index]))
	if err != nil {
		return nil, err
	}
	if entry == nil {
		dir.slots[index] = &missingConfigEntry
		return nil, nil //nolint:nilnil // No config at this location is a normal search miss.
	}
	dir.slots[index] = entry
	return entry, nil
}

func (d *TypeScriptProjectDiscovery) searchBoundary(directory, rootDirectory string) string {
	key := searchBoundaryKey{directory: directory, rootDirectory: rootDirectory}
	if d.hasLastBoundary && d.lastBoundaryKey == key {
		return d.lastBoundary
	}
	if boundary, ok := d.boundaries[key]; ok {
		d.lastBoundaryKey = key
		d.lastBoundary = boundary
		return boundary
	}
	boundary, contains := DirectorySearchBoundary(
		DirectoryIdentity{LexicalPath: directory},
		DirectoryIdentity{LexicalPath: tspath.GetNormalizedAbsolutePath(rootDirectory, "")},
		d.fs,
	)
	if !contains {
		// External targets retain TypeScript's own-ancestor search behavior.
		boundary = tspath.GetPathComponents(directory, "")[0]
	}
	exactBoundary := ExactPathID(boundary)
	if d.hasLastBoundary {
		if d.boundaries == nil {
			d.boundaries = make(map[searchBoundaryKey]string, 4)
			d.boundaries[d.lastBoundaryKey] = d.lastBoundary
		}
		d.boundaries[key] = exactBoundary
	}
	d.lastBoundaryKey = key
	d.lastBoundary = exactBoundary
	d.hasLastBoundary = true
	return exactBoundary
}

// isReferencedRoot identifies source redirects from config metadata. A cycle
// back to the containing config is not a redirect away from that config.
func (d *TypeScriptProjectDiscovery) isReferencedRoot(entry *configEntry, file tspath.Path) (bool, error) {
	r := entry.refs
	if r == nil {
		return false, nil
	}
	st := r.redirect
	if st == nil {
		// Avoid retaining per-config visited slices on configs queried only
		// once (e.g. single-file editor lookups or one-file-per-project chains)
		// and bound total cached redirect entries across the discovery instance.
		if !r.redirectQueried || r.redirectDisabled || d.redirectEntries >= maxDiscoveryRedirectEntries {
			r.redirectQueried = true
			return d.isReferencedRootUncached(entry, file)
		}
		seen := map[tspath.Path]bool{entry.configPath: true}
		if r.lookupPath != entry.configPath {
			seen[r.lookupPath] = true
		}
		st = &redirectSearch{
			stack: []redirectFrame{{entry: entry, index: 0}},
			seen:  seen,
		}
		r.redirect = st
	}
	for _, cfg := range st.visited {
		if _, ok := cfg.fileNames[file]; ok {
			return true, nil
		}
	}
	if st.complete {
		return false, nil
	}
	for len(st.stack) > 0 {
		top := &st.stack[len(st.stack)-1]
		rawRefs := top.entry.rawRefPaths()
		if top.index >= len(rawRefs) {
			st.stack = st.stack[:len(st.stack)-1]
			continue
		}
		ref := top.entry.refAt(d, top.index)
		if st.seen[ref.path] {
			top.index++
			continue
		}
		child, err := d.loadRef(rawRefs[top.index], ref)
		if err != nil {
			if len(st.visited) == 0 {
				r.redirect = nil
			}
			return false, err
		}
		top.index++
		// Missing configs are already cached in d.configs and ref.entry; do not
		// retain them in per-project seen maps outside maxDiscoveryRedirectEntries.
		if child == nil {
			continue
		}
		if child.configPath != ref.path && st.seen[child.configPath] {
			continue
		}
		st.seen[ref.path] = true
		if child.configPath != ref.path {
			st.seen[child.configPath] = true
		}
		if d.redirectEntries >= maxDiscoveryRedirectEntries {
			// Budget reached: disable redirect caching for this config and
			// finish draining the current DFS stack without retaining entries.
			d.redirectEntries -= len(st.visited)
			r.redirect = nil
			r.redirectDisabled = true
			if _, ok := child.fileNames[file]; ok {
				return true, nil
			}
			if child.refs != nil {
				st.stack = append(st.stack, redirectFrame{entry: child, index: 0})
			}
			return d.drainRedirectSearch(st.stack, st.seen, file)
		}
		st.visited = append(st.visited, child)
		d.redirectEntries++
		if child.refs != nil {
			st.stack = append(st.stack, redirectFrame{entry: child, index: 0})
		}
		if _, ok := child.fileNames[file]; ok {
			for len(st.stack) > 0 {
				last := &st.stack[len(st.stack)-1]
				if last.index == 0 || last.index < len(last.entry.rawRefPaths()) {
					break
				}
				st.stack = st.stack[:len(st.stack)-1]
			}
			if len(st.stack) == 0 {
				st.complete = true
				st.stack = nil
				st.seen = nil
			}
			return true, nil
		}
	}
	st.complete = true
	st.stack = nil
	st.seen = nil
	return false, nil
}

func (d *TypeScriptProjectDiscovery) drainRedirectSearch(
	stack []redirectFrame,
	seen map[tspath.Path]bool,
	file tspath.Path,
) (bool, error) {
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		rawRefs := top.entry.rawRefPaths()
		if top.index >= len(rawRefs) {
			stack = stack[:len(stack)-1]
			continue
		}
		ref := top.entry.refAt(d, top.index)
		if seen[ref.path] {
			top.index++
			continue
		}
		child, err := d.loadRef(rawRefs[top.index], ref)
		if err != nil {
			return false, err
		}
		top.index++
		if child == nil {
			continue
		}
		if child.configPath != ref.path && seen[child.configPath] {
			continue
		}
		if _, ok := child.fileNames[file]; ok {
			return true, nil
		}
		if child.refs != nil {
			seen[ref.path] = true
			if child.configPath != ref.path {
				seen[child.configPath] = true
			}
			stack = append(stack, redirectFrame{entry: child, index: 0})
		}
	}
	return false, nil
}

func (d *TypeScriptProjectDiscovery) resetRedirectStack() {
	clear(d.redirectStack[:cap(d.redirectStack)])
	d.redirectStack = d.redirectStack[:0]
}

func (d *TypeScriptProjectDiscovery) isReferencedRootUncached(entry *configEntry, file tspath.Path) (bool, error) {
	if d.redirectSeenGen == nil {
		d.redirectSeenGen = make(map[tspath.Path]uint32)
	} else if d.redirectGen == math.MaxUint32 {
		clear(d.redirectSeenGen)
		d.redirectGen = 0
	}
	d.redirectGen++
	gen := d.redirectGen
	d.redirectSeenGen[entry.configPath] = gen
	if entry.refs.lookupPath != entry.configPath {
		d.redirectSeenGen[entry.refs.lookupPath] = gen
	}
	d.redirectStack = append(d.redirectStack[:0], redirectFrame{entry: entry, index: 0})
	defer d.resetRedirectStack()
	for len(d.redirectStack) > 0 {
		top := &d.redirectStack[len(d.redirectStack)-1]
		rawRefs := top.entry.rawRefPaths()
		if top.index >= len(rawRefs) {
			d.redirectStack = d.redirectStack[:len(d.redirectStack)-1]
			continue
		}
		ref := top.entry.refAt(d, top.index)
		if d.redirectSeenGen[ref.path] == gen {
			top.index++
			continue
		}
		child, err := d.loadRef(rawRefs[top.index], ref)
		if err != nil {
			return false, err
		}
		top.index++
		if child == nil {
			continue
		}
		if child.configPath != ref.path && d.redirectSeenGen[child.configPath] == gen {
			continue
		}
		if _, ok := child.fileNames[file]; ok {
			return true, nil
		}
		if child.refs != nil {
			d.redirectSeenGen[ref.path] = gen
			if child.configPath != ref.path {
				d.redirectSeenGen[child.configPath] = gen
			}
			d.redirectStack = append(d.redirectStack, redirectFrame{entry: child, index: 0})
		}
	}
	return false, nil
}

type typeScriptProjectSearch struct {
	discovery *TypeScriptProjectDiscovery
	file      tspath.Path
	fallback  *tsoptions.ParsedCommandLine
	gen       uint32
}

func (q *typeScriptProjectSearch) ensureGen() {
	if q.gen != 0 {
		return
	}
	d := q.discovery
	if d.seenGen == nil {
		d.seenGen = make(map[tspath.Path]uint32)
	} else if d.searchGen == math.MaxUint32 {
		clear(d.seenGen)
		d.searchGen = 0
	}
	d.searchGen++
	q.gen = d.searchGen
}

func (q *typeScriptProjectSearch) isSeen(path tspath.Path) bool {
	return q.gen != 0 && q.discovery.seenGen[path] == q.gen
}

func (q *typeScriptProjectSearch) markSeen(path tspath.Path) {
	q.ensureGen()
	q.discovery.seenGen[path] = q.gen
}

func (q *typeScriptProjectSearch) direct(cfg *configEntry) (*tsoptions.ParsedCommandLine, error) {
	if _, included := cfg.fileNames[q.file]; !included {
		return nil, nil //nolint:nilnil // Imports do not make a lint target a configured root.
	}
	if cfg.flags&configDisableRedirect != 0 {
		return cfg.parsed, nil
	}
	redirected, err := q.discovery.isReferencedRoot(cfg, q.file)
	if err != nil {
		return nil, err
	}
	if !redirected {
		return cfg.parsed, nil
	}
	if q.fallback == nil {
		q.fallback = cfg.parsed
	}
	return nil, nil //nolint:nilnil // Prefer the referenced project's own compiler options.
}

// references checks immediate references before their subtrees. Disabling
// reference loading stops this discovery step regardless of earlier requests;
// discovery does not reproduce the TypeScript server's warm-project state.
func (q *typeScriptProjectSearch) references(cfg *configEntry) (*tsoptions.ParsedCommandLine, error) {
	if cfg.flags&configDisableRefLoad != 0 || cfg.refs == nil {
		return nil, nil //nolint:nilnil // The config explicitly stops reference discovery or has no references.
	}
	rawRefs := cfg.rawRefPaths()
	var childrenBuf [8]*configEntry
	children := childrenBuf[:0]
	for i, name := range rawRefs {
		ref := cfg.refAt(q.discovery, i)
		if q.isSeen(ref.path) {
			continue
		}
		child, err := q.discovery.loadRef(name, ref)
		if err != nil {
			return nil, err
		}
		if child == nil {
			continue
		}
		if child.configPath != ref.path && q.isSeen(child.configPath) {
			continue
		}
		// Leaf configs cannot recurse or form cycles, and direct(child) is a
		// side-effect-free membership check; skip recording them in seenGen.
		if child.refs != nil {
			q.markSeen(ref.path)
			if child.configPath != ref.path {
				q.markSeen(child.configPath)
			}
		}
		if selected, err := q.direct(child); err != nil || selected != nil {
			return selected, err
		}
		if child.flags&configDisableRefLoad == 0 && child.refs != nil {
			children = append(children, child)
		}
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
// As in TypeScript's project service, a target outside the root searches its
// own ancestor tree. Directory aliases into the root retain the equivalent
// boundary in the target's lexical path.
func (d *TypeScriptProjectDiscovery) Find(fileName, rootDirectory string) (*tsoptions.ParsedCommandLine, error) {
	fileName = tspath.NormalizePath(fileName)
	query := typeScriptProjectSearch{
		discovery: d,
		file:      d.path(fileName),
	}
	directory := tspath.GetDirectoryPath(fileName)
	boundaryExactID := d.searchBoundary(directory, rootDirectory)
	for dir := d.dirEntry(directory); ; dir = d.parentDir(dir) {
		for i := range typeScriptConfigFileNames {
			cfg, err := d.dirSlotConfig(dir, i)
			if err != nil {
				return nil, err
			}
			if cfg == nil {
				continue
			}
			if selected, err := query.direct(cfg); err != nil || selected != nil {
				return selected, err
			}
			if cfg.refs != nil {
				query.markSeen(cfg.configPath)
				if cfg.refs.lookupPath != cfg.configPath {
					query.markSeen(cfg.refs.lookupPath)
				}
				if selected, err := query.references(cfg); err != nil || selected != nil {
					return selected, err
				}
			}
			if cfg.flags&configDisableSolution != 0 {
				return query.fallback, nil
			}
		}
		if dir.exactID == boundaryExactID || dir.isNodeModules || dir.atRoot {
			return query.fallback, nil
		}
	}
}
