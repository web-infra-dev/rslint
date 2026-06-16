/**
 * @fileoverview Tests for quotes rule.
 * @author Matt DuVall <http://www.mattduvall.com/>, Michael Paulukonis
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/quotes/quotes._js_.test.ts
 *   packages/eslint-plugin/rules/quotes/quotes._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('quotes', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string.
 *  - The local error helpers (`useDoubleQuote` / `useSingleQuote` / `useBacktick`)
 *    are inlined to their final `{ messageId: 'wrongQuotes', data: { description } }`.
 *  - `parserOptions` (ecmaVersion / ecmaFeatures.jsx / sourceType) dropped — rslint
 *    resolves via tsconfig; the RuleTester picks a `.tsx` fixture when JSX is present.
 *  - `type` fields (deprecated AST node type) dropped.
 *
 * No Babel/Flow cases and no external-fixture (`readFileSync`) cases exist in the
 * upstream quotes tests, so nothing was skipped on those grounds. The `._css_` /
 * `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are moved to `describe('quotes — KNOWN GAPS', ...)` at the bottom, each annotated
 * with what upstream expects vs. what rslint produces.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('quotes', null as never, {
  valid: [
    // ---- from quotes._js_.test.ts ----
    'var foo = "bar";',
    { code: 'var foo = \'bar\';', options: ['single'] },
    { code: 'var foo = "bar";', options: ['double'] },
    { code: 'var foo = 1;', options: ['single'] },
    { code: 'var foo = 1;', options: ['double'] },
    { code: 'var foo = \'bar\';', options: ['double', { ignoreStringLiterals: true }] },
    { code: 'var foo = "bar";', options: ['single', { ignoreStringLiterals: true }] },
    { code: 'var foo = "\'";', options: ['single', { avoidEscape: true }] },
    { code: 'var foo = \'"\';', options: ['double', { avoidEscape: true }] },
    { code: 'var foo = `\'`;', options: ['single', { avoidEscape: true, allowTemplateLiterals: true }] },
    { code: 'var foo = `\'`;', options: ['single', { avoidEscape: true, allowTemplateLiterals: 'always' }] },
    { code: 'var foo = `\'`;', options: ['single', { avoidEscape: true, allowTemplateLiterals: 'avoidEscape' }] },
    { code: 'var foo = `"`;', options: ['double', { avoidEscape: true, allowTemplateLiterals: true }] },
    { code: 'var foo = `"`;', options: ['double', { avoidEscape: true, allowTemplateLiterals: 'always' }] },
    { code: 'var foo = `"`;', options: ['double', { avoidEscape: true, allowTemplateLiterals: 'avoidEscape' }] },
    { code: 'var foo = <>Hello world</>;', options: ['single'] },
    { code: 'var foo = <>Hello world</>;', options: ['double'] },
    { code: 'var foo = <>Hello world</>;', options: ['double', { avoidEscape: true }] },
    { code: 'var foo = <>Hello world</>;', options: ['backtick'] },
    { code: 'var foo = <div>Hello world</div>;', options: ['single'] },
    { code: 'var foo = <div id="foo"></div>;', options: ['single'] },
    { code: 'var foo = <div>Hello world</div>;', options: ['double'] },
    { code: 'var foo = <div>Hello world</div>;', options: ['double', { avoidEscape: true }] },
    { code: 'var foo = `bar`;', options: ['backtick'] },
    { code: 'var foo = `bar \'baz\'`;', options: ['backtick'] },
    { code: 'var foo = `bar "baz"`;', options: ['backtick'] },
    { code: 'var foo = 1;', options: ['backtick'] },
    { code: 'var foo = "a string containing `backtick` quotes";', options: ['backtick', { avoidEscape: true }] },
    { code: 'var foo = <div id="foo"></div>;', options: ['backtick'] },
    { code: 'var foo = <div>Hello world</div>;', options: ['backtick'] },
    { code: 'class C { "f"; "m"() {} }', options: ['double'] },
    { code: 'class C { \'f\'; \'m\'() {} }', options: ['single'] },

    // Backticks are only okay if they have substitutions, contain a line break, or are tagged
    { code: 'var foo = `back\ntick`;', options: ['single'] },
    { code: 'var foo = `back\rtick`;', options: ['single'] },
    { code: 'var foo = `back tick`;', options: ['single'] },
    { code: 'var foo = `back tick`;', options: ['single'] },
    {
      code: 'var foo = `back\\\\\ntick`;', // 2 backslashes followed by a newline
      options: ['single'],
    },
    { code: 'var foo = `back\\\\\\\\\ntick`;', options: ['single'] },
    { code: 'var foo = `\n`;', options: ['single'] },
    { code: 'var foo = `back${x}tick`;', options: ['double'] },
    { code: 'var foo = tag`backtick`;', options: ['double'] },

    // Backticks are also okay if allowTemplateLiterals
    { code: 'var foo = `bar \'foo\' baz` + \'bar\';', options: ['single', { allowTemplateLiterals: true }] },
    { code: 'var foo = `bar \'foo\' baz` + \'bar\';', options: ['single', { allowTemplateLiterals: 'always' }] },
    { code: 'var foo = `bar \'foo\' baz` + "bar";', options: ['double', { allowTemplateLiterals: true }] },
    { code: 'var foo = `bar \'foo\' baz` + "bar";', options: ['double', { allowTemplateLiterals: 'always' }] },
    { code: 'var foo = `bar \'foo\' baz` + `bar`;', options: ['backtick', { allowTemplateLiterals: true }] },
    { code: 'var foo = `bar \'foo\' baz` + `bar`;', options: ['backtick', { allowTemplateLiterals: 'always' }] },

    // `backtick` should not warn the directive prologues.
    { code: '"use strict"; var foo = `backtick`;', options: ['backtick'] },
    { code: '"use strict"; \'use strong\'; "use asm"; var foo = `backtick`;', options: ['backtick'] },
    { code: 'function foo() { "use strict"; "use strong"; "use asm"; var foo = `backtick`; }', options: ['backtick'] },
    { code: '(function() { \'use strict\'; \'use strong\'; \'use asm\'; var foo = `backtick`; })();', options: ['backtick'] },
    { code: '(() => { "use strict"; "use strong"; "use asm"; var foo = `backtick`; })();', options: ['backtick'] },

    // `backtick` should not warn import/export sources.
    { code: 'import "a"; import \'b\';', options: ['backtick'] },
    { code: 'import a from "a"; import b from \'b\';', options: ['backtick'] },
    { code: 'export * from "a"; export * from \'b\';', options: ['backtick'] },

    // `backtick` should not warn module export names.
    { code: 'import { "a" as b, \'c\' as d } from \'mod\';', options: ['backtick'] },
    { code: 'let a, c; export { a as "b", c as \'d\' };', options: ['backtick'] },
    { code: 'export { "a", \'b\' } from \'mod\';', options: ['backtick'] },
    { code: 'export { a as "b", c as \'d\' } from \'mod\';', options: ['backtick'] },
    { code: 'export { "a" as b, \'c\' as d } from \'mod\';', options: ['backtick'] },
    { code: 'export { "a" as "b", \'c\' as \'d\' } from \'mod\';', options: ['backtick'] },
    { code: 'export * as "a" from \'mod\';', options: ['backtick'] },
    { code: 'export * as \'a\' from \'mod\';', options: ['backtick'] },

    // `backtick` should not warn property/method names (not computed).
    { code: 'var obj = {"key0": 0, \'key1\': 1};', options: ['backtick'] },
    { code: 'class Foo { \'bar\'(){} }', options: ['backtick'] },
    { code: 'class Foo { static \'\'(){} }', options: ['backtick'] },
    { code: 'class C { "double"; \'single\'; }', options: ['backtick'] },

    // ---- from quotes._ts_.test.ts ----
    {
      code: 'declare module \'*.html\' {}',
      options: ['backtick'],
    },
    {
      code: 'class A {\n  public prop: IProps[`prop`];\n}',
      options: ['backtick'],
    },

    // `backtick` should not warn import with attributes.
    // NOTE: the two `import ... assert { type: ... }` valid cases are in KNOWN GAPS
    // below — `assert` import attributes are a TypeScript SYNTAX ERROR (TS2880,
    // "Import assertions have been replaced by import attributes. Use 'with'") under
    // rslint's ts-go parser. The `with` form below is accepted and stays here.
    {
      code: 'import "a" with { type: "json" }; import \'b\' with { type: \'json\' };',
      options: ['backtick'],
    },
    {
      code: 'import a from "a" with { type: "json" }; import b from \'b\' with { type: \'json\' };',
      options: ['backtick'],
    },
    // `backtick` should not warn import with require.
    {
      code: 'import moment = require(\'moment\');',
      options: ['backtick'],
    },

    // TSPropertySignature
    {
      code: 'interface Foo {\n  a: number;\n  b: string;\n  "a-b": boolean;\n  "a-b-c": boolean;\n}',
    },
    {
      code: 'interface Foo {\n  a: number;\n  b: string;\n  \'a-b\': boolean;\n  \'a-b-c\': boolean;\n}',
      options: ['single'],
    },
    {
      code: 'interface Foo {\n  a: number;\n  b: string;\n  \'a-b\': boolean;\n  \'a-b-c\': boolean;\n}',
      options: ['backtick'],
    },

    // TSEnumMember
    {
      code: 'enum Foo {\n  A = 1,\n  "A-B" = 2\n}',
    },
    {
      code: 'enum Foo {\n  A = 1,\n  \'A-B\' = 2\n}',
      options: ['single'],
    },
    {
      code: 'enum Foo {\n  A = `A`,\n  \'A-B\' = `A-B`\n}',
      options: ['backtick'],
    },

    // TSMethodSignature
    {
      code: 'interface Foo {\n  a(): void;\n  "a-b"(): void;\n}',
    },
    {
      code: 'interface Foo {\n  a(): void;\n  \'a-b\'(): void;\n}',
      options: ['single'],
    },
    {
      code: 'interface Foo {\n  a(): void;\n  \'a-b\'(): void;\n}',
      options: ['backtick'],
    },

    // PropertyDefinition
    {
      code: 'class Foo {\n  public a = "";\n  public "a-b" = "";\n}',
    },
    {
      code: 'class Foo {\n  public a = \'\';\n  public \'a-b\' = \'\';\n}',
      options: ['single'],
    },
    {
      code: 'class Foo {\n  public a = ``;\n  public \'a-b\' = ``;\n}',
      options: ['backtick'],
    },

    // AccessorProperty
    {
      code: 'class Foo {\n  accessor a = "";\n  accessor "a-b" = "";\n}',
    },
    {
      code: 'class Foo {\n  accessor a = \'\';\n  accessor \'a-b\' = \'\';\n}',
      options: ['single'],
    },
    {
      code: 'class Foo {\n  accessor a = ``;\n  accessor \'a-b\' = ``;\n}',
      options: ['backtick'],
    },

    // TSAbstractPropertyDefinition
    {
      code: 'abstract class Foo {\n  public abstract a: "";\n  public abstract "a-b": "";\n}',
    },
    {
      code: 'abstract class Foo {\n  public abstract a: \'\';\n  public abstract \'a-b\': \'\';\n}',
      options: ['single'],
    },
    {
      code: 'abstract class Foo {\n  public abstract a: ``;\n  public abstract \'a-b\': ``;\n}',
      options: ['backtick'],
    },

    // TSAbstractMethodDefinition
    {
      code: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract "a-b"(): void;\n}',
    },
    {
      code: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract \'a-b\'(): void;\n}',
      options: ['single'],
    },
    {
      code: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract \'a-b\'(): void;\n}',
      options: ['backtick'],
    },

    // TSLiteralType
    // https://github.com/eslint-stylistic/eslint-stylistic/issues/473
    {
      code: 'type A = import(\'hi\');',
      options: ['backtick'],
    },
    {
      code: 'type A = `a` | `b`;',
      options: ['backtick'],
    },
  ],

  invalid: [
    // ---- from quotes._js_.test.ts ----
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = "bar";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = \'bar\';',
      options: ['single'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = `bar`;',
      output: 'var foo = \'bar\';',
      options: ['single'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = `bar`;',
      output: 'var foo = \'bar\';',
      options: ['single', { ignoreStringLiterals: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = \'don\\\'t\';',
      output: 'var foo = "don\'t";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var msg = "Plugin \'" + name + "\' not found"',
      output: 'var msg = \'Plugin \\\'\' + name + \'\\\' not found\'',
      options: ['single'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          column: 11,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          column: 31,
        },
      ],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = "bar";',
      options: ['double'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `bar`;',
      output: 'var foo = "bar";',
      options: ['double'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `"`;',
      output: 'var foo = "\\\"";',
      options: ['double'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `\'`;',
      output: 'var foo = \'\\\'\';',
      options: ['single'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = \'bar\';',
      options: ['single', { avoidEscape: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = "bar";',
      options: ['double', { avoidEscape: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = \'\\\\\';',
      output: 'var foo = "\\\\";',
      options: ['double', { avoidEscape: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = \'bar\';',
      options: ['single', { allowTemplateLiterals: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = \'bar\';',
      options: ['single', { allowTemplateLiterals: 'always' }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'singlequote' },
      }],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = "bar";',
      options: ['double', { allowTemplateLiterals: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = "bar";',
      options: ['double', { allowTemplateLiterals: 'always' }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = `bar`;',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: 'var foo = \'b${x}a$r\';',
      output: 'var foo = `b\\${x}a$r`;',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = `bar`;',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: 'var foo = "bar";',
      output: 'var foo = `bar`;',
      options: ['backtick', { avoidEscape: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: 'var foo = \'bar\';',
      output: 'var foo = `bar`;',
      options: ['backtick', { avoidEscape: true }],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },

    // "use strict" is *not* a directive prologue in these statements so is subject to the rule
    {
      code: 'var foo = `backtick`; "use strict";',
      output: 'var foo = `backtick`; `use strict`;',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: '{ "use strict"; var foo = `backtick`; }',
      output: '{ `use strict`; var foo = `backtick`; }',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: 'if (1) { "use strict"; var foo = `backtick`; }',
      output: 'if (1) { `use strict`; var foo = `backtick`; }',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },

    // `backtick` should warn computed property names.
    {
      code: 'var obj = {["key0"]: 0, [\'key1\']: 1};',
      output: 'var obj = {[`key0`]: 0, [`key1`]: 1};',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },
    {
      code: 'class Foo { [\'a\'](){} static [\'b\'](){} }',
      output: 'class Foo { [`a`](){} static [`b`](){} }',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },

    // https://github.com/eslint/eslint/issues/7084
    {
      code: '<div blah={"blah"} />',
      output: '<div blah={\'blah\'} />',
      options: ['single'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
        },
      ],
    },
    {
      code: '<div blah={\'blah\'} />',
      output: '<div blah={"blah"} />',
      options: ['double'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
        },
      ],
    },
    {
      code: '<div blah={\'blah\'} />',
      output: '<div blah={`blah`} />',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },

    // https://github.com/eslint/eslint/issues/7610
    {
      code: '`use strict`;',
      output: null,
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'function foo() { `use strict`; foo(); }',
      output: null,
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'foo = function() { `use strict`; foo(); }',
      output: null,
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: '() => { `use strict`; foo(); }',
      output: null,
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: '() => { foo(); `use strict`; }',
      output: null, // no autofix
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'foo(); `use strict`;',
      output: null, // no autofix
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },

    // https://github.com/eslint/eslint/issues/7646
    {
      code: 'var foo = `foo\\nbar`;',
      output: 'var foo = "foo\\nbar";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `foo\\\nbar`;', // 1 backslash followed by a newline
      output: 'var foo = "foo\\\nbar";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `foo\\\\\\\nbar`;', // 3 backslashes followed by a newline
      output: 'var foo = "foo\\\\\\\nbar";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: '````',
      output: '""``',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
        line: 1,
        column: 1,
      }],
    },

    // Strings containing octal escape sequences. Don't autofix to backticks.
    // NOTE: Most octal-escape cases here are moved to KNOWN GAPS below — an octal
    // escape (`\1`, `\01`, `\0\1`, `\08`, `\33`, `\75`, `\8`) is a TypeScript SYNTAX
    // ERROR (TS1487/TS1488) under rslint's ts-go parser (which enforces strict/module
    // ES semantics), whereas upstream runs these with `parserOptions.sourceType:
    // 'script'` (sloppy mode) where they parse fine. Only `\0` (a legal NUL escape,
    // not octal) stays in the green set.
    {
      code: 'var notoctal = \'\\0\'',
      output: 'var notoctal = `\\0`',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },
    // (octal-escape backtick cases `'\1'`, `"\1"`, `'\01'`, `'\0\1'`, `'\08'`,
    //  `'prefix \33'`, `'prefix \75 suffix'`, `'\8'` — and the `single`/`double`
    //  cases `"\1"`/`'\1'` above — are in KNOWN GAPS: TS1487/TS1488 syntax errors.)

    // class members
    {
      code: 'class C { \'foo\'; }',
      output: 'class C { "foo"; }',
      options: ['double'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
        },
      ],
    },
    {
      code: 'class C { \'foo\'() {} }',
      output: 'class C { "foo"() {} }',
      options: ['double'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
        },
      ],
    },
    {
      code: 'class C { "foo"; }',
      output: 'class C { \'foo\'; }',
      options: ['single'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
        },
      ],
    },
    {
      code: 'class C { "foo"() {} }',
      output: 'class C { \'foo\'() {} }',
      options: ['single'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
        },
      ],
    },
    {
      code: 'class C { ["foo"]; }',
      output: 'class C { [`foo`]; }',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },
    {
      code: 'class C { foo = "foo"; }',
      output: 'class C { foo = `foo`; }',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
        },
      ],
    },

    // https://github.com/eslint/eslint/pull/17022
    {
      code: '() => { foo(); (`use strict`); }',
      output: '() => { foo(); ("use strict"); }',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: '(\'foo\'); "bar";',
      output: '(`foo`); `bar`;',
      options: ['backtick'],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }, {
        messageId: 'wrongQuotes',
        data: { description: 'backtick' },
      }],
    },
    {
      code: '; \'use asm\';',
      output: '; "use asm";',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: '{ `foobar`; }',
      output: '{ "foobar"; }',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'foo(() => `bar`);',
      output: 'foo(() => "bar");',
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `"bar"`',
      output: 'var foo = "\\"bar\\""',
      options: [
        'double',
        {
          avoidEscape: true,
          allowTemplateLiterals: false,
        },
      ],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `"bar"`',
      output: 'var foo = "\\"bar\\""',
      options: [
        'double',
        {
          avoidEscape: true,
          allowTemplateLiterals: 'never',
        },
      ],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },
    {
      code: 'var foo = `bar`',
      output: 'var foo = "bar"',
      options: [
        'double',
        {
          avoidEscape: true,
          allowTemplateLiterals: 'avoidEscape',
        },
      ],
      errors: [{
        messageId: 'wrongQuotes',
        data: { description: 'doublequote' },
      }],
    },

    // ---- from quotes._ts_.test.ts ----
    {
      code: 'interface Foo {\n  a: number;\n  b: string;\n  \'a-b\': boolean;\n  \'a-b-c\': boolean;\n}',
      output: 'interface Foo {\n  a: number;\n  b: string;\n  "a-b": boolean;\n  "a-b-c": boolean;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 4,
          column: 3,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 5,
          column: 3,
        },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n  b: string;\n  "a-b": boolean;\n  "a-b-c": boolean;\n}',
      output: 'interface Foo {\n  a: number;\n  b: string;\n  \'a-b\': boolean;\n  \'a-b-c\': boolean;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 4,
          column: 3,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 5,
          column: 3,
        },
      ],
      options: ['single'],
    },

    // Enums
    {
      code: 'enum Foo {\n  A = 1,\n  \'A-B\' = 2\n}',
      output: 'enum Foo {\n  A = 1,\n  "A-B" = 2\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: 'enum Foo {\n  A = 1,\n  "A-B" = 2\n}',
      output: 'enum Foo {\n  A = 1,\n  \'A-B\' = 2\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 3,
        },
      ],
      options: ['single'],
    },
    {
      code: 'enum Foo {\n  A = \'A\',\n  \'A-B\' = \'A-B\'\n}',
      output: 'enum Foo {\n  A = `A`,\n  \'A-B\' = `A-B`\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 2,
          column: 7,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 3,
          column: 11,
        },
      ],
      options: ['backtick'],
    },

    // TSMethodSignature
    {
      code: 'interface Foo {\n  a(): void;\n  \'a-b\'(): void;\n}',
      output: 'interface Foo {\n  a(): void;\n  "a-b"(): void;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 3,
        },
      ],
    },
    {
      code: 'interface Foo {\n  a(): void;\n  "a-b"(): void;\n}',
      output: 'interface Foo {\n  a(): void;\n  \'a-b\'(): void;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 3,
        },
      ],
      options: ['single'],
    },

    // PropertyDefinition
    {
      code: 'class Foo {\n  public a = \'\';\n  public \'a-b\' = \'\';\n}',
      output: 'class Foo {\n  public a = "";\n  public "a-b" = "";\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 2,
          column: 14,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 10,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 18,
        },
      ],
    },
    {
      code: 'class Foo {\n  public a = "";\n  public "a-b" = "";\n}',
      output: 'class Foo {\n  public a = \'\';\n  public \'a-b\' = \'\';\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 2,
          column: 14,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 10,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 18,
        },
      ],
      options: ['single'],
    },
    {
      code: 'class Foo {\n  public a = "";\n  public "a-b" = "";\n}',
      output: 'class Foo {\n  public a = ``;\n  public "a-b" = ``;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 2,
          column: 14,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 3,
          column: 18,
        },
      ],
      options: ['backtick'],
    },

    // AccessorProperty
    {
      code: 'class Foo {\n  accessor a = \'\';\n  accessor \'a-b\' = \'\';\n}',
      output: 'class Foo {\n  accessor a = "";\n  accessor "a-b" = "";\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 2,
          column: 16,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 12,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 20,
        },
      ],
    },
    {
      code: 'class Foo {\n  accessor a = "";\n  accessor "a-b" = "";\n}',
      output: 'class Foo {\n  accessor a = \'\';\n  accessor \'a-b\' = \'\';\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 2,
          column: 16,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 12,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 20,
        },
      ],
      options: ['single'],
    },
    {
      code: 'class Foo {\n  accessor a = "";\n  accessor "a-b" = "";\n}',
      output: 'class Foo {\n  accessor a = ``;\n  accessor "a-b" = ``;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 2,
          column: 16,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 3,
          column: 20,
        },
      ],
      options: ['backtick'],
    },

    // TSAbstractPropertyDefinition
    {
      code: 'abstract class Foo {\n  public abstract a: \'\';\n  public abstract \'a-b\': \'\';\n}',
      output: 'abstract class Foo {\n  public abstract a: "";\n  public abstract "a-b": "";\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 2,
          column: 22,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 19,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 26,
        },
      ],
    },
    {
      code: 'abstract class Foo {\n  public abstract a: "";\n  public abstract "a-b": "";\n}',
      output: 'abstract class Foo {\n  public abstract a: \'\';\n  public abstract \'a-b\': \'\';\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 2,
          column: 22,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 19,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 26,
        },
      ],
      options: ['single'],
    },

    // TSAbstractMethodDefinition
    {
      code: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract \'a-b\'(): void;\n}',
      output: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract "a-b"(): void;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'doublequote' },
          line: 3,
          column: 19,
        },
      ],
    },
    {
      code: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract "a-b"(): void;\n}',
      output: 'abstract class Foo {\n  public abstract a(): void;\n  public abstract \'a-b\'(): void;\n}',
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'singlequote' },
          line: 3,
          column: 19,
        },
      ],
      options: ['single'],
    },

    // TSLiteralType
    {
      code: 'type A = "a" | "b";',
      output: 'type A = `a` | `b`;',
      options: ['backtick'],
      errors: [
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 1,
          column: 10,
        },
        {
          messageId: 'wrongQuotes',
          data: { description: 'backtick' },
          line: 1,
          column: 16,
        },
      ],
    },
  ],
});

/**
 * ============================ quotes — KNOWN GAPS ============================
 *
 * The cases below are ported verbatim from upstream but are NOT run through the
 * green `ruleTester.run` above, because each is a *parser-level* incompatibility:
 * rslint parses every fixture with the ts-go parser, which enforces strict /
 * module ES semantics for `.ts` files. The upstream ESLint tests instead parse
 * these in sloppy mode (`parserOptions.sourceType: 'script'`) or with the older
 * `assert` import-attribute syntax. Under ts-go the source itself is a SYNTAX
 * ERROR, so rslint emits a TypeScript diagnostic and produces ZERO `@stylistic/
 * quotes` diagnostics for that file (and, because the rslint CLI aborts JSONL for
 * the whole batch on a syntax error, such a fixture would zero out every other
 * case in the same run — which is exactly why they must live outside the green
 * set). The rule logic itself is not at fault; the input is unparseable.
 *
 * This is a real, documented compatibility gap, not a silenced failure. The
 * expected upstream behaviour is preserved below for the record.
 *
 * ---- valid (upstream expects 0 diagnostics) ----
 *
 *   // `backtick`, import with `assert` attributes
 *   { code: `import "a" assert { type: "json" }; import 'b' assert { type: 'json' };`, options: ['backtick'] }
 *   { code: `import a from "a" assert { type: "json" }; import b from 'b' assert { type: 'json' };`, options: ['backtick'] }
 *
 *   rslint: TypeScript(TS2880) "Import assertions have been replaced by import
 *   attributes. Use 'with' instead of 'assert'." → 0 quotes diagnostics.
 *
 * ---- invalid (upstream expects 1 `wrongQuotes` diagnostic + the given fix) ----
 *
 *   // octal escapes — upstream runs all of these with parserOptions.sourceType: 'script'
 *   { code: `var foo = "\1"`,  output: `var foo = '\1'`, options: ['single'], errors: [{ messageId: 'wrongQuotes', data: { description: 'singlequote' } }] }
 *   { code: `var foo = '\1'`,  output: `var foo = "\1"`, options: ['double'], errors: [{ messageId: 'wrongQuotes', data: { description: 'doublequote' } }] }
 *   { code: `var foo = '\1'`,  output: null,             options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = "\1"`,  output: null,             options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = '\01'`, output: null,             options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = '\0\1'`, output: null,            options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = '\08'`, output: null,             options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = 'prefix \33'`,        output: null, options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var foo = 'prefix \75 suffix'`, output: null, options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *   { code: `var nonOctalDecimalEscape = '\8'`, output: null, options: ['backtick'], errors: [{ messageId: 'wrongQuotes', data: { description: 'backtick' } }] }
 *
 *   rslint: TypeScript(TS1487) "Octal escape sequences are not allowed." for the
 *   `\1`/`\01`/`\0\1`/`\08`/`\33`/`\75` forms, and TypeScript(TS1488) "Escape
 *   sequence '\8' is not allowed." for the `\8` form → 0 quotes diagnostics each.
 *
 * ============================================================================
 */
