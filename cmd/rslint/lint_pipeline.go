package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/output"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/term"
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

// loadGitignoreAndProjects overlaps two independent CLI preparation steps for
// the single-config JSON path. Program construction itself belongs to the
// loader; this small scheduling decision remains command orchestration.
func loadGitignoreAndProjects(
	config rslintconfig.RslintConfig,
	configDirectory string,
	gitignoreRoot string,
	targetFiles []string,
	targetDirectories []string,
	singleThreaded bool,
	session *loader.Session,
) (rslintconfig.RslintConfig, loader.ProjectSet, error) {
	var (
		configWithIgnores rslintconfig.RslintConfig
		projects          loader.ProjectSet
		programErr        error
	)
	work := core.NewWorkGroup(singleThreaded)
	work.Queue(func() {
		configWithIgnores = rslintconfig.ConfigWithGitignoreForTargetsFromRoot(
			config,
			configDirectory,
			gitignoreRoot,
			session.FS(),
			targetFiles,
			targetDirectories,
		)
	})
	work.Queue(func() {
		projects, programErr = session.BuildProject(configDirectory, config, singleThreaded)
	})
	work.RunAndWait()
	if programErr != nil {
		return config, loader.ProjectSet{}, programErr
	}
	return configWithIgnores, projects, nil
}

