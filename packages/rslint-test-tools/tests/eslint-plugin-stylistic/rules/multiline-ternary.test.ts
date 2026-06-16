/**
 * @fileoverview Enforce newlines between operands of ternary expressions
 * @author Kai Cataldo
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/multiline-ternary/multiline-ternary.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('multiline-ternary', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *    Upstream's three JSX cases (`ignoreJSX: true`) are rendered to their literal
 *    `<>\n  {...}\n</>` strings (trailing whitespace inside the template preserved
 *    verbatim, since the rule is whitespace-sensitive).
 *  - `parserOptions` (ecmaVersion / ecmaFeatures.jsx) dropped — rslint resolves
 *    via tsconfig; the RuleTester routes the JSX fixtures (which carry `<>` / `/>`)
 *    to a `.tsx` file automatically.
 *  - The four messageIds (`expectedTestCons` / `expectedConsAlt` /
 *    `unexpectedTestCons` / `unexpectedConsAlt`) take no `data`, so each maps 1:1
 *    to a fixed message rendered from the plugin's meta.
 *  - `line` / `column` / `endLine` / `endColumn` are asserted wherever upstream
 *    pins them. `output` is asserted on every invalid case (autofix); the one
 *    `output: null` case asserts the source is left unchanged.
 *
 * There are no Babel/Flow, no external-fixture (`readFileSync`), and no
 * output-only (errors-less) cases — every upstream invalid pins explicit
 * `errors`. The `._css_` / `._json_` / `._markdown_` test files don't exist for
 * this rule.
 *
 * There is no rslint<->upstream gap for this rule: every upstream valid/invalid
 * case runs verbatim through the green `ruleTester.run` below and matches,
 * including the multi-line autofix outputs.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('multiline-ternary', null as never, {
  valid: [

    // default "always"
    'a\n? b\n: c',
    'a ?\nb :\nc',
    'a\n? b\n? c\n: d\n: e',
    'a\n? (b\n? c\n: d)\n: e',

    // "always"
    { code: 'a\n? b\n: c', options: ['always'] },
    { code: 'a ?\nb :\nc', options: ['always'] },
    { code: 'a\n? b\n? c\n: d\n: e', options: ['always'] },
    { code: 'a\n? (b\n? c\n: d)\n: e', options: ['always'] },
    { code: '(a\n ? b\n : c)', options: ['always'] },
    { code: '(a)\n? b\n: c', options: ['always'] },
    { code: '((a))\n? b\n: c', options: ['always'] },
    { code: '(a)?\n b\n: c', options: ['always'] },
    { code: '((a))?\n b\n: c', options: ['always'] },
    { code: 'a\n? (b)\n: c', options: ['always'] },
    { code: 'a\n? ((b))\n: c', options: ['always'] },
    { code: 'a ?\n (b)\n: c', options: ['always'] },
    { code: 'a ?\n ((b))\n: c', options: ['always'] },
    { code: 'a \n? b\n: (c)', options: ['always'] },
    { code: 'a \n? b\n: ((c))', options: ['always'] },
    { code: 'a \n? b:\n (c)', options: ['always'] },
    { code: 'a \n? b:\n ((c))', options: ['always'] },
    { code: '(a) \n? (b)\n: (c)', options: ['always'] },
    { code: '((a)) \n? ((b))\n: ((c))', options: ['always'] },
    { code: '((a)) ?\n ((b)):\n ((c))', options: ['always'] },
    {
      code: '<>\n  {a ? <div /> : <div />}\n</>',
      options: ['always', { ignoreJSX: true }],
    },

    // "always-multiline"
    { code: 'a\n? b\n: c', options: ['always-multiline'] },
    { code: 'a ?\nb :\nc', options: ['always-multiline'] },
    { code: 'a\n? b\n? c\n: d\n: e', options: ['always-multiline'] },
    { code: 'a\n? (b\n? c\n: d)\n: e', options: ['always-multiline'] },
    { code: 'a ? b : c', options: ['always-multiline'] },
    { code: 'a ? b ? c : d : e', options: ['always-multiline'] },
    { code: 'a ? (b ? c : d) : e', options: ['always-multiline'] },
    { code: 'a\n? (b ? c : d)\n: e', options: ['always-multiline'] },
    { code: 'a ?\n(b ? c : d) :\ne', options: ['always-multiline'] },
    { code: '(a\n ? b\n : c)', options: ['always-multiline'] },
    { code: '(\na ? b : c\n)', options: ['always-multiline'] },
    { code: '(a)\n? b\n: c', options: ['always-multiline'] },
    { code: '((a))\n? b\n: c', options: ['always-multiline'] },
    { code: '(a)?\n b\n: c', options: ['always-multiline'] },
    { code: '((a))?\n b\n: c', options: ['always-multiline'] },
    { code: 'a\n? (b)\n: c', options: ['always-multiline'] },
    { code: 'a\n? ((b))\n: c', options: ['always-multiline'] },
    { code: 'a ?\n (b)\n: c', options: ['always-multiline'] },
    { code: 'a ?\n ((b))\n: c', options: ['always-multiline'] },
    { code: 'a \n? b\n: (c)', options: ['always-multiline'] },
    { code: 'a \n? b\n: ((c))', options: ['always-multiline'] },
    { code: 'a \n? b:\n (c)', options: ['always-multiline'] },
    { code: 'a \n? b:\n ((c))', options: ['always-multiline'] },
    { code: '(a) \n? (b)\n: (c)', options: ['always-multiline'] },
    { code: '((a)) \n? ((b))\n: ((c))', options: ['always-multiline'] },
    { code: '((a)) ?\n ((b)):\n ((c))', options: ['always-multiline'] },
    { code: '(a) ? b : c', options: ['always-multiline'] },
    { code: '((a)) ? b : c', options: ['always-multiline'] },
    { code: 'a ? (b) : c', options: ['always-multiline'] },
    { code: 'a ? ((b)) : c', options: ['always-multiline'] },
    { code: 'a ? b : (c)', options: ['always-multiline'] },
    { code: 'a ? b : ((c))', options: ['always-multiline'] },
    { code: '(a) ? (b) : (c)', options: ['always-multiline'] },
    { code: '((a)) ? ((b)) : ((c))', options: ['always-multiline'] },
    {
      code: '<>\n  {a ? \n    <div /> : <div />}\n</>',
      options: ['always-multiline', { ignoreJSX: true }],
    },

    // "never"
    { code: 'a ? b : c', options: ['never'] },
    { code: 'a ? b ? c : d : e', options: ['never'] },
    { code: 'a ? (b ? c : d) : e', options: ['never'] },
    { code: 'a  +\n b ? c +\n d : e + \nf', options: ['never'] },
    { code: '(\na ? b : c\n)', options: ['never'] },
    { code: '(a) ? b : c', options: ['never'] },
    { code: '((a)) ? b : c', options: ['never'] },
    { code: 'a ? (b) : c', options: ['never'] },
    { code: 'a ? ((b)) : c', options: ['never'] },
    { code: 'a ? b : (c)', options: ['never'] },
    { code: 'a ? b : ((c))', options: ['never'] },
    { code: '(a) ? (b) : (c)', options: ['never'] },
    { code: '((a)) ? ((b)) : ((c))', options: ['never'] },
    { code: '(a\n) ? b : c', options: ['never'] },
    { code: '((a)\n) ? b : c', options: ['never'] },
    { code: 'a ? (\nb) : c', options: ['never'] },
    { code: 'a ? (\n(b)) : c', options: ['never'] },
    { code: 'a ? (b\n) : c', options: ['never'] },
    { code: 'a ? ((b)\n) : c', options: ['never'] },
    { code: 'a ? b : (\nc)', options: ['never'] },
    { code: 'a ? b : (\n(c))', options: ['never'] },
    { code: '(a\n) ? (\nb\n) : (\nc)', options: ['never'] },
    { code: '((a)\n) ? (\n(b)\n) : (\n(c))', options: ['never'] },
    {
      code: '<>\n  {a \n    ? <div /> \n    : <div />}\n</>',
      options: ['never', { ignoreJSX: true }],
    },
  ],

  invalid: [
    // default "always"
    {
      code: 'a ? b : c',
      output: 'a\n? b\n: c',
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
      }],
    },
    {
      code: 'a\n? b : c',
      output: 'a\n? b\n: c',
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
      }],
    },
    {
      code: 'a ? b\n: c',
      output: 'a\n? b\n: c',
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }],
    },
    {
      code: 'a ? (b ? c : d) : e',
      output: 'a\n? (b\n? c\n: d)\n: e',
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 16,
      }, {
        messageId: 'expectedTestCons',
        line: 1,
        column: 6,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 10,
      }],
    },
    {
      code: 'a ?\n(b ? c : d) :\ne',
      output: 'a ?\n(b\n? c\n: d) :\ne',
      errors: [{
        messageId: 'expectedTestCons',
        line: 2,
        column: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 6,
      }],
    },
    {
      code: 'a ? (b\n? c\n: d) : e',
      output: 'a\n? (b\n? c\n: d)\n: e',
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 3,
        endColumn: 5,
      }],
    },
    {
      code: 'a ?\n(b? c\n: d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 5,
      }, {
        messageId: 'expectedTestCons',
        line: 2,
        column: 2,
      }],
    },
    {
      code: 'a ?\n(b\n? c : d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 9,
      }, {
        messageId: 'expectedConsAlt',
        line: 3,
        column: 3,
      }],
    },
    {
      code: 'a ?\n(b\n? c\n : d) : e',
      output: 'a ?\n(b\n? c\n : d)\n: e',
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 4,
        endColumn: 6,
      }],
    },

    // "always"
    {
      code: 'a ? b : c',
      output: 'a\n? b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'f() ? a + b : c',
      output: 'f()\n? a + b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 4,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 7,
        endLine: 1,
        endColumn: 12,
      }],
    },
    {
      code: 'a\n? b : c',
      output: 'a\n? b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
      }],
    },
    {
      code: 'a ? b\n: c',
      output: 'a\n? b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }],
    },
    {
      code: 'a ? (b ? c : d) : e',
      output: 'a\n? (b\n? c\n: d)\n: e',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 16,
      }, {
        messageId: 'expectedTestCons',
        line: 1,
        column: 6,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 10,
      }],
    },
    {
      code: 'a ?\n(b ? c : d) :\ne',
      output: 'a ?\n(b\n? c\n: d) :\ne',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 2,
        column: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 6,
      }],
    },
    {
      code: 'a ? (b\n? c\n: d) : e',
      output: 'a\n? (b\n? c\n: d)\n: e',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 3,
        endColumn: 5,
      }],
    },
    {
      code: 'a ?\n(b? c\n: d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 5,
      }, {
        messageId: 'expectedTestCons',
        line: 2,
        column: 2,
      }],
    },
    {
      code: 'a ?\n(b\n? c : d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 9,
      }, {
        messageId: 'expectedConsAlt',
        line: 3,
        column: 3,
      }],
    },
    {
      code: 'a ?\n(b\n? c\n : d) : e',
      output: 'a ?\n(b\n? c\n : d)\n: e',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 4,
        endColumn: 6,
      }],
    },
    {
      code: '(a\n) ? b\n: c',
      output: '(a\n)\n? b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '((a)\n) ? b\n: c',
      output: '((a)\n)\n? b\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'a ? (\nb)\n: c',
      output: 'a\n? (\nb)\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 2,
      }],
    },
    {
      code: 'a ? (\n(b))\n: c',
      output: 'a\n? (\n(b))\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? (b\n): c',
      output: 'a\n? (b\n)\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 3,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? ((b)\n): c',
      output: 'a\n? ((b)\n)\n: c',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 3,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? b : (\nc)',
      output: 'a\n? b\n: (\nc)',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a\n? b : (\n(c))',
      output: 'a\n? b\n: (\n(c))',
      options: ['always'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: '(a\n) ? (\nb\n) : (\nc)',
      output: '(a\n)\n? (\nb\n)\n: (\nc)',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 5,
        endLine: 4,
        endColumn: 2,
      }],
    },
    {
      code: '((a)\n) ? (\n(b)\n) : (\n(c))',
      output: '((a)\n)\n? (\n(b)\n)\n: (\n(c))',
      options: ['always'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 5,
        endLine: 4,
        endColumn: 2,
      }],
    },

    // "always-multiline"
    {
      code: 'a\n? b : c',
      output: 'a\n? b\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
      }],
    },
    {
      code: 'a ? b\n: c',
      output: 'a\n? b\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }],
    },
    {
      code: 'a &&\nb ? c : d',
      output: 'a &&\nb\n? c\n: d',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 5,
      }],
    },
    {
      code: 'a ? b +\nc : d',
      output: 'a\n? b +\nc\n: d',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
      }],
    },
    {
      code: 'a ? b : c +\nd',
      output: 'a\n? b\n: c +\nd',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
      }],
    },
    {
      code: 'a ?\n(b ? c : d) : e',
      output: 'a ?\n(b ? c : d)\n: e',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 12,
      }],
    },
    {
      code: 'a ? (b ? c : d) :\ne',
      output: 'a\n? (b ? c : d) :\ne',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }],
    },
    {
      code: 'a ? (b\n? c\n: d) : e',
      output: 'a\n? (b\n? c\n: d)\n: e',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'expectedConsAlt',
        line: 1,
        column: 5,
        endLine: 3,
        endColumn: 5,
      }],
    },
    {
      code: 'a ?\n(b ? c\n: d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 5,
      }, {
        messageId: 'expectedTestCons',
        line: 2,
        column: 2,
      }],
    },
    {
      code: 'a ?\n(b\n? c : d) : e',
      output: 'a ?\n(b\n? c\n: d)\n: e',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 3,
        endColumn: 9,
      }, {
        messageId: 'expectedConsAlt',
        line: 3,
        column: 3,
      }],
    },
    {
      code: 'a ?\n(b\n? c\n : d) : e',
      output: 'a ?\n(b\n? c\n : d)\n: e',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 1,
        endLine: 4,
        endColumn: 6,
      }],
    },
    {
      code: '(a\n) ? b\n: c',
      output: '(a\n)\n? b\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '((a)\n) ? b\n: c',
      output: '((a)\n)\n? b\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'a ? (\nb)\n: c',
      output: 'a\n? (\nb)\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 2,
      }],
    },
    {
      code: 'a ? (\n(b))\n: c',
      output: 'a\n? (\n(b))\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? (b\n): c',
      output: 'a\n? (b\n)\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 3,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? ((b)\n): c',
      output: 'a\n? ((b)\n)\n: c',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 3,
        endColumn: 2,
      }],
    },
    {
      code: 'a\n? b : (\nc)',
      output: 'a\n? b\n: (\nc)',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a\n? b : (\n(c))',
      output: 'a\n? b\n: (\n(c))',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedConsAlt',
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: '(a\n) ? (\nb\n) : (\nc)',
      output: '(a\n)\n? (\nb\n)\n: (\nc)',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 5,
        endLine: 4,
        endColumn: 2,
      }],
    },
    {
      code: '((a)\n) ? (\n(b)\n) : (\n(c))',
      output: '((a)\n)\n? (\n(b)\n)\n: (\n(c))',
      options: ['always-multiline'],
      errors: [{
        messageId: 'expectedTestCons',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'expectedConsAlt',
        line: 2,
        column: 5,
        endLine: 4,
        endColumn: 2,
      }],
    },

    // "never"
    {
      code: 'a\n? b : c',
      output: 'a? b : c',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }],
    },
    {
      code: 'a ? b\n: c',
      output: 'a ? b: c',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedConsAlt',
        line: 1,
        column: 5,
      }],
    },
    {
      code: 'a ?\n(b ? c : d) :\ne',
      output: 'a ?(b ? c : d) :e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 12,
      }],
    },
    {
      code: 'a ? (b\n? c\n: d) : e',
      output: 'a ? (b? c: d) : e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 6,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 3,
      }],
    },
    {
      code: 'a ?\n(b? c\n: d) : e',
      output: 'a ?(b? c: d) : e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 5,
      }],
    },
    {
      code: 'a ?\n(b\n? c : d) : e',
      output: 'a ?(b? c : d) : e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'unexpectedTestCons',
        line: 2,
        column: 2,
      }],
    },
    {
      code: 'a ?\n(b\n? c\n : d) : e',
      output: 'a ?(b? c: d) : e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'unexpectedTestCons',
        line: 2,
        column: 2,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 3,
        column: 3,
      }],
    },
    {
      code: 'a ? (b\n? c\n: d)\n: e',
      output: 'a ? (b? c: d): e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedConsAlt',
        line: 1,
        column: 5,
        endLine: 3,
        endColumn: 5,
      }, {
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 6,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 3,
      }],
    },
    {
      code: 'a\n?\n(b\n?\nc\n:\nd)\n:\ne',
      output: 'a?(b?c:d):e',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 3,
        column: 1,
        endLine: 7,
        endColumn: 3,
      }, {
        messageId: 'unexpectedTestCons',
        line: 3,
        column: 2,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 5,
        column: 1,
      }],
    },
    {
      code: '(a)\n ? b \n : (c)',
      output: '(a)? b: (c)',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 4,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 4,
        endLine: 2,
        endColumn: 5,
      }],
    },
    {
      code: '(a)\n ? (b) \n : (c)',
      output: '(a)? (b): (c)',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 4,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 4,
        endLine: 2,
        endColumn: 7,
      }],
    },
    {
      code: '((a))\n ? ((b)) \n : ((c))',
      output: '((a))? ((b)): ((c))',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedTestCons',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 6,
      }, {
        messageId: 'unexpectedConsAlt',
        line: 2,
        column: 4,
        endLine: 2,
        endColumn: 9,
      }],
    },

    // https://github.com/eslint/eslint/pull/12982#discussion_r409120960
    {
      code: 'a ? // comment\nb : c;',
      output: null,
      errors: [{
        messageId: 'expectedConsAlt',
      }],
    },
  ],
});
