package linter

import (
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// runLinterPositional is a backwards-compatible positional wrapper around
// RunLinter for in-package tests written before the Options refactor. New
// code should construct RunLinterOptions directly; the old positional call
// sites in *_test.go are mechanically renamed from `RunLinter(` to
// `runLinterPositional(`.
func runLinterPositional(
	programs []*compiler.Program,
	singleThreaded bool,
	allowFiles []string,
	allowDirs []string,
	excludedPaths []string,
	getRulesForFile RuleHandler,
	typeCheck bool,
	onDiagnostic DiagnosticHandler,
	typeInfoFiles map[string]struct{},
	fileFilters []func(string) bool,
) (*LintResult, error) {
	var ff []FileFilter
	if fileFilters != nil {
		ff = make([]FileFilter, len(fileFilters))
		for i, f := range fileFilters {
			ff[i] = f
		}
	}
	return RunLinter(RunLinterOptions{
		Programs:         programs,
		SingleThreaded:   singleThreaded,
		Scope:            FileScope{Files: allowFiles, Dirs: allowDirs},
		ExcludePaths:     excludedPaths,
		PerProgramFilter: ff,
		GetRulesForFile:  getRulesForFile,
		TypeInfoFiles:    typeInfoFiles,
		TypeCheck:        typeCheck,
		OnDiagnostic:     onDiagnostic,
	})
}

// RunLinterInProgram is a backwards-compatible test adapter for the
// now-internal runLintRulesInProgram. New code should use RunLinter or
// LintSingleFile.
//
// When typeCheck is true the adapter routes through RunLinter (single
// program) so callers retain the program-level tsc-aligned semantics;
// otherwise it bypasses Phase 2 entirely. lintedFileCount preserves the
// historical return value of files actually visited by lint rules.
func RunLinterInProgram(
	program *compiler.Program,
	allowFiles []string,
	allowDirs []string,
	skipFiles []string,
	getRulesForFile RuleHandler,
	typeCheck bool,
	onDiagnostic DiagnosticHandler,
	typeInfoFiles map[string]struct{},
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
	if typeCheck {
		// Route through RunLinter so the new program-level type-check phase
		// runs. The returned LintResult.LintedFileCount equals what
		// runLintRulesInProgram would have returned for this single program.
		res, _ := RunLinter(RunLinterOptions{
			Programs:        []*compiler.Program{program},
			SingleThreaded:  true,
			Scope:           FileScope{Files: allowFiles, Dirs: allowDirs},
			ExcludePaths:    excludes,
			GetRulesForFile: getRulesForFile,
			TypeInfoFiles:   typeInfoFiles,
			TypeCheck:       true,
			OnDiagnostic:    onDiagnostic,
			PerProgramFilter: func() []FileFilter {
				if ff == nil {
					return nil
				}
				return []FileFilter{ff}
			}(),
		})
		if res == nil {
			return 0
		}
		return res.LintedFileCount
	}
	return runLintRulesInProgram(runProgramOptions{
		Program:         program,
		Scope:           FileScope{Files: allowFiles, Dirs: allowDirs},
		ExcludePaths:    excludes,
		FileFilter:      ff,
		GetRulesForFile: getRulesForFile,
		TypeInfoFiles:   typeInfoFiles,
		OnDiagnostic:    onDiagnostic,
	})
}
