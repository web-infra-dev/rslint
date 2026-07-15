package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"

	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type independentFixEnvironment struct {
	configMap                map[string]rslintconfig.RslintConfig
	config                   rslintconfig.RslintConfig
	currentDirectory         string
	enforcePlugins           bool
	fs                       vfs.FS
	dispatch                 linter.EslintPluginDispatcher
	hasEslintPlugins         bool
	typeCheck                bool
	singleThreaded           bool
	buildSingleConfigProgram bool
}

type independentFixResult struct {
	path        string
	canonical   string
	diagnostics []rule.RuleDiagnostic
	content     string
	didFix      bool
	fixed       int
	executed    map[string]struct{}
	err         error
}

// runIndependentFixes mirrors ESLint's verifyAndFix/outputFixes split. Every
// lexical lint target runs all fix passes against its own in-memory source;
// only after every result settles are outputs written. Two lexical aliases of
// one inode therefore never union their configs or fixes in memory. Like
// ESLint.outputFixes, default-mode writes are independent and concurrent;
// --singleThreaded serializes them as it does the rest of this pipeline.
func runIndependentFixes(
	ctx context.Context,
	diagnostics []rule.RuleDiagnostic,
	targetPlan lintTargetPlan,
	environment independentFixEnvironment,
	maxPasses int,
) ([]rule.RuleDiagnostic, int, map[string]struct{}, error) {
	grouped := groupDiagsByLexicalFile(diagnostics)
	targetByPath := make(map[string]resolvedLintTarget, len(targetPlan.Targets))
	for _, target := range targetPlan.Targets {
		targetByPath[exactHostPathID(target.Path)] = target
	}

	paths := make([]string, 0, len(grouped))
	for path, fileDiagnostics := range grouped {
		if _, exists := targetByPath[path]; exists && diagnosticsHaveFixes(fileDiagnostics) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return diagnostics, 0, nil, nil
	}

	results := make([]independentFixResult, len(paths))
	run := func(index int) {
		target := targetByPath[paths[index]]
		results[index] = environment.fixTarget(ctx, target, grouped[paths[index]], maxPasses)
	}
	if environment.singleThreaded || len(paths) == 1 {
		for index := range paths {
			run(index)
			if results[index].err != nil {
				break
			}
		}
	} else {
		workerCount := min(len(paths), max(1, runtime.GOMAXPROCS(0)))
		work := make(chan int)
		var waitGroup sync.WaitGroup
		waitGroup.Add(workerCount)
		for range workerCount {
			go func() {
				defer waitGroup.Done()
				for index := range work {
					run(index)
				}
			}()
		}
		for index := range paths {
			work <- index
		}
		close(work)
		waitGroup.Wait()
	}

	for _, result := range results {
		if result.err != nil {
			return nil, 0, nil, result.err
		}
	}
	if err := writeIndependentFixOutputs(results, environment.singleThreaded); err != nil {
		return nil, 0, nil, err
	}

	processedPaths := make(map[string]struct{}, len(results))
	processedCanonical := make(map[string]struct{}, len(results))
	fixedCount := 0
	executed := make(map[string]struct{})
	for _, result := range results {
		processedPaths[exactHostPathID(result.path)] = struct{}{}
		processedCanonical[exactHostPathID(result.canonical)] = struct{}{}
		fixedCount += result.fixed
		for name := range result.executed {
			executed[name] = struct{}{}
		}
	}

	finalDiagnostics := make([]rule.RuleDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		if _, replaced := processedPaths[exactHostPathID(diagnostic.FilePath)]; replaced {
			continue
		}
		if diagnostic.Origin == rule.DiagnosticOriginTypeScript {
			if _, replaced := processedCanonical[canonicalHostPathID(diagnostic.FilePath, environment.fs)]; replaced {
				continue
			}
		}
		finalDiagnostics = append(finalDiagnostics, diagnostic)
	}
	for _, result := range results {
		finalDiagnostics = append(finalDiagnostics, result.diagnostics...)
	}
	return finalDiagnostics, fixedCount, executed, nil
}

func diagnosticsHaveFixes(diagnostics []rule.RuleDiagnostic) bool {
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes()) > 0 {
			return true
		}
	}
	return false
}

func (environment independentFixEnvironment) fixTarget(
	ctx context.Context,
	target resolvedLintTarget,
	initialDiagnostics []rule.RuleDiagnostic,
	maxPasses int,
) independentFixResult {
	result := independentFixResult{
		path:        target.Path,
		canonical:   target.CanonicalPath,
		diagnostics: initialDiagnostics,
		executed:    make(map[string]struct{}),
	}
	foundSource := false
	for _, diagnostic := range initialDiagnostics {
		if len(diagnostic.Fixes()) > 0 && diagnostic.SourceFile != nil {
			result.content = diagnostic.SourceFile.Text()
			foundSource = true
			break
		}
	}
	if !foundSource && diagnosticsHaveFixes(initialDiagnostics) {
		result.err = fmt.Errorf("fix target %q has no source text", target.Path)
		return result
	}

	for range maxPasses {
		fixable := make([]rule.RuleDiagnostic, 0, len(result.diagnostics))
		for _, diagnostic := range result.diagnostics {
			if len(diagnostic.Fixes()) > 0 {
				fixable = append(fixable, diagnostic)
			}
		}
		if len(fixable) == 0 {
			break
		}
		fixedContent, unapplied, wasFixed := linter.ApplyRuleFixes(result.content, fixable)
		if !wasFixed {
			break
		}
		result.content = fixedContent
		result.didFix = true
		result.fixed += len(fixable) - len(unapplied)

		diagnostics, executed, err := environment.lintTargetContent(ctx, target, fixedContent)
		if err != nil {
			result.err = err
			return result
		}
		result.diagnostics = diagnostics
		for name := range executed {
			result.executed[name] = struct{}{}
		}
	}
	return result
}

