package main

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
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/inspector"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

type canonicalPathVFS struct {
	vfs.FS
	canonicalPaths map[string]string
}

func (fs *canonicalPathVFS) Realpath(filePath string) string {
	if canonicalPath := fs.canonicalPaths[exactFilesystemPathID(filePath)]; canonicalPath != "" {
		return canonicalPath
	}
	return fs.FS.Realpath(filePath)
}

// programCache holds a cached Program instance for AST info requests
type programCache struct {
	mu              sync.RWMutex
	fileContent     string
	compilerOptions string // JSON serialized for comparison
	program         *compiler.Program
	sourceFile      *ast.SourceFile
}

// Global program cache for AST info requests
var astInfoProgramCache = &programCache{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {
	return h.handleLint(context.Background(), req, nil, nil)
}

type apiConfigModuleLoader struct {
	requester api.Requester
}

func (loader *apiConfigModuleLoader) LoadConfigs(ctx context.Context, request rslintconfig.ConfigLoadBatchRequest) (rslintconfig.ConfigLoadBatchResponse, error) {
	if loader == nil || loader.requester == nil {
		return rslintconfig.ConfigLoadBatchResponse{}, errors.New("reverse config loading is unavailable")
	}
	msg, err := loader.requester.SendRequest(ctx, api.KindLoadConfigs, request)
	if err != nil {
		return rslintconfig.ConfigLoadBatchResponse{}, err
	}
	var response rslintconfig.ConfigLoadBatchResponse
	if err := msg.Decode(&response); err != nil {
		return rslintconfig.ConfigLoadBatchResponse{}, fmt.Errorf("decode loadConfigs result: %w", err)
	}
	return response, nil
}

func (loader *apiConfigModuleLoader) ActivateConfigs(ctx context.Context, request rslintconfig.ConfigActivationRequest) (rslintconfig.ConfigActivationResponse, error) {
	if loader == nil || loader.requester == nil {
		return rslintconfig.ConfigActivationResponse{}, errors.New("reverse config activation is unavailable")
	}
	msg, err := loader.requester.SendRequest(ctx, api.KindActivateConfigs, request)
	if err != nil {
		return rslintconfig.ConfigActivationResponse{}, err
	}
	var response rslintconfig.ConfigActivationResponse
	if err := msg.Decode(&response); err != nil {
		return rslintconfig.ConfigActivationResponse{}, fmt.Errorf("decode activateConfigs result: %w", err)
	}
	return response, nil
}

// HandleLintWithContext enables reverse pluginLint requests when IPCHandler is
// hosted by the bidirectional API service. HandleLint remains available for
// direct/legacy callers that do not need community plugin execution.
func (h *IPCHandler) HandleLintWithContext(ctx context.Context, req api.LintRequest, requester api.Requester) (*api.LintResponse, error) {
	var dispatch linter.EslintPluginDispatcher
	if requester != nil {
		dispatch = func(reqCtx context.Context, pluginReq linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			msg, err := requester.SendRequest(reqCtx, api.KindPluginLint, pluginReq)
			if err != nil {
				return nil, err
			}
			var result linter.EslintPluginLintResult
			if err := msg.Decode(&result); err != nil {
				return nil, fmt.Errorf("decode pluginLint result: %w", err)
			}
			return &result, nil
		}
	}
	return h.handleLint(ctx, req, dispatch, requester)
}

func (h *IPCHandler) handleLint(ctx context.Context, req api.LintRequest, dispatch linter.EslintPluginDispatcher, requester api.Requester) (*api.LintResponse, error) {

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
		canonicalPaths[exactFilesystemPathID(filePath)] = canonicalPath
		canonicalPaths[exactFilesystemPathID(canonicalPath)] = canonicalPath
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

	// Config is the legacy/low-level already-resolved config. High-level native
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
		if optionsErrs := rslintconfig.ValidateRuleOptions(rslintConfig, rslintconfig.GlobalRuleRegistry); len(optionsErrs) > 0 {
			msgs := make([]string, len(optionsErrs))
			for i, optionsErr := range optionsErrs {
				msgs[i] = optionsErr.Error()
			}
			return nil, fmt.Errorf("invalid rule options:\n%s", strings.Join(msgs, "\n"))
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

	var (
		configMap          map[string]rslintconfig.RslintConfig
		configTargetScopes map[string]rslintconfig.LintDiscoveryScope
		catalogPlugins     []rslintconfig.EslintPluginEntry
		originalConfigDir  map[string]string
	)
	if discovery := req.ConfigDiscovery; discovery != nil {
		if requester == nil {
			return nil, errors.New("configDiscovery requires a bidirectional API host")
		}
		if len(discovery.ExplicitFiles) > 0 && len(discovery.ExplicitFiles) != len(req.Files) {
			return nil, errors.New("configDiscovery.explicitFiles must be parallel to files")
		}

		discoveryRequest := rslintconfig.ConfigDiscoveryRequest{
			CWD:                       currentDirectory,
			Fresh:                     true,
			LimitDirectoryWalkToFiles: len(discovery.Directories) > 0,
			ImplicitCWD:               len(req.Files) == 0 && len(discovery.Directories) == 0,
			SingleThreaded:            false,
		}
		switch discovery.Mode {
		case "", "auto":
			discoveryRequest.Mode = rslintconfig.ConfigDiscoveryAuto
		case "explicit":
			discoveryRequest.Mode = rslintconfig.ConfigDiscoveryExplicit
			discoveryRequest.ExplicitConfigPath = resolveRequestPath(discovery.ExplicitConfigPath)
		default:
			return nil, fmt.Errorf("invalid configDiscovery mode %q", discovery.Mode)
		}
		for _, directory := range discovery.Directories {
			discoveryRequest.Directories = append(discoveryRequest.Directories, resolveRequestPath(directory))
		}
		for index, filePath := range req.Files {
			explicit := true
			if len(discovery.ExplicitFiles) > 0 {
				explicit = discovery.ExplicitFiles[index]
			}
			canonicalPath := ""
			if len(req.CanonicalFiles) > 0 {
				canonicalPath = resolveRequestPath(req.CanonicalFiles[index])
			}
			discoveryRequest.Files = append(discoveryRequest.Files, rslintconfig.DiscoveryFile{
				Path:          resolveRequestPath(filePath),
				CanonicalPath: canonicalPath,
				Explicit:      explicit,
			})
		}

		catalog, err := rslintconfig.NewConfigDiscoverySession(
			fs,
			&apiConfigModuleLoader{requester: requester},
		).Build(ctx, discoveryRequest)
		if err != nil {
			return nil, fmt.Errorf("discover config catalog: %w", err)
		}
		if len(catalog.EslintPlugins) > 0 {
			if capabilityRequester, ok := requester.(api.PeerCapabilityRequester); ok &&
				!capabilityRequester.PeerSupportsCapability(api.CapabilityReversePluginLint) {
				return nil, errors.New("API peer does not advertise reversePluginLint capability required by discovered ESLint plugins")
			}
		}

		var overrideConfig rslintconfig.RslintConfig
		if len(discovery.OverrideConfig) > 0 {
			if err := json.Unmarshal(discovery.OverrideConfig, &overrideConfig); err != nil {
				return nil, fmt.Errorf("invalid configDiscovery.overrideConfig: %w", err)
			}
			if err := rslintconfig.ValidateConfig(overrideConfig); err != nil {
				return nil, fmt.Errorf("invalid configDiscovery.overrideConfig: %w", err)
			}
		}

		if len(catalog.Configs) > 0 {
			configDirectories := catalog.ConfigDirectories()
			if len(configDirectories) == 1 && catalog.Sources[configDirectories[0]].ExplicitConfig {
				// overrideConfigFile is invocation-wide. A hierarchical config map
				// would have no owner for a requested file outside cwd and would
				// incorrectly drop it, even though explicit flat-config semantics say
				// the selected module governs the complete supplied target set.
				configDirectory = configDirectories[0]
				rslintConfig = append(rslintconfig.RslintConfig(nil), catalog.Configs[configDirectory]...)
				rslintConfig = append(rslintConfig, overrideConfig...)
				originalConfigDir = map[string]string{configDirectory: configDirectory}
			} else {
				configMap = make(map[string]rslintconfig.RslintConfig, len(catalog.Configs))
				originalConfigDir = make(map[string]string, len(catalog.Configs))
			}
			for ownerDirectory, entries := range catalog.Configs {
				if configMap == nil {
					continue
				}
				effective := make(rslintconfig.RslintConfig, 0, len(entries)+len(overrideConfig))
				effective = append(effective, entries...)
				effective = append(effective, overrideConfig...)
				configMap[ownerDirectory] = effective
				originalConfigDir[ownerDirectory] = ownerDirectory
			}
			if configMap != nil {
				configTargetScopes = catalog.Scopes
				// The API already has an exact target set. Resolve provisional owners
				// with authored config ignores first, then read only each owner's
				// target ancestor chains. Every loaded config remains a source boundary,
				// including one that owns only explicit files.
				filesByOwner := configTargetFilesByOwner(
					configMap,
					configTargetScopes,
					fs,
					allowedFiles,
					false,
				)
				configResolver := rslintconfig.NewConfigOwnerResolver(configMap, fs)
				for ownerDirectory, entries := range configMap {
					ownerFiles := filesByOwner[ownerDirectory]
					if ownerFiles == nil {
						ownerFiles = []string{}
					}
					configMap[ownerDirectory] = rslintconfig.ConfigWithGitignoreWithBoundaries(
						entries,
						ownerDirectory,
						fs,
						ownerFiles,
						configResolver.ChildConfigDirs(ownerDirectory),
					)
				}
			}
		} else {
			// No JS candidate is a valid API state: lint with override entries (or
			// syntax-only with an empty config) rather than falling back to the CLI's
			// rslint.json lookup.
			rslintConfig = overrideConfig
			configDirectory = currentDirectory
		}
		catalogPlugins = catalog.EslintPlugins
	}

	if configMap == nil {
		rslintConfig = rslintconfig.ConfigWithGitignore(rslintConfig, configDirectory, fs, allowedFiles)
	}

	// The registry is process-global, but plugin execution is request-gated by
	// requestPluginRules below so metadata from an earlier API request cannot
	// make a later request dispatch stale plugin rules.
	rslintconfig.RegisterAllRules()
	pluginEntries := append([]rslintconfig.EslintPluginEntry(nil), catalogPlugins...)
	for _, plugin := range req.EslintPlugins {
		pluginEntries = append(pluginEntries, rslintconfig.EslintPluginEntry{
			Prefix:    plugin.Prefix,
			RuleNames: append([]string(nil), plugin.RuleNames...),
		})
	}
	var requestPluginRules map[string]struct{}
	if len(pluginEntries) > 0 {
		requestPluginRules = make(map[string]struct{})
		for _, plugin := range pluginEntries {
			for _, ruleName := range plugin.RuleNames {
				requestPluginRules[plugin.Prefix+"/"+ruleName] = struct{}{}
			}
		}
		rslintconfig.RegisterEslintPluginRules(pluginEntries)
	}

	responsePathBase := configDirectory
	if configMap != nil {
		responsePathBase = currentDirectory
	}
	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          responsePathBase,
		UseCaseSensitiveFileNames: true,
	}

	// Resolve the exact lint target set, bind targets to existing Programs, and
	// append a non-project-backed fallback Program for selected files absent from
	// every Program (the typical lintText/lintFiles in-memory file). Identical to
	// the CLI via the shared helper. Native discovery can supply a hierarchical
	// configMap; low-level config and explicit overrideConfigFile requests remain
	// invocation-wide single-config paths.
	// The --api path never runs the type-check phase (RunLinterOptions.TypeCheck
	// stays false), so there is no per-program type-check skip mask to build.
	targetPlan, err := resolveLintTargetPlan(configMap, rslintConfig, configDirectory, configTargetScopes, fs, allowedFiles, nil, false)
	if err != nil {
		return nil, fmt.Errorf("resolve lint targets: %w", err)
	}
	// A plain API lint only needs type information when at least one target is
	// selected. Resolve the target plan before project paths so an ignored or
	// empty request cannot fail on an inactive project declaration.
	parseCache := utils.NewParseCache()
	var programSet lintProgramSet
	if len(targetPlan.Targets) > 0 {
		if configMap != nil {
			programSet, err = createProgramSetForConfigs(configsForLintTargetPlan(configMap, targetPlan), false, fs, parseCache)
		} else {
			programSet, err = createProgramSetForConfig(configDirectory, rslintConfig, false, fs, parseCache)
		}
		if err != nil {
			return nil, err
		}
	}
	binding, err := bindLintTargetPlan(programSet, targetPlan, configDirectory, fs, parseCache, false)
	if err != nil {
		return nil, err
	}
	programs := binding.Programs
	typeInfoFiles := binding.TypeInfoFiles
	targetsByProgram := binding.TargetsByProgram
	targetPathBySourcePath := binding.TargetPathBySourcePath
	fileConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap:                  configMap,
		Config:                     rslintConfig,
		CurrentDirectory:           configDirectory,
		EnforcePlugins:             true,
		TypeInfoFiles:              typeInfoFiles,
		ConfigPathBySourcePath:     binding.ConfigPathBySourcePath,
		OwnerConfigDirBySourcePath: binding.OwnerConfigDirBySourcePath,
		SourceMappingsCanonical:    true,
		FS:                         fs,
	})
	targetPathForSourcePath := func(sourcePath string) string {
		if targetPath := targetPathBySourcePath[sourcePath]; targetPath != "" {
			return targetPath
		}
		return sourcePath
	}
	responsePathForSourcePath := func(sourcePath string) string {
		return tspath.ConvertToRelativePath(targetPathForSourcePath(sourcePath), comparePathOptions)
	}

	// Collect diagnostics and source files
	var diagnostics []api.Diagnostic
	var diagnosticsLock sync.Mutex
	errorsCount := 0
	warningsCount := 0
	fixableErrorsCount := 0
	fixableWarningsCount := 0
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
		if d.SourceFile != nil {
			sourceFilesLock.Lock()
			filePath := responsePathForSourcePath(d.FilePath)
			if sf, ok := d.SourceFile.(*ast.SourceFile); ok {
				sourceFiles[filePath] = sf
			}
			sourceFilesLock.Unlock()
		}

		diagnosticStart := d.Range.Pos()
		diagnosticEnd := d.Range.End()

		startLine, startColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticStart)
		endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(d.SourceFile, diagnosticEnd)

		diagnostic := api.Diagnostic{
			RuleName:  d.RuleName,
			MessageId: d.Message.Id,
			Message:   d.Message.Description,
			FilePath:  responsePathForSourcePath(d.FilePath),
			Range: api.Range{
				Start: api.Position{
					Line:   startLine + 1, // Convert to 1-based indexing
					Column: int(startColumn) + 1,
				},
				End: api.Position{
					Line:   endLine + 1,
					Column: int(endColumn) + 1,
				},
			},
			Severity: d.Severity.String(),
		}

		// Fix and suggestion ranges are flat UTF-16 offsets (ESLint's unit),
		// converted from the rule's byte offsets via byteOffsetToUTF16. This is
		// a DIFFERENT conversion than the line/column above (which counts UTF-16
		// units from the line start, not a flat file offset).
		fixText := d.SourceFile.Text()

		// Add fixes if available.
		if d.FixesPtr != nil && len(*d.FixesPtr) > 0 {
			var fixes []api.Fix
			for _, fix := range *d.FixesPtr {
				fixes = append(fixes, api.Fix{
					Text:     fix.Text,
					StartPos: byteOffsetToUTF16(fixText, fix.Range.Pos()),
					EndPos:   byteOffsetToUTF16(fixText, fix.Range.End()),
				})
			}
			diagnostic.Fixes = fixes
		}

		// Add suggestions if available — optional, user-selected fixes the
		// editor surfaces (distinct from auto-applied Fixes).
		if d.Suggestions != nil && len(*d.Suggestions) > 0 {
			suggestions := make([]api.Suggestion, 0, len(*d.Suggestions))
			for _, sug := range *d.Suggestions {
				var fixes []api.Fix
				for _, fix := range sug.FixesArr {
					fixes = append(fixes, api.Fix{
						Text:     fix.Text,
						StartPos: byteOffsetToUTF16(fixText, fix.Range.Pos()),
						EndPos:   byteOffsetToUTF16(fixText, fix.Range.End()),
					})
				}
				suggestions = append(suggestions, api.Suggestion{
					MessageId: sug.Message.Id,
					Message:   sug.Message.Description,
					Data:      sug.Message.Data,
					Fixes:     fixes,
				})
			}
			diagnostic.Suggestions = suggestions
		}

		diagnostics = append(diagnostics, diagnostic)

		// Split counts by severity (ESLint semantics): errorCount counts errors
		// only, not the total. fixable*Count counts the fixable subset.
		hasFix := d.FixesPtr != nil && len(*d.FixesPtr) > 0
		switch d.Severity {
		case rule.SeverityError:
			errorsCount++
			if hasFix {
				fixableErrorsCount++
			}
		case rule.SeverityWarning:
			warningsCount++
			if hasFix {
				fixableWarningsCount++
			}
		}

		// Retain the original diagnostic (byte-offset fixes + SourceFile) for the
		// in-band fix pass, grouped by the caller-visible target path.
		if req.Fix && hasFix {
			targetPath := targetPathForSourcePath(d.FilePath)
			diagnosticsByFile[targetPath] = append(diagnosticsByFile[targetPath], d)
		}
	}

	// Every selected target is parsed even when no config entry contributes
	// rules. Global ignores were already removed during target discovery.
	syntaxDiagnostics, syntaxErrorFiles := collectTargetSyntacticDiagnostics(programs, targetsByProgram, nil, false, false)
	for _, diagnostic := range syntaxDiagnostics {
		diagnosticCollector(diagnostic)
	}

	// Build one run descriptor shared by native lint and plugin target
	// collection, keeping both paths on the exact same file/rule selection.
	runOpts := linter.RunLinterOptions{
		Programs:       programs,
		SingleThreaded: false, // Don't use single-threaded mode for IPC
		Scope:          linter.FileScope{Files: allowedFiles},
		TargetFiles:    targetsByProgram,
		// RunLinter repeats the RequiresTypeInfo eligibility check for files
		// outside this set and withholds the project TypeChecker from them.
		TypeInfoFiles:    typeInfoFiles,
		SyntaxErrorFiles: syntaxErrorFiles,
		GetRulesForFile: func(sourceFile *ast.SourceFile) []linter.ConfiguredRule {
			// Track source file for encoding
			sourceFilesLock.Lock()
			filePath := responsePathForSourcePath(sourceFile.FileName())
			sourceFiles[filePath] = sourceFile
			sourceFilesLock.Unlock()

			// GetActiveRulesForFile applies the type-aware gate: when
			// typeInfoFiles is non-nil and this file is not in it (a gap /
			// fallback file with no type information), type-aware rules are
			// filtered out. RunLinter repeats this check at the execution boundary.
			// typeInfoFiles==nil ⇒ no fallback ⇒ every linted file has type
			// info ⇒ nothing to filter. Rules come solely from the resolved
			// config object (config.rules); --api has no separate rule-options
			// surface.
			//
			// enforcePlugins=true: the --api config is a resolved JS-style flat
			// config (plugins + rules), exactly like the CLI's JS/TS config path,
			// so a rule carrying a plugin prefix runs only when its plugin is
			// declared in the config's `plugins` — matching CLI and ESLint
			// semantics (a rule whose plugin is not declared is skipped).
			activeRules := fileConfigResolver.ActiveRulesForFile(sourceFile.FileName())
			// Plugin placeholders live in the process-global registry. Restrict
			// them to this request's metadata so a long-lived API process cannot
			// leak a plugin registered by an earlier lint into a later request.
			for i, configuredRule := range activeRules {
				if !configuredRule.IsEslintPluginRule {
					continue
				}
				if _, ok := requestPluginRules[configuredRule.Name]; ok {
					continue
				}
				filtered := make([]linter.ConfiguredRule, 0, len(activeRules)-1)
				filtered = append(filtered, activeRules[:i]...)
				for _, remainingRule := range activeRules[i+1:] {
					if !remainingRule.IsEslintPluginRule {
						filtered = append(filtered, remainingRule)
						continue
					}
					if _, ok := requestPluginRules[remainingRule.Name]; ok {
						filtered = append(filtered, remainingRule)
					}
				}
				return filtered
			}
			return activeRules
		},
		OnDiagnostic: diagnosticCollector,
	}

	// Metadata is the feature gate: without it there is no plugin target walk,
	// goroutine, or reverse request. With metadata, dispatch starts before the
	// native pass and runs in parallel, matching the CLI pipeline.
	var pluginCh <-chan []rule.RuleDiagnostic
	var cancelPlugin context.CancelFunc
	if len(pluginEntries) > 0 {
		if originalConfigDir == nil {
			wireConfigDirectory := req.PluginConfigDirectory
			if wireConfigDirectory == "" {
				wireConfigDirectory = req.ConfigDirectory
			}
			if wireConfigDirectory == "" {
				wireConfigDirectory = configDirectory
			}
			originalConfigDir = map[string]string{configDirectory: wireConfigDirectory}
		}
		pluginInputs := buildPluginFileInputs(runOpts, pluginConfigResolver{
			lintResolver:      fileConfigResolver,
			originalConfigDir: originalConfigDir,
		})
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
			pluginCh = dispatchPluginLintAsync(pluginCtx, dispatch, pluginInputs, req.Fix, pluginSuggestionsMode(req.Fix))
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
		for _, diagnostic := range <-pluginCh {
			diagnosticCollector(diagnostic)
		}
	}

	if diagnostics == nil {
		diagnostics = []api.Diagnostic{}
	}
	// Sort diagnostics by (file, start position) only — deliberately NO
	// end/rule tie-break: ESLint and the upstream rule tests order
	// same-start diagnostics by emission order (parent reported before
	// nested child), and a file's diagnostics are all emitted by a single
	// worker, so under a STABLE sort this key is already fully
	// deterministic. Keep this comparator in sync with the CLI one in
	// cmd.go (same policy over rule.RuleDiagnostic).
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.FilePath != b.FilePath {
			return a.FilePath < b.FilePath
		}
		if a.Range.Start.Line != b.Range.Start.Line {
			return a.Range.Start.Line < b.Range.Start.Line
		}
		return a.Range.Start.Column < b.Range.Start.Column
	})

	// Apply fixes in-band when requested. ApplyRuleFixes is the same pure fixer
	// the CLI uses (cmd.go applyFixPass), but here the result stays in-memory in
	// Output — the JS side persists it via Rslint.outputFixes. Single pass over
	// each file's fixes (non-overlapping applied, overlapping left for a later
	// lint); no cross-pass re-lint cascade (P1, see design §8).
	var output map[string]string
	if req.Fix && len(diagnosticsByFile) > 0 {
		output = make(map[string]string)
		for filePath, fileDiags := range diagnosticsByFile {
			originalContent := fileDiags[0].SourceFile.Text()
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
		Diagnostics:         diagnostics,
		ErrorCount:          errorsCount,
		WarningCount:        warningsCount,
		FixableErrorCount:   fixableErrorsCount,
		FixableWarningCount: fixableWarningsCount,
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

func configTargetFilesByOwner(
	configMap map[string]rslintconfig.RslintConfig,
	scopes map[string]rslintconfig.LintDiscoveryScope,
	fs vfs.FS,
	allowedFiles []string,
	singleThreaded bool,
) map[string][]string {
	filesByOwner := make(map[string][]string, len(configMap))
	for _, target := range rslintconfig.DiscoverLintTargetsMultiConfig(
		configMap,
		scopes,
		fs,
		allowedFiles,
		nil,
		singleThreaded,
	) {
		filesByOwner[target.ConfigDirectory] = append(
			filesByOwner[target.ConfigDirectory],
			target.Path,
		)
	}
	return filesByOwner
}

func addEquivalentFileContentPaths(fileContents map[string]string, configDirectory string, currentDirectory string, fs vfs.FS) {
	if len(fileContents) == 0 || fs == nil {
		return
	}

	type fileContentEntry struct {
		path    string
		content string
	}
	entries := make([]fileContentEntry, 0, len(fileContents))
	for filePath, content := range fileContents {
		entries = append(entries, fileContentEntry{path: filePath, content: content})
	}

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: true,
	}
	addAlias := func(alias string, content string) {
		if alias == "" {
			return
		}
		if _, exists := fileContents[alias]; exists {
			return
		}
		fileContents[alias] = content
	}
	addDirectoryAlias := func(fromDir string, toDir string, filePath string, content string) {
		if fromDir == "" || toDir == "" || !tspath.ContainsPath(fromDir, filePath, comparePathOptions) {
			return
		}
		relativePath := tspath.GetRelativePathFromDirectory(fromDir, filePath, comparePathOptions)
		if relativePath == "" {
			return
		}
		addAlias(tspath.ResolvePath(toDir, relativePath), content)
	}

	realConfigDirectory := fs.Realpath(configDirectory)
	for _, entry := range entries {
		if realPath := fs.Realpath(entry.path); realPath != "" && realPath != entry.path {
			addAlias(realPath, entry.content)
		}
		if realConfigDirectory != "" && tspath.ComparePaths(configDirectory, realConfigDirectory, comparePathOptions) != 0 {
			addDirectoryAlias(configDirectory, realConfigDirectory, entry.path, entry.content)
			addDirectoryAlias(realConfigDirectory, configDirectory, entry.path, entry.content)
		}
	}
}

// HandleGetAstInfo handles get AST info requests in IPC mode
func (h *IPCHandler) HandleGetAstInfo(req api.GetAstInfoRequest) (*api.GetAstInfoResponse, error) {
	// Fixed user file name for program creation
	const userFileName = "/index.ts"

	// Serialize compiler options for comparison
	compilerOptionsJSON := "{}"
	if req.CompilerOptions != nil {
		jsonBytes, err := json.Marshal(req.CompilerOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal compiler options: %w", err)
		}
		compilerOptionsJSON = string(jsonBytes)
	}

	// Check if we can use cached program
	program, userSourceFile := getCachedProgram(req.FileContent, compilerOptionsJSON)
	if program == nil || userSourceFile == nil {
		// Cache miss - create new program
		var err error
		program, userSourceFile, err = createAndCacheProgram(userFileName, req.FileContent, compilerOptionsJSON, req.CompilerOptions)
		if err != nil {
			return nil, err
		}
	}

	// Get type checker
	typeChecker, done := program.GetTypeChecker(context.Background())
	defer done()

	// Determine which source file to query
	// If FileName is set to an external file, query that file (e.g., lib.d.ts)
	// Otherwise, query the user's source file
	var targetSourceFile *ast.SourceFile
	if req.FileName != "" && req.FileName != userFileName {
		targetSourceFile = program.GetSourceFile(req.FileName)
		if targetSourceFile == nil {
			return &api.GetAstInfoResponse{}, nil
		}
	} else {
		targetSourceFile = userSourceFile
	}

	isExternalFile := targetSourceFile != userSourceFile

	// Build the response
	// Use userSourceFile as the "current" file for the builder
	// This determines which files are considered "external" (fileName will be set for nodes not in userSourceFile)
	builder := api.NewAstInfoBuilder(typeChecker, userSourceFile)
	response := &api.GetAstInfoResponse{}

	// Special case: if requesting SourceFile by kind, build it directly without Node conversion
	if req.Kind > 0 && ast.Kind(req.Kind) == ast.KindSourceFile {
		response.Node = builder.BuildSourceFileNodeInfo(targetSourceFile)
		// SourceFile doesn't have type/symbol/signature/flow, so return early
		return response, nil
	}

	// Find the node at the specified position (with optional end for exact matching)
	node := inspector.FindNodeAtPosition(targetSourceFile, req.Position, req.End, req.Kind)
	if node == nil {
		return &api.GetAstInfoResponse{}, nil
	}

	// Build node info
	response.Node = builder.BuildNodeInfo(node)

	// Build type info
	t := inspector.GetTypeAtNode(typeChecker, node)
	if t != nil {
		response.Type = builder.BuildTypeInfo(t)
	}

	// Build symbol info
	// First try to get symbol directly from node
	symbol := typeChecker.GetSymbolAtLocation(node)
	// If no symbol at node, try to get it from the type
	if symbol == nil && t != nil {
		symbol = t.Symbol()
	}
	if symbol != nil {
		response.Symbol = builder.BuildSymbolInfo(symbol)
	}

	// Build signature info
	sig := inspector.GetSignatureOfNode(typeChecker, node)
	if sig != nil {
		response.Signature = builder.BuildSignatureInfo(sig)
	}

	// Build flow info (only for nodes in user's source file)
	if !isExternalFile {
		flowNode := inspector.GetFlowNodeOfNode(node)
		if flowNode != nil {
			response.Flow = builder.BuildFlowInfo(flowNode)
		}
	}

	return response, nil
}

// getCachedProgram returns the cached program if it matches the current request
func getCachedProgram(fileContent, compilerOptionsJSON string) (*compiler.Program, *ast.SourceFile) {
	astInfoProgramCache.mu.RLock()
	defer astInfoProgramCache.mu.RUnlock()

	if astInfoProgramCache.program == nil {
		return nil, nil
	}

	// Check if cache is valid (only fileContent and compilerOptions matter)
	if astInfoProgramCache.fileContent == fileContent &&
		astInfoProgramCache.compilerOptions == compilerOptionsJSON {
		return astInfoProgramCache.program, astInfoProgramCache.sourceFile
	}

	return nil, nil
}

// createAndCacheProgram creates a new program and caches it
func createAndCacheProgram(fileName, fileContent, compilerOptionsJSON string, compilerOptions map[string]any) (*compiler.Program, *ast.SourceFile, error) {
	// Create a virtual filesystem with the provided file content
	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))

	fileContents := map[string]string{
		fileName: fileContent,
	}
	fs = utils.NewOverlayVFS(fs, fileContents)

	// Build tsconfig from request options or use defaults
	tsconfigContent := buildTsConfigContent(fileName, compilerOptions)
	tsconfigPath := "/tsconfig.json"
	fs = utils.NewOverlayVFS(fs, map[string]string{
		tsconfigPath: tsconfigContent,
	})

	// Create compiler host and program
	host := utils.CreateCompilerHost("/", fs)
	program, err := utils.CreateProgram(false, fs, "/", tsconfigPath, host)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create program: %w", err)
	}

	// Get the source file
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		return nil, nil, errors.New("failed to get source file")
	}

	// Update cache
	astInfoProgramCache.mu.Lock()
	astInfoProgramCache.fileContent = fileContent
	astInfoProgramCache.compilerOptions = compilerOptionsJSON
	astInfoProgramCache.program = program
	astInfoProgramCache.sourceFile = sourceFile
	astInfoProgramCache.mu.Unlock()

	return program, sourceFile, nil
}

