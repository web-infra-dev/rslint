package require_hook_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/require_hook"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireHookExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_hook.RequireHookRule,
		[]rule_tester.ValidTestCase{
			{Code: `setup?.();`},
			{Code: `foo?.bar();`},
			{Code: `import('x');`},
			{Code: `export let value = setup();`},
			{Code: `jest['mock']('../api');`},
			{Code: `describe('a test', () => setup());`},
			{Code: `describe.only('suite', () => {
  beforeEach(() => setup());
});`},
			{
				Code: `helper.setup();`,
				Options: []interface{}{
					map[string]interface{}{
						"allowedFunctionCalls": []interface{}{"helper.setup"},
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `function register() {
  describe('suite', () => {
    setup();
  });
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 3, Column: 5},
				},
			},
			{
				Code: `foo(describe('suite', () => {
  setup();
}));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `if (condition) {
 describe('suite', () => {
  setup();
 });
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 3, Column: 3},
				},
			},
			{
				Code: `const suite = describe('suite', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 2, Column: 3},
				},
			},
			{
				Code: `(setup());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `await using value = setup();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `(foo?.bar)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `helper.setup();`,
				Options: []interface{}{
					map[string]interface{}{
						"allowedFunctionCalls": []interface{}{"setup"},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
			{
				Code: `new NodeExtensionTester()
  .shouldMatch()
  .runTests();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useHook", Line: 1, Column: 1},
				},
			},
		},
	)
}
