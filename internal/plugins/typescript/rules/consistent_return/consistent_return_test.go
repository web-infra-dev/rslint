package consistent_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentReturnRule(t *testing.T) {
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
			Code: `function foo() { if (true) return undefined; else return; }`,
			Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}},
		},
		{
			Code: `function foo() { return undefined; }`,
			Options: []interface{}{map[string]interface{}{"treatUndefinedAsUnspecified": true}},
		},
		{
			Code: `const foo = () => { if (true) return undefined; else return; };`,
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

		// Single return statement
		{Code: `function foo() { if (true) { return 1; } }`},

	}, []rule_tester.InvalidTestCase{
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
				{MessageId: "missingReturnValue"},
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
