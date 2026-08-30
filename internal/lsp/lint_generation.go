package lsp

import (
	"context"
	"errors"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/vfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// rulesSkippedInEditors names rules whose evidence exists only in encoded file
// bytes. Editor documents contain decoded text, so those rules remain CLI-only.
var rulesSkippedInEditors = map[string]bool{
	"unicode-bom": true,
}

func rulesServedToEditors(configuredRules []rule.ConfiguredRule) []rule.ConfiguredRule {
	skipped := func(configured rule.ConfiguredRule) bool {
		return rulesSkippedInEditors[configured.Name]
	}
	if !slices.ContainsFunc(configuredRules, skipped) {
		return configuredRules
	}
	served := make([]rule.ConfiguredRule, 0, len(configuredRules))
	for _, configured := range configuredRules {
		if !skipped(configured) {
			served = append(served, configured)
		}
	}
	return served
}

// documentGenerationProvider adapts the current LSP Session and document
// snapshot to the linter's generation port. It selects and leases Programs but
// never decides how or when lint work runs.
type documentGenerationProvider struct {
	server          *Server
	uri             lsproto.DocumentUri
	snapshot        documentLintSnapshot
	requestPrograms documentProgramRequest
	buildGeneration documentGenerationBuilder
}

type documentProgramRequest func(
	ctx context.Context,
	uri lsproto.DocumentUri,
	target target.File,
) (lintProjectLoaders, linter.ReleaseFunc)

type documentGenerationBuilder func(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	target target.File,
	processCwd string,
	hasTypeInfo bool,
	snapshot documentLintSnapshot,
) linter.Generation

func buildDocumentGeneration(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	target target.File,
	processCwd string,
	hasTypeInfo bool,
	snapshot documentLintSnapshot,
) linter.Generation {
	return newLintGeneration(
		program,
		sourceFile,
		target,
		processCwd,
		hasTypeInfo,
		snapshot.resolvedConfig.EnabledRules,
		pluginFileConfigForLintSnapshot(snapshot),
		nil,
	)
}

func (p *documentGenerationProvider) AcquireGeneration(
	ctx context.Context,
	memory linter.SourceSnapshot,
) (linter.Generation, linter.ReleaseFunc, error) {
	if err := ctx.Err(); err != nil {
		return linter.Generation{}, nil, err
	}
	if p == nil || p.server == nil {
		return linter.Generation{}, nil, errors.New("LSP document generation provider is not configured")
	}
	if !memory.Empty() {
		return linter.Generation{}, nil, errors.New("LSP document diagnostics do not accept autofix snapshots")
	}
	server := p.server
	snapshot := p.snapshot
	if !snapshot.configResolved {
		snapshot = resolveDocumentLintSnapshotConfig(snapshot, server.fs)
	}
	if isDefaultExcludedLintPath(snapshot.target.Path, server.cwd, server.fs) ||
		snapshot.resolvedConfig.GloballyIgnored {
		return emptyLintGeneration(server.cwd), nil, nil
	}

	request := newStandaloneLintProjectRequest(
		snapshot.target,
		func() vfs.FS { return server.currentEditorOverlayFSForTarget(p.uri, snapshot.target) },
	)
	loaders := request.loaders()
	release := linter.ReleaseFunc(nil)
	releasePending := false
	defer func() {
		if releasePending && release != nil {
			release()
		}
	}()
	if p.requestPrograms != nil {
		loaders, release = p.requestPrograms(ctx, p.uri, snapshot.target)
		releasePending = release != nil
	} else if server.lintPrograms != nil && server.lintPrograms.Usable() {
		loadProgram, loadMetadata, finalize := server.lintPrograms.Request(
			ctx,
			p.uri,
			snapshot.target,
		)
		loaders = lintProjectLoaders{
			program:  loadProgram,
			metadata: loadMetadata,
		}
		release = finalize
		releasePending = release != nil
	}

	program, sourceFile, hasTypeInfo, err := selectLintProgram(
		p.uri,
		snapshot.target,
		server.session,
		ctx,
		snapshot.typeScriptConfigPaths,
		server.fs,
		loaders,
		server.lintSessionRoots,
	)
	if err != nil {
		return linter.Generation{}, nil, err
	}
	if sourceFile == nil {
		releasePending = false
		return emptyLintGeneration(server.cwd), release, nil
	}
	buildGeneration := p.buildGeneration
	if buildGeneration == nil {
		buildGeneration = buildDocumentGeneration
	}
	generation := buildGeneration(
		program,
		sourceFile,
		snapshot.target,
		server.cwd,
		hasTypeInfo,
		snapshot,
	)
	// Ownership transfers only after the complete generation has been built.
	// A panic anywhere above this point leaves releasePending armed.
	releasePending = false
	return generation, release, nil
}

func newLintGeneration(
	program *compiler.Program,
	sourceFile *ast.SourceFile,
	target target.File,
	processCwd string,
	hasTypeInfo bool,
	enabledRules []rule.ConfiguredRule,
	pluginConfig *linter.EslintPluginFileConfig,
	readText func(targetPath string, source ast.SourceFileLike) (string, error),
) linter.Generation {
	sourceProgram := lintprogram.NewFromCompiler(program)
	servedRules := rulesServedToEditors(enabledRules)
	if !hasTypeInfo {
		servedRules = rule.FilterNonTypeAwareRules(servedRules)
	}
	if readText == nil {
		readText = func(_ string, source ast.SourceFileLike) (string, error) {
			return source.Text(), nil
		}
	}
	var plugin *linter.PluginGeneration
	if pluginConfig != nil {
		plugin = &linter.PluginGeneration{
			ConfigForFile: func(string) linter.EslintPluginFileConfig {
				return *pluginConfig
			},
			WirePath: func(path string) string {
				if path == sourceFile.FileName() {
					return target.Path
				}
				return path
			},
		}
	}
	return linter.Generation{
		Native: linter.NativeGeneration{
			Programs:         []*lintprogram.Program{sourceProgram},
			TargetsByProgram: [][]string{{sourceFile.FileName()}},
			SingleThreaded:   true,
			Cwd:              processCwd,
			RulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return servedRules
			},
		},
		Target: linter.TargetProjection{
			Path: func(path string) string {
				if path == sourceFile.FileName() {
					return target.Path
				}
				return path
			},
			ReadText: readText,
		},
		Plugin: plugin,
	}
}

func pluginFileConfigForLintSnapshot(snapshot documentLintSnapshot) *linter.EslintPluginFileConfig {
	if !snapshot.configResolved || snapshot.resolvedConfig.MergedConfig == nil {
		return nil
	}
	languageOptions, settings := config.PluginMergedMaps(snapshot.resolvedConfig.MergedConfig)
	return &linter.EslintPluginFileConfig{
		ConfigKey:       snapshot.configKey,
		LanguageOptions: languageOptions,
		Settings:        settings,
	}
}

func emptyLintGeneration(processCwd string) linter.Generation {
	return linter.Generation{Native: linter.NativeGeneration{
		Cwd:            processCwd,
		SingleThreaded: true,
	}}
}
