package lsp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
	"github.com/microsoft/TypeScript/tsc/shim/project"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type lintProgramLoader func(
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error)

type lintProjectMetadataLoader func(
	tsConfigPath string,
) (*lintProjectMetadata, bool, error)

type lintProjectLoaders struct {
	program  lintProgramLoader
	metadata lintProjectMetadataLoader
	fallback func() (*compiler.Program, *ast.SourceFile, error)
}

// lintProjectMetadata is the lightweight, immutable part of one configured
// project generation. A standalone adapter retains the parsed command line so
// selecting the project never requires parsing its config again.
type lintProjectMetadata struct {
	configPath  string
	commandLine *tsoptions.ParsedCommandLine
	rootFiles   *lintprogram.RootFileIndex
}

func newLintProjectMetadata(
	configPath string,
	commandLine *tsoptions.ParsedCommandLine,
	identityFS vfs.FS,
) *lintProjectMetadata {
	if commandLine == nil {
		return nil
	}
	return &lintProjectMetadata{
		configPath:  tspath.NormalizePath(configPath),
		commandLine: commandLine,
		rootFiles: lintprogram.NewRootFileIndex(
			commandLine.FileNames(),
			identityFS,
		),
	}
}

func (metadata *lintProjectMetadata) supportsFileName(fileName string) bool {
	return metadata != nil && lintprogram.CompilerOptionsSupportFileName(
		metadata.commandLine.CompilerOptions(),
		fileName,
	)
}

func (metadata *lintProjectMetadata) Contains(
	fileName string,
	canonicalFileName string,
) bool {
	return metadata != nil && metadata.rootFiles != nil &&
		metadata.rootFiles.Contains(fileName, canonicalFileName)
}

func parseStandaloneLintProject(
	tsConfigPath string,
	parseFS vfs.FS,
	identityFS vfs.FS,
) (*lintProjectMetadata, error) {
	tsConfigPath = tspath.NormalizePath(tsConfigPath)
	if parseFS == nil {
		return nil, fmt.Errorf("cannot parse TypeScript config %q without a filesystem", tsConfigPath)
	}
	// Session URI decoding normalizes the Windows drive through this same
	// tsgo helper. Parse from that base so equivalent snapshots also agree on
	// resolved option paths, while retaining the declared metadata identity.
	volume, path, _ := tspath.SplitVolumePath(tsConfigPath)
	parsePath := volume + path
	configDir := tspath.GetDirectoryPath(parsePath)
	host := utils.CreateCompilerHost(configDir, parseFS)
	parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(
		parsePath,
		&core.CompilerOptions{},
		nil,
		host,
		nil,
	)
	if parsed == nil {
		return nil, fmt.Errorf("no parsed config returned for %q", tsConfigPath)
	}
	if identityFS == nil {
		identityFS = parseFS
	}
	return newLintProjectMetadata(tsConfigPath, parsed, identityFS), nil
}

func createStandaloneLintProgram(
	metadata *lintProjectMetadata,
	fs vfs.FS,
) (*compiler.Program, error) {
	if metadata == nil || metadata.commandLine == nil {
		return nil, errors.New("cannot create a Program without parsed project metadata")
	}
	configDir := tspath.GetDirectoryPath(metadata.configPath)
	host := utils.CreateCompilerHost(configDir, fs)
	return utils.CreateProgramFromParsedConfigLenient(
		true,
		metadata.commandLine,
		host,
	)
}

// selectedLintProject is the configured-project result of the shared LSP
// selection policy. Normal diagnostics and speculative fix passes use the same
// state machine while supplying different Program owners.
type selectedLintProject struct {
	program    *compiler.Program
	sourceFile *ast.SourceFile
	configPath string
	directRoot bool
}

// acquireLintProgram applies the resolved binding policy before either adapter
// acquires a Program. Disabled and unmatched targets share a source-only
// generation; a Session's default graph is not an alternative lint project.
func acquireLintProgram(
	tsConfigPaths []string,
	policy config.ProjectPolicy,
	target target.File,
	fs vfs.FS,
	loaders lintProjectLoaders,
) (*compiler.Program, *ast.SourceFile, bool, error) {
	if !policy.ProjectDisabled {
		selected, found, err := selectConfiguredLintProject(tsConfigPaths, policy.ServiceRootDirectory, target, fs, loaders)
		if err != nil {
			return nil, nil, false, err
		}
		if found {
			return selected.program, selected.sourceFile, true, nil
		}
	}
	program, source, err := loaders.fallback()
	return program, source, false, err
}

