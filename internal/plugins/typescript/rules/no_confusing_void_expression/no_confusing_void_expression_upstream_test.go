package no_confusing_void_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All 53 valid and 57 invalid cases from typescript-eslint v8.67.0.
// Messages and full ranges are checked in addition to upstream outputs and suggestions.
func TestNoConfusingVoidExpressionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingVoidExpressionRule, []rule_tester.ValidTestCase{
		{
			Code: "() => Math.random();",
		},
		{
			Code: "console.log('foo');",
		},
		{
			Code: "foo && console.log(foo);",
		},
		{
			Code: "foo || console.log(foo);",
		},
		{
			Code: "foo ? console.log(true) : console.log(false);",
		},
		{
			Code: "console?.log('foo');",
		},
		{
			Code: `
() => console.log('foo');
      `,
			Options: map[string]any{"ignoreArrowShorthand": true},
		},
		{
			Code: `
foo => foo && console.log(foo);
      `,
			Options: map[string]any{"ignoreArrowShorthand": true},
		},
		{
			Code: `
foo => foo || console.log(foo);
      `,
			Options: map[string]any{"ignoreArrowShorthand": true},
		},
		{
			Code: `
foo => (foo ? console.log(true) : console.log(false));
      `,
			Options: map[string]any{"ignoreArrowShorthand": true},
		},
		{
			Code: `
!void console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
+void (foo && console.log(foo));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
-void (foo || console.log(foo));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
() => void ((foo && void console.log(true)) || console.log(false));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
const x = void (foo ? console.log(true) : console.log(false));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
!(foo && void console.log(foo));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
!!(foo || void console.log(foo));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
const x = (foo && void console.log(true)) || void console.log(false);
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
() => (foo ? void console.log(true) : void console.log(false));
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
return void console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidOperator": true},
		},
		{
			Code: `
function cool(input: string) {
  return (console.log(input), input);
}
    `,
		},
		{
			Code: `
function cool(input: string) {
  return (input, console.log(input), input);
}
      `,
		},
		{
			Code: `
function test(): void {
  return console.log('bar');
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const test = (): void => {
  return console.log('bar');
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const test = (): void => console.log('bar');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
function test(): void {
  {
    return console.log('foo');
  }
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const obj = {
  test(): void {
    return console.log('foo');
  },
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
class Foo {
  test(): void {
    return console.log('foo');
  }
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
function test() {
  function nestedTest(): void {
    return console.log('foo');
  }
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = () => void;
const test = (() => console.log()) as Foo;
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = {
  foo: () => void;
};
const test: Foo = {
  foo: () => console.log(),
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const test = {
  foo: () => console.log(),
} as {
  foo: () => void;
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const test: {
  foo: () => void;
} = {
  foo: () => console.log(),
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = {
  foo: { bar: () => void };
};

const test = {
  foo: { bar: () => console.log() },
} as Foo;
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = {
  foo: { bar: () => void };
};

const test: Foo = {
  foo: { bar: () => console.log() },
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type MethodType = () => void;

class App {
  private method: MethodType = () => console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
interface Foo {
  foo: () => void;
}

function bar(): Foo {
  return {
    foo: () => console.log(),
  };
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = () => () => () => void;
const x: Foo = () => () => () => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = {
  foo: () => void;
};

const test = {
  foo: () => console.log(),
} as Foo;
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = () => void;
const test: Foo = () => console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code:    "const foo = <button onClick={() => console.log()} />;",
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Tsx:     true,
		},
		{
			Code: `
declare function foo(arg: () => void): void;
foo(() => console.log());
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
declare function foo(arg: (() => void) | (() => string)): void;
foo(() => console.log());
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
declare function foo(arg: (() => void) | (() => string) | string): void;
foo(() => console.log());
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
declare function foo(arg: () => void | string): void;
foo(() => console.log());
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
declare function foo(options: { cb: () => void }): void;
foo({ cb: () => console.log() });
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const obj = {
  foo: { bar: () => console.log() },
} as {
  foo: { bar: () => void };
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
function test(): void & void {
  return console.log('foo');
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = void;

declare function foo(): Foo;

function test(): Foo {
  return foo();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
type Foo = void;
const test = (): Foo => console.log('err');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
const test: () => any = (): void => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
function test(): void | string {
  return console.log('bar');
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		},
		{
			Code: `
export function makeDate(): Date;
export function makeDate(m: number): void;
export function makeDate(m?: number): Date | void {
  if (m !== undefined) {
    return console.log('123');
  }
  return new Date();
}

declare const test: (cb: () => void) => void;

test((() => {
  return console.log('123');
}) as typeof makeDate | (() => string));
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
		}}, []rule_tester.InvalidTestCase{
		{
			Code: `
const x = console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 29,
				},
			},
		},
		{
			Code: `
const x = console?.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    11,
					EndLine:   2,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
console.error(console.log('foo'));
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    15,
					EndLine:   2,
					EndColumn: 33,
				},
			},
		},
		{
			Code: `
[console.log('foo')];
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    2,
					EndLine:   2,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
({ x: console.log('foo') });
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    7,
					EndLine:   2,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
void console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    6,
					EndLine:   2,
					EndColumn: 24,
				},
			},
		},
		{
			Code: `
console.log('foo') ? true : false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    1,
					EndLine:   2,
					EndColumn: 19,
				},
			},
		},
		{
			Code: `
(console.log('foo') && true) || false;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    2,
					EndLine:   2,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
(cond && console.log('ok')) || console.log('error');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    10,
					EndLine:   2,
					EndColumn: 27,
				},
			},
		},
		{
			Code: `
!console.log('foo');
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      2,
					Column:    2,
					EndLine:   2,
					EndColumn: 20,
				},
			},
		},
		{
			Code: `
function notcool(input: string) {
  return (input, console.log(input));
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExpr",
					Message:   "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 36,
				},
			},
		},
		{
			Code:   "() => console.log('foo');",
			Output: []string{"() => { console.log('foo'); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    7,
					EndLine:   1,
					EndColumn: 25,
				},
			},
		},
		{
			Code: "foo => foo && console.log(foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 31,
				},
			},
		},
		{
			Code:   "(foo: undefined) => foo && console.log(foo);",
			Output: []string{"(foo: undefined) => { foo && console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 44,
				},
			},
		},
		{
			Code: "foo => foo || console.log(foo);",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 31,
				},
			},
		},
		{
			Code:   "(foo: undefined) => foo || console.log(foo);",
			Output: []string{"(foo: undefined) => { foo || console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    28,
					EndLine:   1,
					EndColumn: 44,
				},
			},
		},
		{
			Code:   "(foo: void) => foo || console.log(foo);",
			Output: []string{"(foo: void) => { foo || console.log(foo); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 39,
				},
			},
		},
		{
			Code:   "foo => (foo ? console.log(true) : console.log(false));",
			Output: []string{"foo => { foo ? console.log(true) : console.log(false); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 32,
				},
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    35,
					EndLine:   1,
					EndColumn: 53,
				},
			},
		},
		{
			Code: `
function f() {
  return console.log('foo');
  console.log('bar');
}
      `,
			Output: []string{`
function f() {
  console.log('foo'); return;
  console.log('bar');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Message:   "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
          console.log('quux')
        }
      `,
			Output: []string{`
        function f() {
          console.log('foo')
          ;['bar', 'baz'].forEach(console.log); return;
          console.log('quux')
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Message:   "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
					Line:      4,
					Column:    18,
					EndLine:   4,
					EndColumn: 53,
				},
			},
		},
		{
			Code: `
function f() {
  console.log('foo');
  return console.log('bar');
}
      `,
			Output: []string{`
function f() {
  console.log('foo');
  console.log('bar');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    10,
					EndLine:   4,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
        }
      `,
			Output: []string{`
        function f() {
          console.log('foo')
          ;['bar', 'baz'].forEach(console.log);
        }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    18,
					EndLine:   4,
					EndColumn: 53,
				},
			},
		},
		{
			Code: `
const f = () => {
  if (cond) {
    return console.error('foo');
  }
  console.log('bar');
};
      `,
			Output: []string{`
const f = () => {
  if (cond) {
    console.error('foo'); return;
  }
  console.log('bar');
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Message:   "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 32,
				},
			},
		},
		{
			Code: `
const f = function () {
  if (cond) return console.error('foo');
  console.log('bar');
};
      `,
			Output: []string{`
const f = function () {
  if (cond) { console.error('foo'); return; }
  console.log('bar');
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Message:   "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
					Line:      3,
					Column:    20,
					EndLine:   3,
					EndColumn: 40,
				},
			},
		},
		{
			Code: `
const f = function () {
  let num = 1;
  return num ? console.log('foo') : num;
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    16,
					EndLine:   4,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
const f = function () {
  let undef = undefined;
  return undef ? console.log('foo') : undef;
};
      `,
			Output: []string{`
const f = function () {
  let undef = undefined;
  undef ? console.log('foo') : undef;
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    18,
					EndLine:   4,
					EndColumn: 36,
				},
			},
		},
		{
			Code: `
const f = function () {
  let num = 1;
  return num || console.log('foo');
};
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
const f = function () {
  let bar = void 0;
  return bar || console.log('foo');
};
      `,
			Output: []string{`
const f = function () {
  let bar = void 0;
  bar || console.log('foo');
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    17,
					EndLine:   4,
					EndColumn: 35,
				},
			},
		},
		{
			Code: `
let num = 1;
const foo = () => (num ? console.log('foo') : num);
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    26,
					EndLine:   3,
					EndColumn: 44,
				},
			},
		},
		{
			Code: `
let bar = void 0;
const foo = () => (bar ? console.log('foo') : bar);
      `,
			Output: []string{`
let bar = void 0;
const foo = () => { bar ? console.log('foo') : bar; };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    26,
					EndLine:   3,
					EndColumn: 44,
				},
			},
		},
		{
			Code:    "return console.log('foo');",
			Options: map[string]any{"ignoreVoidOperator": true},
			Output:  []string{"return void console.log('foo');"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnWrapVoid",
					Message:   "Void expressions returned from a function must be marked explicitly with the `void` operator.",
					Line:      1,
					Column:    8,
					EndLine:   1,
					EndColumn: 26,
				},
			},
		},
		{
			Code:    "console.error(console.log('foo'));",
			Options: map[string]any{"ignoreVoidOperator": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Message:   "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "voidExprWrapVoid", Output: "console.error(void console.log('foo'));"},
					},
				},
			},
		},
		{
			Code:    "console.log('foo') ? true : false;",
			Options: map[string]any{"ignoreVoidOperator": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Message:   "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "voidExprWrapVoid", Output: "void console.log('foo') ? true : false;"},
					},
				},
			},
		},
		{
			Code:    "const x = foo ?? console.log('foo');",
			Options: map[string]any{"ignoreVoidOperator": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Message:   "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "voidExprWrapVoid", Output: "const x = foo ?? void console.log('foo');"},
					},
				},
			},
		},
		{
			Code:    "foo => foo || console.log(foo);",
			Options: map[string]any{"ignoreVoidOperator": true},
			Output:  []string{"foo => foo || void console.log(foo);"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrowWrapVoid",
					Message:   "Void expressions returned from an arrow function shorthand must be marked explicitly with the `void` operator.",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 31,
				},
			},
		},
		{
			Code:    "!!console.log('foo');",
			Options: map[string]any{"ignoreVoidOperator": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprWrapVoid",
					Message:   "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
					Line:      1,
					Column:    3,
					EndLine:   1,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "voidExprWrapVoid", Output: "!!void console.log('foo');"},
					},
				},
			},
		},
		{
			Code: `
function test() {
  return console.log('foo');
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test() {
  console.log('foo');
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code:    "const test = () => console.log('foo');",
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output:  []string{"const test = () => { console.log('foo'); };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      1,
					Column:    20,
					EndLine:   1,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `
const test = () => {
  return console.log('foo');
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
const test = () => {
  console.log('foo');
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
function foo(): void {
  const bar = () => {
    return console.log();
  };
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function foo(): void {
  const bar = () => {
    console.log();
  };
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
(): any => console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
(): any => { console.log('foo'); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      2,
					Column:    12,
					EndLine:   2,
					EndColumn: 30,
				},
			},
		},
		{
			Code: `
(): unknown => console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
(): unknown => { console.log('foo'); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      2,
					Column:    16,
					EndLine:   2,
					EndColumn: 34,
				},
			},
		},
		{
			Code: `
function test(): void {
  () => () => console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(): void {
  () => () => { console.log(); };
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    15,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
type Foo = any;
(): Foo => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = any;
(): Foo => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    12,
					EndLine:   3,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
type Foo = unknown;
(): Foo => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = unknown;
(): Foo => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    12,
					EndLine:   3,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
function test(): any {
  () => () => console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(): any {
  () => () => { console.log(); };
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    15,
					EndLine:   3,
					EndColumn: 28,
				},
			},
		},
		{
			Code: `
function test(): unknown {
  return console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(): unknown {
  console.log();
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
function test(): any {
  return console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(): any {
  console.log();
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      3,
					Column:    10,
					EndLine:   3,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
type Foo = () => any;
(): Foo => () => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = () => any;
(): Foo => () => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
type Foo = () => unknown;
(): Foo => () => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = () => unknown;
(): Foo => () => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    18,
					EndLine:   3,
					EndColumn: 31,
				},
			},
		},
		{
			Code: `
type Foo = () => any;
const test: Foo = () => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = () => any;
const test: Foo = () => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    25,
					EndLine:   3,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `
type Foo = () => unknown;
const test: Foo = () => console.log();
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = () => unknown;
const test: Foo = () => { console.log(); };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprArrow",
					Message:   "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
					Line:      3,
					Column:    25,
					EndLine:   3,
					EndColumn: 38,
				},
			},
		},
		{
			Code: `
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
    console.log();
  }
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      6,
					Column:    12,
					EndLine:   6,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
const foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
const foo = function () {
  function bar() {
    console.log();
  }
};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      4,
					Column:    12,
					EndLine:   4,
					EndColumn: 25,
				},
			},
		},
		{
			Code: `
return console.log('foo');
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
{ console.log('foo'); return; }
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturn",
					Message:   "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
					Line:      2,
					Column:    8,
					EndLine:   2,
					EndColumn: 26,
				},
			},
		},
		{
			Code: `
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  console.log();
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      8,
					Column:    10,
					EndLine:   8,
					EndColumn: 23,
				},
			},
		},
		{
			Code: `
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
			Options: map[string]any{"ignoreVoidReturningFunctions": true},
			Output: []string{`
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  console.log();
}
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidVoidExprReturnLast",
					Message:   "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
					Line:      8,
					Column:    10,
					EndLine:   8,
					EndColumn: 23,
				},
			},
		}})
}
