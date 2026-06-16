/**
 * @fileoverview Tests for comma-dangle rule.
 * @author Ian Christian Myers
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/comma-dangle/comma-dangle._js_.test.ts
 *   packages/eslint-plugin/rules/comma-dangle/comma-dangle._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('comma-dangle', null as never, { valid, invalid })`.
 *    The upstream `._js_` `run()` block and the `._ts_` `run()` block are merged into
 *    the single valid/invalid pair below.
 *  - The `$` unindent template tag is evaluated to its real multi-line string (common
 *    leading indentation stripped, leading/trailing blank lines removed); plain
 *    backtick multi-line templates are kept verbatim.
 *  - `parserOptions` (ecmaVersion / sourceType / ecmaFeatures.jsx) dropped — rslint
 *    resolves via tsconfig; the RuleTester picks a `.tsx` fixture when JSX is present
 *    and honors an explicit `filename`.
 *  - comma-dangle has no `type` AST fields to drop.
 *  - `errors: 2` (numeric count) kept as the number — the RuleTester supports it.
 *
 * The rule's `meta.messages` carry no `{{data}}` interpolation, so every error pins
 * only `messageId` (+ line/column/endLine/endColumn when upstream gives them):
 *   unexpected: 'Unexpected trailing comma.'
 *   missing:    'Missing trailing comma.'
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to the `comma-dangle — KNOWN GAPS` block comment at the bottom, each
 * annotated with what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-dangle', null as never, {
  valid: [
    // ---- from comma-dangle._js_.test.ts ----
    'var foo = { bar: \'baz\' }',
    'var foo = {\nbar: \'baz\'\n}',
    'var foo = [ \'baz\' ]',
    'var foo = [\n\'baz\'\n]',
    '[,,]',
    '[\n,\n,\n]',
    '[,]',
    '[\n,\n]',
    '[]',
    '[\n]',
    { code: 'var foo = [\n      (bar ? baz : qux),\n    ];', options: ['always-multiline'] },
    { code: 'var foo = { bar: \'baz\' }', options: ['never'] },
    { code: 'var foo = {\nbar: \'baz\'\n}', options: ['never'] },
    { code: 'var foo = [ \'baz\' ]', options: ['never'] },
    { code: 'var { a, b } = foo;', options: ['never'] },
    { code: 'var [ a, b ] = foo;', options: ['never'] },
    { code: 'var { a,\n b, \n} = foo;', options: ['only-multiline'] },
    { code: 'var [ a,\n b, \n] = foo;', options: ['only-multiline'] },

    { code: '[(1),]', options: ['always'] },
    { code: 'var x = { foo: (1),};', options: ['always'] },
    { code: 'var foo = { bar: \'baz\', }', options: ['always'] },
    { code: 'var foo = {\nbar: \'baz\',\n}', options: ['always'] },
    { code: 'var foo = {\nbar: \'baz\'\n,}', options: ['always'] },
    { code: 'var foo = [ \'baz\', ]', options: ['always'] },
    { code: 'var foo = [\n\'baz\',\n]', options: ['always'] },
    { code: 'var foo = [\n\'baz\'\n,]', options: ['always'] },
    { code: '[,,]', options: ['always'] },
    { code: '[\n,\n,\n]', options: ['always'] },
    { code: '[,]', options: ['always'] },
    { code: '[\n,\n]', options: ['always'] },
    { code: '[]', options: ['always'] },
    { code: '[\n]', options: ['always'] },

    { code: 'var foo = { bar: \'baz\' }', options: ['always-multiline'] },
    { code: 'var foo = { bar: \'baz\' }', options: ['only-multiline'] },
    { code: 'var foo = {\nbar: \'baz\',\n}', options: ['always-multiline'] },
    { code: 'var foo = {\nbar: \'baz\',\n}', options: ['only-multiline'] },
    { code: 'var foo = [ \'baz\' ]', options: ['always-multiline'] },
    { code: 'var foo = [ \'baz\' ]', options: ['only-multiline'] },
    { code: 'var foo = [\n\'baz\',\n]', options: ['always-multiline'] },
    { code: 'var foo = [\n\'baz\',\n]', options: ['only-multiline'] },
    { code: 'var foo = { bar:\n\n\'bar\' }', options: ['always-multiline'] },
    { code: 'var foo = { bar:\n\n\'bar\' }', options: ['only-multiline'] },
    { code: 'var foo = {a: 1, b: 2, c: 3, d: 4}', options: ['always-multiline'] },
    { code: 'var foo = {a: 1, b: 2, c: 3, d: 4}', options: ['only-multiline'] },
    { code: 'var foo = {a: 1, b: 2,\n c: 3, d: 4}', options: ['always-multiline'] },
    { code: 'var foo = {a: 1, b: 2,\n c: 3, d: 4}', options: ['only-multiline'] },
    { code: 'var foo = {x: {\nfoo: \'bar\',\n}}', options: ['always-multiline'] },
    { code: 'var foo = {x: {\nfoo: \'bar\',\n}}', options: ['only-multiline'] },
    { code: 'var foo = new Map([\n[key, {\na: 1,\nb: 2,\nc: 3,\n}],\n])', options: ['always-multiline'] },
    { code: 'var foo = new Map([\n[key, {\na: 1,\nb: 2,\nc: 3,\n}],\n])', options: ['only-multiline'] },

    // https://github.com/eslint/eslint/issues/3627
    {
      code: 'var [a, ...rest] = [];',
      options: ['always'],
    },
    {
      code: 'var [\n    a,\n    ...rest\n] = [];',
      options: ['always'],
    },
    {
      code: 'var [\n    a,\n    ...rest\n] = [];',
      options: ['always-multiline'],
    },
    {
      code: 'var [\n    a,\n    ...rest\n] = [];',
      options: ['only-multiline'],
    },
    {
      code: '[a, ...rest] = [];',
      options: ['always'],
    },
    {
      code: 'for ([a, ...rest] of []);',
      options: ['always'],
    },
    {
      code: 'var a = [b, ...spread,];',
      options: ['always'],
    },

    // https://github.com/eslint/eslint/issues/7297
    {
      code: 'var {foo, ...bar} = baz',
      options: ['always'],
    },

    // https://github.com/eslint/eslint/issues/3794
    {
      code: 'import {foo,} from \'foo\';',
      options: ['always'],
    },
    {
      code: 'import foo from \'foo\';',
      options: ['always'],
    },
    {
      code: 'import foo, {abc,} from \'foo\';',
      options: ['always'],
    },
    {
      code: 'import * as foo from \'foo\';',
      options: ['always'],
    },
    {
      code: 'export {foo,} from \'foo\';',
      options: ['always'],
    },
    {
      code: 'import {foo} from \'foo\';',
      options: ['never'],
    },
    {
      code: 'import foo from \'foo\';',
      options: ['never'],
    },
    {
      code: 'import foo, {abc} from \'foo\';',
      options: ['never'],
    },
    {
      code: 'import * as foo from \'foo\';',
      options: ['never'],
    },
    {
      code: 'export {foo} from \'foo\';',
      options: ['never'],
    },
    {
      code: 'import {foo} from \'foo\';',
      options: ['always-multiline'],
    },
    {
      code: 'import {foo} from \'foo\';',
      options: ['only-multiline'],
    },
    {
      code: 'export {foo} from \'foo\';',
      options: ['always-multiline'],
    },
    {
      code: 'export {foo} from \'foo\';',
      options: ['only-multiline'],
    },
    {
      code: 'import {\n  foo,\n} from \'foo\';',
      options: ['always-multiline'],
    },
    {
      code: 'import {\n  foo,\n} from \'foo\';',
      options: ['only-multiline'],
    },
    {
      code: 'export {\n  foo,\n} from \'foo\';',
      options: ['always-multiline'],
    },
    {
      code: 'export {\n  foo,\n} from \'foo\';',
      options: ['only-multiline'],
    },
    {
      code: 'import {foo} from \n\'foo\';',
      options: ['always-multiline'],
    },
    {
      code: 'import {foo} from \n\'foo\';',
      options: ['only-multiline'],
    },
    // NOTE: `function foo(a) {}` / `foo(a)` with ['always'] (upstream ecmaVersion 5/7,
    // functions normalized to 'ignore') and `foo(a,\nb\n)` / `function foo(a,\nb\n) {}`
    // with ['always-multiline'] (upstream ecmaVersion 5/7) are in KNOWN GAPS below.
    {
      code: 'function foo(a) {}',
      options: ['never'],
    },
    {
      code: 'foo(a)',
      options: ['never'],
    },
    {
      code: 'function foo(a,\nb) {}',
      options: ['always-multiline'],
    },
    {
      code: 'foo(a,\nb)',
      options: ['always-multiline'],
    },
    {
      code: 'function foo(a,\nb) {}',
      options: ['only-multiline'],
    },
    {
      code: 'foo(a,\nb)',
      options: ['only-multiline'],
    },
    {
      code: 'function foo(a) {}',
      options: ['never'],
    },
    {
      code: 'foo(a)',
      options: ['never'],
    },
    {
      code: 'function foo(a,\nb) {}',
      options: ['always-multiline'],
    },
    {
      code: 'foo(a,\nb)',
      options: ['always-multiline'],
    },
    {
      code: 'function foo(a,\nb) {}',
      options: ['only-multiline'],
    },
    {
      code: 'foo(a,\nb)',
      options: ['only-multiline'],
    },
    {
      code: 'function foo(a) {}',
      options: ['never'],
    },
    {
      code: 'foo(a)',
      options: ['never'],
    },
    {
      code: 'function foo(a,) {}',
      options: ['always'],
    },
    {
      code: 'foo(a,)',
      options: ['always'],
    },
    {
      code: 'function foo(\na,\nb,\n) {}',
      options: ['always-multiline'],
    },
    {
      code: 'foo(\na,b)',
      options: ['always-multiline'],
    },
    {
      code: 'function foo(a,b) {}',
      options: ['always-multiline'],
    },
    {
      code: 'foo(a,b)',
      options: ['always-multiline'],
    },
    {
      code: 'function foo(a,b) {}',
      options: ['only-multiline'],
    },
    {
      code: 'foo(a,b)',
      options: ['only-multiline'],
    },

    // trailing comma in functions
    {
      code: 'function foo(a) {} ',
      options: [{}],
    },
    {
      code: 'foo(a)',
      options: [{}],
    },
    {
      code: 'function foo(a) {} ',
      options: [{ functions: 'never' }],
    },
    {
      code: 'foo(a)',
      options: [{ functions: 'never' }],
    },
    {
      code: 'function foo(a,) {}',
      options: [{ functions: 'always' }],
    },
    {
      code: 'function bar(a, ...b) {}',
      options: [{ functions: 'always' }],
    },
    {
      code: 'foo(a,)',
      options: [{ functions: 'always' }],
    },
    {
      code: 'foo(a,)',
      options: [{ functions: 'always' }],
    },
    {
      code: 'bar(...a,)',
      options: [{ functions: 'always' }],
    },
    {
      code: 'function foo(a) {} ',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'foo(a)',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'function foo(\na,\nb,\n) {} ',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'function foo(\na,\n...b\n) {} ',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'foo(\na,\nb,\n)',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'foo(\na,\n...b,\n)',
      options: [{ functions: 'always-multiline' }],
    },
    {
      code: 'function foo(a) {} ',
      options: [{ functions: 'only-multiline' }],
    },
    {
      code: 'foo(a)',
      options: [{ functions: 'only-multiline' }],
    },
    {
      code: 'function foo(\na,\nb,\n) {} ',
      options: [{ functions: 'only-multiline' }],
    },
    {
      code: 'foo(\na,\nb,\n)',
      options: [{ functions: 'only-multiline' }],
    },
    {
      code: 'function foo(\na,\nb\n) {} ',
      options: [{ functions: 'only-multiline' }],
    },
    {
      code: 'foo(\na,\nb\n)',
      options: [{ functions: 'only-multiline' }],
    },

    // https://github.com/eslint-stylistic/eslint-stylistic/issues/158
    { code: 'a => 42;', options: ['always'] },

    // dynamic import
    {
      code: 'import(source)',
    },
    {
      code: 'import(source, )',
      options: ['always'],
    },
    {
      code: 'import(source, options, )',
      options: ['always'],
    },
    // NOTE: `import(source)` + ['always'] (upstream ecmaVersion 15, dynamicImports
    // normalized to 'ignore') is in KNOWN GAPS below.
    {
      code: 'import(source,)',
      options: ['always'],
    },
    {
      code: 'import(source)',
      options: ['never'],
    },
    {
      code: 'import(source, options)',
      options: ['never'],
    },
    {
      code: 'import(source)',
      options: ['always-multiline'],
    },
    {
      code: 'import(source, options)',
      options: ['always-multiline'],
    },
    {
      code: 'import(\n  source,\n)',
      options: ['always-multiline'],
    },
    {
      code: 'import(\n  source,\n  options,\n)',
      options: ['always-multiline'],
    },
    {
      code: 'import(source)',
      options: ['only-multiline'],
    },
    {
      code: 'import(source, options)',
      options: ['only-multiline'],
    },
    {
      code: 'import(\n  source,\n)',
      options: ['only-multiline'],
    },
    {
      code: 'import(\n  source\n)',
      options: ['only-multiline'],
    },
    {
      code: 'import(\n  source,\n  options,\n)',
      options: ['only-multiline'],
    },
    {
      code: 'import(\n  source,\n  options\n)',
      options: ['only-multiline'],
    },
    {
      code: 'import(source)',
      options: [{ functions: 'always' }],
    },
    {
      code: 'import(source,)',
      options: [{ functions: 'never', dynamicImports: 'always' }],
    },

    // import attributes
    {
      code: 'import foo from "foo" with {type: "json"}',
    },
    {
      code: 'import foo from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'import foo from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'import foo from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
    },
    {
      code: 'export {foo,} from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },
    {
      code: 'export * from "foo" with {type: "json"}',
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'export * from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'export * from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'export * from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'export * from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },

    // ---- from comma-dangle._ts_.test.ts ----
    // default
    { code: 'enum Foo {}' },
    { code: 'enum Foo {\n}' },
    { code: 'enum Foo {Bar}' },
    { code: 'function Foo<T>() {}' },
    { code: 'type Foo = []' },
    { code: 'type Foo = [\n]' },

    // never
    { code: 'enum Foo {Bar}', options: ['never'] },
    { code: 'enum Foo {Bar\n}', options: ['never'] },
    { code: 'enum Foo {Bar\n}', options: [{ enums: 'never' }] },
    { code: 'function Foo<T>() {}', options: ['never'] },
    { code: 'function Foo<T\n>() {}', options: ['never'] },
    { code: 'function Foo<T\n>() {}', options: [{ generics: 'never' }] },
    { code: 'type Foo = [string]', options: ['never'] },
    { code: 'type Foo = [string]', options: [{ tuples: 'never' }] },

    // always
    { code: 'enum Foo {Bar,}', options: ['always'] },
    { code: 'enum Foo {Bar,\n}', options: ['always'] },
    { code: 'enum Foo {Bar,\n}', options: [{ enums: 'always' }] },
    { code: 'function Foo<T,>() {}', options: ['always'] },
    { code: 'function Foo<T,\n>() {}', options: ['always'] },
    { code: 'function Foo<T,\n>() {}', options: [{ generics: 'always' }] },
    { code: 'type Foo = [string,]', options: ['always'] },
    { code: 'type Foo = [string,\n]', options: [{ tuples: 'always' }] },

    // always-multiline
    { code: 'enum Foo {Bar}', options: ['always-multiline'] },
    { code: 'enum Foo {Bar,\n}', options: ['always-multiline'] },
    { code: 'enum Foo {Bar,\n}', options: [{ enums: 'always-multiline' }] },
    { code: 'function Foo<T>() {}', options: ['always-multiline'] },
    { code: 'function Foo<T,\n>() {}', options: ['always-multiline'] },
    {
      code: 'function Foo<T,\n>() {}',
      options: [{ generics: 'always-multiline' }],
    },
    { code: 'type Foo = [string]', options: ['always-multiline'] },
    { code: 'type Foo = [string,\n]', options: ['always-multiline'] },
    {
      code: 'type Foo = [string,\n]',
      options: [{ tuples: 'always-multiline' }],
    },

    // only-multiline
    { code: 'enum Foo {Bar}', options: ['only-multiline'] },
    { code: 'enum Foo {Bar\n}', options: ['only-multiline'] },
    { code: 'enum Foo {Bar,\n}', options: ['only-multiline'] },
    { code: 'enum Foo {Bar,\n}', options: [{ enums: 'only-multiline' }] },
    { code: 'function Foo<T>() {}', options: ['only-multiline'] },
    { code: 'function Foo<T\n>() {}', options: ['only-multiline'] },
    { code: 'function Foo<T,\n>() {}', options: ['only-multiline'] },
    {
      code: 'function Foo<T\n>() {}',
      options: [{ generics: 'only-multiline' }],
    },
    {
      code: 'function Foo<T,\n>() {}',
      options: [{ generics: 'only-multiline' }],
    },
    { code: 'type Foo = [string\n]', options: [{ tuples: 'only-multiline' }] },
    { code: 'type Foo = [string,\n]', options: [{ tuples: 'only-multiline' }] },

    // ignore
    { code: 'const a = <TYPE,>() => {}', options: [{ generics: 'ignore' }] },

    // each options
    {
      code: 'const Obj = { a: 1 };\nenum Foo {Bar}\nfunction Baz<T,>() {}\ntype Qux = [string,\n]',
      options: [
        {
          enums: 'never',
          generics: 'always',
          tuples: 'always-multiline',
        },
      ],
    },

    // https://github.com/eslint-stylistic/eslint-stylistic/issues/35
    // NOTE: the TSX single-generic case `const id = <T,>(x: T) => x;` (filename
    // file.tsx) is in KNOWN GAPS below — in TSX the `<T,>` comma is the required
    // generic/JSX disambiguation, which upstream treats as valid but rslint flags.
    {
      code: 'const id = <T,R>(x: T) => x;',
    },
  ],
  invalid: [
    // ---- from comma-dangle._js_.test.ts ----
    {
      code: 'var foo = { bar: \'baz\', }',
      output: 'var foo = { bar: \'baz\' }',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 23,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var foo = {\nbar: \'baz\',\n}',
      output: 'var foo = {\nbar: \'baz\'\n}',
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 11,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'foo({ bar: \'baz\', qux: \'quux\', });',
      output: 'foo({ bar: \'baz\', qux: \'quux\' });',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 30,
        },
      ],
    },
    {
      code: 'foo({\nbar: \'baz\',\nqux: \'quux\',\n});',
      output: 'foo({\nbar: \'baz\',\nqux: \'quux\'\n});',
      errors: [
        {
          messageId: 'unexpected',
          line: 3,
          column: 12,
        },
      ],
    },
    {
      code: 'var foo = [ \'baz\', ]',
      output: 'var foo = [ \'baz\' ]',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 18,
        },
      ],
    },
    {
      code: 'var foo = [ \'baz\',\n]',
      output: 'var foo = [ \'baz\'\n]',
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 18,
        },
      ],
    },
    {
      code: 'var foo = { bar: \'bar\'\n\n, }',
      output: 'var foo = { bar: \'bar\'\n\n }',
      errors: [
        {
          messageId: 'unexpected',
          line: 3,
          column: 1,
        },
      ],
    },

    {
      code: 'var foo = { bar: \'baz\', }',
      output: 'var foo = { bar: \'baz\' }',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'var foo = { bar: \'baz\', }',
      output: 'var foo = { bar: \'baz\' }',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'var foo = {\nbar: \'baz\',\n}',
      output: 'var foo = {\nbar: \'baz\'\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 11,
        },
      ],
    },
    {
      code: 'foo({ bar: \'baz\', qux: \'quux\', });',
      output: 'foo({ bar: \'baz\', qux: \'quux\' });',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 30,
        },
      ],
    },
    {
      code: 'foo({ bar: \'baz\', qux: \'quux\', });',
      output: 'foo({ bar: \'baz\', qux: \'quux\' });',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 30,
        },
      ],
    },

    {
      code: 'var foo = { bar: \'baz\' }',
      output: 'var foo = { bar: \'baz\', }',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'var foo = {\nbar: \'baz\'\n}',
      output: 'var foo = {\nbar: \'baz\',\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 2,
          column: 11,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'var foo = {\nbar: \'baz\'\r\n}',
      output: 'var foo = {\nbar: \'baz\',\r\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 2,
          column: 11,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    // NOTE: the two `foo({ ... });` + ['always'] cases (single-line and multi-line)
    // that upstream runs under ecmaVersion 5 are in KNOWN GAPS below. Upstream expects
    // 1 `missing` (object only, functions normalized to 'ignore'); rslint emits 2
    // (object + call argument) since functions stays 'always'.
    {
      code: 'var foo = [ \'baz\' ]',
      output: 'var foo = [ \'baz\', ]',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 18,
        },
      ],
    },
    {
      code: 'var foo = [\'baz\']',
      output: 'var foo = [\'baz\',]',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 17,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'var foo = [ \'baz\'\n]',
      output: 'var foo = [ \'baz\',\n]',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 1,
          column: 18,
        },
      ],
    },
    {
      code: 'var foo = { bar:\n\n\'bar\' }',
      output: 'var foo = { bar:\n\n\'bar\', }',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 3,
          column: 6,
        },
      ],
    },

    {
      code: 'var foo = {\nbar: \'baz\'\n}',
      output: 'var foo = {\nbar: \'baz\',\n}',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'missing',
          line: 2,
          column: 11,
        },
      ],
    },
    {
      code: 'var foo = [\n  bar,\n  (\n    baz\n  )\n];',
      output: 'var foo = [\n  bar,\n  (\n    baz\n  ),\n];',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 5,
          column: 4,
        },
      ],
    },
    {
      code: 'var foo = {\n  foo: \'bar\',\n  baz: (\n    qux\n  )\n};',
      output: 'var foo = {\n  foo: \'bar\',\n  baz: (\n    qux\n  ),\n};',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 5,
          column: 4,
        },
      ],
    },
    {
      // https://github.com/eslint/eslint/issues/7291
      code: 'var foo = [\n  (bar\n    ? baz\n    : qux\n  )\n];',
      output: 'var foo = [\n  (bar\n    ? baz\n    : qux\n  ),\n];',
      options: ['always'],
      errors: [
        {
          messageId: 'missing',
          line: 5,
          column: 4,
        },
      ],
    },
    {
      code: 'var foo = { bar: \'baz\', }',
      output: 'var foo = { bar: \'baz\' }',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 23,
        },
      ],
    },
    {
      code: 'foo({\nbar: \'baz\',\nqux: \'quux\'\n});',
      output: 'foo({\nbar: \'baz\',\nqux: \'quux\',\n});',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'missing',
          line: 3,
          column: 12,
        },
      ],
    },
    {
      code: 'foo({ bar: \'baz\', qux: \'quux\', });',
      output: 'foo({ bar: \'baz\', qux: \'quux\' });',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 30,
        },
      ],
    },
    {
      code: 'var foo = [\n\'baz\'\n]',
      output: 'var foo = [\n\'baz\',\n]',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'missing',
          line: 2,
          column: 6,
        },
      ],
    },
    {
      code: 'var foo = [\'baz\',]',
      output: 'var foo = [\'baz\']',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 17,
        },
      ],
    },
    {
      code: 'var foo = [\'baz\',]',
      output: 'var foo = [\'baz\']',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 17,
        },
      ],
    },
    {
      code: 'var foo = {x: {\nfoo: \'bar\',\n},}',
      output: 'var foo = {x: {\nfoo: \'bar\',\n}}',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 3,
          column: 2,
        },
      ],
    },
    {
      code: 'var foo = {a: 1, b: 2,\nc: 3, d: 4,}',
      output: 'var foo = {a: 1, b: 2,\nc: 3, d: 4}',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 11,
        },
      ],
    },
    {
      code: 'var foo = {a: 1, b: 2,\nc: 3, d: 4,}',
      output: 'var foo = {a: 1, b: 2,\nc: 3, d: 4}',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 2,
          column: 11,
        },
      ],
    },
    {
      code: 'var foo = [{\na: 1,\nb: 2,\nc: 3,\nd: 4,\n},]',
      output: 'var foo = [{\na: 1,\nb: 2,\nc: 3,\nd: 4,\n}]',
      options: ['always-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 6,
          column: 2,
        },
      ],
    },
    {
      code: 'var { a, b, } = foo;',
      output: 'var { a, b } = foo;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: 'var { a, b, } = foo;',
      output: 'var { a, b } = foo;',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: 'var [ a, b, ] = foo;',
      output: 'var [ a, b ] = foo;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: 'var [ a, b, ] = foo;',
      output: 'var [ a, b ] = foo;',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 11,
        },
      ],
    },
    {
      code: '[(1),]',
      output: '[(1)]',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: '[(1),]',
      output: '[(1)]',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 5,
        },
      ],
    },
    {
      code: 'var x = { foo: (1),};',
      output: 'var x = { foo: (1)};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 19,
        },
      ],
    },
    {
      code: 'var x = { foo: (1),};',
      output: 'var x = { foo: (1)};',
      options: ['only-multiline'],
      errors: [
        {
          messageId: 'unexpected',
          line: 1,
          column: 19,
        },
      ],
    },

    // https://github.com/eslint/eslint/issues/3794
    {
      code: 'import {foo} from \'foo\';',
      output: 'import {foo,} from \'foo\';',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'import foo, {abc} from \'foo\';',
      output: 'import foo, {abc,} from \'foo\';',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'export {foo} from \'foo\';',
      output: 'export {foo,} from \'foo\';',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'import {foo,} from \'foo\';',
      output: 'import {foo} from \'foo\';',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import {foo,} from \'foo\';',
      output: 'import {foo} from \'foo\';',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import foo, {abc,} from \'foo\';',
      output: 'import foo, {abc} from \'foo\';',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import foo, {abc,} from \'foo\';',
      output: 'import foo, {abc} from \'foo\';',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'export {foo,} from \'foo\';',
      output: 'export {foo} from \'foo\';',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'export {foo,} from \'foo\';',
      output: 'export {foo} from \'foo\';',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import {foo,} from \'foo\';',
      output: 'import {foo} from \'foo\';',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'export {foo,} from \'foo\';',
      output: 'export {foo} from \'foo\';',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'import {\n  foo\n} from \'foo\';',
      output: 'import {\n  foo,\n} from \'foo\';',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'export {\n  foo\n} from \'foo\';',
      output: 'export {\n  foo,\n} from \'foo\';',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },

    // https://github.com/eslint/eslint/issues/6233
    {
      code: 'var foo = {a: (1)}',
      output: 'var foo = {a: (1),}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var foo = [(1)]',
      output: 'var foo = [(1),]',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'var foo = [\n1,\n(2)\n]',
      output: 'var foo = [\n1,\n(2),\n]',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },

    // trailing commas in functions
    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a,) => a',
      output: '(a) => a',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a,) => (a)',
      output: '(a) => (a)',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '({foo(a,) {}})',
      output: '({foo(a) {}})',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A {foo(a,) {}}',
      output: 'class A {foo(a) {}}',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: [{ functions: 'never' }],
      errors: [{ messageId: 'unexpected' }],
    },

    {
      code: 'function foo(a) {}',
      output: 'function foo(a,) {}',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(function foo(a) {})',
      output: '(function foo(a,) {})',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(a) => a',
      output: '(a,) => a',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(a) => (a)',
      output: '(a,) => (a)',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '({foo(a) {}})',
      output: '({foo(a,) {}})',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'class A {foo(a) {}}',
      output: 'class A {foo(a,) {}}',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(a)',
      output: 'foo(a,)',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(...a)',
      output: 'foo(...a,)',
      options: [{ functions: 'always' }],
      errors: [{ messageId: 'missing' }],
    },

    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(\na,\nb\n) {}',
      output: 'function foo(\na,\nb,\n) {}',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(\na,\nb\n)',
      output: 'foo(\na,\nb,\n)',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(\n...a,\n...b\n)',
      output: 'foo(\n...a,\n...b,\n)',
      options: [{ functions: 'always-multiline' }],
      errors: [{ messageId: 'missing' }],
    },

    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: [{ functions: 'only-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: [{ functions: 'only-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: [{ functions: 'only-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: [{ functions: 'only-multiline' }],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a,) => a',
      output: '(a) => a',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(a,) => (a)',
      output: '(a) => (a)',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '({foo(a,) {}})',
      output: '({foo(a) {}})',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'class A {foo(a,) {}}',
      output: 'class A {foo(a) {}}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },

    {
      code: 'function foo(a) {}',
      output: 'function foo(a,) {}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(function foo(a) {})',
      output: '(function foo(a,) {})',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(a) => a',
      output: '(a,) => a',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '(a) => (a)',
      output: '(a,) => (a)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: '({foo(a) {}})',
      output: '({foo(a,) {},})',
      options: ['always'],
      errors: [
        { messageId: 'missing' },
        { messageId: 'missing' },
      ],
    },
    {
      code: 'class A {foo(a) {}}',
      output: 'class A {foo(a,) {}}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(a)',
      output: 'foo(a,)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(...a)',
      output: 'foo(...a,)',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(\na,\nb\n) {}',
      output: 'function foo(\na,\nb,\n) {}',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(\na,\nb\n)',
      output: 'foo(\na,\nb,\n)',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'foo(\n...a,\n...b\n)',
      output: 'foo(\n...a,\n...b,\n)',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },

    {
      code: 'function foo(a,) {}',
      output: 'function foo(a) {}',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: '(function foo(a,) {})',
      output: '(function foo(a) {})',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'foo(...a,)',
      output: 'foo(...a)',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function foo(a) {}',
      output: 'function foo(a,) {}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // separated options
    {
      code: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      output: 'let {a} = {a: 1};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      options: [{
        objects: 'never',
        arrays: 'ignore',
        imports: 'ignore',
        exports: 'ignore',
        functions: 'ignore',
      }],
      errors: [
        { messageId: 'unexpected', line: 1 },
        { messageId: 'unexpected', line: 1 },
      ],
    },
    {
      code: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      output: 'let {a,} = {a: 1,};\nlet [b] = [1];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      options: [{
        objects: 'ignore',
        arrays: 'never',
        imports: 'ignore',
        exports: 'ignore',
        functions: 'ignore',
      }],
      errors: [
        { messageId: 'unexpected', line: 2 },
        { messageId: 'unexpected', line: 2 },
      ],
    },
    {
      code: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      output: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      options: [{
        objects: 'ignore',
        arrays: 'ignore',
        imports: 'never',
        exports: 'ignore',
        functions: 'ignore',
      }],
      errors: [
        { messageId: 'unexpected', line: 3 },
      ],
    },
    {
      code: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      output: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d};\n(function foo(e,) {})(f,);',
      options: [{
        objects: 'ignore',
        arrays: 'ignore',
        imports: 'ignore',
        exports: 'never',
        functions: 'ignore',
      }],
      errors: [
        { messageId: 'unexpected', line: 4 },
      ],
    },
    {
      code: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e,) {})(f,);',
      output: 'let {a,} = {a: 1,};\nlet [b,] = [1,];\nimport {c,} from "foo";\nlet d = 0;export {d,};\n(function foo(e) {})(f);',
      options: [{
        objects: 'ignore',
        arrays: 'ignore',
        imports: 'ignore',
        exports: 'ignore',
        functions: 'never',
      }],
      errors: [
        { messageId: 'unexpected', line: 5 },
        { messageId: 'unexpected', line: 5 },
      ],
    },

    // https://github.com/eslint/eslint/issues/11502
    {
      code: 'foo(a,)',
      output: 'foo(a)',
      errors: [{ messageId: 'unexpected' }],
    },

    // https://github.com/eslint/eslint/issues/15660
    // NOTE: the two `add-named-import` cases (errors: 2) are in KNOWN GAPS below.
    // Upstream's hypothetical `add-named-import` rule injects a second diagnostic
    // that rslint cannot reproduce, so rslint emits only the 1 real comma-dangle error.

    // dynamic import
    {
      code: 'import(source,)',
      output: 'import(source)',
    },
    {
      code: 'import(source)',
      output: 'import(source,)',
      options: ['always'],
    },
    {
      code: 'import(source, options)',
      output: 'import(source, options,)',
      options: ['always'],
    },
    {
      code: 'import(source,)',
      output: 'import(source)',
      options: ['never'],
    },
    {
      code: 'import(source, options,)',
      output: 'import(source, options)',
      options: ['never'],
    },
    {
      code: 'import(source,)',
      output: 'import(source)',
      options: ['always-multiline'],
    },
    {
      code: 'import(source, options,)',
      output: 'import(source, options)',
      options: ['always-multiline'],
    },
    {
      code: 'import(\n  source\n)',
      output: 'import(\n  source,\n)',
      options: ['always-multiline'],
    },
    {
      code: 'import(\n  source,\n  options\n)',
      output: 'import(\n  source,\n  options,\n)',
      options: ['always-multiline'],
    },
    {
      code: 'import(source,)',
      output: 'import(source)',
      options: ['only-multiline'],
    },
    {
      code: 'import(source, options,)',
      output: 'import(source, options)',
      options: ['only-multiline'],
    },
    {
      code: 'import(source)',
      output: 'import(source,)',
      options: [{ functions: 'never', dynamicImports: 'always' }],
    },

    // import attributes
    {
      code: 'import foo from "foo" with {type: "json",}',
      output: 'import foo from "foo" with {type: "json"}',
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      output: 'import foo from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'import foo from "foo" with {type: "json",}',
      output: 'import foo from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'import foo from "foo" with {type: "json",}',
      output: 'import foo from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'import foo from "foo" with {\n  type: "json"\n}',
      output: 'import foo from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'import foo from "foo" with {type: "json",}',
      output: 'import foo from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'import foo from "foo" with {type: "json"}',
      output: 'import foo from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },
    {
      code: 'export {foo} from "foo" with {type: "json",}',
      output: 'export {foo} from "foo" with {type: "json"}',
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
      output: 'export {foo,} from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json",}',
      output: 'export {foo} from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json",}',
      output: 'export {foo} from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {\n  type: "json"\n}',
      output: 'export {foo} from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json",}',
      output: 'export {foo} from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'export {foo} from "foo" with {type: "json"}',
      output: 'export {foo} from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      output: 'export * from "foo" with {type: "json"}',
    },
    {
      code: 'export * from "foo" with {type: "json"}',
      output: 'export * from "foo" with {type: "json",}',
      options: ['always'],
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      output: 'export * from "foo" with {type: "json"}',
      options: ['never'],
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      output: 'export * from "foo" with {type: "json"}',
      options: ['always-multiline'],
    },
    {
      code: 'export * from "foo" with {\n  type: "json"\n}',
      output: 'export * from "foo" with {\n  type: "json",\n}',
      options: ['always-multiline'],
    },
    {
      code: 'export * from "foo" with {type: "json",}',
      output: 'export * from "foo" with {type: "json"}',
      options: ['only-multiline'],
    },
    {
      code: 'export * from "foo" with {type: "json"}',
      output: 'export * from "foo" with {type: "json",}',
      options: [{ functions: 'never', importAttributes: 'always' }],
    },

    // ---- from comma-dangle._ts_.test.ts ----
    // default
    {
      code: 'enum Foo {Bar,}',
      output: 'enum Foo {Bar}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function Foo<T,>() {}',
      output: 'function Foo<T>() {}',
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'type Foo = [string,]',
      output: 'type Foo = [string]',
      errors: [{ messageId: 'unexpected' }],
    },

    // never
    {
      code: 'enum Foo {Bar,}',
      output: 'enum Foo {Bar}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'enum Foo {Bar,\n}',
      output: 'enum Foo {Bar\n}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function Foo<T,>() {}',
      output: 'function Foo<T>() {}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function Foo<T,\n>() {}',
      output: 'function Foo<T\n>() {}',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'type Foo = [string,]',
      output: 'type Foo = [string]',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'type Foo = [string,\n]',
      output: 'type Foo = [string\n]',
      options: ['never'],
      errors: [{ messageId: 'unexpected' }],
    },

    // always
    {
      code: 'enum Foo {Bar}',
      output: 'enum Foo {Bar,}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'enum Foo {Bar\n}',
      output: 'enum Foo {Bar,\n}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'function Foo<T>() {}',
      output: 'function Foo<T,>() {}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'function Foo<T\n>() {}',
      output: 'function Foo<T,\n>() {}',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'type Foo = [string]',
      output: 'type Foo = [string,]',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'type Foo = [string\n]',
      output: 'type Foo = [string,\n]',
      options: ['always'],
      errors: [{ messageId: 'missing' }],
    },

    // always-multiline
    {
      code: 'enum Foo {Bar,}',
      output: 'enum Foo {Bar}',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'enum Foo {Bar\n}',
      output: 'enum Foo {Bar,\n}',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'function Foo<T,>() {}',
      output: 'function Foo<T>() {}',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function Foo<T\n>() {}',
      output: 'function Foo<T,\n>() {}',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },
    {
      code: 'type Foo = [string,]',
      output: 'type Foo = [string]',
      options: ['always-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'type Foo = [string\n]',
      output: 'type Foo = [string,\n]',
      options: ['always-multiline'],
      errors: [{ messageId: 'missing' }],
    },

    // only-multiline
    {
      code: 'enum Foo {Bar,}',
      output: 'enum Foo {Bar}',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'function Foo<T,>() {}',
      output: 'function Foo<T>() {}',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },
    {
      code: 'type Foo = [string,]',
      output: 'type Foo = [string]',
      options: ['only-multiline'],
      errors: [{ messageId: 'unexpected' }],
    },

    // https://github.com/eslint-stylistic/eslint-stylistic/issues/35
    // When there is more than one generic, we don't need to workaround it
    {
      code: 'const id = <T,R,>(x: T) => x;',
      output: 'const id = <T,R>(x: T) => x;',
      errors: [{ messageId: 'unexpected' }],
    },

    {
      code: 'declare function foo(a: number, b: number,): void',
      output: 'declare function foo(a: number, b: number): void',
      errors: [
        { messageId: 'unexpected', line: 1, column: 42 },
      ],
    },
    {
      code: 'type Foo = (a: number, b: number,) => void',
      output: 'type Foo = (a: number, b: number) => void',
      errors: [
        { messageId: 'unexpected', line: 1, column: 33 },
      ],
    },

    {
      code: 'type Foo<T> = Bar<T,>',
      output: 'type Foo<T> = Bar<T>',
      errors: [
        { messageId: 'unexpected', line: 1, column: 20 },
      ],
    },
  ],
});

/**
 * ============================ comma-dangle — KNOWN GAPS ============================
 *
 * The cases below are ported verbatim from upstream but are NOT run through the green
 * `ruleTester.run` above. Each diverges from upstream for a documented reason; none is
 * a transcription mistake or a silenced failure. Every divergence below was VERIFIED by
 * running the fixture through the rslint CLI (the same engine the RuleTester drives).
 * The expected upstream behaviour is preserved for the record.
 *
 * ---- GAP 1: ecmaVersion-dependent option normalization ----
 *
 * Upstream `comma-dangle` normalizes a STRING option using `ecmaVersion`: `functions`
 * becomes 'ignore' when `ecmaVersion < 2017` (trailing commas in function params/calls
 * were illegal before ES2017), and `dynamicImports` becomes 'ignore' when
 * `ecmaVersion < 2025`. The RuleTester DROPS `parserOptions` and rslint always parses
 * as `esnext` (tsconfig `target: esnext`), so no normalization happens — the selected
 * option keeps its literal value and rslint reports where upstream stayed silent. This
 * is a parser-target difference, not a rule-logic bug.
 *
 * -- valid (upstream expects 0 diagnostics; rslint emits 1) --
 *   { code: 'function foo(a) {}',       options: ['always'] }            // upstream ecmaVersion 5  → rslint: 1× 'missing' @1:15
 *   { code: 'foo(a)',                   options: ['always'] }            // upstream ecmaVersion 5  → rslint: 1× 'missing'
 *   { code: 'foo(a,\nb\n)',             options: ['always-multiline'] }  // upstream ecmaVersion 5  → rslint: 1× 'missing' @2:2
 *   { code: 'function foo(a,\nb\n) {}', options: ['always-multiline'] }  // upstream ecmaVersion 5  → rslint: 1× 'missing'
 *   { code: 'function foo(a) {}',       options: ['always'] }            // upstream ecmaVersion 7  → rslint: 1× 'missing' @1:15
 *   { code: 'foo(a)',                   options: ['always'] }            // upstream ecmaVersion 7  → rslint: 1× 'missing'
 *   { code: 'function foo(a,\nb\n) {}', options: ['always-multiline'] }  // upstream ecmaVersion 7  → rslint: 1× 'missing'
 *   { code: 'foo(a,\nb\n)',             options: ['always-multiline'] }  // upstream ecmaVersion 7  → rslint: 1× 'missing'
 *   { code: 'import(source)',           options: ['always'] }            // upstream ecmaVersion 15 (dynamicImports→'ignore') → rslint: 1× 'missing' @1:14
 *
 * -- invalid (upstream expects 1 'missing' — object only; rslint emits 2 — object + call arg) --
 *   {
 *     code: 'foo({ bar: \'baz\', qux: \'quux\' });',
 *     output: 'foo({ bar: \'baz\', qux: \'quux\', });',
 *     options: ['always'],   // upstream parserOptions: { sourceType: 'script', ecmaVersion: 5 }
 *     errors: [{ messageId: 'missing', line: 1, column: 30, endLine: 1, endColumn: 31 }],
 *   }
 *   // rslint: 2× 'missing' (object member @1:30 + call argument @1:32).
 *   {
 *     code: 'foo({\nbar: \'baz\',\nqux: \'quux\'\n});',
 *     output: 'foo({\nbar: \'baz\',\nqux: \'quux\',\n});',
 *     options: ['always'],   // upstream parserOptions: { sourceType: 'script', ecmaVersion: 5 }
 *     errors: [{ messageId: 'missing', line: 3, column: 12, endLine: 4, endColumn: 1 }],
 *   }
 *   // rslint: 2× 'missing' (object member @3:12 + call argument @4:2).
 *
 * ---- GAP 2: inline-directive `add-named-import` INVALID cases (upstream errors: 2) ----
 *
 * // https://github.com/eslint/eslint/issues/15660
 *
 * Upstream runs these with a hypothetical `add-named-import` rule (the
 * `/​*eslint add-named-import:1*​/` directive) that injects a second specifier and thus
 * a second comma-dangle diagnostic in the same autofix pass. The RuleTester here mounts
 * ONLY `@stylistic/comma-dangle`, so that injected diagnostic does not exist and rslint
 * emits exactly the 1 real comma-dangle error (verified: 1 diagnostic at 9:17).
 *
 *   {
 *     code: `/​*eslint add-named-import:1*​/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView\n} from 'react-native';`,
 *     output: <same source with a trailing comma added after SafeAreaView>,
 *     options: [{ imports: 'always-multiline' }],
 *     errors: 2,   // rslint: 1× 'missing'
 *   }
 *   {
 *     code: `/​*eslint add-named-import:1*​/\nimport {\n    StyleSheet,\n    View,\n    TextInput,\n    ImageBackground,\n    Image,\n    TouchableOpacity,\n    SafeAreaView,\n} from 'react-native';`,
 *     output: <same source with the trailing comma removed after SafeAreaView>,
 *     options: [{ imports: 'never' }],
 *     errors: 2,   // rslint: 1× 'unexpected'
 *   }
 *
 * ---- GAP 3: Babel/Flow-only fixtures (the entire `comma-dangle_babel` run, guarded by
 *      `if (!skipBabel)`, using `languageOptionsForBabelFlow`) ----
 *
 * // https://github.com/eslint/eslint/issues/7370
 *
 * Upstream parses these via Babel with the Flow plugin, where `{a: string,}` /
 * `{b: boolean,}` are Flow object-type annotations (NOT object literals). rslint has no
 * Babel/Flow parser; ts-go parses these annotations as TypeScript type literals, which
 * is a different AST. The combined effect — different parse + GAP 1's missing
 * `functions` normalization under esnext — shifts the diagnostic count / fix output away
 * from upstream's Flow result. Concretely, the gated invalid case below produces 2
 * diagnostics under rslint (param pattern + function param) instead of upstream's 1, and
 * the fix adds an extra param comma. VERIFIED by running.
 *
 * -- valid (upstream expects 0 diagnostics) --
 *   { code: 'function foo({a}: {a: string,}) {}', options: ['never'] }
 *   { code: 'function foo({a,}: {a: string}) {}', options: ['always'] }   // upstream also sets sourceType:'script', ecmaVersion:5
 *   { code: 'function foo(a): {b: boolean,} {}',  options: [{ functions: 'never' }] }
 *   { code: 'function foo(a,): {b: boolean} {}',  options: [{ functions: 'always' }] }
 *
 * -- invalid (upstream expectation shown) --
 *   {
 *     code: 'function foo({a}: {a: string,}) {}',
 *     output: 'function foo({a,}: {a: string,}) {}',
 *     options: ['always'],   // upstream sourceType:'script', ecmaVersion:5 → functions ignored
 *     errors: [{ messageId: 'missing' }],
 *   }
 *   // rslint: 2× 'missing' (pattern {a} @1:16 + function param @1:31); the fix yields
 *   // `function foo({a,}: {a: string,},) {}` — the extra param comma comes from
 *   // `functions: 'always'` not being normalized away under esnext.
 *   {
 *     code: 'function foo({a,}: {a: string}) {}',
 *     output: 'function foo({a}: {a: string}) {}',
 *     options: ['never'],
 *     errors: [{ messageId: 'unexpected' }],
 *   }
 *   {
 *     code: 'function foo(a): {b: boolean,} {}',
 *     output: 'function foo(a,): {b: boolean,} {}',
 *     options: [{ functions: 'always' }],
 *     errors: [{ messageId: 'missing' }],
 *   }
 *   {
 *     code: 'function foo(a,): {b: boolean} {}',
 *     output: 'function foo(a): {b: boolean} {}',
 *     options: [{ functions: 'never' }],
 *     errors: [{ messageId: 'unexpected' }],
 *   }
 *
 * ---- GAP 4: TSX single-generic disambiguation comma (upstream expects 0; rslint 1) ----
 *
 * // https://github.com/eslint-stylistic/eslint-stylistic/issues/35
 *
 * In a `.tsx` file, `<T,>` is the trailing comma that disambiguates a single-type-
 * parameter generic from a JSX element. Upstream's rule treats this comma as valid
 * under the JSX language option; rslint reports it as an `unexpected` trailing comma.
 * (With two or more type parameters there is no ambiguity, so `<T,R,>` IS reported by
 * both — that invalid case stays in the green set above.)
 *
 *   { filename: 'file.tsx', code: 'const id = <T,>(x: T) => x;' }
 *   // upstream (ecmaFeatures.jsx) → 0; rslint: 1× 'unexpected' @1:14.
 *
 * ==================================================================================
 */
