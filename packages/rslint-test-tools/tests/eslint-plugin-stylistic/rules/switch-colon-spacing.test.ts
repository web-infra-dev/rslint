/**
 * @fileoverview Tests for switch-colon-spacing rule.
 * @author Toru Nagashima
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/switch-colon-spacing/switch-colon-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('switch-colon-spacing', null as never, { valid, invalid })`
 *  - The four local error helpers are inlined to their final `{ messageId }`:
 *      expectedBeforeError    -> { messageId: 'expectedBefore' }
 *      expectedAfterError     -> { messageId: 'expectedAfter' }
 *      unexpectedBeforeError  -> { messageId: 'unexpectedBefore' }
 *      unexpectedAfterError   -> { messageId: 'unexpectedAfter' }
 *    None of these messages use `{{}}` interpolation, so the RuleTester resolves
 *    each `messageId` to its full static message text from the plugin's meta.
 *  - Every invalid case pins both `errors` and `output`; the RuleTester asserts
 *    the diagnostic count + each message and the `--fix` output. Upstream pins no
 *    line/column for this rule, so none are asserted.
 *
 * The upstream test file is a single `run()` block (no `if (!skipBabel)` block,
 * no Babel/Flow cases, no `$` unindent, no spread/helpers beyond the inlined
 * error objects). Every fixture is a single-line string literal — several embed
 * `\n` inside the switch (e.g. `case 0\n:\nbreak;`), which is valid TS and parses
 * under ts-go. No octal/escape/import-attribute fixtures exist, so nothing trips
 * ts-go's strict/module parser. The `._js_` / `._ts_` / `._css_` / `._json_` /
 * `._markdown_` split test files don't exist for this rule (it ships a single
 * `switch-colon-spacing.test.ts`).
 *
 * No rslint<->upstream gap was found: every case runs in the green
 * `ruleTester.run` below and there is no KNOWN GAPS block.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('switch-colon-spacing', null as never, {
  valid: [
    'switch(a){}',
    '({foo:1,bar : 2});',
    'A:foo(); B : foo();',
    'switch(a){case 0: break;}',
    'switch(a){case 0:}',
    'switch(a){case 0\n:\nbreak;}',
    'switch(a){default: break;}',
    'switch(a){default:}',
    'switch(a){default\n:\nbreak;}',
    { code: 'switch(a){case 0:break;}', options: [{ before: false, after: false }] },
    { code: 'switch(a){case 0:}', options: [{ before: false, after: false }] },
    { code: 'switch(a){case 0\n:\nbreak;}', options: [{ before: false, after: false }] },
    { code: 'switch(a){default:break;}', options: [{ before: false, after: false }] },
    { code: 'switch(a){default:}', options: [{ before: false, after: false }] },
    { code: 'switch(a){default\n:\nbreak;}', options: [{ before: false, after: false }] },
    { code: 'switch(a){case 0: break;}', options: [{ before: false, after: true }] },
    { code: 'switch(a){case 0:}', options: [{ before: false, after: true }] },
    { code: 'switch(a){case 0\n:\nbreak;}', options: [{ before: false, after: true }] },
    { code: 'switch(a){default: break;}', options: [{ before: false, after: true }] },
    { code: 'switch(a){default:}', options: [{ before: false, after: true }] },
    { code: 'switch(a){default\n:\nbreak;}', options: [{ before: false, after: true }] },
    { code: 'switch(a){case 0 :break;}', options: [{ before: true, after: false }] },
    { code: 'switch(a){case 0 :}', options: [{ before: true, after: false }] },
    { code: 'switch(a){case 0\n:\nbreak;}', options: [{ before: true, after: false }] },
    { code: 'switch(a){default :break;}', options: [{ before: true, after: false }] },
    { code: 'switch(a){default :}', options: [{ before: true, after: false }] },
    { code: 'switch(a){default\n:\nbreak;}', options: [{ before: true, after: false }] },
    { code: 'switch(a){case 0 : break;}', options: [{ before: true, after: true }] },
    { code: 'switch(a){case 0 :}', options: [{ before: true, after: true }] },
    { code: 'switch(a){case 0\n:\nbreak;}', options: [{ before: true, after: true }] },
    { code: 'switch(a){default : break;}', options: [{ before: true, after: true }] },
    { code: 'switch(a){default :}', options: [{ before: true, after: true }] },
    { code: 'switch(a){default\n:\nbreak;}', options: [{ before: true, after: true }] },
  ],
  invalid: [
    {
      code: 'switch(a){case 0 :break;}',
      output: 'switch(a){case 0: break;}',
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){default :break;}',
      output: 'switch(a){default: break;}',
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0 : break;}',
      output: 'switch(a){case 0:break;}',
      options: [{ before: false, after: false }],
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'unexpectedAfter' },
      ],
    },
    {
      code: 'switch(a){default : break;}',
      output: 'switch(a){default:break;}',
      options: [{ before: false, after: false }],
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'unexpectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0 :break;}',
      output: 'switch(a){case 0: break;}',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){default :break;}',
      output: 'switch(a){default: break;}',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0: break;}',
      output: 'switch(a){case 0 :break;}',
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'expectedBefore' },
        { messageId: 'unexpectedAfter' },
      ],
    },
    {
      code: 'switch(a){default: break;}',
      output: 'switch(a){default :break;}',
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'expectedBefore' },
        { messageId: 'unexpectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0:break;}',
      output: 'switch(a){case 0 : break;}',
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'expectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){default:break;}',
      output: 'switch(a){default : break;}',
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'expectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0 /**/ :break;}',
      output: 'switch(a){case 0 /**/ : break;}',
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){case 0 :/**/break;}',
      output: 'switch(a){case 0:/**/break;}',
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
    {
      code: 'switch(a){case (0) :break;}',
      output: 'switch(a){case (0): break;}',
      errors: [
        { messageId: 'unexpectedBefore' },
        { messageId: 'expectedAfter' },
      ],
    },
  ],
});