func selectConfiguredLintProject(
	tsConfigPaths []string,
	serviceRootDirectory string,
	target target.File,
	fs vfs.FS,
	loaders lintProjectLoaders,
) (selectedLintProject, bool, error) {
	if serviceRootDirectory != "" {
		metadataByConfig := make(map[string]*lintProjectMetadata)
		discovery := utils.NewTypeScriptProjectDiscovery(fs, func(configPath string) (*tsoptions.ParsedCommandLine, error) {
			metadata, available, err := loaders.metadata(configPath)
			if err != nil {
				return nil, err
			}
			if !available || metadata == nil {
				return nil, fmt.Errorf("no parsed config returned for %q", configPath)
			}
			metadataByConfig[metadata.commandLine.ConfigName()] = metadata
			return metadata.commandLine, nil
		})
		parsed, err := discovery.Find(target.Path, serviceRootDirectory)
		if err != nil || parsed == nil {
			return selectedLintProject{}, false, err
		}
		return loadDirectLintProject(metadataByConfig[parsed.ConfigName()], target, loaders.program)
	}

	metadataByProject := make([]*lintProjectMetadata, len(tsConfigPaths))
	if loaders.metadata != nil {
		for index, tsConfigPath := range tsConfigPaths {
			metadata, available, err := loaders.metadata(tsConfigPath)
			if err != nil {
				return selectedLintProject{}, false, fmt.Errorf(
					"load configured project roots %q: %w",
					tsConfigPath,
					err,
				)
			}
			if !available {
				continue
			}
			metadataByProject[index] = metadata
			if metadata == nil || metadata.rootFiles == nil ||
				!metadata.rootFiles.Contains(target.Path, target.CanonicalPath) {
				continue
			}
			return loadDirectLintProject(metadata, target, loaders.program)
		}
	}

	if loaders.program == nil {
		return selectedLintProject{}, false, nil
	}
	for index, tsConfigPath := range tsConfigPaths {
		metadata := metadataByProject[index]
		if metadata == nil || !metadata.supportsFileName(target.Path) {
			continue
		}
		program, sourceFile, err := loaders.program(metadata)
		if err != nil {
			return selectedLintProject{}, false, fmt.Errorf("load configured project %q: %w", tsConfigPath, err)
		}
		if sourceFile != nil {
			if program == nil {
				return selectedLintProject{}, false, fmt.Errorf("configured project %q returned a source without a Program", tsConfigPath)
			}
			return selectedLintProject{
				program:    program,
				sourceFile: sourceFile,
				configPath: tsConfigPath,
			}, true, nil
		}
	}
	return selectedLintProject{}, false, nil
}

func loadDirectLintProject(
	metadata *lintProjectMetadata,
	target target.File,
	loadProgram lintProgramLoader,
) (selectedLintProject, bool, error) {
	configPath := metadata.configPath
	if loadProgram == nil {
		return selectedLintProject{}, false, fmt.Errorf("configured project root %q cannot load %q", target.Path, configPath)
	}
	program, sourceFile, err := loadProgram(metadata)
	if err != nil {
		return selectedLintProject{}, false, fmt.Errorf("load configured project %q: %w", configPath, err)
	}
	if program == nil || sourceFile == nil {
		return selectedLintProject{}, false, fmt.Errorf("configured project root %q was absent from %q", target.Path, configPath)
	}
	return selectedLintProject{
		program: program, sourceFile: sourceFile, configPath: configPath, directRoot: true,
	}, true, nil
}

// standaloneLintProjectRequest gives one isolated lint pass a stable parsed
// project snapshot. Root probing and Program construction share it, so a
// config cannot be parsed twice or change meaning halfway through selection.
type standaloneLintProjectRequest struct {
	target           target.File
	fs               vfs.FS
	loadFS           func() vfs.FS
	projects         map[string]*lintProjectMetadata
	sourceReferences bool
}

func newStandaloneLintProjectRequest(
	target target.File,
	loadFS func() vfs.FS,
) *standaloneLintProjectRequest {
	target.Path = tspath.NormalizePath(target.Path)
	if target.CanonicalPath != "" {
		target.CanonicalPath = tspath.NormalizePath(target.CanonicalPath)
	}
	return &standaloneLintProjectRequest{
		target:   target,
		loadFS:   loadFS,
		projects: make(map[string]*lintProjectMetadata),
	}
}

func newStandaloneLintProjectRequestWithFS(
	target target.File,
	fs vfs.FS,
) *standaloneLintProjectRequest {
	request := newStandaloneLintProjectRequest(target, nil)
	request.fs = fs
	return request
}

