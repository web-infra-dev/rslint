package loader

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// buildContext owns the services shared by every TypeScript Program
// constructed during one CLI invocation or one API request. Its lifetime is
// deliberately shorter than an LSP session: editor projects have their own
// versioned file/config caches and must not inherit invocation snapshots.
//
// The context keeps three concerns together at the Program-construction
// boundary while leaving their implementations independent:
//   - the final request-specific VFS view;
//   - source text and AST reuse through ParseCache;
//   - successful package.json/registered-tsconfig reads, extended-config
//     parses, and resolved project references through programMetadataFS,
//     programExtendedConfigCache, and programBuildCompilerHost.
//
// RSLINT_DISABLE_PROGRAM_METADATA_CACHE is an undocumented diagnostic escape
// hatch. When set, the Program VFS is not wrapped by either metadata snapshots
// or concurrent-query coalescing, and tsconfig extends and project-reference
// parsing use the upstream no-cache paths. ParseCache keeps its own independent
// escape hatch.
type buildContext struct {
	fs                  vfs.FS
	parseCache          *utils.ParseCache
	metadataFS          *programMetadataFS
	extendedConfigCache *programExtendedConfigCache
	parallelFSOnce      sync.Once
	parallelFS          atomic.Pointer[parallelProgramFS]
}

// newBuildContext creates a context for exactly one final VFS view.
// Callers must apply overlays and canonical-path wrappers before constructing
// the context so cached bytes can never bypass request-local file contents.
func newBuildContext(fs vfs.FS) *buildContext {
	context := &buildContext{
		fs:         fs,
		parseCache: utils.NewParseCache(),
	}
	if os.Getenv("RSLINT_DISABLE_PROGRAM_METADATA_CACHE") != "" {
		return context
	}

	metadataFS := newProgramMetadataFS(fs)
	context.fs = metadataFS
	context.metadataFS = metadataFS
	context.extendedConfigCache = &programExtendedConfigCache{
		metadataFS: metadataFS,
	}
	return context
}

// FS returns the logical VFS view used by non-compiler consumers for this
// invocation/request. Compiler hosts may use a context-owned derived view that
// delegates to this one without changing its path or content semantics.
func (c *buildContext) FS() vfs.FS {
	return c.fs
}

// enableConcurrentProgramQueries switches future compiler hosts in this
// context to one shared VFS view that coalesces their same-path Realpath
// queries. It must be called only after choosing a genuinely parallel Program
// build; FS and every non-Program consumer retain the original view.
func (c *buildContext) enableConcurrentProgramQueries() {
	if c.metadataFS == nil {
		return
	}
	c.parallelFSOnce.Do(func() {
		c.parallelFS.Store(newParallelProgramFS(c.fs))
	})
}

func (c *buildContext) compilerFS() vfs.FS {
	if fs := c.parallelFS.Load(); fs != nil {
		return fs
	}
	return c.fs
}

// newCompilerHostWithCache returns a host wired to every cache owned by the context.
// Project-reference configs are registered and cached by programBuildCompilerHost;
// extended configs are registered by programExtendedConfigCache.
func (c *buildContext) newCompilerHostWithCache(cwd string) compiler.CompilerHost {
	host := c.newCompilerHost(cwd)
	return utils.WithParseCache(host, c.parseCache)
}

// newTransientCompilerHost shares source snapshots with Program construction
// without publishing its bound ASTs into the project-owned parse cache.
func (c *buildContext) newTransientCompilerHost(cwd string) compiler.CompilerHost {
	host := c.newCompilerHost(cwd)
	return utils.WithSourceSnapshots(host, c.parseCache)
}

func (c *buildContext) newCompilerHost(cwd string) compiler.CompilerHost {
	host := compiler.NewCompilerHost(cwd, c.compilerFS(), bundled.LibPath(), c.extendedConfigCacheInterface(), nil, nil)
	if c.metadataFS != nil {
		host = &programBuildCompilerHost{
			CompilerHost: host,
			context:      c,
		}
	}
	return host
}

// createProjectProgram creates a tsconfig-backed Program and preserves the
// lenient syntactic-error behavior used by the CLI and API.
func (c *buildContext) createProjectProgram(singleThreaded bool, cwd string, tsconfigPath string) (*compiler.Program, error) {
	config, err := c.parseConfig(cwd, tsconfigPath)
	if err != nil {
		return nil, err
	}
	return utils.CreateProgramFromParsedConfigLenient(singleThreaded, config, c.newCompilerHostWithCache(cwd))
}

