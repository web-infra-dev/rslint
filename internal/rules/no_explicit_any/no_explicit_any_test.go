package no_explicit_any

import (
	"testing"

	"github.com/typescript-eslint/tsgolint/internal/rule_tester"
	"github.com/typescript-eslint/tsgolint/internal/rules/fixtures"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExplicitAnyRule, []rule_tester.ValidTestCase{
		// Basic valid cases - no any type annotations
		{Code: "const number: number = 1;"},
		{Code: "function greet(): string {}"},
		{Code: "function greet(): Array<string> {}"},
		{Code: "function greet(): string[] {}"},
		{Code: "function greet(): Array<Array<string>> {}"},
		{Code: "function greet(): Array<string[]> {}"},
		{Code: "function greet(param: Array<string>): Array<string> {}"},
		
		// Class examples
		{Code: `class Greeter {
  message: string;
}`},
		{Code: `class Greeter {
  message: Array<string>;
}`},
		{Code: `class Greeter {
  message: string[];
}`},
		{Code: `class Greeter {
  message: Array<Array<string>>;
}`},
		{Code: `class Greeter {
  message: Array<string[]>;
}`},
		
		// Interface examples
		{Code: `interface Greeter {
  message: string;
}`},
		{Code: `interface Greeter {
  message: Array<string>;
}`},
		{Code: `interface Greeter {
  message: string[];
}`},
		{Code: `interface Greeter {
  message: Array<Array<string>>;
}`},
		{Code: `interface Greeter {
  message: Array<string[]>;
}`},
		
		// Type examples
		{Code: `type obj = {
  message: string;
};`},
		{Code: `type obj = {
  message: Array<string>;
};`},
		{Code: `type obj = {
  message: string[];
};`},
		{Code: `type obj = {
  message: Array<Array<string>>;
};`},
		{Code: `type obj = {
  message: Array<string[]>;
};`},
		
		// Union types
		{Code: `type obj = {
  message: string | number;
};`},
		{Code: `type obj = {
  message: string | Array<string>;
};`},
		{Code: `type obj = {
  message: string | string[];
};`},
		{Code: `type obj = {
  message: string | Array<Array<string>>;
};`},
		
		// Intersection types
		{Code: `type obj = {
  message: string & number;
};`},
		{Code: `type obj = {
  message: string & Array<string>;
};`},
		{Code: `type obj = {
  message: string & string[];
};`},
		{Code: `type obj = {
  message: string & Array<Array<string>>;
};`},

		// Rest args with ignoreRestArgs option - these should be valid when the option is enabled
		{Code: `function foo(a: number, ...rest: any[]): void {
  return;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function foo1(...args: any[]) {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const bar1 = function (...args: any[]) {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const baz1 = (...args: any[]) => {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function foo2(...args: readonly any[]) {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const bar2 = function (...args: readonly any[]) {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const baz2 = (...args: readonly any[]) => {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function foo3(...args: Array<any>) {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const bar3 = function (...args: Array<any>) {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const baz3 = (...args: Array<any>) => {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function foo4(...args: ReadonlyArray<any>) {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const bar4 = function (...args: ReadonlyArray<any>) {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "const baz4 = (...args: ReadonlyArray<any>) => {};", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Interface signatures with rest args
		{Code: `interface Qux1 {
  (...args: any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Qux2 {
  (...args: readonly any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Qux3 {
  (...args: Array<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Qux4 {
  (...args: ReadonlyArray<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Function type parameters with rest args
		{Code: "function quux1(fn: (...args: any[]) => void): void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quux2(fn: (...args: readonly any[]) => void): void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quux3(fn: (...args: Array<any>) => void): void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quux4(fn: (...args: ReadonlyArray<any>) => void): void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Return type with rest args
		{Code: "function quuz1(): (...args: any[]) => void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quuz2(): (...args: readonly any[]) => void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quuz3(): (...args: Array<any>) => void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "function quuz4(): (...args: ReadonlyArray<any>) => void {}", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Type aliases with rest args
		{Code: "type Fred1 = (...args: any[]) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Fred2 = (...args: readonly any[]) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Fred3 = (...args: Array<any>) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Fred4 = (...args: ReadonlyArray<any>) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Constructor signatures with rest args
		{Code: "type Corge1 = new (...args: any[]) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Corge2 = new (...args: readonly any[]) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Corge3 = new (...args: Array<any>) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "type Corge4 = new (...args: ReadonlyArray<any>) => void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Interface constructor signatures with rest args
		{Code: `interface Grault1 {
  new (...args: any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Grault2 {
  new (...args: readonly any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Grault3 {
  new (...args: Array<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Grault4 {
  new (...args: ReadonlyArray<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Interface method signatures with rest args
		{Code: `interface Garply1 {
  f(...args: any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Garply2 {
  f(...args: readonly any[]): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Garply3 {
  f(...args: Array<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: `interface Garply4 {
  f(...args: ReadonlyArray<any>): void;
}`, Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Declare function with rest args
		{Code: "declare function waldo1(...args: any[]): void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "declare function waldo2(...args: readonly any[]): void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "declare function waldo3(...args: Array<any>): void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		{Code: "declare function waldo4(...args: ReadonlyArray<any>): void;", Options: map[string]interface{}{"ignoreRestArgs": true}},
		
		// Edge cases - these should not trigger the rule
		{Code: "const notAnyKeyword = 'any';"},
		{Code: "const obj = { any: 'value' };"},
		{Code: "function anyFunc() { return 'not any type'; }"},
	}, []rule_tester.InvalidTestCase{
		// Basic any usage
		{
			Code: "const number: any = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const number: unknown = 1;",
						},
						{
							MessageId: "suggestNever",
							Output:    "const number: never = 1;",
						},
					},
				},
			},
		},
		// Function with any parameter
		{
			Code: "function greet(): any {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function greet(): unknown {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function greet(): never {}",
						},
					},
				},
			},
		},
		// Function with any in Array
		{
			Code: "function greet(): Array<any> {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    25,
					EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function greet(): Array<unknown> {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function greet(): Array<never> {}",
						},
					},
				},
			},
		},
		// Function with any array
		{
			Code: "function greet(): any[] {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function greet(): unknown[] {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function greet(): never[] {}",
						},
					},
				},
			},
		},
		// Function with any parameter
		{
			Code: "function greet(param: any): string {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function greet(param: unknown): string {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function greet(param: never): string {}",
						},
					},
				},
			},
		},
		// Class with any property
		{
			Code: `class Greeter {
  message: any;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    12,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `class Greeter {
  message: unknown;
}`,
						},
						{
							MessageId: "suggestNever",
							Output: `class Greeter {
  message: never;
}`,
						},
					},
				},
			},
		},
		// Interface with any property
		{
			Code: `interface Greeter {
  message: any;
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    12,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `interface Greeter {
  message: unknown;
}`,
						},
						{
							MessageId: "suggestNever",
							Output: `interface Greeter {
  message: never;
}`,
						},
					},
				},
			},
		},
		// Type with any
		{
			Code: `type obj = {
  message: any;
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    12,
					EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `type obj = {
  message: unknown;
};`,
						},
						{
							MessageId: "suggestNever",
							Output: `type obj = {
  message: never;
};`,
						},
					},
				},
			},
		},
		// Union with any
		{
			Code: `type obj = {
  message: string | any;
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    21,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `type obj = {
  message: string | unknown;
};`,
						},
						{
							MessageId: "suggestNever",
							Output: `type obj = {
  message: string | never;
};`,
						},
					},
				},
			},
		},
		// Intersection with any
		{
			Code: `type obj = {
  message: string & any;
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    21,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `type obj = {
  message: string & unknown;
};`,
						},
						{
							MessageId: "suggestNever",
							Output: `type obj = {
  message: string & never;
};`,
						},
					},
				},
			},
		},
		// Rest args without ignoreRestArgs option - should error
		{
			Code: "function foo(...args: any[]) {}",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(...args: unknown[]) {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(...args: never[]) {}",
						},
					},
				},
			},
		},
		// Rest args with ignoreRestArgs: false - should still error
		{
			Code: "function foo(...args: any[]) {}",
			Options: map[string]interface{}{"ignoreRestArgs": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(...args: unknown[]) {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(...args: never[]) {}",
						},
					},
				},
			},
		},
		// keyof any should suggest PropertyKey
		{
			Code: "type Keys = keyof any;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestPropertyKey",
							Output:    "type Keys = PropertyKey;",
						},
					},
				},
			},
		},
		// keyof any in generic context
		{
			Code: `const integer = <
  TKey extends keyof any,
  TTarget extends { [K in TKey]: number },
>(
  target: TTarget,
  key: TKey,
) => {
  /* ... */
};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    22,
					EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestPropertyKey",
							Output: `const integer = <
  TKey extends PropertyKey,
  TTarget extends { [K in TKey]: number },
>(
  target: TTarget,
  key: TKey,
) => {
  /* ... */
};`,
						},
					},
				},
			},
		},
		// fixToUnknown option test for keyof any
		{
			Code: "type Keys = keyof any;",
			Options: map[string]interface{}{"fixToUnknown": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    19,
					EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestPropertyKey",
							Output:    "type Keys = PropertyKey;",
						},
					},
				},
			},
		},
		// fixToUnknown option test for regular any
		{
			Code: "const number: any = 1;",
			Options: map[string]interface{}{"fixToUnknown": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "const number: unknown = 1;",
						},
						{
							MessageId: "suggestNever",
							Output:    "const number: never = 1;",
						},
					},
				},
			},
		},
		// Non-rest args with ignoreRestArgs option should still error
		{
			Code: "function foo(param: any): void {}",
			Options: map[string]interface{}{"ignoreRestArgs": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    21,
					EndColumn: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "function foo(param: unknown): void {}",
						},
						{
							MessageId: "suggestNever",
							Output:    "function foo(param: never): void {}",
						},
					},
				},
			},
		},
		// Rest args that are not arrays should still error with ignoreRestArgs
		{
			Code: "type Corge5 = new (...args: any) => void;",
			Options: map[string]interface{}{"ignoreRestArgs": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    29,
					EndColumn: 32,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "type Corge5 = new (...args: unknown) => void;",
						},
						{
							MessageId: "suggestNever",
							Output:    "type Corge5 = new (...args: never) => void;",
						},
					},
				},
			},
		},
		// Interface with rest args that are not arrays should still error
		{
			Code: `interface Grault5 {
  new (...args: any): void;
}`,
			Options: map[string]interface{}{"ignoreRestArgs": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    17,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `interface Grault5 {
  new (...args: unknown): void;
}`,
						},
						{
							MessageId: "suggestNever",
							Output: `interface Grault5 {
  new (...args: never): void;
}`,
						},
					},
				},
			},
		},
		// Interface method with rest args that are not arrays should still error
		{
			Code: `interface Garply5 {
  f(...args: any): void;
}`,
			Options: map[string]interface{}{"ignoreRestArgs": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      2,
					Column:    14,
					EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output: `interface Garply5 {
  f(...args: unknown): void;
}`,
						},
						{
							MessageId: "suggestNever",
							Output: `interface Garply5 {
  f(...args: never): void;
}`,
						},
					},
				},
			},
		},
		// Declare function with rest args that are not arrays should still error
		{
			Code: "declare function waldo5(...args: any): void;",
			Options: map[string]interface{}{"ignoreRestArgs": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    34,
					EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    "declare function waldo5(...args: unknown): void;",
						},
						{
							MessageId: "suggestNever",
							Output:    "declare function waldo5(...args: never): void;",
						},
					},
				},
			},
		},
	})
}