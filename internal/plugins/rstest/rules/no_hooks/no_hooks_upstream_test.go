// TestNoHooksUpstream migrates the full valid/invalid suite from upstream
// eslint-plugin-jest v29.16.1 src/rules/__tests__/no-hooks.test.ts 1:1.
// Position assertions cover full start/end ranges for every invalid case.
// rslint-specific lock-in cases live in the no_hooks_extras_test.go file.
package no_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoHooksUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_hooks.NoHooksRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid ----
			{Code: `test("foo")`},
			{Code: `describe("foo", () => { it("bar") })`},
			{Code: `test("foo", () => { expect(subject.beforeEach()).toBe(true) })`},
			{
				Code: `afterEach(() => {}); afterAll(() => {});`,
				Options: []any{
					map[string]any{"allow": []any{"afterEach", "afterAll"}},
				},
			},
			{
				Code:    `test("foo")`,
				Options: []any{map[string]any{}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid ----
			{
				Code: `beforeAll(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeAll' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code: `beforeEach(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `afterAll(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterAll' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code: `afterEach(() => {})`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 20,
				}},
			},
			{
				Code: `
import { 'afterEach' as afterEachTest } from '@rstest/core';

afterEachTest(() => {})
`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'afterEach' hook",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 24,
				}},
			},
			{
				Code: `beforeEach(() => {}); afterEach(() => { resetModules() });`,
				Options: []any{
					map[string]any{"allow": []any{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code: `
import { beforeEach as afterEach, afterEach as beforeEach } from '@rstest/core';

afterEach(() => {});
beforeEach(() => { resetModules() });
`,
				Options: []any{
					map[string]any{"allow": []any{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedHook",
					Message:   "Unexpected 'beforeEach' hook",
					Line:      4,
					Column:    1,
					EndLine:   4,
					EndColumn: 20,
				}},
			},
		},
	)
}