// parseConfig reads metadata without binding the source cache to a filesystem.
// Program construction chooses its serial or parallel view before doing that.
func (c *buildContext) parseConfig(cwd string, tsconfigPath string) (*tsoptions.ParsedCommandLine, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !c.compilerFS().FileExists(resolvedConfigPath) {
		return nil, fmt.Errorf("couldn't read tsconfig at %v", resolvedConfigPath)
	}
	c.registerTSConfig(resolvedConfigPath)

	host := c.newCompilerHost(cwd)
	config, _ := tsoptions.GetParsedCommandLineOfConfigFile(
		tsconfigPath,
		&core.CompilerOptions{},
		nil,
		host,
		c.extendedConfigCacheInterface(),
	)
	return config, nil
}

// createCompatibilityProgram creates a source-only compiler Program using the
// same source/metadata view as configured projects. The loader adapts it to the
// same rslint Program facade before returning a result.
func (c *buildContext) createCompatibilityProgram(
	singleThreaded bool,
	cwd string,
	compilerOptions *core.CompilerOptions,
	rootFileNames []string,
) (*compiler.Program, error) {
	return utils.CreateProgramFromOptionsLenient(
		singleThreaded,
		compilerOptions,
		rootFileNames,
		c.newCompilerHostWithCache(cwd),
	)
}

// retainOnlySourceFiles bounds the append-mostly parsed-AST cache to SourceFile
// objects still retained by the supplied Programs.
func (c *buildContext) retainOnlySourceFiles(programs []*compiler.Program) {
	c.parseCache.RetainOnly(programs)
}

// invalidateSourceSnapshots begins a fresh source generation before Programs
// are rebuilt after a fix write. Metadata generations are independent because
// CLI fixes never target package.json or tsconfig files.
func (c *buildContext) invalidateSourceSnapshots() {
	c.parseCache.InvalidateSourceSnapshots()
}

func (c *buildContext) registerTSConfig(fileName string) {
	if c.metadataFS != nil {
		c.metadataFS.registerTSConfig(fileName)
	}
}

// extendedConfigCacheInterface avoids converting a nil concrete pointer into a
// non-nil interface when the diagnostic escape hatch disables metadata caches.
func (c *buildContext) extendedConfigCacheInterface() tsoptions.ExtendedConfigCache {
	if c.extendedConfigCache == nil {
		return nil
	}
	return c.extendedConfigCache
}

// programBuildCompilerHost registers project-reference configs with the
// metadata snapshot VFS and caches resolved command lines for absolute config
// file names across Programs in the same buildContext generation.
//
// Resolution is strictly non-reentrant (GetParsedCommandLineOfConfigFilePath
// takes a ParseConfigHost with no GetResolvedProjectReference method; recursive
// references are walked later by the compiler's project-reference parser), so
// cold misses coalesce under a panic-safe single-flight entry. Each non-nil
// ParsedCommandLine finalizes ParseInputOutputNames() before returning to
// callers or coalesced waiters so source/output maps are immutable across
// concurrent readers. CommonSourceDirectory() remains deferred when
// ParseInputOutputNames() does not trigger it (e.g. in-place emit without
// outDir/declarationDir, or configs whose only inputs are declarations/JSON).
// When a later lazy CommonSourceDirectory() call could still append TS6059 to
// Errors (mayMutateErrorsOnCommonSourceDirectory), the reference bypasses both
// permanent caching and single-flight sharing so each caller receives its own
// private ParsedCommandLine and concurrent CommonSourceDirectory() /
// GetConfigFileParsingDiagnostics() calls remain race-free without eager
// TS6059 diagnostics. Missing/inaccessible configs (nil) and configs with
// unreadable or missing file lookups at any depth bypass permanent caching so
// subsequent lookups retry after access is restored; cached references that
// observed missing package.json probes revalidate those paths purely through
// the VFS on hit and defer committing staged extended configs until clean.
// Relative config paths also bypass caching.
type programBuildCompilerHost struct {
	compiler.CompilerHost
	context *buildContext
}

type programProjectReferenceKey struct {
	fileName string
	path     tspath.Path
}

type programProjectReferenceGeneration struct {
	entries sync.Map // programProjectReferenceKey -> *projectReferenceFlight | *cachedProjectReference
}

type missingPackageProbes struct {
	packageJSONs []string
	packageDirs  []string
}

func (p missingPackageProbes) empty() bool {
	return len(p.packageJSONs) == 0 && len(p.packageDirs) == 0
}

func (p *missingPackageProbes) merge(other missingPackageProbes) {
	p.packageJSONs = appendUniquePaths(p.packageJSONs, other.packageJSONs)
	p.packageDirs = appendUniquePaths(p.packageDirs, other.packageDirs)
}

type packageProbeCache struct {
	files map[string]bool
	dirs  map[string]bool
}

func (c *packageProbeCache) recordMissingFile(path string) {
	if c == nil {
		return
	}
	if c.files == nil {
		c.files = make(map[string]bool)
	}
	c.files[path] = false
}

func (c *packageProbeCache) recordMissingDir(path string) {
	if c == nil {
		return
	}
	if c.dirs == nil {
		c.dirs = make(map[string]bool)
	}
	c.dirs[path] = false
}