// buildTsConfigContent creates a tsconfig.json content string from compiler options
func buildTsConfigContent(fileName string, compilerOptions map[string]any) string {
	// Default compiler options
	opts := map[string]any{
		"target":           "ESNext",
		"module":           "ESNext",
		"strict":           true,
		"strictNullChecks": true,
	}

	// Merge with provided options (provided options override defaults)
	for k, v := range compilerOptions {
		opts[k] = v
	}

	// Serialize compiler options to JSON
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		// Fallback to minimal config on error
		return fmt.Sprintf(`{"compilerOptions":{"target":"ESNext","module":"ESNext","strict":true},"files":["%s"]}`, fileName)
	}

	return fmt.Sprintf(`{"compilerOptions":%s,"files":["%s"]}`, string(optsJSON), fileName)
}

// byteOffsetToUTF16 converts a byte offset within text to a flat UTF-16 code
// unit offset — the unit ESLint uses for fix / suggestion ranges. This is a
// DIFFERENT conversion than line/column (scanner.GetECMALineAndUTF16CharacterOfPosition,
// which counts UTF-16 units from a line start); fix ranges are flat offsets
// from the start of the file.
func byteOffsetToUTF16(text string, byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset >= len(text) {
		return int(core.UTF16Len(text))
	}
	return int(core.UTF16Len(text[:byteOffset]))
}

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := api.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
