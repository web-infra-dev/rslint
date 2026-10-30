package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintLoader "github.com/web-infra-dev/rslint/internal/linter/loader"
	"github.com/web-infra-dev/rslint/internal/program/contentmappers"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func (h *Handler) handleLint(ctx context.Context, req api.LintRequest, dispatch linter.EslintPluginDispatcher, requester api.Requester) (*api.LintResponse, error) {

	// Resolve the working directory WITHOUT os.Chdir: this is a long-lived,
	// reused --api process, so mutating the process-global cwd would leak
	// across requests (and race a future concurrent mode). Everything
	// downstream (resolveRequestPath / config resolution / CreateCompilerHost /
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
	programBaseFS := fs
	var allowedFiles []string
	seenAllowedFiles := make(map[string]struct{})

	resolveRequestPath := func(filePath string) string {
		if tspath.PathIsAbsolute(filePath) {
			return tspath.NormalizePath(filePath)
		}
		return tspath.ResolvePath(currentDirectory, filePath)
	}
	if req.Patterns != nil && (req.Files != nil || req.CanonicalFiles != nil || req.FileScanDirectories != nil ||
		(req.ConfigDiscovery != nil && (req.ConfigDiscovery.Directories != nil || req.ConfigDiscovery.ExplicitFiles != nil))) {
		return nil, errors.New("patterns and exact-file expansion metadata are mutually exclusive")
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

	if req.Files != nil || req.Patterns != nil {
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
			if req.Files == nil && req.Patterns == nil {
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
	// Inline API config has no module directory. Its matching/path base can
	// differ from the invocation cwd and must not become the discovery root.
	if len(rslintConfig) > 0 {
		rslintConfig = rslintconfig.ConfigWithResolvedBasePaths(
			rslintConfig,
			configDirectory,
		)
	}
	if len(fileContents) > 0 {
		addEquivalentFileContentPaths(fileContents, configDirectory, currentDirectory, fs, canonicalPaths)
		fs = utils.NewOverlayVFS(fs, fileContents)
	}
	var inputDirectories, globRoots []string
	scanDirectories := make(map[string][]string)
	if req.Patterns != nil {
		// Raw patterns enumerate the request's filesystem. FileContents remains
		// a read overlay, including unselected dependencies and virtual exact
		// targets; it does not add entries to the pattern-search namespace.
		selection, err := target.SelectInputs(ctx, programBaseFS, currentDirectory, req.Patterns)
		if err != nil {
			return nil, fmt.Errorf("select lint inputs: %w", err)
		}
		allowedFiles = selection.Files
		inputDirectories, globRoots = selection.Directories, selection.GlobRoots
		scanDirectories = selection.ScanDirectories
	}
	if req.FileScanDirectories != nil && len(req.FileScanDirectories) != len(req.Files) {
		return nil, errors.New("fileScanDirectories must be parallel to files")
	}
	explicitFiles := make(map[string]bool)
	seenGlobRoots := make(map[string]bool)
	// Duplicate exact paths contribute a union of scan entries. A literal
	// occurrence still overrides every scanned occurrence below.
	type scanEntry struct{ file, root string }
	seenScanEntries := make(map[scanEntry]bool)
	for index, path := range req.Files {
		path = resolveRequestPath(path)
		if req.FileScanDirectories != nil {
			if req.FileScanDirectories[index] == nil {
				explicitFiles[path] = true
			} else {
				roots := scanDirectories[path]
				if roots == nil {
					// Empty means no scan origins; nil would borrow every containing root.
					roots = make([]string, 0, len(req.FileScanDirectories[index]))
				}
				for _, root := range req.FileScanDirectories[index] {
					root = resolveRequestPath(root)
					entry := scanEntry{path, root}
					if !seenScanEntries[entry] {
						seenScanEntries[entry] = true
						roots = append(roots, root)
					}
					if !seenGlobRoots[root] {
						seenGlobRoots[root] = true
						globRoots = append(globRoots, root)
					}
				}
				scanDirectories[path] = roots
			}
		}
	}
	identityFS := fs
	if len(canonicalPaths) > 0 {
		fs = &canonicalPathVFS{FS: fs, canonicalPaths: canonicalPaths}
	}
	var (
		configMap              map[string]rslintconfig.RslintConfig
		configTargetScopes     map[string]target.OwnerScope
		catalogPlugins         []rslintconfig.EslintPluginEntry
		pluginConfigKeyByOwner map[string]string
		gitPolicy              *rslintconfig.GitPolicy
		pathSpaces             *rslintconfig.PathSpaceSnapshot
	)
	var targetFiles []rslintconfig.PathIdentity
	if allowedFiles != nil {
		targetFiles = make([]rslintconfig.PathIdentity, 0, len(allowedFiles))
		for _, path := range allowedFiles {
			identity := rslintconfig.FreezeFileIdentity(path, identityFS)
			if canonicalPath := canonicalPaths[rslintconfig.ExactPathID(path)]; canonicalPath != "" {
				// The host's file hint is independent of the containing directory.
				identity.CanonicalPath = canonicalPath
			}
			targetFiles = append(targetFiles, identity)
		}
	}
	targetDirectories := rslintconfig.ResolveDirectoryIdentities(inputDirectories, fs)
	discoveryDirectories := append([]rslintconfig.DirectoryIdentity(nil), targetDirectories...)
	discoveryDirectories = append(discoveryDirectories, rslintconfig.ResolveDirectoryIdentities(globRoots, fs)...)
	var recursiveDirectories []rslintconfig.DirectoryIdentity
	limitDirectoryWalk := len(globRoots) > 0
	if limitDirectoryWalk {
		recursiveDirectories = targetDirectories
	}
	if req.ConfigDiscovery != nil {
		for _, directory := range req.ConfigDiscovery.Directories {
			discoveryDirectories = append(discoveryDirectories, rslintconfig.ResolveDirectoryIdentities([]string{resolveRequestPath(directory)}, fs)...)
		}
		limitDirectoryWalk = limitDirectoryWalk || len(req.ConfigDiscovery.Directories) > 0
		if len(req.ConfigDiscovery.ExplicitFiles) > 0 {
			if len(req.ConfigDiscovery.ExplicitFiles) != len(req.Files) {
				return nil, errors.New("configDiscovery.explicitFiles must be parallel to files")
			}
			if req.FileScanDirectories == nil {
				for index, path := range req.Files {
					path = resolveRequestPath(path)
					if req.ConfigDiscovery.ExplicitFiles[index] {
						explicitFiles[path] = true
					} else {
						scanDirectories[path] = nil
					}
				}
			}
		}
	}
	for path := range explicitFiles {
		delete(scanDirectories, path)
	}
	var discoveryFiles []discovery.DiscoveryFile
	if targetFiles != nil {
		discoveryFiles = make([]discovery.DiscoveryFile, 0, len(targetFiles))
	}
	for _, identity := range targetFiles {
		roots, scanned := scanDirectories[identity.Path]
		discoveryFiles = append(discoveryFiles, discovery.DiscoveryFile{
			PathIdentity: identity, Explicit: !scanned, ScanDirectories: roots,
		})
	}
	if req.ConfigDiscovery == nil && (req.Patterns != nil || req.FileScanDirectories != nil) {
		// Inline arrays use the same discovery scope without a module loader.
		catalog, err := discovery.LoadExplicitConfig(ctx, fs, nil, discovery.ExplicitConfigRequest{
			CWD:                       currentDirectory,
			Config:                    append(rslintconfig.RslintConfig{}, rslintConfig...),
			ConfigDirectory:           configDirectory,
			TargetFiles:               discoveryFiles,
			TargetDirectories:         discoveryDirectories,
			RecursiveDirectories:      recursiveDirectories,
			LimitDirectoryWalkToFiles: limitDirectoryWalk,
		})
		if err != nil {
			return nil, fmt.Errorf("resolve inline config inputs: %w", err)
		}
		gitPolicy, pathSpaces, targetFiles = catalog.GitPolicy, catalog.PathSpaces, catalog.TargetFiles
	}
	if configDiscovery := req.ConfigDiscovery; configDiscovery != nil {
		if requester == nil {
			return nil, errors.New("configDiscovery requires a bidirectional API host")
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
			Files:                     discoveryFiles,
			Directories:               discoveryDirectories,
			RecursiveDirectories:      recursiveDirectories,
			FallbackConfig:            rslintconfig.ConfigWithResolvedBasePaths(overrideConfig, currentDirectory),
			Fresh:                     true,
			LimitDirectoryWalkToFiles: limitDirectoryWalk,
			ImplicitCWD:               allowedFiles == nil && len(discoveryDirectories) == 0,
		}

		loader := &apiConfigModuleLoader{
			requester:      requester,
			overrideConfig: overrideConfig,
		}
		var configCatalog *discovery.ConfigCatalog
		var err error
		if configDiscovery.ExplicitConfigPath != "" {
			configCatalog, err = discovery.LoadExplicitConfig(
				ctx,
				fs,
				loader,
				discovery.ExplicitConfigRequest{
					CWD:                       currentDirectory,
					ConfigPath:                resolveRequestPath(configDiscovery.ExplicitConfigPath),
					TargetFiles:               discoveryRequest.Files,
					TargetDirectories:         discoveryRequest.Directories,
					RecursiveDirectories:      discoveryRequest.RecursiveDirectories,
					LimitDirectoryWalkToFiles: discoveryRequest.LimitDirectoryWalkToFiles,
					Fresh:                     true,
				},
			)
		} else {
			configCatalog, err = discovery.DiscoverAutomatic(ctx, fs, loader, discoveryRequest)
		}
		if err != nil {
			return nil, fmt.Errorf("discover config catalog: %w", err)
		}
		if failure := configCatalog.AllConfigsFailedError(); failure != nil {
			return nil, fmt.Errorf("discover config catalog: %w", failure)
		}
		gitPolicy = configCatalog.GitPolicy
		pathSpaces = configCatalog.PathSpaces
		targetFiles = configCatalog.TargetFiles
		if targetFiles == nil {
			targetDirectories = configCatalog.TargetDirectories
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
			// syntax-only with an empty config).
			rslintConfig = rslintconfig.ConfigWithResolvedBasePaths(
				overrideConfig,
				currentDirectory,
			)
			configDirectory = currentDirectory
		}
		catalogPlugins = configCatalog.EslintPlugins
	}

	if gitPolicy == nil {
		if targetFiles == nil {
			targetDirectories = rslintconfig.ResolveDirectoryIdentities([]string{currentDirectory}, fs)
		}
		gitPolicy = rslintconfig.CollectGitPolicy(rslintConfig, configDirectory, fs, targetFiles, targetDirectories)
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
	if pathSpaces != nil {
		configsForBases := configMap
		if configsForBases == nil {
			configsForBases = map[string]rslintconfig.RslintConfig{configDirectory: rslintConfig}
		}
		pathSpaces = pathSpaces.WithConfigBases(configsForBases, fs)
	}
	targetPlan, err := target.Resolve(target.Request{
		ConfigMap:       configMap,
		Config:          rslintConfig,
		ConfigDirectory: configDirectory,
		ScanRoot:        currentDirectory,
		OwnerScopes:     configTargetScopes,
		FS:              fs,
		Files:           targetFiles,
		ScanDirectories: scanDirectories,
		Directories:     targetDirectories,
		GitPolicy:       gitPolicy,
		PathSpaces:      pathSpaces,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve lint targets: %w", err)
	}
	configResolver := configLint.NewResolver(configLint.ResolverOptions{
		ConfigsByOwner:       configMap,
		Config:               rslintConfig,
		ConfigDirectory:      configDirectory,
		DefaultRootDirectory: currentDirectory,
		Catalog:              ruleCatalog,
		PathSpaces:           targetPlan.PathSpaces(),
		GitPolicy:            targetPlan.GitPolicy(),
		FS:                   fs,
	})
	projectPolicies, err := configResolver.ProjectPolicies(targetPlan.Files, false)
	if err != nil {
		return nil, err
	}
	projectConfigs := configMap
	if projectConfigs == nil {
		projectConfigs = map[string]rslintconfig.RslintConfig{configDirectory: rslintConfig}
	}
	projectRequest := loader.ProjectBuildRequest{
		Configs:  projectConfigs,
		Targets:  targetPlan,
		Policies: projectPolicies,
		Scope:    loader.Targeted,
	}
	// Keep exact request buffers above physical IO projection, including when
	// callers supply different contents under two spellings of the same file.
	// Discovery uses its composed view; Program caches see this final view.
	sourceFS, err := loader.WithFrozenIO(programBaseFS, targetPlan.PathSpaces(), targetPlan.Files)
	if err != nil {
		return nil, err
	}
	if len(canonicalPaths) > 0 {
		sourceFS = &canonicalPathVFS{FS: sourceFS, canonicalPaths: canonicalPaths}
	}
	if len(fileContents) > 0 {
		sourceFS = utils.NewOverlayVFS(sourceFS, fileContents)
	}
	// The request owns its mapper processes: fix rounds share them, and they
	// stop when the request ends, so no state outlives it.
	mappers := contentmappers.New(contentmappers.Options{Enabled: req.RunExternalCode})
	defer func() {
		mappers.WriteIgnoredWarnings(os.Stderr, "the runExternalCode option")
		_ = mappers.Close()
	}()
	programSession := loader.NewSession(sourceFS, mappers)
	// The Session whose mapper projects are still open. Each fix round
	// releases the previous round's projects once its own are open.
	mapperSession := programSession
	// A plain API lint only needs type information when at least one target is
	// selected. The stable target/config plan is reused while each source
	// snapshot gets a fresh request-local Program generation.
	buildBinding := func(session *loader.Session) (loader.LoadResult, error) {
		projects, buildErr := session.BuildProjects(projectRequest)
		if buildErr != nil {
			return loader.LoadResult{}, buildErr
		}
		binding, err := session.LoadAPI(projects, targetPlan, configDirectory, false)
		session.ReleaseBuildCache()
		return binding, err
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
			dispatch = linter.EslintPluginDispatchFunc(func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
				return nil, errors.New("bidirectional pluginLint transport is unavailable")
			})
		}
	}
	generationForBinding := func(binding loader.LoadResult, generationFS vfs.FS) linter.Generation {
		fileConfigResolver := configResolver.WithSourceMappings(binding.LintTargetBySourcePath, generationFS, true)
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
		initial: func(context.Context) (linter.Generation, error) {
			binding, err := buildBinding(programSession)
			if err != nil {
				return linter.Generation{}, err
			}
			return generationForBinding(binding, programSession.FS()), nil
		},
		rebuild: func(ctx context.Context, snapshot linter.SourceSnapshot) (linter.Generation, error) {
			if err := ctx.Err(); err != nil {
				return linter.Generation{}, err
			}
			files := snapshot.Files()
			overrides := make(map[string]string, len(files)*2)
			for _, file := range files {
				overrides[tspath.NormalizePath(file.Path)] = file.Text
			}
			addEquivalentFileContentPaths(overrides, configDirectory, currentDirectory, sourceFS, canonicalPaths)
			roundSession := loader.NewSession(utils.NewOverlayVFS(sourceFS, overrides), mappers)
			roundBinding, rebuildErr := buildBinding(roundSession)
			if rebuildErr != nil {
				_ = roundSession.ReleaseContentMapperProjects()
				return linter.Generation{}, rebuildErr
			}
			_ = mapperSession.ReleaseContentMapperProjects()
			mapperSession = roundSession
			if err := ctx.Err(); err != nil {
				return linter.Generation{}, err
			}
			return generationForBinding(roundBinding, roundSession.FS()), nil
		},
	}
	if !req.Fix {
		provider.loads = func(ctx context.Context) ([]linter.GenerationLoad, error) {
			loads, err := programSession.PlanLoads(projectRequest, configDirectory, true)
			if err != nil {
				return nil, err
			}
			return lintLoader.GenerationLoads(loads, programSession.FS(), generationForBinding), nil
		}
	}

	demand := linter.ArtifactDemand{
		Native:      rule.EditDemandAll,
		Plugin:      rule.EditDemandAll,
		LintedFiles: true,
		SourceFiles: req.IncludeEncodedSourceFiles,
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
		var loadError *linter.GenerationLoadError
		if _, fixing := pipelineResult.AppliedFixes(); !fixing && errors.As(err, &loadError) {
			return nil, loadError.Err
		}
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
