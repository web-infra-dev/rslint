/**
 * @fileoverview Tests for function-call-spacing rule.
 * @author Matt DuVall <http://www.mattduvall.com>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0 (both test files merged):
 *   packages/eslint-plugin/rules/function-call-spacing/function-call-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/function-call-spacing/function-call-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('function-call-spacing', null as never, { valid, invalid })`
 *  - `lang: 'js'` dropped; `parserOptions` (ecmaVersion) dropped — rslint resolves via tsconfig.
 *  - The `_ts_` file's spread helper
 *      `...['f<a> ()', ...].map<ValidTestCase>(code => ({ code, options: ['always'] }))`
 *    is evaluated to its nine explicit `{ code, options: ['always'] }` valid cases.
 *
 * Every `code` string is a single-/double-quoted JS string literal whose escapes
 * (`\n`, `\r`, `\u2028`, `\u2029`) are real characters; they are carried over
 * byte-for-byte (no `$`/unindent tag is used upstream). The `_ts_` file's
 * `invalid` array is empty. There are no Babel/Flow cases and no octal /
 * leading-zero fixtures, so nothing is isolated into KNOWN GAPS for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('function-call-spacing', null as never, {
  valid: [
    // ---- from function-call-spacing._js_.test.ts ----
    // default ("never")
    'f();',
    'f(a, b);',
    'f.b();',
    'f.b().c();',
    'f()()',
    '(function() {}())',
    'var f = new Foo()',
    'var f = new Foo',
    'f( (0) )',
    '( f )( 0 )',
    '( (f) )( (0) )',
    '( f()() )(0)',
    '(function(){ if (foo) { bar(); } }());',
    'f(0, (1))',
    'describe/**/(\'foo\', function () {});',
    'new (foo())',
    'import(source)',
    'f?.(a, b);',
    'f?.b()?.c();',
    'f.b?.();',
    'f.b?.().c();',
    '(function() {}?.())',
    'f?.( (0) )',
    '(function(){ if (foo) { bar(); } }?.());',
    'f?.(0, (1))',
    'describe/**/?.(\'foo\', function () {});',
    'describe?./**/(\'foo\', function () {});',
    '( f )?.( 0 )',
    '( (f) )?.( (0) )',
    '( f?.()() )(0)',
    '( f()?.() )(0)',
    '( f?.()?.() )(0)',
    '( f?.()() )?.(0)',
    '( f()?.() )?.(0)',
    '( f?.()?.() )?.(0)',

    // "never"
    {
      code: 'f();',
      options: ['never'],
    },
    {
      code: 'f(a, b);',
      options: ['never'],
    },
    {
      code: 'f.b();',
      options: ['never'],
    },
    {
      code: 'f.b().c();',
      options: ['never'],
    },
    {
      code: 'f()()',
      options: ['never'],
    },
    {
      code: '(function() {}())',
      options: ['never'],
    },
    {
      code: 'var f = new Foo()',
      options: ['never'],
    },
    {
      code: 'var f = new Foo',
      options: ['never'],
    },
    {
      code: 'f( (0) )',
      options: ['never'],
    },
    {
      code: '( f )( 0 )',
      options: ['never'],
    },
    {
      code: '( (f) )( (0) )',
      options: ['never'],
    },
    {
      code: '( f()() )(0)',
      options: ['never'],
    },
    {
      code: '(function(){ if (foo) { bar(); } }());',
      options: ['never'],
    },
    {
      code: 'f(0, (1))',
      options: ['never'],
    },
    {
      code: 'describe/**/(\'foo\', function () {});',
      options: ['never'],
    },
    {
      code: 'new (foo())',
      options: ['never'],
    },
    {
      code: 'import(source)',
      options: ['never'],
    },

    // "always"
    {
      code: 'f ();',
      options: ['always'],
    },
    {
      code: 'f (a, b);',
      options: ['always'],
    },
    {
      code: 'f.b ();',
      options: ['always'],
    },
    {
      code: 'f.b ().c ();',
      options: ['always'],
    },
    {
      code: 'f () ()',
      options: ['always'],
    },
    {
      code: '(function() {} ())',
      options: ['always'],
    },
    {
      code: 'var f = new Foo ()',
      options: ['always'],
    },
    {
      code: 'var f = new Foo',
      options: ['always'],
    },
    {
      code: 'f ( (0) )',
      options: ['always'],
    },
    {
      code: 'f (0) (1)',
      options: ['always'],
    },
    {
      code: '(f) (0)',
      options: ['always'],
    },
    {
      code: '( f ) ( 0 )',
      options: ['always'],
    },
    {
      code: '( (f) ) ( (0) )',
      options: ['always'],
    },
    {
      code: 'f ();\n t   ();',
      options: ['always'],
    },
    {
      code: 'import (source)',
      options: ['always'],
    },
    {
      code: 'f?.b ();',
      options: ['always'],
    },
    {
      code: 'f?.b ()?.c ();',
      options: ['always'],
    },
    {
      code: 'f ?. ();',
      options: ['always'],
    },

    // "always", "allowNewlines": true
    {
      code: 'f\n();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f.b \n ();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\n() ().b \n()\n ()',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'var f = new Foo\n();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f// comment\n()',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f // comment\n ()',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\n/*\n*/\n()',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\r();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\u2028();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\u2029();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\r\n();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'import\n(source)',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f?.b \n ();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\n() ()?.b \n()\n ()',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f ?. ();',
      options: ['always', { allowNewlines: true }],
    },
    {
      code: 'f\n?.\n();',
      options: ['always', { allowNewlines: true }],
    },

    // Optional chaining
    {
      code: 'func?.()',
      options: ['never'],
    },
    {
      code: 'func ?.()',
      options: ['always', { optionalChain: { after: false } }],
    },
    {
      code: 'func?. ()',
      options: ['always', { optionalChain: { before: false } }],
    },
    {
      code: 'func ?. ()',
      options: ['always'],
    },
    {
      code: 'f?.\n();',
      options: ['always', { allowNewlines: true, optionalChain: { before: false } }],
    },
    {
      code: 'f\n?.();',
      options: ['always', { allowNewlines: true, optionalChain: { after: false } }],
    },

    // ---- from function-call-spacing._ts_.test.ts ----
    // default ("never")
    'f<a>()',
    'f<a>(b, b)',
    'f.b<a>(b, b)',
    '(function<T>() {}<a>())',
    '((function<T>() {})<a>())',
    '( f )<a>( 0 )',
    '( (f) )<a>( (0) )',
    '( f()() )<a>(0)',

    // optional call
    'f?.<a>()',
    'f?.<a>(b, b)',
    'f.b?.<a>(b, b)',
    'f?.b<a>(b, b)',
    'f?.b?.<a>(b, b)',
    '(function<T>() {}?.<a>())',
    '((function<T>() {})<a>())',
    '( f )?.<a>( 0 )',
    '( (f) )?.<a>( (0) )',
    '( f()() )?.<a>(0)',

    // (from the `_ts_` spread helper, each with options: ['always'])
    { code: 'f<a> ()', options: ['always'] },
    { code: 'f<a> (b, b)', options: ['always'] },
    { code: 'f.b<a> (b, b)', options: ['always'] },
    { code: '(function<T>() {}<a> ())', options: ['always'] },
    { code: '((function<T>() {})<a> ())', options: ['always'] },
    { code: '( f )<a> ( 0 )', options: ['always'] },
    { code: '( (f) )<a> ( (0) )', options: ['always'] },
    { code: '( f () )<a> (0)', options: ['always'] },

    // optional call
    { code: 'f?.b<a> (b, b)', options: ['always'] },
  ],
  invalid: [
    // ---- from function-call-spacing._js_.test.ts ----
    // default ("never")
    {
      code: 'f ();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f (a, b);',
      output: 'f(a, b);',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f.b ();',
      output: 'f.b();',
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          column: 4,
          line: 1,
          endColumn: 5,
          endLine: 1,
        },
      ],
    },
    {
      code: 'f.b().c ();',
      output: 'f.b().c();',
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          column: 8,
          line: 1,
          endColumn: 9,
          endLine: 1,
        },
      ],
    },
    {
      code: 'f() ()',
      output: 'f()()',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: '(function() {} ())',
      output: '(function() {}())',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'var f = new Foo ()',
      output: 'var f = new Foo()',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f ( (0) )',
      output: 'f( (0) )',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f(0) (1)',
      output: 'f(0)(1)',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: '(f) (0)',
      output: '(f)(0)',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f ();\n t   ();',
      output: 'f();\n t();',
      errors: [
        { messageId: 'unexpectedWhitespace' },
        { messageId: 'unexpectedWhitespace' },
      ],
    },
    {
      code: 'import (source);',
      output: 'import(source);',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },

    {
      code: 'f\n();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f\r();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f\u2028();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f\u2029();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f\r\n();',
      output: 'f();',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'import\n(source);',
      output: 'import(source);',
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },

    // "never"
    {
      code: 'f ();',
      output: 'f();',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f (a, b);',
      output: 'f(a, b);',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f.b  ();',
      output: 'f.b();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          column: 4,
          line: 1,
          endColumn: 6,
          endLine: 1,
        },
      ],
    },
    {
      code: 'f.b().c ();',
      output: 'f.b().c();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          column: 8,
          line: 1,
          endColumn: 9,
          endLine: 1,
        },
      ],
    },
    {
      code: 'f() ()',
      output: 'f()()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: '(function() {} ())',
      output: '(function() {}())',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'var f = new Foo ()',
      output: 'var f = new Foo()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f ( (0) )',
      output: 'f( (0) )',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f(0) (1)',
      output: 'f(0)(1)',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: '(f) (0)',
      output: '(f)(0)',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f ();\n t   ();',
      output: 'f();\n t();',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedWhitespace' },
        { messageId: 'unexpectedWhitespace' },
      ],
    },
    {
      code: 'import (source);',
      output: 'import(source);',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },

    {
      code: 'f\r();',
      output: 'f();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
        },
      ],
    },
    {
      code: 'f\u2028();',
      output: 'f();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
        },
      ],
    },
    {
      code: 'f\u2029();',
      output: 'f();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
        },
      ],
    },
    {
      code: 'f\r\n();',
      output: 'f();',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedWhitespace',
        },
      ],
    },

    // "always"
    {
      code: 'f();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f\n();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f(a, b);',
      output: 'f (a, b);',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f\n(a, b);',
      output: 'f (a, b);',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f.b();',
      output: 'f.b ();',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          column: 4,
          line: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: 'f.b\n();',
      output: 'f.b ();',
      options: ['always'],
      errors: [
        {
          messageId: 'unexpectedNewline',
          column: 4,
          line: 1,
          endColumn: 1,
          endLine: 2,
        },
      ],
    },
    {
      code: 'f.b().c ();',
      output: 'f.b ().c ();',
      options: ['always'],
      errors: [{ messageId: 'missing', column: 4 }],
    },
    {
      code: 'f.b\n().c ();',
      output: 'f.b ().c ();',
      options: ['always'],
      errors: [
        {
          messageId: 'unexpectedNewline',
          column: 4,
          line: 1,
          endColumn: 1,
          endLine: 2,
        },
      ],
    },
    {
      code: 'f() ()',
      output: 'f () ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f\n() ()',
      output: 'f () ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f\n()()',
      output: 'f () ()',
      options: ['always'],
      errors: [
        { messageId: 'unexpectedNewline' },
        { messageId: 'missing' },
      ],
    },
    {
      code: '(function() {}())',
      output: '(function() {} ())',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var f = new Foo()',
      output: 'var f = new Foo ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f( (0) )',
      output: 'f ( (0) )',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f(0) (1)',
      output: 'f (0) (1)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(f)(0)',
      output: '(f) (0)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'import(source);',
      output: 'import (source);',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f();\n t();',
      output: 'f ();\n t ();',
      options: ['always'],
      errors: [
        { messageId: 'missing' },
        { messageId: 'missing' },
      ],
    },
    {
      code: 'f\r();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f\u2028();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f\u2029();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f\r\n();',
      output: 'f ();',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },

    // "always", "allowNewlines": true
    {
      code: 'f();',
      output: 'f ();',
      options: ['always', { allowNewlines: true }],
      errors: [
        { messageId: 'missing' },
      ],
    },
    {
      code: 'f(a, b);',
      output: 'f (a, b);',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f.b();',
      output: 'f.b ();',
      options: ['always', { allowNewlines: true }],
      errors: [
        {
          messageId: 'missing',
          column: 4,
        },
      ],
    },
    {
      code: 'f.b().c ();',
      output: 'f.b ().c ();',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing', column: 4 }],
    },
    {
      code: 'f() ()',
      output: 'f () ()',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(function() {}())',
      output: '(function() {} ())',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var f = new Foo()',
      output: 'var f = new Foo ()',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f( (0) )',
      output: 'f ( (0) )',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f(0) (1)',
      output: 'f (0) (1)',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(f)(0)',
      output: '(f) (0)',
      options: ['always', { allowNewlines: true }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'f();\n t();',
      output: 'f ();\n t ();',
      options: ['always', { allowNewlines: true }],
      errors: [
        { messageId: 'missing' },
        { messageId: 'missing' },
      ],
    },
    {
      code: 'f\n?.();',
      output: 'f\n?. ();',
      options: ['always', { allowNewlines: true }],
      errors: [
        { messageId: 'missing' },
      ],
    },
    {
      code: 'f?.\n();',
      output: 'f ?.\n();',
      options: ['always', { allowNewlines: true }],
      errors: [
        { messageId: 'missing' },
      ],
    },
    {
      code: 'f    ();',
      output: 'f();',
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'f\n ();',
      output: 'f();',
      errors: [
        {
          messageId: 'unexpectedWhitespace',
          line: 1,
          column: 2,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'fn();',
      output: 'fn ();',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'fnn\n (a, b);',
      output: 'fnn (a, b);',
      options: ['always'],
      errors: [
        {
          messageId: 'unexpectedNewline',
          line: 1,
          column: 4,
          endLine: 2,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'f /*comment*/ ()',
      output: null, // Don't remove comments
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f /*\n*/ ()',
      output: null, // Don't remove comments
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'var f = new Foo /* comment */ ()',
      output: null, // Don't remove comments
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'import\n//comment\n(source)',
      output: null, // Don't remove comments
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'f\n//comment\n();',
      output: null,
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'f/*comment*/()',
      output: 'f/*comment*/ ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // Optional chaining
    {
      code: 'func ?.()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func?. ()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func ?. ()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func\n?.()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func\n?.\n()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func ?.\n()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func\n?. ()',
      output: 'func?.()',
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func\n//comment\n?.()',
      output: null, // Don't remove comments
      options: ['never'],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func?.()',
      output: 'func ?. ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func ?.()',
      output: 'func ?. ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func?. ()',
      output: 'func ?. ()',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func\n  ?.()',
      output: 'func   ?. ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func?.\n  ()',
      output: 'func ?.   ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func  ?.\n  ()',
      output: 'func  ?.   ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func\n?.\n()',
      output: 'func ?. ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func\n?. ()',
      output: 'func ?. ()',
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func\n /*comment*/ ?.()',
      output: null, // Don't remove comments
      options: ['always'],
      errors: [{ messageId: 'unexpectedNewline' }],
    },

    {
      code: 'func?.()',
      output: 'func?. ()',
      options: ['always', { optionalChain: { before: false } }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func\n?.()',
      output: 'func?. ()',
      options: ['always', { optionalChain: { before: false } }],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func ?.\n()',
      output: 'func?. ()',
      options: ['always', { optionalChain: { before: false } }],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func?.()',
      output: 'func?. ()',
      options: ['always', { allowNewlines: true, optionalChain: { before: false } }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func\n?.()',
      output: 'func?. ()',
      options: ['always', { allowNewlines: true, optionalChain: { before: false } }],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func ?.\n()',
      output: 'func?.\n()',
      options: ['always', { allowNewlines: true, optionalChain: { before: false } }],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },

    {
      code: 'func?.()',
      output: 'func ?.()',
      options: ['always', { optionalChain: { after: false } }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func?.\n()',
      output: 'func ?.()',
      options: ['always', { optionalChain: { after: false } }],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func\n?. ()',
      output: 'func ?.()',
      options: ['always', { optionalChain: { after: false } }],
      errors: [{ messageId: 'unexpectedNewline' }],
    },
    {
      code: 'func?.()',
      output: 'func ?.()',
      options: ['always', { allowNewlines: true, optionalChain: { after: false } }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'func\n?. ()',
      output: 'func\n?.()',
      options: ['always', { allowNewlines: true, optionalChain: { after: false } }],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
    {
      code: 'func?.\n()',
      output: 'func ?.()',
      options: ['always', { allowNewlines: true, optionalChain: { after: false } }],
      errors: [{ messageId: 'unexpectedWhitespace' }],
    },
  ],
});
