package switch_exhaustiveness_check

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestSwitchExhaustivenessCheckUpstream migrates the full valid/invalid suite
// from upstream typescript-eslint's
// packages/eslint-plugin/tests/rules/switch-exhaustiveness-check.test.ts
// (v8.67.0) 1:1. Position assertions cover line/column for every invalid
// case that upstream itself asserts. rslint-specific lock-in cases live in
// switch_exhaustiveness_check_extras_test.go.
func TestSwitchExhaustivenessCheckUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SwitchExhaustivenessCheckRule,
		[]rule_tester.ValidTestCase{
			// All branches matched
			{Code: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case 'Tuesday': {
    result = 2;
    break;
  }
  case 'Wednesday': {
    result = 3;
    break;
  }
  case 'Thursday': {
    result = 4;
    break;
  }
  case 'Friday': {
    result = 5;
    break;
  }
  case 'Saturday': {
    result = 6;
    break;
  }
  case 'Sunday': {
    result = 7;
    break;
  }
}
`},
			// Other primitive literals work too
			{Code: `
type Num = 0 | 1 | 2;

function test(value: Num): number {
  switch (value) {
    case 0:
      return 0;
    case 1:
      return 1;
    case 2:
      return 2;
  }
}
`},
			{Code: `
type Bool = true | false;

function test(value: Bool): number {
  switch (value) {
    case true:
      return 1;
    case false:
      return 0;
  }
}
`},
			{Code: `
type Mix = 0 | 1 | 'two' | 'three' | true;

function test(value: Mix): number {
  switch (value) {
    case 0:
      return 0;
    case 1:
      return 1;
    case 'two':
      return 2;
    case 'three':
      return 3;
    case true:
      return 4;
  }
}
`},
			// Works with type references
			{Code: `
type A = 'a';
type B = 'b';
type C = 'c';
type Union = A | B | C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case 'b':
      return 2;
    case 'c':
      return 3;
  }
}
`},
			// Works with typeof
			{Code: `
const A = 'a';
const B = 1;
const C = true;

type Union = typeof A | typeof B | typeof C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case 1:
      return 2;
    case true:
      return 3;
  }
}
`},
			// Switch contains default clause.
			{
				Code: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  default: {
    result = 42;
  }
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			// Exhaustiveness check only works for union types...
			{Code: `
const day = 'Monday' as string;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case 'Tuesday': {
    result = 2;
    break;
  }
}
`},
			// ... and enums (at least for now).
			{Code: `
enum Enum {
  A,
  B,
}

function test(value: Enum): number {
  switch (value) {
    case Enum.A:
      return 1;
    case Enum.B:
      return 2;
  }
}
`},
			// Object union types won't work either, unless it's a discriminated union
			{Code: `
type ObjectUnion = { a: 1 } | { b: 2 };

function test(value: ObjectUnion): number {
  switch (value.a) {
    case 1:
      return 1;
  }
}
`},
			// switch with default clause on non-union type
			{
				Code: `
