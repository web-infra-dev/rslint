package no_inferrable_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoInferrableTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoInferrableTypesRule, []rule_tester.ValidTestCase{
		{Code: "const a = 10n;"},
		{Code: "const a = -10n;"},
		{Code: "const a = BigInt(10);"},
		{Code: "const a = -BigInt(10);"},
		{Code: "const a = false;"},
		{Code: "const a = true;"},
		{Code: "const a = Boolean(null);"},
		{Code: "const a = !0;"},
		{Code: "const a = 10;"},
		{Code: "const a = +10;"},
		{Code: "const a = -10;"},
		{Code: "const a = Number('1');"},
		{Code: "const a = +Number('1');"},
		{Code: "const a = -Number('1');"},
		{Code: "const a = Infinity;"},
		{Code: "const a = +Infinity;"},
		{Code: "const a = -Infinity;"},
		{Code: "const a = NaN;"},
		{Code: "const a = +NaN;"},
		{Code: "const a = -NaN;"},
		{Code: "const a = null;"},
		{Code: "const a = /a/;"},
		{Code: "const a = RegExp('a');"},
		{Code: "const a = new RegExp('a');"},
		{Code: "const a = 'str';"},
		{Code: "const a = `str`;"},
		{Code: "const a = String(1);"},
		{Code: "const a = Symbol('a');"},
		{Code: "const a = undefined;"},
		{Code: "const a = void someValue;"},
		{Code: "const fn = (a = 5, b = true, c = 'foo') => {};"},
		{Code: "const fn = function (a = 5, b = true, c = 'foo') {};"},
		{Code: "function fn(a = 5, b = true, c = 'foo') {}"},
		{Code: "function fn(a: number, b: boolean, c: string) {}"},
		{Code: "class Foo { a = 5; b = true; c = 'foo'; }"},
		{Code: "class Foo { readonly a: number = 5; }"},
		{Code: "const a: any = 5;"},
		{Code: "const fn = function (a: any = 5, b: any = true, c: any = 'foo') {};"},
		{
			Code:    "const fn = (a: number = 5, b: boolean = true, c: string = 'foo') => {};",
			Options: map[string]interface{}{"ignoreParameters": true},
		},
		{
			Code:    "function fn(a: number = 5, b: boolean = true, c: string = 'foo') {}",
			Options: map[string]interface{}{"ignoreParameters": true},
		},
		{
			Code:    "const fn = function (a: number = 5, b: boolean = true, c: string = 'foo') {};",
			Options: map[string]interface{}{"ignoreParameters": true},
		},
		{
			Code:    "class Foo { a: number = 5; b: boolean = true; c: string = 'foo'; }",
			Options: map[string]interface{}{"ignoreProperties": true},
		},
		{
			Code: "class Foo { a?: number = 5; b?: boolean = true; c?: string = 'foo'; }",
		},
		{
			Code: "class Foo { constructor(public a = true) {} }",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "const a: bigint = 10n;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: bigint = -10n;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: bigint = BigInt(10);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: bigint = -BigInt(10);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: boolean = false;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: boolean = true;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: boolean = Boolean(null);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: boolean = !0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = 10;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = +10;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = -10;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = Number('1');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = +Number('1');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = -Number('1');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = Infinity;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = +Infinity;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = -Infinity;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = NaN;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = +NaN;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: number = -NaN;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: null = null;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: RegExp = /a/;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: RegExp = RegExp('a');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: RegExp = new RegExp('a');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: string = 'str';",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: string = `str`;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: string = String(1);",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: symbol = Symbol('a');",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: undefined = undefined;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code: "const a: undefined = void someValue;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 7},
			},
		},
		{
			Code:    "const fn = (a?: number = 5) => {};",
			Options: map[string]interface{}{"ignoreParameters": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 13},
			},
		},
		{
			Code:    "class A { a!: number = 1; }",
			Options: map[string]interface{}{"ignoreProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 11},
			},
		},
		{
			Code:    "const fn = (a: number = 5, b: boolean = true, c: string = 'foo') => {};",
			Options: map[string]interface{}{"ignoreParameters": false, "ignoreProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 13},
				{MessageId: "noInferrableType", Line: 1, Column: 28},
				{MessageId: "noInferrableType", Line: 1, Column: 47},
			},
		},
		{
			Code:    "class Foo { a: number = 5; b: boolean = true; c: string = 'foo'; }",
			Options: map[string]interface{}{"ignoreParameters": false, "ignoreProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 13},
				{MessageId: "noInferrableType", Line: 1, Column: 28},
				{MessageId: "noInferrableType", Line: 1, Column: 47},
			},
		},
		{
			Code:    "class Foo { constructor(public a: boolean = true) {} }",
			Options: map[string]interface{}{"ignoreParameters": false, "ignoreProperties": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noInferrableType", Line: 1, Column: 32},
			},
		},
	})
}
