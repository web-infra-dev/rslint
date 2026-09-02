// TestNoConditionalTestsExtras locks in the Rstest-only augmentation the port
// spec requires: the registration source matrix, the registration-shape matrix,
// the parent-walk branch lock-ins that upstream's fixed four-level parent check
// cannot express, and the graceful-degradation shapes of Dimension 4.
package no_conditional_tests

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConditionalTestsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConditionalTestsRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Not a conditional registration at all ----
			{Code: `test("a", () => {});`},
			{Code: `describe("a", () => { test("b", () => {}); });`},
			// The call sits in the condition, so it runs unconditionally.
			{Code: `if (test("a")) {}`},
			{Code: `if (!describe("a")) {}`},
			{Code: `if (x && test("a")) {} else {}`},
			// A condition inside the test body belongs to no-conditional-in-test.
			{Code: `test("a", () => { if (x) { expect(1).toBe(1); } });`},
			{Code: `describe("a", () => { if (x) { helper(); } test("b", () => {}); });`},

			// ---- B. Only `if` counts ----
			{Code: `x ? test("a", () => {}) : null;`},
			{Code: `x && test("a", () => {});`},
			{Code: `x || test("a", () => {});`},
			{Code: `x ?? test("a", () => {});`},
			{Code: `switch (x) { case 1: test("a", () => {}); }`},
			{Code: `for (const c of cs) { test(c, () => {}); }`},
			{Code: `while (x) { test("a", () => {}); }`},
			{Code: `try { test("a", () => {}); } catch {}`},

			// ---- C. Hooks are excluded, matching upstream's name list ----
			{Code: `if (x) { beforeEach(() => {}); }`},
			{Code: `if (x) { afterAll(() => {}); }`},

			// ---- D. The call is not a Rstest registration ----
			{Code: `const test = f; if (x) { test("a", () => {}); }`},
			{Code: `function outer(test) { if (x) { test("a", () => {}); } }`},
			{Code: `import { test } from 'vitest';
if (x) { test("a", () => {}); }`},
			{Code: `import { test } from 'node:test';
if (x) { test("a", () => {}); }`},
			{Code: `import { test } from '@jest/globals';
if (x) { test("a", () => {}); }`},
			{Code: `import { test } from './helpers';
if (x) { test("a", () => {}); }`},
			{Code: `if (x) { expect(1).toBe(1); }`},
			// A non-null assertion or a type assertion around the callee is not
			// a chain the registration parser follows, so the call is not
			// recognised as Rstest and nothing is reported.
			{Code: `if (x) { test!("a", () => {}); }`},
			{Code: `if (x) { (test as any)("a", () => {}); }`},
			{Code: `if (x) { (test satisfies unknown)("a", () => {}); }`},

			// ---- E. The function boundary stops the walk ----
			// The registration is decided by the wrapper's caller, not by the
			// `if` the wrapper happens to sit under.
			{Code: `if (x) { const register = () => { test("a", () => {}); }; register(); }`},
			{Code: `if (x) { function register() { test("a", () => {}); } register(); }`},
			{Code: `if (x) { const o = { register() { test("a", () => {}); } }; }`},
			{Code: `if (x) { const o = { get register() { test("a", () => {}); return 1; } }; }`},
			// The inner registration of a nested pair: its walk stops at the
			// describe callback, so only the outer describe is reported.
			{Code: `describe("a", () => { test("b", () => {}); });`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- 1. The two shapes upstream's fixed parent depth gets wrong ----
			{
				Code: `if (x) { test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				// Braceless: upstream's four-level parent walk misses this.
				Code: `if (x) test("a", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    8,
					EndLine:   1,
					EndColumn: 12,
				}},
			},

			// ---- 2. Branch positions ----
			{
				Code: `if (x) {} else { describe("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 26,
				}},
			},
			{
				Code: `if (x) {} else describe("a", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `if (x) {} else if (y) { test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    25,
					EndLine:   1,
					EndColumn: 29,
				}},
			},
			{
				// The inner `if` puts the call in its condition, but the inner
				// `if` itself is reached through the outer `if`'s then branch,
				// so the registration really is conditional.
				Code: `if (x) { if (test("a")) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 18,
				}},
			},

			// ---- 3. Distance between the `if` and the registration ----
			{
				Code: `if (x) { for (const c of cs) { test(c, () => {}); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    32,
					EndLine:   1,
					EndColumn: 36,
				}},
			},
			{
				Code: `class C { static { if (x) { test("a", () => {}); } } }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    29,
					EndLine:   1,
					EndColumn: 33,
				}},
			},

			// ---- 4. Nesting reports the outermost registration only ----
			{
				Code: `if (x) { describe("a", () => { test("b", () => {}); }); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				// Two sibling registrations under one `if` are two diagnostics.
				Code: `if (x) { test("a", () => {}); test("b", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noConditionalTests",
						Message:   "Avoid using if conditions in a test",
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 14,
					},
					{
						MessageId: "noConditionalTests",
						Message:   "Avoid using if conditions in a test",
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},

			// ---- 5. The registration source matrix ----
			{
				Code: `import { test } from '@rstest/core';
if (x) { test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `import { it as check } from '@rstest/core';
if (x) { check("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 15,
				}},
			},
			{
				Code: `import * as core from '@rstest/core';
if (x) { core.describe("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `const { test } = require('@rstest/core');
if (x) { test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `const core = require('@rstest/core');
if (x) { core.test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { import.meta.rstest.test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 28,
				}},
			},
			{
				Code: `const { describe } = import.meta.rstest;
if (x) { describe("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 18,
				}},
			},
			{
				Code: `import { test } from '@rstest/playwright';
if (x) { test("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 14,
				}},
			},
			{
				Code: `const t = test;
if (x) { t("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 11,
				}},
			},

			// ---- 6. The registration shape matrix ----
			{
				Code: `if (x) { test.skip("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { it.only("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 12,
				}},
			},
			{
				Code: `if (x) { test.todo("a"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { test.each([1, 2])("a %i", (n) => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: "if (x) { describe.each`a`(\"a\", () => {}); }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code: `if (x) { test["skip"]("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { test.concurrent.skipIf(y)("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { test("a", { timeout: 100 }, () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { test("a", () => {}, 100); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) {
  test
    // a comment between the members
    .skip("a", () => {});
}`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      2,
					Column:    3,
					EndLine:   2,
					EndColumn: 7,
				}},
			},

			// ---- 7. Dimension 4: wrapped and degraded call shapes ----
			{
				Code: `if (x) { (test)("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    11,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: `if (x) { ((test))("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    12,
					EndLine:   1,
					EndColumn: 16,
				}},
			},
			{
				Code: `if (x) { test(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				// Optional call and optional member access still register.
				Code: `if (x) { test?.("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				Code: `if (x) { test?.skip("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			// N/A: a malformed or unterminated call cannot be exercised here —
			// the rule tester rejects source with syntactic errors before any
			// rule runs.
			{
				Code: `if (x) { test(...args); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 14,
				}},
			},
			{
				// A comma expression evaluates to its right operand, so this
				// invokes the real Rstest registration.
				Code: `if (x) { (0, test)("a", () => {}); }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noConditionalTests",
					Message:   "Avoid using if conditions in a test",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
		},
	)
}
