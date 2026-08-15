package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestInitDeclarationsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/init-declarations.js (eslint v10.8.1) 1:1 — both
// the default-parser suite (ruleTester) and the @typescript-eslint/parser
// suite (ruleTesterTypeScript). rslint parses every file with tsgo, so both
// suites run as one flat Go suite; TS-syntax cases from the second upstream
// suite are marked accordingly. Position assertions cover line/column for
// every invalid case. rslint-specific lock-in cases live in
// init_declarations_extras_test.go.
func TestInitDeclarationsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&InitDeclarationsRule,
		[]rule_tester.ValidTestCase{
			// ---- ruleTester (default parser) ----
			{Code: `var foo = null;`},
			{Code: `foo = true;`},
			{Code: `var foo = 1, bar = false, baz = {};`},
			{Code: `function foo() { var foo = 0; var bar = []; }`},
			{Code: `var fn = function() {};`},
			{Code: `var foo = bar = 2;`},
			{Code: `for (var i = 0; i < 1; i++) {}`},
			{Code: `for (var foo in []) {}`},
			{Code: `for (var foo of []) {}`},
			{Code: `let a = true;`, Options: []any{"always"}},
			{Code: `const a = {};`, Options: []any{"always"}},
			{Code: `using a = foo();`, Options: []any{"always"}},
			{Code: `await using a = foo();`, Options: []any{"always"}},
			{Code: `function foo() { let a = 1, b = false; if (a) { let c = 3, d = null; } }`, Options: []any{"always"}},
			{Code: `function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }`, Options: []any{"always"}},
			{Code: `function foo() { let a = 1; const b = false; var c = true; }`, Options: []any{"always"}},
			{Code: `var foo;`, Options: []any{"never"}},
			{Code: `var foo, bar, baz;`, Options: []any{"never"}},
			{Code: `function foo() { var foo; var bar; }`, Options: []any{"never"}},
			{Code: `let a;`, Options: []any{"never"}},
			{Code: `const a = 1;`, Options: []any{"never"}},
			{Code: `using a = foo();`, Options: []any{"never"}},
			{Code: `await using a = foo();`, Options: []any{"never"}},
			{Code: `function foo() { let a, b; if (a) { let c, d; } }`, Options: []any{"never"}},
			{Code: `function foo() { const a = 1, b = true; if (a) { const c = 3, d = null; } }`, Options: []any{"never"}},
			{Code: `function foo() { let a; const b = false; var c; }`, Options: []any{"never"}},
			{Code: `for(var i = 0; i < 1; i++){}`, Options: []any{"never", map[string]any{"ignoreForLoopInit": true}}},
			{Code: `for (var foo in []) {}`, Options: []any{"never", map[string]any{"ignoreForLoopInit": true}}},
			{Code: `for (var foo of []) {}`, Options: []any{"never", map[string]any{"ignoreForLoopInit": true}}},

			// ---- ruleTesterTypeScript (@typescript-eslint/parser) ----
			{Code: `declare const foo: number;`, Options: []any{"always"}},
			{Code: `declare const foo: number;`, Options: []any{"never"}},
			{Code: `
	  declare namespace myLib {
		let numberOfGreetings: number;
	  }
			`, Options: []any{"always"}},
			{Code: `
	  declare namespace myLib {
		let numberOfGreetings: number;
	  }
			`, Options: []any{"never"}},
			{Code: `
	  declare namespace myLib {
		let valueInside: number;
	  }
		let valueOutside: number;
			`, Options: []any{"never"}},
			{Code: `
	  interface GreetingSettings {
		greeting: string;
		duration?: number;
		color?: string;
	  }
			`},
			{Code: `
	  interface GreetingSettings {
		greeting: string;
		duration?: number;
		color?: string;
	  }
			`, Options: []any{"never"}},
			{Code: `type GreetingLike = string | (() => string) | Greeter;`},
			{Code: `type GreetingLike = string | (() => string) | Greeter;`, Options: []any{"never"}},
			{Code: `
	  function foo() {
		var bar: string;
	  }
			`, Options: []any{"never"}},
			{Code: `var bar: string;`, Options: []any{"never"}},
			{Code: `
	  var bar: string = function (): string {
		return 'string';
	  };
			`, Options: []any{"always"}},
			{Code: `
	  var bar: string = function (arg1: string): string {
		return 'string';
	  };
			`, Options: []any{"always"}},
			{Code: `function foo(arg1: string = 'string'): void {}`, Options: []any{"never"}},
			{Code: `const foo: string = 'hello';`, Options: []any{"never"}},
			{Code: `
	  const class1 = class NAME {
		constructor() {
		  var name1: string = 'hello';
		}
	  };
			`},
			{Code: `
	  const class1 = class NAME {
		static pi: number = 3.14;
	  };
			`},
			{Code: `
	  const class1 = class NAME {
		static pi: number = 3.14;
	  };
			`, Options: []any{"never"}},
			{Code: `
	  interface IEmployee {
		empCode: number;
		empName: string;
		getSalary: (number) => number; // arrow function
		getManagerName(number): string;
	  }
			`},
			{Code: `
	  interface IEmployee {
		empCode: number;
		empName: string;
		getSalary: (number) => number; // arrow function
		getManagerName(number): string;
	  }
			`, Options: []any{"never"}},
			{Code: `const foo: number = 'asd';`, Options: []any{"always"}},
			{Code: `const foo: number;`, Options: []any{"never"}},
			{Code: `
	  namespace myLib {
		let numberOfGreetings: number;
	  }
			`, Options: []any{"never"}},
			{Code: `
	  namespace myLib {
		let numberOfGreetings: number = 2;
	  }
			`, Options: []any{"always"}},
			{Code: `
	  declare namespace myLib1 {
		const foo: number;
		namespace myLib2 {
		  let bar: string;
		  namespace myLib3 {
			let baz: object;
		  }
		}
	  }
			`, Options: []any{"always"}},
			{Code: `
	  declare namespace myLib1 {
		const foo: number;
		namespace myLib2 {
		  let bar: string;
		  namespace myLib3 {
			let baz: object;
		  }
		}
	  }
			`, Options: []any{"never"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ruleTester (default parser) ----
			{
				Code:    `var foo;`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8}},
			},
			{
				Code:    `for (var a in []) var foo;`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}},
			},
			{
				Code:    `var foo, bar = false, baz;`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 8},
					{MessageId: "initialized", Line: 1, Column: 23, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `function foo() { var foo = 0; var bar; }`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 35, EndLine: 1, EndColumn: 38}},
			},
			{
				Code:    `function foo() { var foo; var bar = foo; }`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 25}},
			},
			{
				Code:    `let a;`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 6}},
			},
			{
				Code:    `function foo() { let a = 1, b; if (a) { let c = 3, d = null; } }`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 29, EndLine: 1, EndColumn: 30}},
			},
			{
				Code:    `function foo() { let a; const b = false; var c; }`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
					{MessageId: "initialized", Line: 1, Column: 46, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:    `var foo = bar = 2;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 18}},
			},
			{
				Code:    `var foo = true;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `var foo, bar = 5, baz = 3;`,
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
					{MessageId: "notInitialized", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    `function foo() { var foo; var bar = foo; }`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 31, EndLine: 1, EndColumn: 40}},
			},
			{
				Code:    `let a = 1;`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 10}},
			},
			{
				Code:    `function foo() { let a = 'foo', b; if (a) { let c, d; } }`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 22, EndLine: 1, EndColumn: 31}},
			},
			{
				Code:    `function foo() { let a; const b = false; var c = 1; }`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 46, EndLine: 1, EndColumn: 51}},
			},
			{
				Code:    `for(var i = 0; i < 1; i++){}`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:    `for (var foo in []) {}`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    `for (var foo of []) {}`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}},
			},

			// ---- ruleTesterTypeScript (@typescript-eslint/parser) ----
			{
				Code:    `let arr: string[] = ['arr', 'ar'];`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 34}},
			},
			{
				Code:    `let arr: string = function () {};`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 33}},
			},
			{
				Code: `
	  const class1 = class NAME {
		constructor() {
		  var name1: string = 'hello';
		}
	  };
			`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 4, Column: 9, EndLine: 4, EndColumn: 32}},
			},
			{
				Code:    `let arr: string;`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}},
			},
			{
				Code: `
	  namespace myLib {
		let numberOfGreetings: number;
	  }
			`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 32}},
			},
			{
				Code: `
	  namespace myLib {
		let numberOfGreetings: number = 2;
	  }
			`,
				Options: []any{"never"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "notInitialized", Line: 3, Column: 7, EndLine: 3, EndColumn: 36}},
			},
			{
				// Not "declare namespace" (unlike the valid case above) — a plain
				// namespace does not exempt its members.
				Code: `
		namespace myLib1 {
		  const foo: number;
			namespace myLib2 {
			  let bar: string;
			  namespace myLib3 {
				let baz: object;
			  }
		  }
		}
			`,
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "initialized", Line: 3, Column: 11, EndLine: 3, EndColumn: 22},
					{MessageId: "initialized", Line: 5, Column: 10, EndLine: 5, EndColumn: 21},
					{MessageId: "initialized", Line: 7, Column: 9, EndLine: 7, EndColumn: 20},
				},
			},
			{
				Code: `
	  declare namespace myLib {
		let valueInside: number;
	  }
		let valueOutside: number;
			`,
				Options: []any{"always"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "initialized", Line: 5, Column: 7, EndLine: 5, EndColumn: 27}},
			},
		},
	)
}
