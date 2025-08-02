package no_unused_expressions

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnusedExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnusedExpressionsRule, []rule_tester.ValidTestCase{
		{Code: `foo()`},
		{Code: `new Foo()`},
		{Code: `foo++`},
		{Code: `delete foo.bar`},
		{Code: `await foo()`},
		{Code: `yield foo`},
		{Code: `"use strict"`},
		{Code: `'use strict'`},
		{Code: `function foo() { "use strict"; return 1; }`},
		{Code: `foo && bar()`},
		{Code: `foo || bar()`},
		{Code: `foo ? bar() : baz()`},
		{Code: `foo?.bar()`},
		{Code: `import('./foo')`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `foo`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression"},
			},
		},
		{
			Code: `foo.bar`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression"},
			},
		},
		{
			Code: `foo[bar]`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression"},
			},
		},
		{
			Code: `1 + 2`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression"},
			},
		},
		{
			Code: `"hello"`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unusedExpression"},
			},
		},
	})
}