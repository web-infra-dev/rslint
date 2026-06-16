/**
 * @fileoverview Tests for object-curly-newline rule.
 * @author Toru Nagashima
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/object-curly-newline/object-curly-newline._js_.test.ts
 *   packages/eslint-plugin/rules/object-curly-newline/object-curly-newline._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('object-curly-newline', null as never, { valid, invalid })`
 *  - The `[...].join('\n')` row builders and the `$` unindent template tag are
 *    evaluated to their real strings.
 *  - The `_ts_` file's `createValidRule` / `createInvalidRule` helpers are
 *    evaluated to their final, fully-expanded cases (each input is emitted twice,
 *    once prefixed `type Foo = ` for a TSTypeLiteral and once `interface Foo `
 *    for a TSInterfaceBody, with a trailing `// <options>` comment and the
 *    line-1 column shifted by the prefix length, exactly as the helper does).
 *  - `parserOptions` (ecmaVersion) and the type generics dropped — rslint
 *    resolves via tsconfig.
 *
 * Merged from both run() blocks: the `_js_` main block, the `_ts_` block, and
 * the `_js_` `skipBabel` block (`object-curly-newline_babel`, Babel/Flow). The
 * `._css_` / `._json_` / `._markdown_` files don't exist for this rule.
 *
 * skipBabel (Babel/Flow) handling — see KNOWN GAPS at the bottom: 6 of the 8
 * Flow cases are valid TS whose behaviour matches rslint exactly and are kept in
 * the green set below; the 2 that diverge (an inline object TYPE `{ a : string,
 * b : string }` which ts-go parses as a TSTypeLiteral the rule then lints, where
 * Babel's Flow parser does not) are isolated into KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('object-curly-newline', null as never, {
  valid: [
    // ==================== from object-curly-newline._js_.test.ts ====================
    'var a = {\n};',
    'var a = {\n   foo\n};',
    'var a = { foo }',
    {
      code: 'var a = {\n};',
      options: ['always'],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      options: ['always'],
    },
    {
      code: 'var c = {\n    a: 1, b: 2\n};',
      options: ['always'],
    },
    {
      code: 'var d = {\n    a: 1,\n    b: 2\n};',
      options: ['always'],
    },
    {
      code: 'var e = {\n    a: function foo() {\n        dosomething();\n    }\n};',
      options: ['always'],
    },
    {
      code: 'var a = {};',
      options: ['never'],
    },
    {
      code: 'var b = {a: 1};',
      options: ['never'],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      options: ['never'],
    },
    {
      code: 'var d = {a: 1,\n    b: 2};',
      options: ['never'],
    },
    {
      code: 'var e = {a: function foo() {\n    dosomething();\n}};',
      options: ['never'],
    },
    {
      code: 'var a = {};',
      options: [{ multiline: true }],
    },
    {
      code: 'var b = {a: 1};',
      options: [{ multiline: true }],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      options: [{ multiline: true }],
    },
    {
      code: 'var d = {\n    a: 1,\n    b: 2\n};',
      options: [{ multiline: true }],
    },
    {
      code: 'var e = {\n    a: function foo() {\n        dosomething();\n    }\n};',
      options: [{ multiline: true }],
    },
    {
      code: 'var obj = {\n    // comment\n    a: 1\n};',
      options: [{ multiline: true }],
    },
    {
      code: 'var obj = { // comment\n    a: 1\n};',
      options: [{ multiline: true }],
    },
    {
      code: 'var a = {};',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'var b = {a: 1};',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'var c = {\n    a: 1, b: 2\n};',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'var d = {\n    a: 1,\n    b: 2\n};',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'var e = {a: function foo() {\n    dosomething();\n}};',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'var a = {};',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'var b = {a: 1};',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'var c = {\n    a: 1, b: 2\n};',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'var d = {\n    a: 1, \n    b: 2\n};',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'var e = {\n    a: function foo() {\n        dosomething();\n    }\n};',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'var c = {\n    a: 1,\n    b: 2\n};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'var e = {a: function() { dosomething();}};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'var e = {\n    a: function() { dosomething();}\n};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {} = {a: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {a} = {a: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {\n} = {a: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {\n    a\n} = {a: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {a, b} = {a: 1, b: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {\n    a, b\n} = {a: 1, b: 1};',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {k = function() {dosomething();}} = obj;',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'let {\n    k = function() {\n        dosomething();\n    }\n} = obj;',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'var c = {a: 1,\nb: 2};',
      options: [{ multiline: false, consistent: true }],
    },
    {
      code: 'let {a,\nb} = {a: 1, b: 1};',
      options: [{ multiline: false, consistent: true }],
    },
    {
      code: 'var c = { a: 1 };',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
    },
    {
      code: 'var c = {\na: 1\n};',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
    },
    {
      code: 'let {a} = {\na: 1\n};',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
    },
    {
      code: 'let {\na\n} = {\na: 1\n};',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
    },
    {
      code: 'let {a, b} = {\n    a: 1, b: 2\n};',
      options: [{ ObjectExpression: 'always', ObjectPattern: 'never' }],
    },
    {
      code: 'import {\n    a,\n b\n} from \'module\';',
      options: [{ ImportDeclaration: 'always' }],
    },
    {
      code: 'import {a as a, b} from \'module\';',
      options: [{ ImportDeclaration: 'never' }],
    },
    {
      code: 'import { a, } from \'module\';',
      options: [{ ImportDeclaration: { multiline: true } }],
    },
    {
      code: 'import {\na, \nb\n} from \'module\';',
      options: [{ ImportDeclaration: { multiline: true } }],
    },
    {
      code: 'import {\n a,\n} from \'module\';',
      options: [{ ImportDeclaration: { consistent: true } }],
    },
    {
      code: 'import { a } from \'module\';',
      options: [{ ImportDeclaration: { consistent: true } }],
    },
    {
      code: 'import {\na, b\n} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 2 } }],
    },
    {
      code: 'import {a, b} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 3 } }],
    },
    {
      code: 'import DefaultExport, {a} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 2 } }],
    },
    {
      code: 'var a = 0, b = 0;\nexport {a,\nb};',
      options: [{ ExportDeclaration: 'never' }],
    },
    {
      code: 'var a = 0, b = 0;\nexport {\na as a, b\n} from \'module\';',
      options: [{ ExportDeclaration: 'always' }],
    },
    {
      code: 'export { a } from \'module\';',
      options: [{ ExportDeclaration: { multiline: true } }],
    },
    {
      code: 'export {\na, \nb\n} from \'module\';',
      options: [{ ExportDeclaration: { multiline: true } }],
    },
    {
      code: 'export {a, \nb} from \'module\';',
      options: [{ ExportDeclaration: { consistent: true } }],
    },
    {
      code: 'export {\na, b\n} from \'module\';',
      options: [{ ExportDeclaration: { minProperties: 2 } }],
    },
    {
      code: 'export {a, b} from \'module\';',
      options: [{ ExportDeclaration: { minProperties: 3 } }],
    },

    // ==================== from object-curly-newline._ts_.test.ts ====================
    {
      code: 'type Foo = {\n};// "always"',
      options: ['always'],
    },
    {
      code: 'type Foo = {\n};// "always"',
      options: [{ TSTypeLiteral: 'always' }],
    },
    {
      code: 'interface Foo {\n};// "always"',
      options: ['always'],
    },
    {
      code: 'interface Foo {\n};// "always"',
      options: [{ TSInterfaceBody: 'always' }],
    },
    {
      code: 'type Foo = {\n    a: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'type Foo = {\n    a: number;\n};// "always"',
      options: [{ TSTypeLiteral: 'always' }],
    },
    {
      code: 'interface Foo {\n    a: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'interface Foo {\n    a: number;\n};// "always"',
      options: [{ TSInterfaceBody: 'always' }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n};// "always"',
      options: [{ TSTypeLiteral: 'always' }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n};// "always"',
      options: [{ TSInterfaceBody: 'always' }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n};// "always"',
      options: [{ TSTypeLiteral: 'always' }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n};// "always"',
      options: ['always'],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n};// "always"',
      options: [{ TSInterfaceBody: 'always' }],
    },
    {
      code: 'type Foo = {}// "never"',
      options: ['never'],
    },
    {
      code: 'type Foo = {}// "never"',
      options: [{ TSTypeLiteral: 'never' }],
    },
    {
      code: 'interface Foo {}// "never"',
      options: ['never'],
    },
    {
      code: 'interface Foo {}// "never"',
      options: [{ TSInterfaceBody: 'never' }],
    },
    {
      code: 'type Foo = { a: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'type Foo = { a: number; }// "never"',
      options: [{ TSTypeLiteral: 'never' }],
    },
    {
      code: 'interface Foo { a: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'interface Foo { a: number; }// "never"',
      options: [{ TSInterfaceBody: 'never' }],
    },
    {
      code: 'type Foo = { a: number; b: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'type Foo = { a: number; b: number; }// "never"',
      options: [{ TSTypeLiteral: 'never' }],
    },
    {
      code: 'interface Foo { a: number; b: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'interface Foo { a: number; b: number; }// "never"',
      options: [{ TSInterfaceBody: 'never' }],
    },
    {
      code: 'type Foo = { a: number;\n    b: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'type Foo = { a: number;\n    b: number; }// "never"',
      options: [{ TSTypeLiteral: 'never' }],
    },
    {
      code: 'interface Foo { a: number;\n    b: number; }// "never"',
      options: ['never'],
    },
    {
      code: 'interface Foo { a: number;\n    b: number; }// "never"',
      options: [{ TSInterfaceBody: 'never' }],
    },
    {
      code: 'type Foo = {}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = {}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo {}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo {}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = { a: number; b: number; }// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = { a: number; b: number; }// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo { a: number; b: number; }// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo { a: number; b: number; }// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = {\n    // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = {\n    // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo {\n    // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo {\n    // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = { // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'type Foo = { // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
    },
    {
      code: 'interface Foo { // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
    },
    {
      code: 'interface Foo { // comment\n    a: number;\n}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
    },
    {
      code: 'type Foo = {}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'type Foo = {}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
    },
    {
      code: 'interface Foo {}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'interface Foo {}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
    },
    {
      code: 'type Foo = { a: number; }// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'type Foo = { a: number; }// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
    },
    {
      code: 'interface Foo { a: number; }// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'interface Foo { a: number; }// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
    },
    {
      code: 'type Foo = {}// default',
    },
    {
      code: 'interface Foo {}// default',
    },
    {
      code: 'type Foo = {}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = {}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo {}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo {}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = {\n}// default',
    },
    {
      code: 'interface Foo {\n}// default',
    },
    {
      code: 'type Foo = {\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = {\n}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo {\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo {\n}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = { a: number; }// default',
    },
    {
      code: 'interface Foo { a: number; }// default',
    },
    {
      code: 'type Foo = { a: number; }// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = { a: number; }// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo { a: number; }// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo { a: number; }// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n}// default',
    },
    {
      code: 'interface Foo {\n    a: number;\n}// default',
    },
    {
      code: 'type Foo = {\n    a: number;\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number; \n}// default',
    },
    {
      code: 'interface Foo {\n    a: number; b: number; \n}// default',
    },
    {
      code: 'type Foo = {\n    a: number; b: number; \n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number; \n}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number; \n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number; \n}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// default',
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// default',
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = { a: number;\nb: number; }// default',
    },
    {
      code: 'interface Foo { a: number;\nb: number; }// default',
    },
    {
      code: 'type Foo = { a: number;\nb: number; }// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'type Foo = { a: number;\nb: number; }// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
    },
    {
      code: 'interface Foo { a: number;\nb: number; }// {"consistent":true}',
      options: [{ consistent: true }],
    },
    {
      code: 'interface Foo { a: number;\nb: number; }// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number; \n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number; b: number; \n}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number; \n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number; b: number; \n}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number; \n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number; \n}// {"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { minProperties: 2, consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number; \n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number; \n}// {"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { minProperties: 2, consistent: true } }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'type Foo = {}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'interface Foo {}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'type Foo = { a: number; }// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'interface Foo { a: number; }// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number; \n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'type Foo = {\n    a: number;\n    b: number; \n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number; \n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },
    {
      code: 'interface Foo {\n    a: number;\n    b: number; \n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2, consistent: true } }],
    },
    {
      code: 'enum Foo {\n  A, B,\n}',
      options: ['always'],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: ['always'],
    },
    {
      code: 'enum Foo { A, B }',
      options: ['never'],
    },
    {
      code: 'enum Foo { A,\n  B }',
      options: ['never'],
    },
    {
      code: 'enum Foo { A, B }',
      options: [{ multiline: true }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ multiline: true }],
    },
    {
      code: 'enum Foo {\n  A, B\n}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ minProperties: 2 }],
    },
    {
      code: 'enum Foo {\n  A, B\n}',
      options: [{ consistent: true }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ consistent: true }],
    },
    {
      code: 'enum Foo {\n  A, B\n}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ multiline: true, minProperties: 2 }],
    },
    {
      code: 'enum Foo {\n  A, B\n}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ multiline: true, consistent: true }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ minProperties: 2, consistent: true }],
    },
    {
      code: 'enum Foo {\n  A,\n  B,\n}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
    },

    // ====== from the _js_ skipBabel block (Babel/Flow) — the 6 cases that ======
    // ====== parse as valid TS and behave identically to upstream.          ======
    // `MyType` annotations carry no object type literal, so rslint lints only
    // the ObjectPattern braces — matching upstream's Flow behaviour exactly.
    {
      code: 'function foo({\n a,\n b\n} : MyType) {}',
      options: ['always'],
    },
    {
      code: 'function foo({ a, b } : MyType) {}',
      options: ['never'],
    },
    // An inline object TYPE that is already single-line ('never') is untouched —
    // rslint's TSTypeLiteral handling and upstream's Flow result coincide here.
    {
      code: 'function foo({ a, b } : { a : string, b : string }) {}',
      options: ['never'],
    },
  ],
  invalid: [
    // ==================== from object-curly-newline._js_.test.ts ====================
    {
      code: 'var a = { a\n};',
      output: 'var a = { a};',
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var a = {};',
      output: 'var a = {\n};',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 10 },
      ],
    },
    {
      code: 'var b = {a: 1};',
      output: 'var b = {\na: 1\n};',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9, endLine: 1, endColumn: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 14, endLine: 1, endColumn: 15 },
      ],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      output: 'var c = {\na: 1, b: 2\n};',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 20 },
      ],
    },
    {
      code: 'var d = {a: 1,\n    b: 2};',
      output: 'var d = {\na: 1,\n    b: 2\n};',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 9 },
      ],
    },
    {
      code: 'var e = {a: function foo() {\n    dosomething();\n}};',
      output: 'var e = {\na: function foo() {\n    dosomething();\n}\n};',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'var a = {\n};',
      output: 'var a = {};',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      output: 'var b = {a: 1};',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9, endLine: 1, endColumn: 10 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1, endLine: 3, endColumn: 2 },
      ],
    },
    {
      code: 'var c = {\n    a: 1, b: 2\n};',
      output: 'var c = {a: 1, b: 2};',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var d = {\n    a: 1,\n    b: 2\n};',
      output: 'var d = {a: 1,\n    b: 2};',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'var e = {\n    a: function foo() {\n        dosomething();\n    }\n};',
      output: 'var e = {a: function foo() {\n        dosomething();\n    }};',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 5, column: 1 },
      ],
    },
    {
      code: 'var a = {\n};',
      output: 'var a = {};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var a = {\n /* comment */ \n};',
      output: null,
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var a = { // comment\n};',
      output: null,
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      output: 'var b = {a: 1};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var b = {\n   a: 1 // comment\n};',
      output: 'var b = {a: 1 // comment\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var c = {\n    a: 1, b: 2\n};',
      output: 'var c = {a: 1, b: 2};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var c = {\n    a: 1, b: 2 // comment\n};',
      output: 'var c = {a: 1, b: 2 // comment\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var d = {a: 1,\n    b: 2};',
      output: 'var d = {\na: 1,\n    b: 2\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 9 },
      ],
    },
    {
      code: 'var d = {a: 1, // comment\n    b: 2};',
      output: 'var d = {\na: 1, // comment\n    b: 2\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 9 },
      ],
    },
    {
      code: 'var e = {a: function foo() {\n    dosomething();\n}};',
      output: 'var e = {\na: function foo() {\n    dosomething();\n}\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'var e = {a: function foo() { // comment\n    dosomething();\n}};',
      output: 'var e = {\na: function foo() { // comment\n    dosomething();\n}\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'var e = {a: 1, /* comment */\n    b: 2, // another comment\n};',
      output: 'var e = {\na: 1, /* comment */\n    b: 2, // another comment\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
      ],
    },
    {
      code: 'var f = { /* comment */ a:\n2\n};',
      output: null,
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
      ],
    },
    {
      code: 'var f = {\n/* comment */\na: 1};',
      output: 'var f = {\n/* comment */\na: 1\n};',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 5 },
      ],
    },
    {
      code: 'var a = {\n};',
      output: 'var a = {};',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      output: 'var b = {a: 1};',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      output: 'var c = {\na: 1, b: 2\n};',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 20 },
      ],
    },
    {
      code: 'var d = {a: 1,\n    b: 2};',
      output: 'var d = {\na: 1,\n    b: 2\n};',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 9 },
      ],
    },
    {
      code: 'var e = {\n    a: function foo() {\n        dosomething();\n    }\n};',
      output: 'var e = {a: function foo() {\n        dosomething();\n    }};',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 5, column: 1 },
      ],
    },
    {
      code: 'var a = {\n};',
      output: 'var a = {};',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var b = {\n    a: 1\n};',
      output: 'var b = {a: 1};',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      output: 'var c = {\na: 1, b: 2\n};',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 20 },
      ],
    },
    {
      code: 'var d = {a: 1, \n    b: 2};',
      output: 'var d = {\na: 1, \n    b: 2\n};',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 9 },
      ],
    },
    {
      code: 'var e = {a: function foo() {\n    dosomething();\n}};',
      output: 'var e = {\na: function foo() {\n    dosomething();\n}\n};',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'var b = {a: 1\n};',
      output: 'var b = {a: 1};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var b = {\na: 1};',
      output: 'var b = {a: 1};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
      ],
    },
    {
      code: 'var c = {a: 1, b: 2\n};',
      output: 'var c = {a: 1, b: 2};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'var c = {\na: 1, b: 2};',
      output: 'var c = {a: 1, b: 2};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
      ],
    },
    {
      code: 'var c = {a: 1,\nb: 2};',
      output: 'var c = {\na: 1,\nb: 2\n};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 5 },
      ],
    },
    {
      code: 'var e = {a: function() {\ndosomething();\n}};',
      output: 'var e = {\na: function() {\ndosomething();\n}\n};',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'let {a\n} = {a: 1}',
      output: 'let {a} = {a: 1}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'let {\na} = {a: 1}',
      output: 'let {a} = {a: 1}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
      ],
    },
    {
      code: 'let {a, b\n} = {a: 1, b: 2}',
      output: 'let {a, b} = {a: 1, b: 2}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'let {\na, b} = {a: 1, b: 2}',
      output: 'let {a, b} = {a: 1, b: 2}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
      ],
    },
    {
      code: 'let {a,\nb} = {a: 1, b: 2}',
      output: 'let {\na,\nb\n} = {a: 1, b: 2}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 2 },
      ],
    },
    {
      code: 'let {e = function() {\ndosomething();\n}} = a;',
      output: 'let {\ne = function() {\ndosomething();\n}\n} = a;',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 2 },
      ],
    },
    {
      code: 'var c = {\na: 1,\nb: 2};',
      output: 'var c = {a: 1,\nb: 2};',
      options: [{ multiline: false, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
      ],
    },
    {
      code: 'var c = {a: 1,\nb: 2\n};',
      output: 'var c = {a: 1,\nb: 2};',
      options: [{ multiline: false, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'let {\na,\nb} = {a: 1, b: 2};',
      output: 'let {a,\nb} = {a: 1, b: 2};',
      options: [{ multiline: false, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
      ],
    },
    {
      code: 'let {a,\nb\n} = {a: 1, b: 2};',
      output: 'let {a,\nb} = {a: 1, b: 2};',
      options: [{ multiline: false, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'var c = {a: 1, b: 2};',
      output: 'var c = {\na: 1, b: 2\n};',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 9 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 20 },
      ],
    },
    {
      code: 'let {a, b} = {\na: 1, b: 2\n};',
      output: 'let {\na, b\n} = {\na: 1, b: 2\n};',
      options: [{ multiline: true, consistent: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 10 },
      ],
    },
    {
      code: 'let {\n    a, b\n} = {a: 1, b: 2};',
      output: 'let {a, b} = {\na: 1, b: 2\n};',
      options: [{ ObjectExpression: 'always', ObjectPattern: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 5 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 3, column: 5 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 3, column: 16 },
      ],
    },
    {
      code: 'import {\n    a,\n b\n} from \'module\';',
      output: 'import {a,\n b} from \'module\';',
      options: [{ ImportDeclaration: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'import {a, b} from \'module\';',
      output: 'import {\na, b\n} from \'module\';',
      options: [{ ImportDeclaration: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 13 },
      ],
    },
    {
      code: 'import {a as c, b} from \'module\';',
      output: 'import {\na as c, b\n} from \'module\';',
      options: [{ ImportDeclaration: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 18 },
      ],
    },
    {
      code: 'import {a, \nb} from \'module\';',
      output: 'import {\na, \nb\n} from \'module\';',
      options: [{ ImportDeclaration: { multiline: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 2 },
      ],
    },
    {
      code: 'import {a, \nb\n} from \'module\';',
      output: 'import {a, \nb} from \'module\';',
      options: [{ ImportDeclaration: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'import {a, b\n} from \'module\';',
      output: 'import {a, b} from \'module\';',
      options: [{ ImportDeclaration: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'import {a, b} from \'module\';',
      output: 'import {\na, b\n} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 13 },
      ],
    },
    {
      code: 'import {\na, b\n} from \'module\';',
      output: 'import {a, b} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 3 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'import DefaultExport, {a, b} from \'module\';',
      output: 'import DefaultExport, {\na, b\n} from \'module\';',
      options: [{ ImportDeclaration: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 23 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 28 },
      ],
    },
    {
      code: 'var a = 0; var b = 0;\nexport {\n    a,\n    b\n};',
      output: 'var a = 0; var b = 0;\nexport {a,\n    b};',
      options: [{ ExportDeclaration: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 2, column: 8 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 5, column: 1 },
      ],
    },
    {
      code: 'export {a as a, b} from \'module\';',
      output: 'export {\na as a, b\n} from \'module\';',
      options: [{ ExportDeclaration: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 18 },
      ],
    },
    {
      code: 'export {a, \nb} from \'module\';',
      output: 'export {\na, \nb\n} from \'module\';',
      options: [{ ExportDeclaration: { multiline: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 2 },
      ],
    },
    {
      code: 'export {a, \nb,\n} from \'module\';',
      output: 'export {a, \nb,} from \'module\';',
      options: [{ ExportDeclaration: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'export {a, b\n} from \'module\';',
      output: 'export {a, b} from \'module\';',
      options: [{ ExportDeclaration: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'export {a, b,} from \'module\';',
      output: 'export {\na, b,\n} from \'module\';',
      options: [{ ExportDeclaration: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 14 },
      ],
    },
    {
      code: 'export {\na, b\n} from \'module\';',
      output: 'export {a, b} from \'module\';',
      options: [{ ExportDeclaration: { minProperties: 3 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 8 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },

    // ==================== from object-curly-newline._ts_.test.ts ====================
    {
      code: 'type Foo = {}// "always"',
      output: 'type Foo = {\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 13 },
      ],
    },
    {
      code: 'type Foo = {}// "always"',
      output: 'type Foo = {\n}// "always"',
      options: [{ TSTypeLiteral: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 13 },
      ],
    },
    {
      code: 'interface Foo {}// "always"',
      output: 'interface Foo {\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 16 },
      ],
    },
    {
      code: 'interface Foo {}// "always"',
      output: 'interface Foo {\n}// "always"',
      options: [{ TSInterfaceBody: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 16 },
      ],
    },
    {
      code: 'type Foo = {a: number;}// "always"',
      output: 'type Foo = {\na: number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 23 },
      ],
    },
    {
      code: 'type Foo = {a: number;}// "always"',
      output: 'type Foo = {\na: number;\n}// "always"',
      options: [{ TSTypeLiteral: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 23 },
      ],
    },
    {
      code: 'interface Foo {a: number;}// "always"',
      output: 'interface Foo {\na: number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 26 },
      ],
    },
    {
      code: 'interface Foo {a: number;}// "always"',
      output: 'interface Foo {\na: number;\n}// "always"',
      options: [{ TSInterfaceBody: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 26 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// "always"',
      output: 'type Foo = {\na: number;b:number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// "always"',
      output: 'type Foo = {\na: number;b:number;\n}// "always"',
      options: [{ TSTypeLiteral: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// "always"',
      output: 'interface Foo {\na: number;b:number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// "always"',
      output: 'interface Foo {\na: number;b:number;\n}// "always"',
      options: [{ TSInterfaceBody: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// "always"',
      output: 'type Foo = {\na: number;\n  b:number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// "always"',
      output: 'type Foo = {\na: number;\n  b:number;\n}// "always"',
      options: [{ TSTypeLiteral: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// "always"',
      output: 'interface Foo {\na: number;\n  b:number;\n}// "always"',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// "always"',
      output: 'interface Foo {\na: number;\n  b:number;\n}// "always"',
      options: [{ TSInterfaceBody: 'always' }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\n}// "never"',
      output: 'type Foo = {}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n}// "never"',
      output: 'type Foo = {}// "never"',
      options: [{ TSTypeLiteral: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// "never"',
      output: 'interface Foo {}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// "never"',
      output: 'interface Foo {}// "never"',
      options: [{ TSInterfaceBody: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// "never"',
      output: 'type Foo = {a: number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// "never"',
      output: 'type Foo = {a: number;}// "never"',
      options: [{ TSTypeLiteral: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// "never"',
      output: 'interface Foo {a: number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// "never"',
      output: 'interface Foo {a: number;}// "never"',
      options: [{ TSInterfaceBody: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na: number;b:number;\n}// "never"',
      output: 'type Foo = {a: number;b:number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na: number;b:number;\n}// "never"',
      output: 'type Foo = {a: number;b:number;}// "never"',
      options: [{ TSTypeLiteral: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\na: number;b:number;\n}// "never"',
      output: 'interface Foo {a: number;b:number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\na: number;b:number;\n}// "never"',
      output: 'interface Foo {a: number;b:number;}// "never"',
      options: [{ TSInterfaceBody: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n  b:number;\n}// "never"',
      output: 'type Foo = {a: number;\n  b:number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n  b:number;\n}// "never"',
      output: 'type Foo = {a: number;\n  b:number;}// "never"',
      options: [{ TSTypeLiteral: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n  b:number;\n}// "never"',
      output: 'interface Foo {a: number;\n  b:number;}// "never"',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n  b:number;\n}// "never"',
      output: 'interface Foo {a: number;\n  b:number;}// "never"',
      options: [{ TSInterfaceBody: 'never' }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true}',
      output: 'type Foo = {}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true}',
      output: 'type Foo = {}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true}',
      output: 'interface Foo {}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true}',
      output: 'interface Foo {}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"multiline":true}',
      output: 'type Foo = {a: number;}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"multiline":true}',
      output: 'type Foo = {a: number;}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"multiline":true}',
      output: 'interface Foo {a: number;}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"multiline":true}',
      output: 'interface Foo {a: number;}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;b:number;\n}// {"multiline":true}',
      output: 'type Foo = {a: number;b:number;}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;b:number;\n}// {"multiline":true}',
      output: 'type Foo = {a: number;b:number;}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;b:number;\n}// {"multiline":true}',
      output: 'interface Foo {a: number;b:number;}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;b:number;\n}// {"multiline":true}',
      output: 'interface Foo {a: number;b:number;}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true}',
      options: [{ TSTypeLiteral: { multiline: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true}',
      options: [{ TSInterfaceBody: { multiline: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"minProperties":2}',
      output: 'type Foo = {}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"minProperties":2}',
      output: 'type Foo = {}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"minProperties":2}',
      output: 'interface Foo {}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"minProperties":2}',
      output: 'interface Foo {}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"minProperties":2}',
      output: 'type Foo = {a: number;}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"minProperties":2}',
      output: 'type Foo = {a: number;}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"minProperties":2}',
      output: 'interface Foo {a: number;}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"minProperties":2}',
      output: 'interface Foo {a: number;}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"minProperties":2}',
      output: 'type Foo = {\na: number;b:number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"minProperties":2}',
      output: 'type Foo = {\na: number;b:number;\n}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"minProperties":2}',
      output: 'interface Foo {\na: number;b:number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"minProperties":2}',
      output: 'interface Foo {\na: number;b:number;\n}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"minProperties":2}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"minProperties":2}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"minProperties":2}',
      options: [{ TSTypeLiteral: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"minProperties":2}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"minProperties":2}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"minProperties":2}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"minProperties":2}',
      options: [{ TSInterfaceBody: { minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a:number;\n}// default',
      output: 'type Foo = {a:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;\n}// default',
      output: 'interface Foo {a:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;\n}// {"consistent":true}',
      output: 'type Foo = {a:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;\n}// {"consistent":true}',
      output: 'type Foo = {a:number;}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;\n}// {"consistent":true}',
      output: 'interface Foo {a:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;\n}// {"consistent":true}',
      output: 'interface Foo {a:number;}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na:number;}// default',
      output: 'type Foo = {a:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;}// default',
      output: 'interface Foo {a:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {\na:number;}// {"consistent":true}',
      output: 'type Foo = {a:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\na:number;}// {"consistent":true}',
      output: 'type Foo = {a:number;}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;}// {"consistent":true}',
      output: 'interface Foo {a:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'interface Foo {\na:number;}// {"consistent":true}',
      output: 'interface Foo {a:number;}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {a:number;b:number;\n}// default',
      output: 'type Foo = {a:number;b:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;b:number;\n}// default',
      output: 'interface Foo {a:number;b:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;b:number;\n}// {"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;b:number;\n}// {"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;b:number;\n}// {"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;b:number;\n}// {"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na:number;b:number;}// default',
      output: 'type Foo = {a:number;b:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;b:number;}// default',
      output: 'interface Foo {a:number;b:number;}// default',
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {\na:number;b:number;}// {"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\na:number;b:number;}// {"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"consistent":true}',
      options: [{ TSTypeLiteral: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;b:number;}// {"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"consistent":true}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'interface Foo {\na:number;b:number;}// {"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"consistent":true}',
      options: [{ TSInterfaceBody: { consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {a: number;}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\n  a: number;\n}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {a: number;}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {a: number;}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'interface Foo {\n  a: number;\n}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {a: number;}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"minProperties":2}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ multiline: true, minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"minProperties":2}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2 } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a:number;\n}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;\n}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;\n}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;\n}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\na:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'interface Foo {\na:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {a:number;b:number;\n}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {a:number;b:number;\n}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;b:number;\n}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'interface Foo {a:number;b:number;\n}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'type Foo = {\na:number;b:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'type Foo = {\na:number;b:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
      ],
    },
    {
      code: 'interface Foo {\na:number;b:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'interface Foo {\na:number;b:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {a:number;b:number;}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"consistent":true}',
      options: [{ multiline: true, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'type Foo = {a: number;b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 32 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'interface Foo {a: number;b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 35 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'type Foo = {a: number;\n  b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'type Foo = {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSTypeLiteral: { multiline: true, minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 12 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'interface Foo {a: number;\n  b:number;}// {"multiline":true,"minProperties":2,"consistent":true}',
      output: 'interface Foo {\na: number;\n  b:number;\n}// {"multiline":true,"minProperties":2,"consistent":true}',
      options: [{ TSInterfaceBody: { multiline: true, minProperties: 2, consistent: true } }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 15 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 12 },
      ],
    },
    {
      code: 'enum Foo {}',
      output: 'enum Foo {\n}',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 11 },
      ],
    },
    {
      code: 'enum Foo {A,B}',
      output: 'enum Foo {\nA,B\n}',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 14 },
      ],
    },
    {
      code: 'enum Foo {A,\n  B}',
      output: 'enum Foo {\nA,\n  B\n}',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 4 },
      ],
    },
    {
      code: 'enum Foo {\n}',
      output: 'enum Foo {}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 2, column: 1 },
      ],
    },
    {
      code: 'enum Foo {\n  A,B\n}',
      output: 'enum Foo {A,B}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'enum Foo {\n  A,\n  B\n}',
      output: 'enum Foo {A,\n  B}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'enum Foo {\n  A,B\n}',
      output: 'enum Foo {A,B}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 3, column: 1 },
      ],
    },
    {
      code: 'enum Foo {A,\n  B}',
      output: 'enum Foo {\nA,\n  B\n}',
      options: [{ multiline: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 4 },
      ],
    },
    {
      code: 'enum Foo {A,B}',
      output: 'enum Foo {\nA,B\n}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 14 },
      ],
    },
    {
      code: 'enum Foo {A,\n  B}',
      output: 'enum Foo {\nA,\n  B\n}',
      options: [{ minProperties: 2 }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 4 },
      ],
    },
    {
      code: 'enum Foo {\n  A,B}',
      output: 'enum Foo {A,B}',
      options: [{ consistent: true }],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
      ],
    },
    {
      code: 'enum Foo {A,B}',
      output: 'enum Foo {\nA,B\n}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 14 },
      ],
    },
    {
      code: 'enum Foo {A,\n  B}',
      output: 'enum Foo {\nA,\n  B\n}',
      options: [{ multiline: true, minProperties: 2, consistent: true }],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 10 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 2, column: 4 },
      ],
    },

    // ====== from the _js_ skipBabel block (Babel/Flow) — the 3 cases that ======
    // ====== parse as valid TS and behave identically to upstream.         ======
    {
      code: 'function foo({ a, b } : MyType) {}',
      output: 'function foo({\n a, b \n} : MyType) {}',
      options: ['always'],
      errors: [
        { messageId: 'expectedLinebreakAfterOpeningBrace', line: 1, column: 14 },
        { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 21 },
      ],
    },
    {
      code: 'function foo({\n a,\n b\n} : MyType) {}',
      output: 'function foo({a,\n b} : MyType) {}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 14 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
    {
      code: 'function foo({\n a,\n b\n} : { a : string, b : string }) {}',
      output: 'function foo({a,\n b} : { a : string, b : string }) {}',
      options: ['never'],
      errors: [
        { messageId: 'unexpectedLinebreakAfterOpeningBrace', line: 1, column: 14 },
        { messageId: 'unexpectedLinebreakBeforeClosingBrace', line: 4, column: 1 },
      ],
    },
  ],
});

/**
 * ===================== object-curly-newline — KNOWN GAPS =====================
 *
 * From the `_js_` `skipBabel` block (`object-curly-newline_babel`). Upstream
 * parses these with `@babel/eslint-parser` + the Flow plugin, where the inline
 * object TYPE annotation `{ a : string, b : string }` on a destructured param is
 * a Flow type node the rule does NOT visit — so upstream lints only the
 * ObjectPattern `{ a, b }`.
 *
 * rslint parses with ts-go, where `{ a : string, b : string }` is a
 * `TSTypeLiteral` — a node this very rule is designed to lint (see the whole
 * `type Foo = { ... }` / `interface Foo { ... }` suite above). rslint therefore
 * (correctly, for TypeScript) ALSO reports brace placement on that type literal,
 * producing extra diagnostics relative to the Babel/Flow expectation. This is a
 * parser-semantics divergence, not a rule bug: the rule logic is right; the two
 * parsers expose different node trees for the same source.
 *
 * Only the two `always`-option cases whose inline object type is single-line
 * diverge (the type literal then wants linebreaks added). The other six
 * Babel/Flow cases coincide with rslint and are kept in the green set above.
 *
 * ---- valid (upstream expects 0 diagnostics) ----
 *
 *   {
 *     code: 'function foo({\n a,\n b\n} : { a : string, b : string }) {}',
 *     options: ['always'],
 *   }
 *
 *   rslint: the ObjectPattern is already multiline (0 diags there, matching
 *   upstream) but the single-line type literal `{ a : string, b : string }` is a
 *   TSTypeLiteral, so 'always' makes rslint emit 2 diagnostics on it:
 *     expectedLinebreakAfterOpeningBrace  (line 4, column 5)
 *     expectedLinebreakBeforeClosingBrace (line 4, column 30)
 *
 * ---- invalid (upstream expects exactly 2 diagnostics on the ObjectPattern) ----
 *
 *   {
 *     code: 'function foo({ a, b } : { a : string, b : string }) {}',
 *     output: [
 *       'function foo({',
 *       ' a, b ',
 *       '} : { a : string, b : string }) {}',
 *     ].join('\n'),
 *     options: ['always'],
 *     errors: [
 *       { messageId: 'expectedLinebreakAfterOpeningBrace',  line: 1, column: 14 },
 *       { messageId: 'expectedLinebreakBeforeClosingBrace', line: 1, column: 21 },
 *     ],
 *   }
 *
 *   rslint emits the upstream pair on the ObjectPattern (line 1, columns 14/21)
 *   PLUS two more on the TSTypeLiteral `{ a : string, b : string }`:
 *     expectedLinebreakAfterOpeningBrace  (line 1, column 25)
 *     expectedLinebreakBeforeClosingBrace (line 1, column 50)
 *   i.e. 4 diagnostics total, so the exact-count assertion can't match upstream.
 *
 * ============================================================================
 */
