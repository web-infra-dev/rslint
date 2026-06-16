/**
 * @fileoverview Tests for no-extra-semi rule.
 * @author Nicholas C. Zakas
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-extra-semi/no-extra-semi._js_.test.ts
 *   packages/eslint-plugin/rules/no-extra-semi/no-extra-semi._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('no-extra-semi', null as never, { valid, invalid })`
 *  - `rule` / the `#test` import / `name` / `lang` dropped; the rule is resolved
 *    from the mounted live plugin by id.
 *  - Errors use the `{ messageId: 'unexpected' }` form. Upstream's _ts_ file pins
 *    several errors as the bare string `'unexpected'` — that is upstream's
 *    `eslint-vitest-rule-tester` shorthand for a *messageId*. This repo's RuleTester
 *    reads a bare-string error as a literal *message* instead, so each such entry is
 *    normalized to its equivalent `{ messageId: 'unexpected' }` object form (the same
 *    objectification every other ported file applies). The `unexpected` message takes
 *    no `data`, rendering to the fixed string 'Unnecessary semicolon.'.
 *  - `parserOptions` (`ecmaVersion` / `sourceType: 'module'`) dropped — rslint
 *    resolves via tsconfig (esnext/module), which already enables class fields,
 *    static blocks, `for..of`, and module syntax used by these fixtures.
 *  - The `$` unindent template tag (used only in the _ts_ file) is evaluated to its
 *    real multi-line string and written here with explicit `\n` line breaks, the
 *    exact text after stripping the common 8-space indent and dropping the leading
 *    and trailing blank lines.
 *  - `type` fields: none present upstream, nothing dropped.
 *
 * Each upstream file has exactly ONE `run()` block (no `if (!skipBabel)` block, no
 * Babel/Flow cases, no second block). The `._css_` / `._json_` / `._markdown_` test
 * files don't exist for this rule. The _js_ cases come first, then the _ts_ cases.
 *
 * Every invalid case upstream pins `errors`; the cases that pin `output: null` (the
 * directive-prologue cases where removing the semicolon would turn a following
 * string literal into a directive — see no-extra-semi.ts `isFixable`) keep that pin,
 * and the RuleTester asserts the source is left unchanged. There are NO output-only
 * invalid cases.
 *
 * NO case surfaces a real rslint<->upstream gap. Verified empirically against the
 * rslint CLI: the `with(...)` fixtures (sloppy-mode-only in spec terms) parse under
 * ts-go and report/fix identically; the sloppy directive-prologue cases (`; 'use
 * strict'`, `debugger;\n;\n'use strict'`, etc.) report the same diagnostic count,
 * lines, and (no-)fix as upstream; the TS class-body `$` cases (accessor / abstract
 * props / abstract method overloads) match diagnostic count, the pinned columns, and
 * `--fix` output byte-for-byte. The `KNOWN GAPS` block at the bottom is therefore
 * empty (documented as such).
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-extra-semi', null as never, {
  valid: [
    // ---- from no-extra-semi._js_.test.ts ----
    'var x = 5;',
    'function foo(){}',
    'for(;;);',
    'while(0);',
    'do;while(0);',
    'for(a in b);',
    { code: 'for(a of b);' },
    'if(true);',
    'if(true); else;',
    'foo: ;',

    // Class body.
    { code: 'class A { }' },
    { code: 'var A = class { };' },
    { code: 'class A { a() { this; } }' },
    { code: 'var A = class { a() { this; } };' },
    { code: 'class A { } a;' },
    { code: 'class A { field; }' },
    { code: 'class A { field = 0; }' },
    { code: 'class A { static { foo; } }' },

    // modules
    { code: 'export const x = 42;' },
    { code: 'export default 42;' },

    // ---- from no-extra-semi._ts_.test.ts ----
    'with(foo);',

    // Class Property
    {
      code: 'export class Foo {\n  public foo: number = 0;\n  accessor bar: number = 1;\n  accessor [baz]: number = 2;\n}',
    },
    {
      code: 'export class Foo {\n  public foo: number = 0; public bar: number = 1;\n  accessor bar: number = 1; accessor [baz]: number = 2;\n}',
    },
  ],
  invalid: [
    // ---- from no-extra-semi._js_.test.ts ----
    {
      code: 'var x = 5;;',
      output: 'var x = 5;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(){};',
      output: 'function foo(){}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for(;;);;',
      output: 'for(;;);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'while(0);;',
      output: 'while(0);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'do;while(0);;',
      output: 'do;while(0);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for(a in b);;',
      output: 'for(a in b);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'for(a of b);;',
      output: 'for(a of b);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true);;',
      output: 'if(true);',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true){} else;;',
      output: 'if(true){} else;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'if(true){;} else {;}',
      output: 'if(true){} else {}',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'foo:;;',
      output: 'foo:;',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A { static { ; } }',
      output: 'class A { static {  } }',
      errors: [{ messageId: 'unexpected', column: 20 }],
    },
    {
      code: 'class A { static { a;; } }',
      output: 'class A { static { a; } }',
      errors: [{ messageId: 'unexpected', column: 22 }],
    },

    // Class body.
    {
      code: 'class A { ; }',
      output: 'class A {  }',
      errors: [{ messageId: 'unexpected', column: 11 }],
    },
    {
      code: 'class A { /*a*/; }',
      output: 'class A { /*a*/ }',
      errors: [{ messageId: 'unexpected', column: 16 }],
    },
    {
      code: 'class A { ; a() {} }',
      output: 'class A {  a() {} }',
      errors: [{ messageId: 'unexpected', column: 11 }],
    },
    {
      code: 'class A { a() {}; }',
      output: 'class A { a() {} }',
      errors: [{ messageId: 'unexpected', column: 17 }],
    },
    {
      code: 'class A { a() {}; b() {} }',
      output: 'class A { a() {} b() {} }',
      errors: [{ messageId: 'unexpected', column: 17 }],
    },
    {
      code: 'class A {; a() {}; b() {}; }',
      output: 'class A { a() {} b() {} }',
      errors: [
        { messageId: 'unexpected', column: 10 },
        { messageId: 'unexpected', column: 18 },
        { messageId: 'unexpected', column: 26 },
      ],
    },
    {
      code: 'class A { a() {}; get b() {} }',
      output: 'class A { a() {} get b() {} }',
      errors: [{ messageId: 'unexpected', column: 17 }],
    },
    {
      code: 'class A { field;; }',
      output: 'class A { field; }',
      errors: [{ messageId: 'unexpected', column: 17 }],
    },
    {
      code: 'class A { static {}; }',
      output: 'class A { static {} }',
      errors: [{ messageId: 'unexpected', column: 20 }],
    },
    {
      code: 'class A { static { a; }; foo(){} }',
      output: 'class A { static { a; } foo(){} }',
      errors: [{ messageId: 'unexpected', column: 24 }],
    },

    // https://github.com/eslint/eslint/issues/16988
    {
      code: '; \'use strict\'',
      output: null,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '; ; \'use strict\'',
      output: ' ; \'use strict\'',
      errors: [{ messageId: 'unexpected' }, { messageId: 'unexpected' }],
    },
    {
      code: 'debugger;\n;\n\'use strict\'',
      output: null,
      errors: [{ messageId: 'unexpected', line: 2 }],
    },
    {
      code: 'function foo() { ; \'bar\'; }',
      output: null,
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '{ ; \'foo\'; }',
      output: '{  \'foo\'; }',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '; (\'use strict\');',
      output: ' (\'use strict\');',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '; 1;',
      output: ' 1;',
      errors: [{ messageId: 'unexpected' }],
    },

    // ---- from no-extra-semi._ts_.test.ts ----
    {
      code: 'with(foo);;',
      output: 'with(foo);',
      errors: [
        { messageId: 'unexpected' },
      ],
    },
    {
      code: 'with(foo){;}',
      output: 'with(foo){}',
      errors: [
        { messageId: 'unexpected' },
      ],
    },

    // Class Property
    {
      code: 'class Foo {\n  public foo: number = 0;;\n}',
      output: 'class Foo {\n  public foo: number = 0;\n}',
      errors: [
        {
          messageId: 'unexpected',
          column: 26,
        },
      ],
    },
    {
      code: 'class Foo {\n  public foo: number = 0;; public bar: number = 1;;\n  public baz: number = 1;;\n}',
      output: 'class Foo {\n  public foo: number = 0; public bar: number = 1;\n  public baz: number = 1;\n}',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },

    // accessor property
    {
      code: 'class Foo {\n  accessor foo: number;; accessor [bar]: number;;\n  accessor baz: number;;\n}',
      output: 'class Foo {\n  accessor foo: number; accessor [bar]: number;\n  accessor baz: number;\n}',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },

    // abstract prop/method
    {
      code: 'class Foo {\n  abstract foo: number;; abstract bar: number;;\n  abstract baz: number;;\n}',
      output: 'class Foo {\n  abstract foo: number; abstract bar: number;\n  abstract baz: number;\n}',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },
    {
      code: 'class Foo {\n  abstract foo();; abstract bar();;\n  abstract baz();;\n  abstract foo(): void;; abstract bar(): void;;\n  abstract baz(): void;;\n}',
      output: 'class Foo {\n  abstract foo(); abstract bar();\n  abstract baz();\n  abstract foo(): void; abstract bar(): void;\n  abstract baz(): void;\n}',
      errors: [
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
        { messageId: 'unexpected' },
      ],
    },
  ],
});

