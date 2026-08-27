package lsp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/tsoptions"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config/target"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
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
	tsConfigPaths []string,
	target target.File,
	loaders lintProjectLoaders,
) (selectedLintProject, bool, error) {
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
			if loaders.program == nil {
				return selectedLintProject{}, false, fmt.Errorf("configured project root %q cannot load %q", target.Path, tsConfigPath)
			}
			program, sourceFile, err := loaders.program(tsConfigPath)
			if err != nil {
				return selectedLintProject{}, false, fmt.Errorf("load configured project %q: %w", tsConfigPath, err)
			}
			if program == nil || sourceFile == nil {
				return selectedLintProject{}, false, fmt.Errorf(
					"configured project root %q was absent from %q",
					target.Path,
					tsConfigPath,
				)
			}
			return selectedLintProject{
				program:    program,
				sourceFile: sourceFile,
				configPath: tsConfigPath,
				directRoot: true,
			}, true, nil
		}
	}

	if loaders.program == nil {
		return selectedLintProject{}, false, nil
	}
	for index, tsConfigPath := range tsConfigPaths {
		metadata := metadataByProject[index]
		if metadata != nil && !metadata.supportsFileName(target.Path) {
			continue
		}
		program, sourceFile, err := loaders.program(tsConfigPath)
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
	return program, sourceFileForTarget(program, request.target, request.filesystem()), nil
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
	key := string(lintProgramLexicalPathID(configPath, fs))
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

// selectLintProgram chooses type information according to the authored
// parserOptions.project order while adapting already-loaded Session Programs
// and LSP-owned standalone Programs through one project-selection policy.
func selectLintProgram(
	uri lsproto.DocumentUri,
	target target.File,
	session *project.Session,
	ctx context.Context,
	tsConfigPaths []string,
	fs vfs.FS,
	fallbackLoaders lintProjectLoaders,
	sessionRoots *lintSessionProjectRootCache,
) (*compiler.Program, *ast.SourceFile, bool, error) {
	_, languageService, loadedProjects, err := session.GetLanguageServiceAndProjectsForFile(ctx, uri)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to get language service: %w", err)
	}
	program := languageService.GetProgram()

	type loadedLintProject struct {
		program     *compiler.Program
		commandLine *tsoptions.ParsedCommandLine
	}
	loadedByConfig := make(map[tspath.Path]loadedLintProject, len(loadedProjects))
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
			program:     candidateProgram,
			commandLine: commandLine,
		}
	}
	loaders := lintProjectLoaders{
		metadata: func(tsConfigPath string) (*lintProjectMetadata, bool, error) {
			if loadedProject, ok := loadedByConfig[lintProgramLexicalPathID(tsConfigPath, fs)]; ok {
				metadata := sessionRoots.metadata(tsConfigPath, loadedProject.commandLine, fs)
				return metadata, metadata != nil, nil
			}
			if fallbackLoaders.metadata == nil {
				return nil, false, nil
			}
			return fallbackLoaders.metadata(tsConfigPath)
		},
		program: func(tsConfigPath string) (*compiler.Program, *ast.SourceFile, error) {
			if loadedProject, ok := loadedByConfig[lintProgramLexicalPathID(tsConfigPath, fs)]; ok {
				return loadedProject.program, sourceFileForTarget(loadedProject.program, target, fs), nil
			}
			if fallbackLoaders.program == nil {
				return nil, nil, nil
			}
			return fallbackLoaders.program(tsConfigPath)
		},
	}
	selected, found, err := selectConfiguredLintProject(tsConfigPaths, target, loaders)
	if err != nil {
		return nil, nil, false, err
	}
	if found {
		return selected.program, selected.sourceFile, true, nil
	}
	return program, sourceFileForTarget(program, target, fs), false, nil
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

func createStandaloneFallbackProgram(filename string, cwd string, fs vfs.FS) (*compiler.Program, error) {
	host := utils.CreateCompilerHost(cwd, fs)
	return utils.CreateProgramFromOptionsLenient(true, &core.CompilerOptions{
		Target:    core.ScriptTargetESNext,
		Module:    core.ModuleKindESNext,
		Jsx:       core.JsxEmitPreserve,
		AllowJs:   core.TSTrue,
		NoLib:     core.TSTrue,
		NoResolve: core.TSTrue,
	}, []string{filename}, host)
}
