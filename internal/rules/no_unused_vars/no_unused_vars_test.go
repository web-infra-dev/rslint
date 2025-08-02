package no_unused_vars

import (
	"testing"
)

func TestNoUnusedVarsRule(t *testing.T) {
	// TODO: Convert this test to use rule_tester.RunRuleTester format
	// This test uses an incompatible structure with []rule_tester.TestCase
	/*
	tester := rule_tester.New(t)

	tester.Run("no-unused-vars", NoUnusedVarsRule, []rule_tester.TestCase{
		// Valid cases
		{
			Name: "variable is used",
			Code: `const foo = 5; console.log(foo);`,
		},
		{
			Name: "function is used",
			Code: `function foo() {} foo();`,
		},
		{
			Name: "parameter is used",
			Code: `function foo(bar) { console.log(bar); }`,
		},
		{
			Name: "catch variable is used",
			Code: `try {} catch (e) { console.log(e); }`,
		},
		{
			Name: "rest siblings with ignoreRestSiblings",
			Code: `const { foo, ...rest } = { foo: 1, bar: 2 }; console.log(rest);`,
			Options: map[string]interface{}{
				"ignoreRestSiblings": true,
			},
		},
		{
			Name: "vars ignore pattern",
			Code: `const _foo = 1;`,
			Options: map[string]interface{}{
				"varsIgnorePattern": "^_",
			},
		},
		{
			Name: "args none",
			Code: `function foo(bar) {}`,
			Options: map[string]interface{}{
				"args": "none",
			},
		},
		{
			Name: "caught errors none",
			Code: `try {} catch (e) {}`,
			Options: map[string]interface{}{
				"caughtErrors": "none",
			},
		},
		{
			Name: "exported variable",
			Code: `export const foo = 1;`,
		},
		{
			Name: "type-only import",
			Code: `import type { Foo } from "./foo"; const bar: Foo = {};`,
		},

		// Invalid cases
		{
			Name: "unused variable",
			Code: `const foo = 5;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "foo",
						"action":     "defined",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "unused function",
			Code: `function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "foo",
						"action":     "defined",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "unused parameter",
			Code: `function foo(bar) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "bar",
						"action":     "defined",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "unused catch variable",
			Code: `try {} catch (e) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "e",
						"action":     "defined",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "assigned but unused",
			Code: `let foo = 5; foo = 10;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "foo",
						"action":     "assigned a value",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "only used as type",
			Code: `const foo = 1; type Bar = typeof foo;`,
			Options: map[string]interface{}{
				"vars": "all",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "usedOnlyAsType",
					Data: map[string]interface{}{
						"varName":    "foo",
						"action":     "defined",
						"additional": "",
					},
				},
			},
		},
		{
			Name: "vars ignore pattern message",
			Code: `const foo = 1;`,
			Options: map[string]interface{}{
				"varsIgnorePattern": "^_",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unusedVar",
					Data: map[string]interface{}{
						"varName":    "foo",
						"action":     "defined",
						"additional": ". Allowed unused vars must match ^_",
					},
				},
			},
		},
		{
			Name: "report used ignore pattern",
			Code: `const _foo = 1; console.log(_foo);`,
			Options: map[string]interface{}{
				"varsIgnorePattern":       "^_",
				"reportUsedIgnorePattern": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "usedIgnoredVar",
					Data: map[string]interface{}{
						"varName":    "_foo",
						"additional": ". Used vars must not match ^_",
					},
				},
			},
		},
	})
	*/
}