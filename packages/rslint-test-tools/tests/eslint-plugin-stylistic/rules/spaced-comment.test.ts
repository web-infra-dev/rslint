/**
 * @fileoverview Test for spaced-comments
 * @author Gyandeep Singh
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/spaced-comment/spaced-comment.test.ts
 *
 * The upstream file has a single `run({ name, rule, valid, invalid })` block
 * (no second skipBabel block). Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('spaced-comment', null as never, { valid, invalid })`
 *  - The `validShebangProgram` const is inlined to its string value.
 *  - Upstream uses no `$` unindent tag; every `code` is a plain string literal
 *    (with `\n`, `\r\n` and `\u2028` escapes) and is copied byte-for-byte.
 *  - The `linterOptions: { reportUnusedDisableDirectives: false }` field on one
 *    valid case is dropped — it tunes ESLint's own unused-directive reporting,
 *    not the spaced-comment rule; rslint reports under the rule id only, so this
 *    fixture's `eslint-enable`/`eslint-disable` marker behaviour is unaffected
 *    (verified: 0 spaced-comment diagnostics).
 *
 * No Babel/Flow cases, no external-fixture (`readFileSync`) cases, and no
 * `._css_`/`._json_`/`._markdown_` files exist for this rule. Every fixture is
 * pure source-with-comments and parses cleanly under ts-go, so nothing is
 * isolated into KNOWN GAPS — see the note at the bottom confirming the empty
 * gap set.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const validShebangProgram = '#!/path/to/node\nvar a = 3;';

ruleTester.run('spaced-comment', null as never, {
  valid: [
    {
      code: '// A valid comment starting with space\nvar a = 1;',
      options: ['always'],
    },
    {
      code: '//   A valid comment starting with tab\nvar a = 1;',
      options: ['always'],
    },
    {
      code: '//A valid comment NOT starting with space\nvar a = 2;',
      options: ['never'],
    },

    // exceptions - line comments
    {
      code: '//-----------------------\n// A comment\n//-----------------------',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
    },
    {
      code: '//-----------------------\n// A comment\n//-----------------------',
      options: ['always', {
        line: { exceptions: ['-', '=', '*', '#', '!@#'] },
      }],
    },
    {
      code: '//===========\n// A comment\n//*************',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
    },
    {
      code: '//######\n// A comment',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
    },
    {
      code: '//!@#!@#!@#\n// A comment\n//!@#',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
    },

    // exceptions - block comments
    {
      code: 'var a = 1; /*######*/',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
    },
    {
      code: 'var a = 1; /*######*/',
      options: ['always', {
        block: { exceptions: ['-', '=', '*', '#', '!@#'] },
      }],
    },
    {
      code: '/*****************\n * A comment\n *****************/',
      options: ['always', {
        exceptions: ['*'],
      }],
    },
    {
      code: '/*++++++++++++++\n * A comment\n +++++++++++++++++*/',
      options: ['always', {
        exceptions: ['+'],
      }],
    },
    {
      code: '/*++++++++++++++\n + A comment\n * B comment\n - C comment\n----------------*/',
      options: ['always', {
        exceptions: ['+', '-'],
      }],
    },

    // markers - line comments
    {
      code: '//!< docblock style comment',
      options: ['always', {
        markers: ['/', '!<'],
      }],
    },
    {
      code: '//!< docblock style comment',
      options: ['always', {
        line: { markers: ['/', '!<'] },
      }],
    },
    {
      code: '//----\n// a comment\n//----\n/// xmldoc style comment\n//!< docblock style comment',
      options: ['always', {
        exceptions: ['-'],
        markers: ['/', '!<'],
      }],
    },
    {
      code: '/* x*/',
      options: ['always', {
        markers: ['/', '!<'],
      }],
    },
    {
      code: '///xmldoc style comment',
      options: ['never', {
        markers: ['/', '!<'],
      }],
    },

    // markers - block comments
    {
      code: 'var a = 1; /*# This is an example of a marker in a block comment\nsubsequent lines do not count*/',
      options: ['always', {
        markers: ['#'],
      }],
    },
    {
      code: '/*!\n *comment\n */',
      options: ['always', { markers: ['!'] }],
    },
    {
      code: '/*!\n *comment\n */',
      options: ['always', { block: { markers: ['!'] } }],
    },
    {
      code: '/**\n *jsdoc\n */',
      options: ['always', { markers: ['*'] }],
    },
    {
      code: '/*global ABC*/',
      options: ['always', { markers: ['global'] }],
    },
    {
      code: '/*eslint eqeqeq:0, curly: 2*/',
      options: ['always', { markers: ['eslint'] }],
    },
    {
      code: '/*eslint-disable no-alert, no-console */\nalert()\nconsole.log()\n/*eslint-enable no-alert */',
      options: ['always', { markers: ['eslint-enable', 'eslint-disable'] }],
    },

    // misc. variations
    {
      code: validShebangProgram,
      options: ['always'],
    },
    {
      code: validShebangProgram,
      options: ['never'],
    },
    {
      code: '//',
      options: ['always'],
    },
    {
      code: '//\n',
      options: ['always'],
    },
    {
      code: '// space only at start; valid since balanced doesn\'t apply to line comments',
      options: ['always', { block: { balanced: true } }],
    },
    {
      code: '//space only at end; valid since balanced doesn\'t apply to line comments ',
      options: ['never', { block: { balanced: true } }],
    },

    // block comments
    {
      code: 'var a = 1; /* A valid comment starting with space */',
      options: ['always'],
    },
    {
      code: 'var a = 1; /*A valid comment NOT starting with space */',
      options: ['never'],
    },
    {
      code: 'function foo(/* height */a) { \n }',
      options: ['always'],
    },
    {
      code: 'function foo(/*height */a) { \n }',
      options: ['never'],
    },
    {
      code: 'function foo(a/* height */) { \n }',
      options: ['always'],
    },
    {
      code: '/*\n * Test\n */',
      options: ['always'],
    },
    {
      code: '/*\n *Test\n */',
      options: ['never'],
    },
    {
      code: '/*     \n *Test\n */',
      options: ['always'],
    },
    {
      code: '/*\r\n *Test\r\n */',
      options: ['never'],
    },
    {
      code: '/*     \r\n *Test\r\n */',
      options: ['always'],
    },
    {
      code: '/**\n *jsdoc\n */',
      options: ['always'],
    },
    {
      code: '/**\r\n *jsdoc\r\n */',
      options: ['always'],
    },
    {
      code: '/**\n *jsdoc\n */',
      options: ['never'],
    },
    {
      code: '/**   \n *jsdoc \n */',
      options: ['always'],
    },

    // balanced block comments
    {
      code: 'var a = 1; /* comment */',
      options: ['always', { block: { balanced: true } }],
    },
    {
      code: 'var a = 1; /*comment*/',
      options: ['never', { block: { balanced: true } }],
    },
    {
      code: 'function foo(/* height */a) { \n }',
      options: ['always', { block: { balanced: true } }],
    },
    {
      code: 'function foo(/*height*/a) { \n }',
      options: ['never', { block: { balanced: true } }],
    },
    {
      code: 'var a = 1; /*######*/',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
        block: { balanced: true },
      }],
    },
    {
      code: '/*****************\n * A comment\n *****************/',
      options: ['always', {
        exceptions: ['*'],
        block: { balanced: true },
      }],
    },
    {
      code: '/*! comment */',
      options: ['always', { markers: ['!'], block: { balanced: true } }],
    },
    {
      code: '/*!comment*/',
      options: ['never', { markers: ['!'], block: { balanced: true } }],
    },
    {
      code: '/*!\n *comment\n */',
      options: ['always', { markers: ['!'], block: { balanced: true } }],
    },
    {
      code: '/*global ABC */',
      options: ['always', { markers: ['global'], block: { balanced: true } }],
    },

    // markers & exceptions
    {
      code: '///--------\r\n/// test\r\n///--------',
      options: ['always', { markers: ['/'], exceptions: ['-'] }],
    },
    {
      code: '///--------\r\n/// test\r\n///--------\r\n/* blah */',
      options: ['always', { markers: ['/'], exceptions: ['-'], block: { markers: [] } }],
    },
    {
      code: '/*** */',
      options: ['always', { exceptions: ['*'] }],
    },

    // ignore marker-only comments, https://github.com/eslint/eslint/issues/12036
    {
      code: '//#endregion',
      options: ['always', { line: { markers: ['#endregion'] } }],
    },
    {
      code: '/*foo*/',
      options: ['always', { block: { markers: ['foo'] } }],
    },
    {
      code: '/*foo*/',
      options: ['always', { block: { markers: ['foo'], balanced: true } }],
    },
    {
      code: '/*foo*/ /*bar*/',
      options: ['always', { markers: ['foo', 'bar'] }],
    },
    {
      code: '//foo\n//bar',
      options: ['always', { markers: ['foo', 'bar'] }],
    },
    {
      code: '/* foo */',
      options: ['never', { markers: [' foo '] }],
    },
    {
      code: '// foo ',
      options: ['never', { markers: [' foo '] }],
    },
    {
      code: '//*', // "*" is a marker by default
      options: ['always'],
    },
    {
      code: '/***/', // "*" is a marker by default
      options: ['always'],
    },

    // ignore typescript triple-slash directive
    {
      code: '/// <reference types="node" />',
      options: ['always'],
    },
    {
      code: '/// <reference path="path/to/file" />',
      options: ['always'],
    },
    {
      code: '/// <amd-module name="moduleName" />',
      options: ['always'],
    },
  ],

  invalid: [
    {
      code: '//An invalid comment NOT starting with space\nvar a = 1;',
      output: '// An invalid comment NOT starting with space\nvar a = 1;',
      options: ['always'],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '//' },
      }],
    },
    {
      code: '// An invalid comment starting with space\nvar a = 2;',
      output: '//An invalid comment starting with space\nvar a = 2;',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: { refChar: '//' },
      }],
    },
    {
      code: '//   An invalid comment starting with tab\nvar a = 2;',
      output: '//An invalid comment starting with tab\nvar a = 2;',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: { refChar: '//' },
      }],
    },
    {

      /**
       * note that the first line in the comment is not a valid exception
       * block pattern because of the minus sign at the end of the line:
       * `//*********************-`
       */
      code: '//*********************-\n// Comment Block 3\n//***********************',
      output: '//* ********************-\n// Comment Block 3\n//***********************',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
      errors: [{
        messageId: 'expectedExceptionAfter',
        data: { refChar: '//*' },
      }],
    },
    {
      code: '//-=-=-=-=-=-=\n// A comment\n//-=-=-=-=-=-=',
      output: '// -=-=-=-=-=-=\n// A comment\n// -=-=-=-=-=-=',
      options: ['always', {
        exceptions: ['-', '=', '*', '#', '!@#'],
      }],
      errors: [
        {
          messageId: 'expectedExceptionAfter',
          data: { refChar: '//' },
        },
        {
          messageId: 'expectedExceptionAfter',
          data: { refChar: '//' },
        },
      ],
    },
    {
      code: '//!<docblock style comment',
      output: '//!< docblock style comment',
      options: ['always', {
        markers: ['/', '!<'],
      }],
      errors: 1,
    },
    {
      code: '//!< docblock style comment',
      output: '//!<docblock style comment',
      options: ['never', {
        markers: ['/', '!<'],
      }],
      errors: 1,
    },
    {
      code: 'var a = 1; /* A valid comment starting with space */',
      output: 'var a = 1; /*A valid comment starting with space */',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: 'var a = 1; /*######*/',
      output: 'var a = 1; /* ######*/',
      options: ['always', {
        exceptions: ['-', '=', '*', '!@#'],
      }],
      errors: [{
        messageId: 'expectedExceptionAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: 'var a = 1; /*A valid comment NOT starting with space */',
      output: 'var a = 1; /* A valid comment NOT starting with space */',
      options: ['always'],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: 'function foo(/* height */a) { \n }',
      output: 'function foo(/*height */a) { \n }',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: 'function foo(/*height */a) { \n }',
      output: 'function foo(/* height */a) { \n }',
      options: ['always'],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: 'function foo(a/*height */) { \n }',
      output: 'function foo(a/* height */) { \n }',
      options: ['always'],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '/*     \n *Test\n */',
      output: '/*\n *Test\n */',
      options: ['never'],
      errors: [{
        messageId: 'unexpectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '//-----------------------\n// A comment\n//-----------------------',
      output: '// -----------------------\n// A comment\n// -----------------------',
      options: ['always', {
        block: { exceptions: ['-', '=', '*', '#', '!@#'] },
      }],
      errors: [
        { messageId: 'expectedSpaceAfter', data: { refChar: '//' } },
        { messageId: 'expectedSpaceAfter', data: { refChar: '//' } },
      ],
    },
    {
      code: 'var a = 1; /*######*/',
      output: 'var a = 1; /* ######*/',
      options: ['always', {
        line: { exceptions: ['-', '=', '*', '#', '!@#'] },
      }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '//!< docblock style comment',
      output: '// !< docblock style comment',
      options: ['always', {
        block: { markers: ['/', '!<'] },
      }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '//' },
      }],
    },
    {
      code: '/*!\n *comment\n */',
      output: '/* !\n *comment\n */',
      options: ['always', { line: { markers: ['!'] } }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '///--------\r\n/// test\r\n///--------\r\n/*/ blah *//*-----*/',
      output: '///--------\r\n/// test\r\n///--------\r\n/* / blah *//*-----*/',
      options: ['always', { markers: ['/'], exceptions: ['-'], block: { markers: [] } }],
      errors: [{
        messageId: 'expectedExceptionAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '///--------\r\n/// test\r\n///--------\r\n/*/ blah */ /*-----*/',
      output: '///--------\r\n/// test\r\n///--------\r\n/* / blah */ /* -----*/',
      options: ['always', { line: { markers: ['/'], exceptions: ['-'] } }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
        line: 4,
        column: 1,
      }, {
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
        line: 4,
        column: 13,
      }],
    },

    // balanced block comments
    {
      code: 'var a = 1; /* A balanced comment starting with space*/',
      output: 'var a = 1; /* A balanced comment starting with space */',
      options: ['always', { block: { balanced: true } }],
      errors: [{
        messageId: 'expectedSpaceBefore',
        data: { refChar: '/**' },
      }],
    },
    {
      code: 'var a = 1; /*A balanced comment NOT starting with space */',
      output: 'var a = 1; /*A balanced comment NOT starting with space*/',
      options: ['never', { block: { balanced: true } }],
      errors: [{
        messageId: 'unexpectedSpaceBefore',
        data: { refChar: '*/' },
      }],
    },
    {
      code: 'function foo(/* height*/a) { \n }',
      output: 'function foo(/* height */a) { \n }',
      options: ['always', { block: { balanced: true } }],
      errors: [{
        messageId: 'expectedSpaceBefore',
        data: { refChar: '/**' },
      }],
    },
    {
      code: 'function foo(/*height */a) { \n }',
      output: 'function foo(/*height*/a) { \n }',
      options: ['never', { block: { balanced: true } }],
      errors: [{
        messageId: 'unexpectedSpaceBefore',
        data: { refChar: '*/' },
      }],
    },
    {
      code: '/*! comment*/',
      output: '/*! comment */',
      options: ['always', { markers: ['!'], block: { balanced: true } }],
      errors: [{
        messageId: 'expectedSpaceBefore',
        data: { refChar: '/**' },
      }],
    },
    {
      code: '/*!comment */',
      output: '/*!comment*/',
      options: ['never', { markers: ['!'], block: { balanced: true } }],
      errors: [{
        messageId: 'unexpectedSpaceBefore',
        data: { refChar: '*/' },
      }],
    },

    // not a marker-only comment, regression tests for https://github.com/eslint/eslint/issues/12036
    {
      code: '//#endregionfoo',
      output: '//#endregion foo',
      options: ['always', { line: { markers: ['#endregion'] } }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '//#endregion' },
      }],
    },
    {
      code: '/*#endregion*/',
      output: '/* #endregion*/', // not an allowed marker for block comments
      options: ['always', { line: { markers: ['#endregion'] } }],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/*' },
      }],
    },
    {
      code: '/****/',
      output: '/** **/',
      options: ['always'],
      errors: [{
        messageId: 'expectedSpaceAfter',
        data: { refChar: '/**' },
      }],
    },
    {
      code: '/****/',
      output: '/** * */',
      options: ['always', { block: { balanced: true } }],
      errors: [
        {
          messageId: 'expectedSpaceAfter',
          data: { refChar: '/**' },
        },
        {
          messageId: 'expectedSpaceBefore',
          data: { refChar: '*/' },
        },
      ],
    },
    {
      code: '/* foo */',
      output: '/*foo*/',
      options: ['never', { block: { markers: ['foo'], balanced: true } }], // not " foo "
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { refChar: '/*' },
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { refChar: '*/' },
        },
      ],
    },
  ],
});

/**
 * ======================== spaced-comment — KNOWN GAPS =======================
 *
 * None. Every upstream fixture is source-with-comments that parses cleanly under
 * the ts-go parser. The only upstream-specific metadata dropped was
 * `linterOptions: { reportUnusedDisableDirectives: false }` on one valid case
 * (the `eslint-disable`/`eslint-enable` marker fixture), which controls ESLint's
 * own unused-directive reporting and not the spaced-comment rule — rslint reports
 * only under the rule id, so dropping it changes nothing (verified: 0
 * spaced-comment diagnostics). No octal / sloppy-mode / Babel / Flow /
 * import-attribute cases exist in this rule's tests.
 * ============================================================================
 */
