package lsp

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
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
	// URI decoding normalizes the Windows drive through this same tsgo
	// helper. Parse from that base so equivalent snapshots also agree on
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

// selectedLintProject is the configured-project result of the shared LSP
// selection policy. Normal diagnostics and speculative fix passes use the same
// state machine while supplying different Program owners.
type selectedLintProject struct {
	program    *compiler.Program
	sourceFile *ast.SourceFile
	configPath string
}

// acquireLintProgram applies the resolved binding policy before either adapter
// acquires a Program. Disabled and unmatched targets share a source-only
// generation; an imported source is not an alternative configured root.
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

	if loaders.metadata != nil {
		for _, tsConfigPath := range tsConfigPaths {
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
			if !metadata.Contains(target.Path, target.CanonicalPath) {
				continue
			}
			return loadDirectLintProject(metadata, target, loaders.program)
		}
	}

	// A Program's imported sources cannot supply lint ownership. The caller
	// creates a source-only generation using the same document overlay.
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
		program: program, sourceFile: sourceFile, configPath: configPath,
	}, true, nil
}

// standaloneLintProjectRequest gives one isolated lint pass a stable parsed
// project snapshot. Root probing and Program construction share it, so a
// config cannot be parsed twice or change meaning halfway through selection.
type standaloneLintProjectRequest struct {
	target   target.File
	fs       vfs.FS
	loadFS   func() vfs.FS
	projects map[string]*lintProjectMetadata
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
	program, err := utils.CreateProgramFromParsedConfigLenientWithProjectReferences(
		false, singleCheckerLintCommandLine(metadata.commandLine),
		newLintProjectReferenceHost(
			utils.CreateCompilerHost(tspath.GetDirectoryPath(metadata.configPath), request.filesystem()),
			request.projects,
		),
	)
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForTarget(program, request.target, request.filesystem()), nil
}

// LSP file requests use one checker even when Program loading is parallel.
func singleCheckerLintCommandLine(commandLine *tsoptions.ParsedCommandLine) *tsoptions.ParsedCommandLine {
	commandLine = commandLine.WithFileNames(commandLine.FileNames())
	options := *commandLine.CompilerOptions()
	checkerCount := 1
	options.Checkers = &checkerCount
	commandLine.SetCompilerOptions(&options)
	return commandLine
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

func sourceFileForPath(program *compiler.Program, filename string, fs vfs.FS) *ast.SourceFile {
	return utils.NewProgramSourceLookup(program, fs).SourceFileForPath(filename, fs)
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
	program, err := utils.CreateProgramFromOptionsLenient(true, lintprogram.SourceOnlyCompilerOptions(), []string{target.Path}, host)
	if err != nil {
		return nil, nil, fmt.Errorf("create fallback lint program: %w", err)
	}
	return program, sourceFileForTarget(program, target, fs), nil
}