/**
 * ====================== no-extra-semi — KNOWN GAPS ======================
 *
 * None. Verified empirically against the rslint CLI (the engine this RuleTester
 * drives), every upstream v5.10.0 fixture — _js_ and _ts_ — produces identical
 * output:
 *
 *  - `with(foo)` parses under ts-go (no TS1101 strict-mode rejection here); the
 *    `with(foo);` valid case reports 0, and `with(foo);;` / `with(foo){;}` report 1
 *    and fix to `with(foo);` / `with(foo){}`, matching upstream.
 *  - The eslint#16988 directive-prologue cases (`; 'use strict'`, `; ; 'use strict'`,
 *    `debugger;\n;\n'use strict'`, `function foo() { ; 'bar'; }`, `{ ; 'foo'; }`,
 *    `; ('use strict');`, `; 1;`) all match upstream on diagnostic count, the pinned
 *    line, and the `isFixable` no-op (`output: null`) vs. fix behaviour.
 *  - The TS `$` class-body cases (accessor properties, abstract props, abstract
 *    method/overload signatures) match diagnostic count, the pinned columns (col 26
 *    on the first `public` case), and `--fix` output byte-for-byte.
 *
 * No fixture hits an octal/escape edge case, an `assert`/`with` import attribute, or
 * JSX, so nothing is unparseable under ts-go's strict/module semantics. Every invalid
 * case pins `errors`; there are NO output-only invalid cases.
 * ========================================================================
 */
