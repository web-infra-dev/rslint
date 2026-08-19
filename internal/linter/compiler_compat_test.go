package linter

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

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
	var filters []FileFilter
	if fileFilters != nil {
		filters = make([]FileFilter, len(lintPrograms))
		for index, owner := range owners {
			if owner < len(fileFilters) {
				filters[index] = fileFilters[owner]
			}
		}
	}
	return RunLinter(RunLinterOptions{
		Programs:         lintPrograms,
		SingleThreaded:   singleThreaded,
		Scope:            FileScope{Files: allowFiles, Dirs: allowDirs},
		ExcludePaths:     excludedPaths,
		PerProgramFilter: filters,
		TargetFiles:      targetFiles,
		GetRulesForFile:  getRulesForFile,
		TypeCheck:        typeCheck,
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
	var filters []FileFilter
	if fileFilter != nil {
		filters = make([]FileFilter, len(programs))
		for index := range filters {
			filters[index] = fileFilter
		}
	}
	runOpts := RunLinterOptions{
		Programs:         programs,
		SingleThreaded:   true,
		Scope:            FileScope{Files: allowFiles, Dirs: allowDirs},
		ExcludePaths:     excludes,
		PerProgramFilter: filters,
		TargetFiles:      targets,
		GetRulesForFile:  getRulesForFile,
		TypeCheck:        typeCheck,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: onDiagnostic,
		},
	}
	prepared := &LintPlan{programs: make([]programLintPlan, len(programs))}
	for index, sourceProgram := range programs {
		planOpts := programPlanOptions{
			Program:         sourceProgram,
			Scope:           runOpts.Scope,
			ExcludePaths:    excludes,
			GetRulesForFile: getRulesForFile,
			SkipSyntaxCheck: true,
		}
		if filters != nil {
			planOpts.FileFilter = filters[index]
		}
		if targets != nil {
			planOpts.HasTargetFiles = true
			planOpts.TargetFiles = targets[index]
		}
		plan, err := prepareProgramLintPlan(planOpts)
		if err != nil {
			panic(err)
		}
		prepared.programs[index] = plan
	}
	runOpts.PreparedPlan = prepared
	result, err := RunLinter(runOpts)
	if err != nil || result == nil {
		return 0
	}
	return result.LintedFileCount
}
