package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
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
	// selected. Resolve the target plan before project paths so an ignored or
	// empty request cannot fail on an inactive project declaration.
	var projectSet loader.ProjectSet
	if len(targetPlan.Files) > 0 {
		if configMap != nil {
			projectSet, err = programSession.BuildTargetProjects(
				configMap,
				targetPlan,
				false,
			)
		} else {
			projectSet, err = programSession.BuildTargetProject(
				configDirectory,
				rslintConfig,
				targetPlan,
				false,
			)
		}
		if err != nil {
			return nil, err
		}
	}
	binding, err := programSession.LoadAPI(projectSet, targetPlan, configDirectory, false)
	if err != nil {
		return nil, err
	}
	programs := binding.Programs
	targetsByProgram := binding.TargetsByProgram
	fileConfigResolver := configLint.NewResolver(configLint.ResolverOptions{
		ConfigsByOwner:                      configMap,
		Config:                              rslintConfig,
		ConfigDirectory:                     configDirectory,
		Catalog:                             ruleCatalog,
		EnforcePlugins:                      true,
		TargetsBySourcePath:                 binding.LintTargetBySourcePath,
		SourceMappingsIncludeCanonicalPaths: true,
		PathSpaces:                          targetPlan.PathSpaces(),
		FS:                                  fs,
	})
	targetPathForSourcePath := func(sourcePath string) string {
		if target, bound := fileConfigResolver.TargetForSourcePath(sourcePath); bound {
			return target.Path
		}
		return sourcePath
	}
	responsePathForSourcePath := func(sourcePath string) string {
		return tspath.ConvertToRelativePath(targetPathForSourcePath(sourcePath), comparePathOptions)
	}

	// Collect diagnostics in the shared internal model. Each diagnostic is
	// copied into the API's caller-visible path space before the completed set
	// is sorted and projected to wire fields below.
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsLock sync.Mutex
	// When Fix is requested, the original RuleDiagnostics (byte-offset fixes +
	// their SourceFile) are retained per file for the in-band fix pass below.
	var diagnosticsByFile map[string][]rule.RuleDiagnostic
	if req.Fix {
		diagnosticsByFile = make(map[string][]rule.RuleDiagnostic)
	}

	// Track source files for encoding
	sourceFiles := make(map[string]*ast.SourceFile)
	var sourceFilesLock sync.Mutex

	// Create collector function
	diagnosticCollector := func(d rule.RuleDiagnostic) {
		diagnosticsLock.Lock()
		defer diagnosticsLock.Unlock()
		responsePath := responsePathForSourcePath(d.FilePath)
		if d.SourceFile != nil {
			sourceFilesLock.Lock()
			if sf, ok := d.SourceFile.(*ast.SourceFile); ok {
				sourceFiles[responsePath] = sf
			}
			sourceFilesLock.Unlock()
		}

		hasFix := d.FixesPtr != nil && len(*d.FixesPtr) > 0
		// Retain the original diagnostic (byte-offset fixes + SourceFile) for the
		// in-band fix pass, grouped by the caller-visible target path.
		if req.Fix && hasFix {
			targetPath := targetPathForSourcePath(d.FilePath)
			diagnosticsByFile[targetPath] = append(diagnosticsByFile[targetPath], d)
		}

		d.FilePath = responsePath
		diagnostics = append(diagnostics, d)
	}

	// Every selected target is parsed even when no config entry contributes
	// rules. Global ignores were already removed during target discovery.
	syntaxDiagnostics := linter.CollectTargetSyntacticDiagnostics(programs, targetsByProgram, false)
	for _, diagnostic := range syntaxDiagnostics {
		diagnosticCollector(diagnostic)
	}

	// Build one run descriptor and prepared plan shared by native lint and
	// plugin dispatch, keeping both paths on the exact same file/rule selection.
	runOpts := linter.RunLinterOptions{
		Programs:       programs,
		SingleThreaded: false, // Don't use single-threaded mode for IPC
		Cwd:            currentDirectory,
		Scope:          linter.FileScope{Files: allowedFiles},
		TargetFiles:    targetsByProgram,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
			// Track source file for encoding
			sourceFilesLock.Lock()
			filePath := responsePathForSourcePath(sourceFile.FileName())
			sourceFiles[filePath] = sourceFile
			sourceFilesLock.Unlock()

			// Rules come solely from the resolved config object (config.rules).
			// Program capability filtering happens once while PrepareLintPlan
			// freezes the shared native/plugin execution plan.
			//
			// enforcePlugins=true: the --api config is a resolved JS-style flat
			// config (plugins + rules), exactly like the CLI's JS/TS config path,
			// so a rule carrying a plugin prefix runs only when its plugin is
			// declared in the config's `plugins` — matching CLI and ESLint
			// semantics (a rule whose plugin is not declared is skipped).
			return fileConfigResolver.EnabledRulesForSourcePath(sourceFile.FileName())
		},
		// The API returns concrete fixes, suggestions, and fixable counts
		// independently of whether req.Fix later applies autofixes.
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: diagnosticCollector,
		},
	}
	preparedPlan, err := linter.PrepareLintPlan(runOpts)
	if err != nil {
		return nil, fmt.Errorf("error preparing lint plan: %w", err)
	}
	runOpts.PreparedPlan = preparedPlan

	// Metadata is the feature gate: without it there is no plugin target walk,
	// goroutine, or reverse request. With metadata, dispatch starts before the
	// native pass and runs in parallel, matching the CLI pipeline.
	var pluginCh <-chan linter.EslintPluginDispatchOutcome
	var cancelPlugin context.CancelFunc
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
		pluginInputs := linter.BuildEslintPluginFileInputs(runOpts.PreparedPlan, eslintPluginConfigResolver{
			lintResolver:           fileConfigResolver,
			pluginConfigKeyByOwner: pluginConfigKeyByOwner,
		}.resolve)
		for i := range pluginInputs {
			// Programmatic lint supports in-memory overlays. Always send the exact
			// parsed source frame instead of asking the host to re-read disk.
			if pluginInputs[i].SourceFile != nil {
				text := pluginInputs[i].SourceFile.Text()
				pluginInputs[i].Text = &text
			}
		}
		if len(pluginInputs) > 0 {
			pluginCtx := ctx
			pluginCtx, cancelPlugin = context.WithCancel(pluginCtx)
			if dispatch == nil {
				dispatch = func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
					return nil, errors.New("bidirectional pluginLint transport is unavailable")
				}
			}
			suggestionsMode := linter.SuggestionsModeOff
			if req.Fix {
				suggestionsMode = linter.SuggestionsModeEager
			}
			pluginCh = linter.DispatchEslintPluginRulesAsync(pluginCtx, dispatch, pluginInputs, req.Fix, suggestionsMode, nil, reportEslintPluginDispatchOutcome)
		}
	}
	if cancelPlugin != nil {
		defer cancelPlugin()
	}

	// Run native rules while community plugin batches execute in the host.
	lintResult, err := linter.RunLinter(runOpts)
	if err != nil {
		return nil, fmt.Errorf("error running linter: %w", err)
	}
	if pluginCh != nil {
		for _, diagnostic := range (<-pluginCh).Diagnostics {
			diagnosticCollector(diagnostic)
		}
	}

	linter.StableSortDiagnosticsByFileAndStart(diagnostics)
	diagnosticProjection := projectLintDiagnostics(diagnostics)

	// Apply fixes in-band when requested. ApplyRuleFixes is the same pure fixer
	// the CLI uses through applyFixPass, but here the result stays in-memory in
	// Output — the JS side persists it via Rslint.outputFixes. Single pass over
	// each file's fixes (non-overlapping applied, overlapping left for a later
	// lint); no cross-pass re-lint cascade (P1, see design §8).
	var output map[string]string
	if req.Fix && len(diagnosticsByFile) > 0 {
		output = make(map[string]string)
		for filePath, fileDiags := range diagnosticsByFile {
			// The parsed text has no byte order mark — reading the file
			// consumed it — so put it back before fixing. Output is what the
			// JS side writes to disk, and it has to be the whole file.
			originalContent := fileDiags[0].SourceFile.Text()
			if utils.SourceHasBOM(programSession.FS(), filePath) {
				originalContent = utils.BOM + originalContent
			}
			fixedContent, _, didFix := linter.ApplyRuleFixes(originalContent, fileDiags)
			if didFix {
				output[tspath.ConvertToRelativePath(filePath, comparePathOptions)] = fixedContent
			}
		}
		if len(output) == 0 {
			output = nil
		}
	}

	// The files actually linted (target discovery already excluded global
	// ignores and gitignore entries). sourceFiles was populated by
	// GetRulesForFile for every linted file under its caller-visible target
	// path, relative to configDirectory. This keeps
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
