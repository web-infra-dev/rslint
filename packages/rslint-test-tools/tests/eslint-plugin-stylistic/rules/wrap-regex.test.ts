/**
 * @fileoverview Tests for wrap-regex rule.
 * @author Nicholas C. Zakas
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/wrap-regex/wrap-regex.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('wrap-regex', null as never, { valid, invalid })`
 *    (the `name`/`rule` fields are dropped — the RuleTester resolves the rule by id).
 *
 * The upstream test has no `options`, no `parserOptions`, no `$`/unindent tag, no
 * spreads or custom error helpers, and no Babel/Flow or external-fixture cases —
 * nothing required inlining or evaluation. There is a single `run()` block (no
 * trailing skipBabel block). The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * No rslint<->upstream gap surfaced: every fixture is plain ES regex/member
 * syntax that parses identically under rslint's ts-go parser, so there is no
 * KNOWN GAPS section.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('wrap-regex', null as never, {
  valid: [
    '(/foo/).test(bar);',
    '(/foo/ig).test(bar);',
    '/foo/;',
    'var f = 0;',
    'a[/b/];',
  ],
  invalid: [
    {
      code: '/foo/.test(bar);',
      output: '(/foo/).test(bar);',
      errors: [{ messageId: 'requireParens' }],
    },
    {
      code: '/foo/ig.test(bar);',
      output: '(/foo/ig).test(bar);',
      errors: [{ messageId: 'requireParens' }],
    },

    // https://github.com/eslint/eslint/issues/10573
    {
      code: 'if(/foo/ig.test(bar));',
      output: 'if((/foo/ig).test(bar));',
      errors: [{ messageId: 'requireParens' }],
    },
  ],
});
