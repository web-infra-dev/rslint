package require_hook_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/require_hook"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireHookUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_hook.RequireHookRule,
		[]rule_tester.ValidTestCase{
			{Code: `function helper() { setup(); }
test('x', () => {});`},
			{Code: `let x;
x = 1;
test('x', () => {});`},
			{Code: `new Helper();
test('x', () => {});`},
			{Code: `describe('empty', () => {});`},
			{Code: `var x;
test('x', () => {});`},
			{Code: `let x = (null);
test('x', () => {});`},
			{Code: `let x = (undefined);
test('x', () => {});`},
			{Code: `type T = number;
interface I { x: number }
test('x', () => {});`},
			{Code: `jest.anythingCustom();`},
			{Code: `const utils = require('./utils');
test('x', () => {});`},
			{Code: `describe('title only');`},
			{Code: `describe('title', 'not a function');`},
			{
				Code: `enableAutoDestroy(afterEach);
test('x', () => {});`,
				Options: []interface{}{
					map[string]interface{}{
						"allowedFunctionCalls": []interface{}{"enableAutoDestroy"},
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe('suite', function () {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 2, Column: 3},
				},
			},
			{
				Code: `var value = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `foo.bar();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `describe.only('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 2, Column: 3},
				},
			},
			{
				Code: `test('x', () => {
  setup();
});
setup();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 4, Column: 1},
				},
			},
		},
	)
}
