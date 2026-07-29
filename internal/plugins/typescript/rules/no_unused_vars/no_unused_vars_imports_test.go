package no_unused_vars

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUnusedVarsImports(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// import: used
		{Code: `import type { Foo } from "./foo"; const bar: Foo = {} as any; console.log(bar);`},
		// namespace import: used
		{Code: `import * as path from "path"; console.log(path.join("a", "b"));`},
		// import equals: used
		{Code: `import path = require("path"); console.log(path.join("a", "b"));`},
		// side-effect import: no binding, not affected
		{Code: `import "path";`},
		// import then re-export: used
		{Code: `import { join } from "path"; export { join };`},
		// re-export with rename: used
		{Code: `import { join } from "path"; export { join as myJoin };`},
		// namespace import re-exported
		{Code: `import * as path from "path"; export { path };`},
		// import used via export default
		{Code: `import { join } from "path"; export default join;`},
		// direct re-export (no local binding, rule doesn't apply)
		{Code: `export { join } from "path";`},
		// import used in multiple places
		{Code: `import { join, resolve } from "path"; console.log(join("a"), resolve("b"));`},
		// default + named import, both used
		{Code: `import Def, { join } from "path"; console.log(Def, join("a"));`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// --- import: unused (with suggestion fixers) ---
		{
			Code: `import { Foo } from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// namespace import: unused
		{
			Code: `import * as ns from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// import equals: unused
		{
			Code: `import path = require("path");`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// type import unused → remove entire line
		{
			Code: `import type { Foo } from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// default import unused → remove entire line
		{
			Code: `import Foo from "./foo";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// alias import unused — reported at the local name (r), not the original name
		{
			Code: `import { resolve as r } from "path";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 21,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// unrelated namespace re-export in the same file should not panic while checking unused imports
		{
			Code: `import { join } from "path"; export * as pathNs from "path";`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ` export * as pathNs from "path";`,
				}},
			}},
		},

		// --- enableAutofixRemoval.imports = true: fix applied as autofix ---
		{
			Code:    `import { Foo } from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
			Output:  []string{``},
		},
		// enableAutofixRemoval.imports = false (explicit): fix applied as suggestion
		{
			Code:    `import { Foo } from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": false}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedImportDeclaration",
					Output:    ``,
				}},
			}},
		},
		// autofix mode: namespace import
		{
			Code:    `import * as ns from "./foo";`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 13}},
			Output:  []string{``},
		},
		// autofix mode: import equals
		{
			Code:    `import path = require("path");`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 8}},
			Output:  []string{``},
		},

		// --- import fix: partial removal ---
		// first specifier unused, second used → remove first + trailing comma
		{
			Code: `import { resolve, join } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// second specifier unused, first used → remove leading comma + second
		{
			Code: `import { join, resolve } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// partial removal with autofix mode
		{
			Code:    `import { join, resolve } from "path"; console.log(join("a", "b"));`,
			Options: map[string]interface{}{"enableAutofixRemoval": map[string]interface{}{"imports": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 16}},
			Output:  []string{`import { join } from "path"; console.log(join("a", "b"));`},
		},
		// middle specifier unused (three specifiers)
		{
			Code: `import { join, resolve, basename } from "path"; console.log(join("a", "b"), basename("c"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join, basename } from "path"; console.log(join("a", "b"), basename("c"));`,
				}},
			}},
		},
		// alias import: partial removal (alias unused, another specifier used)
		{
			Code: `import { join, resolve as r } from "path"; console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; console.log(join("a", "b"));`,
				}},
			}},
		},
		// partial re-export: join re-exported (used), resolve not → resolve reported
		{
			Code: `import { join, resolve } from "path"; export { join };`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 1, Column: 16,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output:    `import { join } from "path"; export { join };`,
				}},
			}},
		},
		// multiline import: partial removal
		{
			Code: `import {
  join,
  resolve
} from "path";
console.log(join("a", "b"));`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unusedVar", Line: 3, Column: 3,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "removeUnusedVar",
					Output: `import {
  join
} from "path";
console.log(join("a", "b"));`,
				}},
			}},
		},
		// three specifiers, two unused → each gets its own suggestion
		{
			Code: `import { join, resolve, basename } from "path"; console.log(resolve("/"));`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { resolve, basename } from "path"; console.log(resolve("/"));`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { join, resolve } from "path"; console.log(resolve("/"));`,
					}},
				},
			},
		},
		// four specifiers, three unused
		{
			Code: `import { a, b, c, d } from "./foo"; console.log(b);`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedVar", Line: 1, Column: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { b, c, d } from "./foo"; console.log(b);`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { a, b, d } from "./foo"; console.log(b);`,
					}},
				},
				{MessageId: "unusedVar", Line: 1, Column: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "removeUnusedVar",
						Output:    `import { a, b, c } from "./foo"; console.log(b);`,
					}},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}

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
			options:         rule_tester.ResolveTestCaseOptions(t, &NoUnusedVarsRule, nil),
			requestedDemand: rule.EditDemandSuggestion,
			otherDemand:     rule.EditDemandAutofix,
		},
		{
			name: "autofix",
			options: rule_tester.ResolveTestCaseOptions(t, &NoUnusedVarsRule, map[string]interface{}{
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

func createNoUnusedVarsProgram(t testing.TB, fileName string, code string) (*compiler.Program, *ast.SourceFile) {
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

func noUnusedVarsConfiguredRules(options []any) func(*ast.SourceFile) []linter.ConfiguredRule {
	return func(*ast.SourceFile) []linter.ConfiguredRule {
		return []linter.ConfiguredRule{{
			Name:             NoUnusedVarsRule.Name,
			Severity:         rule.SeverityError,
			RequiresTypeInfo: true,
			Run: func(ctx rule.RuleContext) rule.RuleListeners {
				return NoUnusedVarsRule.Run(ctx, options)
			},
		}}
	}
}
