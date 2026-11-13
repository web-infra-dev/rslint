// Upstream: typescript-eslint v8.67.0, no-confusing-void-expression.test.ts.
// All upstream cases and outputs are preserved; full ranges come from the upstream parser.
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'input.ts',
  languageOptions: { parserOptions: { project: './tsconfig.json' } },
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: { strict: true, target: 'esnext', jsx: 'preserve', types: [] },
    }),
    'globals.d.ts': 'declare namespace JSX { interface IntrinsicElements { button: { onClick?: () => void } } }',
  },
});

ruleTester.run('no-confusing-void-expression', {
  valid: [
    '() => Math.random();',
    "console.log('foo');",
    'foo && console.log(foo);',
    'foo || console.log(foo);',
    'foo ? console.log(true) : console.log(false);',
    "console?.log('foo');",

    {
      code: `
() => console.log('foo');
      `,
      options: [{ ignoreArrowShorthand: true }],
    },
    {
      code: `
foo => foo && console.log(foo);
      `,
      options: [{ ignoreArrowShorthand: true }],
    },
    {
      code: `
foo => foo || console.log(foo);
      `,
      options: [{ ignoreArrowShorthand: true }],
    },
    {
      code: `
foo => (foo ? console.log(true) : console.log(false));
      `,
      options: [{ ignoreArrowShorthand: true }],
    },

    {
      code: `
!void console.log('foo');
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
+void (foo && console.log(foo));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
-void (foo || console.log(foo));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
() => void ((foo && void console.log(true)) || console.log(false));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
const x = void (foo ? console.log(true) : console.log(false));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
!(foo && void console.log(foo));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
!!(foo || void console.log(foo));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
const x = (foo && void console.log(true)) || void console.log(false);
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
() => (foo ? void console.log(true) : void console.log(false));
      `,
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
return void console.log('foo');
      `,
      options: [{ ignoreVoidOperator: true }],
    },

    `
function cool(input: string) {
  return (console.log(input), input);
}
    `,
    {
      code: `
function cool(input: string) {
  return (input, console.log(input), input);
}
      `,
    },
    {
      code: `
function test(): void {
  return console.log('bar');
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const test = (): void => {
  return console.log('bar');
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const test = (): void => console.log('bar');
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
function test(): void {
  {
    return console.log('foo');
  }
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const obj = {
  test(): void {
    return console.log('foo');
  },
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
class Foo {
  test(): void {
    return console.log('foo');
  }
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
function test() {
  function nestedTest(): void {
    return console.log('foo');
  }
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = () => void;
const test = (() => console.log()) as Foo;
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = {
  foo: () => void;
};
const test: Foo = {
  foo: () => console.log(),
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const test = {
  foo: () => console.log(),
} as {
  foo: () => void;
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const test: {
  foo: () => void;
} = {
  foo: () => console.log(),
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = {
  foo: { bar: () => void };
};

const test = {
  foo: { bar: () => console.log() },
} as Foo;
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = {
  foo: { bar: () => void };
};

const test: Foo = {
  foo: { bar: () => console.log() },
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type MethodType = () => void;

class App {
  private method: MethodType = () => console.log();
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
interface Foo {
  foo: () => void;
}

function bar(): Foo {
  return {
    foo: () => console.log(),
  };
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = () => () => () => void;
const x: Foo = () => () => () => console.log();
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = {
  foo: () => void;
};

const test = {
  foo: () => console.log(),
} as Foo;
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = () => void;
const test: Foo = () => console.log('foo');
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      filename: 'jsx.tsx',
      code: 'const foo = <button onClick={() => console.log()} />;',
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
declare function foo(arg: () => void): void;
foo(() => console.log());
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
declare function foo(arg: (() => void) | (() => string)): void;
foo(() => console.log());
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
declare function foo(arg: (() => void) | (() => string) | string): void;
foo(() => console.log());
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
declare function foo(arg: () => void | string): void;
foo(() => console.log());
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
declare function foo(options: { cb: () => void }): void;
foo({ cb: () => console.log() });
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const obj = {
  foo: { bar: () => console.log() },
} as {
  foo: { bar: () => void };
};
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
function test(): void & void {
  return console.log('foo');
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = void;

declare function foo(): Foo;

function test(): Foo {
  return foo();
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
type Foo = void;
const test = (): Foo => console.log('err');
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
const test: () => any = (): void => console.log();
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
function test(): void | string {
  return console.log('bar');
}
      `,
      options: [{ ignoreVoidReturningFunctions: true }],
    },
    {
      code: `
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
      options: [{ ignoreVoidReturningFunctions: true }],
    },
  ],

  invalid: [
    {
      code: `
const x = console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
    {
      code: `
const x = console?.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: `
console.error(console.log('foo'));
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 33,
        },
      ],
    },
    {
      code: `
[console.log('foo')];
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
({ x: console.log('foo') });
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
void console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 6,
          endLine: 2,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
console.log('foo') ? true : false;
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
(console.log('foo') && true) || false;
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
(cond && console.log('ok')) || console.log('error');
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      code: `
!console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },

    {
      code: `
function notcool(input: string) {
  return (input, console.log(input));
}
      `,
      errors: [
        {
          messageId: "invalidVoidExpr",
          message: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 36,
        },
      ],
    },
    {
      code: "() => console.log('foo');",
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 25,
        },
      ],
      output: `() => { console.log('foo'); };`,
    },
    {
      code: 'foo => foo && console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '(foo: undefined) => foo && console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 44,
        },
      ],
      output: `(foo: undefined) => { foo && console.log(foo); };`,
    },
    {
      code: 'foo => foo || console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '(foo: undefined) => foo || console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 44,
        },
      ],
      output: `(foo: undefined) => { foo || console.log(foo); };`,
    },
    {
      code: '(foo: void) => foo || console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 39,
        },
      ],
      output: `(foo: void) => { foo || console.log(foo); };`,
    },
    {
      code: 'foo => (foo ? console.log(true) : console.log(false));',
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 53,
        },
      ],
      output: `foo => { foo ? console.log(true) : console.log(false); };`,
    },
    {
      code: `
function f() {
  return console.log('foo');
  console.log('bar');
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturn",
          message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 28,
        },
      ],
      output: `
function f() {
  console.log('foo'); return;
  console.log('bar');
}
      `,
    },
    {
      code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
          console.log('quux')
        }
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturn",
          message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 53,
        },
      ],
      output: `
        function f() {
          console.log('foo')
          ;['bar', 'baz'].forEach(console.log); return;
          console.log('quux')
        }
      `,
    },
    {
      code: `
function f() {
  console.log('foo');
  return console.log('bar');
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 28,
        },
      ],
      output: `
function f() {
  console.log('foo');
  console.log('bar');
}
      `,
    },
    {
      code: `
        function f() {
          console.log('foo')
          return ['bar', 'baz'].forEach(console.log)
        }
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 53,
        },
      ],
      output: `
        function f() {
          console.log('foo')
          ;['bar', 'baz'].forEach(console.log);
        }
      `,
    },
    {
      code: `
const f = () => {
  if (cond) {
    return console.error('foo');
  }
  console.log('bar');
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturn",
          message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 32,
        },
      ],
      output: `
const f = () => {
  if (cond) {
    console.error('foo'); return;
  }
  console.log('bar');
};
      `,
    },
    {
      code: `
const f = function () {
  if (cond) return console.error('foo');
  console.log('bar');
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturn",
          message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 40,
        },
      ],
      output: `
const f = function () {
  if (cond) { console.error('foo'); return; }
  console.log('bar');
};
      `,
    },
    {
      code: `
const f = function () {
  let num = 1;
  return num ? console.log('foo') : num;
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 34,
        },
      ],
    },
    {
      code: `
const f = function () {
  let undef = undefined;
  return undef ? console.log('foo') : undef;
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 36,
        },
      ],
      output: `
const f = function () {
  let undef = undefined;
  undef ? console.log('foo') : undef;
};
      `,
    },
    {
      code: `
const f = function () {
  let num = 1;
  return num || console.log('foo');
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
const f = function () {
  let bar = void 0;
  return bar || console.log('foo');
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 35,
        },
      ],
      output: `
const f = function () {
  let bar = void 0;
  bar || console.log('foo');
};
      `,
    },
    {
      code: `
let num = 1;
const foo = () => (num ? console.log('foo') : num);
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 44,
        },
      ],
    },
    {
      code: `
let bar = void 0;
const foo = () => (bar ? console.log('foo') : bar);
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 26,
          endLine: 3,
          endColumn: 44,
        },
      ],
      output: `
let bar = void 0;
const foo = () => { bar ? console.log('foo') : bar; };
      `,
    },
    {
      code: "return console.log('foo');",
      errors: [
        {
          messageId: "invalidVoidExprReturnWrapVoid",
          message: "Void expressions returned from a function must be marked explicitly with the `void` operator.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 26,
        },
      ],
      options: [{ ignoreVoidOperator: true }],
      output: "return void console.log('foo');",
    },
    {
      code: "console.error(console.log('foo'));",
      errors: [
        {
          messageId: "invalidVoidExprWrapVoid",
          message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 33,
          suggestions: [
            {
              messageId: "voidExprWrapVoid",
              desc: "Mark with an explicit `void` operator.",
              output: "console.error(void console.log('foo'));",
            },
          ],
        },
      ],
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: "console.log('foo') ? true : false;",
      errors: [
        {
          messageId: "invalidVoidExprWrapVoid",
          message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
          suggestions: [
            {
              messageId: "voidExprWrapVoid",
              desc: "Mark with an explicit `void` operator.",
              output: "void console.log('foo') ? true : false;",
            },
          ],
        },
      ],
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: "const x = foo ?? console.log('foo');",
      errors: [
        {
          messageId: "invalidVoidExprWrapVoid",
          message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 36,
          suggestions: [
            {
              messageId: "voidExprWrapVoid",
              desc: "Mark with an explicit `void` operator.",
              output: "const x = foo ?? void console.log('foo');",
            },
          ],
        },
      ],
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: 'foo => foo || console.log(foo);',
      errors: [
        {
          messageId: "invalidVoidExprArrowWrapVoid",
          message: "Void expressions returned from an arrow function shorthand must be marked explicitly with the `void` operator.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 31,
        },
      ],
      options: [{ ignoreVoidOperator: true }],
      output: 'foo => foo || void console.log(foo);',
    },
    {
      code: "!!console.log('foo');",
      errors: [
        {
          messageId: "invalidVoidExprWrapVoid",
          message: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 21,
          suggestions: [
            {
              messageId: "voidExprWrapVoid",
              desc: "Mark with an explicit `void` operator.",
              output: "!!void console.log('foo');",
            },
          ],
        },
      ],
      options: [{ ignoreVoidOperator: true }],
    },
    {
      code: `
function test() {
  return console.log('foo');
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 28,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test() {
  console.log('foo');
}
      `,
    },
    {
      code: "const test = () => console.log('foo');",
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 38,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: "const test = () => { console.log('foo'); };",
    },
    {
      code: `
const test = () => {
  return console.log('foo');
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 28,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
const test = () => {
  console.log('foo');
};
      `,
    },
    {
      code: `
function foo(): void {
  const bar = () => {
    return console.log();
  };
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 25,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function foo(): void {
  const bar = () => {
    console.log();
  };
}
      `,
    },
    {
      code: `
(): any => console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 30,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
(): any => { console.log('foo'); };
      `,
    },
    {
      code: `
(): unknown => console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 2,
          column: 16,
          endLine: 2,
          endColumn: 34,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
(): unknown => { console.log('foo'); };
      `,
    },
    {
      code: `
function test(): void {
  () => () => console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 28,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(): void {
  () => () => { console.log(); };
}
      `,
    },
    {
      code: `
type Foo = any;
(): Foo => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 25,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = any;
(): Foo => { console.log(); };
      `,
    },
    {
      code: `
type Foo = unknown;
(): Foo => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 25,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = unknown;
(): Foo => { console.log(); };
      `,
    },
    {
      code: `
function test(): any {
  () => () => console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 28,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(): any {
  () => () => { console.log(); };
}
      `,
    },
    {
      code: `
function test(): unknown {
  return console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 23,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(): unknown {
  console.log();
}
      `,
    },
    {
      code: `
function test(): any {
  return console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 23,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(): any {
  console.log();
}
      `,
    },
    {
      code: `
type Foo = () => any;
(): Foo => () => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = () => any;
(): Foo => () => { console.log(); };
      `,
    },
    {
      code: `
type Foo = () => unknown;
(): Foo => () => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 18,
          endLine: 3,
          endColumn: 31,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = () => unknown;
(): Foo => () => { console.log(); };
      `,
    },
    {
      code: `
type Foo = () => any;
const test: Foo = () => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 38,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = () => any;
const test: Foo = () => { console.log(); };
      `,
    },
    {
      code: `
type Foo = () => unknown;
const test: Foo = () => console.log();
      `,
      errors: [
        {
          messageId: "invalidVoidExprArrow",
          message: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
          line: 3,
          column: 25,
          endLine: 3,
          endColumn: 38,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = () => unknown;
const test: Foo = () => { console.log(); };
      `,
    },
    {
      code: `
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 6,
          column: 12,
          endLine: 6,
          endColumn: 25,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
type Foo = () => void;

const foo: Foo = function () {
  function bar() {
    console.log();
  }
};
      `,
    },
    {
      code: `
const foo = function () {
  function bar() {
    return console.log();
  }
};
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 25,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
const foo = function () {
  function bar() {
    console.log();
  }
};
      `,
    },
    {
      code: `
return console.log('foo');
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturn",
          message: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 26,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
{ console.log('foo'); return; }
      `,
    },
    {
      code: `
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 8,
          column: 10,
          endLine: 8,
          endColumn: 23,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(): void;
function test(arg: string): any;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  console.log();
}
      `,
    },
    {
      code: `
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  return console.log();
}
      `,
      errors: [
        {
          messageId: "invalidVoidExprReturnLast",
          message: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
          line: 8,
          column: 10,
          endLine: 8,
          endColumn: 23,
        },
      ],
      options: [{ ignoreVoidReturningFunctions: true }],
      output: `
function test(arg: string): any;
function test(): void;
function test(arg?: string): any | void {
  if (arg) {
    return arg;
  }
  console.log();
}
      `,
    },
  ],
});
