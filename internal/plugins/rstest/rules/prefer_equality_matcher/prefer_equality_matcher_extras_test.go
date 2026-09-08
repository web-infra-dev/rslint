// TestPreferEqualityMatcherExtras locks in branches and edge shapes that the
// upstream test suite does not exercise: the full truth table, Rstest expect
// sources, factory and matcher boundaries, safe source edits, and edit-demand
// behavior. The upstream mirror lives in prefer_equality_matcher_upstream_test.go.
package prefer_equality_matcher

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func equalityExtrasError(column int, output func(string) string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId:   "useEqualityMatcher",
		Line:        1,
		Column:      column,
		Suggestions: equalitySuggestions(output),
	}
}

func TestPreferEqualityMatcherExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferEqualityMatcherRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream comparison gate: loose and non-equality operators.
			{Code: `expect(a == b).toBe(true);`},
			{Code: `expect(a != b).toEqual(false);`},
			{Code: `expect(a > b).toStrictEqual(true);`},
			// Locks in upstream matcher gate and first-argument boolean gate.
			{Code: `expect(a === b).toContain(true);`},
			{Code: `expect(a === b).toBe();`},
			{Code: `expect(a === b).toBe(flag);`},
			{Code: `expect(a === b).toBe(...flags);`},
			{Code: `expect(a === b).toBe(true satisfies boolean);`},
			{Code: `expect(a === b).toBe(true!);`},

			// ---- Dimension 4: matcher shape and access forms ----
			{Code: `expect(a === b).toBe;`},
			{Code: `expect(a === b).to.be.equal(true);`},
			{Code: `expect(a === b)[matcherName](true);`},
			{Code: `expect(a === b)[0](true);`},

			// ---- Dimension 4: receiver wrappers are parser boundaries ----
			{Code: `expect(a === b)!.toBe(true);`},
			{Code: `(expect(a === b) as any).toBe(true);`},
			{Code: `(expect(a === b) satisfies Assertion).toBe(true);`},

			// ---- Rstest source and shadowing boundaries ----
			{Code: `import { expect } from 'vitest'; expect(a === b).toBe(true);`},
			{Code: `import { expect } from '@jest/globals'; expect(a === b).toBe(true);`},
			{Code: `import { expect } from 'chai'; expect(a === b).toBe(true);`},
			{Code: `const expect = makeExpect(); expect(a === b).toBe(true);`},
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(a === b).toBe(true); }`},

			// No Entry gate: poll is rejected because its subject is a function,
			// while element naturally remains outside the strict-comparison shape.
			{Code: `expect.poll(() => a === b).toBe(true);`},
			{Code: `expect.element(locator).toBe(true);`},

			// N/A: declaration/container and function-kind rows do not apply to
			// this rule's call-expression listener. Spread/rest and empty-body
			// forms cannot be the strict binary subject selected by the rule.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Complete 2 x 2 x 2 truth table ----
			{Code: `expect(a === b).toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Message: "Prefer using one of the equality matchers instead", Line: 1, Column: 17, EndLine: 1, EndColumn: 24, Suggestions: equalitySuggestions(func(m string) string { return `expect(a).` + m + `(b);` })}}},
			{Code: `expect(a === b).toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).not.` + m + `(b);` })}},
			{Code: `expect(a !== b).toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).not.` + m + `(b);` })}},
			{Code: `expect(a !== b).toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `(b);` })}},
			{Code: `expect(a === b).not.toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(21, func(m string) string { return `expect(a).not.` + m + `(b);` })}},
			{Code: `expect(a === b).not.toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(21, func(m string) string { return `expect(a).` + m + `(b);` })}},
			{Code: `expect(a !== b).not.toEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(21, func(m string) string { return `expect(a).` + m + `(b);` })}},
			{Code: `expect(a !== b).not.toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(21, func(m string) string { return `expect(a).not.` + m + `(b);` })}},
			{
				Code: `expect(a === b)
  .resolves
  .toStrictEqual(true);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useEqualityMatcher",
					Line:      3,
					Column:    4,
					EndLine:   3,
					EndColumn: 17,
					Suggestions: equalitySuggestions(func(m string) string {
						return `expect(a)
  .resolves
  .` + m + `(b);`
					}),
				}},
			},

			// ---- Dimension 4: parentheses and type assertions ----
			{Code: `expect(((a === b))).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(21, func(m string) string { return `expect(((a))).` + m + `(b);` })}},
			{Code: `expect((a as number) === (b as number)).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(41, func(m string) string { return `expect(a as number).` + m + `(b as number);` })}},
			{Code: `expect(a === b).toBe((true as boolean));`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `((b));` })}},
			// A type assertion written for the boolean literal cannot survive on
			// the operand that replaces it: `b as const` does not compile.
			{Code: `expect(a === b).toBe(true as const);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `(b);` })}},
			{Code: `expect(a === b).toBe(<const>true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `(b);` })}},
			// Parentheses around a comma expression are load-bearing: dropping
			// them would splice a second argument into the rewritten call.
			{Code: `expect((f(), a) === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(24, func(m string) string { return `expect((f(), a)).` + m + `(b);` })}},
			{Code: `expect(a === (f(), b)).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(24, func(m string) string { return `expect(a).` + m + `((f(), b));` })}},

			// ---- Accessor syntax and modifier preservation ----
			{Code: `expect(a === b)['toBe'](true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a)['` + m + `'](b);` })}},
			{Code: "expect(a === b)[`toBe`](true);", Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return "expect(a)[`" + m + "`](b);" })}},
			{Code: `expect(a === b)?.toBe(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(18, func(m string) string { return `expect(a)?.not.` + m + `(b);` })}},
			{Code: `expect(a === b).toBe?.(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `?.(b);` })}},
			{Code: `expect(a !== b).rejects.not.toStrictEqual(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(29, func(m string) string { return `expect(a).rejects.` + m + `(b);` })}},
			{Code: `expect(a === b).not.resolves.toBe(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(30, func(m string) string { return `expect(a).resolves.` + m + `(b);` })}},
			// ---- UTF-16 report and edit ranges ----
			{Code: `expect("😀" === value).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(24, func(m string) string { return `expect("😀").` + m + `(value);` })}},

			// ---- Rstest factories and expect roots ----
			{Code: `expect.soft(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(22, func(m string) string { return `expect.soft(a).` + m + `(b);` })}},
			// No factory-specific gate: a binary argument is handled uniformly,
			// even for factories whose ordinary argument has another shape.
			{Code: `expect.poll(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(22, func(m string) string { return `expect.poll(a).` + m + `(b);` })}},
			{Code: `expect.element(a !== b).toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(25, func(m string) string { return `expect.element(a).` + m + `(b);` })}},
			{Code: `expect(a === b, 'values should match').toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(40, func(m string) string { return `expect(a, 'values should match').` + m + `(b);` })}},
			{Code: `import { expect as check } from '@rstest/core';
check(a === b, 'message').toBe(false);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 2, Column: 27, Suggestions: equalitySuggestions(func(m string) string {
				return `import { expect as check } from '@rstest/core';
check(a, 'message').not.` + m + `(b);`
			})}}},
			{Code: `const { expect } = require('@rstest/core');
expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 2, Column: 17, Suggestions: equalitySuggestions(func(m string) string {
				return `const { expect } = require('@rstest/core');
expect(a).` + m + `(b);`
			})}}},
			{Code: `import * as core from '@rstest/core';
core.expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 2, Column: 22, Suggestions: equalitySuggestions(func(m string) string {
				return `import * as core from '@rstest/core';
core.expect(a).` + m + `(b);`
			})}}},
			{Code: `import { expect } from '@rstest/playwright';
expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 2, Column: 17, Suggestions: equalitySuggestions(func(m string) string {
				return `import { expect } from '@rstest/playwright';
expect(a).` + m + `(b);`
			})}}},
			{Code: `import.meta.rstest.expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(36, func(m string) string { return `import.meta.rstest.expect(a).` + m + `(b);` })}},
			{Code: `const api = import.meta.rstest;
api.expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 2, Column: 21, Suggestions: equalitySuggestions(func(m string) string {
				return `const api = import.meta.rstest;
api.expect(a).` + m + `(b);`
			})}}},
			{Code: `test('compares values', ({ expect }) => expect(a !== b).toBe(false));`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(57, func(m string) string { return `test('compares values', ({ expect }) => expect(a).` + m + `(b));` })}},
			{Code: `test('compares values', ctx => ctx.expect(a === b).toBe(true));`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(52, func(m string) string { return `test('compares values', ctx => ctx.expect(a).` + m + `(b));` })}},

			// ---- Chai multi-matcher boundary ----
			// Only the first call-style matcher is inspected. A later matcher is
			// left outside the three edits.
			{Code: `expect(a === b).toBe(true).toEqual(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `(b).toEqual(false);` })}},

			// ---- Real-user: issue #2006 empty expect argument regression ----
			// The parser sees a real matcher, but the shared subject gate leaves
			// `expect()` alone without indexing a missing argument.
			// (The invalid companion proves a later valid assertion still runs.)
			{Code: `expect(); expect(a === b).toBe(true);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(27, func(m string) string { return `expect(); expect(a).` + m + `(b);` })}},
			// ---- Real-user: preserve comments and trailing commas ----
			{Code: `expect(a /* left */ === b /* right */, 'message').toBe(true,);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 1, Column: 51}}},
			{Code: `expect(a === b).resolves /* keep */ .toBe(false,);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(38, func(m string) string { return `expect(a).resolves.not /* keep */ .` + m + `(b,);` })}},
			{Code: `expect(a === b).not /* keep */ .toBe(false);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(33, func(m string) string { return `expect(a) /* keep */ .` + m + `(b);` })}},
			{Code: `expect(a === b).toBe(/* expected */ true,);`, Errors: []rule_tester.InvalidTestCaseError{equalityExtrasError(17, func(m string) string { return `expect(a).` + m + `(/* expected */ b,);` })}},
			// The whole type-asserted argument is replaced, so a comment inside
			// it would be lost; the diagnostic is kept without suggestions.
			{Code: `expect(a === b).toBe(true /* keep */ as const);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useEqualityMatcher", Line: 1, Column: 17}}},
		},
	)
}

// TestPreferEqualityMatcherEditDemand locks in that the diagnostic is stable
// and the three suggestions are constructed only when requested.
func TestPreferEqualityMatcherEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(a === b, 'message').resolves.not['toBe'](false,);`,
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
					Name:     PreferEqualityMatcherRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferEqualityMatcherRule.Run(ctx, nil)
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

	diagnostics := map[rule.EditDemand]rule.RuleDiagnostic{
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
	want := withoutEdits(diagnostics[rule.EditDemandAll])
	for demand, diagnostic := range diagnostics {
		if got := withoutEdits(diagnostic); !reflect.DeepEqual(got, want) {
			t.Errorf("demand %d changed diagnostic identity:\ngot  %#v\nwant %#v", demand, got, want)
		}
		if diagnostic.FixesPtr != nil {
			t.Errorf("demand %d unexpectedly materialized autofixes", demand)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix} {
		if diagnostics[demand].Suggestions != nil {
			t.Errorf("demand %d unexpectedly materialized suggestions", demand)
		}
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandSuggestion, rule.EditDemandAll} {
		suggestions := diagnostics[demand].Suggestions
		if suggestions == nil || len(*suggestions) != 3 {
			t.Fatalf("demand %d suggestions = %#v, want three", demand, suggestions)
		}
		for index, suggestion := range *suggestions {
			if len(suggestion.FixesArr) < 3 {
				t.Errorf("demand %d suggestion %d has %d fixes, want at least 3", demand, index, len(suggestion.FixesArr))
			}
		}
	}
}
