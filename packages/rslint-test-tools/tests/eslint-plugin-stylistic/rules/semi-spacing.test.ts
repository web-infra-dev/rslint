/**
 * @fileoverview Tests for semi-spacing.
 * @author Mathias Schreck
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/semi-spacing/semi-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/semi-spacing/semi-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('semi-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint always parses at
 *    esnext under module semantics, where class fields, accessors, static blocks
 *    and `import`/`export` statements are all accepted.
 *  - `type` fields are not used by these upstream cases.
 *  - The rule's messageIds (`unexpectedWhitespaceBefore` / `unexpectedWhitespaceAfter`
 *    / `missingWhitespaceBefore` / `missingWhitespaceAfter`) take no `data`, so they
 *    map 1:1 to a fixed message; the RuleTester renders them from the plugin's meta.
 *
 * Neither upstream file imports the `$` unindent tag: every `code`/`output` is a
 * plain string (or, in the `._ts_` file, a plain multi-line template literal whose
 * literal indentation is preserved verbatim — it has no leading indentation). Each
 * file is a single `run()` block: no `if (!skipBabel)` block, no Babel/Flow cases,
 * no external-fixture `readFileSync` cases, no `suggestions`. Every invalid case
 * pins an explicit `errors` array — there are NO output-only cases. The `._css_` /
 * `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * KNOWN GAPS: none. Every fixture — including the sloppy/module `import`/`export`
 * statements, the `for(a ; ; )` empty-clause loops, class fields, `accessor`
 * members and `static {}` blocks, plus the `._ts_` type/interface/declare/abstract
 * cases — parses under rslint's ts-go parser and produces byte-identical
 * diagnostics and autofix output. No octal/`\8` syntax, no `assert` import
 * attributes, no ecmaVersion-normalization edge cases exist for this rule, so
 * nothing is isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('semi-spacing', null as never, {
  valid: [
    // ---- from semi-spacing._js_.test.ts ----
    'var a = \'b\';',
    'var a = \'b ; c\';',
    'var a = \'b\',\nc = \'d\';',
    'var a = function() {};',
    ';(function(){}());',
    'var a = \'b\'\n;(function(){}())',
    'debugger\n;(function(){}())',
    'while (true) { break; }',
    'while (true) { continue; }',
    'debugger;',
    'function foo() { return; }',
    'throw new Error(\'foo\');',
    'for (var i = 0; i < 10; i++) {}',
    'for (;;) {}',
    {
      code: 'var a = \'b\' ;',
      options: [{ before: true, after: true }],
    },
    {
      code: 'var a = \'b\';c = \'d\';',
      options: [{ before: false, after: false }],
    },
    {
      code: 'for (var i = 0 ;i < 10 ;i++) {}',
      options: [{ before: true, after: false }],
    },
    {
      code: 'for (var i = 0 ; i < 10 ; i++) {}',
      options: [{ before: true, after: true }],
    },

    // https://github.com/eslint/eslint/issues/3721
    'function foo(){return 2;}',
    'for(var i = 0; i < results.length;) {}',
    { code: 'function foo() { return 2; }', options: [{ after: false }] },
    { code: 'for ( var i = 0;i < results.length; ) {}', options: [{ after: false }] },

    'do {} while (true); foo',

    // Class fields
    {
      code: 'class C { foo; bar; method() {} }',
    },
    {
      code: 'class C { foo }',
    },
    'class C { accessor foo; accessor [bar]; }',
    'class C { accessor foo }',

    // Empty are ignored (`no-extra-semi` rule will remove those)
    'foo; ;;;;;;;;;',
    {
      code: 'class C { foo; ;;;;;;;;;; }',
    },

    // ---- from semi-spacing._ts_.test.ts ----
    'type Union = number | string;',
    `interface Foo { name: string; greet: () => string; }`,
    'declare function example();',
    'declare function example(): void;',
    `function foo(a: number): void;
function foo(a: string): void;
function foo(a: string | number): void {}`,
    'interface Example { new (): number; }',
    'abstract class Example { abstract prop: string; }',
    'abstract class Example { abstract method(): void; }',
    {
      code: 'type A = Record<string, number>;type B = string;',
      options: [{ before: false, after: false }],
    },
  ],
  invalid: [
    // ---- from semi-spacing._js_.test.ts ----
    {
      code: 'var a = \'b\'  ;',
      output: 'var a = \'b\';',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var a = \'b\' ;',
      output: 'var a = \'b\';',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var a = \'b\',\nc = \'d\' ;',
      output: 'var a = \'b\',\nc = \'d\';',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 2, column: 8 }],
    },
    {
      code: 'var a = function() {} ;',
      output: 'var a = function() {};',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 22 }],
    },
    {
      code: 'var a = function() {\n} ;',
      output: 'var a = function() {\n};',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 2, column: 2 }],
    },
    {
      code: '/^a$/.test(\'b\') ;',
      output: '/^a$/.test(\'b\');',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 16 }],
    },
    {
      code: ';(function(){}()) ;',
      output: ';(function(){}());',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 18 }],
    },
    {
      code: 'while (true) { break ; }',
      output: 'while (true) { break; }',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 21 }],
    },
    {
      code: 'while (true) { continue ; }',
      output: 'while (true) { continue; }',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 24 }],
    },
    {
      code: 'debugger ;',
      output: 'debugger;',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 9 }],
    },
    {
      code: 'function foo() { return ; }',
      output: 'function foo() { return; }',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 24 }],
    },
    {
      code: 'throw new Error(\'foo\') ;',
      output: 'throw new Error(\'foo\');',
      errors: [{ messageId: 'unexpectedWhitespaceBefore', line: 1, column: 23 }],
    },
    {
      code: 'for (var i = 0 ; i < 10 ; i++) {}',
      output: 'for (var i = 0; i < 10; i++) {}',
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 15 },
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 24 },
      ],
    },
    {
      code: 'var a = \'b\';c = \'d\';',
      output: 'var a = \'b\'; c = \'d\';',
      errors: [
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var a = \'b\';',
      output: 'var a = \'b\' ;',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missingWhitespaceBefore',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var a = \'b\'; c = \'d\';',
      output: 'var a = \'b\';c = \'d\';',
      options: [{ before: false, after: false }],
      errors: [
        {
          messageId: 'unexpectedWhitespaceAfter',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var a = \'b\';   c = \'d\';',
      output: 'var a = \'b\';c = \'d\';',
      options: [{ before: false, after: false }],
      errors: [
        {
          messageId: 'unexpectedWhitespaceAfter',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'for (var i = 0;i < 10;i++) {}',
      output: 'for (var i = 0; i < 10; i++) {}',
      errors: [
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'for (var i = 0; i < 10; i++) {}',
      output: 'for (var i = 0 ; i < 10 ; i++) {}',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missingWhitespaceBefore',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
        {
          messageId: 'missingWhitespaceBefore',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'for (var i = 0; i < 10; i++) {}',
      output: 'for (var i = 0;i < 10;i++) {}',
      options: [{ before: false, after: false }],
      errors: [
        { messageId: 'unexpectedWhitespaceAfter', line: 1, column: 16 },
        { messageId: 'unexpectedWhitespaceAfter', line: 1, column: 24 },
      ],
    },
    {
      code: 'import Foo from \'bar\' ;',
      output: 'import Foo from \'bar\';',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 22 },
      ],
    },
    {
      code: 'import * as foo from \'bar\' ;',
      output: 'import * as foo from \'bar\';',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 27 },
      ],
    },
    {
      code: 'var foo = 0; export {foo} ;',
      output: 'var foo = 0; export {foo};',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 26 },
      ],
    },
    {
      code: 'export * from \'foo\' ;',
      output: 'export * from \'foo\';',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 20 },
      ],
    },
    {
      code: 'export default foo ;',
      output: 'export default foo;',
      options: [{ before: false, after: true }],
      errors: [
        { messageId: 'unexpectedWhitespaceBefore', line: 1, column: 19 },
      ],
    },
    {
      code: 'while(foo) {continue   ;}',
      output: 'while(foo) {continue;}',
      options: [{ before: false, after: true }],
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'if(foo) {throw new Error()   ;  }',
      output: 'if(foo) {throw new Error();  }',
      options: [{ before: false, after: false }],
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'for(a ; ; );',
      output: 'for(a;; );',
      options: [{ before: false, after: false }],
      errors: [{
        messageId: 'unexpectedWhitespaceBefore',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }, {
        messageId: 'unexpectedWhitespaceAfter',
        line: 1,
        column: 8,
        endLine: 1,
        endColumn: 9,
      }],
    },
    {
      code: 'for(a ; \n ; );',
      output: 'for(a; \n ; );',
      options: [{ before: false, after: false }],
      errors: [{
        messageId: 'unexpectedWhitespaceBefore',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: 'do {} while (true) ;',
      output: 'do {} while (true);',
      errors: [{
        messageId: 'unexpectedWhitespaceBefore',
        line: 1,
        column: 19,
        endLine: 1,
        endColumn: 20,
      }],
    },
    {
      code: 'do {} while (true);foo',
      output: 'do {} while (true); foo',
      errors: [{
        messageId: 'missingWhitespaceAfter',
        line: 1,
        column: 19,
        endLine: 1,
        endColumn: 20,
      }],
    },
    {
      code: 'do {} while (true);  foo',
      output: 'do {} while (true) ;foo',
      options: [{ before: true, after: false }],
      errors: [{
        messageId: 'missingWhitespaceBefore',
        line: 1,
        column: 19,
        endLine: 1,
        endColumn: 20,
      }, {
        messageId: 'unexpectedWhitespaceAfter',
        line: 1,
        column: 20,
        endLine: 1,
        endColumn: 22,
      }],
    },

    // Class fields
    {
      code: 'class C { foo ;bar;}',
      output: 'class C { foo; bar;}',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C { foo; bar ; }',
      output: 'class C { foo ;bar ; }',
      options: [{ before: true, after: false }],
      errors: [
        {
          messageId: 'missingWhitespaceBefore',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'unexpectedWhitespaceAfter',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'class C { foo;static {}}',
      output: 'class C { foo; static {}}',
      errors: [{
        messageId: 'missingWhitespaceAfter',
        line: 1,
        column: 14,
        endLine: 1,
        endColumn: 15,
      }],
    },
    {
      code: 'class C { accessor foo ;accessor [bar] ;}',
      output: 'class C { accessor foo; accessor [bar];}',
      errors: [
        { messageId: 'unexpectedWhitespaceBefore' },
        { messageId: 'missingWhitespaceAfter' },
        { messageId: 'unexpectedWhitespaceBefore' },
      ],
    },
    {
      code: 'class C { accessor foo; accessor [bar]; }',
      output: 'class C { accessor foo ;accessor [bar] ; }',
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'missingWhitespaceBefore' },
        { messageId: 'unexpectedWhitespaceAfter' },
        { messageId: 'missingWhitespaceBefore' },
      ],
    },

    // ---- from semi-spacing._ts_.test.ts ----
    {
      code: 'type Union = number | string ;',
      output: 'type Union = number | string;',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'interface Foo { name: string;greet: () => string ; }',
      output: 'interface Foo { name: string; greet: () => string; }',
      errors: [
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'declare function example() ;',
      output: 'declare function example();',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'declare function example(): void ;',
      output: 'declare function example(): void;',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: `function foo(a: number): void;function foo(a: string): void ;
function foo(a: string | number): void {}`,
      output: `function foo(a: number): void; function foo(a: string): void;
function foo(a: string | number): void {}`,
      errors: [
        {
          messageId: 'missingWhitespaceAfter',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'interface Example { new (): number ; }',
      output: 'interface Example { new (): number; }',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'abstract class Example { abstract prop: string ; }',
      output: 'abstract class Example { abstract prop: string; }',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: 'abstract class Example { abstract method(): void ; }',
      output: 'abstract class Example { abstract method(): void; }',
      errors: [
        {
          messageId: 'unexpectedWhitespaceBefore',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
  ],
});
