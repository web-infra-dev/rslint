package utils

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// ProgramBuildContext owns the services shared by every TypeScript Program
// constructed during one CLI invocation or one API request. Its lifetime is
// deliberately shorter than an LSP session: editor projects have their own
// versioned file/config caches and must not inherit invocation snapshots.
//
// The context keeps three concerns together at the Program-construction
// boundary while leaving their implementations independent:
//   - the final request-specific VFS view;
//   - source text and AST reuse through ParseCache;
//   - successful package.json/registered-tsconfig reads and extended-config
//     parses through programMetadataFS and programExtendedConfigCache.
//
// RSLINT_DISABLE_PROGRAM_METADATA_CACHE is an undocumented diagnostic escape
// hatch. When set, the VFS is not wrapped and tsconfig extends parsing uses the
// upstream no-cache path. ParseCache keeps its own independent escape hatch.
type ProgramBuildContext struct {
	fs                  vfs.FS
	parseCache          *ParseCache
	metadataFS          *programMetadataFS
	extendedConfigCache *programExtendedConfigCache
}

// NewProgramBuildContext creates a context for exactly one final VFS view.
// Callers must apply overlays and canonical-path wrappers before constructing
// the context so cached bytes can never bypass request-local file contents.
func NewProgramBuildContext(fs vfs.FS) *ProgramBuildContext {
	context := &ProgramBuildContext{
		fs:         fs,
		parseCache: NewParseCache(),
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

// FS returns the exact VFS view that must be used by config discovery, Program
// construction, target binding, and linting for this invocation/request.
func (c *ProgramBuildContext) FS() vfs.FS {
	return c.fs
}

// NewCompilerHost returns a host wired to every cache owned by the context.
// Project-reference configs are registered before ts-go reads them; extended
// configs are registered by programExtendedConfigCache.
func (c *ProgramBuildContext) NewCompilerHost(cwd string) compiler.CompilerHost {
	host := compiler.NewCompilerHost(cwd, c.fs, bundled.LibPath(), c.extendedConfigCacheInterface(), nil)
	if c.metadataFS != nil {
		host = &programBuildCompilerHost{
			CompilerHost: host,
			context:      c,
		}
	}
	return WithParseCache(host, c.parseCache)
}

// CreateProgram creates a tsconfig-backed Program and preserves the strict
// syntactic-error behavior of CreateProgram.
func (c *ProgramBuildContext) CreateProgram(singleThreaded bool, cwd string, tsconfigPath string) (*compiler.Program, error) {
	host, config, err := c.parseConfig(cwd, tsconfigPath)
	if err != nil {
		return nil, err
	}
	return createProgramFromConfig(singleThreaded, config, host)
}

// CreateProgramLenient creates a tsconfig-backed Program and preserves the
// lenient syntactic-error behavior used by the CLI and API.
func (c *ProgramBuildContext) CreateProgramLenient(singleThreaded bool, cwd string, tsconfigPath string) (*compiler.Program, error) {
	host, config, err := c.parseConfig(cwd, tsconfigPath)
	if err != nil {
		return nil, err
	}
	return createProgramFromConfigLenient(singleThreaded, config, host)
}

func (c *ProgramBuildContext) parseConfig(cwd string, tsconfigPath string) (compiler.CompilerHost, *tsoptions.ParsedCommandLine, error) {
	resolvedConfigPath := tspath.ResolvePath(cwd, tsconfigPath)
	if !c.fs.FileExists(resolvedConfigPath) {
		return nil, nil, configNotFoundError(resolvedConfigPath)
	}
	c.registerTSConfig(resolvedConfigPath)

	host := c.NewCompilerHost(cwd)
	config, _ := tsoptions.GetParsedCommandLineOfConfigFile(
		tsconfigPath,
		&core.CompilerOptions{},
		nil,
		host,
		c.extendedConfigCacheInterface(),
	)
	return host, config, nil
}

// CreateProgramFromOptionsLenient creates a non-project fallback Program using
// the same source/metadata view as the tsconfig-backed Programs.
func (c *ProgramBuildContext) CreateProgramFromOptionsLenient(
	singleThreaded bool,
	cwd string,
	compilerOptions *core.CompilerOptions,
	rootFileNames []string,
) (*compiler.Program, error) {
	return CreateProgramFromOptionsLenient(
		singleThreaded,
		compilerOptions,
		rootFileNames,
		c.NewCompilerHost(cwd),
	)
}

// RetainOnlySourceFiles bounds the append-mostly parsed-AST cache to SourceFile
// objects still retained by the supplied Programs.
func (c *ProgramBuildContext) RetainOnlySourceFiles(programs []*compiler.Program) {
	c.parseCache.RetainOnly(programs)
}

// InvalidateSourceSnapshots begins a fresh source generation before Programs
// are rebuilt after a fix write. Metadata generations are independent because
// CLI fixes never target package.json or tsconfig files.
func (c *ProgramBuildContext) InvalidateSourceSnapshots() {
	c.parseCache.InvalidateSourceSnapshots()
}

func (c *ProgramBuildContext) registerTSConfig(fileName string) {
	if c.metadataFS != nil {
		c.metadataFS.registerTSConfig(fileName)
	}
}

// extendedConfigCacheInterface avoids converting a nil concrete pointer into a
// non-nil interface when the diagnostic escape hatch disables metadata caches.
func (c *ProgramBuildContext) extendedConfigCacheInterface() tsoptions.ExtendedConfigCache {
	if c.extendedConfigCache == nil {
		return nil
	}
	return c.extendedConfigCache
}

// programBuildCompilerHost only adds project-reference registration to the
// upstream compiler host. All other CompilerHost behavior remains delegated.
type programBuildCompilerHost struct {
	compiler.CompilerHost
	context *ProgramBuildContext
}

func (h *programBuildCompilerHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	h.context.registerTSConfig(fileName)
	commandLine, _ := tsoptions.GetParsedCommandLineOfConfigFilePath(
		fileName,
		path,
		nil,
		nil,
		h,
		h.context.extendedConfigCacheInterface(),
	)
	return commandLine
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
	entries sync.Map // exact fileName -> *tsoptions.ExtendedConfigCacheEntry
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
	if cached, ok := generation.entries.Load(fileName); ok {
		return mustExtendedConfigCacheEntry(cached)
	}

	candidate := tsoptions.ParseExtendedConfig(fileName, path, resolutionStack, host, c)
	actual, _ := generation.entries.LoadOrStore(fileName, candidate)
	return mustExtendedConfigCacheEntry(actual)
}

func mustExtendedConfigCacheEntry(value any) *tsoptions.ExtendedConfigCacheEntry {
	entry, ok := value.(*tsoptions.ExtendedConfigCacheEntry)
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
	configPaths              sync.Map // exact fileName -> struct{}
	generation               atomic.Pointer[programMetadataGeneration]
	extendedConfigGeneration atomic.Pointer[programExtendedConfigGeneration]
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

func (f *programMetadataFS) invalidate() {
	// Publish the raw-read generation first. A concurrent config parse that
	// captures the old parsed-config generation can observe either raw
	// generation, but its result is discarded when the parsed generation is
	// replaced below. Publishing in the opposite order could let an old raw
	// read populate the new parsed generation.
	f.generation.Store(&programMetadataGeneration{})
	f.extendedConfigGeneration.Store(&programExtendedConfigGeneration{})
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
