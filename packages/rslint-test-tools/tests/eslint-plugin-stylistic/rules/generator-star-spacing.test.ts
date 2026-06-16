/**
 * @fileoverview Tests for generator-star-spacing rule.
 * @author Jamund Ferguson
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/generator-star-spacing/generator-star-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('generator-star-spacing', null as never, { valid, invalid })`
 *  - The local error helpers (`missingBeforeError` / `missingAfterError` /
 *    `unexpectedBeforeError` / `unexpectedAfterError`) are inlined to their final
 *    `{ messageId: ... }` objects. These messages carry no `data`, so no
 *    interpolation is involved.
 *  - `parserOptions` (`ecmaVersion: 8` on the six async valid cases) dropped —
 *    rslint resolves syntax via tsconfig (`target/module: esnext`), so async
 *    functions / methods parse without an explicit `ecmaVersion`.
 *
 * The upstream test file is a single `run()` block (no `if (!skipBabel)` block,
 * no Babel/Flow cases, no `$` unindent — every fixture is a single-line string
 * literal). There are no octal/escape or import-attribute fixtures, so nothing
 * trips ts-go's strict/module parser. The `._js_` / `._ts_` / `._jsx_` /
 * `._css_` / `._json_` / `._markdown_` split test files don't exist for this
 * rule (it ships a single `generator-star-spacing.test.ts`).
 *
 * No rslint<->upstream gap was found: every case runs in the green
 * `ruleTester.run` below and there is no KNOWN GAPS block.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('generator-star-spacing', null as never, {
  valid: [

    // Default ("before")
    'function foo(){}',
    'function *foo(){}',
    'function *foo(arg1, arg2){}',
    'var foo = function *foo(){};',
    'var foo = function *(){};',
    'var foo = { *foo(){} };',
    'var foo = {*foo(){} };',
    'class Foo { *foo(){} }',
    'class Foo {*foo(){} }',
    'class Foo { static *foo(){} }',
    'var foo = {*[ foo ](){} };',
    'class Foo {*[ foo ](){} }',

    // "before"
    {
      code: 'function foo(){}',
      options: ['before'],
    },
    {
      code: 'function *foo(){}',
      options: ['before'],
    },
    {
      code: 'function *foo(arg1, arg2){}',
      options: ['before'],
    },
    {
      code: 'var foo = function *foo(){};',
      options: ['before'],
    },
    {
      code: 'var foo = function *(){};',
      options: ['before'],
    },
    {
      code: 'var foo = { *foo(){} };',
      options: ['before'],
    },
    {
      code: 'var foo = {*foo(){} };',
      options: ['before'],
    },
    {
      code: 'class Foo { *foo(){} }',
      options: ['before'],
    },
    {
      code: 'class Foo {*foo(){} }',
      options: ['before'],
    },
    {
      code: 'class Foo { static *foo(){} }',
      options: ['before'],
    },
    {
      code: 'class Foo {*[ foo ](){} }',
      options: ['before'],
    },
    {
      code: 'var foo = {*[ foo ](){} };',
      options: ['before'],
    },

    // "after"
    {
      code: 'function foo(){}',
      options: ['after'],
    },
    {
      code: 'function* foo(){}',
      options: ['after'],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      options: ['after'],
    },
    {
      code: 'var foo = function* foo(){};',
      options: ['after'],
    },
    {
      code: 'var foo = function* (){};',
      options: ['after'],
    },
    {
      code: 'var foo = {* foo(){} };',
      options: ['after'],
    },
    {
      code: 'var foo = { * foo(){} };',
      options: ['after'],
    },
    {
      code: 'class Foo {* foo(){} }',
      options: ['after'],
    },
    {
      code: 'class Foo { * foo(){} }',
      options: ['after'],
    },
    {
      code: 'class Foo { static* foo(){} }',
      options: ['after'],
    },
    {
      code: 'var foo = {* [foo](){} };',
      options: ['after'],
    },
    {
      code: 'class Foo {* [foo](){} }',
      options: ['after'],
    },

    // "both"
    {
      code: 'function foo(){}',
      options: ['both'],
    },
    {
      code: 'function * foo(){}',
      options: ['both'],
    },
    {
      code: 'function * foo(arg1, arg2){}',
      options: ['both'],
    },
    {
      code: 'var foo = function * foo(){};',
      options: ['both'],
    },
    {
      code: 'var foo = function * (){};',
      options: ['both'],
    },
    {
      code: 'var foo = { * foo(){} };',
      options: ['both'],
    },
    {
      code: 'var foo = {* foo(){} };',
      options: ['both'],
    },
    {
      code: 'class Foo { * foo(){} }',
      options: ['both'],
    },
    {
      code: 'class Foo {* foo(){} }',
      options: ['both'],
    },
    {
      code: 'class Foo { static * foo(){} }',
      options: ['both'],
    },
    {
      code: 'var foo = {* [foo](){} };',
      options: ['both'],
    },
    {
      code: 'class Foo {* [foo](){} }',
      options: ['both'],
    },

    // "neither"
    {
      code: 'function foo(){}',
      options: ['neither'],
    },
    {
      code: 'function*foo(){}',
      options: ['neither'],
    },
    {
      code: 'function*foo(arg1, arg2){}',
      options: ['neither'],
    },
    {
      code: 'var foo = function*foo(){};',
      options: ['neither'],
    },
    {
      code: 'var foo = function*(){};',
      options: ['neither'],
    },
    {
      code: 'var foo = {*foo(){} };',
      options: ['neither'],
    },
    {
      code: 'var foo = { *foo(){} };',
      options: ['neither'],
    },
    {
      code: 'class Foo {*foo(){} }',
      options: ['neither'],
    },
    {
      code: 'class Foo { *foo(){} }',
      options: ['neither'],
    },
    {
      code: 'class Foo { static*foo(){} }',
      options: ['neither'],
    },
    {
      code: 'var foo = {*[ foo ](){} };',
      options: ['neither'],
    },
    {
      code: 'class Foo {*[ foo ](){} }',
      options: ['neither'],
    },

    // {"before": true, "after": false}
    {
      code: 'function foo(){}',
      options: [{ before: true, after: false }],
    },
    {
      code: 'function *foo(){}',
      options: [{ before: true, after: false }],
    },
    {
      code: 'function *foo(arg1, arg2){}',
      options: [{ before: true, after: false }],
    },
    {
      code: 'var foo = function *foo(){};',
      options: [{ before: true, after: false }],
    },
    {
      code: 'var foo = function *(){};',
      options: [{ before: true, after: false }],
    },
    {
      code: 'var foo = { *foo(){} };',
      options: [{ before: true, after: false }],
    },
    {
      code: 'var foo = {*foo(){} };',
      options: [{ before: true, after: false }],
    },
    {
      code: 'class Foo { *foo(){} }',
      options: [{ before: true, after: false }],
    },
    {
      code: 'class Foo {*foo(){} }',
      options: [{ before: true, after: false }],
    },
    {
      code: 'class Foo { static *foo(){} }',
      options: [{ before: true, after: false }],
    },

    // {"before": false, "after": true}
    {
      code: 'function foo(){}',
      options: [{ before: false, after: true }],
    },
    {
      code: 'function* foo(){}',
      options: [{ before: false, after: true }],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      options: [{ before: false, after: true }],
    },
    {
      code: 'var foo = function* foo(){};',
      options: [{ before: false, after: true }],
    },
    {
      code: 'var foo = function* (){};',
      options: [{ before: false, after: true }],
    },
    {
      code: 'var foo = {* foo(){} };',
      options: [{ before: false, after: true }],
    },
    {
      code: 'var foo = { * foo(){} };',
      options: [{ before: false, after: true }],
    },
    {
      code: 'class Foo {* foo(){} }',
      options: [{ before: false, after: true }],
    },
    {
      code: 'class Foo { * foo(){} }',
      options: [{ before: false, after: true }],
    },
    {
      code: 'class Foo { static* foo(){} }',
      options: [{ before: false, after: true }],
    },

    // {"before": true, "after": true}
    {
      code: 'function foo(){}',
      options: [{ before: true, after: true }],
    },
    {
      code: 'function * foo(){}',
      options: [{ before: true, after: true }],
    },
    {
      code: 'function * foo(arg1, arg2){}',
      options: [{ before: true, after: true }],
    },
    {
      code: 'var foo = function * foo(){};',
      options: [{ before: true, after: true }],
    },
    {
      code: 'var foo = function * (){};',
      options: [{ before: true, after: true }],
    },
    {
      code: 'var foo = { * foo(){} };',
      options: [{ before: true, after: true }],
    },
    {
      code: 'var foo = {* foo(){} };',
      options: [{ before: true, after: true }],
    },
    {
      code: 'class Foo { * foo(){} }',
      options: [{ before: true, after: true }],
    },
    {
      code: 'class Foo {* foo(){} }',
      options: [{ before: true, after: true }],
    },
    {
      code: 'class Foo { static * foo(){} }',
      options: [{ before: true, after: true }],
    },

    // {"before": false, "after": false}
    {
      code: 'function foo(){}',
      options: [{ before: false, after: false }],
    },
    {
      code: 'function*foo(){}',
      options: [{ before: false, after: false }],
    },
    {
      code: 'function*foo(arg1, arg2){}',
      options: [{ before: false, after: false }],
    },
    {
      code: 'var foo = function*foo(){};',
      options: [{ before: false, after: false }],
    },
    {
      code: 'var foo = function*(){};',
      options: [{ before: false, after: false }],
    },
    {
      code: 'var foo = {*foo(){} };',
      options: [{ before: false, after: false }],
    },
    {
      code: 'var foo = { *foo(){} };',
      options: [{ before: false, after: false }],
    },
    {
      code: 'class Foo {*foo(){} }',
      options: [{ before: false, after: false }],
    },
    {
      code: 'class Foo { *foo(){} }',
      options: [{ before: false, after: false }],
    },
    {
      code: 'class Foo { static*foo(){} }',
      options: [{ before: false, after: false }],
    },

    // full configurability
    {
      code: 'function * foo(){}',
      options: [{ before: false, after: false, named: 'both' }],
    },
    {
      code: 'var foo = function * (){};',
      options: [{ before: false, after: false, anonymous: 'both' }],
    },
    {
      code: 'class Foo { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
    },
    {
      code: 'var foo = { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
    },
    {
      code: 'var foo = { bar: function * () {} }',
      options: [{ before: false, after: false, anonymous: 'both' }],
    },
    {
      code: 'class Foo { static * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
    },
    {
      code: 'var foo = { * foo(){} }',
      options: [{ before: false, after: false, shorthand: 'both' }],
    },

    // default to top level "before"
    {
      code: 'function *foo(){}',
      options: [{ method: 'both' }],
    },

    // don't apply unrelated override
    {
      code: 'function*foo(){}',
      options: [{ before: false, after: false, method: 'both' }],
    },

    // ensure using object-type override works
    {
      code: 'function * foo(){}',
      options: [{ before: false, after: false, named: { before: true, after: true } }],
    },

    // unspecified option uses default
    {
      code: 'function *foo(){}',
      options: [{ before: false, after: false, named: { before: true } }],
    },

    // https://github.com/eslint/eslint/issues/7101#issuecomment-246080531
    {
      code: 'async function foo() { }',
    },
    {
      code: '(async function() { })',
    },
    {
      code: 'async () => { }',
    },
    {
      code: '({async foo() { }})',
    },
    {
      code: 'class A {async foo() { }}',
    },
    {
      code: '(class {async foo() { }})',
    },
  ],

  invalid: [

    // Default ("before")
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      output: 'function *foo(arg1, arg2){}',
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function*foo(){};',
      output: 'var foo = function *foo(){};',
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'var foo = function* (){};',
      output: 'var foo = function *(){};',
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = {* foo(){} };',
      output: 'var foo = {*foo(){} };',
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo {* foo(){} }',
      output: 'class Foo {*foo(){} }',
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { static* foo(){} }',
      output: 'class Foo { static *foo(){} }',
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },

    // "before"
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      output: 'function *foo(arg1, arg2){}',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function*foo(){};',
      output: 'var foo = function *foo(){};',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'var foo = function* (){};',
      output: 'var foo = function *(){};',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = {* foo(){} };',
      output: 'var foo = {*foo(){} };',
      options: ['before'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo {* foo(){} }',
      output: 'class Foo {*foo(){} }',
      options: ['before'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = {* [ foo ](){} };',
      output: 'var foo = {*[ foo ](){} };',
      options: ['before'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo {* [ foo ](){} }',
      output: 'class Foo {*[ foo ](){} }',
      options: ['before'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },

    // "after"
    {
      code: 'function*foo(){}',
      output: 'function* foo(){}',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(arg1, arg2){}',
      output: 'function* foo(arg1, arg2){}',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function *foo(){};',
      output: 'var foo = function* foo(){};',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function *(){};',
      output: 'var foo = function* (){};',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { *foo(){} };',
      output: 'var foo = { * foo(){} };',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { *foo(){} }',
      output: 'class Foo { * foo(){} }',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static *foo(){} }',
      output: 'class Foo { static* foo(){} }',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { *[foo](){} };',
      output: 'var foo = { * [foo](){} };',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { *[foo](){} }',
      output: 'class Foo { * [foo](){} }',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },

    // "both"
    {
      code: 'function*foo(){}',
      output: 'function * foo(){}',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'function*foo(arg1, arg2){}',
      output: 'function * foo(arg1, arg2){}',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function*foo(){};',
      output: 'var foo = function * foo(){};',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function*(){};',
      output: 'var foo = function * (){};',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = {*foo(){} };',
      output: 'var foo = {* foo(){} };',
      options: ['both'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo {*foo(){} }',
      output: 'class Foo {* foo(){} }',
      options: ['both'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static*foo(){} }',
      output: 'class Foo { static * foo(){} }',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = {*[foo](){} };',
      output: 'var foo = {* [foo](){} };',
      options: ['both'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo {*[foo](){} }',
      output: 'class Foo {* [foo](){} }',
      options: ['both'],
      errors: [{ messageId: 'missingAfter' }],
    },

    // "neither"
    {
      code: 'function * foo(){}',
      output: 'function*foo(){}',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function * foo(arg1, arg2){}',
      output: 'function*foo(arg1, arg2){}',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function * foo(){};',
      output: 'var foo = function*foo(){};',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function * (){};',
      output: 'var foo = function*(){};',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = { * foo(){} };',
      output: 'var foo = { *foo(){} };',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { * foo(){} }',
      output: 'class Foo { *foo(){} }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { static * foo(){} }',
      output: 'class Foo { static*foo(){} }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = { * [ foo ](){} };',
      output: 'var foo = { *[ foo ](){} };',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { * [ foo ](){} }',
      output: 'class Foo { *[ foo ](){} }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },

    // {"before": true, "after": false}
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'function* foo(arg1, arg2){}',
      output: 'function *foo(arg1, arg2){}',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function*foo(){};',
      output: 'var foo = function *foo(){};',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'var foo = function* (){};',
      output: 'var foo = function *(){};',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = {* foo(){} };',
      output: 'var foo = {*foo(){} };',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo {* foo(){} }',
      output: 'class Foo {*foo(){} }',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'unexpectedAfter' }],
    },

    // {"before": false, "after": true}
    {
      code: 'function*foo(){}',
      output: 'function* foo(){}',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(arg1, arg2){}',
      output: 'function* foo(arg1, arg2){}',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function *foo(){};',
      output: 'var foo = function* foo(){};',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function *(){};',
      output: 'var foo = function* (){};',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { *foo(){} };',
      output: 'var foo = { * foo(){} };',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { *foo(){} }',
      output: 'class Foo { * foo(){} }',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static *foo(){} }',
      output: 'class Foo { static* foo(){} }',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },

    // {"before": true, "after": true}
    {
      code: 'function*foo(){}',
      output: 'function * foo(){}',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'function*foo(arg1, arg2){}',
      output: 'function * foo(arg1, arg2){}',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function*foo(){};',
      output: 'var foo = function * foo(){};',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function*(){};',
      output: 'var foo = function * (){};',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = {*foo(){} };',
      output: 'var foo = {* foo(){} };',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo {*foo(){} }',
      output: 'class Foo {* foo(){} }',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static*foo(){} }',
      output: 'class Foo { static * foo(){} }',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },

    // {"before": false, "after": false}
    {
      code: 'function * foo(){}',
      output: 'function*foo(){}',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function * foo(arg1, arg2){}',
      output: 'function*foo(arg1, arg2){}',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function * foo(){};',
      output: 'var foo = function*foo(){};',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = function * (){};',
      output: 'var foo = function*(){};',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'var foo = { * foo(){} };',
      output: 'var foo = { *foo(){} };',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { * foo(){} }',
      output: 'class Foo { *foo(){} }',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { static * foo(){} }',
      output: 'class Foo { static*foo(){} }',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },

    // full configurability
    {
      code: 'function*foo(){}',
      output: 'function * foo(){}',
      options: [{ before: false, after: false, named: 'both' }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = function*(){};',
      output: 'var foo = function * (){};',
      options: [{ before: false, after: false, anonymous: 'both' }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { *foo(){} }',
      output: 'class Foo { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { *foo(){} }',
      output: 'var foo = { * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { bar: function*() {} }',
      output: 'var foo = { bar: function * () {} }',
      options: [{ before: false, after: false, anonymous: 'both' }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static*foo(){} }',
      output: 'class Foo { static * foo(){} }',
      options: [{ before: false, after: false, method: 'both' }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'var foo = { *foo(){} }',
      output: 'var foo = { * foo(){} }',
      options: [{ before: false, after: false, shorthand: 'both' }],
      errors: [{ messageId: 'missingAfter' }],
    },

    // default to top level "before"
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      options: [{ method: 'both' }],
      errors: [{ messageId: 'missingBefore' }],
    },

    // don't apply unrelated override
    {
      code: 'function * foo(){}',
      output: 'function*foo(){}',
      options: [{ before: false, after: false, method: 'both' }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },

    // ensure using object-type override works
    {
      code: 'function*foo(){}',
      output: 'function * foo(){}',
      options: [{ before: false, after: false, named: { before: true, after: true } }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },

    // unspecified option uses default
    {
      code: 'function*foo(){}',
      output: 'function *foo(){}',
      options: [{ before: false, after: false, named: { before: true } }],
      errors: [{ messageId: 'missingBefore' }],
    },

    // async generators
    {
      code: '({ async * foo(){} })',
      output: '({ async*foo(){} })',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: '({ async*foo(){} })',
      output: '({ async * foo(){} })',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { async * foo(){} }',
      output: 'class Foo { async*foo(){} }',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { async*foo(){} }',
      output: 'class Foo { async * foo(){} }',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'class Foo { static async * foo(){} }',
      output: 'class Foo { static async*foo(){} }',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'class Foo { static async*foo(){} }',
      output: 'class Foo { static async * foo(){} }',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },

  ],
});