func (request *standaloneLintProjectRequest) filesystem() vfs.FS {
	if request.fs == nil && request.loadFS != nil {
		request.fs = request.loadFS()
		request.loadFS = nil
	}
	return request.fs
}

func (request *standaloneLintProjectRequest) metadata(
	tsConfigPath string,
) (*lintProjectMetadata, error) {
	tsConfigPath = tspath.NormalizePath(tsConfigPath)
	if metadata := request.projects[tsConfigPath]; metadata != nil {
		return metadata, nil
	}
	fs := request.filesystem()
	metadata, err := parseStandaloneLintProject(tsConfigPath, fs, fs)
	if err != nil {
		return nil, err
	}
	request.projects[tsConfigPath] = metadata
	return metadata, nil
}

func (request *standaloneLintProjectRequest) program(
	metadata *lintProjectMetadata,
) (*compiler.Program, *ast.SourceFile, error) {
	var program *compiler.Program
	var err error
	if request.sourceReferences {
		program, err = utils.CreateProgramFromParsedConfigLenientWithProjectReferences(
			true, metadata.commandLine,
			newLintProjectReferenceHost(
				utils.CreateCompilerHost(tspath.GetDirectoryPath(metadata.configPath), request.filesystem()),
				request.projects,
			),
		)
	} else {
		program, err = createStandaloneLintProgram(metadata, request.filesystem())
	}
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForTarget(program, request.target, request.filesystem()), nil
}

func (request *standaloneLintProjectRequest) loadMetadata(
	tsConfigPath string,
) (*lintProjectMetadata, bool, error) {
	metadata, err := request.metadata(tsConfigPath)
	return metadata, err == nil && metadata != nil, err
}

func (request *standaloneLintProjectRequest) fallback() (*compiler.Program, *ast.SourceFile, error) {
	return createStandaloneFallbackProgram(request.target, request.filesystem())
}

func (request *standaloneLintProjectRequest) loaders() lintProjectLoaders {
	return lintProjectLoaders{
		program:  request.program,
		metadata: request.loadMetadata,
		fallback: request.fallback,
	}
}

// lintProjectReferenceHost retains already selected reference snapshots. Other
// references keep the compiler host's normal parsing and failed-lookup tracking.
// The immutable host never retains a request, its overlay supplier, or watchers.
type lintProjectReferenceHost struct {
	compiler.CompilerHost
	configs map[tspath.Path]*tsoptions.ParsedCommandLine
}

func newLintProjectReferenceHost(host compiler.CompilerHost, projects map[string]*lintProjectMetadata) compiler.CompilerHost {
	configs := make(map[tspath.Path]*tsoptions.ParsedCommandLine, len(projects))
	for _, metadata := range projects {
		configs[lintProgramLexicalPathID(metadata.configPath, host.FS())] = metadata.commandLine
	}
	return &lintProjectReferenceHost{CompilerHost: host, configs: configs}
}

func (host *lintProjectReferenceHost) GetResolvedProjectReference(fileName string, path tspath.Path) *tsoptions.ParsedCommandLine {
	if config := host.configs[path]; config != nil {
		return config
	}
	return host.CompilerHost.GetResolvedProjectReference(fileName, path)
}

// lintProjectSnapshotsEqual compares construction inputs, not mutable parsed
// config caches or AST identity. Extends are already reflected in these inputs.
func lintProjectSnapshotsEqual(selected, built *tsoptions.ParsedCommandLine, fs vfs.FS) bool {
	if selected == built {
		return true
	}
	if selected == nil || built == nil ||
		!reflect.DeepEqual(selected.CompilerOptions(), built.CompilerOptions()) ||
		!slices.EqualFunc(selected.GetConfigFileParsingDiagnostics(), built.GetConfigFileParsingDiagnostics(), ast.EqualDiagnostics) {
		return false
	}
	samePath := func(a, b string) bool {
		return lintProgramLexicalPathID(a, fs) == lintProgramLexicalPathID(b, fs)
	}
	return slices.EqualFunc(selected.FileNames(), built.FileNames(), samePath) &&
		slices.EqualFunc(selected.ProjectReferences(), built.ProjectReferences(), func(a, b *core.ProjectReference) bool {
			return samePath(a.Path, b.Path) && a.Circular == b.Circular
		})
}

