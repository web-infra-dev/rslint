// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = (code: string) => ({ code, filename });

const fixed = (code: string, output: string) => ({
  code,
  filename,
  output,
  errors: [
    {
      messageId: 'no-typeof-undefined/error',
      message: 'Compare with `undefined` directly instead of using `typeof`.',
      suggestions: [],
    },
  ],
});

const globalSuggestion = (
  code: string,
  output: string,
  operator: '===' | '!==',
) => ({
  code,
  filename,
  options: [{ checkGlobalVariables: true }],
  errors: [
    {
      messageId: 'no-typeof-undefined/error',
      message: 'Compare with `undefined` directly instead of using `typeof`.',
      suggestions: [
        {
          messageId: 'no-typeof-undefined/suggestion',
          desc: 'Switch to `… ' + operator + ' undefined`.',
          output,
        },
      ],
    },
  ],
});

ruleTester.run('no-typeof-undefined', null as never, {
  valid: [
    valid('typeof a.b'),
    valid('typeof a.b > "undefined"'),
    valid('a.b === "undefined"'),
    valid('void a.b === "undefined"'),
    valid('+a.b === "undefined"'),
    valid('++a.b === "undefined"'),
    valid('a.b++ === "undefined"'),
    valid('foo === undefined'),
    valid('typeof a.b === "string"'),
    valid('typeof foo === "undefined"'),
    valid('foo = 2; typeof foo === "undefined"'),
    valid('/* globals foo: readonly */ typeof foo === "undefined"'),
    valid(
      '/* globals globalThis: readonly */ typeof globalThis === "undefined"',
    ),
    valid("function parse() {\n\tswitch (typeof value === 'undefined') {}\n}"),
    valid(
      "/* globals value: readonly */\nfunction parse() {\n\tswitch (typeof value === 'undefined') {}\n}",
    ),
    valid('"undefined" === typeof a.b'),
    valid('const UNDEFINED = "undefined"; typeof a.b === UNDEFINED'),
    valid('typeof a.b === `undefined`'),
  ],
  invalid: [
    fixed('typeof a.b === "undefined"', 'a.b === undefined'),
    fixed('typeof a.b !== "undefined"', 'a.b !== undefined'),
    fixed('typeof a.b == "undefined"', 'a.b === undefined'),
    fixed('typeof a.b != "undefined"', 'a.b !== undefined'),
    fixed("typeof a.b == 'undefined'", 'a.b === undefined'),
    fixed('let foo; typeof foo === "undefined"', 'let foo; foo === undefined'),
    fixed(
      'const foo = 1; typeof foo === "undefined"',
      'const foo = 1; foo === undefined',
    ),
    fixed('var foo; typeof foo === "undefined"', 'var foo; foo === undefined'),
    fixed(
      'var foo; var foo; typeof foo === "undefined"',
      'var foo; var foo; foo === undefined',
    ),
    fixed(
      'for (const foo of bar) typeof foo === "undefined";',
      'for (const foo of bar) foo === undefined;',
    ),
    fixed(
      'let foo;\nfunction bar() {\n\ttypeof foo === "undefined";\n}',
      'let foo;\nfunction bar() {\n\tfoo === undefined;\n}',
    ),
    fixed(
      'function foo() {typeof foo === "undefined"}',
      'function foo() {foo === undefined}',
    ),
    fixed(
      'function foo(bar) {typeof bar === "undefined"}',
      'function foo(bar) {bar === undefined}',
    ),
    fixed(
      'function foo({bar}) {typeof bar === "undefined"}',
      'function foo({bar}) {bar === undefined}',
    ),
    fixed(
      'function foo([bar]) {typeof bar === "undefined"}',
      'function foo([bar]) {bar === undefined}',
    ),
    fixed('typeof foo.bar === "undefined"', 'foo.bar === undefined'),
    fixed(
      'import foo from \'foo\';\ntypeof foo.bar === "undefined"',
      "import foo from 'foo';\nfoo.bar === undefined",
    ),
    fixed('foo\ntypeof [] === "undefined";', 'foo\n;[] === undefined;'),
    fixed(
      'foo\ntypeof (a ? b : c) === "undefined";',
      'foo\n;(a ? b : c) === undefined;',
    ),
    fixed(
      "function a() {\n\treturn typeof // comment\n\t\ta.b === 'undefined';\n}",
      'function a() {\n\treturn ( // comment\n\t\ta.b === undefined);\n}',
    ),
    fixed(
      "function a() {\n\treturn (typeof // ReturnStatement argument is parenthesized\n\t\ta.b === 'undefined');\n}",
      'function a() {\n\treturn (// ReturnStatement argument is parenthesized\n\t\ta.b === undefined);\n}',
    ),
    fixed(
      "function a() {\n\treturn (typeof // UnaryExpression is parenthesized\n\t\ta.b) === 'undefined';\n}",
      'function a() {\n\treturn (// UnaryExpression is parenthesized\n\t\ta.b) === undefined;\n}',
    ),
    fixed(
      "function parse(value) {\n\tswitch (typeof value === 'undefined') {}\n}",
      'function parse(value) {\n\tswitch (value === undefined) {}\n}',
    ),
    globalSuggestion(
      'typeof undefinedVariableIdentifier === "undefined"',
      'undefinedVariableIdentifier === undefined',
      '===',
    ),
    globalSuggestion(
      'typeof Array !== "undefined"',
      'Array !== undefined',
      '!==',
    ),
    globalSuggestion(
      "function parse() {\n\tswitch (typeof value === 'undefined') {}\n}",
      'function parse() {\n\tswitch (value === undefined) {}\n}',
      '===',
    ),
    globalSuggestion(
      "/* globals value: readonly */\nfunction parse() {\n\tswitch (typeof value === 'undefined') {}\n}",
      '/* globals value: readonly */\nfunction parse() {\n\tswitch (value === undefined) {}\n}',
      '===',
    ),
  ],
});
