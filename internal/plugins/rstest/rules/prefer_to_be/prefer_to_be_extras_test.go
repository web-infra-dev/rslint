// TestPreferToBeExtras locks in branches and edge shapes that the upstream
// suite does not exercise: Rstest expect sources, every relevant tsgo wrapper,
// safe-fix boundaries, semantic corrections and edit-demand behavior.
package prefer_to_be

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

func TestPreferToBeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToBeRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: non-null and satisfies wrappers stay opaque ----
			{Code: `expect(value).toEqual(null!);`},
			{Code: `expect(value).toEqual(1 satisfies number);`},
			// ---- Dimension 4: regexp/object/array values require deep equality ----
			{Code: `expect(value).toEqual(/value/);`},
			{Code: `expect(value).toStrictEqual({ value: 1 });`},
			{Code: `expect(value).toEqual([1]);`},
			// ---- Dimension 4: dynamic, numeric and symbol accessors are not names ----
			{Code: `expect(value)[matcher](1);`},
			{Code: `expect(value)[0](1);`},
			{Code: `expect(value)[Symbol.iterator](1);`},
			// ---- Dimension 4: empty matcher calls ----
			{Code: `expect(value).toEqual();`},
			{Code: `expect(value).to.equal(1);`},
			{Code: `expect.toEqual(1);`},
			{Code: `expect(value).resolves.rejects.toEqual(1);`},
			// ---- Dimension 4: receiver type wrappers are parse boundaries ----
			{Code: `(expect(value) as any).toEqual(1);`},
			{Code: `expect(value)!.toEqual(1);`},
			{Code: `(expect(value) satisfies Assertion).toEqual(1);`},
			// ---- Rstest reality: element assertions do not expose these matchers ----
			{Code: `expect.element(locator).toEqual(1);`},
			// ---- Rstest reality: other assertion libraries and local shadows ----
			{Code: `import { expect } from 'vitest'; expect(value).toEqual(1);`},
			{Code: `import { expect } from '@jest/globals'; expect(value).toEqual(1);`},
			{Code: `import { expect } from 'chai'; expect(value).toEqual(1);`},
			{Code: `const expect = createExpect(); expect(value).toEqual(1);`},
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(value).toEqual(1); }`},
			// ---- Rstest semantic correction: locally shadowed special values ----
			{Code: `const undefined = 1; expect(value).toEqual(undefined);`},
			{Code: `function f(NaN: number) { expect(value).toEqual(NaN); }`},
			// ---- Vitest policy: positive and negative fractional literals ----
			{Code: `expect(value).toEqual(0.3);`},
			{Code: `expect(value).toEqual(-0.3);`},
			// N/A: declaration/container and function/class forms; this rule
			// inspects only parsed assertion calls.
			// N/A: spread/rest in object or binding patterns, empty bodies and
			// overload signatures cannot be matcher arguments of the inspected kind.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: nested parentheses are transparent ----
			{
				Code:   `expect(value).toEqual(((1)));`,
				Output: []string{`expect(value).toBe(((1)));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toEqual((undefined));`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 15}},
			},
			// ---- Dimension 4: type assertions are transparent ----
			{
				Code:   `expect(value).toEqual(<number>1);`,
				Output: []string{`expect(value).toBe(<number>1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// A type-only declaration does not shadow the runtime built-in.
			{
				Code:   `type NaN = number; expect(value).toEqual(NaN);`,
				Output: []string{`type NaN = number; expect(value).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 34}},
			},
			// ---- Dimension 4: optional chain flags and accessor forms ----
			{
				Code:   `expect(value)?.toEqual?.(1);`,
				Output: []string{`expect(value)?.toBe?.(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 16}},
			},
			{
				Code:   `(expect(value)).toEqual(1);`,
				Output: []string{`(expect(value)).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 17}},
			},
			{
				Code:   `(expect(value)?.toEqual)(1);`,
				Output: []string{`(expect(value)?.toBe)(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 17}},
			},
			{
				Code:   `expect(value)["toEqual"](1);`,
				Output: []string{`expect(value)["toBe"](1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value)[`toEqual`](1);",
				Output: []string{"expect(value)[`toBe`](1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// Diverges from @vitest/eslint-plugin: the dedicated matchers declare
			// no type parameters, so type arguments go with the value arguments
			// instead of leaving the fixed assertion failing with TS2558.
			{
				Code:   `expect(null).toEqual<null>(null);`,
				Output: []string{`expect(null).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 14}},
			},
			{
				Code:   `expect(value).not.toBeDefined<never>();`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 19}},
			},
			// toBe shares toEqual's single type parameter, so its type argument
			// stays.
			{
				Code:   `expect(value).toEqual<number>(1);`,
				Output: []string{`expect(value).toBe<number>(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// A comment inside the removed type arguments withholds the fix, as
			// it does inside the removed value arguments.
			{
				Code:   `expect(null).toEqual</* keep */ null>(null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 14}},
			},
			// ---- Dimension 4: a trailing call cannot replace matcher arguments ----
			{
				Code:   `expect(value).toEqual(1)();`,
				Output: []string{`expect(value).toBe(1)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Dimension 3: comments outside edited arguments survive ----
			{
				Code:   `expect(value). /* keep */ toEqual(null);`,
				Output: []string{`expect(value). /* keep */ toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 27}},
			},
			{
				Code:   `expect(value).toEqual /* keep */ (null);`,
				Output: []string{`expect(value).toBeNull /* keep */ ();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 15}},
			},
			// ---- Dimension 3: comments inside removed arguments withhold fixes ----
			{
				Code:   `expect(value).toEqual(/* keep */ null);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).not /* keep */ .toEqual(undefined);`,
				Output: []string{`expect(value) /* keep */ .toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeDefined", Line: 1, Column: 31}},
			},
			// Locks in the upstream argument-count branch: only the first
			// argument selects toBe, and toBe keeps the authored argument list.
			{
				Code:   `expect(value).toEqual(1, 2);`,
				Output: []string{`expect(value).toBe(1, 2);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Real-user: eslint-plugin-jest#1260 negative literal regression ----
			{
				Code:   `expect(value).toEqual(-1);`,
				Output: []string{`expect(value).toBe(-1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Real-user: eslint-plugin-jest#1131 computed accessor regression ----
			{
				Code:   `expect(value)['toEqual'](false);`,
				Output: []string{`expect(value)['toBe'](false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 15}},
			},
			// ---- Real-user: eslint-plugin-jest#1282 trailing-comma regression ----
			{
				Code:   `expect(value).toEqual(null,);`,
				Output: []string{`expect(value).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 1, Column: 15}},
			},
			// ---- Rstest expect factories ----
			{
				Code:   `expect.soft(value).toEqual(1);`,
				Output: []string{`expect.soft(value).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 20}},
			},
			{
				Code:   `expect(value, 'status').toStrictEqual(1);`,
				Output: []string{`expect(value, 'status').toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 25}},
			},
			{
				Code:   `await expect.poll(() => value).toEqual(1);`,
				Output: []string{`await expect.poll(() => value).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 32}},
			},
			// ---- Rstest source resolution ----
			{
				Code: `import { expect as check } from '@rstest/core';
check(value).toEqual(null);`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(value).toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Line: 2, Column: 14}},
			},
			{
				Code: `import { expect } from 'rstack/test';
expect(value).toEqual(1);`,
				Output: []string{`import { expect } from 'rstack/test';
expect(value).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 2, Column: 15}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(value).toEqual(1);`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(value).toBe(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 2, Column: 15}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(value).toEqual(undefined);`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 2, Column: 20}},
			},
			{
				Code:   `import.meta.rstest.expect(value).toStrictEqual(NaN);`,
				Output: []string{`import.meta.rstest.expect(value).toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Line: 1, Column: 34}},
			},
			{
				Code: `const { expect: check } = import.meta.rstest;
check(value).toEqual('ready');`,
				Output: []string{`const { expect: check } = import.meta.rstest;
check(value).toBe('ready');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 2, Column: 14}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(value).toEqual(true);`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 2, Column: 15}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(value).toEqual(1));`,
				Output: []string{`test('x', ctx => ctx.expect(value).toBe(1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 36}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(value).toEqual(1));`,
				Output: []string{`test('x', ({ expect }) => expect(value).toBe(1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Line: 1, Column: 41}},
			},
			// ---- Dimension 4: nested assertions remain independent ----
			{
				Code:   `expect(expect(value).toEqual(1)).toEqual(true);`,
				Output: []string{`expect(expect(value).toBe(1)).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useToBe", Line: 1, Column: 34},
					{MessageId: "useToBe", Line: 1, Column: 22},
				},
			},
			// Locks in primitive-literal arm and its exact multiline range.
			{
				Code: `expect(value)
  .toEqual('ready');`,
				Output: []string{`expect(value)
  .toBe('ready');`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBe", Message: "Use `toBe` instead", Line: 2, Column: 4, EndLine: 2, EndColumn: 11}},
			},
			// Locks in not.toBeDefined arm: inversion removes not.
			{
				Code:   `expect(value).not.toBeDefined();`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Message: "Use `toBeUndefined()` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 30}},
			},
			// Locks in the special-matcher argument-removal branch.
			{
				Code:   `expect(value).not.toBeDefined('extra');`,
				Output: []string{`expect(value).toBeUndefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeUndefined", Line: 1, Column: 19}},
			},
			// Locks in undefined-with-not arm and fixes an upstream Vitest semantic bug.
			{
				Code: `expect(value)
  .not.toEqual(undefined);`,
				Output: []string{`expect(value).toBeDefined();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeDefined", Message: "Use `toBeDefined()` instead", Line: 2, Column: 8, EndLine: 2, EndColumn: 15}},
			},
			// Locks in null and NaN arms: not remains semantically meaningful.
			{
				Code:   `expect(value).not.toEqual(null);`,
				Output: []string{`expect(value).not.toBeNull();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNull", Message: "Use `toBeNull()` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}},
			},
			{
				Code:   `expect(value).not.toEqual(NaN);`,
				Output: []string{`expect(value).not.toBeNaN();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useToBeNaN", Message: "Use `toBeNaN()` instead", Line: 1, Column: 19, EndLine: 1, EndColumn: 26}},
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
					Name: PreferToBeRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferToBeRule.Run(ctx, nil)
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
