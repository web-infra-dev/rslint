package no_duplicates_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_duplicates"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoDuplicatesEditDemand(t *testing.T) {
	t.Parallel()

	const code = "import { first } from './shared';\nimport { second } from './shared';\n"
	program, sourceFile := createNoDuplicatesProgram(t, "edit-demand.ts", code)

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:         program,
			File:            sourceFile.FileName(),
			HasTypeInfo:     true,
			GetRulesForFile: noDuplicatesConfiguredRules,
			ExcludePaths:    []string{},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixes := run(rule.EditDemandAutofix)
	suggestionsOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	for name, diagnostics := range map[string][]rule.RuleDiagnostic{
		"diagnostics only": diagnosticsOnly,
		"autofixes":        autofixes,
		"suggestions only": suggestionsOnly,
		"all edits":        allEdits,
	} {
		if len(diagnostics) != 2 {
			t.Fatalf("%s: got %d diagnostics, want 2", name, len(diagnostics))
		}
		for i, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Errorf("%s: diagnostic %d unexpectedly has suggestions", name, i)
			}
		}
	}

	for i := range allEdits {
		want := diagnosticWithoutEdits(allEdits[i])
		for name, diagnostics := range map[string][]rule.RuleDiagnostic{
			"diagnostics only": diagnosticsOnly,
			"autofixes":        autofixes,
			"suggestions only": suggestionsOnly,
		} {
			if got := diagnosticWithoutEdits(diagnostics[i]); !reflect.DeepEqual(got, want) {
				t.Errorf("%s: diagnostic %d changed:\ngot  %#v\nwant %#v", name, i, got, want)
			}
		}
	}

	if got := len(autofixes[0].Fixes()); got == 0 {
		t.Fatal("autofixes: first diagnostic has no merge fix")
	}
	if !reflect.DeepEqual(autofixes[0].Fixes(), allEdits[0].Fixes()) {
		t.Fatalf("autofix and all-edits modes produced different fixes:\nautofix %#v\nall     %#v", autofixes[0].Fixes(), allEdits[0].Fixes())
	}
	for name, diagnostics := range map[string][]rule.RuleDiagnostic{
		"diagnostics only": diagnosticsOnly,
		"suggestions only": suggestionsOnly,
	} {
		for i, diagnostic := range diagnostics {
			if got := len(diagnostic.Fixes()); got != 0 {
				t.Errorf("%s: diagnostic %d has %d fixes, want 0", name, i, got)
			}
		}
	}
	for i := 1; i < len(autofixes); i++ {
		if got := len(autofixes[i].Fixes()); got != 0 {
			t.Errorf("autofixes: diagnostic %d has %d fixes, want 0", i, got)
		}
	}

	const unfixableCode = "import first from './shared';\nimport second from './shared';\n"
	unfixableProgram, unfixableSourceFile := createNoDuplicatesProgram(t, "unfixable.ts", unfixableCode)
	var unfixableDiagnostics []rule.RuleDiagnostic
	linter.LintSingleFile(linter.LintSingleFileOptions{
		Program:         unfixableProgram,
		File:            unfixableSourceFile.FileName(),
		HasTypeInfo:     true,
		GetRulesForFile: noDuplicatesConfiguredRules,
		ExcludePaths:    []string{},
		Consumer: rule.DiagnosticConsumer{
			Demand: rule.EditDemandAll,
			Report: func(diagnostic rule.RuleDiagnostic) {
				unfixableDiagnostics = append(unfixableDiagnostics, diagnostic)
			},
		},
	})
	if len(unfixableDiagnostics) != 2 {
		t.Fatalf("unfixable imports: got %d diagnostics, want 2", len(unfixableDiagnostics))
	}
	for i, diagnostic := range unfixableDiagnostics {
		if diagnostic.FixesPtr != nil {
			t.Errorf("unfixable imports: diagnostic %d has a non-nil fix payload", i)
		}
	}
}

type diagnosticIdentity struct {
	Range    [2]int
	RuleName string
	Message  rule.RuleMessage
	FilePath string
	Severity rule.DiagnosticSeverity
}

func diagnosticWithoutEdits(diagnostic rule.RuleDiagnostic) diagnosticIdentity {
	return diagnosticIdentity{
		Range:    [2]int{diagnostic.Range.Pos(), diagnostic.Range.End()},
		RuleName: diagnostic.RuleName,
		Message:  diagnostic.Message,
		FilePath: diagnostic.FilePath,
		Severity: diagnostic.Severity,
	}
}

func createNoDuplicatesProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
	t.Helper()

	rootDir := fixtures.GetRootDir()
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), code)
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatalf("failed to create program: %v", err)
	}
	sourceFile := program.GetSourceFile(fileName)
	if sourceFile == nil {
		t.Fatalf("source file %q not found", fileName)
	}
	return program, sourceFile
}

func noDuplicatesConfiguredRules(*ast.SourceFile) []linter.ConfiguredRule {
	return []linter.ConfiguredRule{{
		Name:     no_duplicates.NoDuplicatesRule.Name,
		Severity: rule.SeverityError,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return no_duplicates.NoDuplicatesRule.Run(ctx, nil)
		},
	}}
}