func (environment independentFixEnvironment) lintTargetContent(
	ctx context.Context,
	target resolvedLintTarget,
	content string,
) ([]rule.RuleDiagnostic, map[string]struct{}, error) {
	overlayFiles := map[string]string{target.Path: content}
	if target.CanonicalPath != "" {
		overlayFiles[target.CanonicalPath] = content
	}
	overlayFS := utils.NewOverlayVFS(environment.fs, overlayFiles)
	parseCache := utils.NewParseCache()
	singlePlan := lintTargetPlan{Targets: []resolvedLintTarget{target}}

	var programSet lintProgramSet
	var err error
	if environment.configMap != nil {
		programSet, err = createProgramSetForConfigs(
			configsForLintTargetPlan(environment.configMap, singlePlan),
			environment.singleThreaded,
			overlayFS,
			parseCache,
		)
	} else if environment.buildSingleConfigProgram {
		programSet, err = createProgramSetForConfig(
			environment.currentDirectory,
			environment.config,
			environment.singleThreaded,
			overlayFS,
			parseCache,
		)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("rebuild Program for fixed target %q: %w", target.Path, err)
	}
	binding, err := bindLintTargetPlan(
		programSet,
		singlePlan,
		environment.currentDirectory,
		overlayFS,
		parseCache,
		environment.singleThreaded,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("bind fixed target %q: %w", target.Path, err)
	}

	resolverOptions := lintConfigResolverOptions{
		ConfigMap:        environment.configMap,
		Config:           environment.config,
		CurrentDirectory: environment.currentDirectory,
		EnforcePlugins:   environment.enforcePlugins,
		FS:               overlayFS,
	}
	var diagnostics []rule.RuleDiagnostic
	var diagnosticsMu sync.Mutex
	appendDiagnostic := func(diagnostic rule.RuleDiagnostic) {
		diagnosticsMu.Lock()
		diagnostics = append(diagnostics, diagnostic)
		diagnosticsMu.Unlock()
	}
	views, resolvers, err := buildBindingLintProgramViews(binding, resolverOptions, appendDiagnostic)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve fixed target config %q: %w", target.Path, err)
	}
	syntaxDiagnostics, syntaxErrorFiles := collectTargetViewSyntacticDiagnostics(binding, environment.typeCheck, false)
	diagnostics = append(diagnostics, syntaxDiagnostics...)
	runOptions := linter.RunLinterOptions{
		Programs:              binding.Programs,
		ProgramViews:          views,
		SingleThreaded:        environment.singleThreaded,
		SyntaxErrorFiles:      syntaxErrorFiles,
		TypeCheck:             environment.typeCheck,
		SkipTypeCheckPrograms: buildTypeCheckSkipMask(binding.Programs),
		OnDiagnostic:          appendDiagnostic,
		ExcludePaths:          []string{"bundled:"},
	}

	var pluginChannel <-chan []rule.RuleDiagnostic
	if environment.hasEslintPlugins {
		inputs := buildPluginFileInputs(runOptions, pluginConfigResolver{lintResolvers: resolvers})
		for index := range inputs {
			text := inputs[index].SourceFile.Text()
			inputs[index].Text = &text
		}
		pluginChannel = dispatchPluginLintAsync(ctx, environment.dispatch, inputs, true, pluginSuggestionsMode(true))
	}
	lintResult, lintErr := linter.RunLinter(runOptions)
	if pluginChannel != nil {
		diagnostics = append(diagnostics, (<-pluginChannel)...)
	}
	if lintErr != nil {
		return nil, nil, fmt.Errorf("lint fixed target %q: %w", target.Path, lintErr)
	}

	canonicalTarget := exactHostPathID(target.CanonicalPath)
	filtered := diagnostics[:0]
	for _, diagnostic := range diagnostics {
		if diagnostic.Origin == rule.DiagnosticOriginTypeScript {
			if canonicalHostPathID(diagnostic.FilePath, overlayFS) != canonicalTarget {
				continue
			}
			diagnostic.FilePath = target.Path
		}
		filtered = append(filtered, diagnostic)
	}
	return filtered, lintResult.ExecutedRules, nil
}

func writeIndependentFixOutputs(results []independentFixResult, singleThreaded bool) error {
	errorsByIndex := make([]error, len(results))
	write := func(index int) {
		if err := os.WriteFile(results[index].path, []byte(results[index].content), 0o644); err != nil {
			errorsByIndex[index] = fmt.Errorf("write fixed file %q: %w", results[index].path, err)
		}
	}
	if singleThreaded {
		for index := range results {
			if results[index].didFix {
				write(index)
			}
		}
	} else {
		var waitGroup sync.WaitGroup
		for index := range results {
			if !results[index].didFix {
				continue
			}
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				write(index)
			}()
		}
		waitGroup.Wait()
	}
	return errors.Join(errorsByIndex...)
}
