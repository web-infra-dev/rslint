// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, caseSensitiveFileSystem } from '../config';
new RuleTester(fixtureConfig()).run('named', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import { foo } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: 'import { foo } from "./empty-module"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: 'import bar from "./bar.js"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import bar, { foo } from "./bar.js"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: 'import {a, b, d} from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import {ExportedClass} from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 6',
      code: 'import { destructingAssign } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'import { destructingRenamedAssign } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: 'import { ActionTypes } from "./qc"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: 'import {a, b, c, d} from "./re-export"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 10',
      code: 'import {a, b, c} from "./re-export-common-star"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: 'import {RuleTester} from "./re-export-node_modules"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'import { jsxFoo } from "./jsx/AnotherComponent"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/resolve': {
          extensions: ['.js', '.jsx'],
        },
      },
    },
    {
      name: 'run 0 valid 13',
      code: 'import {a, b, d} from "./common"; // eslint-disable-line import/named',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'import { foo, bar } from "./re-export-names"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'import { foo, bar } from "./common"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['common'],
      },
    },
    {
      name: 'run 0 valid 16',
      code: 'import { foo } from "crypto"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'import { zoob } from "a"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 18',
      code: 'import { someThing } from "./test-module"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 19',
      code: 'export { foo } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 20',
      code: 'export { foo as bar } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 21',
      code: 'export { foo } from "./does-not-exist"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 22',
      code: 'import { foo, bar } from "./named-trampoline"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 23',
      code: 'let foo; export { foo as bar }',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 24',
      code: 'import { destructuredProp } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 25',
      code: 'import { arrayKeyProp } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 26',
      code: 'import { deepProp } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 27',
      code: 'import { deepSparseElement } from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 28',
      code: 'import type { MissingType } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 29',
      code: 'import type { MyOpaqueType } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 30',
      code: 'import { type MyOpaqueType, MyClass } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 31',
      code: 'export type { MissingType } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 32',
      code: 'export type { MyOpaqueType } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 33',
      code: '/*jsnext*/ import { createStore } from "./redux"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': [],
      },
    },
    {
      name: 'run 0 valid 34',
      code: '/*jsnext*/ import { createStore } from "./redux"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 35',
      code: 'import { foo } from "./es6-module"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 36',
      code: 'import { me, soGreat } from "./narcissist"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 37',
      code: 'import { foo, bar, baz } from "./re-export-default"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 38',
      code: 'import { common } from "./re-export-default"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['common'],
      },
    },
    {
      name: 'run 0 valid 39',
      code: 'import {a, b, d} from "./common"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 40',
      code: 'import { baz } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['bar'],
      },
    },
    {
      name: 'run 0 valid 41',
      code: 'import { common } from "./re-export-default"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 42',
      code: 'const { destructuredProp } = require("./named-exports")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 43',
      code: 'let { arrayKeyProp } = require("./named-exports")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 44',
      code: 'const { deepProp } = require("./named-exports")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 45',
      code: 'const { foo, bar } = require("./re-export-names")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 46',
      code: 'const { baz } = require("./bar")',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 47',
      code: 'const { baz } = require("./bar")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: false,
        },
      ],
    },
    {
      name: 'run 0 valid 48',
      code: 'const { default: defExport } = require("./bar")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 49',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 50',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 51',
      code: 'const { x, y } = bar',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 52',
      code: 'const { x, y, ...z } = bar',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 53',
      code: 'let x; export { x }',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 54',
      code: 'let x; export { x as y }',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 55',
      code: 'export const x = null',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 56',
      code: 'export var x = null',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 57',
      code: 'export let x = null',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 58',
      code: 'export default x',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 59',
      code: 'export default class x {}',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 60',
      code: 'import json from "./data.json"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 61',
      code: 'import foo from "./foobar.json";',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 62',
      code: 'import foo from "./foobar";',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 63',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 64',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 65',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 66',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 67',
      code: "import { ExtfieldModel, Extfield2Model } from './models';",
      filename: 'fixtures/named-rule/export-star/downstream.js',
    },
    {
      name: 'run 0 valid 68',
      code: 'const { something } = require("./dynamic-import-in-commonjs")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
    },
    {
      name: 'run 0 valid 69',
      code: 'import { something } from "./dynamic-import-in-commonjs"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 70',
      code: 'import { "foo" as foo } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 71',
      code: 'import { "foo" as foo } from "./empty-module"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 1 valid 0',
      code: 'import { b } from "./Named-Exports"',
      filename: 'fixtures/named-rule/consumer.ts',
      skip: caseSensitiveFileSystem
        ? 'This case requires case-insensitive lookup of Named-Exports.js; the fixture filesystem is case-sensitive.'
        : undefined,
    },
    {
      name: 'run 2 valid 0',
      code: 'import { foo } from "./export-all"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 0',
      code: "import x from './typescript-export-assign-object'",
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 1',
      code: 'import { MyType } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 2',
      code: 'import { Foo } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 3',
      code: 'import { Bar } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 4',
      code: 'import { getFoo } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 5',
      code: 'import { MyEnum } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 6',
      code: '\n              import { MyModule } from "./typescript"\n              MyModule.ModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 7',
      code: '\n              import { MyNamespace } from "./typescript"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 8',
      code: 'import { MyType } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 9',
      code: 'import { Foo } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 10',
      code: 'import { Bar } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 11',
      code: 'import { getFoo } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 12',
      code: 'import { MyEnum } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 13',
      code: '\n              import { MyModule } from "./typescript-declare"\n              MyModule.ModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 14',
      code: '\n              import { MyNamespace } from "./typescript-declare"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 15',
      code: 'import { MyType } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 16',
      code: 'import { Foo } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 17',
      code: 'import { Bar } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 18',
      code: 'import { getFoo } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 19',
      code: 'import { MyEnum } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 20',
      code: '\n              import { MyModule } from "./typescript-export-assign-namespace"\n              MyModule.ModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 21',
      code: '\n              import { MyNamespace } from "./typescript-export-assign-namespace"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 22',
      code: 'import { MyType } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 23',
      code: 'import { Foo } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 24',
      code: 'import { Bar } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 25',
      code: 'import { getFoo } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 26',
      code: 'import { MyEnum } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 27',
      code: '\n              import { MyModule } from "./typescript-export-assign-namespace-merged"\n              MyModule.ModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 3 valid 28',
      code: '\n              import { MyNamespace } from "./typescript-export-assign-namespace-merged"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 4 valid 0',
      code: 'export const foo = "I\'m so foo"',
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 4 valid 1',
      code: "import { foo } from './foo'",
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 4 valid 2',
      code: "export { foo as bar } from './foo'",
      filename: 'fixtures/named-rule/consumer.ts',
    },
    {
      name: 'run 4 valid 3',
      code: "import { SomeNonsenseThatDoesntExist } from 'react'",
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['node_modules'],
      },
    },
    {
      name: 'run 4 valid 4',
      code: "import { notWhatever } from './whatever'",
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': ['node_modules', '\\.coffee$'],
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import { somethingElse } from "./test-module"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "somethingElse not found in './test-module'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import { baz } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import { baz, bop } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: '',
          message: "bop not found in './bar'",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import {a, b, c} from "./named-exports"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "c not found in './named-exports'",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import { a } from "./default-export"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "a not found in './default-export'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import { ActionTypess } from "./qc"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "ActionTypess not found in './qc'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import {a, b, c, d, e} from "./re-export"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "e not found in './re-export'",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import { a } from "./re-export-names"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "a not found in './re-export-names'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'export { bar } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "bar not found in './bar'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'import { foo, bar, baz } from "./named-trampoline"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "baz not found in './named-trampoline'",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'import { baz } from "./broken-trampoline"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: 'baz not found via broken-trampoline.js -> named-exports.js',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'const { baz } = require("./bar")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'let { baz } = require("./bar")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'const { baz: bar, bop } = require("./bar"), { a } = require("./re-export-names")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 12,
        },
        {
          messageId: '',
          message: "bop not found in './bar'",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: '',
          message: "a not found in './re-export-names'",
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'const { default: defExport } = require("./named-exports")',
      filename: 'fixtures/named-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: '',
          message: "default not found in './named-exports'",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'import  { type MyOpaqueType, MyMissingClass } from "./flowtypes"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "MyMissingClass not found in './flowtypes'",
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: '/*jsnext*/ import { createSnorlax } from "./redux"',
      filename: 'fixtures/named-rule/consumer.ts',
      settings: {
        'import/ignore': [],
      },
      errors: [
        {
          messageId: '',
          message: "createSnorlax not found in './redux'",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: '/*jsnext*/ import { createSnorlax } from "./redux"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "createSnorlax not found in './redux'",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'import { baz } from "./es6-module"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "baz not found in './es6-module'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'import { foo, bar, bap } from "./re-export-default"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "bap not found in './re-export-default'",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'import { default as barDefault } from "./re-export"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "default not found in './re-export'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: 'import { "somethingElse" as somethingElse } from "./test-module"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "somethingElse not found in './test-module'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: 'import { "baz" as baz, "bop" as bop } from "./bar"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "baz not found in './bar'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: '',
          message: "bop not found in './bar'",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: 'import { "default" as barDefault } from "./re-export"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "default not found in './re-export'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 1 invalid 0',
      code: 'import { foo } from "./Named-Exports"',
      filename: 'fixtures/named-rule/consumer.ts',
      skip: caseSensitiveFileSystem
        ? 'This case requires case-insensitive lookup of Named-Exports.js; the fixture filesystem is case-sensitive.'
        : undefined,
      errors: [
        {
          messageId: '',
          message: "foo not found in './Named-Exports'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 2 invalid 0',
      code: 'import { bar } from "./export-all"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "bar not found in './export-all'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      name: 'run 3 invalid 0',
      code: "import { NotExported } from './typescript-export-assign-object'",
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message:
            "NotExported not found in './typescript-export-assign-object'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 1',
      code: "import { FooBar } from './typescript-export-assign-object'",
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "FooBar not found in './typescript-export-assign-object'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 3 invalid 2',
      code: 'import { MissingType } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "MissingType not found in './typescript'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 3',
      code: 'import { NotExported } from "./typescript"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "NotExported not found in './typescript'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 4',
      code: 'import { MissingType } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "MissingType not found in './typescript-declare'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 5',
      code: 'import { NotExported } from "./typescript-declare"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "NotExported not found in './typescript-declare'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 6',
      code: 'import { MissingType } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message:
            "MissingType not found in './typescript-export-assign-namespace'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 7',
      code: 'import { NotExported } from "./typescript-export-assign-namespace"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message:
            "NotExported not found in './typescript-export-assign-namespace'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 8',
      code: 'import { MissingType } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message:
            "MissingType not found in './typescript-export-assign-namespace-merged'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 3 invalid 9',
      code: 'import { NotExported } from "./typescript-export-assign-namespace-merged"',
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message:
            "NotExported not found in './typescript-export-assign-namespace-merged'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 4 invalid 0',
      code: "import { notFoo } from './foo'",
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "notFoo not found in './foo'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 4 invalid 1',
      code: "export { notFoo as defNotBar } from './foo'",
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "notFoo not found in './foo'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 4 invalid 2',
      code: "import { dontCreateStore } from './redux'",
      filename: 'fixtures/named-rule/consumer.ts',
      errors: [
        {
          messageId: '',
          message: "dontCreateStore not found in './redux'",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
  ],
});