// executeLintPipeline runs the full lint flow (config load → program build →
// lint target planning/Program loading → lint → optional --fix loop → report) and
// returns the process exit code. Shared by the IPC entry (runCLI) and the wasm
// native fallback.
func executeLintPipeline(args lintArgs, ctx context.Context, dispatch linter.EslintPluginDispatcher) int {
	// Unpack into locals so the pipeline body below stays verbatim — only the
	// flag-parse front matter lives in parseLintFlags.
	init := args.Init
	config := args.Config
	configCatalog := args.ConfigCatalog
	usesJSConfig := configCatalog != nil && (configCatalog.Explicit || len(configCatalog.Configs) > 0)
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

	fs := args.FS
	if fs == nil {
		fs = bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	}

	// The run-scoped loader owns source snapshots, compiler metadata, Program
	// construction, target binding, and fix-generation rebuilds. Integrations
	// consume only its unified Program results.
	programSession := loader.NewSession(fs)
	fs = programSession.FS()

	var eslintPlugins []rslintconfig.EslintPluginEntry
	if usesJSConfig {
		eslintPlugins = configCatalog.EslintPlugins
	}
	ruleCatalog, shadowedPluginRules := deriveRuleCatalog(eslintPlugins)
	reportShadowedPluginRules(shadowedPluginRules)
	var rslintConfig rslintconfig.RslintConfig

	// configMap holds per-directory configs for automatically discovered JS/TS
	// configs. Explicit JS/TS and JSON/JSONC configs use rslintConfig instead.
	var configMap map[string]rslintconfig.RslintConfig

	var configTargetScopes map[string]target.OwnerScope

	// Program-wide type checking builds every configured project. Plain linting
	// waits for target discovery and builds only the projects owned by configs
	// that govern at least one selected target.
	var projectSet loader.ProjectSet
	buildAllPrograms := typeCheck || typeCheckOnly

	if usesJSConfig {
		configDirectories := configCatalog.ConfigDirectories()
		if configCatalog.Explicit {
			if len(configDirectories) != 1 {
				fmt.Fprintf(os.Stderr, "error: explicit config catalog contains %d configs, want exactly one\n", len(configDirectories))
				return 1
			}
			currentDirectory = configDirectories[0]
			rslintConfig = slices.Clone(configCatalog.Configs[currentDirectory])
			if buildAllPrograms {
				projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
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

			if buildAllPrograms {
				projectSet, err = programSession.BuildProjects(configMap, singleThreaded)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					return 1
				}
			}
		}
	} else {
		// Load configuration from file (JSON config path, isJSConfig stays false)
		loader := rslintconfig.NewConfigLoader(fs, currentDirectory, rules.All())
		rslintConfig, currentDirectory, err = loader.LoadRslintConfiguration(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if typeCheckOnly {
			projectSet, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
		} else if buildAllPrograms {
			rslintConfig, projectSet, err = loadGitignoreAndProjects(
				rslintConfig, currentDirectory, workingDirectory, allowFiles, allowDirs, singleThreaded, programSession,
			)
		} else {
			rslintConfig = rslintconfig.ConfigWithGitignoreForTargetsFromRoot(
				rslintConfig,
				currentDirectory,
				workingDirectory,
				fs,
				allowFiles,
				allowDirs,
			)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
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
	var optionsMessages []string
	configMap, rslintConfig, optionsMessages = validateResolvedRuleOptions(configMap, rslintConfig, ruleCatalog)
	if len(optionsMessages) > 0 {
		for _, message := range optionsMessages {
			fmt.Fprintf(os.Stderr, "error: %s\n", message)
		}
		return 1
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
	programConfigMap := configMap
	buildSingleConfigPrograms := buildAllPrograms
	var (
		targetPlan             target.Plan
		loadedPrograms         loader.LoadResult
		targetsByProgram       [][]string
		lintTargetBySourcePath map[string]target.File
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
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		if !buildAllPrograms {
			if configMap != nil {
				programConfigMap = configsForOwners(configMap, targetPlan.ActiveOwners())
				if broadProjectLoad {
					projectSet, err = programSession.BuildProjects(programConfigMap, singleThreaded)
				} else {
					projectSet, err = programSession.BuildTargetProjects(programConfigMap, targetPlan, singleThreaded)
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
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return 1
			}
		}
		loadedPrograms, err = programSession.LoadCLI(projectSet, targetPlan, currentDirectory, singleThreaded)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return 1
		}
		programs = loadedPrograms.Programs
		targetsByProgram = loadedPrograms.TargetsByProgram
		lintTargetBySourcePath = loadedPrograms.LintTargetBySourcePath
	}

	// Rebuild ts-go Programs and bind the original stable target plan again on
	// every fix pass. A target can move between rslint Programs when fixes
	// change the import graph.
	createPrograms := func() (loader.LoadResult, error) {
		var rebuilt loader.ProjectSet
		var err error
		if configMap != nil {
			if buildAllPrograms || broadProjectLoad {
				rebuilt, err = programSession.BuildProjects(programConfigMap, singleThreaded)
			} else {
				rebuilt, err = programSession.BuildTargetProjects(programConfigMap, targetPlan, singleThreaded)
			}
		} else if buildSingleConfigPrograms {
			if buildAllPrograms || broadProjectLoad {
				rebuilt, err = programSession.BuildProject(currentDirectory, rslintConfig, singleThreaded)
			} else {
				rebuilt, err = programSession.BuildTargetProject(currentDirectory, rslintConfig, targetPlan, singleThreaded)
			}
		}
		if err != nil {
			return loader.LoadResult{}, err
		}
		return programSession.LoadCLI(rebuilt, targetPlan, currentDirectory, singleThreaded)
	}

	// Phase 1: Collect all diagnostics (no printing yet).
	// Like ESLint, diagnostics are collected first, then printed at the end.
	// This ensures --fix only shows remaining unfixed issues.
	var allDiags []rule.RuleDiagnostic
	var diagsMu sync.Mutex
	fixedCount := 0

	diagnosticsChan := make(chan rule.RuleDiagnostic, 4096)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range diagnosticsChan {
			allDiags = append(allDiags, d)
		}
	}()

	enforcePlugins := usesJSConfig
	var fileConfigResolver *lintConfigResolver

	// Target discovery already excluded default paths, global ignores, and
	// .gitignore entries. Target ownership and deduplication were already
	// resolved in targetsByProgram.
	// Source-only Programs expose no checker or
	// program-wide diagnostic capability; the linter consumes that distinction
	// through the common Program contract rather than a parallel skip mask.
	syntaxDiagnostics := collectTargetSyntacticDiagnostics(
		programs,
		targetsByProgram,
		typeCheck,
		typeCheckOnly,
	)
	for _, diagnostic := range syntaxDiagnostics {
		diagnosticsChan <- diagnostic
	}

	// In --type-check-only mode, skip the lint phase entirely by passing
	// nil for GetRulesForFile. RunLinter's Phase 1 is gated on this being
	// non-nil; Phase 2 (type-check) runs independently.
	var rulesForFile linter.RuleHandler
	if !typeCheckOnly {
		fileConfigResolver = newLintConfigResolver(lintConfigResolverOptions{
			ConfigMap:               configMap,
			Config:                  rslintConfig,
			CurrentDirectory:        currentDirectory,
			RuleCatalog:             ruleCatalog,
			EnforcePlugins:          enforcePlugins,
			LintTargetBySourcePath:  lintTargetBySourcePath,
			SourceMappingsCanonical: true,
			PathSpaces:              targetPlan.PathSpaces(),
			FS:                      fs,
		})
		rulesForFile = func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
			return fileConfigResolver.EnabledRulesForFile(sourceFile.FileName())
		}
	}

	nativeEditDemand := rule.EditDemandNone
	if fix {
		nativeEditDemand = rule.EditDemandAutofix
	}
	runOpts := linter.RunLinterOptions{
		Programs:        programs,
		SingleThreaded:  singleThreaded,
		Cwd:             cwd,
		Scope:           linter.FileScope{Files: allowFiles, Dirs: allowDirs},
		TargetFiles:     targetsByProgram,
		GetRulesForFile: rulesForFile,
		TypeCheck:       typeCheck,
		Timing:          timingCollector,
		Consumer: rule.DiagnosticConsumer{
			Demand: nativeEditDemand,
			Report: func(d rule.RuleDiagnostic) {
				diagnosticsChan <- d
			},
		},
	}
	preparedPlan, planErr := linter.PrepareLintPlan(runOpts)
	if planErr != nil {
		close(diagnosticsChan)
		wg.Wait()
		fmt.Fprintf(os.Stderr, "error preparing lint plan: %v\n", planErr)
		return 1
	}
	runOpts.PreparedPlan = preparedPlan
	// Dispatch eslint-plugin rules to the Node worker in parallel with the
	// native lint pass; results are awaited + merged before output / --fix.
	// ONLY when plugins are actually configured — otherwise the whole reverse-
	// dispatch is skipped so the native-only path pays nothing for the feature.
	// Both paths consume the same prepared file/rule plan.
	hasEslintPlugins := !typeCheckOnly && len(eslintPlugins) > 0
	pluginResolver := pluginConfigResolver{
		lintResolver: fileConfigResolver,
	}
	var pluginCh <-chan []rule.RuleDiagnostic
	if hasEslintPlugins {
		pluginInputs := buildPluginFileInputs(runOpts.PreparedPlan, pluginResolver)
		pluginCh = dispatchPluginLintAsync(ctx, dispatch, pluginInputs, fix, pluginSuggestionsMode(fix), timingCollector)
	}

	lintResult, err := linter.RunLinter(runOpts)

	close(diagnosticsChan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running linter: %v\n", err)
		return 1
	}

	lintedfileCount := lintResult.LintedFileCount

	wg.Wait()
	// Merge eslint-plugin diagnostics (dispatched in parallel) now that the
	// native diagnostics goroutine has drained.
	if pluginCh != nil {
		allDiags = append(allDiags, (<-pluginCh)...)
	}
	remapDiagnosticTargetPaths(allDiags, lintTargetBySourcePath, fs)

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
	if shouldShortCircuitOutput(typeCheckOnly, typeCheck, scopeRestricted, lintedfileCount) {
		return 0
	}

	// Phase 2: Apply fixes if --fix flag is enabled.
	// Uses multi-pass fixing: after applying fixes, rebuild programs and re-lint
	// to catch cascading issues (e.g. no-wrapper-object-types fix triggers no-inferrable-types).
	// After fixing, allDiags is replaced with remaining (unfixed) diagnostics.
	const maxFixPasses = 10
	if fix && len(allDiags) > 0 {
		diagnosticsByFile := groupDiagsByFile(allDiags)
		passFixed, fixErr := applyFixPass(diagnosticsByFile, fs)
		// Replace the entire source generation after every write attempt and
		// before any Program rebuild. os.WriteFile may truncate or partially
		// mutate a file even when it ultimately returns an error, and whole-
		// generation invalidation also covers caller/source/symlink aliases.
		programSession.InvalidateSourceSnapshots()
		if fixErr != nil {
			fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
			return 1
		}
		fixedCount += passFixed

		// Re-lint → fix → re-lint → fix → ... until stable or maxFixPasses.
		// Skip if nothing was fixed in the first pass (no need to re-lint).
		for pass := 1; pass <= maxFixPasses && fixedCount > 0; pass++ {
			newBinding, err := createPrograms()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error rebuilding Programs after fixes: %v\n", err)
				return 1
			}
			if len(newBinding.Programs) == 0 {
				fmt.Fprintln(os.Stderr, "error rebuilding Programs after fixes: no Program returned")
				return 1
			}

			// Re-lint using the fresh binding derived from the stable target plan.
			fixTargetsByProgram := newBinding.TargetsByProgram
			fixLintTargetBySourcePath := newBinding.LintTargetBySourcePath
			fixConfigResolver := newLintConfigResolver(lintConfigResolverOptions{
				ConfigMap:               configMap,
				Config:                  rslintConfig,
				CurrentDirectory:        currentDirectory,
				RuleCatalog:             ruleCatalog,
				EnforcePlugins:          enforcePlugins,
				LintTargetBySourcePath:  newBinding.LintTargetBySourcePath,
				SourceMappingsCanonical: true,
				PathSpaces:              targetPlan.PathSpaces(),
				FS:                      fs,
			})
			fixGetRulesForFile := func(sourceFile *ast.SourceFile) []rule.ConfiguredRule {
				return fixConfigResolver.EnabledRulesForFile(sourceFile.FileName())
			}
			var fixRulesForFile linter.RuleHandler
			if !typeCheckOnly {
				fixRulesForFile = fixGetRulesForFile
			}
			passEditDemand := rule.EditDemandAutofix
			if pass == maxFixPasses {
				// This pass only verifies the bytes written by the final
				// allowed write pass; no further fixes can be applied.
				passEditDemand = rule.EditDemandNone
			}
			var passDiags []rule.RuleDiagnostic
			fixSyntaxDiagnostics := collectTargetSyntacticDiagnostics(
				newBinding.Programs,
				fixTargetsByProgram,
				typeCheck,
				typeCheckOnly,
			)
			passDiags = append(passDiags, fixSyntaxDiagnostics...)
			fixRunOpts := linter.RunLinterOptions{
				Programs:        newBinding.Programs,
				SingleThreaded:  singleThreaded,
				Cwd:             cwd,
				Scope:           linter.FileScope{Files: allowFiles, Dirs: allowDirs},
				TargetFiles:     fixTargetsByProgram,
				GetRulesForFile: fixRulesForFile,
				TypeCheck:       typeCheck,
				Timing:          timingCollector,
				Consumer: rule.DiagnosticConsumer{
					Demand: passEditDemand,
					Report: func(d rule.RuleDiagnostic) {
						diagsMu.Lock()
						passDiags = append(passDiags, d)
						diagsMu.Unlock()
					},
				},
			}
			fixPreparedPlan, planErr := linter.PrepareLintPlan(fixRunOpts)
			if planErr != nil {
				fmt.Fprintf(os.Stderr, "error preparing lint plan after fixes: %v\n", planErr)
				return 1
			}
			fixRunOpts.PreparedPlan = fixPreparedPlan
			// Re-dispatch plugin rules each pass (only when configured): the
			// worker re-reads the post-fix file content, and merging here keeps
			// plugin diagnostics from being lost when allDiags is replaced.
			// Each pass prepares a fresh plan for the rebuilt target binding.
			var fixPluginCh <-chan []rule.RuleDiagnostic
			if hasEslintPlugins {
				fixPluginInputs := buildPluginFileInputs(fixRunOpts.PreparedPlan, pluginConfigResolver{
					lintResolver: fixConfigResolver,
				})
				fixPluginCh = dispatchPluginLintAsync(ctx, dispatch, fixPluginInputs, fix, pluginSuggestionsMode(fix), timingCollector)
			}
			passResult, passErr := linter.RunLinter(fixRunOpts)
			var fixPluginDiags []rule.RuleDiagnostic
			if fixPluginCh != nil {
				fixPluginDiags = <-fixPluginCh
			}
			if passErr != nil {
				fmt.Fprintf(os.Stderr, "error running linter after fixes: %v\n", passErr)
				return 1
			}
			if passResult != nil {
				for name := range passResult.ExecutedRules {
					lintResult.ExecutedRules[name] = struct{}{}
				}
			}
			// Merge this pass's plugin diagnostics before applying fixes so
			// plugin fixes participate and plugin diagnostics survive.
			passDiags = append(passDiags, fixPluginDiags...)
			remapDiagnosticTargetPaths(passDiags, fixLintTargetBySourcePath, fs)

			// Replace allDiags with latest post-fix diagnostics.
			allDiags = passDiags
			if pass == maxFixPasses {
				// The maximum number of write passes has already run (the initial
				// pass plus maxFixPasses-1 loop passes). This extra pass is the
				// required final verification of the bytes written by pass 10.
				break
			}

			passFixed, fixErr := applyFixPass(groupDiagsByFile(passDiags), fs)
			// See the first fix pass above: invalidate before inspecting the
			// result so a partially successful write can never feed a rebuild.
			programSession.InvalidateSourceSnapshots()
			if fixErr != nil {
				fmt.Fprintf(os.Stderr, "error applying fixes: %v\n", fixErr)
				return 1
			}
			if passFixed == 0 {
				break // Stable — allDiags reflect final state
			}
			fixedCount += passFixed
		}
	}

	allDiags = deduplicateTypeScriptDiagnostics(allDiags, fs, targetPlan.PreferredCallerPaths())

	// Diagnostics arrive in completion order — programs and, within a
	// program, file shards run in parallel — so impose a deterministic
	// order before printing. The key is (file, start position) only,
	// deliberately with NO end/rule tie-break: ESLint orders same-start
	// diagnostics by emission order (parent reported before nested child),
	// and a file's diagnostics are all emitted by a single worker, so under
	// a STABLE sort this key is already fully deterministic. Keep this
	// comparator in sync with the --api one in api_lint.go (same policy over
	// api.Diagnostic).
	slices.SortStableFunc(allDiags, func(a, b rule.RuleDiagnostic) int {
		if c := strings.Compare(a.FilePath, b.FilePath); c != 0 {
			return c
		}
		return cmp.Compare(a.Range.Pos(), b.Range.Pos())
	})

	// Phase 3: Build one report from the final post-fix diagnostics, then let
	// the CLI output subsystem own format dispatch, colors, and summary text.
	mode := output.ModeLint
	if typeCheckOnly {
		mode = output.ModeTypeCheckOnly
	} else if typeCheck {
		mode = output.ModeLintAndTypeCheck
	}

	typeCheckedFileCount := 0
	if typeCheck {
		// Count compiler-capable Program root files (tsconfig include/files),
		// not transitive declarations, for every summary that includes type-check.
		seen := make(map[string]struct{})
		for _, prog := range programs {
			if !prog.CanProvideProgramDiagnostics() {
				continue
			}
			for _, fileName := range prog.RootFileNames() {
				seen[fileName] = struct{}{}
			}
		}
		typeCheckedFileCount = len(seen)
	}

	threadsCount := 1
	if !singleThreaded {
		threadsCount = runtime.GOMAXPROCS(0)
	}

	report := output.NewReport(allDiags, output.Metadata{
		Mode:             mode,
		LintedFiles:      int(lintedfileCount),
		TypeCheckedFiles: typeCheckedFileCount,
		Rules:            len(lintResult.ExecutedRules),
		Threads:          threadsCount,
		FixedIssues:      fixedCount,
		StartedAt:        timeBefore,
	})
	if err := output.Render(os.Stdout, report, output.Options{
		Format:       format,
		ComparePaths: comparePathOptions,
		Quiet:        quiet,
		ColorEnabled: colorEnabled,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error writing lint report: %v\n", err)
		return 1
	}
	// The timing table goes to stderr so machine-readable stdout formats
	// (jsonline/github/gitlab) stay parseable with --timing enabled.
	if timingCollector != nil {
		table := output.FormatRuleTimingTable(timingCollector.Timings(), timingLimit)
		if args.DeferTimingTable != nil {
			args.DeferTimingTable(table)
		} else {
			fmt.Fprint(os.Stderr, table)
		}
	}
	counts := report.Counts()

	tooManyWarnings := maxWarnings >= 0 && counts.Warnings > maxWarnings

	if counts.Errors == 0 && tooManyWarnings {
		fmt.Fprintf(os.Stderr, "Rslint found too many warnings (maximum: %d).\n", maxWarnings)
	}

	// Exit with non-zero status code if errors were found
	if counts.Errors > 0 || tooManyWarnings {
		return 1
	}
	return 0
}
