/**
 * @fileoverview tests to validate spacing before and after comma.
 * @author Vignesh Anand.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/comma-spacing/comma-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/comma-spacing/comma-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('comma-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion / sourceType / ecmaFeatures.jsx) dropped — rslint
 *    resolves via tsconfig; the RuleTester picks a `.tsx` fixture when JSX is present.
 *  - `type` fields (deprecated AST node type) dropped (none were present).
 *  - Upstream errors pin either `messageId` (`missing` / `unexpected`) + `data.loc`,
 *    or a literal `message`. Both forms are kept verbatim; the RuleTester resolves a
 *    `messageId` through the plugin's own `meta.messages`
 *    (`unexpected`: "There should be no space {{loc}} ','.",
 *     `missing`: "A space is required {{loc}} ','.").
 *
 * No Babel/Flow cases and no external-fixture (`readFileSync`) cases exist in the
 * upstream comma-spacing tests, so nothing was skipped on those grounds. The
 * `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `KNOWN GAPS` block comment at the bottom, each annotated with
 * what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-spacing', null as never, {
  valid: [
    // ---- from comma-spacing._js_.test.ts ----
    'myfunc(404, true/* bla bla bla */, \'hello\');',
    'myfunc(404, true /* bla bla bla */, \'hello\');',
    'myfunc(404, true/* bla bla bla *//* hi */, \'hello\');',
    'myfunc(404, true/* bla bla bla */ /* hi */, \'hello\');',
    'myfunc(404, true, /* bla bla bla */ \'hello\');',
    'myfunc(404, // comment\n true, /* bla bla bla */ \'hello\');',
    { code: 'myfunc(404, // comment\n true,/* bla bla bla */ \'hello\');', options: [{ before: false, after: false }] },
    'var a = 1, b = 2;',
    'var arr = [,];',
    'var arr = [, ];',
    'var arr = [ ,];',
    'var arr = [ , ];',
    'var arr = [1,];',
    'var arr = [1, ];',
    'var arr = [, 2];',
    'var arr = [ , 2];',
    'var arr = [1, 2];',
    'var arr = [,,];',
    'var arr = [ ,,];',
    'var arr = [, ,];',
    'var arr = [,, ];',
    'var arr = [ , ,];',
    'var arr = [ ,, ];',
    'var arr = [, , ];',
    'var arr = [ , , ];',
    'var arr = [1, , ];',
    'var arr = [, 2, ];',
    'var arr = [, , 3];',
    'var arr = [,, 3];',
    'var arr = [1, 2, ];',
    'var arr = [, 2, 3];',
    'var arr = [1, , 3];',
    'var arr = [1, 2, 3];',
    'var arr = [1, 2, 3,];',
    'var arr = [1, 2, 3, ];',
    'var obj = {\'foo\':\'bar\', \'baz\':\'qur\'};',
    'var obj = {\'foo\':\'bar\', \'baz\':\'qur\', };',
    'var obj = {\'foo\':\'bar\', \'baz\':\'qur\',};',
    'var obj = {\'foo\':\'bar\', \'baz\':\n\'qur\'};',
    'var obj = {\'foo\':\n\'bar\', \'baz\':\n\'qur\'};',
    'function foo(a, b){}',
    { code: 'function foo(a, b = 1){}' },
    { code: 'function foo(a = 1, b, c){}' },
    { code: 'var foo = (a, b) => {}' },
    { code: 'var foo = (a=1, b) => {}' },
    { code: 'var foo = a => a + 2' },
    'a, b',
    'var a = (1 + 2, 2);',
    'a(b, c)',
    'new A(b, c)',
    'foo((a), b)',
    'var b = ((1 + 2), 2);',
    'parseInt((a + b), 10)',
    'go.boom((a + b), 10)',
    'go.boom((a + b), 10, (4))',
    'var x = [ (a + c), (b + b) ]',
    '[\'  ,  \']',
    { code: '[`  ,  `]' },
    { code: '`${[1, 2]}`' },
    { code: 'fn(a, b,)' }, // #11295
    { code: 'const fn = (a, b,) => {}' }, // #11295
    { code: 'const fn = function (a, b,) {}' }, // #11295
    { code: 'function fn(a, b,) {}' }, // #11295
    'foo(/,/, \'a\')',
    'var x = \',,,,,\';',
    'var code = \'var foo = 1, bar = 3;\'',
    '[\'apples\', \n \'oranges\'];',
    '{x: \'var x,y,z\'}',
    { code: 'var obj = {\'foo\':\n\'bar\' ,\'baz\':\n\'qur\'};', options: [{ before: true, after: false }] },
    { code: 'var a = 1 ,b = 2;', options: [{ before: true, after: false }] },
    { code: 'function foo(a ,b){}', options: [{ before: true, after: false }] },
    { code: 'var arr = [,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [, ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ , ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 , ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,2];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 ,2];', options: [{ before: true, after: false }] },
    { code: 'var arr = [,,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [, ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [,, ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ , ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,, ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [, , ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ , , ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 , ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,2 ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [,2 , ];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ , ,3];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 ,2 ,];', options: [{ before: true, after: false }] },
    { code: 'var arr = [ ,2 ,3];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 , ,3];', options: [{ before: true, after: false }] },
    { code: 'var arr = [1 ,2 ,3];', options: [{ before: true, after: false }] },
    { code: 'var obj = {\'foo\':\'bar\' , \'baz\':\'qur\'};', options: [{ before: true, after: true }] },
    { code: 'var obj = {\'foo\':\'bar\' ,\'baz\':\'qur\' , };', options: [{ before: true, after: false }] },
    { code: 'var a = 1 , b = 2;', options: [{ before: true, after: true }] },
    { code: 'var arr = [, ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [,,];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [ , 2];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , 2];', options: [{ before: true, after: true }] },
    { code: 'var arr = [, , ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , , ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [ , 2 , ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [ , , 3];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , 2 , ];', options: [{ before: true, after: true }] },
    { code: 'var arr = [, 2 , 3];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , , 3];', options: [{ before: true, after: true }] },
    { code: 'var arr = [1 , 2 , 3];', options: [{ before: true, after: true }] },
    { code: 'a , b', options: [{ before: true, after: true }] },
    { code: 'var arr = [,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [ ,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [,2];', options: [{ before: false, after: false }] },
    { code: 'var arr = [ ,2];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,2];', options: [{ before: false, after: false }] },
    { code: 'var arr = [,,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [ , , ];', options: [{ before: false, after: false }] },
    { code: 'var arr = [ ,,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [,2,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [ ,2,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [,,3];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,2,];', options: [{ before: false, after: false }] },
    { code: 'var arr = [,2,3];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,,3];', options: [{ before: false, after: false }] },
    { code: 'var arr = [1,2,3];', options: [{ before: false, after: false }] },
    { code: 'var a = (1 + 2,2)', options: [{ before: false, after: false }] },
    { code: 'var a; console.log(`${a}`, "a");' },
    { code: 'var [a, b] = [1, 2];' },
    { code: 'var [a, b, ] = [1, 2];' },
    { code: 'var [a, b,] = [1, 2];' },
    { code: 'var [a, , b] = [1, 2, 3];' },
    { code: 'var [a,, b] = [1, 2, 3];' },
    { code: 'var [ , b] = a;' },
    { code: 'var [, b] = a;' },
    { code: 'var { a,} = a;' },
    { code: 'import { a,} from \'mod\';' },
    { code: '<a>,</a>' },
    { code: '<a>  ,  </a>' },
    { code: '<a>Hello, world</a>', options: [{ before: true, after: false }] },

    // For backwards compatibility. Ignoring spacing between a comment and comma of a null element was possibly unintentional.
    { code: '[a, /**/ , ]', options: [{ before: false, after: true }] },
    { code: '[a , /**/, ]', options: [{ before: true, after: true }] },
    { code: '[a, /**/ , ] = foo', options: [{ before: false, after: true }] },
    { code: '[a , /**/, ] = foo', options: [{ before: true, after: true }] },

    // ---- from comma-spacing._ts_.test.ts ----
    'const Foo = <T,>(foo: T) => {}',
    'function foo<T,>() {}',
    'class Foo<T, T1> {}',
    'type Foo<T, P,> = Bar<T, P>',
    'interface Foo<T, T1,>{}',
    'let foo,',
  ],

  invalid: [
    // ---- from comma-spacing._js_.test.ts ----
    {
      code: 'a(b,c)',
      output: 'a(b , c)',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'new A(b,c)',
      output: 'new A(b , c)',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var a = 1 ,b = 2;',
      output: 'var a = 1, b = 2;',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var arr = [1 , 2];',
      output: 'var arr = [1, 2];',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1 , ];',
      output: 'var arr = [1, ];',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1 ,2];',
      output: 'var arr = [1, 2];',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var arr = [(1) , 2];',
      output: 'var arr = [(1), 2];',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1, 2];',
      output: 'var arr = [1 ,2];',
      options: [{ before: true, after: false }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          message: 'There should be no space after \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1\n  , 2];',
      output: 'var arr = [1\n  ,2];',
      options: [{ before: false, after: false }],
      errors: [
        {
          message: 'There should be no space after \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1,\n  2];',
      output: 'var arr = [1 ,\n  2];',
      options: [{ before: true, after: false }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
      ],
    },
    {
      code: 'var obj = {\'foo\':\n\'bar\', \'baz\':\n\'qur\'};',
      output: 'var obj = {\'foo\':\n\'bar\' ,\'baz\':\n\'qur\'};',
      options: [{ before: true, after: false }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          message: 'There should be no space after \',\'.',
        },
      ],
    },
    {
      code: 'var obj = {a: 1\n  ,b: 2};',
      output: 'var obj = {a: 1\n  , b: 2};',
      options: [{ before: false, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var obj = {a: 1 ,\n  b: 2};',
      output: 'var obj = {a: 1,\n  b: 2};',
      options: [{ before: false, after: false }],
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
      ],
    },
    {
      code: 'var arr = [1 ,2];',
      output: 'var arr = [1 , 2];',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var arr = [1,2];',
      output: 'var arr = [1 , 2];',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var obj = {\'foo\':\n\'bar\',\'baz\':\n\'qur\'};',
      output: 'var obj = {\'foo\':\n\'bar\' , \'baz\':\n\'qur\'};',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var arr = [1 , 2];',
      output: 'var arr = [1,2];',
      options: [{ before: false, after: false }],
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          message: 'There should be no space after \',\'.',
        },
      ],
    },
    {
      code: 'a ,b',
      output: 'a, b',
      options: [{ before: false, after: true }],
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'function foo(a,b){}',
      output: 'function foo(a , b){}',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var foo = (a,b) => {}',
      output: 'var foo = (a , b) => {}',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'var foo = (a = 1,b) => {}',
      output: 'var foo = (a = 1 , b) => {}',
      options: [{ before: true, after: true }],
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'function foo(a = 1 ,b = 2) {}',
      output: 'function foo(a = 1, b = 2) {}',
      options: [{ before: false, after: true }],
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: '<a>{foo(1 ,2)}</a>',
      output: '<a>{foo(1, 2)}</a>',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'myfunc(404, true/* bla bla bla */ , \'hello\');',
      output: 'myfunc(404, true/* bla bla bla */, \'hello\');',
      errors: [
        {
          message: 'There should be no space before \',\'.',
        },
      ],
    },
    {
      code: 'myfunc(404, true,/* bla bla bla */ \'hello\');',
      output: 'myfunc(404, true, /* bla bla bla */ \'hello\');',
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'myfunc(404,// comment\n true, \'hello\');',
      output: 'myfunc(404, // comment\n true, \'hello\');',
      errors: [
        {
          messageId: 'missing',
          data: { loc: 'after' },
        },
      ],
    },

    // ---- from comma-spacing._ts_.test.ts ----
    {
      code: 'function Foo<T,T1>() {}',
      output: 'function Foo<T, T1>() {}',
      errors: [
        {
          messageId: 'missing',
          column: 15,
          line: 1,
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'function Foo<T , T1>() {}',
      output: 'function Foo<T, T1>() {}',
      errors: [
        {
          messageId: 'unexpected',
          column: 16,
          line: 1,
          data: { loc: 'before' },
        },
      ],
    },
    {
      code: 'function Foo<T ,T1>() {}',
      output: 'function Foo<T, T1>() {}',
      errors: [
        {
          messageId: 'unexpected',
          column: 16,
          line: 1,
          data: { loc: 'before' },
        },
        {
          messageId: 'missing',
          column: 16,
          line: 1,
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'function Foo<T, T1>() {}',
      output: 'function Foo<T,T1>() {}',
      options: [{ before: false, after: false }],
      errors: [
        {
          messageId: 'unexpected',
          column: 15,
          line: 1,
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: 'function Foo<T,T1>() {}',
      output: 'function Foo<T ,T1>() {}',
      options: [{ before: true, after: false }],
      errors: [
        {
          messageId: 'missing',
          column: 15,
          line: 1,
          data: { loc: 'before' },
        },
      ],
    },
    // NOTE: the `let foo ,` invalid case is in KNOWN GAPS below — rslint's ts-go
    // parser accepts a trailing comma in a single-binding variable declaration
    // without error but emits no comma-spacing diagnostic for it, whereas upstream
    // (with @typescript-eslint/parser) reports the space before the trailing comma.
    {
      code: 'type Foo<T,P,> = Bar<T,P>',
      output: 'type Foo<T, P,> = Bar<T, P>',
      errors: [
        { messageId: 'missing', column: 11, line: 1, data: { loc: 'after' } },
        { messageId: 'missing', column: 23, line: 1, data: { loc: 'after' } },
      ],
    },
  ],
});

/**
 * ========================= comma-spacing — KNOWN GAPS =========================
 *
 * The case below is ported verbatim from upstream but is NOT run through the green
 * `ruleTester.run` above, because it surfaces a real rslint<->upstream behavioural
 * gap (not a parser-level abort — the fixture parses cleanly under ts-go).
 *
 * ---- invalid (upstream expects 1 `unexpected` diagnostic + the given fix) ----
 *
 *   { code: 'let foo ,', output: 'let foo,',
 *     errors: [{ messageId: 'unexpected', column: 9, line: 1, data: { loc: 'before' } }] }
 *
 *   upstream (@typescript-eslint/parser, what `lang: 'ts'` selects): parses
 *   `let foo ,` as a VariableDeclaration whose single declarator is followed by a
 *   trailing comma token, and reports "There should be no space before ','." at
 *   line 1, column 9.
 *
 *   rslint (ts-go parser): accepts `let foo ,` WITHOUT a syntax error (exit 0, no
 *   TS diagnostic) but emits ZERO @stylistic/comma-spacing diagnostics — the rule
 *   does not flag the space before a trailing comma at the end of a
 *   variable-declaration list. The corresponding valid case `'let foo,'` agrees
 *   between rslint and upstream (0 diagnostics) and so stays in the green set
 *   above; only the spacing variant differs. Note this is specific to the trailing
 *   comma: commas BETWEEN declarators (e.g. `var a = 1 ,b = 2;`) are flagged
 *   correctly by rslint and remain in the green set.
 *
 * ==============================================================================
 */

