package lsp

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// speculativeLintEnvironment captures the base filesystem identity, process
// directory, and open-editor text used to rebuild isolated fix-all generations.
type speculativeLintEnvironment struct {
	baseFS     vfs.FS
	processCwd string
	openFiles  map[string]string
}

func (s *Server) freezeSpeculativeLintEnvironment(
	uri lsproto.DocumentUri,
	target target.File,
) speculativeLintEnvironment {
	openFiles, _ := s.currentEditorOverlayFilesForFrozenTarget(uri, target, "", false)
	return speculativeLintEnvironment{
		baseFS:     s.fs,
		processCwd: s.cwd,
		openFiles:  openFiles,
	}
}

func acquireSpeculativeGeneration(
	ctx context.Context,
	content string,
	snapshot documentLintSnapshot,
	environment speculativeLintEnvironment,
) (linter.Generation, linter.ReleaseFunc, error) {
	if err := ctx.Err(); err != nil {
		return linter.Generation{}, nil, err
	}
	if !snapshot.configResolved {
		snapshot = resolveDocumentLintSnapshotConfig(snapshot, environment.baseFS)
	}
	target := snapshot.target
	if isDefaultExcludedLintPath(target.Path, environment.processCwd, environment.baseFS) ||
		snapshot.resolvedConfig.GloballyIgnored {
		return emptyLintGeneration(environment.processCwd), nil, nil
	}

	files := make(map[string]string, len(environment.openFiles)+2)
	for path, text := range environment.openFiles {
		files[path] = text
	}
	addEditorOverlayTarget(files, target, content)
	overlayFS := newFrozenLintTargetOverlayFS(environment.baseFS, files, target)

	request := newStandaloneLintProjectRequestWithFS(target, overlayFS)
	selected, found, err := selectConfiguredLintProject(
		snapshot.typeScriptConfigPaths,
		target,
		request.loaders(),
	)
	if err != nil {
		return linter.Generation{}, nil, err
	}
	newGeneration := func(program *compiler.Program, sourceFile *ast.SourceFile, hasTypeInfo bool) linter.Generation {
		readText := func(path string, _ ast.SourceFileLike) (string, error) {
			if path != target.Path {
				return "", fmt.Errorf("unexpected LSP fix-all target %q", path)
			}
			return content, nil
		}
		return newLintGeneration(
			program,
			sourceFile,
			target,
			environment.processCwd,
			hasTypeInfo,
			snapshot.resolvedConfig.EnabledRules,
			pluginFileConfigForLintSnapshot(snapshot),
			readText,
		)
	}
	if found {
		if selected.sourceFile == nil {
			return emptyLintGeneration(environment.processCwd), nil, nil
		}
		return newGeneration(selected.program, selected.sourceFile, true), nil, nil
	}

	program, err := createStandaloneFallbackProgram(target.Path, target.ConfigDirectory, overlayFS)
	if err != nil {
		return linter.Generation{}, nil, fmt.Errorf("create fallback lint program: %w", err)
	}
	sourceFile := sourceFileForTarget(program, target, overlayFS)
	if sourceFile == nil {
		return emptyLintGeneration(environment.processCwd), nil, nil
	}
	return newGeneration(program, sourceFile, false), nil, nil
}

type speculativeGenerationAcquire func(
	context.Context,
	string,
	documentLintSnapshot,
	speculativeLintEnvironment,
) (linter.Generation, linter.ReleaseFunc, error)

// speculativeGenerationProvider maps the pipeline-owned memory snapshot onto an
// isolated editor overlay. It never advances fix rounds and never mutates the
// TypeScript Session, open-document store, or diagnostics cache.
type speculativeGenerationProvider struct {
	uri             lsproto.DocumentUri
	targetPath      string
	snapshot        documentLintSnapshot
	environment     speculativeLintEnvironment
	acquire         speculativeGenerationAcquire
	originalContent string
}

func (s *Server) newSpeculativeGenerationProvider(
	uri lsproto.DocumentUri,
	originalContent string,
	snapshot documentLintSnapshot,
) *speculativeGenerationProvider {
	acquire := speculativeGenerationAcquire(acquireSpeculativeGeneration)
	if s.speculativeGeneration != nil {
		acquire = s.speculativeGeneration
	}
	return &speculativeGenerationProvider{
		uri:             uri,
		targetPath:      snapshot.target.Path,
		snapshot:        snapshot,
		environment:     s.freezeSpeculativeLintEnvironment(uri, snapshot.target),
		acquire:         acquire,
		originalContent: originalContent,
	}
}

func (p *speculativeGenerationProvider) AcquireGeneration(
	ctx context.Context,
	snapshot linter.SourceSnapshot,
) (linter.Generation, linter.ReleaseFunc, error) {
	if p == nil {
		return linter.Generation{}, nil, errors.New("LSP fix-all generation provider is not configured")
	}
	if p.acquire == nil {
		return linter.Generation{}, nil, fmt.Errorf("LSP fix-all generation provider for %s is not configured", p.uri)
	}
	if err := ctx.Err(); err != nil {
		return linter.Generation{}, nil, err
	}
	content := p.originalContent
	files := snapshot.Files()
	if len(files) > 0 {
		if len(files) != 1 || files[0].Path != p.targetPath {
			return linter.Generation{}, nil, fmt.Errorf("LSP fix-all snapshot for %s must contain only %q", p.uri, p.targetPath)
		}
		content = files[0].Text
	}
	return p.acquire(ctx, content, p.snapshot, p.environment)
}
