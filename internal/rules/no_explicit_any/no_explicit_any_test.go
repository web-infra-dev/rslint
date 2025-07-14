package no_explicit_any

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExplicitAnyRule, []rule_tester.ValidTestCase{
		// Valid cases - no any type annotations
		{Code: "const x: string = 'hello';"},
		{Code: "const x: number = 42;"},
		{Code: "const x: boolean = true;"},
		{Code: "function foo(x: string): number { return 1; }"},
		{Code: "const x: unknown = someValue;"},
		{Code: "const x: never = (() => { throw new Error(); })();"},
		{Code: "const x: PropertyKey = 'key';"},
		{Code: "interface Foo { bar: string; }"},
		{Code: "type Foo = string | number;"},
		{Code: "const x: object = {};"},
		{Code: "const x: {} = {};"},
		{Code: "class Foo { constructor(public prop: string) {} }"},
		{Code: "const arrow = (x: string): number => 1;"},
		{Code: "type Union = string | number | boolean;"},
		{Code: "type Intersection = { a: string } & { b: number };"},
		// Edge cases
		{Code: "const notAnyKeyword = 'any';"},
		{Code: "const obj = { any: 'value' };"},
		{Code: "function anyFunc() { return 'not any type'; }"},
	}, []rule_tester.InvalidTestCase{
		// Basic any usage
		{
			Code: "const x: any = 'hello';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    10,
					EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const x: unknown = 'hello';",
						},
						{
							MessageId: "suggestNever",
							Output:    "const x: never = 'hello';",
						},
					},
				},
			},
		},
		// Function parameter
		{
			Code: "function foo(x: any): void {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    17,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(x: unknown): void {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(x: never): void {}",
						},
					},
				},
			},
		},
		// Function return type
		{
			Code: "function foo(): any { return 1; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    17,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(): unknown { return 1; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(): never { return 1; }",
						},
					},
				},
			},
		},
		// Array type
		{
			Code: "const x: any[] = [1, 2, 3];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    10,
					EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const x: unknown[] = [1, 2, 3];",
						},
						{
							MessageId: "suggestNever",
							Output:    "const x: never[] = [1, 2, 3];",
						},
					},
				},
			},
		},
		// Generic type
		{
			Code: "const x: Array<any> = [1, 2, 3];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const x: Array<unknown> = [1, 2, 3];",
						},
						{
							MessageId: "suggestNever",
							Output:    "const x: Array<never> = [1, 2, 3];",
						},
					},
				},
			},
		},
		// Interface property
		{
			Code: "interface Foo { bar: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    22,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "interface Foo { bar: unknown; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "interface Foo { bar: never; }",
						},
					},
				},
			},
		},
		// Type alias
		{
			Code: "type Foo = any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    12,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "type Foo = unknown;",
						},
						{
							MessageId: "suggestNever",
							Output:    "type Foo = never;",
						},
					},
				},
			},
		},
		// Union type
		{
			Code: "const x: string | any = 'hello';",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const x: string | unknown = 'hello';",
						},
						{
							MessageId: "suggestNever",
							Output:    "const x: string | never = 'hello';",
						},
					},
				},
			},
		},
		// Intersection type
		{
			Code: "const x: { a: string } & any = {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    26,
					EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const x: { a: string } & unknown = {};",
						},
						{
							MessageId: "suggestNever",
							Output:    "const x: { a: string } & never = {};",
						},
					},
				},
			},
		},
		// Arrow function parameter
		{
			Code: "const fn = (x: any) => x;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const fn = (x: unknown) => x;",
						},
						{
							MessageId: "suggestNever",
							Output:    "const fn = (x: never) => x;",
						},
					},
				},
			},
		},
		// Arrow function return type
		{
			Code: "const fn = (): any => 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const fn = (): unknown => 1;",
						},
						{
							MessageId: "suggestNever",
							Output:    "const fn = (): never => 1;",
						},
					},
				},
			},
		},
		// Class property
		{
			Code: "class Foo { prop: any; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "class Foo { prop: unknown; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "class Foo { prop: never; }",
						},
					},
				},
			},
		},
		// Constructor parameter
		{
			Code: "class Foo { constructor(public prop: any) {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    38,
					EndColumn: 41,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "class Foo { constructor(public prop: unknown) {} }",
						},
						{
							MessageId: "suggestNever",
							Output:    "class Foo { constructor(public prop: never) {} }",
						},
					},
				},
			},
		},
		// Method parameter
		{
			Code: "class Foo { method(param: any): void {} }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    27,
					EndColumn: 30,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "class Foo { method(param: unknown): void {} }",
						},
						{
							MessageId: "suggestNever",
							Output:    "class Foo { method(param: never): void {} }",
						},
					},
				},
			},
		},
		// Method return type
		{
			Code: "class Foo { method(): any { return 1; } }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "class Foo { method(): unknown { return 1; } }",
						},
						{
							MessageId: "suggestNever",
							Output:    "class Foo { method(): never { return 1; } }",
						},
					},
				},
			},
		},
		// Multiple any types in one declaration
		{
			Code: "function foo(a: any, b: any): any { return 1; }",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    17,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(a: unknown, b: any): any { return 1; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(a: never, b: any): any { return 1; }",
						},
					},
				},
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    25,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(a: any, b: unknown): any { return 1; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(a: any, b: never): any { return 1; }",
						},
					},
				},
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    31,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(a: any, b: any): unknown { return 1; }",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(a: any, b: any): never { return 1; }",
						},
					},
				},
			},
		},
	})
}