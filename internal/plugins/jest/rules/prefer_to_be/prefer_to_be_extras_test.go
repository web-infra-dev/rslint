// TestPreferToBeExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch, Dimension 4 row or real-user regression it covers, so
// future refactors cannot silently regress it without breaking a named lock-in.
package prefer_to_be_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_to_be"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToBeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_to_be.PreferToBeRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: type wrappers not followed by upstream ----
			{Code: `expect(value).toEqual(null!);`},
			{Code: `expect(value).toEqual(1 satisfies number);`},
			// ---- Dimension 4: dynamic and non-string element access ----
			{Code: `expect(value)[matcher](1);`},
			{Code: `expect(value)[0](1);`},
			{Code: `expect(value)[Symbol.iterator](1);`},
			// ---- Dimension 4: graceful degradation for empty arguments ----
			{Code: `expect(value).toEqual();`},
			// ---- Dimension 4: parenthesized optional-chain boundaries ----
			// ESTree exposes a ChainExpression here, so upstream's member walk
			// stops instead of treating the outer call as a Jest assertion.
			{Code: `(expect(value)?.toEqual)(1);`},
			{Code: `(expect(value)?.not).toBeUndefined();`},
			// ---- Dimension 4: object, array and regexp literals retain deep equality ----
			{Code: `expect(value).toEqual({ value: 1 });`},
			{Code: `expect(value).toEqual([1]);`},
			{Code: `expect(value).toEqual(/value/);`},
			// N/A: declaration/container forms; the rule targets call expressions.
			// N/A: class/function nesting; each assertion is parsed independently.
			// N/A: object spread, binding rest, empty bodies and overload signatures;
			// none can occupy a matcher-argument expression inspected by this rule.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single and nested parenthesized arguments ----
			{
				Code:   `expect(value).toEqual(((1)));`,
				Output: []string{`expect(value).toBe(((1)));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toEqual((null));`,
				Output: []string{`expect(value).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toEqual((NaN));`,
				Output: []string{`expect(value).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toEqual((undefined));`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 15}},
			},
			// ---- Dimension 4: optional member and optional call chains ----
			{
				Code:   `expect(value)?.toEqual?.(1);`,
				Output: []string{`expect(value)?.toBe?.(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 16}},
			},
			{
				Code:   `expect(value)?.not.toBeUndefined?.();`,
				Output: []string{`expect(value)?.toBeDefined?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeDefined", Line: 1, Column: 20}},
			},
			{
				Code:   `expect("a string")["not"]["toBe"](undefined);`,
				Output: []string{`expect("a string")['toBeDefined']();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeDefined", Line: 1, Column: 27}},
			},
			// ---- Dimension 4: type assertions are transparent ----
			{
				Code:   `expect(value).toEqual(<number>1);`,
				Output: []string{`expect(value).toBe(<number>1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Dimension 4: string and template-literal accessor forms ----
			{
				Code:   "expect(value)[`toEqual`](1);",
				Output: []string{"expect(value)['toBe'](1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// Diverges from upstream: replaceAccessorFixer rewrites the key
			// identifier in place, producing a reference to an undeclared
			// `toBe`. A computed key is quoted so the fixed code still runs.
			{
				Code:   `const toEqual = 'toEqual'; expect(value)[toEqual](1);`,
				Output: []string{`const toEqual = 'toEqual'; expect(value)['toBe'](1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 42}},
			},
			// Diverges from upstream: the dedicated matchers declare no type
			// parameters, so type arguments are removed with the value
			// arguments instead of leaving the assertion failing with TS2558.
			{
				Code:   `expect(null).toEqual<null>(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 14}},
			},
			{
				Code:   `expect(value).toEqual<number>(1);`,
				Output: []string{`expect(value).toBe<number>(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// Diverges from upstream: an outer call in the same chain resolves
			// back to this matcher call, which upstream reports twice at the
			// same location.
			{
				Code:   `expect(promise).resolves.toEqual(null).then(done);`,
				Output: []string{`expect(promise).resolves.toBeNull().then(done);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 26}},
			},
			// ---- Dimension 4: same-kind nesting reports each independent assertion ----
			{
				Code: `expect(expect(value).toEqual(1)).toEqual(true);`,
				Output: []string{
					`expect(expect(value).toBe(1)).toBe(true);`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 34},
					{MessageId: "useToBe", Line: 1, Column: 22},
				},
			},
			// ---- Real-user: eslint-plugin-jest#1260 negative literal regression ----
			{
				Code:   `expect(value).toEqual(-1n);`,
				Output: []string{`expect(value).toBe(-1n);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Real-user: eslint-plugin-jest#1131 computed accessor regression ----
			{
				Code:   `expect(value)["toEqual"](false);`,
				Output: []string{`expect(value)['toBe'](false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Real-user: eslint-plugin-jest#1282 trailing-comma regression ----
			{
				Code:   `expect(value).toEqual(undefined,);`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 15}},
			},
			// Locks in removeExtraArgumentsFixer semantics: leading comments are
			// outside the removed range, while trailing argument trivia is removed.
			{
				Code:   `expect(value).toEqual(/* keep */ null /* drop */);`,
				Output: []string{`expect(value).toBeNull(/* keep */ );`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 15}},
			},
			// Locks in reportPreferToBe special-matcher branch: exact messages and ranges.
			{
				Code: `expect(value)
  .toEqual(null);`,
				Output: []string{`expect(value)
  .toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToBeNull", Message: "Use `toBeNull` instead",
					Line: 2, Column: 4, EndLine: 2, EndColumn: 11,
				}},
			},
			// Locks in not.toBeDefined arm: inversion removes not.
			{
				Code:   `expect(value).not.toBeDefined();`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Message: "Use `toBeUndefined` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 30}},
			},
			// Locks in reportPreferToBe's conditional argument removal for a
			// no-argument matcher reached through the pre-equality branch.
			{
				Code:   `expect(value).not.toBeDefined('extra');`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 19}},
			},
			// Locks in undefined-with-not arm: inversion becomes toBeDefined.
			{
				Code:   `expect(value).not.toEqual(undefined);`,
				Output: []string{`expect(value).toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeDefined", Message: "Use `toBeDefined` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}},
			},
			// Locks in NaN arm: not is retained.
			{
				Code:   `expect(value).not.toEqual(NaN);`,
				Output: []string{`expect(value).not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Message: "Use `toBeNaN` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}},
			},
			// Locks in primitive arm: fractional numbers remain eligible for Jest.
			{
				Code:   `expect(value).toEqual(-0.3);`,
				Output: []string{`expect(value).toBe(-0.3);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Message: "Use `toBe` when expecting primitive literals", Line: 1, Column: 15, EndLine: 1, EndColumn: 22}},
			},
		},
	)
}

func TestPreferToBeEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(value).not.toEqual(undefined);`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name: prefer_to_be.PreferToBeRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return prefer_to_be.PreferToBeRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			}},
		})
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d: diagnostics = %d, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	want := withoutEdits(allEdits)
	for demand, diagnostic := range map[rule.EditDemand]rule.RuleDiagnostic{
		rule.EditDemandNone: diagnosticsOnly, rule.EditDemandAutofix: autofixOnly,
		rule.EditDemandSuggestion: suggestionOnly,
	} {
		if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d diagnostic changed:\ngot  %#v\nwant %#v", demand, got, want)
		}
	}
	if diagnosticsOnly.FixesPtr != nil || suggestionOnly.FixesPtr != nil {
		t.Fatal("autofixes were materialized without being requested")
	}
	if autofixOnly.FixesPtr == nil || allEdits.FixesPtr == nil ||
		!reflect.DeepEqual(*autofixOnly.FixesPtr, *allEdits.FixesPtr) {
		t.Fatal("requested autofixes do not match all-edits output")
	}
	for _, diagnostic := range []rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		if diagnostic.Suggestions != nil {
			t.Fatal("prefer-to-be unexpectedly materialized suggestions")
		}
	}
}
