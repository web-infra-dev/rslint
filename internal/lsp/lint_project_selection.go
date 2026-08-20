package lsp

import (
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/program/projectselection"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type lintProgramLoader func(
	tsConfigPath string,
) (*compiler.Program, *ast.SourceFile, error)

type lintProjectMetadataLoader func(
	tsConfigPath string,
) (*lintProjectMetadata, bool, error)

type lintProjectLoaders struct {
	program  lintProgramLoader
	metadata lintProjectMetadataLoader
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

func (metadata *lintProjectMetadata) DirectRoot(target projectselection.Target) bool {
	return metadata.Contains(target.Path, target.CanonicalPath)
}

func (metadata *lintProjectMetadata) Supports(target projectselection.Target) bool {
	return metadata.supportsFileName(target.Path)
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
	configDir := tspath.GetDirectoryPath(tsConfigPath)
	host := utils.CreateCompilerHost(configDir, parseFS)
	parsed, _ := tsoptions.GetParsedCommandLineOfConfigFile(
		tsConfigPath,
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

func selectConfiguredLintProject(
	projectPaths []string,
	targetPath string,
	loaders lintProjectLoaders,
) (selectedLintProject, bool, error) {
	if len(projectPaths) == 0 {
		return selectedLintProject{}, false, nil
	}
	projects := make([]int, len(projectPaths))
	for index := range projects {
		projects[index] = index
	}
	target := projectselection.Target{Path: tspath.NormalizePath(targetPath), Projects: projects}
	if loaders.metadata == nil {
		return selectedLintProject{}, false, nil
	}
	programByProject := make([]*compiler.Program, len(projectPaths))
	sourceByProject := make([]*ast.SourceFile, len(projectPaths))
	loadedProjects := make([]bool, len(projectPaths))
	loadMetadata := func(project int) (projectselection.Metadata, bool, error) {
		metadata, available, err := loaders.metadata(projectPaths[project])
		if err != nil {
			return nil, false, fmt.Errorf("load configured project roots %q: %w", projectPaths[project], err)
		}
		return metadata, available && metadata != nil, nil
	}
	loadProject := func(project int) (bool, error) {
		if loadedProjects[project] {
			return programByProject[project] != nil, nil
		}
		loadedProjects[project] = true
		if loaders.program == nil {
			return false, nil
		}
		program, sourceFile, err := loaders.program(projectPaths[project])
		if err != nil {
			return false, fmt.Errorf("load configured project %q: %w", projectPaths[project], err)
		}
		if sourceFile != nil && program == nil {
			return false, fmt.Errorf("configured project %q returned a source without a Program", projectPaths[project])
		}
		programByProject[project] = program
		sourceByProject[project] = sourceFile
		return program != nil, nil
	}
	contains := func(project int, _ projectselection.Target) bool {
		return sourceByProject[project] != nil
	}
	bindings, err := projectselection.Resolve(
		projectselection.Plan{Targets: []projectselection.Target{target}},
		loadMetadata,
		loadProject,
		contains,
	)
	if err != nil {
		var absent *projectselection.DirectRootAbsentError
		if errors.As(err, &absent) {
			return selectedLintProject{}, false, fmt.Errorf(
				"configured project root %q was absent from %q",
				targetPath,
				projectPaths[absent.Project],
			)
		}
		var unavailable *projectselection.DirectProjectUnavailableError
		if errors.As(err, &unavailable) {
			return selectedLintProject{}, false, fmt.Errorf(
				"configured project root %q was absent from %q",
				targetPath,
				projectPaths[unavailable.Project],
			)
		}
		return selectedLintProject{}, false, err
	}
	if len(bindings) == 0 || bindings[0].Project == projectselection.NoProject {
		return selectedLintProject{}, false, nil
	}
	project := bindings[0].Project
	return selectedLintProject{
		program:    programByProject[project],
		sourceFile: sourceByProject[project],
		configPath: projectPaths[project],
		directRoot: bindings[0].Tier == projectselection.TierDirect,
	}, true, nil
}

// standaloneLintProjectRequest gives one isolated lint pass a stable parsed
// project snapshot. Root probing and Program construction share it, so a
// config cannot be parsed twice or change meaning halfway through selection.
type standaloneLintProjectRequest struct {
	targetPath string
	fs         vfs.FS
	loadFS     func() vfs.FS
	projects   map[string]*lintProjectMetadata
}

func newStandaloneLintProjectRequest(
	targetPath string,
	loadFS func() vfs.FS,
) *standaloneLintProjectRequest {
	return &standaloneLintProjectRequest{
		targetPath: tspath.NormalizePath(targetPath),
		loadFS:     loadFS,
		projects:   make(map[string]*lintProjectMetadata),
	}
}

func newStandaloneLintProjectRequestWithFS(
	targetPath string,
	fs vfs.FS,
) *standaloneLintProjectRequest {
	request := newStandaloneLintProjectRequest(targetPath, nil)
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
	tsConfigPath string,
) (*compiler.Program, *ast.SourceFile, error) {
	metadata, err := request.metadata(tsConfigPath)
	if err != nil {
		return nil, nil, err
	}
	program, err := createStandaloneLintProgram(metadata, request.filesystem())
	if err != nil {
		return nil, nil, err
	}
	return program, sourceFileForPath(program, request.targetPath, request.filesystem()), nil
}

func (request *standaloneLintProjectRequest) loadMetadata(
	tsConfigPath string,
) (*lintProjectMetadata, bool, error) {
	metadata, err := request.metadata(tsConfigPath)
	return metadata, err == nil && metadata != nil, err
}

func (request *standaloneLintProjectRequest) loaders() lintProjectLoaders {
	return lintProjectLoaders{
		program:  request.program,
		metadata: request.loadMetadata,
	}
}

type lintSessionProjectRootCache struct {
	mu      sync.Mutex
	entries map[string]lintSessionProjectRootEntry
}

type lintSessionProjectRootEntry struct {
	commandLine *tsoptions.ParsedCommandLine
	metadata    *lintProjectMetadata
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
	key := string(lintProjectDeclarationPathID(configPath))
	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry := cache.entries[key]
	if entry.commandLine == commandLine && entry.metadata != nil {
		return entry.metadata
	}
	metadata := newLintProjectMetadata(configPath, commandLine, fs)
	cache.entries[key] = lintSessionProjectRootEntry{
		commandLine: commandLine,
		metadata:    metadata,
	}
	return metadata
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
