package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (h *Handler) handleLint(ctx context.Context, req api.LintRequest, dispatch linter.EslintPluginDispatcher, requester api.Requester) (*api.LintResponse, error) {

	// Resolve the working directory WITHOUT os.Chdir: this is a long-lived,
	// reused --api process, so mutating the process-global cwd would leak
	// across requests (and race a future concurrent mode). Everything
	// downstream (resolveRequestPath / config loader / CreateCompilerHost /
	// CreateProgram) takes this directory explicitly, so a local var suffices.
	currentDirectory := req.WorkingDirectory
	if currentDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current directory: %w", err)
		}
		currentDirectory = cwd
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)

	// Create filesystem
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	var allowedFiles []string
	seenAllowedFiles := make(map[string]struct{})

	resolveRequestPath := func(filePath string) string {
		if tspath.PathIsAbsolute(filePath) {
			return tspath.NormalizePath(filePath)
		}
		return tspath.ResolvePath(currentDirectory, filePath)
	}
	if len(req.CanonicalFiles) > 0 && len(req.CanonicalFiles) != len(req.Files) {
		return nil, errors.New("canonicalFiles must be parallel to files")
	}
	canonicalPaths := make(map[string]string, len(req.CanonicalFiles))
	for index, canonicalPath := range req.CanonicalFiles {
		filePath := resolveRequestPath(req.Files[index])
		canonicalPath = resolveRequestPath(canonicalPath)
		canonicalPaths[rslintconfig.ExactPathID(filePath)] = canonicalPath
		canonicalPaths[rslintconfig.ExactPathID(canonicalPath)] = canonicalPath
	}

	addAllowedFile := func(filePath string) string {
		normalizedPath := resolveRequestPath(filePath)
		if _, exists := seenAllowedFiles[normalizedPath]; exists {
			return normalizedPath
		}
		seenAllowedFiles[normalizedPath] = struct{}{}
		allowedFiles = append(allowedFiles, normalizedPath)
		return normalizedPath
	}

	if req.Files != nil {
		allowedFiles = make([]string, 0, len(req.Files)+len(req.FileContents))
		for _, filePath := range req.Files {
			addAllowedFile(filePath)
		}
	}
	// Apply file contents if provided
	var fileContents map[string]string
	if len(req.FileContents) > 0 {
		if allowedFiles == nil {
			allowedFiles = make([]string, 0, len(req.FileContents))
		}
		fileContents = make(map[string]string, len(req.FileContents))
		for k, v := range req.FileContents {
			normalizedPath := resolveRequestPath(k)
			// Preserve the low-level IPC contract: when Files is omitted,
			// FileContents supplies the in-memory target set. When Files is
			// present, FileContents is overlay-only dependency/config data and
			// must not widen the explicit lint target set.
			if req.Files == nil {
				addAllowedFile(normalizedPath)
			}
			fileContents[normalizedPath] = v
		}
	}

	if len(req.Config) > 0 && req.ConfigDiscovery != nil {
		return nil, errors.New("config and configDiscovery are mutually exclusive")
	}

	// Config is the current low-level already-resolved config. High-level native
	// API callers instead send ConfigDiscovery: Go discovers ownership and asks
	// the host to evaluate only the staged candidate frontier.
	var rslintConfig rslintconfig.RslintConfig
	if len(req.Config) > 0 {
		if err := json.Unmarshal(req.Config, &rslintConfig); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		if err := rslintconfig.ValidateConfig(rslintConfig); err != nil {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
	}
	configDirectory := req.ConfigDirectory
	if configDirectory == "" {
		configDirectory = currentDirectory
	}
	configDirectory = tspath.NormalizePath(configDirectory)
	if len(fileContents) > 0 {
		addEquivalentFileContentPaths(fileContents, configDirectory, currentDirectory, fs)
		fs = utils.NewOverlayVFS(fs, fileContents)
	}
	if len(canonicalPaths) > 0 {
		fs = &canonicalPathVFS{FS: fs, canonicalPaths: canonicalPaths}
	}
	sourceFS := fs
	// The Program context must wrap the fully composed request VFS so metadata
	// snapshots observe overlay contents and canonical aliases exactly as the
	// rest of this request does. A new context is created for every API call.
	programSession := loader.NewSession(fs)
	fs = programSession.FS()

	var (
		configMap              map[string]rslintconfig.RslintConfig
		configTargetScopes     map[string]target.OwnerScope
		catalogPlugins         []rslintconfig.EslintPluginEntry
		pluginConfigKeyByOwner map[string]string
		configGitignoreFrozen  bool
	)
	if configDiscovery := req.ConfigDiscovery; configDiscovery != nil {
		if requester == nil {
			return nil, errors.New("configDiscovery requires a bidirectional API host")
		}
		if len(configDiscovery.ExplicitFiles) > 0 && len(configDiscovery.ExplicitFiles) != len(req.Files) {
			return nil, errors.New("configDiscovery.explicitFiles must be parallel to files")
		}
		var overrideConfig rslintconfig.RslintConfig
		if len(configDiscovery.OverrideConfig) > 0 {
			if err := json.Unmarshal(configDiscovery.OverrideConfig, &overrideConfig); err != nil {
				return nil, fmt.Errorf("invalid configDiscovery.overrideConfig: %w", err)
			}
			if err := rslintconfig.ValidateConfig(overrideConfig); err != nil {
				return nil, fmt.Errorf("invalid configDiscovery.overrideConfig: %w", err)
			}
			overrideConfig = rslintconfig.ConfigWithAuthoredPathBase(overrideConfig, currentDirectory)
		}
		discoveryRequest := discovery.ConfigDiscoveryRequest{
			CWD:                       currentDirectory,
			Fresh:                     true,
			LimitDirectoryWalkToFiles: len(configDiscovery.Directories) > 0,
			ImplicitCWD:               len(req.Files) == 0 && len(configDiscovery.Directories) == 0,
			SingleThreaded:            false,
		}
		for _, directory := range configDiscovery.Directories {
			discoveryRequest.Directories = append(discoveryRequest.Directories, resolveRequestPath(directory))
		}
		for index, filePath := range req.Files {
			explicit := true
			if len(configDiscovery.ExplicitFiles) > 0 {
				explicit = configDiscovery.ExplicitFiles[index]
			}
			canonicalPath := ""
			if len(req.CanonicalFiles) > 0 {
				canonicalPath = resolveRequestPath(req.CanonicalFiles[index])
			}
			discoveryRequest.Files = append(discoveryRequest.Files, discovery.DiscoveryFile{
				Path:          resolveRequestPath(filePath),
				CanonicalPath: canonicalPath,
				Explicit:      explicit,
			})
		}

		loader := &apiConfigModuleLoader{
			requester:      requester,
			overrideConfig: overrideConfig,
		}
		var configCatalog *discovery.ConfigCatalog
		var err error
		if configDiscovery.ExplicitConfigPath != "" {
			var targetFiles []discovery.DiscoveryFile
			if allowedFiles != nil {
				targetFiles = make([]discovery.DiscoveryFile, 0, len(allowedFiles))
				for _, filePath := range allowedFiles {
					targetFiles = append(targetFiles, discovery.DiscoveryFile{
						Path:          filePath,
						CanonicalPath: canonicalPaths[rslintconfig.ExactPathID(filePath)],
					})
				}
			}
			configCatalog, err = discovery.LoadExplicitConfig(
				ctx,
				fs,
				loader,
				discovery.ExplicitConfigRequest{
					CWD:               currentDirectory,
					ConfigPath:        resolveRequestPath(configDiscovery.ExplicitConfigPath),
					TargetFiles:       targetFiles,
					TargetDirectories: append([]string(nil), discoveryRequest.Directories...),
					Fresh:             true,
				},
			)
		} else {
			configCatalog, err = discovery.DiscoverAutomatic(ctx, fs, loader, discoveryRequest)
		}
		if err != nil {
			return nil, fmt.Errorf("discover config catalog: %w", err)
		}
		if len(configCatalog.Failures) > 0 {
			printConfigDiscoveryFailures(configCatalog.Failures)
		}
		if len(configCatalog.EslintPlugins) > 0 {
			if capabilityRequester, ok := requester.(api.PeerCapabilityRequester); ok &&
				!capabilityRequester.PeerSupportsCapability(api.CapabilityReversePluginLint) {
				return nil, errors.New("API peer does not advertise reversePluginLint capability required by discovered ESLint plugins")
			}
		}

		if len(configCatalog.Configs) > 0 {
			configDirectories := configCatalog.ConfigDirectories()
			if len(configDirectories) == 1 && configCatalog.Explicit {
				// overrideConfigFile is invocation-wide. A hierarchical config map
				// would have no owner for a requested file outside cwd and would
				// incorrectly drop it, even though explicit flat-config semantics say
				// the selected module governs the complete supplied target set.
				configDirectory = configDirectories[0]
				rslintConfig = append(rslintconfig.RslintConfig(nil), configCatalog.Configs[configDirectory]...)
				pluginConfigKeyByOwner = map[string]string{configDirectory: configDirectory}
				configGitignoreFrozen = true
			} else {
				configMap = make(map[string]rslintconfig.RslintConfig, len(configCatalog.Configs))
				pluginConfigKeyByOwner = make(map[string]string, len(configCatalog.Configs))
			}
			for ownerDirectory, entries := range configCatalog.Configs {
				if configMap == nil {
					continue
				}
				configMap[ownerDirectory] = append(rslintconfig.RslintConfig(nil), entries...)
				pluginConfigKeyByOwner[ownerDirectory] = ownerDirectory
			}
			if configMap != nil {
				configTargetScopes = configCatalog.Scopes
			}
		} else {
			// No JS candidate is a valid API state: lint with override entries (or
			// syntax-only with an empty config) rather than falling back to the CLI's
			// rslint.json lookup.
			rslintConfig = overrideConfig
			configDirectory = currentDirectory
		}
		catalogPlugins = configCatalog.EslintPlugins
	}

	if configMap == nil && !configGitignoreFrozen {
		rslintConfig = rslintconfig.ConfigWithGitignoreForTargetsFromRoot(
			rslintConfig,
			configDirectory,
			currentDirectory,
			fs,
			allowedFiles,
			nil,
		)
	}
	pluginEntries := append([]rslintconfig.EslintPluginEntry(nil), catalogPlugins...)
	for _, plugin := range req.EslintPlugins {
		pluginEntries = append(pluginEntries, rslintconfig.EslintPluginEntry{
			Prefix:    plugin.Prefix,
			RuleNames: append([]string(nil), plugin.RuleNames...),
		})
	}
	ruleCatalog, shadowedPluginRules := rules.All().ForESLintPlugins(pluginEntries)
	var optionErrors []rslintconfig.ResolvedRuleOptionsError
	configMap, rslintConfig, optionErrors = rslintconfig.ValidateResolvedRuleOptions(configMap, rslintConfig, ruleCatalog)
	if len(optionErrors) > 0 {
		messages := make([]string, len(optionErrors))
		for index, optionError := range optionErrors {
			messages[index] = optionError.Error()
		}
		return nil, fmt.Errorf("invalid rule options:\n%s", strings.Join(messages, "\n"))
	}
	reportShadowedPluginRules(shadowedPluginRules)

	responsePathBase := configDirectory
	if req.ConfigDiscovery != nil {
		// ConfigDiscovery is the high-level API: request and response paths use
		// WorkingDirectory even when an explicitly selected config lives elsewhere.
		// The low-level Config API keeps its ConfigDirectory-relative contract.
		responsePathBase = currentDirectory
	}
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          responsePathBase,
		UseCaseSensitiveFileNames: true,
	}

	// Resolve the exact lint target set and load one unified Program sequence.
	// Selected files outside configured projects (the typical lintText/lintFiles
	// in-memory file) are represented by source-only Programs. Native discovery
	// can supply a hierarchical
	// configMap; low-level config and explicit overrideConfigFile requests remain
	// invocation-wide single-config paths.
	// The --api path never runs the type-check phase (RunLinterOptions.TypeCheck
	// stays false), so there is no per-program type-check skip mask to build.
	targetPlan, err := target.Resolve(target.Request{
		ConfigMap:       configMap,
		Config:          rslintConfig,
		ConfigDirectory: configDirectory,
		ScanRoot:        currentDirectory,
		OwnerScopes:     configTargetScopes,
		FS:              fs,
		Files:           allowedFiles,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve lint targets: %w", err)
	}
	// A plain API lint only needs type information when at least one target is
	// selected. The stable target/config plan is reused while each source
	// snapshot gets a fresh request-local Program generation.
	buildBinding := func(session *loader.Session) (loader.LoadResult, error) {
		var projects loader.ProjectSet
		var buildErr error
		if len(targetPlan.Files) > 0 {
			if configMap != nil {
				projects, buildErr = session.BuildTargetProjects(configMap, targetPlan, false)
			} else {
				projects, buildErr = session.BuildTargetProject(configDirectory, rslintConfig, targetPlan, false)
			}
			if buildErr != nil {
				return loader.LoadResult{}, buildErr
			}
		}
		return session.LoadAPI(projects, targetPlan, configDirectory, false)
	}
	binding, err := buildBinding(programSession)
	if err != nil {
		return nil, err
	}

	if len(pluginEntries) > 0 {
		if pluginConfigKeyByOwner == nil {
			wireConfigDirectory := req.PluginConfigDirectory
			if wireConfigDirectory == "" {
				wireConfigDirectory = req.ConfigDirectory
			}
			if wireConfigDirectory == "" {
				wireConfigDirectory = configDirectory
			}
			pluginConfigKeyByOwner = map[string]string{configDirectory: wireConfigDirectory}
		}
		if dispatch == nil {
			dispatch = func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
				return nil, errors.New("bidirectional pluginLint transport is unavailable")
			}
		}
	}
	generationForBinding := func(binding loader.LoadResult, generationFS vfs.FS) linter.Generation {
		fileConfigResolver := configLint.NewResolver(configLint.ResolverOptions{
			ConfigsByOwner:                      configMap,
			Config:                              rslintConfig,
			ConfigDirectory:                     configDirectory,
			Catalog:                             ruleCatalog,
			EnforcePlugins:                      true,
			TargetsBySourcePath:                 binding.LintTargetBySourcePath,
			SourceMappingsIncludeCanonicalPaths: true,
			PathSpaces:                          targetPlan.PathSpaces(),
			FS:                                  generationFS,
		})
		targetPathForSourcePath := func(sourcePath string) string {
			if lintTarget, bound := fileConfigResolver.TargetForSourcePath(sourcePath); bound {
				return lintTarget.Path
			}
			return sourcePath
		}
		var plugin *linter.PluginGeneration
		if len(pluginEntries) > 0 {
			plugin = &linter.PluginGeneration{
				ConfigForFile: eslintPluginConfigResolver{
					lintResolver:           fileConfigResolver,
					pluginConfigKeyByOwner: pluginConfigKeyByOwner,
				}.resolve,
			}
		}
		return linter.Generation{
			Native: linter.NativeGeneration{
				Programs:         binding.Programs,
				TargetsByProgram: binding.TargetsByProgram,
				SingleThreaded:   false, // Don't use single-threaded mode for IPC
				Cwd:              currentDirectory,
				RulesForFile: func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
					return fileConfigResolver.EnabledRulesForSourcePath(sourceFile.FileName())
				},
			},
			Target: linter.TargetProjection{
				Path: targetPathForSourcePath,
				ReadText: func(path string, source ast.SourceFileLike) (string, error) {
					return utils.RestoreSourceBOM(generationFS, path, source.Text()), nil
				},
			},
			Plugin: plugin,
		}
	}
	provider := &apiGenerationProvider{
		initial: generationForBinding(binding, programSession.FS()),
		rebuild: func(ctx context.Context, snapshot linter.SourceSnapshot) (linter.Generation, error) {
			if err := ctx.Err(); err != nil {
				return linter.Generation{}, err
			}
			files := snapshot.Files()
			overrides := make(map[string]string, len(files)*2)
			for _, file := range files {
				overrides[tspath.NormalizePath(file.Path)] = file.Text
			}
			addEquivalentFileContentPaths(overrides, configDirectory, currentDirectory, sourceFS)
			roundSession := loader.NewSession(utils.NewOverlayVFS(sourceFS, overrides))
			roundBinding, rebuildErr := buildBinding(roundSession)
			if rebuildErr != nil {
				return linter.Generation{}, rebuildErr
			}
			if err := ctx.Err(); err != nil {
				return linter.Generation{}, err
			}
			return generationForBinding(roundBinding, roundSession.FS()), nil
		},
	}
	demand := linter.ArtifactDemand{
		Native:      rule.EditDemandAll,
		LintedFiles: true,
	}
	if req.Fix {
		demand.Plugin = rule.EditDemandAll
	}
	policy := linter.ObservationPolicy{
		Demand:        demand,
		Plugin:        linter.PluginConcurrentJoined,
		PluginFailure: linter.PluginKeepPartialWithSynthetic,
	}
	var pipelineRequest linter.PipelineRequest
	if req.Fix {
		pipelineRequest = linter.NewAutofixRequest(
			provider,
			policy,
			linter.AutofixPolicy{
				VerifyAfterLastRound: true,
				VerificationDemand:   demand,
			},
			dispatch,
		)
	} else {
		pipelineRequest = linter.NewLintRequest(provider, policy, dispatch)
	}
	pipelineResult, err := linter.RunPipeline(ctx, pipelineRequest)
	for _, pluginRecord := range pipelineResult.PluginOutcomes() {
		reportEslintPluginDispatchOutcome(linter.EslintPluginDispatchOutcome{
			Notices:       pluginRecord.Notices,
			DispatchError: pluginRecord.DispatchError,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("error running linter: %w", err)
	}
	diagnostics, complete := pipelineResult.Observation.CompleteDiagnostics()
	if !complete {
		return nil, errors.New("error running linter: API lint returned an incomplete observation")
	}
	lintResult := pipelineResult.Observation.Native.Lint

	// The shared result uses absolute stable target identities. Copy it into the
	// API's caller-visible relative path space only after all producers joined.
	for index := range diagnostics {
		diagnostics[index].FilePath = tspath.ConvertToRelativePath(diagnostics[index].FilePath, comparePathOptions)
	}

	// Track every selected target, including syntax-error and zero-rule files,
	// from the prepared plan rather than diagnostic/GetRules side effects.
	sourceFiles := make(map[string]*ast.SourceFile, len(pipelineResult.Observation.Native.Files))
	for _, lintedFile := range pipelineResult.Observation.Native.Files {
		if lintedFile.SourceFile == nil {
			continue
		}
		responsePath := tspath.ConvertToRelativePath(lintedFile.Path, comparePathOptions)
		sourceFiles[responsePath] = lintedFile.SourceFile
	}

	linter.StableSortDiagnosticsByFileAndStart(diagnostics)
	diagnosticProjection := projectLintDiagnostics(diagnostics)

	// The core applies every bounded round to request-local memory and returns a
	// final observation over the resulting source. This adapter maps every
	// successfully fixed file's complete final source into Output; the JS API
	// keeps ownership of any later physical persistence.
	var output map[string]string
	if req.Fix {
		applied, ok := pipelineResult.AppliedFixes()
		if !ok {
			return nil, errors.New("error running linter: API fix did not return an in-memory result")
		}
		if len(applied.FinalSources) > 0 {
			output = make(map[string]string, len(applied.FinalSources))
			for _, source := range applied.FinalSources {
				output[tspath.ConvertToRelativePath(source.Path, comparePathOptions)] = source.Text
			}
		}
		if len(output) == 0 {
			output = nil
		}
	}

	// The files actually linted (target discovery already excluded global
	// ignores and gitignore entries). sourceFiles comes from the prepared plan's
	// complete selected-file projection under each caller-visible target path,
	// relative to configDirectory. This keeps
	// Diagnostic.FilePath, LintedFiles, Output, and EncodedSourceFiles in one path
	// space even when a Program represents a requested symlink target by a
	// different source-file path. Sorted for a deterministic response.
	lintedFiles := make([]string, 0, len(sourceFiles))
	for filePath := range sourceFiles {
		lintedFiles = append(lintedFiles, filePath)
	}
	sort.Strings(lintedFiles)

	// Create response
	response := &api.LintResponse{
		Diagnostics:         diagnosticProjection.diagnostics,
		ErrorCount:          diagnosticProjection.errorCount,
		WarningCount:        diagnosticProjection.warningCount,
		FixableErrorCount:   diagnosticProjection.fixableErrorCount,
		FixableWarningCount: diagnosticProjection.fixableWarningCount,
		// FileCount mirrors the unique caller-visible LintedFiles result set.
		FileCount:   len(lintedFiles),
		RuleCount:   len(lintResult.ExecutedRules),
		LintedFiles: lintedFiles,
		Output:      output,
	}
	// Only include encoded source files if requested
	if req.IncludeEncodedSourceFiles {
		encodedSourceFiles := make(map[string]api.ByteArray)
		for filePath, sourceFile := range sourceFiles {
			encoded, err := api.EncodeAST(sourceFile, filePath)

			if err != nil {
				// Log error but don't fail the entire request
				fmt.Fprintf(os.Stderr, "warning: failed to encode source file %s: %v\n", filePath, err)
				continue
			}
			encodedSourceFiles[filePath] = encoded
		}
		response.EncodedSourceFiles = encodedSourceFiles
	}
	return response, nil
}
