/**
 * @fileoverview Tests for new-parens rule.
 * @author Ilya Volodin
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/new-parens/new-parens.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('new-parens', null as never, { valid, invalid })`
 *  - The local error helpers (`error` / `neverError`) are inlined to their final
 *    `{ messageId: 'missing' }` / `{ messageId: 'unnecessary' }`.
 *  - `parser: tsParser` (the `@typescript-eslint/parser` override on one valid
 *    case) dropped — rslint always parses TS; the case is valid TS as written.
 *
 * The whole upstream test file is a single `run()` block (no trailing skipBabel
 * block). No Babel/Flow, external-fixture (`readFileSync`), or suggestion cases
 * exist. The `._css_` / `._json_` / `._markdown_` test files don't exist for
 * this rule. Every fixture is valid TS under ts-go's strict/module semantics,
 * so nothing was isolated into KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('new-parens', null as never, {
  valid: [
    // Default (Always)
    'var a = new Date();',
    'var a = new Date(function() {});',
    'var a = new (Date)();',
    'var a = new ((Date))();',
    'var a = (new Date());',
    'var a = new foo.Bar();',
    'var a = (new Foo()).bar;',
    {
      code: 'new Storage<RootState>(\'state\');',
    },

    // Explicit Always
    { code: 'var a = new Date();', options: ['always'] },
    { code: 'var a = new foo.Bar();', options: ['always'] },
    { code: 'var a = (new Foo()).bar;', options: ['always'] },

    // Never
    { code: 'var a = new Date;', options: ['never'] },
    { code: 'var a = new Date(function() {});', options: ['never'] },
    { code: 'var a = new (Date);', options: ['never'] },
    { code: 'var a = new ((Date));', options: ['never'] },
    { code: 'var a = (new Date);', options: ['never'] },
    { code: 'var a = new foo.Bar;', options: ['never'] },
    { code: 'var a = (new Foo).bar;', options: ['never'] },
    { code: 'var a = new Person(\'Name\')', options: ['never'] },
    { code: 'var a = new Person(\'Name\', 12)', options: ['never'] },
    { code: 'var a = new ((Person))(\'Name\');', options: ['never'] },
  ],
  invalid: [
    // Default (Always)
    {
      code: 'var a = new Date;',
      output: 'var a = new Date();',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new Date',
      output: 'var a = new Date()',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new (Date);',
      output: 'var a = new (Date)();',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new (Date)',
      output: 'var a = new (Date)()',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = (new Date)',
      output: 'var a = (new Date())',
      errors: [{ messageId: 'missing' }],
    },
    {
      // This `()` is `CallExpression`'s. This is a call of the result of `new Date`.
      code: 'var a = (new Date)()',
      output: 'var a = (new Date())()',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new foo.Bar;',
      output: 'var a = new foo.Bar();',
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = (new Foo).bar;',
      output: 'var a = (new Foo()).bar;',
      errors: [{ messageId: 'missing' }],
    },

    // Explicit always
    {
      code: 'var a = new Date;',
      output: 'var a = new Date();',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new foo.Bar;',
      output: 'var a = new foo.Bar();',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = (new Foo).bar;',
      output: 'var a = (new Foo()).bar;',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var a = new new Foo()',
      output: 'var a = new new Foo()()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // Never
    {
      code: 'var a = new Date();',
      output: 'var a = (new Date);',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = new Date()',
      output: 'var a = (new Date)',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = new (Date)();',
      output: 'var a = (new (Date));',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = new (Date)()',
      output: 'var a = (new (Date))',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = (new Date())',
      output: 'var a = ((new Date))',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = (new Date())()',
      output: 'var a = ((new Date))()',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = new foo.Bar();',
      output: 'var a = (new foo.Bar);',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = (new Foo()).bar;',
      output: 'var a = ((new Foo)).bar;',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: 'var a = new new Foo()',
      output: 'var a = new (new Foo)',
      options: ['never'],
      errors: [{ messageId: 'unnecessary' }],
    },
  ],
});
