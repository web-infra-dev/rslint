package first_test

import (
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestFirstEditDemand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		code           string
		wantFixPayload []bool
	}{
		{
			name: "all misplaced imports batch together",
			code: "const before = 1;\n" +
				"import { first } from './first';\n" +
				"const between = 2;\n" +
				"import { second } from './second';\n" +
				"import { third } from './third';\n",
			wantFixPayload: []bool{true, true, true},
		},
		{
			name: "reference cuts off the current fix batch",
			code: "const before = 1;\n" +
				"import { first } from './first';\n" +
				"const usedBeforeImport = second;\n" +
				"import { second } from './second';\n" +
				"import { third } from './third';\n",
			wantFixPayload: []bool{true, false, false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createFirstProgram(t, "edit-demand.ts", test.code)
			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				t.Helper()

				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:         program,
					File:            sourceFile.FileName(),
					HasTypeInfo:     true,
					GetRulesForFile: firstConfiguredRules,
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
				if len(diagnostics) != len(test.wantFixPayload) {
					t.Fatalf("%s: got %d diagnostics, want %d", name, len(diagnostics), len(test.wantFixPayload))
				}
				for i, diagnostic := range diagnostics {
					if diagnostic.Suggestions != nil {
						t.Errorf("%s: diagnostic %d unexpectedly has suggestions", name, i)
					}
				}
			}

			for i := range allEdits {
				wantIdentity := firstDiagnosticWithoutEdits(allEdits[i])
				for name, diagnostics := range map[string][]rule.RuleDiagnostic{
					"diagnostics only": diagnosticsOnly,
					"autofixes":        autofixes,
					"suggestions only": suggestionsOnly,
				} {
					if got := firstDiagnosticWithoutEdits(diagnostics[i]); !reflect.DeepEqual(got, wantIdentity) {
						t.Errorf("%s: diagnostic %d changed:\ngot  %#v\nwant %#v", name, i, got, wantIdentity)
					}
				}

				if got := autofixes[i].FixesPtr != nil; got != test.wantFixPayload[i] {
					t.Errorf("autofixes: diagnostic %d fix-payload presence = %v, want %v", i, got, test.wantFixPayload[i])
				}
				if !reflect.DeepEqual(autofixes[i].Fixes(), allEdits[i].Fixes()) {
					t.Errorf("diagnostic %d differs between autofix and all-edits modes", i)
				}
				if diagnosticsOnly[i].FixesPtr != nil {
					t.Errorf("diagnostics only: diagnostic %d unexpectedly has a fix payload", i)
				}
				if suggestionsOnly[i].FixesPtr != nil {
					t.Errorf("suggestions only: diagnostic %d unexpectedly has a fix payload", i)
				}
			}
		})
	}
}

type firstDiagnosticIdentity struct {
	Range    [2]int
	RuleName string
	Message  rule.RuleMessage
	FilePath string
	Severity rule.DiagnosticSeverity
}

func firstDiagnosticWithoutEdits(diagnostic rule.RuleDiagnostic) firstDiagnosticIdentity {
	return firstDiagnosticIdentity{
		Range:    [2]int{diagnostic.Range.Pos(), diagnostic.Range.End()},
		RuleName: diagnostic.RuleName,
		Message:  diagnostic.Message,
		FilePath: diagnostic.FilePath,
		Severity: diagnostic.Severity,
	}
}
