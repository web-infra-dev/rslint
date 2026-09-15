package testutil

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// DefaultExcludedPathSubstrings preserves the historical exclusions used by
// lenient cross-package lint fixtures.
var DefaultExcludedPathSubstrings = []string{"/node_modules/", "bundled:"}

// LintProgramOptions describes the exact files exercised by LintProgram.
// Files contains Program source-file names; nil selects the complete source
// universe, while an explicit empty slice is invalid.
type LintProgramOptions struct {
	Program                *program.Program
	Files                  []string
	ExcludedPathSubstrings []string
	GetRulesForFile        linter.RuleHandler
	OnDiagnostic           linter.DiagnosticHandler
}

// LintProgram preserves the parser-recovery behavior needed by a small number
// of cross-package tests without exposing test-only target selection from the
// product linter package. It fails when the requested fixture selects no files.
func LintProgram(t testing.TB, opts LintProgramOptions) {
	t.Helper()
	if opts.Program == nil || !opts.Program.IsValid() {
		t.Fatal("LintProgram requires a valid Program")
	}
	if opts.GetRulesForFile == nil {
		t.Fatal("LintProgram requires GetRulesForFile")
	}
	if opts.OnDiagnostic == nil {
		t.Fatal("LintProgram requires OnDiagnostic")
	}

	files := selectProgramFiles(t, opts.Program, opts.Files)
	consumer := rule.DiagnosticConsumer{
		Demand: rule.EditDemandAll,
		Report: opts.OnDiagnostic,
	}
	linted := 0
	for _, file := range files {
		if pathContainsAny(string(file.Path()), opts.ExcludedPathSubstrings) {
			continue
		}
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:         opts.Program,
			File:            file.FileName(),
			HasTypeInfo:     opts.Program.CanProvideTypeChecker(file),
			GetRulesForFile: opts.GetRulesForFile,
			Consumer:        consumer,
		})
		linted++
	}
	if linted == 0 {
		t.Fatal("LintProgram selected no files after exclusions")
	}
}

func selectProgramFiles(t testing.TB, sourceProgram *program.Program, fileNames []string) []*ast.SourceFile {
	t.Helper()
	if fileNames == nil {
		files := sourceProgram.SourceFiles()
		if len(files) == 0 {
			t.Fatal("LintProgram received an empty Program")
		}
		return files
	}
	if len(fileNames) == 0 {
		t.Fatal("LintProgram requires at least one file when Files is non-nil")
	}

	files := make([]*ast.SourceFile, 0, len(fileNames))
	seen := make(map[string]struct{}, len(fileNames))
	for _, fileName := range fileNames {
		file := sourceProgram.GetSourceFile(fileName)
		if file == nil {
			t.Fatalf("LintProgram file %q is not in Program", fileName)
		}
		if _, ok := seen[file.FileName()]; ok {
			continue
		}
		seen[file.FileName()] = struct{}{}
		files = append(files, file)
	}
	return files
}

func pathContainsAny(path string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(path, substring) {
			return true
		}
	}
	return false
}