func (c *packageProbeCache) recordPresentFile(path string) {
	if c == nil || c.files == nil {
		return
	}
	if _, seen := c.files[path]; seen {
		c.files[path] = true
	}
}

func (c *packageProbeCache) recordPresentDir(path string) {
	if c == nil || c.dirs == nil {
		return
	}
	if _, seen := c.dirs[path]; seen {
		c.dirs[path] = true
	}
}

func (c *packageProbeCache) fileExists(fsys vfs.FS, path string) bool {
	if c != nil {
		if exists, ok := c.files[path]; ok {
			return exists
		}
	}
	exists := fsys.FileExists(path)
	if c != nil {
		if c.files == nil {
			c.files = make(map[string]bool)
		}
		c.files[path] = exists
	}
	return exists
}

func (c *packageProbeCache) directoryExists(fsys vfs.FS, path string) bool {
	if c != nil {
		if exists, ok := c.dirs[path]; ok {
			return exists
		}
	}
	exists := fsys.DirectoryExists(path)
	if c != nil {
		if c.dirs == nil {
			c.dirs = make(map[string]bool)
		}
		c.dirs[path] = exists
	}
	return exists
}

type cachedProjectReference struct {
	commandLine   *tsoptions.ParsedCommandLine
	missingProbes missingPackageProbes
}

type projectReferenceFlight struct {
	ready       chan struct{}
	commandLine *tsoptions.ParsedCommandLine
	completed   bool
	unshared    bool
}

func (h *programBuildCompilerHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	if !tspath.PathIsAbsolute(fileName) {
		commandLine, _, _, _ := h.parseProjectReference(fileName, path)
		return commandLine
	}

	generation := h.context.metadataFS.currentProjectReferenceGeneration()
	key := programProjectReferenceKey{
		fileName: fileName,
		path:     path,
	}

	for {
		existing, ok := generation.entries.Load(key)
		if !ok {
			flight := &projectReferenceFlight{ready: make(chan struct{})}
			actual, loaded := generation.entries.LoadOrStore(key, flight)
			if !loaded {
				return h.resolveAndPublishProjectReference(generation, key, flight)
			}
			existing = actual
		}

		switch entry := existing.(type) {
		case *cachedProjectReference:
			if hasRecoveredPackageProbe(h.FS(), entry.missingProbes, nil) {
				generation.entries.CompareAndDelete(key, entry)
				continue
			}
			return entry.commandLine
		case *projectReferenceFlight:
			<-entry.ready
			if entry.unshared {
				return h.parseUnsharedProjectReference(fileName, path)
			}
			if entry.completed {
				return entry.commandLine
			}
		default:
			panic("program project reference cache contains an unexpected value")
		}
	}
}

func hasRecoveredPackageProbe(fsys vfs.FS, missing missingPackageProbes, cache *packageProbeCache) bool {
	for _, pkg := range missing.packageJSONs {
		if cache.fileExists(fsys, pkg) {
			return true
		}
	}
	for _, dir := range missing.packageDirs {
		if cache.directoryExists(fsys, dir) {
			return true
		}
	}
	return false
}

