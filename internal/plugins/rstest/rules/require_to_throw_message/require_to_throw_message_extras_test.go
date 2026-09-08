// TestRequireToThrowMessageExtras covers Rstest expect sources, parser
// boundaries, tsgo-specific call shapes, and branch lock-ins beyond the
// upstream suite in require_to_throw_message_upstream_test.go.
package require_to_throw_message

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireToThrowMessageExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireToThrowMessageRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream condition arm 1: unrelated matcher names do not report.
			{Code: `expect(run).toBeDefined();`},
			{Code: `expect(run).throw();`},
			{Code: `expect(run).throws();`},
			{Code: `expect(run).throw().toThrow();`},
			{Code: `expect(run).toBeDefined().toThrow();`},
			// Locks in upstream condition arm 2: any argument is an expected error.
			{Code: `expect(run).toThrow(Error);`},
			{Code: `expect(run).toThrow(undefined);`},
			{Code: `expect(run).toThrow(null);`},
			{Code: `expect(run).toThrow(...messages);`},
			// Locks in upstream condition arm 3: not exempts an empty matcher call.
			{Code: `await expect(run()).rejects.not.toThrow();`},
			{Code: `expect.soft(run).not.toThrowError();`},
			{Code: `expect(run)['not'].toThrow();`},
			{Code: "expect(run)[`not`].toThrowError();"},

			// ---- Dimension 4: dynamic and unsupported accessor keys ----
			{Code: `expect(run)[matcher]();`},
			{Code: `expect(run)[0]();`},
			{Code: `expect(run)[Symbol.iterator]();`},
			// N/A: a private identifier cannot legally access the result of an
			// expect call, so there is no valid private matcher spelling.

			// ---- Dimension 4: TypeScript receiver wrappers are parser boundaries ----
			{Code: `expect(run)!.toThrow();`},
			{Code: `(expect(run) as any).toThrow();`},
			{Code: `(expect(run) satisfies Assertion).toThrow();`},
			{Code: `(<Assertion>expect(run)).toThrow();`},

			// ---- Dimension 4: local and foreign expect roots ----
			{Code: `const expect = createAssertionLibrary(); expect(run).toThrow();`},
			{Code: `import { expect } from 'vitest'; expect(run).toThrow();`},
			{Code: `import { expect } from '@jest/globals'; expect(run).toThrow();`},
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(run).toThrow(); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(run).toThrow(); }`},

			// ---- Dimension 4: malformed or non-call matcher shapes degrade safely ----
			{Code: `expect(run).toThrow;`},
			{Code: `expect.toThrow();`},
			{Code: `expect.not.toThrow();`},
			{Code: `expect(run).resolves.rejects.toThrow();`},
			// N/A: declaration/container/function variants, nesting boundaries,
			// spread assignments, destructuring and body-less declarations are not
			// inspected by this call-expression-only rule. The empty matcher argument
			// list is the rule's triggering branch and is covered below.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream matcher-name arm 1.
			{Code: `expect(run).toThrow();`, Errors: missingError(1, 13, 20, "toThrow")},
			// Locks in upstream matcher-name arm 2.
			{Code: `expect(run).toThrowError();`, Errors: missingError(1, 13, 25, "toThrowError")},
			// Locks in upstream modifier predicate's non-not branch.
			{Code: `await expect(run()).rejects.toThrow();`, Errors: missingError(1, 29, 36, "toThrow")},

			// ---- Dimension 4: parenthesized receivers and matcher callees ----
			{Code: `(expect(run)).toThrow();`, Errors: missingError(1, 15, 22, "toThrow")},
			{Code: `((expect(run))).toThrowError();`, Errors: missingError(1, 17, 29, "toThrowError")},
			{Code: `(expect(run).toThrow)();`, Errors: missingError(1, 14, 21, "toThrow")},

			// ---- Dimension 4: static accessor forms ----
			{Code: `expect(run)['toThrow']();`, Errors: missingError(1, 13, 22, "toThrow")},
			{Code: "expect(run)[`toThrowError`]();", Errors: missingError(1, 13, 27, "toThrowError")},
			// ---- Dimension 4: optional matcher calls ----
			{Code: `expect(run).toThrow?.();`, Errors: missingError(1, 13, 20, "toThrow")},
			{Code: `expect(run)?.toThrow();`, Errors: missingError(1, 14, 21, "toThrow")},
			{Code: `expect?.(run).toThrowError();`, Errors: missingError(1, 15, 27, "toThrowError")},
			{Code: `expect().toThrow();`, Errors: missingError(1, 10, 17, "toThrow")},
			{Code: `expect(run).toThrow(/* expected error */);`, Errors: missingError(1, 13, 20, "toThrow")},

			// ---- Rstest assertion factories ----
			{Code: `expect.soft(run).toThrow();`, Errors: missingError(1, 18, 25, "toThrow")},
			// Rstest's types exclude these matchers from poll/element, but this rule
			// intentionally keeps Vitest's syntax-only contract with no Entry gate.
			{Code: `expect.poll(run).toThrow();`, Errors: missingError(1, 18, 25, "toThrow")},
			{Code: `expect.element(run).toThrowError();`, Errors: missingError(1, 21, 33, "toThrowError")},

			// ---- Rstest expect sources ----
			{Code: `import { expect } from '@rstest/core';
