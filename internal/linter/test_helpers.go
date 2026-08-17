package linter

import (
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// RunLinterInProgram is a compatibility adapter for cross-package rule tests.
// It deliberately accepts only rslint's unified Program; raw ts-go assembly
// belongs in the test that creates the source generation, just as it does at
// CLI/API/LSP boundaries. New tests should prefer RunLinter or LintSingleFile.
//
// When typeCheck is true the adapter routes through RunLinter (single
// program) so callers retain the program-level tsc-aligned semantics;
// otherwise it bypasses Phase 2 entirely. lintedFileCount preserves the
// historical return value of files actually visited by lint rules.
func RunLinterInProgram(
	sourceProgram *lintprogram.Program,
	allowFiles []string,
	allowDirs []string,
	skipFiles []string,
	getRulesForFile RuleHandler,
	typeCheck bool,
	onDiagnostic DiagnosticHandler,
	fileFilter func(string) bool,
) int32 {
	excludes := skipFiles
	if excludes == nil {
		excludes = []string{}
	}
	var ff FileFilter
	if fileFilter != nil {
		ff = fileFilter
	}
	if onDiagnostic == nil {
		onDiagnostic = func(rule.RuleDiagnostic) {}
	}
	programs := []*lintprogram.Program{sourceProgram}
	var filters []FileFilter
	if ff != nil {
		filters = []FileFilter{ff}
	}
	runOpts := RunLinterOptions{
		Programs:         programs,
		SingleThreaded:   true,
		Scope:            FileScope{Files: allowFiles, Dirs: allowDirs},
		ExcludePaths:     excludes,
		GetRulesForFile:  getRulesForFile,
		PerProgramFilter: filters,
		TypeCheck:        typeCheck,
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: onDiagnostic,
		},
	}
	// Preserve this legacy test adapter's parser-recovery behavior without
	// exposing a syntax-override map in the production API.
	planOpts := programPlanOptions{
		Program:         sourceProgram,
		Scope:           runOpts.Scope,
		ExcludePaths:    excludes,
		FileFilter:      ff,
		GetRulesForFile: getRulesForFile,
		SkipSyntaxCheck: true,
	}
	plan, err := prepareProgramLintPlan(planOpts)
	if err != nil {
		panic(err)
	}
	runOpts.PreparedPlan = &LintPlan{programs: []programLintPlan{plan}}
	res, err := RunLinter(runOpts)
	if err != nil || res == nil {
		return 0
	}
	return res.LintedFileCount
}