// isNodeModulesPackageDir reports whether path names a direct package root
// directory underneath a node_modules folder (.../node_modules/<pkg> or
// .../node_modules/@<scope>/<pkg>), excluding bare node_modules folders and
// deeper nested paths inside a package (such as .../node_modules/@org/pkg/base.json,
// which TypeScript probes with DirectoryExists during normal extends resolution).
// Tracking missing package root directories catches temporarily inaccessible
// nearer package directories or symlinks without adding probes when packages
// have no local node_modules directory (when a package has a local node_modules
// folder alongside a hoisted preset, its absent local package dir is revalidated
// on cache hit; bare node_modules symlinks and deeper nested paths inside an
// accessible package remain outside this pure-VFS filter to avoid probing every
// healthy hit).
func isNodeModulesPackageDir(path string) bool {
	trimmed := strings.TrimRight(path, `/\`)
	const segment = "node_modules"
	lastRest := -1
	for offset := 0; offset < len(trimmed); {
		idx := strings.Index(trimmed[offset:], segment)
		if idx < 0 {
			break
		}
		pos := offset + idx
		end := pos + len(segment)
		leftOK := pos == 0 || trimmed[pos-1] == '/' || trimmed[pos-1] == '\\'
		rightOK := end < len(trimmed) && (trimmed[end] == '/' || trimmed[end] == '\\')
		if leftOK && rightOK && end+1 < len(trimmed) {
			lastRest = end + 1
		}
		offset = end
	}
	if lastRest < 0 {
		return false
	}
	rest := trimmed[lastRest:]
	sep := strings.IndexAny(rest, `/\`)
	if rest[0] == '@' {
		// Scoped package: require exactly "@scope/<pkg>" with non-empty parts.
		if sep <= 1 || sep+1 >= len(rest) {
			return false
		}
		return !strings.ContainsAny(rest[sep+1:], `/\`)
	}
	// Unscoped package: require exactly "<pkg>" with no deeper subpath.
	return sep < 0
}

type transientReadTrackerFS struct {
	vfs.FS
	hadTransientRead bool
	missingProbes    missingPackageProbes
	ownedProbes      packageProbeCache
	probes           *packageProbeCache
}

func (f *transientReadTrackerFS) probeCache() *packageProbeCache {
	if f.probes != nil {
		return f.probes
	}
	return &f.ownedProbes
}

func (f *transientReadTrackerFS) ReadFile(path string) (string, bool) {
	content, ok := f.FS.ReadFile(path)
	if !ok {
		f.hadTransientRead = true
	}
	return content, ok
}

func (f *transientReadTrackerFS) FileExists(path string) bool {
	if f.FS.FileExists(path) {
		f.probeCache().recordPresentFile(path)
		return true
	}
	if isPackageJSONPath(path) {
		if !slices.Contains(f.missingProbes.packageJSONs, path) {
			f.missingProbes.packageJSONs = append(f.missingProbes.packageJSONs, path)
		}
		f.probeCache().recordMissingFile(path)
	}
	return false
}

func (f *transientReadTrackerFS) DirectoryExists(path string) bool {
	if f.FS.DirectoryExists(path) {
		f.probeCache().recordPresentDir(path)
		return true
	}
	if isNodeModulesPackageDir(path) {
		if !slices.Contains(f.missingProbes.packageDirs, path) {
			f.missingProbes.packageDirs = append(f.missingProbes.packageDirs, path)
		}
		f.probeCache().recordMissingDir(path)
	}
	return false
}

func appendUniquePaths(dst []string, src []string) []string {
	for _, p := range src {
		if !slices.Contains(dst, p) {
			dst = append(dst, p)
		}
	}
	return dst
}

type transientReadTrackerHost struct {
	tsoptions.ParseConfigHost
	fs *transientReadTrackerFS
}

func (h *transientReadTrackerHost) FS() vfs.FS {
	return h.fs
}

// stagedExtendedConfigCache buffers cold extended-config misses locally during
// parseProjectReference until the project reference resolves cleanly without
// transient reads or lookup diagnostics. Each entry retains the transitive
// missing package.json and node_modules package-directory probes observed
// during its parse so shared hits can revalidate cleanly against the live VFS
// without blocking publication of healthy nested presets (e.g.
// presets/browser.json without presets/package.json).
type stagedExtendedConfigCache struct {
	metadataFS *programMetadataFS
	generation *programExtendedConfigGeneration
	tracker    *transientReadTrackerFS
	staged     map[string]*cachedExtendedConfig
}

func (c *buildContext) newStagedExtendedConfigCache(tracker *transientReadTrackerFS) *stagedExtendedConfigCache {
	if c.extendedConfigCache == nil || c.metadataFS == nil {
		return nil
	}
	return &stagedExtendedConfigCache{
		metadataFS: c.metadataFS,
		generation: c.metadataFS.currentExtendedConfigGeneration(),
		tracker:    tracker,
	}
}

func (s *stagedExtendedConfigCache) cacheInterface() tsoptions.ExtendedConfigCache {
	if s == nil {
		return nil
	}
	return s
}

func (s *stagedExtendedConfigCache) GetExtendedConfig(
	fileName string,
	path tspath.Path,
	resolutionStack []tspath.Path,
	host tsoptions.ParseConfigHost,
) *tsoptions.ExtendedConfigCacheEntry {
	s.metadataFS.registerTSConfig(fileName)
	if cached, ok := s.staged[fileName]; ok {
		s.tracker.missingProbes.merge(cached.missingProbes)
		return cached.entry
	}
	if raw, ok := s.generation.entries.Load(fileName); ok {
		cached := mustExtendedConfigCacheEntry(raw)
		if cached.missingProbes.empty() || !hasRecoveredPackageProbe(s.tracker.FS, cached.missingProbes, s.tracker.probeCache()) {
			s.tracker.missingProbes.merge(cached.missingProbes)
			return cached.entry
		}
		s.generation.entries.CompareAndDelete(fileName, cached)
	}
	outerMissing := s.tracker.missingProbes
	s.tracker.missingProbes = missingPackageProbes{}
	entry := tsoptions.ParseExtendedConfig(fileName, path, resolutionStack, host, s)
	entryMissing := s.tracker.missingProbes
	s.tracker.missingProbes = outerMissing
	s.tracker.missingProbes.merge(entryMissing)
	if s.staged == nil {
		s.staged = make(map[string]*cachedExtendedConfig)
	}
	s.staged[fileName] = &cachedExtendedConfig{
		entry:         entry,
		missingProbes: entryMissing,
	}
	return entry
}

func (s *stagedExtendedConfigCache) commit() {
	if s == nil {
		return
	}
	probes := s.tracker.probeCache()
	for fileName, entry := range s.staged {
		publishExtendedConfig(s.generation, s.tracker.FS, fileName, entry, probes)
	}
}

func (h *programBuildCompilerHost) parseProjectReference(
	fileName string,
	path tspath.Path,
) (*tsoptions.ParsedCommandLine, bool, missingPackageProbes, *stagedExtendedConfigCache) {
	h.context.registerTSConfig(fileName)
	tracker := &transientReadTrackerFS{FS: h.FS()}
	staged := h.context.newStagedExtendedConfigCache(tracker)
	commandLine, _ := tsoptions.GetParsedCommandLineOfConfigFilePath(
		fileName,
		path,
		nil,
		nil,
		&transientReadTrackerHost{ParseConfigHost: h, fs: tracker},
		staged.cacheInterface(),
	)
	return commandLine, tracker.hadTransientRead, tracker.missingProbes, staged
}

func (h *programBuildCompilerHost) parseUnsharedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	commandLine, hadTransientRead, _, staged := h.parseProjectReference(fileName, path)
	if commandLine != nil {
		commandLine.ParseInputOutputNames()
		if !hadTransientRead && !hasTransientConfigLookupDiagnostic(commandLine) {
			staged.commit()
		}
	}
	return commandLine
}

func (h *programBuildCompilerHost) resolveAndPublishProjectReference(
	generation *programProjectReferenceGeneration,
	key programProjectReferenceKey,
	flight *projectReferenceFlight,
) *tsoptions.ParsedCommandLine {
	closed := false
	defer func() {
		if !closed {
			// A panicking underlying VFS must not strand concurrent waiters.
			generation.entries.CompareAndDelete(key, flight)
			close(flight.ready)
		}
	}()

	candidate, hadTransientRead, missingProbes, staged := h.parseProjectReference(key.fileName, key.path)
	if candidate == nil {
		// Never permanently cache or coalesce nil (an inaccessible file or
		// directory makes ReadFile/FileExists fail transiently): return nil to
		// this caller while concurrent waiters retry independently in parallel
		// instead of re-serializing through successive flights.
		flight.unshared = true
		generation.entries.CompareAndDelete(key, flight)
		closed = true
		close(flight.ready)
		return nil
	}

	// Finalize lazy source/output maps (and CommonSourceDirectory when outDir
	// or declarationDir is configured with a non-declaration, non-JSON input
	// file) before publishing so concurrent readers never race on
	// ParseInputOutputNames(). Avoid eagerly calling CommonSourceDirectory()
	// otherwise: it scans every input file and appends TS6059 to
	// candidate.Errors during loading, diverging from uncached tsgo loading.
	candidate.ParseInputOutputNames()

	if hadTransientRead || hasTransientConfigLookupDiagnostic(candidate) {
		// Never permanently cache or coalesce configs whose extends/file lookup
		// failed transiently (e.g. an unreadable package.json or nested extends
		// leaving outDir missing): return candidate to this caller while
		// concurrent waiters retry independently in parallel instead of
		// serializing persistent lookup failures across successive rounds.
		flight.unshared = true
		generation.entries.CompareAndDelete(key, flight)
		closed = true
		close(flight.ready)
		return candidate
	}

	staged.commit()

	if mayMutateErrorsOnCommonSourceDirectory(candidate) {
		// CommonSourceDirectory() was not finalized by ParseInputOutputNames()
		// and a later lazy call will append TS6059 to candidate.Errors, racing
		// with concurrent GetConfigFileParsingDiagnostics() readers. Neither
		// cache nor share this ParsedCommandLine across concurrent waiters.
		flight.unshared = true
		flight.completed = true
		generation.entries.CompareAndDelete(key, flight)
		closed = true
		close(flight.ready)
		return candidate
	}

	cached := &cachedProjectReference{
		commandLine:   candidate,
		missingProbes: missingProbes,
	}
	if missingProbes.empty() {
		flight.commandLine = candidate
		flight.completed = true
	}
	// When missingProbes is non-empty, leave flight.completed false so any
	// concurrent waiter re-checks generation.entries and revalidates
	// cached.missingProbes against the live VFS before reusing candidate.
	generation.entries.CompareAndSwap(key, flight, cached)
	closed = true
	close(flight.ready)
	return candidate
}

// mayMutateErrorsOnCommonSourceDirectory reports whether a later lazy call to
// cmd.CommonSourceDirectory() can still run checkSourceFilesBelongToPath and
// append TS6059 to cmd.Errors. ParseInputOutputNames() must already have run;
// when outDir or declarationDir is set and cmd contains any non-declaration,
// non-JSON source file, ParseInputOutputNames() has already finalized
// CommonSourceDirectory() before publication.
func mayMutateErrorsOnCommonSourceDirectory(cmd *tsoptions.ParsedCommandLine) bool {
	if cmd == nil {
		return false
	}
	options := cmd.CompilerOptions()
	if options == nil {
		return false
	}
	rootDirectory := options.RootDir
	if rootDirectory == "" {
		if options.ConfigFilePath == "" {
			return false
		}
		rootDirectory = tspath.GetDirectoryPath(options.ConfigFilePath)
	}
	hasOutputDir := options.DeclarationDir != "" || options.OutDir != ""
	useCaseSensitive := cmd.UseCaseSensitiveFileNames()
	compareOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cmd.GetCurrentDirectory(),
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	cleanRoot, rootLength, hasCleanRoot := prepareCleanAbsoluteRoot(rootDirectory)
	hasOutsideFile := false
	for _, file := range cmd.FileNames() {
		if tspath.IsDeclarationFileName(file) {
			continue
		}
		if hasOutputDir && !tspath.FileExtensionIs(file, tspath.ExtensionJson) {
			// ParseInputOutputNames() already invoked CommonSourceDirectory()
			// for this non-declaration, non-JSON input before publication.
			return false
		}
		if options.NoEmitForJsFiles.IsTrue() && tspath.HasJSFileExtension(file) {
			continue
		}
		if hasCleanRoot && containsCleanAbsolutePath(cleanRoot, rootLength, file, useCaseSensitive) {
			continue
		}
		if !tspath.ContainsPath(rootDirectory, file, compareOptions) {
			hasOutsideFile = true
			if !hasOutputDir {
				return true
			}
		}
	}
	return hasOutsideFile
}

// prepareCleanAbsoluteRoot validates and normalizes root once before scanning
// cmd.FileNames() so the zero-alloc containment fast path only runs on rooted
// forward-slash disk paths without relative or empty segments.
func prepareCleanAbsoluteRoot(root string) (cleanRoot string, rootLength int, ok bool) {
	rootLength = tspath.GetEncodedRootLength(root)
	if rootLength <= 0 {
		return "", 0, false
	}
	if len(root) > rootLength && root[len(root)-1] == '/' {
		root = root[:len(root)-1]
	}
	if len(root) == rootLength && (root[len(root)-1] != '/' || root == "//") {
		return "", 0, false
	}
	if strings.IndexByte(root, '\\') >= 0 {
		return "", 0, false
	}
	if len(root) > rootLength && !isCleanRelativePath(root[rootLength:]) {
		return "", 0, false
	}
	return root, rootLength, true
}

// containsCleanAbsolutePath is a zero-allocation fast path for
// tspath.ContainsPath(cleanRoot, file, opts). Returning true guarantees that
// tspath.ContainsPath returns true; returning false asks the caller to fall
// back to tspath.ContainsPath.
func containsCleanAbsolutePath(cleanRoot string, rootLength int, file string, useCaseSensitive bool) bool {
	if len(file) < len(cleanRoot) {
		return false
	}
	if file[:len(cleanRoot)] != cleanRoot {
		if useCaseSensitive {
			return false
		}
		for i := range len(cleanRoot) {
			b := file[i]
			c := cleanRoot[i]
			if b == c {
				continue
			}
			if b >= 0x80 || c >= 0x80 || asciiToLower(b) != asciiToLower(c) {
				return false
			}
		}
	}
	rest := file[len(cleanRoot):]
	if rest == "" {
		return true
	}
	if len(cleanRoot) > rootLength {
		if rest[0] != '/' {
			return false
		}
		rest = rest[1:]
	}
	return isCleanRelativePath(rest)
}

func isCleanRelativePath(s string) bool {
	if s == "" {
		return false
	}
	segStart := 0
	for i := range len(s) + 1 {
		if i < len(s) {
			if s[i] == '\\' {
				return false
			}
			if s[i] != '/' {
				continue
			}
		}
		segLen := i - segStart
		if segLen == 0 {
			return false
		}
		if segLen == 1 && s[segStart] == '.' {
			return false
		}
		if segLen == 2 && s[segStart] == '.' && s[segStart+1] == '.' {
			return false
		}
		segStart = i + 1
	}
	return true
}

func asciiToLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func hasTransientConfigLookupDiagnostic(cmd *tsoptions.ParsedCommandLine) bool {
	for _, diag := range cmd.GetConfigFileParsingDiagnostics() {
		switch diag.Code() {
		case 5012, 5083, 6053: // TS5012/TS5083 (Cannot read file) / TS6053 (File not found)
			return true
		}
	}
	return false
}

// programExtendedConfigCache is request-local and append-only. Parsing happens
// outside sync.Map so recursively extended configs never hold one entry lock
// while requesting another (which can deadlock under adversarial cross-cycles).
// Concurrent misses may duplicate parsing, but LoadOrStore makes every caller
// use one immutable winning result. Raw config reads remain single-flight in
// programMetadataFS even in that rare case.
type programExtendedConfigCache struct {
	metadataFS *programMetadataFS
}

type programExtendedConfigGeneration struct {
	entries sync.Map // exact fileName -> *cachedExtendedConfig
}

type cachedExtendedConfig struct {
	entry         *tsoptions.ExtendedConfigCacheEntry
	missingProbes missingPackageProbes
}

var _ tsoptions.ExtendedConfigCache = (*programExtendedConfigCache)(nil)

func (c *programExtendedConfigCache) GetExtendedConfig(
	fileName string,
	path tspath.Path,
	resolutionStack []tspath.Path,
	host tsoptions.ParseConfigHost,
) *tsoptions.ExtendedConfigCacheEntry {
	c.metadataFS.registerTSConfig(fileName)
	generation := c.metadataFS.currentExtendedConfigGeneration()
	baseHost := host
	var outerTracker *transientReadTrackerFS
	var probes *packageProbeCache
	if tracked, ok := host.(*transientReadTrackerHost); ok {
		baseHost = tracked.ParseConfigHost
		outerTracker = tracked.fs
		probes = outerTracker.probeCache()
	}
	baseFS := baseHost.FS()

	if raw, ok := generation.entries.Load(fileName); ok {
		cached := mustExtendedConfigCacheEntry(raw)
		if cached.missingProbes.empty() || !hasRecoveredPackageProbe(baseFS, cached.missingProbes, probes) {
			if outerTracker != nil {
				outerTracker.missingProbes.merge(cached.missingProbes)
			}
			return cached.entry
		}
		generation.entries.CompareAndDelete(fileName, cached)
	}

	tracker := &transientReadTrackerFS{FS: baseFS, probes: probes}
	entry := tsoptions.ParseExtendedConfig(
		fileName,
		path,
		resolutionStack,
		&transientReadTrackerHost{ParseConfigHost: baseHost, fs: tracker},
		c,
	)
	if tracker.hadTransientRead {
		if outerTracker != nil {
			outerTracker.hadTransientRead = true
			outerTracker.missingProbes.merge(tracker.missingProbes)
		}
		return entry
	}
	published := publishExtendedConfig(generation, baseFS, fileName, &cachedExtendedConfig{
		entry:         entry,
		missingProbes: tracker.missingProbes,
	}, tracker.probeCache())
	if outerTracker != nil {
		outerTracker.missingProbes.merge(published.missingProbes)
	}
	return published.entry
}

// publishExtendedConfig installs candidate into generation and replaces any
// incumbent whose missing package probes have already recovered so concurrent
// recovery cannot resurrect or return stale extended configs. Later readers
// revalidate candidate.missingProbes against the live VFS on hit.
func publishExtendedConfig(
	generation *programExtendedConfigGeneration,
	fsys vfs.FS,
	fileName string,
	candidate *cachedExtendedConfig,
	probes *packageProbeCache,
) *cachedExtendedConfig {
	for {
		actual, loaded := generation.entries.LoadOrStore(fileName, candidate)
		if !loaded {
			return candidate
		}
		existing := mustExtendedConfigCacheEntry(actual)
		if existing.missingProbes.empty() || !hasRecoveredPackageProbe(fsys, existing.missingProbes, probes) {
			return existing
		}
		if generation.entries.CompareAndSwap(fileName, existing, candidate) {
			return candidate
		}
	}
}

func mustExtendedConfigCacheEntry(value any) *cachedExtendedConfig {
	entry, ok := value.(*cachedExtendedConfig)
	if !ok {
		panic("program extended config cache contains an unexpected value")
	}
	return entry
}

// programMetadataFS snapshots only successful reads of package.json and
// explicitly registered tsconfig files. Source files, .gitignore files, rslint
// configs, and arbitrary JSON remain delegated to the underlying VFS.
//
// Cache keys are the exact path strings received by ReadFile. Detection may
// inspect a cross-platform basename, but it never cleans, case-folds, resolves,
// or resolves the key to a real path; lexical, symlink, case, and overlay
// aliases therefore cannot be merged accidentally.
type programMetadataFS struct {
	vfs.FS
	configPaths                sync.Map // exact fileName -> struct{}
	generation                 atomic.Pointer[programMetadataGeneration]
	extendedConfigGeneration   atomic.Pointer[programExtendedConfigGeneration]
	projectReferenceGeneration atomic.Pointer[programProjectReferenceGeneration]
}

type programMetadataGeneration struct {
	entries sync.Map // exact fileName -> *programMetadataRead
}

type programMetadataRead struct {
	ready   chan struct{}
	content string
	ok      bool
}

var _ vfs.FS = (*programMetadataFS)(nil)

func newProgramMetadataFS(fs vfs.FS) *programMetadataFS {
	result := &programMetadataFS{FS: fs}
	result.generation.Store(&programMetadataGeneration{})
	result.extendedConfigGeneration.Store(&programExtendedConfigGeneration{})
	result.projectReferenceGeneration.Store(&programProjectReferenceGeneration{})
	return result
}

func (f *programMetadataFS) registerTSConfig(fileName string) {
	f.configPaths.Store(fileName, struct{}{})
}

func (f *programMetadataFS) shouldCache(fileName string) bool {
	if isPackageJSONPath(fileName) {
		return true
	}
	_, registered := f.configPaths.Load(fileName)
	return registered
}

func isPackageJSONPath(fileName string) bool {
	if separator := strings.LastIndexAny(fileName, `/\`); separator >= 0 {
		fileName = fileName[separator+1:]
	}
	return fileName == "package.json"
}

func (f *programMetadataFS) ReadFile(fileName string) (string, bool) {
	if !f.shouldCache(fileName) {
		return f.FS.ReadFile(fileName)
	}

	generation := f.currentGeneration()
	candidate := &programMetadataRead{ready: make(chan struct{})}
	actual, loaded := generation.entries.LoadOrStore(fileName, candidate)
	if loaded {
		read := mustProgramMetadataRead(actual)
		<-read.ready
		return read.content, read.ok
	}

	return f.readAndPublish(generation, fileName, candidate)
}

// ReadSourceBOM forwards to the wrapped VFS. Embedding vfs.FS promotes only
// that interface's methods, so this layer has to pass [BOMSource] through by
// hand or every overlay identity below it becomes invisible.
func (f *programMetadataFS) ReadSourceBOM(path string) (bool, error) {
	return utils.ReadSourceBOM(f.FS, path)
}

func mustProgramMetadataRead(value any) *programMetadataRead {
	read, ok := value.(*programMetadataRead)
	if !ok {
		panic("program metadata cache contains an unexpected value")
	}
	return read
}

func (f *programMetadataFS) readAndPublish(
	generation *programMetadataGeneration,
	fileName string,
	read *programMetadataRead,
) (content string, ok bool) {
	completed := false
	defer func() {
		if !completed {
			// A panicking underlying VFS must not strand concurrent waiters.
			generation.entries.CompareAndDelete(fileName, read)
			close(read.ready)
		}
	}()

	content, ok = f.FS.ReadFile(fileName)
	read.content = content
	read.ok = ok
	if !ok {
		// Failures are observations, not snapshots. A later call may recover
		// after an overlay or transient filesystem condition changes.
		generation.entries.CompareAndDelete(fileName, read)
	}
	close(read.ready)
	completed = true
	return content, ok
}

func (f *programMetadataFS) currentGeneration() *programMetadataGeneration {
	for {
		if generation := f.generation.Load(); generation != nil {
			return generation
		}
		generation := &programMetadataGeneration{}
		if f.generation.CompareAndSwap(nil, generation) {
			return generation
		}
	}
}

func (f *programMetadataFS) currentExtendedConfigGeneration() *programExtendedConfigGeneration {
	for {
		if generation := f.extendedConfigGeneration.Load(); generation != nil {
			return generation
		}
		generation := &programExtendedConfigGeneration{}
		if f.extendedConfigGeneration.CompareAndSwap(nil, generation) {
			return generation
		}
	}
}

func (f *programMetadataFS) currentProjectReferenceGeneration() *programProjectReferenceGeneration {
	for {
		if generation := f.projectReferenceGeneration.Load(); generation != nil {
			return generation
		}
		generation := &programProjectReferenceGeneration{}
		if f.projectReferenceGeneration.CompareAndSwap(nil, generation) {
			return generation
		}
	}
}

func (f *programMetadataFS) invalidate() {
	// Publish the raw-read generation first, then extended configs, then
	// resolved project references. A concurrent config parse that captures an
	// older derived generation can observe either raw generation, but its
	// result is discarded when that derived generation is replaced below.
	// Publishing in the opposite order could let an old raw read populate a
	// newer derived generation.
	f.generation.Store(&programMetadataGeneration{})
	f.extendedConfigGeneration.Store(&programExtendedConfigGeneration{})
	f.projectReferenceGeneration.Store(&programProjectReferenceGeneration{})
}

// Writes are not used by the current CLI/API Program pipeline, but delegating
// them without invalidation would make this VFS abstraction unsafe for future
// callers. Swap generations both before and after the mutation: an in-flight
// old read may complete, while every read begun after the write returns does
// so against a fresh generation.
func (f *programMetadataFS) WriteFile(path string, data string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.WriteFile(path, data)
}

func (f *programMetadataFS) AppendFile(path string, data string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.AppendFile(path, data)
}

func (f *programMetadataFS) Remove(path string) error {
	f.invalidate()
	defer f.invalidate()
	return f.FS.Remove(path)
}
