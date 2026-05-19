// TestNoUnnecessaryTypeConversionUpstream migrates the full valid/invalid
// suite from upstream
// packages/eslint-plugin/tests/rules/no-unnecessary-type-conversion.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// no_unnecessary_type_conversion_extras_test.go.
package no_unnecessary_type_conversion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeConversionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeConversionRule,
		[]rule_tester.ValidTestCase{
			// ---- standard type conversions are valid ----
			{Code: `String(1);`},
			{Code: `(1).toString();`},
			{Code: "`${1}`;"},
			{Code: `'' + 1;`},
			{Code: `1 + '';`},
			{Code: `
let str = 1;
str += '';
`},
			{Code: `Number('2');`},
			{Code: `+'2';`},
			{Code: `~~'2';`},
			{Code: `~~1.1;`},
			{Code: `~~-1.1;`},
			{Code: `~~(1.5 + 2.3);`},
			{Code: `~~(1 / 3);`},
			{Code: `Boolean(0);`},
			{Code: `!!0;`},
			{Code: `BigInt(3);`},

			// ---- things that look similar but are not type-conversion idioms ----
			{Code: `new String('asdf');`},
			{Code: `new Number(2);`},
			{Code: `new Boolean(true);`},
			{Code: `!false;`},
			{Code: `~2;`},
			{Code: `
function String(value: unknown) {
  return value;
}
String('asdf');
export {};
`},
			{Code: `
function Number(value: unknown) {
  return value;
}
Number(2);
export {};
`},
			{Code: `
function Boolean(value: unknown) {
  return value;
}
Boolean(true);
export {};
`},
			{Code: `
function BigInt(value: unknown) {
  return value;
}
BigInt(3n);
export {};
`},
			{Code: `
function toString(value: unknown) {
  return value;
}
toString('asdf');
`},
			{Code: `
export {};
declare const toString: string;
toString.toUpperCase();
`},

			// ---- using conversion idioms to unbox boxed primitives is valid ----
			{Code: `String(new String());`},
			{Code: `new String().toString();`},
			{Code: `'' + new String();`},
			{Code: `new String() + '';`},
			{Code: `
let str = new String();
str += '';
`},
			{Code: `Number(new Number());`},
			{Code: `+new Number();`},
			{Code: `~~new Number();`},
			{Code: `Boolean(new Boolean());`},
			{Code: `!!new Boolean();`},
			{Code: `
enum CustomIds {
  Id1 = 'id1',
  Id2 = 'id2',
}
const customId = 'id1';
const compareWithToString = customId === CustomIds.Id1.toString();
`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `String('asdf');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			{
				Code: `'asdf'.toString();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    8,
						EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			{
				Code: `'' + 'asdf';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			{
				Code: `'asdf' + '';`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    7,
						EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `'asdf';`},
							{MessageId: "suggestSatisfies", Output: `'asdf' satisfies string;`},
						},
					},
				},
			},
			{
				Code: `
let str = 'asdf';
str += '';
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
let str = 'asdf';

`},
							{MessageId: "suggestSatisfies", Output: `
let str = 'asdf';
str satisfies string;
`},
						},
					},
				},
			},
			{
				Code: `
let str = 'asdf';
'asdf' + (str += '');
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    11,
						EndLine:   3,
						EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
let str = 'asdf';
'asdf' + (str);
`},
							{MessageId: "suggestSatisfies", Output: `
let str = 'asdf';
'asdf' + (str satisfies string);
`},
						},
					},
				},
			},
			{
				Code: `Number(123);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `123;`},
							{MessageId: "suggestSatisfies", Output: `123 satisfies number;`},
						},
					},
				},
			},
			{
				Code: `+123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 2,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `123;`},
							{MessageId: "suggestSatisfies", Output: `123 satisfies number;`},
						},
					},
				},
			},
			{
				Code: `~~123;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `123;`},
							{MessageId: "suggestSatisfies", Output: `123 satisfies number;`},
						},
					},
				},
			},
			{
				Code: `Boolean(true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `true;`},
							{MessageId: "suggestSatisfies", Output: `true satisfies boolean;`},
						},
					},
				},
			},
			{
				Code: `!!true;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `true;`},
							{MessageId: "suggestSatisfies", Output: `true satisfies boolean;`},
						},
					},
				},
			},
			{
				Code: `BigInt(3n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `3n;`},
							{MessageId: "suggestSatisfies", Output: `3n satisfies bigint;`},
						},
					},
				},
			},

			// ---- generics that extend a primitive ----
			{
				Code: `
function f<T extends string>(x: T) {
  return String(x);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
function f<T extends string>(x: T) {
  return x;
}
`},
							{MessageId: "suggestSatisfies", Output: `
function f<T extends string>(x: T) {
  return x satisfies string;
}
`},
						},
					},
				},
			},
			{
				Code: `
function f<T extends number>(x: T) {
  return Number(x);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
function f<T extends number>(x: T) {
  return x;
}
`},
							{MessageId: "suggestSatisfies", Output: `
function f<T extends number>(x: T) {
  return x satisfies number;
}
`},
						},
					},
				},
			},
			{
				Code: `
function f<T extends boolean>(x: T) {
  return Boolean(x);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
function f<T extends boolean>(x: T) {
  return x;
}
`},
							{MessageId: "suggestSatisfies", Output: `
function f<T extends boolean>(x: T) {
  return x satisfies boolean;
}
`},
						},
					},
				},
			},
			{
				Code: `
function f<T extends bigint>(x: T) {
  return BigInt(x);
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    10,
						EndLine:   3,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
function f<T extends bigint>(x: T) {
  return x;
}
`},
							{MessageId: "suggestSatisfies", Output: `
function f<T extends bigint>(x: T) {
  return x satisfies bigint;
}
`},
						},
					},
				},
			},

			// ---- fixes preserve parentheses where precedence requires it ----
			{
				Code: `String('a' + 'b').length;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `('a' + 'b').length;`},
							{MessageId: "suggestSatisfies", Output: `(('a' + 'b') satisfies string).length;`},
						},
					},
				},
			},
			{
				Code: `('a' + 'b').toString().length;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    13,
						EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `('a' + 'b').length;`},
							{MessageId: "suggestSatisfies", Output: `(('a' + 'b') satisfies string).length;`},
						},
					},
				},
			},
			{
				Code: `2 * +(2 + 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    5,
						EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `2 * (2 + 2);`},
							{MessageId: "suggestSatisfies", Output: `2 * ((2 + 2) satisfies number);`},
						},
					},
				},
			},
			{
				Code: `2 * Number(2 + 2);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    5,
						EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `2 * (2 + 2);`},
							{MessageId: "suggestSatisfies", Output: `2 * ((2 + 2) satisfies number);`},
						},
					},
				},
			},
			{
				Code: `false && !!(false || true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    10,
						EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `false && (false || true);`},
							{MessageId: "suggestSatisfies", Output: `false && ((false || true) satisfies boolean);`},
						},
					},
				},
			},
			{
				Code: `false && Boolean(false || true);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    10,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `false && (false || true);`},
							{MessageId: "suggestSatisfies", Output: `false && ((false || true) satisfies boolean);`},
						},
					},
				},
			},
			{
				Code: `2n * BigInt(2n + 2n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    6,
						EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `2n * (2n + 2n);`},
							{MessageId: "suggestSatisfies", Output: `2n * ((2n + 2n) satisfies bigint);`},
						},
					},
				},
			},

			// ---- suggestions add parens around the satisfies expression where syntax requires it ----
			{
				Code: `
let str = 'asdf';
String(str).length;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndLine:   3,
						EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
let str = 'asdf';
str.length;
`},
							{MessageId: "suggestSatisfies", Output: `
let str = 'asdf';
(str satisfies string).length;
`},
						},
					},
				},
			},
			{
				Code: `
let str = 'asdf';
str.toString().length;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    5,
						EndLine:   3,
						EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
let str = 'asdf';
str.length;
`},
							{MessageId: "suggestSatisfies", Output: `
let str = 'asdf';
(str satisfies string).length;
`},
						},
					},
				},
			},
			{
				Code: `~~1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `1;`},
							{MessageId: "suggestSatisfies", Output: `1 satisfies number;`},
						},
					},
				},
			},
			{
				Code: `~~-1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      1,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `(-1);`},
							{MessageId: "suggestSatisfies", Output: `(-1) satisfies number;`},
						},
					},
				},
			},
			{
				Code: `
declare const threeOrFour: 3 | 4;
~~threeOrFour;
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unnecessaryTypeConversion",
						Line:      3,
						Column:    1,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestRemove", Output: `
declare const threeOrFour: 3 | 4;
threeOrFour;
`},
							{MessageId: "suggestSatisfies", Output: `
declare const threeOrFour: 3 | 4;
threeOrFour satisfies number;
`},
						},
					},
				},
			},
		},
	)
}
