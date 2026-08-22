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

	"github.com/web-infra-dev/rslint/internal/config"
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
	target config.DiscoveredLintTarget,
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
	target   config.DiscoveredLintTarget
	fs       vfs.FS
	loadFS   func() vfs.FS
	projects map[string]*lintProjectMetadata
}

func newStandaloneLintProjectRequest(
	target config.DiscoveredLintTarget,
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
	target config.DiscoveredLintTarget,
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
