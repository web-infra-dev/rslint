/**
 * @fileoverview Tests for yield-star-spacing rule.
 * @author Bryan Smith
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/yield-star-spacing/yield-star-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('yield-star-spacing', null as never, { valid, invalid })`
 *    (the `name`/`rule` fields are dropped — the RuleTester resolves the rule by id).
 *  - The four local error helpers are inlined to their final `{ messageId }` form:
 *      missingBeforeError    -> { messageId: 'missingBefore' }
 *      missingAfterError     -> { messageId: 'missingAfter' }
 *      unexpectedBeforeError -> { messageId: 'unexpectedBefore' }
 *      unexpectedAfterError  -> { messageId: 'unexpectedAfter' }
 *    (none carry `data`; the rule's messages are static strings.)
 *
 * The upstream test has no `parserOptions`, no `$`/unindent tag, no spreads, and
 * no Babel/Flow or external-fixture cases. There is a single `run()` block (no
 * trailing skipBabel block). The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * No rslint<->upstream gap surfaced: every fixture is plain ES generator
 * (`yield*`) syntax that parses identically under rslint's ts-go parser, so there
 * is no KNOWN GAPS section.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('yield-star-spacing', null as never, {
  valid: [

    // default (after)
    'function *foo(){ yield foo; }',
    'function *foo(){ yield* foo; }',

    // after
    {
      code: 'function *foo(){ yield foo; }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ yield* foo; }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ yield* foo(); }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ yield* 0 }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ yield* []; }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ var result = yield* foo(); }',
      options: ['after'],
    },
    {
      code: 'function *foo(){ var result = yield* (foo()); }',
      options: ['after'],
    },

    // before
    {
      code: 'function *foo(){ yield foo; }',
      options: ['before'],
    },
    {
      code: 'function *foo(){ yield *foo; }',
      options: ['before'],
    },
    {
      code: 'function *foo(){ yield *foo(); }',
      options: ['before'],
    },
    {
      code: 'function *foo(){ yield *0 }',
      options: ['before'],
    },
    {
      code: 'function *foo(){ yield *[]; }',
      options: ['before'],
    },
    {
      code: 'function *foo(){ var result = yield *foo(); }',
      options: ['before'],
    },

    // both
    {
      code: 'function *foo(){ yield foo; }',
      options: ['both'],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      options: ['both'],
    },
    {
      code: 'function *foo(){ yield * foo(); }',
      options: ['both'],
    },
    {
      code: 'function *foo(){ yield * 0 }',
      options: ['both'],
    },
    {
      code: 'function *foo(){ yield * []; }',
      options: ['both'],
    },
    {
      code: 'function *foo(){ var result = yield * foo(); }',
      options: ['both'],
    },

    // neither
    {
      code: 'function *foo(){ yield foo; }',
      options: ['neither'],
    },
    {
      code: 'function *foo(){ yield*foo; }',
      options: ['neither'],
    },
    {
      code: 'function *foo(){ yield*foo(); }',
      options: ['neither'],
    },
    {
      code: 'function *foo(){ yield*0 }',
      options: ['neither'],
    },
    {
      code: 'function *foo(){ yield*[]; }',
      options: ['neither'],
    },
    {
      code: 'function *foo(){ var result = yield*foo(); }',
      options: ['neither'],
    },

    // object option
    {
      code: 'function *foo(){ yield* foo; }',
      options: [{ before: false, after: true }],
    },
    {
      code: 'function *foo(){ yield *foo; }',
      options: [{ before: true, after: false }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      options: [{ before: true, after: true }],
    },
    {
      code: 'function *foo(){ yield*foo; }',
      options: [{ before: false, after: false }],
    },
  ],

  invalid: [

    // default (after)
    {
      code: 'function *foo(){ yield *foo1; }',
      output: 'function *foo(){ yield* foo1; }',
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },

    // after
    {
      code: 'function *foo(){ yield *foo1; }',
      output: 'function *foo(){ yield* foo1; }',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      output: 'function *foo(){ yield* foo; }',
      options: ['after'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },
    {
      code: 'function *foo(){ yield*foo2; }',
      output: 'function *foo(){ yield* foo2; }',
      options: ['after'],
      errors: [{ messageId: 'missingAfter' }],
    },

    // before
    {
      code: 'function *foo(){ yield* foo; }',
      output: 'function *foo(){ yield *foo; }',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      output: 'function *foo(){ yield *foo; }',
      options: ['before'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function *foo(){ yield*foo; }',
      output: 'function *foo(){ yield *foo; }',
      options: ['before'],
      errors: [{ messageId: 'missingBefore' }],
    },

    // both
    {
      code: 'function *foo(){ yield* foo; }',
      output: 'function *foo(){ yield * foo; }',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }],
    },
    {
      code: 'function *foo(){ yield *foo3; }',
      output: 'function *foo(){ yield * foo3; }',
      options: ['both'],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(){ yield*foo4; }',
      output: 'function *foo(){ yield * foo4; }',
      options: ['both'],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },

    // neither
    {
      code: 'function *foo(){ yield* foo; }',
      output: 'function *foo(){ yield*foo; }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function *foo(){ yield *foo; }',
      output: 'function *foo(){ yield*foo; }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      output: 'function *foo(){ yield*foo; }',
      options: ['neither'],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },

    // object option
    {
      code: 'function *foo(){ yield*foo; }',
      output: 'function *foo(){ yield* foo; }',
      options: [{ before: false, after: true }],
      errors: [{ messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      output: 'function *foo(){ yield *foo; }',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'unexpectedAfter' }],
    },
    {
      code: 'function *foo(){ yield*foo; }',
      output: 'function *foo(){ yield * foo; }',
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missingBefore' }, { messageId: 'missingAfter' }],
    },
    {
      code: 'function *foo(){ yield * foo; }',
      output: 'function *foo(){ yield*foo; }',
      options: [{ before: false, after: false }],
      errors: [{ messageId: 'unexpectedBefore' }, { messageId: 'unexpectedAfter' }],
    },
  ],

});
