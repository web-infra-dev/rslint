/**
 * @fileoverview This rule should require or disallow spaces before or after unary operations.
 * @author Marcin Kumorek
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/space-unary-ops/space-unary-ops.test.ts
 *
 * The upstream file has a single `run({ name, rule, valid, invalid })` block
 * (no second skipBabel block). Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('space-unary-ops', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string
 *    (common leading indentation stripped, leading/trailing blank lines dropped),
 *    rendered here as a single-quoted string with `\n` line breaks.
 *  - `parserOptions: { ecmaVersion: 6 | 8 | 2022 }` dropped — rslint resolves
 *    language level via tsconfig (target/module esnext), so yield/await/private-
 *    field syntax all parse.
 *
 * No Babel/Flow cases, no external-fixture (`readFileSync`) cases, and no
 * `._css_`/`._json_`/`._markdown_` files exist for this rule. Every fixture is
 * valid TypeScript (the `a!` non-null assertion cases are TS-specific and parse
 * cleanly under ts-go), so nothing is isolated into KNOWN GAPS — see the note at
 * the bottom of the file confirming the empty gap set.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('space-unary-ops', null as never, {
  valid: [
    {
      code: '++this.a',
      options: [{ words: true }],
    },
    {
      code: '--this.a',
      options: [{ words: true }],
    },
    {
      code: 'this.a++',
      options: [{ words: true }],
    },
    {
      code: 'this.a--',
      options: [{ words: true }],
    },
    'foo .bar++',
    {
      code: 'foo.bar --',
      options: [{ nonwords: true }],
    },

    {
      code: 'delete foo.bar',
      options: [{ words: true }],
    },
    {
      code: 'delete foo["bar"]',
      options: [{ words: true }],
    },
    {
      code: 'delete foo.bar',
      options: [{ words: false }],
    },
    {
      code: 'delete(foo.bar)',
      options: [{ words: false }],
    },

    {
      code: 'new Foo',
      options: [{ words: true }],
    },
    {
      code: 'new Foo()',
      options: [{ words: true }],
    },
    {
      code: 'new [foo][0]',
      options: [{ words: true }],
    },
    {
      code: 'new[foo][0]',
      options: [{ words: false }],
    },

    {
      code: 'typeof foo',
      options: [{ words: true }],
    },
    {
      code: 'typeof{foo:true}',
      options: [{ words: false }],
    },
    {
      code: 'typeof {foo:true}',
      options: [{ words: true }],
    },
    {
      code: 'typeof (foo)',
      options: [{ words: true }],
    },
    {
      code: 'typeof(foo)',
      options: [{ words: false }],
    },
    {
      code: 'typeof!foo',
      options: [{ words: false }],
    },

    {
      code: 'void 0',
      options: [{ words: true }],
    },
    {
      code: '(void 0)',
      options: [{ words: true }],
    },
    {
      code: '(void (0))',
      options: [{ words: true }],
    },
    {
      code: 'void foo',
      options: [{ words: true }],
    },
    {
      code: 'void foo',
      options: [{ words: false }],
    },
    {
      code: 'void(foo)',
      options: [{ words: false }],
    },

    {
      code: '-1',
      options: [{ nonwords: false }],
    },
    {
      code: '!foo',
      options: [{ nonwords: false }],
    },
    {
      code: '!!foo',
      options: [{ nonwords: false }],
    },
    {
      code: 'foo++',
      options: [{ nonwords: false }],
    },
    {
      code: 'foo ++',
      options: [{ nonwords: true }],
    },
    {
      code: '++foo',
      options: [{ nonwords: false }],
    },
    {
      code: '++ foo',
      options: [{ nonwords: true }],
    },
    {
      code: 'function *foo () { yield (0) }',
    },
    {
      code: 'function *foo() { yield +1 }',
    },
    {
      code: 'function *foo() { yield* 0 }',
    },
    {
      code: 'function *foo() { yield * 0 }',
    },
    {
      code: 'function *foo() { (yield)*0 }',
    },
    {
      code: 'function *foo() { (yield) * 0 }',
    },
    {
      code: 'function *foo() { yield*0 }',
    },
    {
      code: 'function *foo() { yield *0 }',
    },
    {
      code: 'async function foo() { await {foo: 1} }',
    },
    {
      code: 'async function foo() { await {bar: 2} }',
    },
    {
      code: 'async function foo() { await{baz: 3} }',
      options: [{ words: false }],
    },
    {
      code: 'async function foo() { await {qux: 4} }',
      options: [{ words: false, overrides: { await: true } }],
    },
    {
      code: 'async function foo() { await{foo: 5} }',
      options: [{ words: true, overrides: { await: false } }],
    },
    {
      code: 'foo++',
      options: [{ nonwords: true, overrides: { '++': false } }],
    },
    {
      code: 'foo++',
      options: [{ nonwords: false, overrides: { '++': false } }],
    },
    {
      code: '++foo',
      options: [{ nonwords: true, overrides: { '++': false } }],
    },
    {
      code: '++foo',
      options: [{ nonwords: false, overrides: { '++': false } }],
    },
    {
      code: '!foo',
      options: [{ nonwords: true, overrides: { '!': false } }],
    },
    {
      code: '!foo',
      options: [{ nonwords: false, overrides: { '!': false } }],
    },
    {
      code: 'new foo',
      options: [{ words: true, overrides: { new: false } }],
    },
    {
      code: 'new foo',
      options: [{ words: false, overrides: { new: false } }],
    },
    {
      code: 'function *foo () { yield(0) }',
      options: [{ words: true, overrides: { yield: false } }],
    },
    {
      code: 'function *foo () { yield(0) }',
      options: [{ words: false, overrides: { yield: false } }],
    },
    {
      code: 'class C { #x; *foo(bar) { yield#x in bar; } }',
      options: [{ words: false }],
    },
    {
      code: 'a!.b!.c\n!a.b.c',
      options: [{ nonwords: false }],
    },
    {
      code: 'a !.b !.c\n! a.b.c',
      options: [{ nonwords: true }],
    },
    {
      code: 'a !.b !.c\n! a.b.c',
      options: [{ nonwords: false, overrides: { '!': true } }],
    },
    {
      code: 'a!.b!.c\n! a.b.c',
      options: [{ nonwords: false, overrides: { 'ts-non-null': false, '!': true } }],
    },
    {
      code: 'a !.b !.c\n!a.b.c',
      options: [{ nonwords: false, overrides: { 'ts-non-null': true } }],
    },
    {
      code: 'a!.b!.c\n! a.b.c',
      options: [{ nonwords: true, overrides: { 'ts-non-null': false } }],
    },
  ],

  invalid: [
    {
      code: 'delete(foo.bar)',
      output: 'delete (foo.bar)',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'delete' },
      }],
    },
    {
      code: 'delete(foo["bar"]);',
      output: 'delete (foo["bar"]);',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'delete' },
      }],
    },
    {
      code: 'delete (foo.bar)',
      output: 'delete(foo.bar)',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'delete' },
      }],
    },
    {
      code: 'new(Foo)',
      output: 'new (Foo)',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'new' },
      }],
    },
    {
      code: 'new (Foo)',
      output: 'new(Foo)',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'new' },
      }],
    },
    {
      code: 'new(Foo())',
      output: 'new (Foo())',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'new' },
      }],
    },
    {
      code: 'new [foo][0]',
      output: 'new[foo][0]',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'new' },
      }],
    },

    {
      code: 'typeof(foo)',
      output: 'typeof (foo)',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof (foo)',
      output: 'typeof(foo)',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof[foo]',
      output: 'typeof [foo]',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof [foo]',
      output: 'typeof[foo]',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof{foo:true}',
      output: 'typeof {foo:true}',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof {foo:true}',
      output: 'typeof{foo:true}',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'typeof' },
      }],
    },
    {
      code: 'typeof!foo',
      output: 'typeof !foo',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'typeof' },
      }],
    },

    {
      code: 'void(0);',
      output: 'void (0);',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'void' },
      }],
    },
    {
      code: 'void(foo);',
      output: 'void (foo);',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'void' },
      }],
    },
    {
      code: 'void[foo];',
      output: 'void [foo];',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'void' },
      }],
    },
    {
      code: 'void{a:0};',
      output: 'void {a:0};',
      options: [{ words: true }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'void' },
      }],
    },
    {
      code: 'void (foo)',
      output: 'void(foo)',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'void' },
      }],
    },
    {
      code: 'void [foo]',
      output: 'void[foo]',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'void' },
      }],
    },

    {
      code: '! foo',
      output: '!foo',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '!' },
      }],
    },
    {
      code: '!foo',
      output: '! foo',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '!' },
      }],
    },

    {
      code: '!! foo',
      output: '!!foo',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '!' },
        line: 1,
        column: 2,
      }],
    },
    {
      code: '!!foo',
      output: '!! foo',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '!' },
        line: 1,
        column: 2,
      }],
    },

    {
      code: '- 1',
      output: '-1',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '-' },
      }],
    },
    {
      code: '-1',
      output: '- 1',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '-' },
      }],
    },

    {
      code: 'foo++',
      output: 'foo ++',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireBefore',
        data: { operator: '++' },
      }],
    },
    {
      code: 'foo ++',
      output: 'foo++',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedBefore',
        data: { operator: '++' },
      }],
    },
    {
      code: '++ foo',
      output: '++foo',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '++' },
      }],
    },
    {
      code: '++foo',
      output: '++ foo',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '++' },
      }],
    },
    {
      code: 'foo .bar++',
      output: 'foo .bar ++',
      options: [{ nonwords: true }],
      errors: [{
        messageId: 'requireBefore',
        data: { operator: '++' },
      }],
    },
    {
      code: 'foo.bar --',
      output: 'foo.bar--',
      errors: [{
        messageId: 'unexpectedBefore',
        data: { operator: '--' },
      }],
    },
    {
      code: '+ +foo',
      output: null,
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '+' },
      }],
    },
    {
      code: '+ ++foo',
      output: null,
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '+' },
      }],
    },
    {
      code: '- -foo',
      output: null,
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '-' },
      }],
    },
    {
      code: '- --foo',
      output: null,
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '-' },
      }],
    },
    {
      code: '+ -foo',
      output: '+-foo',
      options: [{ nonwords: false }],
      errors: [{
        messageId: 'unexpectedAfter',
        data: { operator: '+' },
      }],
    },
    {
      code: 'function *foo() { yield(0) }',
      output: 'function *foo() { yield (0) }',
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 19,
      }],
    },
    {
      code: 'function *foo() { yield (0) }',
      output: 'function *foo() { yield(0) }',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 19,
      }],
    },
    {
      code: 'function *foo() { yield+0 }',
      output: 'function *foo() { yield +0 }',
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 19,
      }],
    },
    {
      code: 'foo++',
      output: 'foo ++',
      options: [{ nonwords: true, overrides: { '++': true } }],
      errors: [{
        messageId: 'requireBefore',
        data: { operator: '++' },
      }],
    },
    {
      code: 'foo++',
      output: 'foo ++',
      options: [{ nonwords: false, overrides: { '++': true } }],
      errors: [{
        messageId: 'requireBefore',
        data: { operator: '++' },
      }],
    },
    {
      code: '++foo',
      output: '++ foo',
      options: [{ nonwords: true, overrides: { '++': true } }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '++' },
      }],
    },
    {
      code: '++foo',
      output: '++ foo',
      options: [{ nonwords: false, overrides: { '++': true } }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '++' },
      }],
    },
    {
      code: '!foo',
      output: '! foo',
      options: [{ nonwords: true, overrides: { '!': true } }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '!' },
      }],
    },
    {
      code: '!foo',
      output: '! foo',
      options: [{ nonwords: false, overrides: { '!': true } }],
      errors: [{
        messageId: 'requireAfter',
        data: { operator: '!' },
      }],
    },
    {
      code: 'new(Foo)',
      output: 'new (Foo)',
      options: [{ words: true, overrides: { new: true } }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'new' },
      }],
    },
    {
      code: 'new(Foo)',
      output: 'new (Foo)',
      options: [{ words: false, overrides: { new: true } }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'new' },
      }],
    },
    {
      code: 'function *foo() { yield(0) }',
      output: 'function *foo() { yield (0) }',
      options: [{ words: true, overrides: { yield: true } }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 19,
      }],
    },
    {
      code: 'function *foo() { yield(0) }',
      output: 'function *foo() { yield (0) }',
      options: [{ words: false, overrides: { yield: true } }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 19,
      }],
    },
    {
      code: 'async function foo() { await{foo: \'bar\'} }',
      output: 'async function foo() { await {foo: \'bar\'} }',
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'await' },
        line: 1,
        column: 24,
      }],
    },
    {
      code: 'async function foo() { await{baz: \'qux\'} }',
      output: 'async function foo() { await {baz: \'qux\'} }',
      options: [{ words: false, overrides: { await: true } }],
      errors: [{
        messageId: 'requireAfterWord',
        data: { word: 'await' },
        line: 1,
        column: 24,
      }],
    },
    {
      code: 'async function foo() { await {foo: 1} }',
      output: 'async function foo() { await{foo: 1} }',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'await' },
        line: 1,
        column: 24,
      }],
    },
    {
      code: 'async function foo() { await {bar: 2} }',
      output: 'async function foo() { await{bar: 2} }',
      options: [{ words: true, overrides: { await: false } }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'await' },
        line: 1,
        column: 24,
      }],
    },
    {
      code: 'class C { #x; *foo(bar) { yield #x in bar; } }',
      output: 'class C { #x; *foo(bar) { yield#x in bar; } }',
      options: [{ words: false }],
      errors: [{
        messageId: 'unexpectedAfterWord',
        data: { word: 'yield' },
        line: 1,
        column: 27,
      }],
    },
    {
      code: 'const w = func()!',
      output: 'const w = func() !',
      options: [{ nonwords: false, overrides: { '!': true } }],
      errors: [
        { messageId: 'requireBefore', data: { operator: '!' } },
      ],
    },
    {
      code: 'a  !  .b  !  .c',
      output: 'a!  .b!  .c',
      options: [{ nonwords: false }],
      errors: [
        { messageId: 'unexpectedBefore', data: { operator: '!' } },
        { messageId: 'unexpectedBefore', data: { operator: '!' } },
      ],
    },
    {
      code: 'a!',
      output: 'a !',
      options: [{ nonwords: false, overrides: { 'ts-non-null': true } }],
      errors: [
        { messageId: 'requireBefore', data: { operator: '!' } },
      ],
    },
    {
      code: 'a !',
      output: 'a!',
      options: [{ nonwords: true, overrides: { 'ts-non-null': false } }],
      errors: [
        { messageId: 'unexpectedBefore', data: { operator: '!' } },
      ],
    },
  ],
});

/**
 * ======================= space-unary-ops — KNOWN GAPS =======================
 *
 * None. Every upstream fixture is valid TypeScript and parses cleanly under the
 * ts-go parser (including the `a!` non-null-assertion and private-field
 * `#x`/`yield#x in bar` cases). The only upstream-specific metadata dropped was
 * `parserOptions: { ecmaVersion: 6 | 8 | 2022 }`, which rslint supplies via the
 * generated tsconfig (target/module esnext) — not a behavioural gap. No octal /
 * sloppy-mode / Babel / Flow / import-attribute cases exist in this rule's tests.
 * ============================================================================
 */
