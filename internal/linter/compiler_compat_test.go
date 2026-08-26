package linter

import (
	"context"
	"os"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var legacyDefaultExcludedPathSubstrings = []string{"/node_modules/", "bundled:"}

// runLinterPositional preserves the positional convention of older in-package
// tests. Product code and cross-package tests enter through unified Programs.
func runLinterPositional(
	programs []*compiler.Program,
	singleThreaded bool,
	allowFiles []string,
	allowDirs []string,
	excludedPaths []string,
	getRulesForFile RuleHandler,
	typeCheck bool,
	onDiagnostic DiagnosticHandler,
	checkerFiles map[string]struct{},
	fileFilters []func(string) bool,
) (*LintResult, error) {
	lintPrograms, targetFiles, owners := testProgramsByCheckerCapability(programs, checkerFiles)
	var filters []legacyFileFilter
	if fileFilters != nil {
		filters = make([]legacyFileFilter, len(lintPrograms))
		for index, owner := range owners {
			if owner < len(fileFilters) {
				filters[index] = fileFilters[owner]
			}
		}
	}
	prepared := &LintPlan{programs: make([]programLintPlan, len(lintPrograms))}
	for index, sourceProgram := range lintPrograms {
		var plan programLintPlan
		var err error
		if targetFiles != nil {
			targets := filterTestTargets(targetFiles[index], filterAt(filters, index))
			plan, err = prepareExactTestProgramLintPlan(sourceProgram, targets, excludedPaths, false, getRulesForFile)
		} else {
			plan, err = prepareLegacyProgramLintPlan(legacyProgramPlanOptions{
				program:      sourceProgram,
				scope:        legacyFileScope{files: allowFiles, dirs: allowDirs},
				excludePaths: excludedPaths,
				fileFilter: composeOwnedFilter(
					filterAt(filters, index),
					buildOwnedFileSet(sourceProgram),
				),
				getRulesForFile: getRulesForFile,
			})
		}
		if err != nil {
			return nil, err
		}
		prepared.programs[index] = plan
	}
	return RunLinter(RunLinterOptions{
		SingleThreaded: singleThreaded,
		LintPlan:       prepared,
		TypeCheck:      typeCheck,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: onDiagnostic,
		},
	})
}

// testProgramsByCheckerCapability translates the retired checker-file map
// used by older in-package tests into explicit source generations.
func testProgramsByCheckerCapability(
	rawPrograms []*compiler.Program,
	checkerFiles map[string]struct{},
) ([]*lintprogram.Program, [][]string, []int) {
	if checkerFiles == nil {
		owners := make([]int, len(rawPrograms))
		for index := range owners {
			owners[index] = index
		}
		return lintprogram.NewFromCompilers(rawPrograms), nil, owners
	}

	programs := make([]*lintprogram.Program, 0, len(rawPrograms)*2)
	targets := make([][]string, 0, len(rawPrograms)*2)
	owners := make([]int, 0, len(rawPrograms)*2)
	for rawIndex, raw := range rawPrograms {
		programs = append(programs, lintprogram.NewFromCompiler(raw))
		owners = append(owners, rawIndex)
		var typedTargets []string
		var sourceOnlyFiles []*ast.SourceFile
		var sourceOnlyTargets []string
		for _, fileName := range raw.CommandLine().FileNames() {
			file := raw.GetSourceFile(fileName)
			if file == nil {
				continue
			}
			if _, ok := checkerFiles[file.FileName()]; ok {
				typedTargets = append(typedTargets, file.FileName())
				continue
			}
			sourceOnlyFiles = append(sourceOnlyFiles, file)
			sourceOnlyTargets = append(sourceOnlyTargets, file.FileName())
		}
		targets = append(targets, typedTargets)
		if len(sourceOnlyFiles) == 0 {
			continue
		}
		sourceOnly, err := lintprogram.NewFromBoundSources(raw, sourceOnlyFiles)
		if err != nil {
			panic(err)
		}
		programs = append(programs, sourceOnly)
		targets = append(targets, sourceOnlyTargets)
		owners = append(owners, rawIndex)
	}
	return programs, targets, owners
}

func runLinterInCompilerProgram(
	raw *compiler.Program,
	allowFiles []string,
	allowDirs []string,
	skipFiles []string,
	getRulesForFile RuleHandler,
	typeCheck bool,
	onDiagnostic DiagnosticHandler,
	checkerFiles map[string]struct{},
	fileFilter func(string) bool,
) int32 {
	programs, targets, _ := testProgramsByCheckerCapability([]*compiler.Program{raw}, checkerFiles)
	excludes := skipFiles
	if excludes == nil {
		excludes = []string{}
	}
	if onDiagnostic == nil {
		onDiagnostic = func(rule.RuleDiagnostic) {}
	}
	var filters []legacyFileFilter
	if fileFilter != nil {
		filters = make([]legacyFileFilter, len(programs))
		for index := range filters {
			filters[index] = fileFilter
		}
	}
	runOpts := RunLinterOptions{
		SingleThreaded: true,
		TypeCheck:      typeCheck,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: onDiagnostic,
		},
	}
	prepared := &LintPlan{programs: make([]programLintPlan, len(programs))}
	for index, sourceProgram := range programs {
		var plan programLintPlan
		var err error
		if targets != nil {
			plan, err = prepareExactTestProgramLintPlan(
				sourceProgram,
				filterTestTargets(targets[index], filterAt(filters, index)),
				excludes,
				true,
				getRulesForFile,
			)
		} else {
			plan, err = prepareLegacyProgramLintPlan(legacyProgramPlanOptions{
				program:         sourceProgram,
				scope:           legacyFileScope{files: allowFiles, dirs: allowDirs},
				excludePaths:    excludes,
				fileFilter:      filterAt(filters, index),
				skipSyntaxCheck: true,
				getRulesForFile: getRulesForFile,
			})
		}
		if err != nil {
			panic(err)
		}
		prepared.programs[index] = plan
	}
	runOpts.LintPlan = prepared
	result, err := RunLinter(runOpts)
	if err != nil || result == nil {
		return 0
	}
	return result.LintedFileCount
}

