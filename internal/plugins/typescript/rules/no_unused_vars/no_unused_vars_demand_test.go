package no_unused_vars

import (
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNoUnusedVarsImportEditDemand(t *testing.T) {
	t.Parallel()

	const code = `import DefaultValue, { usedValue, unusedOne, unusedTwo } from './named';
import * as unusedNamespace from './namespace';
import unusedEquals = require('./equals');
console.log(usedValue);
`

	tests := []struct {
		name            string
		options         []any
		requestedDemand rule.EditDemand
		otherDemand     rule.EditDemand
		wantFixes       bool
	}{
		{
			name:            "suggestions",
			options:         rule.NormalizeOptions(nil),
			requestedDemand: rule.EditDemandSuggestion,
			otherDemand:     rule.EditDemandAutofix,
		},
		{
			name: "autofix",
			options: rule.NormalizeOptions(map[string]interface{}{
				"enableAutofixRemoval": map[string]interface{}{"imports": true},
			}),
			requestedDemand: rule.EditDemandAutofix,
			otherDemand:     rule.EditDemandSuggestion,
			wantFixes:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			program, sourceFile := createNoUnusedVarsProgram(t, "edit-demand-"+test.name+".ts", code)
			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				t.Helper()

				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program:         program,
					File:            sourceFile.FileName(),
					HasTypeInfo:     true,
					GetRulesForFile: noUnusedVarsConfiguredRules(test.options),
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
			requested := run(test.requestedDemand)
			otherCategory := run(test.otherDemand)
			allEdits := run(rule.EditDemandAll)

			const wantDiagnostics = 5
			for name, diagnostics := range map[string][]rule.RuleDiagnostic{
				"diagnostics only": diagnosticsOnly,
				"requested":        requested,
				"other category":   otherCategory,
				"all edits":        allEdits,
			} {
				if len(diagnostics) != wantDiagnostics {
					t.Fatalf("%s: got %d diagnostics, want %d", name, len(diagnostics), wantDiagnostics)
				}
			}

			for i := range allEdits {
				wantIdentity := noUnusedVarsDiagnosticWithoutEdits(allEdits[i])
				for name, diagnostics := range map[string][]rule.RuleDiagnostic{
					"diagnostics only": diagnosticsOnly,
					"requested":        requested,
					"other category":   otherCategory,
				} {
					if got := noUnusedVarsDiagnosticWithoutEdits(diagnostics[i]); !reflect.DeepEqual(got, wantIdentity) {
						t.Errorf("%s: diagnostic %d changed:\ngot  %#v\nwant %#v", name, i, got, wantIdentity)
					}
				}

				assertNoUnusedVarsDiagnosticHasNoEdits(t, "diagnostics only", i, diagnosticsOnly[i])
				assertNoUnusedVarsDiagnosticHasNoEdits(t, "other category", i, otherCategory[i])

				if test.wantFixes {
					if requested[i].FixesPtr == nil {
						t.Errorf("requested: diagnostic %d has no autofix payload", i)
					}
					if !reflect.DeepEqual(requested[i].Fixes(), allEdits[i].Fixes()) {
						t.Errorf("diagnostic %d differs between autofix and all-edits modes", i)
					}
					if requested[i].Suggestions != nil || allEdits[i].Suggestions != nil {
						t.Errorf("diagnostic %d unexpectedly has suggestions in autofix configuration", i)
					}
				} else {
					if requested[i].Suggestions == nil {
						t.Errorf("requested: diagnostic %d has no suggestion payload", i)
					}
					if !reflect.DeepEqual(requested[i].Suggestions, allEdits[i].Suggestions) {
						t.Errorf("diagnostic %d differs between suggestion and all-edits modes", i)
					}
					if requested[i].FixesPtr != nil || allEdits[i].FixesPtr != nil {
						t.Errorf("diagnostic %d unexpectedly has fixes in suggestion configuration", i)
					}
				}
			}
		})
	}
}

func assertNoUnusedVarsDiagnosticHasNoEdits(t *testing.T, mode string, index int, diagnostic rule.RuleDiagnostic) {
	t.Helper()
	if diagnostic.FixesPtr != nil {
		t.Errorf("%s: diagnostic %d unexpectedly has a fix payload", mode, index)
	}
	if diagnostic.Suggestions != nil {
		t.Errorf("%s: diagnostic %d unexpectedly has a suggestion payload", mode, index)
	}
}

type noUnusedVarsDiagnosticIdentity struct {
	Range    [2]int
	RuleName string
	Message  rule.RuleMessage
	FilePath string
	Severity rule.DiagnosticSeverity
}

func noUnusedVarsDiagnosticWithoutEdits(diagnostic rule.RuleDiagnostic) noUnusedVarsDiagnosticIdentity {
	return noUnusedVarsDiagnosticIdentity{
		Range:    [2]int{diagnostic.Range.Pos(), diagnostic.Range.End()},
		RuleName: diagnostic.RuleName,
		Message:  diagnostic.Message,
		FilePath: diagnostic.FilePath,
		Severity: diagnostic.Severity,
	}
}
