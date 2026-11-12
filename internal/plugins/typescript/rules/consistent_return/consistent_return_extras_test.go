package consistent_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentReturnExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentReturnRule, []rule_tester.ValidTestCase{
		// Basic valid cases - consistent returns
		{Code: `function foo() { return; }`},
		{Code: `function foo() { return 1; }`},
		{Code: `const foo = () => true;`},
		{Code: `const foo = () => { return true; };`},
		{Code: `function foo() { if (true) return 1; else return 2; }`},

		// Void functions can have empty returns
		{Code: `function foo(): void { return; }`},
		{Code: `function foo(): void { if (true) return; }`},
		{Code: `function foo(): void { if (true) return; else return; }`},
		{Code: `const foo = (): void => { return; };`},
		{Code: `const foo = (): void => { if (true) return; };`},

		// Void functions can call other void functions
		{Code: `
			function bar(): void { return; }
			function foo(): void { return bar(); }
		`},
		{Code: `
			const bar = (): void => {};
			const foo = (): void => { return bar(); };
		`},

		// Async functions returning Promise<void>
		{Code: `async function foo(): Promise<void> { return; }`},
		{Code: `async function foo(): Promise<void> { if (true) return; }`},
		{Code: `const foo = async (): Promise<void> => { return; };`},
		{Code: `async function foo(): Promise<void> { return Promise.resolve(); }`},

		// Functions with undefined in return type union
		{Code: `function foo(): number | undefined { if (true) return 1; else return undefined; }`},
		{Code: `function foo(): void | undefined { if (true) return; else return undefined; }`},

		// treatUndefinedAsUnspecified option
		{
			Code:    `function foo() { if (true) return undefined; else return; }`,
			Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}},
		},
		{
			Code:    `function foo() { return undefined; }`,
			Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}},
		},
		{
			Code:    `const foo = () => { if (true) return undefined; else return; };`,
			Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}},
		},

		// Nested functions
		{Code: `
			function foo() {
				function bar() { return 1; }
				return 2;
			}
		`},
		{Code: `
			function foo() {
				const bar = () => 1;
				return;
			}
		`},

		// Class methods
		{Code: `
			class Foo {
				bar(): void { return; }
			}
		`},
		{Code: `
			class Foo {
				bar() { return 1; }
			}
		`},

		// Overload signatures
		{Code: `
			function foo(x: number): number;
			function foo(x: string): string;
			function foo(x: any): any {
				return x;
			}
		`},

		// Functions that always throw
		{Code: `function foo() { throw new Error('error'); }`},
	}, []rule_tester.InvalidTestCase{
		// Upstream checks this function's implicit return, even with a literal
		// true condition. The original Rslint extras incorrectly marked it valid.
		{
			Code: `function foo() { if (true) { return 1; } }`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "missingReturn",
				Message:   "Expected to return a value at the end of function 'foo'.",
				Line:      1, Column: 10, EndLine: 1, EndColumn: 13,
			}},
		},
		// Basic inconsistent returns
		{
			Code: `function foo() { if (true) return 1; return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},
		{
			Code: `function foo() { if (true) return; return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue"},
			},
		},
		{
			Code: `const foo = () => { if (true) return 1; return; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Functions typed as any mixing returns
		{
			Code: `function foo(): any { if (true) return 1; return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Nested function breaking parent function's return contract
		{
			Code: `
				function foo() {
					function bar() { if (true) return 1; return; }
					return 2;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Async function with Promise<void> having incomplete returns
		{
			Code: `async function foo(): Promise<string> { if (true) return 'test'; return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Async function mixing value and empty returns
		{
			Code: `async function foo() { if (true) return Promise.resolve(1); return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Class methods with inconsistent returns
		{
			Code: `
				class Foo {
					bar() { if (true) return 1; return; }
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},

		// Multiple code paths with inconsistent returns
		{
			Code: `
				function foo(x: number) {
					if (x > 0) return 1;
					else if (x < 0) return;
					return 2;
				}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue"},
			},
		},
	})
}

// Additional contracts verified independently with typescript-eslint v8.67.0
// and ESLint v9.39.4; messages and ranges come from upstream, not Rslint.
func TestConsistentReturnTypeAwareExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentReturnRule, []rule_tester.ValidTestCase{
		// void union, value first.
		{
			Code: `function f(b: boolean): number | void { if (b) return 1; return; }`,
		},
		// void union, bare return first.
		{
			Code: `function f(b: boolean): number | void { if (b) return; return 1; }`,
		},
		// nested Promise with a void union.
		{
			Code: `async function f(b: boolean): Promise<Promise<number | void>> { if (b) return; return 1; }`,
		},
		// generic thenable whose first type argument is void.
		{
			Code: `interface Box<T> { then(cb: (value: T) => unknown): unknown; } async function f(b: boolean): Box<void> { if (b) return; return undefined; }`,
		},
		// any overload returning void.
		{
			Code: `function f(): number; function f(b: boolean): void; function f(b?: boolean): any { if (b) return; return 1; }`,
		},
		// undefined variable with the option enabled.
		{
			Code:    `declare const u: undefined; function f(b: boolean): undefined { if (b) return u; return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
		},
		// parenthesized undefined call with the option enabled.
		{
			Code:    `declare function u(): undefined; function f(b: boolean) { if (b) return (u()); return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
		},
		// parenthesized void syntax with the option enabled.
		{
			Code:    `function f(b: boolean) { if (b) return (void 0); return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
		},
		// nested function inside a void function.
		{
			Code: `function f(b: boolean): void { function inner() { return 1; } if (b) return inner(); return; }`,
		},
		// ES5 constructor implicit return.
		{
			Code: `function F(b: boolean) { if (b) return 1; }`,
		},
		// unreachable end after throw.
		{
			Code: `function f(b: boolean) { if (b) return 1; throw new Error(); }`,
		},
		// void expression type is not exactly undefined.
		{
			Code:    `declare const v: void; function f(b: boolean) { if (b) return v; return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
		},
		// unreachable end after infinite loop.
		{
			Code: `function f(b: boolean) { if (b) return 1; while (true) {} }`,
		},
		// try/catch end reachability.
		{
			Code: `function f(b: boolean) { if (b) return 1; try { return 2; } catch {} }`,
		},
		// shadowed undefined retains the core syntactic behavior.
		{
			Code:    `function f(b: boolean) { var undefined = 1; if (b) return (undefined); return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
		},
		// async void/undefined union.
		{
			Code: `async function f(b: boolean): Promise<void | undefined> { if (b) return undefined; return; }`,
		},
		// multiple ignored bare returns do not set the expected direction.
		{
			Code: `async function f(b: boolean): Promise<void> { if (b) return; return; return Promise.resolve(); }`,
		},
	}, []rule_tester.InvalidTestCase{
		// undefined calls are specified by default.
		{
			Code: `declare function u(): undefined; function f(b: boolean): undefined { if (b) return u(); return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 89, EndLine: 1, EndColumn: 96},
			},
		},
		// undefined return, bare return first.
		{
			Code: `function f(b: boolean): undefined { if (b) return; return undefined; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 52, EndLine: 1, EndColumn: 69},
			},
		},
		// async Promise<undefined> is not Promise<void>.
		{
			Code: `async function f(b: boolean): Promise<undefined> { if (b) return; return undefined; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Async function 'f' expected no return value.", Line: 1, Column: 67, EndLine: 1, EndColumn: 84},
			},
		},
		// synchronous Promise<void> is not exempt.
		{
			Code: `function f(b: boolean): Promise<void> { if (b) return Promise.resolve(); return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 74, EndLine: 1, EndColumn: 81},
			},
		},
		// union of Promises is not a type reference.
		{
			Code: `async function f(b: boolean): Promise<void> | Promise<number> { if (b) return; return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Async function 'f' expected no return value.", Line: 1, Column: 80, EndLine: 1, EndColumn: 89},
			},
		},
		// void annotation does not suppress an implicit return report.
		{
			Code: `function f(b: boolean): void { if (b) return void 0; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of function 'f'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// async void does not suppress an implicit return report.
		{
			Code: `async function f(b: boolean): Promise<void> { if (b) return void 0; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of async function 'f'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
			},
		},
		// union containing undefined remains specified.
		{
			Code:    `declare const u: number | undefined; function f(b: boolean) { if (b) return; return u; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 78, EndLine: 1, EndColumn: 87},
			},
		},
		// each later bare return conflicts.
		{
			Code: `function f(b: boolean) { if (b) return 1; return; return; return 2; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 51, EndLine: 1, EndColumn: 58},
			},
		},
		// each later valued return conflicts.
		{
			Code: `function f(b: boolean) { if (b) return; return 1; return 2; return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 41, EndLine: 1, EndColumn: 50},
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 51, EndLine: 1, EndColumn: 60},
			},
		},
		// nested function keeps its own return contract.
		{
			Code: `function f(b: boolean): void { function inner(b: boolean): undefined { if (b) return; return undefined; } if (b) return inner(b); return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'inner' expected no return value.", Line: 1, Column: 87, EndLine: 1, EndColumn: 104},
			},
		},
		// reachable implicit return.
		{
			Code: `function f(b: boolean) { if (b) return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of function 'f'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
			},
		},
		// arrow implicit return range.
		{
			Code: `const f = (b: boolean) => { if (b) return 1; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of arrow function.", Line: 1, Column: 24, EndLine: 1, EndColumn: 26},
			},
		},
		// object method implicit return range.
		{
			Code: `const o = { m(b: boolean) { if (b) return 1; } };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of method 'm'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
			},
		},
		// class method implicit return range.
		{
			Code: `class C { m(b: boolean) { if (b) return 1; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of method 'm'.", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
			},
		},
		// async arrow message.
		{
			Code: `const f = async (b: boolean): Promise<string> => { if (b) return; return 'x'; };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Async arrow function expected no return value.", Line: 1, Column: 67, EndLine: 1, EndColumn: 78},
			},
		},
		// async method message.
		{
			Code: `class C { async m(b: boolean): Promise<string> { if (b) return; return 'x'; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Async method 'm' expected no return value.", Line: 1, Column: 65, EndLine: 1, EndColumn: 76},
			},
		},
		// getter return classification.
		{
			Code: `const o = { get x(): undefined { if (b) return; return undefined; } };`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Getter 'x' expected no return value.", Line: 1, Column: 49, EndLine: 1, EndColumn: 66},
			},
		},
		// setter owns its returns.
		{
			Code: `class C { set x(v: number) { if (v) return; return undefined; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Setter 'x' expected no return value.", Line: 1, Column: 45, EndLine: 1, EndColumn: 62},
			},
		},
		// constructor owns its returns.
		{
			Code: `class C { constructor(b: boolean) { if (b) return; return {}; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Constructor expected no return value.", Line: 1, Column: 52, EndLine: 1, EndColumn: 62},
			},
		},
		// static constructor is an ordinary method.
		{
			Code: `class C { static constructor(b: boolean) { if (b) return 1; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of static method 'constructor'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 29},
			},
		},
		// type parameter constrained to void is not intrinsic void.
		{
			Code: `function f<T extends void>(b: boolean): T { if (b) return; return undefined as T; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 60, EndLine: 1, EndColumn: 82},
			},
		},
		// unknown return type.
		{
			Code: `function f(b: boolean): unknown { if (b) return 1; return; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 52, EndLine: 1, EndColumn: 59},
			},
		},
		// any return type.
		{
			Code: `function f(b: boolean): any { if (b) return; return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 46, EndLine: 1, EndColumn: 55},
			},
		},
		// never return type.
		{
			Code: `function f(b: boolean): never { if (b) return; return 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedReturnValue", Message: "Function 'f' expected no return value.", Line: 1, Column: 48, EndLine: 1, EndColumn: 57},
			},
		},
		// nested and enclosing diagnostics sort by source position.
		{
			Code: `function f(b: boolean) { const inner = () => { if (b) return; return 1; }; if (b) return 2; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturn", Message: "Expected to return a value at the end of function 'f'.", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				{MessageId: "unexpectedReturnValue", Message: "Arrow function expected no return value.", Line: 1, Column: 63, EndLine: 1, EndColumn: 72},
			},
		},
		// finally can conflict with the first return.
		{
			Code: `function f(b: boolean) { if (b) return 1; try {} finally { return; } }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 60, EndLine: 1, EndColumn: 67},
			},
		},
		// any cast of undefined remains specified.
		{
			Code:    `function f(b: boolean) { if (b) return undefined as any; return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 58, EndLine: 1, EndColumn: 65},
			},
		},
		// null is not undefined.
		{
			Code:    `declare const n: null; function f(b: boolean) { if (b) return n; return; }`,
			Options: map[string]any{"treatUndefinedAsUnspecified": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnValue", Message: "Function 'f' expected a return value.", Line: 1, Column: 66, EndLine: 1, EndColumn: 73},
			},
		},
	})
}
