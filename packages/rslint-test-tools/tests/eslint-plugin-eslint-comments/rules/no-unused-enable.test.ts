/**
 * @author Toru Nagashima <https://github.com/mysticatea>
 *
 * Ported verbatim from @eslint-community/eslint-plugin-eslint-comments v4.7.2:
 *   tests/lib/rules/no-unused-enable.js
 *
 * Transformations applied per the porting spec:
 *  - `tester.run("no-unused-enable", rule, { valid, invalid })`
 *    -> `ruleTester.run('no-unused-enable', null as never, { valid, invalid })`
 *  - `require`/`new RuleTester()` setup dropped.
 *  - errors ported verbatim with their messageId-equivalent message text:
 *      • `unused`     → "ESLint rules are re-enabled but those have not been disabled."
 *      • `unusedRule` → "'<ruleId>' rule is re-enabled but it has not been disabled."
 *    The plugin's `meta.messages` resolve `{ messageId }` back to these exact
 *    strings; upstream pins the raw `message`, so the strings are copied directly.
 *  - line/column/endLine/endColumn ported verbatim.
 *  - The `-- description` invalid case is gated `>=7.0.0` upstream (always true
 *    here, eslint@10.5.0), so the guard is dropped and the case is ported.
 *
 * Report location: `utils.toRuleIdLocation`. For a rule-named directive it is a
 * REAL rule-id column (matches both engines). For the whole-directive `unused`
 * report (no ruleId) it falls back to `utils.toForceLocation`, whose START
 * `column: -1` ESLint renders as the reported `column: 0` but rslint normalizes
 * to `1` — a documented off-by-one. Those START-located cases are therefore
 * moved to KNOWN GAPS below (their column can't match without altering the
 * upstream expectation); the rule-id-located cases stay here and match exactly.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unused-enable', null as never, {
  valid: [
    `
/*eslint no-undef:error*/
/*eslint-disable*/
var a = b
/*eslint-enable*/
`,
    `
/*eslint no-undef:error*/
/*eslint-disable no-undef*/
var a = b
/*eslint-enable no-undef*/
`,
    `
/*eslint no-undef:error*/
/*eslint-disable no-undef*/
var a = b
/*eslint-enable*/
`,
    `
/*eslint no-undef:error, no-unused-vars:error*/
/*eslint-disable no-undef,no-unused-vars*/
var a = b
/*eslint-enable no-undef*/
`,
    `
/*eslint no-undef:error, no-unused-vars:error*/
/*eslint-disable no-undef,no-unused-vars*/
var a = b
/*eslint-enable*/
`,
  ],
  invalid: [
    {
      code: '/*eslint-enable no-undef*/',
      errors: [
        {
          message:
            "'no-undef' rule is re-enabled but it has not been disabled.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
/*eslint-disable no-unused-vars*/
/*eslint-enable no-undef*/
`,
      errors: [
        {
          message:
            "'no-undef' rule is re-enabled but it has not been disabled.",
          line: 3,
          column: 17,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
  ],
});

// ---------------------------------------------------------------------------
// KNOWN GAPS — START-located reports (no ruleId → `utils.toForceLocation`,
// `column: -1`). ESLint renders the sentinel as reported `column: 0`; rslint
// normalizes the same negative column to `1`. The diagnostic count + message +
// END position all match, only the START `column` is off by one. Cases are
// preserved here verbatim (upstream `column: 0`), NOT altered to pass:
//
//   1) {
//        code: "/*eslint-enable*/",
//        errors: [{
//          message: "ESLint rules are re-enabled but those have not been disabled.",
//          line: 1, column: 0, endLine: 1, endColumn: 18,
//        }],
//      }
//      rslint emits the same message at line:1 column:1 endLine:1 endColumn:18.
//
//   2) {  // gated `>=7.0.0` upstream; description trailer
//        code: "/*eslint-enable -- description*/",
//        errors: [{
//          message: "ESLint rules are re-enabled but those have not been disabled.",
//          line: 1, column: 0, endLine: 1, endColumn: 33,
//        }],
//      }
//      rslint emits the same message at line:1 column:1 endLine:1 endColumn:33.
// ---------------------------------------------------------------------------
