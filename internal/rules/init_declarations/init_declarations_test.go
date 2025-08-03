package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestInitDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &InitDeclarationsRule, []rule_tester.ValidTestCase{
		// Basic valid cases - always mode (default)
		{Code: "var foo = null;"},
		{Code: "foo = true;"},
		{Code: "var foo = 1, bar = false, baz = {};"},
		{Code: "function foo() { var foo = 0; var bar = []; }"},
		{Code: "var fn = function () {};"},
		{Code: "var foo = (bar = 2);"},
		{Code: "for (var i = 0; i < 1; i++) {}"},
		{Code: "for (var foo in []) {}"},
		{Code: "for (var foo of []) {}"},
		{Code: "let a = true;"},
		{Code: "const a = {};"},
		{Code: "function foo() { let a = 1, b = false; if (a) { let c = 3, d = null; } }"},
		{Code: "function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }"},
		{Code: "function foo() { let a = 1; const b = false; var c = true; }"},

		// Basic valid cases - never mode
		{Code: "var foo;", Options: []interface{}{"never"}},
		{Code: "var foo, bar, baz;", Options: []interface{}{"never"}},
		{Code: "function foo() { var foo; var bar; }", Options: []interface{}{"never"}},
		{Code: "let a;", Options: []interface{}{"never"}},
		{Code: "const a = 1;", Options: []interface{}{"never"}}, // const always requires init
		{Code: "function foo() { let a, b; if (a) { let c, d; } }", Options: []interface{}{"never"}},
		{Code: "function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }", Options: []interface{}{"never"}},
		{Code: "function foo() { let a; const b = false; var c; }", Options: []interface{}{"never"}},

		// ignoreForLoopInit option
		{Code: "for (var i = 0; i < 1; i++) {}", Options: []interface{}{"never", map[string]interface{}{"ignoreForLoopInit": true}}},
		{Code: "for (var foo in []) {}", Options: []interface{}{"never", map[string]interface{}{"ignoreForLoopInit": true}}},
		{Code: "for (var foo of []) {}", Options: []interface{}{"never", map[string]interface{}{"ignoreForLoopInit": true}}},

		// TypeScript-specific valid cases
		{Code: "declare const foo: number;", Options: []interface{}{"always"}},
		{Code: "declare const foo: number;", Options: []interface{}{"never"}},
		{Code: "declare namespace myLib { let numberOfGreetings: number; }", Options: []interface{}{"always"}},
		{Code: "declare namespace myLib { let numberOfGreetings: number; }", Options: []interface{}{"never"}},
		{Code: "interface GreetingSettings { greeting: string; duration?: number; color?: string; }", Options: []interface{}{"always"}},
		{Code: "interface GreetingSettings { greeting: string; duration?: number; color?: string; }", Options: []interface{}{"never"}},
		{Code: "type GreetingLike = string | (() => string) | Greeter;", Options: []interface{}{"always"}},
		{Code: "type GreetingLike = string | (() => string) | Greeter;", Options: []interface{}{"never"}},
		{Code: "function foo() { var bar: string; }", Options: []interface{}{"never"}},
		{Code: "var bar: string;", Options: []interface{}{"never"}},
		{Code: "var bar: string = function (): string { return 'string'; };", Options: []interface{}{"always"}},
		{Code: "var bar: string = function (arg1: string): string { return 'string'; };", Options: []interface{}{"always"}},
		{Code: "function foo(arg1: string = 'string'): void {}", Options: []interface{}{"never"}},
		{Code: "const foo: string = 'hello';", Options: []interface{}{"never"}},
		{Code: "const foo: number = 123;", Options: []interface{}{"always"}},
		{Code: "const foo: number;", Options: []interface{}{"never"}}, // const must be initialized
		{Code: "namespace myLib { let numberOfGreetings: number; }", Options: []interface{}{"never"}},
		{Code: "namespace myLib { let numberOfGreetings: number = 2; }", Options: []interface{}{"always"}},
		{Code: "declare namespace myLib1 { const foo: number; namespace myLib2 { let bar: string; namespace myLib3 { let baz: object; } } }", Options: []interface{}{"always"}},
		{Code: "declare namespace myLib1 { const foo: number; namespace myLib2 { let bar: string; namespace myLib3 { let baz: object; } } }", Options: []interface{}{"never"}},
	}, []rule_tester.InvalidTestCase{
		// Basic invalid cases - always mode
		{
			Code:    "var foo;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
			},
		},
		{
			Code:    "for (var a in []) var foo;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 26},
			},
		},
		{
			Code:    "var foo, bar = false, baz;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
				{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 26},
			},
		},
		{
			Code:    "function foo() { var foo = 0; var bar; }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 35, EndLine: 1, EndColumn: 38},
			},
		},
		{
			Code:    "function foo() { var foo; var bar = foo; }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
			},
		},
		{
			Code:    "let a;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
			},
		},
		{
			Code:    "function foo() { let a = 1, b; if (a) { let c = 3, d = null; } }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
			},
		},
		{
			Code:    "function foo() { let a; const b = false; var c; }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				{MessageId: "initialized", Line: 1, Column: 46, EndLine: 1, EndColumn: 47},
			},
		},

		// Basic invalid cases - never mode
		{
			Code:    "var foo = (bar = 2);",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 20},
			},
		},
		{
			Code:    "var foo = true;",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 15},
			},
		},
		{
			Code:    "var foo, bar = 5, baz = 3;",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				{MessageId: "notInitialized", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
			},
		},
		{
			Code:    "function foo() { var foo; var bar = foo; }",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 31, EndLine: 1, EndColumn: 40},
			},
		},
		{
			Code:    "let a = 1;",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 10},
			},
		},
		{
			Code:    "function foo() { let a = 'foo', b; if (a) { let c, d; } }",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
			},
		},
		{
			Code:    "function foo() { let a; const b = false; var c = 1; }",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 46, EndLine: 1, EndColumn: 51},
			},
		},

		// For loop init without ignoreForLoopInit
		{
			Code:    "for (var i = 0; i < 1; i++) {}",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
			},
		},
		{
			Code:    "for (var foo in []) {}",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code:    "for (var foo of []) {}",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
			},
		},
		{
			Code:    "function foo() { var bar; }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
			},
		},

		// TypeScript-specific invalid cases
		{
			Code:    "let arr: string[] = ['arr', 'ar'];",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 34},
			},
		},
		{
			Code:    "let arr: string = function () {};",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 33},
			},
		},
		{
			Code:    "let arr: string;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
			},
		},
		{
			Code:    "namespace myLib { let numberOfGreetings: number; }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 40},
			},
		},
		{
			Code:    "namespace myLib { let numberOfGreetings: number = 2; }",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "notInitialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 52},
			},
		},
		{
			Code:    "namespace myLib1 { const foo: number; namespace myLib2 { let bar: string; namespace myLib3 { let baz: object; } } }",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "initialized", Line: 1, Column: 26, EndLine: 1, EndColumn: 29},
				{MessageId: "initialized", Line: 1, Column: 62, EndLine: 1, EndColumn: 65},
				{MessageId: "initialized", Line: 1, Column: 98, EndLine: 1, EndColumn: 101},
			},
		},
	})
}
