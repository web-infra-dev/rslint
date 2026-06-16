/**
 * @fileoverview Tests for template-tag-spacing rule.
 * @author Jonathan Wilsson
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/template-tag-spacing/template-tag-spacing.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('template-tag-spacing', null as never, { valid, invalid })`
 *    (the `name` / `rule` / `#test` import are dropped).
 *  - No `$` unindent tag, no spread/helper errors, no `parserOptions`, no `type`
 *    fields, and no suggestions exist in the upstream cases — nothing to evaluate
 *    or drop on those grounds.
 *
 * The upstream file has a single `run()` block (no skipBabel / second block) and
 * no `._css_` / `._json_` / `._markdown_` companion files.
 *
 * Alignment status (verified against the live rslint CLI):
 *  - All 28 valid cases report 0 diagnostics.
 *  - All 35 invalid cases match upstream on diagnostic count, message,
 *    line/column/endLine/endColumn, and autofix `output` (the 2 `output: null`
 *    cases — multi-line tag/literal separations the rule refuses to autofix —
 *    are asserted to leave the source unchanged).
 * No rslint<->upstream gap surfaced, so there is no KNOWN GAPS section.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('template-tag-spacing', null as never, {
  valid: [
    'tag`name`',
    { code: 'tag`name`', options: ['never'] },
    { code: 'tag `name`', options: ['always'] },
    'tag`hello ${name}`',
    { code: 'tag`hello ${name}`', options: ['never'] },
    { code: 'tag `hello ${name}`', options: ['always'] },
    'tag/*here\'s a comment*/`Hello world`',
    { code: 'tag/*here\'s a comment*/`Hello world`', options: ['never'] },
    { code: 'tag /*here\'s a comment*/`Hello world`', options: ['always'] },
    { code: 'tag/*here\'s a comment*/ `Hello world`', options: ['always'] },
    'new tag`name`',
    { code: 'new tag`name`', options: ['never'] },
    { code: 'new tag `name`', options: ['always'] },
    'new tag`hello ${name}`',
    { code: 'new tag`hello ${name}`', options: ['never'] },
    { code: 'new tag `hello ${name}`', options: ['always'] },
    '(tag)`name`',
    { code: '(tag)`name`', options: ['never'] },
    { code: '(tag) `name`', options: ['always'] },
    '(tag)`hello ${name}`',
    { code: '(tag)`hello ${name}`', options: ['never'] },
    { code: '(tag) `hello ${name}`', options: ['always'] },
    'new (tag)`name`',
    { code: 'new (tag)`name`', options: ['never'] },
    { code: 'new (tag) `name`', options: ['always'] },
    'new (tag)`hello ${name}`',
    { code: 'new (tag)`hello ${name}`', options: ['never'] },
    { code: 'new (tag) `hello ${name}`', options: ['always'] },
  ],
  invalid: [
    {
      code: 'tag `name`',
      output: 'tag`name`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'tag `name`',
      output: 'tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'tag`name`',
      output: 'tag `name`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 4,
      }],
    },
    {
      code: 'tag /*here\'s a comment*/`Hello world`',
      output: 'tag/*here\'s a comment*/`Hello world`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 25,
      }],
    },
    {
      code: 'tag/*here\'s a comment*/ `Hello world`',
      output: 'tag/*here\'s a comment*/`Hello world`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 25,
      }],
    },
    {
      code: 'tag/*here\'s a comment*/`Hello world`',
      output: 'tag /*here\'s a comment*/`Hello world`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 24,
      }],
    },
    {
      code: 'tag // here\'s a comment \n`bar`',
      output: null,
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 2,
        endColumn: 1,
      }],
    },
    {
      code: 'tag // here\'s a comment \n`bar`',
      output: null,
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 2,
        endColumn: 1,
      }],
    },
    {
      code: 'tag `hello ${name}`',
      output: 'tag`hello ${name}`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'tag `hello ${name}`',
      output: 'tag`hello ${name}`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'tag`hello ${name}`',
      output: 'tag `hello ${name}`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 4,
      }],
    },
    {
      code: 'new tag `name`',
      output: 'new tag`name`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 8,
        endLine: 1,
        endColumn: 9,
      }],
    },
    {
      code: 'new tag `name`',
      output: 'new tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 8,
        endLine: 1,
        endColumn: 9,
      }],
    },
    {
      code: 'new tag`name`',
      output: 'new tag `name`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 8,
      }],
    },
    {
      code: 'new tag `hello ${name}`',
      output: 'new tag`hello ${name}`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 8,
        endLine: 1,
        endColumn: 9,
      }],
    },
    {
      code: 'new tag `hello ${name}`',
      output: 'new tag`hello ${name}`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 8,
        endLine: 1,
        endColumn: 9,
      }],
    },
    {
      code: 'new tag`hello ${name}`',
      output: 'new tag `hello ${name}`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 8,
      }],
    },
    {
      code: '(tag) `name`',
      output: '(tag)`name`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: '(tag) `name`',
      output: '(tag)`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: '(tag)`name`',
      output: '(tag) `name`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: '(tag) `hello ${name}`',
      output: '(tag)`hello ${name}`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: '(tag) `hello ${name}`',
      output: '(tag)`hello ${name}`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: '(tag)`hello ${name}`',
      output: '(tag) `hello ${name}`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'new (tag) `name`',
      output: 'new (tag)`name`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 10,
        endLine: 1,
        endColumn: 11,
      }],
    },
    {
      code: 'new (tag) `name`',
      output: 'new (tag)`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 10,
        endLine: 1,
        endColumn: 11,
      }],
    },
    {
      code: 'new (tag)`name`',
      output: 'new (tag) `name`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 10,
      }],
    },
    {
      code: 'new (tag) `hello ${name}`',
      output: 'new (tag)`hello ${name}`',
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 10,
        endLine: 1,
        endColumn: 11,
      }],
    },
    {
      code: 'new (tag) `hello ${name}`',
      output: 'new (tag)`hello ${name}`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 10,
        endLine: 1,
        endColumn: 11,
      }],
    },
    {
      code: 'new (tag)`hello ${name}`',
      output: 'new (tag) `hello ${name}`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 10,
      }],
    },
    {
      code: 'tag   `name`',
      output: 'tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: 'tag\n`name`',
      output: 'tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 2,
        endColumn: 1,
      }],
    },
    {
      code: 'tag \n  `name`',
      output: 'tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'tag\n\n`name`',
      output: 'tag`name`',
      options: ['never'],
      errors: [{
        messageId: 'unexpected',
        line: 1,
        column: 4,
        endLine: 3,
        endColumn: 1,
      }],
    },
    {
      code: 'foo\n  .bar`Hello world`',
      output: 'foo\n  .bar `Hello world`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 2,
        endColumn: 7,
      }],
    },
    {
      code: 'foo(\n  bar\n)`Hello world`',
      output: 'foo(\n  bar\n) `Hello world`',
      options: ['always'],
      errors: [{
        messageId: 'missing',
        line: 1,
        column: 1,
        endLine: 3,
        endColumn: 2,
      }],
    },
  ],
});