func filterAt(filters []legacyFileFilter, index int) legacyFileFilter {
	if index >= len(filters) {
		return nil
	}
	return filters[index]
}

func filterTestTargets(targets []string, filter legacyFileFilter) []string {
	if filter == nil {
		return targets
	}
	filtered := make([]string, 0, len(targets))
	for _, target := range targets {
		if filter(target) {
			filtered = append(filtered, target)
		}
	}
	return filtered
}

type legacyFileScope struct {
	files []string
	dirs  []string
}

type legacyFileFilter func(absPath string) bool

type legacyProgramPlanOptions struct {
	program         *lintprogram.Program
	scope           legacyFileScope
	excludePaths    []string
	fileFilter      legacyFileFilter
	skipSyntaxCheck bool
	getRulesForFile RuleHandler
}

func prepareExactTestProgramLintPlan(
	sourceProgram *lintprogram.Program,
	targets []string,
	excludePaths []string,
	skipSyntaxCheck bool,
	getRulesForFile RuleHandler,
) (programLintPlan, error) {
	files := make([]*ast.SourceFile, 0, len(targets))
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		file := sourceProgram.GetSourceFile(target)
		if file == nil || pathContainsAny(string(file.Path()), excludePaths) {
			continue
		}
		if _, ok := seen[file.FileName()]; ok {
			continue
		}
		seen[file.FileName()] = struct{}{}
		files = append(files, file)
	}
	return prepareProgramLintPlanForFiles(programRulePlanOptions{
		Program:         sourceProgram,
		SkipSyntaxCheck: skipSyntaxCheck,
		GetRulesForFile: getRulesForFile,
	}, files)
}

func prepareLegacyProgramLintPlan(opts legacyProgramPlanOptions) (programLintPlan, error) {
	plan, err := newProgramLintPlanForFiles(opts.program, collectLegacyFilesToLint(opts))
	if err != nil {
		return programLintPlan{}, err
	}
	ruleOpts := programRulePlanOptions{
		Program:         opts.program,
		SkipSyntaxCheck: opts.skipSyntaxCheck,
		GetRulesForFile: opts.getRulesForFile,
	}
	for fileIndex := range plan.files {
		resolveProgramLintPlanFile(ruleOpts, &plan, fileIndex, context.Background())
	}
	return plan, nil
}

func collectLegacyFilesToLint(opts legacyProgramPlanOptions) []*ast.SourceFile {
	var allowFileInfos []os.FileInfo
	if opts.scope.files != nil {
		allowFileInfos = precomputeAllowFileInfos(opts.scope.files)
	}
	var filesToLint []*ast.SourceFile
	for _, file := range opts.program.SourceFiles() {
		if filePassesLegacyProjection(opts, file, allowFileInfos) {
			filesToLint = append(filesToLint, file)
		}
	}
	return filesToLint
}

func filePassesLegacyProjection(opts legacyProgramPlanOptions, file *ast.SourceFile, allowFileInfos []os.FileInfo) bool {
	if pathContainsAny(string(file.Path()), opts.excludePaths) {
		return false
	}
	if opts.scope.files != nil || opts.scope.dirs != nil {
		fileAllowed := opts.scope.files != nil && isFileAllowed(file.FileName(), opts.scope.files, allowFileInfos)
		dirAllowed := opts.scope.dirs != nil && isDirAllowed(file.FileName(), opts.scope.dirs)
		if !fileAllowed && !dirAllowed {
			return false
		}
	}
	return opts.fileFilter == nil || opts.fileFilter(file.FileName())
}

func pathContainsAny(path string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(path, substring) {
			return true
		}
	}
	return false
}

func isFileAllowed(fileName string, allowFiles []string, allowFileInfos []os.FileInfo) bool {
	for _, filePath := range allowFiles {
		if filePath == fileName {
			return true
		}
	}
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	for _, info := range allowFileInfos {
		if os.SameFile(fileInfo, info) {
			return true
		}
	}
	return false
}

func precomputeAllowFileInfos(allowFiles []string) []os.FileInfo {
	infos := make([]os.FileInfo, 0, len(allowFiles))
	for _, file := range allowFiles {
		if info, err := os.Stat(file); err == nil {
			infos = append(infos, info)
		}
	}
	return infos
}

func isDirAllowed(fileName string, allowDirs []string) bool {
	for _, dirPath := range allowDirs {
		if tspath.StartsWithDirectory(fileName, dirPath, true) {
			return true
		}
	}
	return false
}

func composeOwnedFilter(extra legacyFileFilter, owned map[string]struct{}) legacyFileFilter {
	if extra == nil && owned == nil {
		return nil
	}
	return func(name string) bool {
		if extra != nil && !extra(name) {
			return false
		}
		if owned != nil {
			if _, ok := owned[name]; !ok {
				return false
			}
		}
		return true
	}
}

func buildOwnedFileSet(sourceProgram *lintprogram.Program) map[string]struct{} {
	fileNames := sourceProgram.RootFileNames()
	if len(fileNames) == 0 {
		return nil
	}
	owned := make(map[string]struct{}, len(fileNames))
	for _, fileName := range fileNames {
		owned[fileName] = struct{}{}
	}
	return owned
}
