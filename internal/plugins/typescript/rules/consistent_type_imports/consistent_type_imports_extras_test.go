package consistent_type_imports

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestConsistentTypeImportsExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. The upstream migration lives in
// consistent_type_imports_upstream_test.go; each case below names its branch or
// Dimension 4 shape.
func TestConsistentTypeImportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentTypeImportsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: a parenthesized value reference stays a value use ----
		{Code: `import Foo from 'foo'; const value = (((Foo)));`},
		// ---- Dimension 4: element access in value position stays a value use ----
		{Code: `import * as Foo from 'foo'; Foo['run']();`},
		// ---- Non-FP: a computed type key's argument is a runtime value ----
		{Code: `import key from 'foo'; type T = { [key()]: string };`},
		// Locks in upstream ImportDeclaration arm 4: an unused binding does not become a type-only binding.
		{Code: `import { Unused, Value } from 'foo'; const value = Value;`},
		// ---- Real-user: typescript-eslint#7527 import attributes prevent an unsafe rewrite ----
		{Code: `import Type from 'foo' with { type: 'json' }; type T = Type;`},
		// ---- Real-user: typescript-eslint#2455 the classic JSX factory is an implicit value use ----
		{Code: `import React from 'react'; export const C: React.FC = () => <div />;`, FileName: "test.tsx"},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: nested qualified type name ----
		{
			Code:   `import Foo from 'foo'; type T = Foo.Bar.Baz;`,
			Output: []string{`import type Foo from 'foo'; type T = Foo.Bar.Baz;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "typeOverValue", Line: 1, Column: 1}},
		},
		// Locks in upstream fixToTypeImportDeclaration() namespace arm: split a type namespace from a value default.
		{
			Code: `import Value, * as Type from 'foo'; Value(); type T = Type.Foo;`,
			Output: []string{`import type * as Type from 'foo';
import Value from 'foo'; Value(); type T = Type.Foo;`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "someImportsAreOnlyTypes", Line: 1, Column: 1}},
		},
		// Locks in upstream fixRemoveTypeSpecifierFromImportDeclaration(): preserve a following comment.
		{
			Code:    `import type /* keep */ Foo from 'foo'; type T = Foo;`,
			Options: map[string]interface{}{"prefer": "no-type-imports"},
			Output:  []string{`import /* keep */ Foo from 'foo'; type T = Foo;`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "avoidImportType", Line: 1, Column: 1}},
		},
		// Locks in upstream fixRemoveTypeSpecifierFromImportSpecifier(): preserve a following comment.
		{
			Code:    `import { type /* keep */ Foo } from 'foo'; type T = Foo;`,
			Options: map[string]interface{}{"prefer": "no-type-imports"},
			Output:  []string{`import { /* keep */ Foo } from 'foo'; type T = Foo;`},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "avoidImportType", Line: 1, Column: 10}},
		},
	})
}

func TestConsistentTypeImportsEditDemand(t *testing.T) {
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`import { A, B } from 'foo'; const value: A = B();`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program),
			File:    sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     ConsistentTypeImportsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return ConsistentTypeImportsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnostics := map[rule.EditDemand][]rule.RuleDiagnostic{
		rule.EditDemandNone:       run(rule.EditDemandNone),
		rule.EditDemandAutofix:    run(rule.EditDemandAutofix),
		rule.EditDemandSuggestion: run(rule.EditDemandSuggestion),
		rule.EditDemandAll:        run(rule.EditDemandAll),
	}
	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	wantIdentity := withoutEdits(diagnostics[rule.EditDemandAll][0])
	for demand, result := range diagnostics {
		if got := withoutEdits(result[0]); !reflect.DeepEqual(got, wantIdentity) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, wantIdentity)
		}
		if result[0].Suggestions != nil {
			t.Errorf("demand %d materialized suggestions", demand)
		}
	}
	if diagnostics[rule.EditDemandNone][0].FixesPtr != nil || diagnostics[rule.EditDemandSuggestion][0].FixesPtr != nil {
		t.Error("non-autofix demand materialized fixes")
	}
	if !reflect.DeepEqual(diagnostics[rule.EditDemandAutofix][0].FixesPtr, diagnostics[rule.EditDemandAll][0].FixesPtr) {
		t.Error("autofix and all-edits demands produced different fixes")
	}
	if fixes := diagnostics[rule.EditDemandAll][0].FixesPtr; fixes == nil || len(*fixes) == 0 {
		t.Error("all-edits demand produced no fixes")
	}
}