func lintSessionProgramMatchesConfig(
	program *compiler.Program,
	metadata *lintProjectMetadata,
	loadMetadata lintProjectMetadataLoader,
	fs vfs.FS,
) bool {
	if !lintProjectSnapshotsEqual(metadata.commandLine, program.CommandLine(), fs) {
		return false
	}
	compatible := true
	program.RangeResolvedProjectReference(func(_ tspath.Path, built, parent *tsoptions.ParsedCommandLine, index int) bool {
		// A canonical path is an identity key, not a parser base. Preserve the
		// declaration's spelling so case-insensitive hosts retain equal options.
		configPath := core.ResolveProjectReferencePath(parent.ProjectReferences()[index])
		if built != nil {
			configPath = built.ConfigName()
		}
		current, available, err := loadMetadata(configPath)
		if err != nil || !available || current == nil || !lintProjectSnapshotsEqual(current.commandLine, built, fs) {
			compatible = false
		}
		return compatible
	})
	return compatible
}

type lintSessionProjectRootCache struct {
	mu      sync.Mutex
	entries map[string]lintSessionProjectRootEntry
}

type lintSessionProjectRootEntry struct {
	commandLine *tsoptions.ParsedCommandLine
	metadata    *lintProjectMetadata
	// A compatibility fact must not keep a Session's AST graph alive.
	program           weak.Pointer[compiler.Program]
	serviceCompatible bool
}

func newLintSessionProjectRootCache() *lintSessionProjectRootCache {
	return &lintSessionProjectRootCache{
		entries: make(map[string]lintSessionProjectRootEntry),
	}
}

func (cache *lintSessionProjectRootCache) metadata(
	configPath string,
	commandLine *tsoptions.ParsedCommandLine,
	fs vfs.FS,
) *lintProjectMetadata {
	if commandLine == nil {
		return nil
	}
	if cache == nil {
		return newLintProjectMetadata(configPath, commandLine, fs)
	}
	key := string(lintProgramLexicalPathID(configPath, fs))
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry := cache.entries[key]
	if entry.commandLine == commandLine && entry.metadata != nil {
		return entry.metadata
	}
	metadata := newLintProjectMetadata(configPath, commandLine, fs)
	entry.commandLine = commandLine
	entry.metadata = metadata
	cache.entries[key] = entry
	return metadata
}

// canUseServiceProgram checks the known construction difference between a
// configured Session Program and service construction. The result belongs to
// the current Program generation, not merely to its unchanged tsconfig roots.
func (cache *lintSessionProjectRootCache) canUseServiceProgram(configPath string, program *compiler.Program, fs vfs.FS) bool {
	if cache == nil {
		return lintSessionProgramSupportsService(program)
	}
	key := string(lintProgramLexicalPathID(configPath, fs))
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry := cache.entries[key]
	if entry.program.Value() == program {
		return entry.serviceCompatible
	}
	entry.program = weak.Make(program)
	entry.serviceCompatible = lintSessionProgramSupportsService(program)
	cache.entries[key] = entry
	return entry.serviceCompatible
}

func lintSessionProgramSupportsService(program *compiler.Program) bool {
	if program.Options().AllowNonTsExtensions.IsTrue() {
		return true
	}
	supported := func(fileName string) bool {
		return lintprogram.CompilerOptionsSupportFileName(program.Options(),
			tspath.GetCanonicalFileName(fileName, program.UseCaseSensitiveFileNames()))
	}
	for _, fileName := range program.CommandLine().FileNames() {
		if !supported(fileName) {
			return false
		}
	}
	// A supported target can depend on an explicit JS path reference that the
	// Session rejected. Inspect existing reference metadata without loading any
	// additional source or reproducing module resolution.
	for _, source := range program.GetSourceFiles() {
		for _, reference := range source.ReferencedFiles {
			if !supported(reference.FileName) {
				return false
			}
		}
	}
	return true
}

