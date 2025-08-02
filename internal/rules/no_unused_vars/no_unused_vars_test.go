package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnusedVarsRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: `const foo = 5; console.log(foo);`},
		{Code: `function foo() {} foo();`},
		{Code: `function foo(bar) { console.log(bar); }`},
		{Code: `try {} catch (e) { console.log(e); }`},
		{Code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
		{Code: `const _foo = 1;`, Options: map[string]interface{}{"varsIgnorePattern": "^_"}},
		{Code: `function foo(bar) {}`, Options: map[string]interface{}{"args": "none"}},
		{Code: `try {} catch (e) {}`, Options: map[string]interface{}{"caughtErrors": "none"}},
		{Code: `export const foo = 1;`},
		{Code: `import type { Foo } from "./foo"; const bar: Foo = {};`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `const foo = 5;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		{
			Code: `function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 10}},
		},
		{
			Code: `function foo(bar) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 14}},
		},
		{
			Code: `try {} catch (e) {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 15}},
		},
		{
			Code: `let foo = 5; foo = 10;`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 5}},
		},
		{
			Code: `const foo = 1; type Bar = typeof foo;`,
			Options: map[string]interface{}{"vars": "all"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedOnlyAsType", Line: 1, Column: 7}},
		},
		{
			Code: `const foo = 1;`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unusedVar", Line: 1, Column: 7}},
		},
		{
			Code: `const _foo = 1; console.log(_foo);`,
			Options: map[string]interface{}{"varsIgnorePattern": "^_", "reportUsedIgnorePattern": true},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedIgnoredVar", Line: 1, Column: 7}},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedVarsRule, validTestCases, invalidTestCases)
}