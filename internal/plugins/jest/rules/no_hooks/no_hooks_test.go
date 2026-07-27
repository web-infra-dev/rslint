package no_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoHooksRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_hooks.NoHooksRule,
		[]rule_tester.ValidTestCase{
			{Code: "test(\"foo\")"},
			{Code: "describe(\"foo\", () => { it(\"bar\") })"},
			{Code: "test(\"foo\", () => { expect(subject.beforeEach()).toBe(true) })"},
			{
				Code: "afterEach(() => {}); afterAll(() => {});",
				Options: []interface{}{
					map[string]interface{}{"allow": []interface{}{"afterEach", "afterAll"}},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "beforeAll(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 1, Column: 1},
				},
			},
			{
				Code: "beforeEach(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 1, Column: 1},
				},
			},
			{
				Code: "afterAll(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 1, Column: 1},
				},
			},
			{
				Code: "afterEach(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `
			        import { 'afterEach' as afterEachTest } from '@jest/globals';

			        afterEachTest(() => {})
			    `,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 4, Column: 12},
				},
			},
			{
				Code: "beforeEach(() => {}); afterEach(() => { jest.resetModules() });",
				Options: []interface{}{
					map[string]interface{}{"allow": []interface{}{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `
			        import { beforeEach as afterEach, afterEach as beforeEach } from '@jest/globals';

			        afterEach(() => {});
			        beforeEach(() => { jest.resetModules() });
			    `,
				Options: []interface{}{
					map[string]interface{}{"allow": []interface{}{"afterEach"}},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedHook", Line: 4, Column: 12},
				},
			},
		},
	)
}

// TestNoHooksAllowSchema locks in the divergence from upstream's `meta.schema`,
// which constrains `allow` with `contains: ["beforeAll", ...]`. `contains` is a
// draft-6 keyword whose value must be a schema, so under the draft-4 dialect
// ESLint (and rslint) validate with, it is an unknown keyword and ignored
// entirely — leaving `allow` to accept any array at all. Every element must be
// one of the four hook names the rule can act on, which is what `items`/`enum`
// says, so a typo fails validation instead of silently never matching.
func TestNoHooksAllowSchema(t *testing.T) {
	valid := []any{map[string]any{"allow": []any{"beforeEach", "afterAll"}}}
	if err := no_hooks.NoHooksRule.Schema.Validate(valid); err != nil {
		t.Errorf("expected hook names to pass schema validation, got: %v", err)
	}
	invalid := []any{map[string]any{"allow": []any{"beforeeach"}}}
	if err := no_hooks.NoHooksRule.Schema.Validate(invalid); err == nil {
		t.Error("expected a non-hook name to fail schema validation")
	}
}
