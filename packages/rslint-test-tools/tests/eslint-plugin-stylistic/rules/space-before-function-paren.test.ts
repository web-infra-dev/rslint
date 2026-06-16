/**
 * @fileoverview Tests for space-before-function-paren.
 * @author Mathias Schreck <https://github.com/lo1tuma>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/space-before-function-paren/space-before-function-paren._js_.test.ts
 *   packages/eslint-plugin/rules/space-before-function-paren/space-before-function-paren._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('space-before-function-paren', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - `parserOptions` (ecmaVersion) dropped — rslint always parses at esnext.
 *  - The rule's messageIds (`missingSpace` / `unexpectedSpace`) take no `data`, so
 *    they map 1:1 to a fixed message; the RuleTester renders them from the plugin's
 *    meta. The two cases that pin a literal `message` instead of a `messageId` are
 *    carried verbatim as `message`.
 *
 * The `._js_` file ends with an `if (!skipBabel)` block holding a single Flow
 * valid case (`type TransformFunction = (el, code) => string;`). That fixture is
 * byte-identical to the first valid case of the `._ts_` suite and is plain, legal
 * TypeScript (a type alias whose value is a function type) — it parses cleanly
 * under ts-go and behaves identically, so it is kept in the green `valid` set
 * below rather than isolated. (It therefore appears once from `._ts_` and once
 * from the Babel/Flow block.)
 *
 * KNOWN GAPS: none. Every upstream fixture parses under rslint's ts-go parser and
 * produces byte-identical diagnostics and autofix output. No octal/`\8` syntax,
 * no `assert` import attributes, no sloppy-mode-only fixtures, and no output-only
 * invalid cases exist for this rule, so nothing was isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('space-before-function-paren', null as never, {
  valid: [
    // ---- from space-before-function-paren._js_.test.ts ----
    'function foo () {}',
    'var foo = function () {}',
    'var bar = function foo () {}',
    'var bar = function foo/**/ () {}',
    'var bar = function foo /**/() {}',
    'var bar = function foo/**/\n() {}',
    'var bar = function foo\n/**/() {}',
    'var bar = function foo//\n() {}',
    'var obj = { get foo () {}, set foo (val) {} };',
    {
      code: 'var obj = { foo () {} };',
    },
    { code: 'function* foo () {}' },
    { code: 'var foo = function *() {};' },

    { code: 'function foo() {}', options: ['never'] },
    { code: 'var foo = function() {}', options: ['never'] },
    { code: 'var foo = function/**/() {}', options: ['never'] },
    { code: 'var foo = function/* */() {}', options: ['never'] },
    { code: 'var foo = function/* *//*  */() {}', options: ['never'] },
    { code: 'var bar = function foo() {}', options: ['never'] },
    { code: 'var obj = { get foo() {}, set foo(val) {} };', options: ['never'] },
    {
      code: 'var obj = { foo() {} };',
      options: ['never'],
    },
    {
      code: 'function* foo() {}',
      options: ['never'],
    },
    {
      code: 'var foo = function*() {};',
      options: ['never'],
    },

    {
      code:
        'function foo() {}\n' +
        'var bar = function () {}\n' +
        'function* baz() {}\n' +
        'var bat = function*() {};\n' +
        'var obj = { get foo() {}, set foo(val) {}, bar() {} };',
      options: [{ named: 'never', anonymous: 'always' }],
    },
    {
      code:
        'function foo () {}\n' +
        'var bar = function() {}\n' +
        'function* baz () {}\n' +
        'var bat = function* () {};\n' +
        'var obj = { get foo () {}, set foo (val) {}, bar () {} };',
      options: [{ named: 'always', anonymous: 'never' }],
    },
    {
      code: 'class Foo { constructor() {} *method() {} }',
      options: [{ named: 'never', anonymous: 'always' }],
    },
    {
      code: 'class Foo { constructor () {} *method () {} }',
      options: [{ named: 'always', anonymous: 'never' }],
    },
    {
      code: 'var foo = function() {}',
      options: [{ named: 'always', anonymous: 'ignore' }],
    },
    {
      code: 'var foo = function () {}',
      options: [{ named: 'always', anonymous: 'ignore' }],
    },
    {
      code: 'var bar = function foo() {}',
      options: [{ named: 'ignore', anonymous: 'always' }],
    },
    {
      code: 'var bar = function foo () {}',
      options: [{ named: 'ignore', anonymous: 'always' }],
    },

    // Async arrow functions
    { code: '() => 1' },
    { code: 'async a => a' },
    { code: 'async a => a', options: [{ asyncArrow: 'always' }] },
    { code: 'async a => a', options: [{ asyncArrow: 'never' }] },
    { code: 'async () => 1', options: [{ asyncArrow: 'always' }] },
    { code: 'async() => 1', options: [{ asyncArrow: 'never' }] },
    { code: 'async () => 1', options: [{ asyncArrow: 'ignore' }] },
    { code: 'async() => 1', options: [{ asyncArrow: 'ignore' }] },
    { code: 'async () => 1' },
    { code: 'async () => 1', options: ['always'] },
    { code: 'async() => 1', options: ['never'] },

    // Catch clause
    { code: 'try {} catch (e) {}' },
    { code: 'try {} catch (e) {}', options: ['always'] },
    { code: 'try {} catch(e) {}', options: ['never'] },
    { code: 'try {} catch (e) {}', options: [{ catch: 'always' }] },
    { code: 'try {} catch(e) {}', options: [{ catch: 'never' }] },
    { code: 'try {} catch (e) {}', options: [{ catch: 'ignore' }] },
    { code: 'try {} catch(e) {}', options: [{ catch: 'ignore' }] },

    // ---- from space-before-function-paren._ts_.test.ts ----
    'type TransformFunction = (el: ASTElement, code: string) => string;',
    'var f = function <T> () {};',
    'function foo<T extends () => {}> () {}',
    'async <T extends () => {}> () => {}',
    'async <T>() => {}',
    {
      code: 'function foo<T extends Record<string, () => {}>>() {}',
      options: ['never'],
    },

    'abstract class Foo { constructor () {} abstract method () }',
    {
      code: 'abstract class Foo { constructor() {} abstract method() }',
      options: ['never'],
    },
    {
      code: 'abstract class Foo { constructor() {} abstract method() }',
      options: [{ anonymous: 'always', named: 'never' }],
    },
    'function foo ();',
    {
      code: 'function foo();',
      options: ['never'],
    },
    {
      code: 'function foo();',
      options: [{ anonymous: 'always', named: 'never' }],
    },

    // ---- from the `._js_` `if (!skipBabel)` Babel/Flow block ----
    // Plain, legal TypeScript (a function-type alias) — parses identically under
    // ts-go, so kept green rather than gapped.
    'type TransformFunction = (el: ASTElement, code: string) => string;',
  ],

  invalid: [
    // ---- from space-before-function-paren._js_.test.ts ----
    {
      code: 'function foo() {}',
      output: 'function foo () {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 13,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'function foo/* */() {}',
      output: 'function foo /* */() {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 18,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'var foo = function() {}',
      output: 'var foo = function () {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 19,
        },
      ],
    },
    {
      code: 'var bar = function foo() {}',
      output: 'var bar = function foo () {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'var obj = { get foo() {}, set foo(val) {} };',
      output: 'var obj = { get foo () {}, set foo (val) {} };',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 20,
        },
        {
          messageId: 'missingSpace',
          line: 1,
          column: 34,
        },
      ],
    },
    {
      code: 'var obj = { foo() {} };',
      output: 'var obj = { foo () {} };',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 16,
        },
      ],
    },
    {
      code: 'function* foo() {}',
      output: 'function* foo () {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 14,
        },
      ],
    },

    {
      code: 'function foo () {}',
      output: 'function foo() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'function foo /* */ () {}',
      output: 'function foo/* */() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'function foo/* block comment */ () {}',
      output: 'function foo/* block comment */() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'function foo/* 1 */ /* 2 */ \n /* 3 */\n/* 4 */ () {}',
      output: 'function foo/* 1 *//* 2 *//* 3 *//* 4 */() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'function foo  () {}',
      output: 'function foo() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'function foo//\n() {}',
      output: null,
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'function foo // line comment \n () {}',
      output: null,
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'function foo\n//\n() {}',
      output: null,
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
      ],
    },
    {
      code: 'var foo = function () {}',
      output: 'var foo = function() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 19,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'var bar = function foo () {}',
      output: 'var bar = function foo() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'var obj = { get foo () {}, set foo (val) {} };',
      output: 'var obj = { get foo() {}, set foo(val) {} };',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 20,
        },
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 35,
        },
      ],
    },
    {
      code: 'var obj = { foo () {} };',
      output: 'var obj = { foo() {} };',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 16,
        },
      ],
    },
    {
      code: 'function* foo () {}',
      output: 'function* foo() {}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 14,
        },
      ],
    },

    {
      code:
        'function foo () {}\n' +
        'var bar = function() {}\n' +
        'var obj = { get foo () {}, set foo (val) {}, bar () {} };',
      output:
        'function foo() {}\n' +
        'var bar = function () {}\n' +
        'var obj = { get foo() {}, set foo(val) {}, bar() {} };',
      options: [{ named: 'never', anonymous: 'always' }],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 13,
        },
        {
          messageId: 'missingSpace',
          line: 2,
          column: 19,
        },
        {
          messageId: 'unexpectedSpace',
          line: 3,
          column: 20,
        },
        {
          messageId: 'unexpectedSpace',
          line: 3,
          column: 35,
        },
        {
          messageId: 'unexpectedSpace',
          line: 3,
          column: 49,
        },
      ],
    },
    {
      code: 'class Foo { constructor () {} *method () {} }',
      output: 'class Foo { constructor() {} *method() {} }',
      options: [{ named: 'never', anonymous: 'always' }],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 24,
        },
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 38,
        },
      ],
    },
    {
      code: 'var foo = { bar () {} }',
      output: 'var foo = { bar() {} }',
      options: [{ named: 'never', anonymous: 'always' }],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 16,
        },
      ],
    },
    {
      code:
        'function foo() {}\n' +
        'var bar = function () {}\n' +
        'var obj = { get foo() {}, set foo(val) {}, bar() {} };',
      output:
        'function foo () {}\n' +
        'var bar = function() {}\n' +
        'var obj = { get foo () {}, set foo (val) {}, bar () {} };',
      options: [{ named: 'always', anonymous: 'never' }],
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 13,
        },
        {
          messageId: 'unexpectedSpace',
          line: 2,
          column: 19,
        },
        {
          messageId: 'missingSpace',
          line: 3,
          column: 20,
        },
        {
          messageId: 'missingSpace',
          line: 3,
          column: 34,
        },
        {
          messageId: 'missingSpace',
          line: 3,
          column: 47,
        },
      ],
    },
    {
      code: 'var foo = function() {}',
      output: 'var foo = function () {}',
      options: [{ named: 'ignore', anonymous: 'always' }],
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 19,
        },
      ],
    },
    {
      code: 'var foo = function () {}',
      output: 'var foo = function() {}',
      options: [{ named: 'ignore', anonymous: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 19,
        },
      ],
    },
    {
      code: 'var bar = function foo() {}',
      output: 'var bar = function foo () {}',
      options: [{ named: 'always', anonymous: 'ignore' }],
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'var bar = function foo () {}',
      output: 'var bar = function foo() {}',
      options: [{ named: 'never', anonymous: 'ignore' }],
      errors: [
        {
          messageId: 'unexpectedSpace',
          line: 1,
          column: 23,
        },
      ],
    },

    // Async arrow functions
    {
      code: 'async() => 1',
      output: 'async () => 1',
      options: [{ asyncArrow: 'always' }],
      errors: [{ message: 'Missing space before function parentheses.' }],
    },
    {
      code: 'async () => 1',
      output: 'async() => 1',
      options: [{ asyncArrow: 'never' }],
      errors: [{ message: 'Unexpected space before function parentheses.' }],
    },
    {
      code: 'async() => 1',
      output: 'async () => 1',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'async() => 1',
      output: 'async () => 1',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'async () => 1',
      output: 'async() => 1',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },

    // Catch clause
    {
      code: 'try {} catch(e) {}',
      output: 'try {} catch (e) {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'try {} catch(e) {}',
      output: 'try {} catch (e) {}',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'try {} catch (e) {}',
      output: 'try {} catch(e) {}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'try {} catch(e) {}',
      output: 'try {} catch (e) {}',
      options: [{ catch: 'always' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'try {} catch (e) {}',
      output: 'try {} catch(e) {}',
      options: [{ catch: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },

    // ---- from space-before-function-paren._ts_.test.ts ----
    {
      code: 'function foo<T extends () => {}>() {}',
      output: 'function foo<T extends () => {}> () {}',
      errors: [
        {
          messageId: 'missingSpace',
          line: 1,
          column: 33,
        },
      ],
    },
  ],
});