declare const value: number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
			},
			// switch with default clause on string type + "allowDefaultCaseForExhaustiveSwitch" option
			{
				Code: `
declare const value: string;
switch (value) {
  case 'foo':
    return 0;
  case 'bar':
    return 1;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			// switch with default clause on number type + "allowDefaultCaseForExhaustiveSwitch" option
			{
				Code: `
declare const value: number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			// switch with default clause on bigint type + "allowDefaultCaseForExhaustiveSwitch" option
			{
				Code: `
declare const value: bigint;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			// switch with default clause on symbol type + "allowDefaultCaseForExhaustiveSwitch" option
			{
				Code: `
declare const value: symbol;
const foo = Symbol('foo');
switch (value) {
  case foo:
    return 0;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			// switch with default clause on union with number + "allowDefaultCaseForExhaustiveSwitch" option
			{
				Code: `
declare const value: 0 | 1 | number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: 'literal';
switch (value) {
  case 'literal':
    return 0;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: null;
switch (value) {
  case null:
    return 0;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: undefined;
switch (value) {
  case undefined:
    return 0;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: null | undefined;
switch (value) {
  case null:
    return 0;
  case undefined:
    return 0;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: 'literal' & { _brand: true };
switch (value) {
  case 'literal':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: ('literal' & { _brand: true }) | 1;
switch (value) {
  case 'literal':
    break;
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: (1 & { _brand: true }) | 'literal' | null;
switch (value) {
  case 'literal':
    break;
  case 1:
    break;
  case null:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | (number & { foo: 'bar' });
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: number | null | undefined;
switch (value) {
  case null:
    break;
  case undefined:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"considerDefaultExhaustiveForUnions":  true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"considerDefaultExhaustiveForUnions":  true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: (string & { foo: 'bar' }) | '1';
switch (value) {
  case '1':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
const a = Symbol('a');
declare const value: typeof a | 2;
switch (value) {
  case a:
    break;
  case 2:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: string | number;
switch (value) {
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: string | number;
switch (value) {
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: string | number;
switch (value) {
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: number;
declare const a: number;
switch (value) {
  case a:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: bigint;
switch (value) {
  case 10n:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  case a:
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"considerDefaultExhaustiveForUnions":  true,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: boolean | 1;
switch (value) {
  case 1:
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"considerDefaultExhaustiveForUnions":  true,
					"requireDefaultForNonUnion":           true,
				},
			},
			{
				Code: `
declare const value: boolean | 1;
switch (value) {
  case 1:
    break;
  case true:
    break;
  case false:
    break;
  default:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
  case Aaa.Bar:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  case 'b':
    break;
}
`,
				Options: map[string]any{
					"considerDefaultExhaustiveForUnions": true,
					"requireDefaultForNonUnion":          true,
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  default:
    break;
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			{
				Code: `
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  case 'b':
    break;
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
			},
			{
				Code: `
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
    break;
  case MyEnum.Bar:
    break;
  default: {
    break;
  }
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			{
				Code: `
declare const value: boolean;
switch (value) {
  case false:
    break;
  default: {
    break;
  }
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			{
				Code: `
function foo(x: string[]) {
  switch (x[0]) {
    case 'hi':
      break;
    case undefined:
      break;
  }
}
`,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
			{
				Code: `
function foo(x: string[], y: string | undefined) {
  const a = x[0];
  if (typeof a === 'string') {
    return;
  }
  switch (y) {
    case 'hi':
      break;
    case a:
      break;
  }
}
`,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
			},
			{
				Code: `
declare const value: number;
switch (value) {
  case 0:
    break;
  case 1:
    break;
  // no default
}
`,
				Options: map[string]any{"requireDefaultForNonUnion": true},
			},
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // no default
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": true},
			},
			{
				Code: `
declare const value: 'a' | 'b';
switch (value) {
  case 'a':
    break;
  // skip default
}
`,
				Options: map[string]any{
					"considerDefaultExhaustiveForUnions": true,
					"defaultCaseCommentPattern":          `^skip\sdefault`,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
declare const value: 'literal';
switch (value) {
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: 'literal';
switch (value) {
case "literal": { throw new Error('Not implemented yet: "literal" case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: 'literal' & { _brand: true };
switch (value) {
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: 'literal' & { _brand: true };
switch (value) {
case "literal": { throw new Error('Not implemented yet: "literal" case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: ('literal' & { _brand: true }) | 1;
switch (value) {
  case 'literal':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: ('literal' & { _brand: true }) | 1;
switch (value) {
  case 'literal':
    break;
  case 1: { throw new Error('Not implemented yet: 1 case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case "2": { throw new Error('Not implemented yet: "2" case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case "2": { throw new Error('Not implemented yet: "2" case') }
}
`,
							},
						},
					},
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: (string & { foo: 'bar' }) | '1';
switch (value) {
  case '1':
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: (string & { foo: 'bar' }) | '1';
switch (value) {
  case '1':
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: (string & { foo: 'bar' }) | '1' | 1 | null | undefined;
switch (value) {
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: (string & { foo: 'bar' }) | '1' | 1 | null | undefined;
switch (value) {
case undefined: { throw new Error('Not implemented yet: undefined case') }
case null: { throw new Error('Not implemented yet: null case') }
case "1": { throw new Error('Not implemented yet: "1" case') }
case 1: { throw new Error('Not implemented yet: 1 case') }
}
`,
							},
						},
					},
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: (string & { foo: 'bar' }) | '1' | 1 | null | undefined;
switch (value) {
default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: string | number;
switch (value) {
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: string | number;
switch (value) {
  case 1:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: number;
declare const a: number;
switch (value) {
  case a:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: number;
declare const a: number;
switch (value) {
  case a:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: bigint;
switch (value) {
  case 10n:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: bigint;
switch (value) {
  case 10n:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
const a = Symbol('aa');
const b = Symbol('bb');
declare const value: typeof a | typeof b | 1;
switch (value) {
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      5,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
const a = Symbol('aa');
const b = Symbol('bb');
declare const value: typeof a | typeof b | 1;
switch (value) {
  case 1:
    break;
  case a: { throw new Error('Not implemented yet: a case') }
  case b: { throw new Error('Not implemented yet: b case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  case a:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  case a:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: boolean;
switch (value) {
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: boolean;
switch (value) {
case false: { throw new Error('Not implemented yet: false case') }
case true: { throw new Error('Not implemented yet: true case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: boolean | 1;
switch (value) {
  case false:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: boolean | 1;
switch (value) {
  case false:
    break;
  case true: { throw new Error('Not implemented yet: true case') }
  case 1: { throw new Error('Not implemented yet: 1 case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: boolean | number;
switch (value) {
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: boolean | number;
switch (value) {
  case 1:
    break;
  case false: { throw new Error('Not implemented yet: false case') }
  case true: { throw new Error('Not implemented yet: true case') }
}
`,
							},
						},
					},
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: boolean | number;
switch (value) {
  case 1:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: object;
switch (value) {
  case 1:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: object;
switch (value) {
  case 1:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1 | string;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      7,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1 | string;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
  case Aaa.Bar: { throw new Error('Not implemented yet: Aaa.Bar case') }
}
`,
							},
						},
					},
					{
						MessageId: "switchIsNotExhaustive",
						Line:      7,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1 | string;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				// Matched only one branch out of seven.
				Code: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   `Switch is not exhaustive. Cases not matched: "Tuesday" | "Wednesday" | "Thursday" | "Friday" | "Saturday" | "Sunday"`,
						Line:      14,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case "Tuesday": { throw new Error('Not implemented yet: "Tuesday" case') }
  case "Wednesday": { throw new Error('Not implemented yet: "Wednesday" case') }
  case "Thursday": { throw new Error('Not implemented yet: "Thursday" case') }
  case "Friday": { throw new Error('Not implemented yet: "Friday" case') }
  case "Saturday": { throw new Error('Not implemented yet: "Saturday" case') }
  case "Sunday": { throw new Error('Not implemented yet: "Sunday" case') }
}
`,
							},
						},
					},
				},
			},
			{
				// Didn't match all enum variants
				Code: `
enum Enum {
  A,
  B,
}

function test(value: Enum): number {
  switch (value) {
    case Enum.A:
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: Enum.B",
						Line:      8,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum Enum {
  A,
  B,
}

function test(value: Enum): number {
  switch (value) {
    case Enum.A:
      return 1;
    case Enum.B: { throw new Error('Not implemented yet: Enum.B case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
type A = 'a';
type B = 'b';
type C = 'c';
type Union = A | B | C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   `Switch is not exhaustive. Cases not matched: "b" | "c"`,
						Line:      8,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type A = 'a';
type B = 'b';
type C = 'c';
type Union = A | B | C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case "b": { throw new Error('Not implemented yet: "b" case') }
    case "c": { throw new Error('Not implemented yet: "c" case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
const A = 'a';
const B = 1;
const C = true;

type Union = typeof A | typeof B | typeof C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: true | 1",
						Line:      9,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
const A = 'a';
const B = 1;
const C = true;

type Union = typeof A | typeof B | typeof C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case true: { throw new Error('Not implemented yet: true case') }
    case 1: { throw new Error('Not implemented yet: 1 case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
type DiscriminatedUnion = { type: 'A'; a: 1 } | { type: 'B'; b: 2 };

function test(value: DiscriminatedUnion): number {
  switch (value.type) {
    case 'A':
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   `Switch is not exhaustive. Cases not matched: "B"`,
						Line:      5,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type DiscriminatedUnion = { type: 'A'; a: 1 } | { type: 'B'; b: 2 };

function test(value: DiscriminatedUnion): number {
  switch (value.type) {
    case 'A':
      return 1;
    case "B": { throw new Error('Not implemented yet: "B" case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				// Still complains with empty switch
				Code: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;

switch (day) {
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   `Switch is not exhaustive. Cases not matched: "Monday" | "Tuesday" | "Wednesday" | "Thursday" | "Friday" | "Saturday" | "Sunday"`,
						Line:      13,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;

switch (day) {
case "Monday": { throw new Error('Not implemented yet: "Monday" case') }
case "Tuesday": { throw new Error('Not implemented yet: "Tuesday" case') }
case "Wednesday": { throw new Error('Not implemented yet: "Wednesday" case') }
case "Thursday": { throw new Error('Not implemented yet: "Thursday" case') }
case "Friday": { throw new Error('Not implemented yet: "Friday" case') }
case "Saturday": { throw new Error('Not implemented yet: "Saturday" case') }
case "Sunday": { throw new Error('Not implemented yet: "Sunday" case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
const a = Symbol('a');
const b = Symbol('b');
const c = Symbol('c');

type T = typeof a | typeof b | typeof c;

function test(value: T): number {
  switch (value) {
    case a:
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: typeof b | typeof c",
						Line:      9,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
const a = Symbol('a');
const b = Symbol('b');
const c = Symbol('c');

type T = typeof a | typeof b | typeof c;

function test(value: T): number {
  switch (value) {
    case a:
      return 1;
    case b: { throw new Error('Not implemented yet: b case') }
    case c: { throw new Error('Not implemented yet: c case') }
  }
}
`,
							},
						},
					},
				},
			},
			// Provides suggestions to add missing cases
			{
				// with existing cases present
				Code: `
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
    case 1:
      return 1;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
    case 1:
      return 1;
    case 2: { throw new Error('Not implemented yet: 2 case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				// without existing cases
				Code: `
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
  case 1: { throw new Error('Not implemented yet: 1 case') }
  case 2: { throw new Error('Not implemented yet: 2 case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				// keys include special characters
				Code: `
export enum Enum {
  'test-test' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
export enum Enum {
  'test-test' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  case Enum['test-test']: { throw new Error('Not implemented yet: Enum[\'test-test\'] case') }
  case Enum.test: { throw new Error('Not implemented yet: Enum.test case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				// keys include empty string
				Code: `
export enum Enum {
  '' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
export enum Enum {
  '' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  case Enum['']: { throw new Error('Not implemented yet: Enum[\'\'] case') }
  case Enum.test: { throw new Error('Not implemented yet: Enum.test case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				// keys include number as first character
				Code: `
export enum Enum {
  '9test' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
export enum Enum {
  '9test' = 'test-test',
  test = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  case Enum['9test']: { throw new Error('Not implemented yet: Enum[\'9test\'] case') }
  case Enum.test: { throw new Error('Not implemented yet: Enum.test case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
const value: number = Math.floor(Math.random() * 3);
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": true,
					"requireDefaultForNonUnion":           true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
const value: number = Math.floor(Math.random() * 3);
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default: { throw new Error('default case') }
}
`,
							},
						},
					},
				},
			},
			{
				// Enum member name containing literal backticks and escaped
				// double-quotes. Verified against v8.67.0's noFormat test case
				// via Node (see PR description) since the raw text can't be
				// authored directly inside a Go backtick string.
				Code: `
        enum Enum {
          'a' = 1,
          "'a' ` + "`" + `b` + "`" + ` \"c\"" = 2,
        }

        declare const a: Enum;

        switch (a) {}
      `,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
        enum Enum {
          'a' = 1,
          "'a' ` + "`" + `b` + "`" + ` \"c\"" = 2,
        }

        declare const a: Enum;

        switch (a) {
        case Enum.a: { throw new Error('Not implemented yet: Enum.a case') }
        case Enum['\'a\' ` + "`" + `b` + "`" + ` "c"']: { throw new Error('Not implemented yet: Enum[\'\\\'a\\\' ` + "`" + `b` + "`" + ` "c"\'] case') }
        }
      `,
							},
						},
					},
				},
			},
			{
				// superfluous switch with a string-based union
				Code: `
type MyUnion = 'foo' | 'bar' | 'baz';

declare const myUnion: MyUnion;

switch (myUnion) {
  case 'foo':
  case 'bar':
  case 'baz': {
    break;
  }
  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase", Line: 12, Column: 3, EndLine: 14, EndColumn: 4},
				},
			},
			{
				// superfluous switch with a string-based enum
				Code: `
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
  case MyEnum.Bar:
  case MyEnum.Baz: {
    break;
  }
  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				// superfluous switch with a number-based enum
				Code: `
enum MyEnum {
  Foo,
  Bar,
  Baz,
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
  case MyEnum.Bar:
  case MyEnum.Baz: {
    break;
  }
  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				// superfluous switch with a boolean
				Code: `
declare const myBoolean: boolean;

switch (myBoolean) {
  case true:
  case false: {
    break;
  }
  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				// superfluous switch with undefined
				Code: `
declare const myValue: undefined;

switch (myValue) {
  case undefined: {
    break;
  }

  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				// superfluous switch with null
				Code: `
declare const myValue: null;

switch (myValue) {
  case null: {
    break;
  }

  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				// superfluous switch with union of various types
				Code: `
declare const myValue: 'foo' | boolean | undefined | null;

switch (myValue) {
  case 'foo':
  case true:
  case false:
  case undefined:
  case null: {
    break;
  }

  default: {
    break;
  }
}
`,
				Options: map[string]any{
					"allowDefaultCaseForExhaustiveSwitch": false,
					"requireDefaultForNonUnion":           false,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
  default:
    break;
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  default:
    break;
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b';

switch (literal) {
  default:
  case 'a':
    break;
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b';

switch (literal) {
  case "b": { throw new Error('Not implemented yet: "b" case') }
  default:
  case 'a':
    break;
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  default:
    break;
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  case "c": { throw new Error('Not implemented yet: "c" case') }
  default:
    break;
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
    break;
  default: {
    break;
  }
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      10,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
    break;
  case MyEnum.Bar: { throw new Error('Not implemented yet: MyEnum.Bar case') }
  case MyEnum.Baz: { throw new Error('Not implemented yet: MyEnum.Baz case') }
  default: {
    break;
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const value: boolean;
switch (value) {
  default: {
    break;
  }
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const value: boolean;
switch (value) {
  case false: { throw new Error('Not implemented yet: false case') }
  case true: { throw new Error('Not implemented yet: true case') }
  default: {
    break;
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
function foo(x: string[]) {
  switch (x[0]) {
    case 'hi':
      break;
  }
}
`,
				TSConfig: "tsconfig.noUncheckedIndexedAccess.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      3,
						Column:    11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
function foo(x: string[]) {
  switch (x[0]) {
    case 'hi':
      break;
    case undefined: { throw new Error('Not implemented yet: undefined case') }
  }
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const myValue: 'a' | 'b';
switch (myValue) {
  case 'a':
    return 'a';
  case 'b':
    return 'b';
  // no default
}
`,
				Options: map[string]any{"allowDefaultCaseForExhaustiveSwitch": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "dangerousDefaultCase"},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  // no default
}
`,
				Options: map[string]any{"considerDefaultExhaustiveForUnions": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  case "c": { throw new Error('Not implemented yet: "c" case') }
  // no default
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  // skip default
}
`,
				Options: map[string]any{
					"considerDefaultExhaustiveForUnions": false,
					"defaultCaseCommentPattern":          `^skip\sdefault`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  case "b": { throw new Error('Not implemented yet: "b" case') }
  case "c": { throw new Error('Not implemented yet: "c" case') }
  // skip default
}
`,
							},
						},
					},
				},
			},
			{
				Code: `
export namespace A {
  export enum B {
    C,
    D,
  }
}
declare const foo: A.B;
switch (foo) {
  case A.B.C: {
    break;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: A.B.D",
						Line:      9,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
export namespace A {
  export enum B {
    C,
    D,
  }
}
declare const foo: A.B;
switch (foo) {
  case A.B.C: {
    break;
  }
  case A.B.D: { throw new Error('Not implemented yet: A.B.D case') }
}
`,
							},
						},
					},
				},
			},
			{
				FileName: "switch-exhaustiveness-check-import.ts",
				Code: `
import { A } from './switch-exhaustiveness-check';
declare const foo: A.B;
switch (foo) {
  case A.B.C: {
    break;
  }
}
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "switchIsNotExhaustive",
						Message:   "Switch is not exhaustive. Cases not matched: A.B.D",
						Line:      4,
						Column:    9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "addMissingCases",
								Output: `
import { A } from './switch-exhaustiveness-check';
declare const foo: A.B;
switch (foo) {
  case A.B.C: {
    break;
  }
  case A.B.D: { throw new Error('Not implemented yet: A.B.D case') }
}
`,
							},
						},
					},
				},
			},
		},
	)
}