func (cache *lintSessionProjectRootCache) Invalidate() bool {
	if cache == nil {
		return false
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	hadState := len(cache.entries) != 0
	clear(cache.entries)
	return hadState
}

// selectLintProgram adapts already-loaded Session Programs and LSP-owned
// standalone Programs through the same configured-project selection policy.
func selectLintProgram(
	uri lsproto.DocumentUri,
	target target.File,
	session *project.Session,
	ctx context.Context,
	tsConfigPaths []string,
	policy config.ProjectPolicy,
	fs vfs.FS,
	fallbackLoaders lintProjectLoaders,
	sessionRoots *lintSessionProjectRootCache,
) (*compiler.Program, *ast.SourceFile, bool, error) {
	serviceRootDirectory := policy.ServiceRootDirectory
	type loadedLintProject struct {
		program     *compiler.Program
		commandLine *tsoptions.ParsedCommandLine
	}
	loadedByConfig := make(map[tspath.Path]loadedLintProject)
	var sessionSnapshot *project.Snapshot
	sessionLoaded := false
	loadSession := func() error {
		if session == nil || sessionLoaded {
			return nil
		}
		_, _, loadedProjects, err := session.GetLanguageServiceAndProjectsForFile(ctx, uri)
		if err != nil {
			return fmt.Errorf("failed to get language service: %w", err)
		}
		sessionLoaded = true
		if serviceRootDirectory != "" {
			sessionSnapshot = session.Snapshot()
		}
		for _, candidate := range loadedProjects {
			if candidate == nil || candidate.GetProgram() == nil {
				continue
			}
			candidateProgram := candidate.GetProgram()
			configPath := string(candidate.Id())
			if configPath == "" {
				continue
			}
			commandLine := candidateProgram.CommandLine()
			if sessionProject, ok := candidate.(*project.Project); ok && sessionProject.CommandLine != nil {
				commandLine = sessionProject.CommandLine
			}
			loadedByConfig[lintProgramLexicalPathID(configPath, fs)] = loadedLintProject{
				program: candidateProgram, commandLine: commandLine,
			}
		}
		return nil
	}
	findLoadedProject := func(configPath string) (loadedLintProject, bool) {
		key := lintProgramLexicalPathID(configPath, fs)
		if loaded, ok := loadedByConfig[key]; ok {
			return loaded, true
		}
		// A selected direct root can be absent from a Session Program when
		// source redirects are disabled. Reuse that exact project so the
		// missing-source error is preserved instead of building a substitute.
		if sessionSnapshot != nil {
			if candidate := sessionSnapshot.ProjectCollection.ConfiguredProject(key); candidate != nil && candidate.GetProgram() != nil {
				loaded := loadedLintProject{program: candidate.GetProgram(), commandLine: candidate.CommandLine}
				loadedByConfig[key] = loaded
				return loaded, true
			}
		}
		return loadedLintProject{}, false
	}
	loaders := lintProjectLoaders{
		fallback: fallbackLoaders.fallback,
		metadata: func(tsConfigPath string) (*lintProjectMetadata, bool, error) {
			// Service discovery reads the current overlay before asking Session
			// to update any Programs. Its previous command line can predate a
			// pending config change; existing standalone metadata has its own
			// watcher invalidation and remains safe to reuse here.
			if serviceRootDirectory == "" {
				if err := loadSession(); err != nil {
					return nil, false, err
				}
				if loadedProject, ok := findLoadedProject(tsConfigPath); ok {
					metadata := sessionRoots.metadata(tsConfigPath, loadedProject.commandLine, fs)
					return metadata, metadata != nil, nil
				}
			}
			if fallbackLoaders.metadata == nil {
				return nil, false, nil
			}
			return fallbackLoaders.metadata(tsConfigPath)
		},
		program: func(metadata *lintProjectMetadata) (*compiler.Program, *ast.SourceFile, error) {
			tsConfigPath := metadata.configPath
			if err := loadSession(); err != nil {
				return nil, nil, err
			}
			if loadedProject, ok := findLoadedProject(tsConfigPath); ok {
				if serviceRootDirectory == "" ||
					(lintSessionProgramMatchesConfig(loadedProject.program, metadata, fallbackLoaders.metadata, fs) &&
						sessionRoots.canUseServiceProgram(tsConfigPath, loadedProject.program, fs)) {
					return loadedProject.program, sourceFileForTarget(loadedProject.program, target, fs), nil
				}
			}
			if fallbackLoaders.program == nil {
				return nil, nil, nil
			}
			return fallbackLoaders.program(metadata)
		},
	}
	return acquireLintProgram(tsConfigPaths, policy, target, fs, loaders)
}

func sourceFileForPath(program *compiler.Program, filename string, fs vfs.FS) *ast.SourceFile {
	return utils.NewProgramSourceLookup(program, fs).SourceFileForPath(filename)
}

func sourceFileForTarget(
	program *compiler.Program,
	target target.File,
	fs vfs.FS,
) *ast.SourceFile {
	return utils.NewProgramSourceLookup(program, fs).
		SourceFileForTarget(target.Path, target.CanonicalPath)
}

// createStandaloneFallbackProgram parses exactly the frozen target when no
// configured project owns it. Both editor adapters supply the same overlay
// used for project selection and mark this generation as lacking type info.
func createStandaloneFallbackProgram(target target.File, fs vfs.FS) (*compiler.Program, *ast.SourceFile, error) {
	host := utils.CreateCompilerHost(target.ConfigDirectory, fs)
	program, err := utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{target.Path}, host)
	if err != nil {
		return nil, nil, fmt.Errorf("create fallback lint program: %w", err)
	}
	return program, sourceFileForTarget(program, target, fs), nil
}
