package method_signature_style

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMethodSignatureStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MethodSignatureStyleRule, []rule_tester.ValidTestCase{
		// =============================================
		// Property mode (default) — valid
		// =============================================
		{Code: `interface Test { f: (a: string) => number; }`},
		{Code: "interface Test { ['f']: (a: boolean) => void; }"},
		{Code: `interface Test { f: <T>(a: T) => T; }`},
		{Code: "interface Test { ['f']: <T extends {}>(a: T, b: T) => T; }"},
		{Code: `interface Test { get f(): number; }`},
		{Code: `interface Test { set f(value: number): void; }`},
		{Code: `type Test = { readonly f: (a: string) => number };`},
		{Code: "type Test = { ['f']?: (a: boolean) => void };"},
		{Code: `type Test = { readonly f?: <T>(a?: T) => T };`},
		{Code: "type Test = { readonly ['f']?: <T>(a: T, b: T) => T };"},
		{Code: `type Test = { get f(): number };`},
		{Code: `type Test = { set f(value: number): void };`},

		// Non-method members should be ignored in property mode
		{Code: `interface Test { (): void; }`},                    // call signature
		{Code: `interface Test { new (): Test; }`},                // construct signature
		{Code: `interface Test { [key: string]: any; }`},          // index signature
		{Code: `interface Test { x: number; y: string; }`},        // regular properties
		{Code: `interface Test { f: (a: string) => number; g: number; }`}, // mix

		// =============================================
		// Method mode — valid
		// =============================================
		{Code: `interface Test { f(a: string): number; }`, Options: []interface{}{"method"}},
		{Code: "interface Test { ['f'](a: boolean): void; }", Options: []interface{}{"method"}},
		{Code: `interface Test { f<T>(a: T): T; }`, Options: []interface{}{"method"}},
		{Code: "interface Test { ['f']<T extends {}>(a: T, b: T): T; }", Options: []interface{}{"method"}},
		{Code: `type Test = { f(a: string): number };`, Options: []interface{}{"method"}},
		{Code: "type Test = { ['f']?(a: boolean): void };", Options: []interface{}{"method"}},
		{Code: `type Test = { f?<T>(a?: T): T };`, Options: []interface{}{"method"}},
		{Code: "type Test = { ['f']?<T>(a: T, b: T): T };", Options: []interface{}{"method"}},
		{Code: `interface Test { get f(): number; }`, Options: []interface{}{"method"}},
		{Code: `interface Test { set f(value: number): void; }`, Options: []interface{}{"method"}},
		{Code: `type Test = { get f(): number };`, Options: []interface{}{"method"}},
		{Code: `type Test = { set f(value: number): void };`, Options: []interface{}{"method"}},

		// Non-function-type properties should be ignored in method mode
		{Code: `interface Test { f: string; }`, Options: []interface{}{"method"}},
		{Code: `interface Test { f: number; }`, Options: []interface{}{"method"}},
		{Code: `interface Test { f: string | number; }`, Options: []interface{}{"method"}},
		// Union/intersection containing function types — not a bare FunctionType, should be ignored
		{Code: `interface Test { f: (() => void) | string; }`, Options: []interface{}{"method"}},
		{Code: `interface Test { f: (() => void) & (() => string); }`, Options: []interface{}{"method"}},
		// Array of functions — not a bare FunctionType
		{Code: `interface Test { f: (() => void)[]; }`, Options: []interface{}{"method"}},
		// Parenthesized function type — not a bare FunctionType
		{Code: `interface Test { f: ((a: string) => void); }`, Options: []interface{}{"method"}},
	}, []rule_tester.InvalidTestCase{
		// =============================================
		// Property mode: method → property (basic)
		// =============================================
		{
			Code:   `interface Test { f(a: string): number; }`,
			Output: []string{`interface Test { f: (a: string) => number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code:   "interface Test { ['f'](a: boolean): void; }",
			Output: []string{"interface Test { ['f']: (a: boolean) => void; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code:   `interface Test { f<T>(a: T): T; }`,
			Output: []string{`interface Test { f: <T>(a: T) => T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code:   "interface Test { ['f']<T extends {}>(a: T, b: T): T; }",
			Output: []string{"interface Test { ['f']: <T extends {}>(a: T, b: T) => T; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code:   `type Test = { f(a: string): number };`,
			Output: []string{`type Test = { f: (a: string) => number };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 15},
			},
		},
		{
			Code:   "type Test = { ['f']?(a: boolean): void };",
			Output: []string{"type Test = { ['f']?: (a: boolean) => void };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 15},
			},
		},
		{
			Code:   `type Test = { f?<T>(a?: T): T };`,
			Output: []string{`type Test = { f?: <T>(a?: T) => T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 15},
			},
		},
		{
			Code:   "type Test = { ['f']?<T>(a: T, b: T): T };",
			Output: []string{"type Test = { ['f']?: <T>(a: T, b: T) => T };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 15},
			},
		},
		// Implicit return type → any
		{
			Code:   `interface MyInterface { methodReturningImplicitAny(); }`,
			Output: []string{`interface MyInterface { methodReturningImplicitAny: () => any; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 25},
			},
		},
		// Comments preserved in type params and params
		{
			Code:   "interface Test { 'f!'</* a */ T>(/* b */ x: any /* c */): void; }",
			Output: []string{"interface Test { 'f!': </* a */ T>(/* b */ x: any /* c */) => void; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		// NOTE: `readonly` on method signatures is invalid TypeScript (parse error
		// in ESLint parser), so no readonly-method test case here.

		// =============================================
		// Delimiters: semicolon, comma, none
		// =============================================
		{
			Code: "interface Foo {\n  semi(arg: string): void;\n  comma(arg: string): void,\n  none(arg: string): void\n}",
			Output: []string{"interface Foo {\n  semi: (arg: string) => void;\n  comma: (arg: string) => void,\n  none: (arg: string) => void\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},

		// =============================================
		// Multi-line parameters
		// =============================================
		{
			Code: "interface Foo {\n  x(\n    args: Pick<\n      Bar,\n      'one' | 'two' | 'three'\n    >,\n  ): Baz;\n  y(\n    foo: string,\n    bar: number,\n  ): void;\n}",
			Output: []string{"interface Foo {\n  x: (\n    args: Pick<\n      Bar,\n      'one' | 'two' | 'three'\n    >,\n  ) => Baz;\n  y: (\n    foo: string,\n    bar: number,\n  ) => void;\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 8},
			},
		},

		// =============================================
		// Method mode: property → method (basic)
		// =============================================
		{
			Code:    `interface Test { f: (a: string) => number; }`,
			Options: []interface{}{"method"},
			Output:  []string{`interface Test { f(a: string): number; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},
		{
			Code:    "interface Test { ['f']: (a: boolean) => void; }",
			Options: []interface{}{"method"},
			Output:  []string{"interface Test { ['f'](a: boolean): void; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},
		{
			Code:    `interface Test { f: <T>(a: T) => T; }`,
			Options: []interface{}{"method"},
			Output:  []string{`interface Test { f<T>(a: T): T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},
		{
			Code:    "interface Test { ['f']: <T extends {}>(a: T, b: T) => T; }",
			Options: []interface{}{"method"},
			Output:  []string{"interface Test { ['f']<T extends {}>(a: T, b: T): T; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},
		{
			Code:    `type Test = { f: (a: string) => number };`,
			Options: []interface{}{"method"},
			Output:  []string{`type Test = { f(a: string): number };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 15},
			},
		},
		{
			Code:    "type Test = { ['f']?: (a: boolean) => void };",
			Options: []interface{}{"method"},
			Output:  []string{"type Test = { ['f']?(a: boolean): void };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 15},
			},
		},
		{
			Code:    `type Test = { f?: <T>(a?: T) => T };`,
			Options: []interface{}{"method"},
			Output:  []string{`type Test = { f?<T>(a?: T): T };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 15},
			},
		},
		{
			Code:    "type Test = { ['f']?: <T>(a: T, b: T) => T };",
			Options: []interface{}{"method"},
			Output:  []string{"type Test = { ['f']?<T>(a: T, b: T): T };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 15},
			},
		},
		// Readonly property with function type → method
		{
			Code:    `type Test = { readonly f: (a: string) => number };`,
			Options: []interface{}{"method"},
			Output:  []string{`type Test = { readonly f(a: string): number };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 15},
			},
		},
		// Delimiter preservation in method mode
		{
			Code:    "interface Foo {\n  semi: (arg: string) => void;\n  comma: (arg: string) => void,\n  none: (arg: string) => void\n}",
			Options: []interface{}{"method"},
			Output:  []string{"interface Foo {\n  semi(arg: string): void;\n  comma(arg: string): void,\n  none(arg: string): void\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 2},
				{MessageId: "errorProperty", Line: 3},
				{MessageId: "errorProperty", Line: 4},
			},
		},
		// Comments preserved in method mode
		{
			Code:    "interface Test { 'f!': </* a */ T>(/* b */ x: any /* c */) => void; }",
			Options: []interface{}{"method"},
			Output:  []string{"interface Test { 'f!'</* a */ T>(/* b */ x: any /* c */): void; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},

		// =============================================
		// this return type — no autofix
		// =============================================
		{
			Code: `interface Test { f(): this; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code: `interface Test { f(): this | undefined; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code: `interface Test { f(): Promise<this>; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		{
			Code: `interface Test { f(value: number): Promise<this | undefined>; }`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		// Overloaded methods returning this — both should report, no fix
		{
			Code: "interface Test {\n  foo(): this;\n  foo(): Promise<this>;\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Overloads — merge into intersection type
		// =============================================
		// 3 overloads, same params
		{
			Code: "interface Test {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}",
			Output: []string{"interface Test {\n  foo: (() => one) & (() => two) & (() => three);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},
		// Overloads with different parameters
		{
			Code: "interface Test {\n  foo(bar: string): one;\n  foo(bar: number, baz: string): two;\n  foo(): three;\n}",
			Output: []string{"interface Test {\n  foo: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},
		// Overloads with computed property names
		{
			Code: "interface Foo {\n  [foo](bar: string): one;\n  [foo](bar: number, baz: string): two;\n  [foo](): three;\n}",
			Output: []string{"interface Foo {\n  [foo]: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},
		// Two independent overload groups in one interface
		{
			Code: "interface Foo {\n  [foo](bar: string): one;\n  [foo](bar: number, baz: string): two;\n  [foo](): three;\n  bar(arg: string): void;\n  bar(baz: number): Foo;\n}",
			Output: []string{"interface Foo {\n  [foo]: ((bar: string) => one) & ((bar: number, baz: string) => two) & (() => three);\n  bar: ((arg: string) => void) & ((baz: number) => Foo);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
				{MessageId: "errorMethod", Line: 5},
				{MessageId: "errorMethod", Line: 6},
			},
		},
		// Overloads in type literal
		{
			Code: "type Foo = {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}",
			Output: []string{"type Foo = {\n  foo: (() => one) & (() => two) & (() => three);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},
		// Overloads in declare const
		{
			Code: "declare const Foo: {\n  foo(): one;\n  foo(): two;\n  foo(): three;\n}",
			Output: []string{"declare const Foo: {\n  foo: (() => one) & (() => two) & (() => three);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},

		// =============================================
		// Module/namespace declaration — no autofix
		// =============================================
		{
			Code: "declare global {\n  namespace jest {\n    interface Matchers<R> {\n      toHaveProperties(): R;\n    }\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 4},
			},
		},
		// Nested namespace — multiple methods, all unfixable
		{
			Code: "declare global {\n  namespace jest {\n    interface Matchers<R, T> {\n      toHaveProp<K extends keyof T>(name: K, value?: T[K]): R;\n      toHaveProps(props: Partial<T>): R;\n    }\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 4},
				{MessageId: "errorMethod", Line: 5},
			},
		},

		// =============================================
		// Rest parameters
		// =============================================
		{
			Code:   `interface Test { f(...args: any[]): void; }`,
			Output: []string{`interface Test { f: (...args: any[]) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Multiple type parameters
		// =============================================
		{
			Code:   `interface Test { f<T, U>(a: T, b: U): [T, U]; }`,
			Output: []string{`interface Test { f: <T, U>(a: T, b: U) => [T, U]; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// No parameters, explicit void return
		// =============================================
		{
			Code:   `interface Test { f(): void; }`,
			Output: []string{`interface Test { f: () => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Numeric literal key
		// =============================================
		{
			Code:   `interface Test { 0(a: string): void; }`,
			Output: []string{`interface Test { 0: (a: string) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Nested type literal (inner method should also be flagged)
		// =============================================
		{
			Code:   `type Outer = { inner: { method(): void } };`,
			Output: []string{`type Outer = { inner: { method: () => void } };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 25},
			},
		},

		// =============================================
		// Method in generic constraint
		// =============================================
		{
			Code:   `interface Foo<T extends { bar(): void }> { baz(): T; }`,
			Output: []string{`interface Foo<T extends { bar: () => void }> { baz: () => T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 27},
				{MessageId: "errorMethod", Line: 1, Column: 44},
			},
		},

		// =============================================
		// Complex return types
		// =============================================
		// Conditional type return
		{
			Code:   `interface Test { f<T>(a: T): T extends string ? number : boolean; }`,
			Output: []string{`interface Test { f: <T>(a: T) => T extends string ? number : boolean; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		// Tuple return
		{
			Code:   `interface Test { f(a: number): [string, number]; }`,
			Output: []string{`interface Test { f: (a: number) => [string, number]; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Method mode: no-param function property → method
		// =============================================
		{
			Code:    `interface Test { f: () => void; }`,
			Options: []interface{}{"method"},
			Output:  []string{`interface Test { f(): void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Multiple members — only methods are flagged
		// =============================================
		{
			Code:   `interface Test { x: number; f(a: string): void; y: string; }`,
			Output: []string{`interface Test { x: number; f: (a: string) => void; y: string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 29},
			},
		},

		// =============================================
		// Non-adjacent overloads (interspersed with other members)
		// Two fix rounds: round 1 merges foo overloads (removal spans past bar),
		// round 2 converts bar.
		// =============================================
		{
			Code: "interface Test {\n  foo(): one;\n  bar(): void;\n  foo(): two;\n}",
			Output: []string{
				"interface Test {\n  foo: (() => one) & (() => two);\n  bar(): void;\n}",
				"interface Test {\n  foo: (() => one) & (() => two);\n  bar: () => void;\n}",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
				{MessageId: "errorMethod", Line: 4},
			},
		},

		// =============================================
		// Minimum overload: exactly 2
		// =============================================
		{
			Code:   "interface Test {\n  f(a: string): void;\n  f(a: number): void;\n}",
			Output: []string{"interface Test {\n  f: ((a: string) => void) & ((a: number) => void);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Overload: first returns this → all skip fix
		// =============================================
		{
			Code: "interface Test {\n  foo(): this;\n  foo(a: string): string;\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// `this` as parameter (NOT return type) — should fix normally
		// =============================================
		{
			Code:   `interface Test { f(this: Foo, a: string): void; }`,
			Output: []string{`interface Test { f: (this: Foo, a: string) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// No params but has type params
		// =============================================
		{
			Code:   `interface Test { f<T>(): T; }`,
			Output: []string{`interface Test { f: <T>() => T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Default type parameter
		// =============================================
		{
			Code:   `interface Test { f<T = string>(): T; }`,
			Output: []string{`interface Test { f: <T = string>() => T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Method in intersection type's type literal
		// =============================================
		{
			Code:   `type Test = A & { method(): void };`,
			Output: []string{`type Test = A & { method: () => void };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 19},
			},
		},

		// =============================================
		// declare module (not declare global) — still a module, no fix
		// =============================================
		{
			Code: "declare module 'foo' {\n  interface Bar {\n    baz(): void;\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Deeply nested type literals
		// =============================================
		{
			Code:   `type T = { a: { b: { c(): void } } };`,
			Output: []string{`type T = { a: { b: { c: () => void } } };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},

		// =============================================
		// String key with special characters
		// =============================================
		{
			Code:   "interface Test { 'foo-bar'(x: number): void; }",
			Output: []string{"interface Test { 'foo-bar': (x: number) => void; }"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Method mode: generic no-param function property
		// =============================================
		{
			Code:    `interface Test { f: <T>() => T; }`,
			Options: []interface{}{"method"},
			Output:  []string{`interface Test { f<T>(): T; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Parameter with function type (nested parens in param type)
		// The ')' scan must find the correct outer paren.
		// =============================================
		{
			Code:   `interface Test { f(callback: (x: number) => void): void; }`,
			Output: []string{`interface Test { f: (callback: (x: number) => void) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		// Multiple params, last param has function type
		{
			Code:   `interface Test { f(a: string, callback: (x: number) => void): void; }`,
			Output: []string{`interface Test { f: (a: string, callback: (x: number) => void) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},
		// Method mode: property with function type param
		{
			Code:    `interface Test { f: (callback: (x: number) => void) => void; }`,
			Options: []interface{}{"method"},
			Output:  []string{`interface Test { f(callback: (x: number) => void): void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorProperty", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Overloads with different generic type params
		// =============================================
		{
			Code:   "interface Test {\n  f<T>(a: T): T;\n  f<T, U>(a: T, b: U): [T, U];\n}",
			Output: []string{"interface Test {\n  f: (<T>(a: T) => T) & (<T, U>(a: T, b: U) => [T, U]);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Optional method overloads
		// =============================================
		{
			Code:   "interface Test {\n  f?(): void;\n  f?(a: string): string;\n}",
			Output: []string{"interface Test {\n  f?: (() => void) & ((a: string) => string);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// TypeLiteral in various contexts
		// =============================================
		// In function parameter type
		{
			Code:   `function foo(x: { method(): void }) {}`,
			Output: []string{`function foo(x: { method: () => void }) {}`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In type assertion
		{
			Code:   `const x = {} as { method(): void };`,
			Output: []string{`const x = {} as { method: () => void };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In return type
		{
			Code:   `function foo(): { method(): void } { return {} as any; }`,
			Output: []string{`function foo(): { method: () => void } { return {} as any; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In variable type annotation
		{
			Code:   `let x: { method(): void };`,
			Output: []string{`let x: { method: () => void };`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In conditional type branch
		{
			Code:   `type T = true extends true ? { method(): void } : never;`,
			Output: []string{`type T = true extends true ? { method: () => void } : never;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In tuple element
		{
			Code:   `type T = [{ method(): void }];`,
			Output: []string{`type T = [{ method: () => void }];`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
		// In mapped type value
		{
			Code:   "type T = { [K in string]: { method(): void } };",
			Output: []string{"type T = { [K in string]: { method: () => void } };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},

		// =============================================
		// Single-line overloads
		// =============================================
		{
			Code:   `interface Test { f(): one; f(): two; }`,
			Output: []string{`interface Test { f: (() => one) & (() => two); }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
				{MessageId: "errorMethod", Line: 1},
			},
		},

		// =============================================
		// Overload: only second returns this → first generates fix (includes this)
		// Matches ESLint behavior: skipFix is per-node, not per-group.
		// =============================================
		{
			Code: "interface Test {\n  foo(a: string): string;\n  foo(): this;\n}",
			Output: []string{"interface Test {\n  foo: ((a: string) => string) & (() => this);\n}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 2},
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Destructured parameter
		// =============================================
		{
			Code:   `interface Test { f({ a, b }: Options): void; }`,
			Output: []string{`interface Test { f: ({ a, b }: Options) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Trailing comma in parameter list
		// =============================================
		{
			Code:   `interface Test { f(a: string,): void; }`,
			Output: []string{`interface Test { f: (a: string,) => void; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1, Column: 18},
			},
		},

		// =============================================
		// Non-declare namespace (still a ModuleDeclaration — no fix)
		// =============================================
		{
			Code: "namespace Foo {\n  interface Bar {\n    method(): void;\n  }\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 3},
			},
		},

		// =============================================
		// Type predicate return type
		// =============================================
		{
			Code:   `interface Guard { check(x: unknown): x is string; }`,
			Output: []string{`interface Guard { check: (x: unknown) => x is string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},

		// =============================================
		// Asserts return type
		// =============================================
		{
			Code:   `interface Guard { assert(x: unknown): asserts x is string; }`,
			Output: []string{`interface Guard { assert: (x: unknown) => asserts x is string; }`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "errorMethod", Line: 1},
			},
		},
	})
}
