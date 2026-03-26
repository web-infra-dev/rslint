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
