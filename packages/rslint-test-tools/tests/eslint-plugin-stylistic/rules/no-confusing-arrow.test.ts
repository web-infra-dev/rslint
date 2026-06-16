/**
 * @fileoverview Tests for no-confusing-arrow rule.
 * @author Jxck <https://github.com/Jxck>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-confusing-arrow/no-confusing-arrow.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('no-confusing-arrow', null as never, { valid, invalid })`
 *  - `rule` / the `#test` import / `name` dropped; the rule is resolved from the
 *    mounted live plugin by id, not passed in.
 *  - The single `{ messageId: 'confusing' }` error is kept as-is; the rule's
 *    `confusing` message takes no `data`, so it renders to a fixed string
 *    ('Arrow function used ambiguously with a conditional expression.').
 *  - No `parserOptions` / `type` / `$` unindent / spread / custom helpers exist in
 *    this upstream file, so nothing was expanded or stripped on those grounds.
 *
 * The upstream file has exactly ONE `run()` block (no `if (!skipBabel)` block, no
 * Babel/Flow cases, no second block). The `._css_` / `._json_` / `._markdown_`
 * test files don't exist for this rule.
 *
 * Every invalid case upstream pins `errors`; the three cases that pin `output: null`
 * (the `allowParens: false` variants, where the rule's fixer returns `null`) keep
 * that pin — the RuleTester asserts the source is left unchanged. There are NO
 * output-only invalid cases.
 *
 * NO case surfaces a real rslint<->upstream gap: all fixtures are valid TypeScript
 * under ts-go (no JSX, no sloppy-mode-only syntax, no octal/escape edge cases), the
 * diagnostics, columns, and fix outputs all match upstream. The `KNOWN GAPS` block
 * at the bottom is therefore empty (documented as such).
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-confusing-arrow', null as never, {
  valid: [
    'a => { return 1 ? 2 : 3; }',
    { code: 'a => { return 1 ? 2 : 3; }', options: [{ allowParens: false }] },

    'var x = a => { return 1 ? 2 : 3; }',
    { code: 'var x = a => { return 1 ? 2 : 3; }', options: [{ allowParens: false }] },

    'var x = (a) => { return 1 ? 2 : 3; }',
    { code: 'var x = (a) => { return 1 ? 2 : 3; }', options: [{ allowParens: false }] },

    'var x = a => (1 ? 2 : 3)',
    { code: 'var x = a => (1 ? 2 : 3)', options: [{ allowParens: true }] },

    'var x = (a,b) => (1 ? 2 : 3)',
    { code: '() => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
    { code: '(a, b) => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
    { code: '(a = b) => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
    { code: '({ a }) => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
    { code: '([a]) => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
    { code: '(...a) => 1 ? 2 : 3', options: [{ onlyOneSimpleParam: true }] },
  ],
  invalid: [
    {
      code: 'a => 1 ? 2 : 3',
      output: 'a => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'a => 1 ? 2 : 3',
      output: 'a => (1 ? 2 : 3)',
      options: [{ allowParens: true }],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'a => 1 ? 2 : 3',
      output: null,
      options: [{ allowParens: false }],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = a => 1 ? 2 : 3',
      output: 'var x = a => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = a => 1 ? 2 : 3',
      output: 'var x = a => (1 ? 2 : 3)',
      options: [{ allowParens: true }],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = a => 1 ? 2 : 3',
      output: null,
      options: [{ allowParens: false }],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = (a) => 1 ? 2 : 3',
      output: 'var x = (a) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = () => 1 ? 2 : 3',
      output: 'var x = () => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = () => 1 ? 2 : 3',
      output: 'var x = () => (1 ? 2 : 3)',
      options: [{}],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = () => 1 ? 2 : 3',
      output: 'var x = () => (1 ? 2 : 3)',
      options: [{ onlyOneSimpleParam: false }],
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = (a, b) => 1 ? 2 : 3',
      output: 'var x = (a, b) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = (a = b) => 1 ? 2 : 3',
      output: 'var x = (a = b) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = ({ a }) => 1 ? 2 : 3',
      output: 'var x = ({ a }) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = ([a]) => 1 ? 2 : 3',
      output: 'var x = ([a]) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
    {
      code: 'var x = (...a) => 1 ? 2 : 3',
      output: 'var x = (...a) => (1 ? 2 : 3)',
      errors: [{ messageId: 'confusing' }],
    },
  ],
});

/**
 * ====================== no-confusing-arrow — KNOWN GAPS ======================
 *
 * None. Every upstream v5.10.0 fixture for this rule is valid TypeScript under
 * rslint's ts-go parser — plain arrow functions with conditional bodies, no JSX,
 * no sloppy-mode-only syntax, no octal/escape edge cases, no `assert`/`with` import
 * attributes. The diagnostic count (always 1 `confusing`), the rendered message,
 * and the `--fix` output (paren-wrap when `allowParens` is truthy, no-op / unchanged
 * source when `allowParens: false`) all match upstream exactly, so no case is
 * isolated. There are NO output-only invalid cases (every invalid case pins `errors`).
 * =============================================================================
 */
