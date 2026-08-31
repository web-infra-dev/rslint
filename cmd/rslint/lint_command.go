package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"slices"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/output"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/term"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func isBroadProjectLoadScope(
	allowFiles []string,
	allowDirs []string,
	currentDirectory string,
	useCaseSensitive bool,
) bool {
	if len(allowDirs) == 0 {
		return len(allowFiles) == 0
	}
	options := tspath.ComparePathsOptions{
		CurrentDirectory:          currentDirectory,
		UseCaseSensitiveFileNames: useCaseSensitive,
	}
	for _, directory := range allowDirs {
		if tspath.ContainsPath(directory, currentDirectory, options) {
			return true
		}
	}
	return false
}

// resolveStartTime returns the start time for timing output.
// If startTimeMs (epoch millis from the Node.js entry point) is positive,
// it is used so the reported duration covers end-to-end execution.
// Otherwise falls back to time.Now().
func resolveStartTime(startTimeMs int64) time.Time {
	if startTimeMs > 0 {
		return time.UnixMilli(startTimeMs)
	}
	return time.Now()
}

// handleLintCommand handles one CLI invocation: it prepares command-owned
// config/targets/Programs, delegates lint execution to linter.RunPipeline, and
// projects the result to the selected output format.
func handleLintCommand(args lintArgs, ctx context.Context, dispatch linter.EslintPluginDispatcher) int {
	// Unpack into locals so the command body below stays focused — only the
	// flag-parse front matter lives in parseLintFlags.
	init := args.Init
	configCatalog := args.ConfigCatalog
	fix := args.Fix
	typeCheck := args.TypeCheck
	typeCheckOnly := args.TypeCheckOnly
	traceOut := args.TraceOut
	cpuprofOut := args.CpuprofOut
	singleThreaded := args.SingleThreaded
	quiet := args.Quiet
	maxWarnings := args.MaxWarnings
	startTimeMs := args.StartTimeMs
	ruleFlags := args.RuleFlags
	allowFiles := args.AllowFiles
	allowDirs := args.AllowDirs
	// --timing enables per-rule timing. One collector is shared across the
	// initial lint and every --fix re-lint pass, so the table reflects total
	// rule cost for the whole run.
	var timingCollector *linter.TimingCollector
	timingLimit := args.TimingLimit
	if args.Timing && !typeCheckOnly {
		timingCollector = linter.NewTimingCollector()
	}
	format := output.FormatDefault
	if !init {
		var formatErr error
		format, formatErr = output.ParseFormat(args.Format)
		if formatErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", formatErr)
			return 2
		}
	}
	mode := output.ModeLint
	if typeCheckOnly {
		mode = output.ModeTypeCheckOnly
	} else if typeCheck {
		mode = output.ModeLintAndTypeCheck
	}

	// Resolve color against the real output destination. In native CLI runs
	// the Go process writes to an IPC pipe, so the Node host supplies the TTY
	// fact for the stdout that ultimately receives forwarded output frames.
	forceColorEnv, forceColorEnvSet := os.LookupEnv("FORCE_COLOR")
	_, noColorEnvSet := os.LookupEnv("NO_COLOR")
	colorEnabled := term.ResolveColorEnabled(term.ColorInputs{
		NoColorFlag:      args.NoColor,
		ForceColorFlag:   args.ForceColor,
		ForceColorEnv:    forceColorEnv,
		ForceColorEnvSet: forceColorEnvSet,
		NoColorEnvSet:    noColorEnvSet,
		GithubActionsEnv: os.Getenv("GITHUB_ACTIONS"),
		Term:             os.Getenv("TERM"),
		StdoutIsTTY:      args.StdoutIsTTY,
	})

	enableVirtualTerminalProcessing()
	timeBefore := resolveStartTime(startTimeMs)

	if traceOut != "" {
		f, err := os.Create(traceOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating trace file: %v\n", err)
			return 1
		}
		defer f.Close()
		trace.Start(f)
		defer trace.Stop()
	}
	if cpuprofOut != "" {
		f, err := os.Create(cpuprofOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating cpuprof file: %v\n", err)
			return 1
		}
		defer f.Close()
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error starting cpu profiling: %v\n", err)
			return 1
		}
		defer pprof.StopCPUProfile()
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
		return 1
	}
	currentDirectory = tspath.NormalizePath(currentDirectory)
	workingDirectory := currentDirectory

	if init {
		if err := rslintconfig.InitDefaultConfig(currentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "error initializing config: %v\n", err)
			return 1
		}
		return 0
	}
	if configCatalog == nil {
		fmt.Fprintln(os.Stderr, "error: config discovery did not produce a catalog")
		return 1
	}
	if !configCatalog.Explicit && len(configCatalog.Configs) == 0 {
		fmt.Fprintln(os.Stderr, "error: no rslint config found; run `rslint --init` to create one")
		return 1
	}

	// Only the production disk-backed CLI shares its initial source medium with
	// the Node plugin host. Injected VFS implementations must send exact text.
	pluginHostReadsInitialText := args.FS == nil
	fs := args.FS
	if fs == nil {
		fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	}
	sourceFS := fs

	// The run-scoped loader owns source snapshots, compiler metadata, Program
	// construction, target binding, and fix-generation rebuilds. Integrations
	// consume only its unified Program results.
	programSession := loader.NewSession(fs)
	fs = programSession.FS()

	eslintPlugins := configCatalog.EslintPlugins
	ruleCatalog, shadowedPluginRules := rules.All().ForESLintPlugins(eslintPlugins)
	reportShadowedPluginRules(shadowedPluginRules)
	var rslintConfig rslintconfig.RslintConfig

	// configMap holds per-directory configs for automatically discovered
	// configs. An explicit config uses rslintConfig instead.
	var configMap map[string]rslintconfig.RslintConfig

	var configTargetScopes map[string]target.OwnerScope

	// Program-wide type checking builds every configured project. Plain linting
	// waits for target discovery and builds only the projects owned by configs
	// that govern at least one selected target.
	var projectSet loader.ProjectSet
	buildAllPrograms := typeCheck || typeCheckOnly

	configDirectories := configCatalog.ConfigDirectories()
	if configCatalog.Explicit {
		if len(configDirectories) != 1 {
			fmt.Fprintf(os.Stderr, "error: explicit config catalog contains %d configs, want exactly one\n", len(configDirectories))
			return 1
		}
		currentDirectory = configDirectories[0]
		rslintConfig = slices.Clone(configCatalog.Configs[currentDirectory])
	} else {
		configMap = make(map[string]rslintconfig.RslintConfig, len(configCatalog.Configs))
		configTargetScopes = make(map[string]target.OwnerScope, len(configCatalog.Scopes))
		for _, configDir := range configDirectories {
			configMap[configDir] = slices.Clone(configCatalog.Configs[configDir])
			if scope, ok := configCatalog.Scopes[configDir]; ok {
				scope.ExplicitFiles = slices.Clone(scope.ExplicitFiles)
				configTargetScopes[configDir] = scope
			}
		}
	}

	targetConfigMap := cloneConfigMap(configMap)
	targetRslintConfig := slices.Clone(rslintConfig)

	// Apply --rule CLI overrides by appending a synthetic ConfigEntry. Target
	// discovery below intentionally uses the pre-override config snapshots
	// above, so --rule overlays already-selected lint targets without widening
	// discovery by itself.
	if len(ruleFlags) > 0 {
		cliEntry, err := rslintconfig.BuildCLIRuleEntry(ruleFlags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if cliEntry != nil {
			if configMap != nil {
				for dir, cfg := range configMap {
					configMap[dir] = append(cfg, *cliEntry)
				}
			} else {
				rslintConfig = append(rslintConfig, *cliEntry)
			}
		}
	}

	// Validate every configured rule's options against its declared schema —
	// a separate step after configuration is fully resolved (including --rule
	// overrides) and before any linting starts, so a bad config fails fast
	// with every failure reported at once instead of surfacing mid-lint.
	var optionErrors []rslintconfig.ResolvedRuleOptionsError
	configMap, rslintConfig, optionErrors = rslintconfig.ValidateResolvedRuleOptions(configMap, rslintConfig, ruleCatalog)
	if len(optionErrors) > 0 {
		for _, optionError := range optionErrors {
			fmt.Fprintf(os.Stderr, "error: %s\n", optionError.Error())
		}
		return 1
	}

	outputOptions := output.Options{
		Format:       format,
		Quiet:        quiet,
		ColorEnabled: colorEnabled,
	}
	startWriter := args.StartWriter
	if startWriter == nil {
		startWriter = os.Stdout
	}
	if err := output.RenderStart(startWriter, mode, outputOptions); err != nil {
		if ctx.Err() == nil {
			fmt.Fprintf(os.Stderr, "error writing lint report: %v\n", err)
		}
		return 1
	}
	abortRun := func(reason, legacyMessage string) int {
		// The IPC owner maps a canceled command to exit 130. Do not make an
		// interrupted run look like an ordinary completed failure.
		if ctx.Err() != nil {
			return 1
		}
		if format == output.FormatDefault {
			if err := output.RenderAbort(os.Stdout, mode, timeBefore, reason, outputOptions); err != nil {
				fmt.Fprintf(os.Stderr, "error writing lint report: %v\n", err)
			}
		} else {
			fmt.Fprintln(os.Stderr, legacyMessage)
		}
		return 1
	}

	// Program-wide type checking builds every configured project. Delay that
	// expensive work until preflight has succeeded and the interactive start
	// line is visible. The pre-override snapshots preserve the prior project
	// selection semantics: --rule changes rules, not project discovery.
	if buildAllPrograms {
		if targetConfigMap != nil {
			projectSet, err = programSession.BuildProjects(targetConfigMap, singleThreaded)
		} else {
			projectSet, err = programSession.BuildProject(currentDirectory, targetRslintConfig, singleThreaded)
		}
		if err != nil {
			return abortRun(err.Error(), fmt.Sprintf("error: %v", err))
		}
	}

	// Use CWD for display paths (not any config directory).
	// In multi-config mode, currentDirectory was never reassigned from os.Getwd(),
	// so it already holds the normalized CWD.
	cwd := workingDirectory

	comparePathOptions := tspath.ComparePathsOptions{
		CurrentDirectory:          cwd,
		UseCaseSensitiveFileNames: true,
	}
	broadProjectLoad := isBroadProjectLoadScope(
		allowFiles,
		allowDirs,
		cwd,
		fs.UseCaseSensitiveFileNames(),
	)

	// No args → implicit CWD scoping (same as `rslint .`), matching ESLint.
	// This keeps an explicit --config outside the current directory from
	// widening the scanned root to the config file's directory.
	if len(allowFiles) == 0 && len(allowDirs) == 0 {
		allowDirs = []string{cwd}
	}

	// --- Lint target discovery and Program loading ---
	programs := projectSet.Programs()
	buildSingleConfigPrograms := buildAllPrograms
	var (
		targetPlan     target.Plan
		loadedPrograms loader.LoadResult
	)
	// --type-check-only is program-wide and pays no lint-target discovery,
	// target binding/parsing, config-resolution, or Program-loading cost.
	if !typeCheckOnly {
		targetPlan, err = target.Resolve(target.Request{
			ConfigMap:       targetConfigMap,
			Config:          targetRslintConfig,
			ConfigDirectory: currentDirectory,
			ScanRoot:        workingDirectory,
			OwnerScopes:     configTargetScopes,
			FS:              fs,
			Files:           allowFiles,
			Directories:     allowDirs,
			SingleThreaded:  singleThreaded,
		})
		if err != nil {
			return abortRun(err.Error(), fmt.Sprintf("error: %v", err))
		}
		if !buildAllPrograms {
			if configMap != nil {
				if broadProjectLoad {
					projectSet, err = programSession.BuildProjectsForTargetOwners(configMap, targetPlan, singleThreaded)
				} else {
					projectSet, err = programSession.BuildTargetProjects(configMap, targetPlan, singleThreaded)
				}
			} else if len(targetPlan.Files) > 0 {
				buildSingleConfigPrograms = true
				if broadProjectLoad {
					projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
				} else {
					projectSet, err = programSession.BuildTargetProject(currentDirectory, rslintConfig, targetPlan, singleThreaded)
				}
			}
			if err != nil {
				return abortRun(err.Error(), fmt.Sprintf("error: %v", err))
			}
		}
		loadedPrograms, err = programSession.LoadCLI(projectSet, targetPlan, currentDirectory, singleThreaded)
		if err != nil {
			return abortRun(err.Error(), fmt.Sprintf("error: %v", err))
		}
		programs = loadedPrograms.Programs
	}

	// Rebuild ts-go Programs against a caller-supplied immutable VFS generation
	// and bind the original stable target plan again. A target can move between
	// rslint Programs when in-memory fixes change the import graph.
	createPrograms := func(session *loader.Session) (loader.LoadResult, error) {
		var rebuilt loader.ProjectSet
		var err error
		if configMap != nil {
			if buildAllPrograms {
				rebuilt, err = session.BuildProjects(configMap, singleThreaded)
			} else if broadProjectLoad {
				rebuilt, err = session.BuildProjectsForTargetOwners(configMap, targetPlan, singleThreaded)
			} else {
				rebuilt, err = session.BuildTargetProjects(configMap, targetPlan, singleThreaded)
			}
		} else if buildSingleConfigPrograms {
			if buildAllPrograms || broadProjectLoad {
				rebuilt, err = session.BuildProject(currentDirectory, rslintConfig, singleThreaded)
			} else {
				rebuilt, err = session.BuildTargetProject(currentDirectory, rslintConfig, targetPlan, singleThreaded)
			}
		}
		if err != nil {
			return loader.LoadResult{}, err
		}
		return session.LoadCLI(rebuilt, targetPlan, currentDirectory, singleThreaded)
	}

	// The command owns config/target/Program construction and physical writes.
	// It supplies the current immutable source generation to the linter-owned
	// pipeline, which alone schedules native/plugin work and fix rounds.
	hasEslintPlugins := !typeCheckOnly && len(eslintPlugins) > 0
	generationForBinding := func(
		binding loader.LoadResult,
		generationFS vfs.FS,
	) linter.Generation {
		var fileConfigResolver *configLint.Resolver
		var rulesForFile linter.RuleHandler
		if !typeCheckOnly {
			fileConfigResolver = configLint.NewResolver(configLint.ResolverOptions{
				ConfigsByOwner:                      configMap,
				Config:                              rslintConfig,
				ConfigDirectory:                     currentDirectory,
				Catalog:                             ruleCatalog,
				TargetsBySourcePath:                 binding.LintTargetBySourcePath,
				SourceMappingsIncludeCanonicalPaths: true,
				PathSpaces:                          targetPlan.PathSpaces(),
				FS:                                  generationFS,
			})
			rulesForFile = func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
				return fileConfigResolver.EnabledRulesForSourcePath(sourceFile.FileName())
			}
		}
		targetPath := func(sourcePath string) string {
			if lintTarget, ok := target.LookupSourceTarget(binding.LintTargetBySourcePath, sourcePath, generationFS); ok {
				return lintTarget.Path
			}
			return sourcePath
		}
		var plugin *linter.PluginGeneration
		if hasEslintPlugins {
			plugin = &linter.PluginGeneration{
				ConfigForFile: eslintPluginConfigResolver{
					lintResolver: fileConfigResolver,
				}.resolve,
				HostReadsInitialText: pluginHostReadsInitialText,
			}
		}
		return linter.Generation{
			Native: linter.NativeGeneration{
				Programs:         binding.Programs,
				TargetsByProgram: binding.TargetsByProgram,
				RulesForFile:     rulesForFile,
				Cwd:              cwd,
				TypeCheck:        typeCheck,
				SingleThreaded:   singleThreaded,
				Timing:           timingCollector,
			},
			Target: linter.TargetProjection{
				Path: targetPath,
				ReadText: func(path string, source ast.SourceFileLike) (string, error) {
					return utils.RestoreSourceBOM(generationFS, path, source.Text()), nil
				},
			},
			Plugin: plugin,
		}
	}

	initialBinding := loadedPrograms
	if typeCheckOnly {
		initialBinding = loader.LoadResult{Programs: programs}
	}
	provider := &cliGenerationProvider{
		initial:   initialBinding,
		initialFS: programSession.FS(),
		rebuild: func(ctx context.Context, snapshot linter.SourceSnapshot) (loader.LoadResult, vfs.FS, error) {
			if err := ctx.Err(); err != nil {
				return loader.LoadResult{}, nil, err
			}
			files := snapshot.Files()
			overrides := make(map[string]string, len(files)*2)
			for _, file := range files {
				path := tspath.NormalizePath(file.Path)
				overrides[path] = file.Text
				if realPath := sourceFS.Realpath(path); realPath != "" {
					overrides[tspath.NormalizePath(realPath)] = file.Text
				}
			}
			roundSession := loader.NewSession(utils.NewOverlayVFS(sourceFS, overrides))
			binding, err := createPrograms(roundSession)
			if err != nil {
				return loader.LoadResult{}, nil, err
			}
			if err := ctx.Err(); err != nil {
				return loader.LoadResult{}, nil, err
			}
			return binding, roundSession.FS(), nil
		},
		generation: generationForBinding,
	}
	demand := linter.ArtifactDemand{}
	if fix {
		demand.Native = rule.EditDemandAutofix
		demand.Plugin = rule.EditDemandAll
	}
	observationPolicy := linter.ObservationPolicy{
		Demand:        demand,
		Plugin:        linter.PluginConcurrentJoined,
		PluginFailure: linter.PluginKeepPartialWithSynthetic,
	}
	var pipelineRequest linter.PipelineRequest
	if fix {
		pipelineRequest = linter.NewAutofixRequestWithCommitter(
			provider,
			cliFinalChangeCommitter{},
			observationPolicy,
			linter.AutofixPolicy{
				VerifyAfterLastRound: true,
				VerificationDemand: linter.ArtifactDemand{
					Native: rule.EditDemandNone,
					Plugin: rule.EditDemandAll,
				},
			},
			dispatch,
		)
	} else {
		pipelineRequest = linter.NewLintRequest(provider, observationPolicy, dispatch)
	}
	pipelineResult, err := linter.RunPipeline(ctx, pipelineRequest)
	for _, pluginRecord := range pipelineResult.PluginOutcomes() {
		reportEslintPluginDispatchOutcome(linter.EslintPluginDispatchOutcome{
			Notices:       pluginRecord.Notices,
			DispatchError: pluginRecord.DispatchError,
		})
	}
	if err != nil {
		operation := "running linter"
		reason := err.Error()
		if _, autofixStarted := pipelineResult.AppliedFixes(); autofixStarted {
			operation = "applying fixes"
			reason = operation + ": " + reason
		}
		return abortRun(reason, fmt.Sprintf("error %s: %v", operation, err))
	}
	allDiags, complete := pipelineResult.Observation.CompleteDiagnostics()
	if !complete {
		const reason = "CLI lint returned an incomplete observation"
		return abortRun(reason, "error running linter: "+reason)
	}
	lintResult := pipelineResult.Observation.Native.Lint
	initialObservation := pipelineResult.Observation
	fixedCount := 0
	if applied, ok := pipelineResult.AppliedFixes(); ok {
		initialObservation = applied.Initial
		for _, round := range applied.Rounds {
			fixedCount += round.AppliedDiagnostics
		}
	}
	lintedfileCount := initialObservation.Native.Lint.LintedFileCount
	lintResult.ExecutedRules = pipelineResult.ExecutedRules()

	// Emit per-file warnings for CLI-specified files that won't be linted.
	// Distinguishes "not found on disk" vs "ignored by pattern", aligned
	// with ESLint v10's warning behavior. Skipped in --type-check-only mode:
	// these are lint-phase concepts and would mislead users about Phase 2
	// (which runs program-wide regardless of CLI scope and rslint ignores).
	if !typeCheckOnly {
		warnings := collectAllowFileWarnings(targetPlan.ExplicitFileOutcomes)
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, formatAllowFileWarning(w, comparePathOptions))
		}
	}
	scopeRestricted := len(allowFiles) > 0 || len(allowDirs) > 0
	if format != output.FormatDefault && shouldShortCircuitMachineOutput(typeCheckOnly, typeCheck, scopeRestricted, lintedfileCount) {
		return 0
	}

	allDiags = deduplicateTypeScriptDiagnostics(allDiags, fs, targetPlan.PreferredCallerPaths())

	// Paths have already been remapped into the caller-visible target space.
	// Sort the completed set before rendering; same-start diagnostics retain
	// emission order.
	linter.StableSortDiagnosticsByFileAndStart(allDiags)

	// Phase 3: Build one report from the final post-fix diagnostics, then let
	// the CLI output subsystem own format dispatch, colors, and status text.
	var typeCheckedRoots []string
	if typeCheck {
		// Compiler-capable Program roots represent the user-authored type-check
		// set; transitive declarations are intentionally excluded.
		for _, prog := range programs {
			if !prog.CanProvideProgramDiagnostics() {
				continue
			}
			typeCheckedRoots = append(typeCheckedRoots, prog.RootFileNames()...)
		}
	}

	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}

	report := output.NewReport(allDiags, output.Metadata{
		Mode:        mode,
		Files:       lintReportFileCount(mode, targetPlan.Files, typeCheckedRoots, fs),
		Rules:       len(lintResult.ExecutedRules),
		Threads:     threadsCount,
		FixedIssues: fixedCount,
		StartedAt:   timeBefore,
	})
	counts := report.Counts()
	outcome := lintReportOutcome(counts, maxWarnings)
	outputOptions.ComparePaths = comparePathOptions
	if err := renderLintReport(ctx, os.Stdout, report, outcome, outputOptions); err != nil {
		return abortRun(
			"writing lint report: "+err.Error(),
			fmt.Sprintf("error writing lint report: %v", err),
		)
	}

	// The default status already explains the warning-budget failure. Preserve
	// the legacy stderr line for machine formats without contaminating stdout.
	if format != output.FormatDefault && outcome.Kind == output.OutcomeWarningLimitExceeded {
		fmt.Fprintf(os.Stderr, "Rslint found too many warnings (maximum: %d).\n", maxWarnings)
	}

	// The timing table goes to stderr so machine-readable stdout formats
	// (jsonline/github/gitlab) stay parseable with --timing enabled. It is
	// deliberately emitted after every result message so it remains last.
	if timingCollector != nil {
		table := output.FormatRuleTimingTable(timingCollector.Timings(), timingLimit)
		if args.DeferTimingTable != nil {
			args.DeferTimingTable(table)
		} else {
			fmt.Fprint(os.Stderr, table)
		}
	}

	if outcome.Kind != output.OutcomePassed {
		return 1
	}
	return 0
}