expect(run).toThrow();`, Errors: missingError(2, 13, 20, "toThrow")},
			{Code: `import { expect as check } from '@rstest/core';
check(run).toThrowError();`, Errors: missingError(2, 12, 24, "toThrowError")},
			{Code: `const { expect } = require('@rstest/core');
expect(run).toThrow();`, Errors: missingError(2, 13, 20, "toThrow")},
			{Code: `import * as core from '@rstest/core';
core.expect(run).toThrowError();`, Errors: missingError(2, 18, 30, "toThrowError")},
			{Code: `const core = require('@rstest/core');
core.expect(run).toThrow();`, Errors: missingError(2, 18, 25, "toThrow")},
			{Code: `import.meta.rstest.expect(run).toThrow();`, Errors: missingError(1, 32, 39, "toThrow")},
			{Code: `const { expect } = import.meta.rstest;
expect(run).toThrowError();`, Errors: missingError(2, 13, 25, "toThrowError")},
			{Code: `const api = import.meta.rstest;
api.expect(run).toThrow();`, Errors: missingError(2, 17, 24, "toThrow")},
			{Code: `import { expect } from '@rstest/playwright';
expect(run).toThrow();`, Errors: missingError(2, 13, 20, "toThrow")},
			{Code: `import { expect } from 'rstack/test';
expect(run).toThrowError();`, Errors: missingError(2, 13, 25, "toThrowError")},
			{Code: `test('x', ({ expect }) => expect(run).toThrow());`, Errors: missingError(1, 39, 46, "toThrow")},
			{Code: `test('x', ({ expect: check }) => check(run).toThrowError());`, Errors: missingError(1, 45, 57, "toThrowError")},
			{Code: `test('x', ctx => ctx.expect(run).toThrow());`, Errors: missingError(1, 34, 41, "toThrow")},

			// ---- First call-style matcher only ----
			{Code: `expect(run).toThrow().throw();`, Errors: missingError(1, 13, 20, "toThrow")},

			// ---- Diagnostic range: multiline and UTF-16 prefix ----
			{Code: `expect(run)
  .toThrow();`, Errors: missingError(2, 4, 11, "toThrow")},
			{Code: `const label = '😀'; expect(run).toThrow();`, Errors: missingError(1, 33, 40, "toThrow")},

			// ---- Real-user: nested assertion from Rstest's expect e2e suite ----
			{Code: `expect(() => expect(mock).returned(undefined)).toThrow();`, Errors: missingError(1, 48, 55, "toThrow")},
			// ---- Real-user: async rejection assertion used in CLI tests ----
			{Code: `await expect(run()).rejects.toThrow();`, Errors: missingError(1, 29, 36, "toThrow")},
		},
	)
}

func missingError(line, column, endColumn int, matcher string) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "addErrorMessage",
		Message:   "Add an error message to " + matcher + "()",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}}
}

func TestRequireToThrowMessageSchema(t *testing.T) {
	for _, options := range [][]any{nil, {}} {
		if err := RequireToThrowMessageRule.Schema.Validate(options); err != nil {
			t.Errorf("expected options %#v to pass schema validation: %v", options, err)
		}
	}
	if err := RequireToThrowMessageRule.Schema.Validate([]any{map[string]any{}}); err == nil {
		t.Error("expected a non-empty options array to fail schema validation")
	}
}
