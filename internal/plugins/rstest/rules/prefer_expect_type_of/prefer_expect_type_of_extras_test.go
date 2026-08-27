// TestPreferExpectTypeOfExtras locks in the Rstest-only augmentation the port
// spec requires: the expect source matrix, the assertion-factory boundary, the
// matcher-argument contract, trivia-preserving surgical fixes, real-user
// shapes, and graceful degradation. Edit-demand parity lives in
// TestPreferExpectTypeOfEditDemand.
package prefer_expect_type_of

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferExpectTypeOfExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferExpectTypeOfRule,
		[]rule_tester.ValidTestCase{
			// ---- A. The target form already in use ----
			{Code: `expect(value).toBeTypeOf("string");`},
			{Code: `expect.soft(value).toBeTypeOf("string");`},
			{Code: `expect(typeof value).toBeTypeOf("string");`},

			// ---- B. The head argument is not a typeof expression ----
			{Code: `expect(typeof value === "string").toBe(true);`},
			{Code: `expect(!(typeof value)).toBe(true);`},
			{Code: `expect(value).toBe("string");`},
			{Code: `expect().toBe("string");`},
			{Code: `expect(...args).toBe("string");`},
			// SkipParentheses is the spec's unwrap: a TypeScript assertion
			// around the argument is left to the assertion-aware rules.
			{Code: `expect(typeof value as unknown).toBe("string");`},
			{Code: `expect((typeof value)! ).toBe("string");`},

			// ---- C. Assertion factories that cannot carry a typeof value ----
			// poll takes a callback and element takes a locator, so neither
			// shape is a typeof assertion even when written like one.
			{Code: `expect.poll(() => typeof value).toBe("string");`},
			{Code: `expect.element(typeof value).toBe("string");`},
			{Code: `expect.assertions(typeof value);`},
			{Code: `expect.unreachable(typeof value);`},

			// ---- D. Matcher contract ----
			{Code: `expect(typeof value).toMatch("string");`},
			{Code: `expect(typeof value).toStrictEqual("string");`},
			{Code: `expect(typeof value).toBe();`},
			{Code: `expect(typeof value).toBe("string", "extra");`},
			{Code: `expect(typeof value).toBe(...types);`},
			{Code: `expect(typeof value).toEqual(...types);`},
			// Broken chains belong to rstest/valid-expect, which owns Reason.
			{Code: `expect(typeof value).toBe;`},
			{Code: `expect(typeof value);`},
			{Code: `expect(typeof value).resolves.resolves.toBe("string");`},
			// Chai property assertions never execute toBe.
			{Code: `expect(typeof value).to.be.ok;`},
			{Code: `expect(typeof value).to.be.a("string");`},

			// ---- E. A computed identifier key names its matcher at runtime ----
			// The parser records the variable's own text as the member name, so
			// matching it here would report whatever assertion the variable
			// holds. Fixing is impossible for the same reason, and the fix
			// would have to rewrite the variable name.
			{Code: `const toBe = "toStrictEqual";
expect(typeof value)[toBe]("string");`},
			{Code: `const toEqual = "toMatch";
expect(typeof value)[toEqual]("string");`},
			{Code: `expect(typeof value)[matcherName]("string");`},
			{Code: `expect(typeof value)[matchers.toBe]("string");`},

			// ---- F. Reverse sources: foreign and local expects ----
			{Code: `import { expect } from 'vitest';
expect(typeof value).toBe("string");`},
			{Code: `import { expect } from '@jest/globals';
expect(typeof value).toBe("string");`},
			{Code: `import { expect } from '@playwright/test';
expect(typeof value).toBe("string");`},
			{Code: `import { expect } from 'chai';
expect(typeof value).toBe("string");`},
			{Code: `const expect = createAssertionLibrary();
expect(typeof value).toBe("string");`},
			{Code: `custom.expect(typeof value).toBe("string");`},
			// A same-file alias of expect is not followed by the parser, so the
			// call is not a recognised Rstest expect.
			{Code: `import { expect } from '@rstest/core';
const check = expect;
check(typeof value).toBe("string");`},

			// ---- F. Shadowing ----
			{Code: `import { expect } from '@rstest/core';
function run(expect: any) { expect(typeof value).toBe("string"); }`},
			{Code: `import { expect as check } from '@rstest/core';
function run(check: any) { check(typeof value).toBe("string"); }`},
			{Code: `import * as core from '@rstest/core';
function run() { const core = helper(); core.expect(typeof value).toBe("string"); }`},

			// ---- G. Dimension 4: shapes the shared parser declines ----
			// The rule reports only what analysis.ParseExpectCall resolves, so
			// an unrecognised root, a chain that reads a property of the
			// assertion's own result, and an unknown member before the matcher
			// are all silent rather than guessed at.
			{Code: `broken.expect?.(typeof value).toBe("string");`},
			{Code: `expect(typeof value).toBe("string")[0];`},
			{Code: `expect(typeof value).unknownMember.toBe("string");`},

			// N/A: declaration and container variants, function kinds, class
			// members, and overload signatures are unrelated to this rule's
			// single call-expression listener over analysis.ParseExpectCall.
		},
		[]rule_tester.InvalidTestCase{
			// ---- H. Expect source matrix ----
			{
				Code: `import { expect } from '@rstest/core';
expect(typeof value).toBe("string");`,
				Output: []string{`import { expect } from '@rstest/core';
expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 36,
				}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(typeof value).toEqual("string");`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(typeof value).toBe("string");`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(typeof value).toBe("string");`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(typeof value).toBe("string");`,
				Output: []string{`const core = require('@rstest/core');
core.expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code:   `import.meta.rstest.expect(typeof value).toBe("string");`,
				Output: []string{`import.meta.rstest.expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(typeof value).toBe("string");`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(typeof value).toBe("string");`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 1}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(typeof value).toBe("string"));`,
				Output: []string{`test('x', ({ expect }) => expect(value).toBeTypeOf("string"));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 27}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(typeof value).toBe("string"));`,
				Output: []string{`test('x', ctx => ctx.expect(value).toBeTypeOf("string"));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 18}},
			},

			// ---- I. The parts of the call the surgical fix must preserve ----
			{
				// The message argument only expect(...) accepts survives.
				Code:   `expect(typeof value, "value should be a string").toBe("string");`,
				Output: []string{`expect(value, "value should be a string").toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect(value).toBeTypeOf(\"string\")` instead of `expect(typeof value).toBe(\"string\")`",
					Line:      1,
					Column:    1,
				}},
			},
			{
				// expect.soft stays soft rather than being rewritten to expect.
				Code:   `expect.soft(typeof value).toBe("string");`,
				Output: []string{`expect.soft(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `expect(typeof value).not.toBe("string");`,
				Output: []string{`expect(value).not.toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `await expect(typeof value).resolves.toBe("string");`,
				Output: []string{`await expect(value).resolves.toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 7}},
			},
			{
				Code:   `expect(typeof value).toEqual(expectedType);`,
				Output: []string{`expect(value).toBeTypeOf(expectedType);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				// The operand keeps its own parentheses.
				Code:   `expect(typeof (value)).toBe("string");`,
				Output: []string{`expect((value)).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferExpectTypeOf",
					Message:   "Use `expect((value)).toBeTypeOf(\"string\")` instead of `expect(typeof (value)).toBe(\"string\")`",
					Line:      1,
					Column:    1,
				}},
			},
			{
				// Parentheses around the typeof expression are the argument, so
				// only the operator is removed.
				Code:   `expect((typeof value)).toBe("string");`,
				Output: []string{`expect((value)).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `expect(typeof value.deep.property).toBe("function");`,
				Output: []string{`expect(value.deep.property).toBeTypeOf("function");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `expect(typeof value).toBe(` + "`string`" + `);`,
				Output: []string{`expect(value).toBeTypeOf(` + "`string`" + `);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},

			// ---- J. Accessor spellings are preserved ----
			{
				Code:   `expect(typeof value)['toBe']("string");`,
				Output: []string{`expect(value)['toBeTypeOf']("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   "expect(typeof value)[`toEqual`](\"string\");",
				Output: []string{"expect(value)[`toBeTypeOf`](\"string\");"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `expect(typeof value).not["toBe"]("string");`,
				Output: []string{`expect(value).not["toBeTypeOf"]("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},

			// ---- K. Trivia ----
			{
				// A comment between the operator and the operand sits inside the
				// removed span and goes with it; every other comment stays.
				Code: `expect(
  typeof /* why */ value, // reason
).toBe(
  "string", // the type
);`,
				Output: []string{`expect(
  value, // reason
).toBeTypeOf(
  "string", // the type
);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				// A comment between the matcher accessor and its argument list
				// is outside both edits.
				Code:   `expect(typeof value).toBe /* here */ ("string");`,
				Output: []string{`expect(value).toBeTypeOf /* here */ ("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code: `expect(typeof value)
  .not
  .toBe("string");`,
				Output: []string{`expect(value)
  .not
  .toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},

			// ---- L. Optional call and optional chain links ----
			{
				Code:   `expect?.(typeof value).toBe("string");`,
				Output: []string{`expect?.(value).toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},
			{
				Code:   `expect(typeof value)?.toBe("string");`,
				Output: []string{`expect(value)?.toBeTypeOf("string");`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 1, Column: 1}},
			},

			// ---- M. Real-user shapes ----
			{
				Code: `import { expect, test } from '@rstest/core';

test('exposes a string id', () => {
  const user = createUser();
  expect(typeof user.id).toBe('string');
  expect(user.name).toBe('ada');
});`,
				Output: []string{`import { expect, test } from '@rstest/core';

test('exposes a string id', () => {
  const user = createUser();
  expect(user.id).toBeTypeOf('string');
  expect(user.name).toBe('ada');
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 5, Column: 3}},
			},
			{
				Code: `test.each(['a', 1])('checks %s', (value) => {
  expect(typeof value).toEqual(typeof value);
});`,
				Output: []string{`test.each(['a', 1])('checks %s', (value) => {
  expect(value).toBeTypeOf(typeof value);
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferExpectTypeOf", Line: 2, Column: 3}},
			},
			{
				// Two assertions in one file are reported and fixed independently.
				Code: `expect(typeof first).toBe("string");
expect(typeof second).toEqual("number");`,
				Output: []string{`expect(first).toBeTypeOf("string");
expect(second).toBeTypeOf("number");`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferExpectTypeOf", Line: 1, Column: 1},
					{MessageId: "preferExpectTypeOf", Line: 2, Column: 1},
				},
			},
		},
	)
}

// TestPreferExpectTypeOfEditDemand locks in that the diagnostic is identical
// under every edit demand and that only the autofix demands materialize fixes.
func TestPreferExpectTypeOfEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(typeof value).toBe("string");
expect.soft(typeof other)['toEqual']("number");`,
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
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferExpectTypeOfRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferExpectTypeOfRule.Run(ctx, nil)
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
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
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
	for index := range allEdits {
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	if diagnosticsOnly[0].FixesPtr != nil || diagnosticsOnly[1].FixesPtr != nil {
		t.Fatal("EditDemandNone unexpectedly materialized fixes")
	}
	if suggestionOnly[0].FixesPtr != nil || suggestionOnly[1].FixesPtr != nil {
		t.Fatal("EditDemandSuggestion unexpectedly materialized fixes")
	}
	for index, diagnostic := range allEdits {
		if diagnostic.Suggestions != nil {
			t.Fatalf("diagnostic %d unexpectedly materialized suggestions", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, diagnostic.FixesPtr) {
			t.Fatalf("diagnostic %d produced inconsistent fixes between autofix and all demands", index)
		}
		if fixes := *autofixOnly[index].FixesPtr; len(fixes) != 2 {
			t.Fatalf("diagnostic %d produced %d fixes, want 2", index, len(fixes))
		}
	}
}
