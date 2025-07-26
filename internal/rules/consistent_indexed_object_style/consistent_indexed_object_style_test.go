package consistent_indexed_object_style

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestConsistentIndexedObjectStyleRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConsistentIndexedObjectStyleRule, []rule_tester.ValidTestCase{
		// 'record' (default)
		// Record
		{Code: "type Foo = Record<string, any>;"},

		// Interface
		{Code: "interface Foo {}"},
		{Code: `
interface Foo {
  bar: string;
}
		`},
		{Code: `
interface Foo {
  bar: string;
  [key: string]: any;
}
		`},
		{Code: `
interface Foo {
  [key: string]: any;
  bar: string;
}
		`},
		// circular
		{Code: "type Foo = { [key: string]: string | Foo };"},
		{Code: "type Foo = { [key: string]: Foo };"},
		{Code: "type Foo = { [key: string]: Foo } | Foo;"},
		{Code: "type Foo = { [key in string]: Foo };"},
		{Code: `
interface Foo {
  [key: string]: Foo;
}
		`},
		{Code: `
interface Foo<T> {
  [key: string]: Foo<T>;
}
		`},
		{Code: `
interface Foo<T> {
  [key: string]: Foo<T> | string;
}
		`},
		{Code: `
interface Foo {
  [s: string]: Foo & {};
}
		`},
		{Code: `
interface Foo {
  [s: string]: Foo | string;
}
		`},
		{Code: `
interface Foo<T> {
  [s: string]: Foo extends T ? string : number;
}
		`},
		{Code: `
interface Foo<T> {
  [s: string]: T extends Foo ? string : number;
}
		`},
		{Code: `
interface Foo<T> {
  [s: string]: T extends true ? Foo : number;
}
		`},
		{Code: `
interface Foo<T> {
  [s: string]: T extends true ? string : Foo;
}
		`},
		{Code: `
interface Foo {
  [s: string]: Foo[number];
}
		`},
		{Code: `
interface Foo {
  [s: string]: {}[Foo];
}
		`},

		// circular (indirect)
		{Code: `
interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo1;
}
		`},
		{Code: `
interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo1;
}
		`},
		{Code: `
interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Record<string, Foo1>;
}
		`},
		{Code: `
type Foo1 = {
  [key: string]: Foo2;
};

type Foo2 = {
  [key: string]: Foo3;
};

type Foo3 = {
  [key: string]: Foo1;
};
		`},
		{Code: `
interface Foo1 {
  [key: string]: Foo2;
}

type Foo2 = {
  [key: string]: Foo3;
};

interface Foo3 {
  [key: string]: Foo1;
}
		`},
		{Code: `
type Foo1 = {
  [key: string]: Foo2;
};

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo1;
}
		`},
		{Code: `
type ExampleUnion = boolean | number;

type ExampleRoot = ExampleUnion | ExampleObject;

interface ExampleObject {
  [key: string]: ExampleRoot;
}
		`},
		{Code: `
type Bar<K extends string = never> = {
  [k in K]: Bar;
};
		`},
		{Code: `
type Bar<K extends string = never> = {
  [k in K]: Foo;
};

type Foo = Bar;
		`},

		// Type literal
		{Code: "type Foo = {};"},
		{Code: `
type Foo = {
  bar: string;
  [key: string]: any;
};
		`},
		{Code: `
type Foo = {
  bar: string;
};
		`},
		{Code: `
type Foo = {
  [key: string]: any;
  bar: string;
};
		`},

		// Generic
		{Code: `
type Foo = Generic<{
  [key: string]: any;
  bar: string;
}>;
		`},

		// Function types
		{Code: "function foo(arg: { [key: string]: any; bar: string }) {}"},
		{Code: "function foo(): { [key: string]: any; bar: string } {}"},

		// Invalid syntax allowed by the parser
		{Code: "type Foo = { [key: string] };"},
		{Code: "type Foo = { [] };"},
		{Code: `
interface Foo {
  [key: string];
}
		`},
		{Code: `
interface Foo {
  [];
}
		`},

		// 'index-signature'
		// Unhandled type
		{
			Code:    "type Foo = Misc<string, unknown>;",
			Options: []interface{}{"index-signature"},
		},

		// Invalid record
		{
			Code:    "type Foo = Record;",
			Options: []interface{}{"index-signature"},
		},
		{
			Code:    "type Foo = Record<string>;",
			Options: []interface{}{"index-signature"},
		},
		{
			Code:    "type Foo = Record<string, number, unknown>;",
			Options: []interface{}{"index-signature"},
		},

		// Type literal
		{
			Code:    "type Foo = { [key: string]: any };",
			Options: []interface{}{"index-signature"},
		},

		// Generic
		{
			Code:    "type Foo = Generic<{ [key: string]: any }>;",
			Options: []interface{}{"index-signature"},
		},

		// Function types
		{
			Code:    "function foo(arg: { [key: string]: any }) {}",
			Options: []interface{}{"index-signature"},
		},
		{
			Code:    "function foo(): { [key: string]: any } {}",
			Options: []interface{}{"index-signature"},
		},

		// Namespace
		{
			Code:    "type T = A.B;",
			Options: []interface{}{"index-signature"},
		},

		{
			// mapped type that uses the key cannot be converted to record
			Code: "type T = { [key in Foo]: key | number };",
		},
		{
			Code: "function foo(e: { readonly [key in PropertyKey]-?: key }) {}",
		},

		{
			// `in keyof` mapped types are not convertible to Record.
			Code: `
function f(): {
  // intentionally not using a Record to preserve optionals
  [k in keyof ParseResult]: unknown;
} {
  return {};
}
			`,
		},
	}, []rule_tester.InvalidTestCase{
		// Interface
		{
			Code: `
interface Foo {
  [key: string]: any;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, any>;
			`},
		},

		// Readonly interface
		{
			Code: `
interface Foo {
  readonly [key: string]: any;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Readonly<Record<string, any>>;
			`},
		},

		// Interface with generic parameter
		{
			Code: `
interface Foo<A> {
  [key: string]: A;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<A> = Record<string, A>;
			`},
		},

		// Interface with generic parameter and default value
		{
			Code: `
interface Foo<A = any> {
  [key: string]: A;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<A = any> = Record<string, A>;
			`},
		},

		// Interface with extends
		{
			Code: `
interface B extends A {
  [index: number]: unknown;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
		},
		// Readonly interface with generic parameter
		{
			Code: `
interface Foo<A> {
  readonly [key: string]: A;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<A> = Readonly<Record<string, A>>;
			`},
		},

		// Interface with multiple generic parameters
		{
			Code: `
interface Foo<A, B> {
  [key: A]: B;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<A, B> = Record<A, B>;
			`},
		},

		// Readonly interface with multiple generic parameters
		{
			Code: `
interface Foo<A, B> {
  readonly [key: A]: B;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<A, B> = Readonly<Record<A, B>>;
			`},
		},

		// Type literal
		{
			Code: "type Foo = { [key: string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 12},
			},
			Output: []string{"type Foo = Record<string, any>;"},
		},

		// Readonly type literal
		{
			Code: "type Foo = { readonly [key: string]: any };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 12},
			},
			Output: []string{"type Foo = Readonly<Record<string, any>>;"},
		},

		// Generic
		{
			Code: "type Foo = Generic<{ [key: string]: any }>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 20},
			},
			Output: []string{"type Foo = Generic<Record<string, any>>;"},
		},

		// Readonly Generic
		{
			Code: "type Foo = Generic<{ readonly [key: string]: any }>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 20},
			},
			Output: []string{"type Foo = Generic<Readonly<Record<string, any>>>;"},
		},

		// Function types
		{
			Code: "function foo(arg: { [key: string]: any }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 19},
			},
			Output: []string{"function foo(arg: Record<string, any>) {}"},
		},
		{
			Code: "function foo(): { [key: string]: any } {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 17},
			},
			Output: []string{"function foo(): Record<string, any> {}"},
		},

		// Readonly function types
		{
			Code: "function foo(arg: { readonly [key: string]: any }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 19},
			},
			Output: []string{"function foo(arg: Readonly<Record<string, any>>) {}"},
		},
		{
			Code: "function foo(): { readonly [key: string]: any } {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 17},
			},
			Output: []string{"function foo(): Readonly<Record<string, any>> {}"},
		},

		// Never
		// Type literal
		{
			Code: "type Foo = Record<string, any>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 12},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"type Foo = { [key: string]: any };"},
		},

		// Type literal with generic parameter
		{
			Code: "type Foo<T> = Record<string, T>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 15},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"type Foo<T> = { [key: string]: T };"},
		},

		// Circular
		{
			Code: "type Foo = { [k: string]: A.Foo };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 12},
			},
			Output: []string{"type Foo = Record<string, A.Foo>;"},
		},
		{
			Code: "type Foo = { [key: string]: AnotherFoo };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 12},
			},
			Output: []string{"type Foo = Record<string, AnotherFoo>;"},
		},
		{
			Code: "type Foo = { [key: string]: { [key: string]: Foo } };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 29},
			},
			Output: []string{"type Foo = { [key: string]: Record<string, Foo> };"},
		},
		{
			Code: "type Foo = { [key: string]: string } | Foo;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 12},
			},
			Output: []string{"type Foo = Record<string, string> | Foo;"},
		},
		{
			Code: `
interface Foo<T> {
  [k: string]: T;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo<T> = Record<string, T>;
			`},
		},
		{
			Code: `
interface Foo {
  [k: string]: A.Foo;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, A.Foo>;
			`},
		},
		{
			Code: `
interface Foo {
  [k: string]: { [key: string]: Foo };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 3, Column: 16},
			},
			Output: []string{`
interface Foo {
  [k: string]: Record<string, Foo>;
}
			`},
		},
		{
			Code: `
interface Foo {
  [key: string]: { foo: Foo };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, { foo: Foo }>;
			`},
		},
		{
			Code: `
interface Foo {
  [key: string]: Foo[];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, Foo[]>;
			`},
		},
		{
			Code: `
interface Foo {
  [key: string]: () => Foo;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, () => Foo>;
			`},
		},
		{
			Code: `
interface Foo {
  [s: string]: [Foo];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo = Record<string, [Foo]>;
			`},
		},

		// Circular (indirect)
		{
			Code: `
interface Foo1 {
  [key: string]: Foo2;
}

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo2;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo1 = Record<string, Foo2>;

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo2;
}
			`},
		},
		{
			Code: `
interface Foo1 {
  [key: string]: Record<string, Foo2>;
}

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo2;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 1},
			},
			Output: []string{`
type Foo1 = Record<string, Record<string, Foo2>>;

interface Foo2 {
  [key: string]: Foo3;
}

interface Foo3 {
  [key: string]: Foo2;
}
			`},
		},
		{
			Code: `
type Foo1 = {
  [key: string]: { foo2: Foo2 };
};

type Foo2 = {
  [key: string]: Foo3;
};

type Foo3 = {
  [key: string]: Record<string, Foo1>;
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 13},
			},
			Output: []string{`
type Foo1 = Record<string, { foo2: Foo2 }>;

type Foo2 = {
  [key: string]: Foo3;
};

type Foo3 = {
  [key: string]: Record<string, Foo1>;
};
			`},
		},
		{
			Code: `
type Foos<K extends string = never> = {
  [k in K]: { foo: Foo };
};

type Foo = Foos;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 39},
			},
			Output: []string{`
type Foos<K extends string = never> = Record<K, { foo: Foo }>;

type Foo = Foos;
			`},
		},
		{
			Code: `
type Foos<K extends string = never> = {
  [k in K]: Foo[];
};

type Foo = Foos;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 39},
			},
			Output: []string{`
type Foos<K extends string = never> = Record<K, Foo[]>;

type Foo = Foos;
			`},
		},

		// Generic
		{
			Code: "type Foo = Generic<Record<string, any>>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 20},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"type Foo = Generic<{ [key: string]: any }>;"},
		},

		// Record with an index node that may potentially break index-signature style
		{
			Code: "type Foo = Record<string | number, any>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIndexSignature",
					Line:     1,
					Column:   12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferIndexSignatureSuggestion",
							Output:    "type Foo = { [key: string | number]: any };",
						},
					},
				},
			},
			Options: []interface{}{"index-signature"},
		},
		{
			Code: "type Foo = Record<Exclude<'a' | 'b' | 'c', 'a'>, any>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferIndexSignature",
					Line:     1,
					Column:   12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "preferIndexSignatureSuggestion",
							Output:    "type Foo = { [key: Exclude<'a' | 'b' | 'c', 'a'>]: any };",
						},
					},
				},
			},
			Options: []interface{}{"index-signature"},
		},

		// Record with valid index node should use an auto-fix
		{
			Code: "type Foo = Record<number, any>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 12},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"type Foo = { [key: number]: any };"},
		},
		{
			Code: "type Foo = Record<symbol, any>;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 12},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"type Foo = { [key: symbol]: any };"},
		},

		// Function types
		{
			Code: "function foo(arg: Record<string, any>) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 19},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"function foo(arg: { [key: string]: any }) {}"},
		},
		{
			Code: "function foo(): Record<string, any> {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferIndexSignature", Line: 1, Column: 17},
			},
			Options: []interface{}{"index-signature"},
			Output:  []string{"function foo(): { [key: string]: any } {}"},
		},
		{
			Code: "type T = { readonly [key in string]: number };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 10},
			},
			Output: []string{"type T = Readonly<Record<string, number>>;"},
		},
		{
			Code: "type T = { +readonly [key in string]: number };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 10},
			},
			Output: []string{"type T = Readonly<Record<string, number>>;"},
		},
		{
			// There is no fix, since there isn't a builtin Mutable<T> :(
			Code: "type T = { -readonly [key in string]: number };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 10},
			},
		},
		{
			Code: "type T = { [key in string]: number };",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 1, Column: 10},
			},
			Output: []string{"type T = Record<string, number>;"},
		},
		{
			Code: "function foo(e: { [key in PropertyKey]?: string }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     1,
					Column:   17,
					EndLine:  1,
					EndColumn: 50,
				},
			},
			Output: []string{"function foo(e: Partial<Record<PropertyKey, string>>) {}"},
		},
		{
			Code: "function foo(e: { [key in PropertyKey]+?: string }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     1,
					Column:   17,
					EndLine:  1,
					EndColumn: 51,
				},
			},
			Output: []string{"function foo(e: Partial<Record<PropertyKey, string>>) {}"},
		},
		{
			Code: "function foo(e: { [key in PropertyKey]-?: string }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     1,
					Column:   17,
					EndLine:  1,
					EndColumn: 51,
				},
			},
			Output: []string{"function foo(e: Required<Record<PropertyKey, string>>) {}"},
		},
		{
			Code: "function foo(e: { readonly [key in PropertyKey]-?: string }) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     1,
					Column:   17,
					EndLine:  1,
					EndColumn: 60,
				},
			},
			Output: []string{"function foo(e: Readonly<Required<Record<PropertyKey, string>>>) {}"},
		},
		{
			Code: `
type Options = [
  { [Type in (typeof optionTesters)[number]['option']]?: boolean } & {
    allow?: TypeOrValueSpecifier[];
  },
];
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     3,
					Column:   3,
					EndLine:  3,
					EndColumn: 67,
				},
			},
			Output: []string{`
type Options = [
  Partial<Record<(typeof optionTesters)[number]['option'], boolean>> & {
    allow?: TypeOrValueSpecifier[];
  },
];
			`},
		},
		{
			Code: `
export type MakeRequired<Base, Key extends keyof Base> = {
  [K in Key]-?: NonNullable<Base[Key]>;
} & Omit<Base, Key>;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     2,
					Column:   58,
					EndLine:  4,
					EndColumn: 2,
				},
			},
			Output: []string{`
export type MakeRequired<Base, Key extends keyof Base> = Required<Record<Key, NonNullable<Base[Key]>>> & Omit<Base, Key>;
			`},
		},
		{
			// in parenthesized expression is convertible to Record
			Code: `
function f(): {
  [k in (keyof ParseResult)]: unknown;
} {
  return {};
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     2,
					Column:   15,
					EndLine:  4,
					EndColumn: 2,
				},
			},
			Output: []string{`
function f(): Record<keyof ParseResult, unknown> {
  return {};
}
			`},
		},

		// missing index signature type annotation while checking for a recursive type
		{
			Code: `
interface Foo {
  [key: string]: Bar;
}

interface Bar {
  [key: string];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "preferRecord",
					Line:     2,
					Column:   1,
					EndLine:  4,
					EndColumn: 2,
				},
			},
			Output: []string{`
type Foo = Record<string, Bar>;

interface Bar {
  [key: string];
}
			`},
		},

		{
			Code: `
type Foo = {
  [k in string];
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "preferRecord", Line: 2, Column: 12},
			},
			Output: []string{`
type Foo = Record<string, any>;
			`},
		},
	})
}