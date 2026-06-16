/**
 * @fileoverview Tests for object-curly-spacing rule.
 * @author Jamund Ferguson
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/object-curly-spacing/object-curly-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/object-curly-spacing/object-curly-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('object-curly-spacing', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via tsconfig.
 *  - `type` fields (deprecated AST node type) dropped (none were present).
 *  - The few upstream `code` values written as backtick template literals are
 *    single-line strings (not the `$` unindent tag); they are kept as single-
 *    quoted strings here.
 *
 * The upstream files contain NO `$` unindent template tags, NO spread/helper
 * error builders, NO `readFileSync` external-fixture cases, and NO `suggestions`.
 * The `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * The two `if (!skipBabel)` `run()` blocks (`object-curly-spacing_babel`, using
 * `languageOptionsForBabelFlow`) are Babel/Flow parser suites — those cases are
 * moved to `KNOWN GAPS` at the bottom (rslint has no Babel/Flow parser).
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('object-curly-spacing', null as never, {
  valid: [
    // ---- from object-curly-spacing._js_.test.ts ----

    // always - object literals
    { code: 'var obj = { foo: bar, baz: qux };', options: ['always'] },
    { code: 'var obj = { foo: { bar: quxx }, baz: qux };', options: ['always'] },
    { code: 'var obj = {\nfoo: bar,\nbaz: qux\n};', options: ['always'] },
    { code: 'var obj = { /**/foo:bar/**/ };', options: ['always'] },
    { code: 'var obj = { //\nfoo:bar };', options: ['always'] },

    // always - destructuring
    { code: 'var { x } = y', options: ['always'] },
    { code: 'var { x, y } = y', options: ['always'] },
    { code: 'var { x,y } = y', options: ['always'] },
    { code: 'var {\nx,y } = y', options: ['always'] },
    { code: 'var {\nx,y\n} = z', options: ['always'] },
    { code: 'var { /**/x/**/ } = y', options: ['always'] },
    { code: 'var { //\nx } = y', options: ['always'] },
    { code: 'var { x = 10, y } = y', options: ['always'] },
    { code: 'var { x: { z }, y } = y', options: ['always'] },
    { code: 'var {\ny,\n} = x', options: ['always'] },
    { code: 'var { y, } = x', options: ['always'] },
    { code: 'var { y: x } = x', options: ['always'] },

    // always - import / export
    { code: 'import door from \'room\'', options: ['always'] },
    { code: 'import * as door from \'room\'', options: ['always'] },
    { code: 'import { door } from \'room\'', options: ['always'] },
    { code: 'import {\ndoor } from \'room\'', options: ['always'] },
    { code: 'import { /**/door/**/ } from \'room\'', options: ['always'] },
    { code: 'import { //\ndoor } from \'room\'', options: ['always'] },
    { code: 'export { door } from \'room\'', options: ['always'] },
    { code: 'import { house, mouse } from \'caravan\'', options: ['always'] },
    { code: 'import house, { mouse } from \'caravan\'', options: ['always'] },
    { code: 'import door, { house, mouse } from \'caravan\'', options: ['always'] },
    { code: 'var door = 0;export { door }', options: ['always'] },
    { code: 'import \'room\'', options: ['always'] },
    { code: 'import { bar as x } from \'foo\';', options: ['always'] },
    { code: 'import { x, } from \'foo\';', options: ['always'] },
    { code: 'import {\nx,\n} from \'foo\';', options: ['always'] },
    { code: 'export { x, } from \'foo\';', options: ['always'] },
    { code: 'export {\nx,\n} from \'foo\';', options: ['always'] },
    { code: 'export { /**/x/**/ } from \'foo\';', options: ['always'] },
    { code: 'export { //\nx } from \'foo\';', options: ['always'] },
    { code: 'var x = 1;\nexport { /**/x/**/ };', options: ['always'] },
    { code: 'var x = 1;\nexport { //\nx };', options: ['always'] },

    // always - empty object
    { code: 'var foo = {};', options: ['always'] },

    // always - objectsInObjects
    { code: 'var obj = { \'foo\': { \'bar\': 1, \'baz\': 2 }};', options: ['always', { objectsInObjects: false }] },
    { code: 'var a = { noop: function () {} };', options: ['always', { objectsInObjects: false }] },
    { code: 'var { y: { z }} = x', options: ['always', { objectsInObjects: false }] },

    // always - arraysInObjects
    { code: 'var obj = { \'foo\': [ 1, 2 ]};', options: ['always', { arraysInObjects: false }] },
    { code: 'var a = { thingInList: list[0] };', options: ['always', { arraysInObjects: false }] },

    // always - arraysInObjects, objectsInObjects
    { code: 'var obj = { \'qux\': [ 1, 2 ], \'foo\': { \'bar\': 1, \'baz\': 2 }};', options: ['always', { arraysInObjects: false, objectsInObjects: false }] },

    // always - arraysInObjects, objectsInObjects (reverse)
    { code: 'var obj = { \'foo\': { \'bar\': 1, \'baz\': 2 }, \'qux\': [ 1, 2 ]};', options: ['always', { arraysInObjects: false, objectsInObjects: false }] },

    // never
    { code: 'var obj = {foo: bar,\nbaz: qux\n};', options: ['never'] },
    { code: 'var obj = {\nfoo: bar,\nbaz: qux};', options: ['never'] },

    // never - object literals
    { code: 'var obj = {foo: bar, baz: qux};', options: ['never'] },
    { code: 'var obj = {foo: {bar: quxx}, baz: qux};', options: ['never'] },
    { code: 'var obj = {foo: {\nbar: quxx}, baz: qux\n};', options: ['never'] },
    { code: 'var obj = {foo: {\nbar: quxx\n}, baz: qux};', options: ['never'] },
    { code: 'var obj = {\nfoo: bar,\nbaz: qux\n};', options: ['never'] },
    { code: 'var obj = {foo: bar, baz: qux /* */};', options: ['never'] },
    { code: 'var obj = {/* */ foo: bar, baz: qux};', options: ['never'] },
    { code: 'var obj = {//\n foo: bar};', options: ['never'] },
    { code: 'var obj = { // line comment exception\n foo: bar};', options: ['never'] },

    // never - destructuring
    { code: 'var {x} = y', options: ['never'] },
    { code: 'var {x, y} = y', options: ['never'] },
    { code: 'var {x,y} = y', options: ['never'] },
    { code: 'var {\nx,y\n} = y', options: ['never'] },
    { code: 'var {x = 10} = y', options: ['never'] },
    { code: 'var {x = 10, y} = y', options: ['never'] },
    { code: 'var {x: {z}, y} = y', options: ['never'] },
    { code: 'var {\nx: {z\n}, y} = y', options: ['never'] },
    { code: 'var {\ny,\n} = x', options: ['never'] },
    { code: 'var {y,} = x', options: ['never'] },
    { code: 'var {y:x} = x', options: ['never'] },
    { code: 'var {/* */ y} = x', options: ['never'] },
    { code: 'var {y /* */} = x', options: ['never'] },
    { code: 'var {//\n y} = x', options: ['never'] },
    { code: 'var { // line comment exception\n y} = x', options: ['never'] },

    // never - import / export
    { code: 'import door from \'room\'', options: ['never'] },
    { code: 'import * as door from \'room\'', options: ['never'] },
    { code: 'import {door} from \'room\'', options: ['never'] },
    { code: 'export {door} from \'room\'', options: ['never'] },
    { code: 'import {/* */ door} from \'room\'', options: ['never'] },
    { code: 'export {/* */ door} from \'room\'', options: ['never'] },
    { code: 'import {door /* */} from \'room\'', options: ['never'] },
    { code: 'export {door /* */} from \'room\'', options: ['never'] },
    { code: 'import {//\n door} from \'room\'', options: ['never'] },
    { code: 'export {//\n door} from \'room\'', options: ['never'] },
    { code: 'var door = foo;\nexport {//\n door}', options: ['never'] },
    { code: 'import { // line comment exception\n door} from \'room\'', options: ['never'] },
    { code: 'export { // line comment exception\n door} from \'room\'', options: ['never'] },
    { code: 'var door = foo; export { // line comment exception\n door}', options: ['never'] },
    { code: 'import {\ndoor} from \'room\'', options: ['never'] },
    { code: 'export {\ndoor\n} from \'room\'', options: ['never'] },
    { code: 'import {house,mouse} from \'caravan\'', options: ['never'] },
    { code: 'import {house, mouse} from \'caravan\'', options: ['never'] },
    { code: 'var door = 0;export {door}', options: ['never'] },
    { code: 'import \'room\'', options: ['never'] },
    { code: 'import x, {bar} from \'foo\';', options: ['never'] },
    { code: 'import x, {bar, baz} from \'foo\';', options: ['never'] },
    { code: 'import {bar as y} from \'foo\';', options: ['never'] },
    { code: 'import {x,} from \'foo\';', options: ['never'] },
    { code: 'import {\nx,\n} from \'foo\';', options: ['never'] },
    { code: 'export {x,} from \'foo\';', options: ['never'] },
    { code: 'export {\nx,\n} from \'foo\';', options: ['never'] },

    // never - empty object
    { code: 'var foo = {};', options: ['never'] },

    // never - objectsInObjects
    { code: 'var obj = {\'foo\': {\'bar\': 1, \'baz\': 2} };', options: ['never', { objectsInObjects: true }] },

    /**
     * https://github.com/eslint/eslint/issues/3658
     * Empty cases.
     */
    { code: 'var {} = foo;' },
    { code: 'var [] = foo;' },
    { code: 'var {a: {}} = foo;' },
    { code: 'var {a: []} = foo;' },
    { code: 'import {} from \'foo\';' },
    { code: 'export {} from \'foo\';' },
    { code: 'export {};' },
    { code: 'var {} = foo;', options: ['never'] },
    { code: 'var [] = foo;', options: ['never'] },
    { code: 'var {a: {}} = foo;', options: ['never'] },
    { code: 'var {a: []} = foo;', options: ['never'] },
    { code: 'import {} from \'foo\';', options: ['never'] },
    { code: 'export {} from \'foo\';', options: ['never'] },
    { code: 'export {};', options: ['never'] },

    // https://github.com/eslint-stylistic/eslint-stylistic/issues/906
    'import foo, * as bar from \'mod\'',
    {
      code: 'var obj = {   /*comment*/   };',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'var obj = {/*comment*/};',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'var obj = { /* comment */ \nfoo: bar\n};',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'var obj = {};',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'var obj = { };',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'var {} = y;',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'var { } = y;',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'import {} from "room";',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'import { } from "room";',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'export {}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'export { }',
      options: ['never', { emptyObjects: 'always' }],
    },

    // ---- from object-curly-spacing._ts_.test.ts ----

    // default - object literal types
    {
      code: 'const x:{}',
    },
    {
      code: 'const x:{ }',
    },
    {
      code: 'const x:{f: number}',
    },
    {
      code: 'const x:{ // line-comment\nf: number\n}',
    },
    {
      code: 'const x:{// line-comment\nf: number\n}',
    },
    {
      code: 'const x:{/* inline-comment */f: number/* inline-comment */}',
    },
    {
      code: 'const x:{\nf: number\n}',
    },
    {
      code: 'const x:{f: {g: number}}',
    },
    {
      code: 'const x:{f: [number]}',
    },
    {
      code: 'const x:{[key: string]: value}',
    },
    {
      code: 'const x:{[key: string]: [number]}',
    },

    // default - mapped types
    {
      code: 'const x:{[k in \'union\']: number}',
    },
    {
      code: 'const x:{ // line-comment\n[k in \'union\']: number\n}',
    },
    {
      code: 'const x:{// line-comment\n[k in \'union\']: number\n}',
    },
    {
      code: 'const x:{/* inline-comment */[k in \'union\']: number/* inline-comment */}',
    },
    {
      code: 'const x:{\n[k in \'union\']: number\n}',
    },
    {
      code: 'const x:{[k in \'union\']: {[k in \'union\']: number}}',
    },
    {
      code: 'const x:{[k in \'union\']: [number]}',
    },
    {
      code: 'const x:{[k in \'union\']: value}',
    },

    // never - mapped types
    {
      code: 'const x:{[k in \'union\']: {[k in \'union\']: number} }',
      options: ['never', { objectsInObjects: true }],
    },
    {
      code: 'const x:{[k in \'union\']: {[k in \'union\']: number}}',
      options: ['never', { objectsInObjects: false }],
    },
    {
      code: 'const x:{[k in \'union\']: () => {[k in \'union\']: number} }',
      options: ['never', { objectsInObjects: true }],
    },
    {
      code: 'const x:{[k in \'union\']: () => {[k in \'union\']: number}}',
      options: ['never', { objectsInObjects: false }],
    },
    {
      code: 'const x:{[k in \'union\']: [ number ]}',
      options: ['never', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [k in \'union\']: value}',
      options: ['never', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[k in \'union\']: value}',
      options: ['never', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [k in \'union\']: [number] }',
      options: ['never', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[k in \'union\']: [number]}',
      options: ['never', { arraysInObjects: false }],
    },

    // never - object literal types
    {
      code: 'const x:{f: {g: number} }',
      options: ['never', { objectsInObjects: true }],
    },
    {
      code: 'const x:{f: {g: number}}',
      options: ['never', { objectsInObjects: false }],
    },
    {
      code: 'const x:{f: () => {g: number} }',
      options: ['never', { objectsInObjects: true }],
    },
    {
      code: 'const x:{f: () => {g: number}}',
      options: ['never', { objectsInObjects: false }],
    },
    {
      code: 'const x:{f: [number] }',
      options: ['never', { arraysInObjects: true }],
    },
    {
      code: 'const x:{f: [ number ]}',
      options: ['never', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [key: string]: value}',
      options: ['never', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[key: string]: value}',
      options: ['never', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [key: string]: [number] }',
      options: ['never', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[key: string]: [number]}',
      options: ['never', { arraysInObjects: false }],
    },

    // always - mapped types
    {
      code: 'const x:{ [k in \'union\']: number }',
      options: ['always'],
    },
    {
      code: 'const x:{ // line-comment\n[k in \'union\']: number\n}',
      options: ['always'],
    },
    {
      code: 'const x:{ /* inline-comment */ [k in \'union\']: number /* inline-comment */ }',
      options: ['always'],
    },
    {
      code: 'const x:{\n[k in \'union\']: number\n}',
      options: ['always'],
    },
    {
      code: 'const x:{ [k in \'union\']: [number] }',
      options: ['always'],
    },

    // always - mapped types - objectsInObjects
    {
      code: 'const x:{ [k in \'union\']: { [k in \'union\']: number } }',
      options: ['always', { objectsInObjects: true }],
    },
    {
      code: 'const x:{ [k in \'union\']: { [k in \'union\']: number }}',
      options: ['always', { objectsInObjects: false }],
    },
    {
      code: 'const x:{ [k in \'union\']: () => { [k in \'union\']: number } }',
      options: ['always', { objectsInObjects: true }],
    },
    {
      code: 'const x:{ [k in \'union\']: () => { [k in \'union\']: number }}',
      options: ['always', { objectsInObjects: false }],
    },

    // always - mapped types - arraysInObjects
    {
      code: 'type x = { [k in \'union\']: number }',
      options: ['always'],
    },
    {
      code: 'const x:{ [k in \'union\']: [number] }',
      options: ['always', { arraysInObjects: true }],
    },
    {
      code: 'const x:{ [k in \'union\']: value }',
      options: ['always', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[k in \'union\']: value }',
      options: ['always', { arraysInObjects: false }],
    },
    {
      code: 'const x:{[k in \'union\']: [number]}',
      options: ['always', { arraysInObjects: false }],
    },

    // always - object literal types
    {
      code: 'const x:{}',
      options: ['always'],
    },
    {
      code: 'const x:{ }',
      options: ['always'],
    },
    {
      code: 'const x:{ f: number }',
      options: ['always'],
    },
    {
      code: 'const x:{ // line-comment\nf: number\n}',
      options: ['always'],
    },
    {
      code: 'const x:{ /* inline-comment */ f: number /* inline-comment */ }',
      options: ['always'],
    },
    {
      code: 'const x:{\nf: number\n}',
      options: ['always'],
    },
    {
      code: 'const x:{ f: [number] }',
      options: ['always'],
    },

    // always - literal types - objectsInObjects
    {
      code: 'const x:{ f: { g: number } }',
      options: ['always', { objectsInObjects: true }],
    },
    {
      code: 'const x:{ f: { g: number }}',
      options: ['always', { objectsInObjects: false }],
    },
    {
      code: 'const x:{ f: () => { g: number } }',
      options: ['always', { objectsInObjects: true }],
    },
    {
      code: 'const x:{ f: () => { g: number }}',
      options: ['always', { objectsInObjects: false }],
    },

    // always - literal types - arraysInObjects
    {
      code: 'const x:{ f: [number] }',
      options: ['always', { arraysInObjects: true }],
    },
    {
      code: 'const x:{ f: [ number ]}',
      options: ['always', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [key: string]: value }',
      options: ['always', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[key: string]: value }',
      options: ['always', { arraysInObjects: false }],
    },
    {
      code: 'const x:{ [key: string]: [number] }',
      options: ['always', { arraysInObjects: true }],
    },
    {
      code: 'const x:{[key: string]: [number]}',
      options: ['always', { arraysInObjects: false }],
    },

    // default - TSInterfaceBody
    {
      code: 'interface x {f: number}',
    },
    // always - TSInterfaceBody
    {
      code: 'interface x { f: number }',
      options: ['always'],
    },
    // never - TSInterfaceBody
    {
      code: 'interface x {f: number}',
      options: ['never'],
    },
    // default - TSEnumBody
    {
      code: 'enum Foo {ONE, TWO,}',
    },
    // always - TSEnumBody
    {
      code: 'enum Foo { ONE, TWO = 2 }',
      options: ['always'],
    },
    // never - TSEnumBody
    {
      code: 'enum Foo {ONE, TWO,}',
      options: ['never'],
    },
    {
      code: 'import pkgJson from \'package.json\' with {type: \'json\'}',
    },
    {
      code: 'export { name } from \'package.json\' with { type: \'json\' }',
      options: ['always'],
    },
    {
      code: 'export * from \'package.json\' with {type: \'json\'}',
      options: ['never'],
    },
    {
      code: 'import {name, version} from \'package.json\' with { type: \'json\' }',
      options: ['never', {
        overrides: {
          ImportAttributes: 'always',
        },
      }],
    },

    {
      code: 'const x:{/* comment */}',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'const x:{  /* comment */  }',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'const x:{}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'const x:{ }',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'const x:{f: {}}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'const x:{f: { }}',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'interface x {}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'interface x { }',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'interface x { /* comment */ \n foo: string \n}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'import {} from "package.json" with {}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'import { } from "package.json" with { }',
      options: ['never', { emptyObjects: 'always' }],
    },
    {
      code: 'enum Foo {}',
      options: ['never', { emptyObjects: 'never' }],
    },
    {
      code: 'enum Foo { }',
      options: ['never', { emptyObjects: 'always' }],
    },
  ],

  invalid: [
    // ---- from object-curly-spacing._js_.test.ts ----
    {
      code: 'import {bar} from \'foo.js\';',
      output: 'import { bar } from \'foo.js\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'import { bar as y} from \'foo.js\';',
      output: 'import { bar as y } from \'foo.js\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'import {bar as y} from \'foo.js\';',
      output: 'import { bar as y } from \'foo.js\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'import { bar} from \'foo.js\';',
      output: 'import { bar } from \'foo.js\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'import x, { bar} from \'foo\';',
      output: 'import x, { bar } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },

      ],
    },
    {
      code: 'import x, { bar/* */} from \'foo\';',
      output: 'import x, { bar/* */ } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'import x, {/* */bar } from \'foo\';',
      output: 'import x, { /* */bar } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'import x, {//\n bar } from \'foo\';',
      output: 'import x, { //\n bar } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'import x, { bar, baz} from \'foo\';',
      output: 'import x, { bar, baz } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },

      ],
    },
    {
      code: 'import x, {bar} from \'foo\';',
      output: 'import x, { bar } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },

      ],
    },
    {
      code: 'import x, {bar, baz} from \'foo\';',
      output: 'import x, { bar, baz } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: 'import {bar,} from \'foo\';',
      output: 'import { bar, } from \'foo\';',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },

      ],
    },
    {
      code: 'import { bar, } from \'foo\';',
      output: 'import {bar,} from \'foo\';',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'import { /* */ bar, /* */ } from \'foo\';',
      output: 'import {/* */ bar, /* */} from \'foo\';',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: 'var bar = 0;\nexport {bar};',
      output: 'var bar = 0;\nexport { bar };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 2,
          column: 12,
        },
      ],
    },
    {
      code: 'var bar = 0;\nexport {/* */ bar /* */};',
      output: 'var bar = 0;\nexport { /* */ bar /* */ };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'var bar = 0;\nexport {//\n bar };',
      output: 'var bar = 0;\nexport { //\n bar };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'var bar = 0;\nexport { /* */ bar /* */ };',
      output: 'var bar = 0;\nexport {/* */ bar /* */};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 2,
          column: 9,
          endLine: 2,
          endColumn: 10,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 26,
        },
      ],
    },

    // always - arraysInObjects
    {
      code: 'var obj = { \'foo\': [ 1, 2 ] };',
      output: 'var obj = { \'foo\': [ 1, 2 ]};',
      options: ['always', { arraysInObjects: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'var obj = { \'foo\': [ 1, 2 ] , \'bar\': [ \'baz\', \'qux\' ] };',
      output: 'var obj = { \'foo\': [ 1, 2 ] , \'bar\': [ \'baz\', \'qux\' ]};',
      options: ['always', { arraysInObjects: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // always-objectsInObjects
    {
      code: 'var obj = { \'foo\': { \'bar\': 1, \'baz\': 2 } };',
      output: 'var obj = { \'foo\': { \'bar\': 1, \'baz\': 2 }};',
      options: ['always', { objectsInObjects: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'var obj = { \'foo\': [ 1, 2 ] , \'bar\': { \'baz\': 1, \'qux\': 2 } };',
      output: 'var obj = { \'foo\': [ 1, 2 ] , \'bar\': { \'baz\': 1, \'qux\': 2 }};',
      options: ['always', { objectsInObjects: false }],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 60,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },

    // always-destructuring trailing comma
    {
      code: 'var { a,} = x;',
      output: 'var { a, } = x;',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'var {a, } = x;',
      output: 'var {a,} = x;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'var {a:b } = x;',
      output: 'var {a:b} = x;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'var { a:b } = x;',
      output: 'var {a:b} = x;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 7,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var {  a:b  } = x;',
      output: 'var {a:b} = x;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 8,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var {   a:b    } = x;',
      output: 'var {a:b} = x;',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 9,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },

    // never-objectsInObjects
    {
      code: 'var obj = {\'foo\': {\'bar\': 1, \'baz\': 2}};',
      output: 'var obj = {\'foo\': {\'bar\': 1, \'baz\': 2} };',
      options: ['never', { objectsInObjects: true }],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'var obj = {\'foo\': [1, 2] , \'bar\': {\'baz\': 1, \'qux\': 2}};',
      output: 'var obj = {\'foo\': [1, 2] , \'bar\': {\'baz\': 1, \'qux\': 2} };',
      options: ['never', { objectsInObjects: true }],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 55,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },

    // always & never
    {
      code: 'var obj = {foo: bar, baz: qux};',
      output: 'var obj = { foo: bar, baz: qux };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'var obj = {foo: bar, baz: qux };',
      output: 'var obj = { foo: bar, baz: qux };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var obj = {/* */foo: bar, baz: qux };',
      output: 'var obj = { /* */foo: bar, baz: qux };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var obj = {//\n foo: bar };',
      output: 'var obj = { //\n foo: bar };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var obj = { foo: bar, baz: qux};',
      output: 'var obj = { foo: bar, baz: qux };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = { foo: bar, baz: qux/* */};',
      output: 'var obj = { foo: bar, baz: qux/* */ };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'var obj = { foo: bar, baz: qux };',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {  foo: bar, baz: qux };',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'var obj = {foo: bar, baz: qux };',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: 'var obj = {foo: bar, baz: qux  };',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'var obj = {foo: bar, baz: qux /* */ };',
      output: 'var obj = {foo: bar, baz: qux /* */};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'var obj = { foo: bar, baz: qux};',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var obj = {  foo: bar, baz: qux};',
      output: 'var obj = {foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'var obj = { /* */ foo: bar, baz: qux};',
      output: 'var obj = {/* */ foo: bar, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var obj = { // line comment exception\n foo: bar };',
      output: 'var obj = { // line comment exception\n foo: bar};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 2,
          column: 10,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var obj = { foo: { bar: quxx}, baz: qux};',
      output: 'var obj = {foo: {bar: quxx}, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: 'var obj = {foo: {bar: quxx }, baz: qux };',
      output: 'var obj = {foo: {bar: quxx}, baz: qux};',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'export const thing = {value: 1 };',
      output: 'export const thing = { value: 1 };',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // destructuring
    {
      code: 'var {x, y} = y',
      output: 'var { x, y } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var { x, y} = y',
      output: 'var { x, y } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var { x, y/* */} = y',
      output: 'var { x, y/* */ } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'var {/* */x, y } = y',
      output: 'var { /* */x, y } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'var {//\n x } = y',
      output: 'var { //\n x } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'var { x, y } = y',
      output: 'var {x, y} = y',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 7,
        },
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var {x, y } = y',
      output: 'var {x, y} = y',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var {x, y/* */ } = y',
      output: 'var {x, y/* */} = y',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'var { /* */x, y} = y',
      output: 'var {/* */x, y} = y',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'var { x=10} = y',
      output: 'var { x=10 } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var {x=10 } = y',
      output: 'var { x=10 } = y',
      options: ['always'],
      errors: [
        {
          messageId: 'requireSpaceAfter',
          data: { token: '{' },
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },

    // never - arraysInObjects
    {
      code: 'var obj = {\'foo\': [1, 2]};',
      output: 'var obj = {\'foo\': [1, 2] };',
      options: ['never', { arraysInObjects: true }],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'var obj = {\'foo\': [1, 2] , \'bar\': [\'baz\', \'qux\']};',
      output: 'var obj = {\'foo\': [1, 2] , \'bar\': [\'baz\', \'qux\'] };',
      options: ['never', { arraysInObjects: true }],
      errors: [
        {
          messageId: 'requireSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    {
      code: 'var obj = {};',
      output: 'var obj = { };',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ObjectExpression' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'var {} = y;',
      output: 'var { } = y;',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ObjectPattern' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'import {} from "room";',
      output: 'import { } from "room";',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ImportDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'export {}',
      output: 'export { }',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ExportNamedDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'var obj = { };',
      output: 'var obj = {};',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ObjectExpression' },
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'var { } = y;',
      output: 'var {} = y;',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ObjectPattern' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: 'import { } from "room";',
      output: 'import {} from "room";',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ImportDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'export {      }',
      output: 'export {}',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ExportNamedDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'var {      } = y;',
      output: 'var { } = y;',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ObjectPattern' },
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'import {    } from "room";',
      output: 'import { } from "room";',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ImportDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'export {      }',
      output: 'export { }',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ExportNamedDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },

    // ---- from object-curly-spacing._ts_.test.ts ----
    // https://github.com/eslint/eslint/issues/6940
    {
      code: 'function foo ({a, b }: Props) {\n}',
      output: 'function foo ({a, b}: Props) {\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'unexpectedSpaceBefore',
          data: { token: '}' },
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // object literal types
    // never - literal types
    {
      code: 'type x = { f: number }',
      output: 'type x = {f: number}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'type x = { f: number}',
      output: 'type x = {f: number}',
      errors: [{ messageId: 'unexpectedSpaceAfter' }],
    },
    {
      code: 'type x = {f: number }',
      output: 'type x = {f: number}',
      errors: [{ messageId: 'unexpectedSpaceBefore' }],
    },
    // always - literal types
    {
      code: 'type x = {f: number}',
      output: 'type x = { f: number }',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
    {
      code: 'type x = {f: number }',
      output: 'type x = { f: number }',
      options: ['always'],
      errors: [{ messageId: 'requireSpaceAfter' }],
    },
    {
      code: 'type x = { f: number}',
      output: 'type x = { f: number }',
      options: ['always'],
      errors: [{ messageId: 'requireSpaceBefore' }],
    },

    // never - mapped types
    {
      code: 'type x = { [k in \'union\']: number }',
      output: 'type x = {[k in \'union\']: number}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'type x = { [k in \'union\']: number}',
      output: 'type x = {[k in \'union\']: number}',
      errors: [{ messageId: 'unexpectedSpaceAfter' }],
    },
    {
      code: 'type x = {[k in \'union\']: number }',
      output: 'type x = {[k in \'union\']: number}',
      errors: [{ messageId: 'unexpectedSpaceBefore' }],
    },
    // always - mapped types
    {
      code: 'type x = {[k in \'union\']: number}',
      output: 'type x = { [k in \'union\']: number }',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
    {
      code: 'type x = {[k in \'union\']: number }',
      output: 'type x = { [k in \'union\']: number }',
      options: ['always'],
      errors: [{ messageId: 'requireSpaceAfter' }],
    },
    {
      code: 'type x = { [k in \'union\']: number}',
      output: 'type x = { [k in \'union\']: number }',
      options: ['always'],
      errors: [{ messageId: 'requireSpaceBefore' }],
    },
    // Mapped and literal types mix
    {
      code: 'type x = { [k in \'union\']: { [k: string]: number } }',
      output: 'type x = {[k in \'union\']: {[k: string]: number}}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    // TSInterfaceBody
    {
      code: 'interface x { f: number }',
      output: 'interface x {f: number}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'interface x {f: number}',
      output: 'interface x { f: number }',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
    // TSEnumBody
    {
      code: 'enum Foo { ONE, TWO = 2 }',
      output: 'enum Foo {ONE, TWO = 2}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'enum Foo {ONE, TWO,}',
      output: 'enum Foo { ONE, TWO, }',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
    {
      code: 'import pkgJson from \'package.json\' with { type: \'json\' }',
      output: 'import pkgJson from \'package.json\' with {type: \'json\'}',
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'export { name } from \'package.json\' with {type: \'json\'}',
      output: 'export { name } from \'package.json\' with { type: \'json\' }',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
    {
      code: 'export * from \'package.json\' with { type: \'json\' }',
      output: 'export * from \'package.json\' with {type: \'json\'}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'import {name, version} from \'package.json\' with { type: \'json\' }',
      output: 'import { name, version } from \'package.json\' with {type: \'json\'}',
      options: ['always', {
        overrides: {
          ImportAttributes: 'never',
        },
      }],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
        { messageId: 'unexpectedSpaceAfter' },
        { messageId: 'unexpectedSpaceBefore' },
      ],
    },
    {
      code: 'import { } from "package.json" with { }',
      output: 'import {} from "package.json" with {}',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ImportDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ImportAttributes' },
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'enum Foo {}',
      output: 'enum Foo { }',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'TSEnumBody' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'const x:{f: {}}',
      output: 'const x:{f: { }}',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'TSTypeLiteral' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'interface x {    }',
      output: 'interface x {}',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'TSInterfaceBody' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'export {   } from "package.json" with {   }',
      output: 'export {} from "package.json" with {}',
      options: ['never', { emptyObjects: 'never' }],
      errors: [
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ExportNamedDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'unexpectedSpaceInEmptyObject',
          data: { node: 'ImportAttributes' },
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'import {   } from "package.json" with {   }',
      output: 'import { } from "package.json" with { }',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ImportDeclaration' },
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'ImportAttributes' },
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'enum Foo {   }',
      output: 'enum Foo { }',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'TSEnumBody' },
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: 'const x:{f: {   }}',
      output: 'const x:{f: { }}',
      options: ['never', { emptyObjects: 'always' }],
      errors: [
        {
          messageId: 'requiredSpaceInEmptyObject',
          data: { node: 'TSTypeLiteral' },
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'const foo = ({str}: { str: string }) => null',
      output: 'const foo = ({ str }: { str: string }) => null',
      options: ['always'],
      errors: [
        { messageId: 'requireSpaceAfter' },
        { messageId: 'requireSpaceBefore' },
      ],
    },
  ],
});

/**
 * ===================== object-curly-spacing — KNOWN GAPS =====================
 *
 * The cases below come from the two upstream `if (!skipBabel)` `run()` blocks
 * (`object-curly-spacing_babel`), which parse with `languageOptionsForBabelFlow`
 * — i.e. the Babel parser with the Flow plugin. rslint has no Babel/Flow parser;
 * it parses every fixture with the ts-go parser. Both fixtures use Flow's
 * `({ ... }: Props)` destructuring-parameter typing as exercised only under
 * Babel/Flow upstream. They are preserved here verbatim for the record, NOT run
 * through the green `ruleTester.run` above.
 *
 * NOTE: the identical-looking `function foo ({a, b }: Props) {...}` invalid case
 * from the `._ts_` suite (with NO `languageOptions`) IS included in the green
 * `invalid` set above — that one is a genuine TypeScript fixture; only the
 * Babel/Flow-parsed copies below are gapped.
 *
 * ---- from object-curly-spacing._js_.test.ts (Babel/Flow) ----
 *
 *   // valid (upstream expects 0 diagnostics)
 *   { code: 'function foo ({a, b}: Props) {\n}', options: ['never'] }   // Babel/Flow
 *
 *   // invalid (upstream expects the given diagnostic + fix)
 *   {
 *     code:   'function foo ({a, b }: Props) {\n}',
 *     output: 'function foo ({a, b}: Props) {\n}',
 *     options: ['never'],   // Babel/Flow
 *     errors: [
 *       { messageId: 'unexpectedSpaceBefore', data: { token: '}' }, line: 1, column: 20, endLine: 1, endColumn: 21 },
 *     ],
 *   }
 *
 * =============================================================================
 */
