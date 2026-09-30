import { existsSync } from 'node:fs';
import { RuleTester } from '../rule-tester.js';
import { testFixturePath } from '../utils.js';

// eslint-plugin-import v2.32.0 tests/src/rules/named.js (all semantic cases),
// expanded SYNTAX_CASES, and docs/rules/named.md. Package entries are localized;
// dependency-only Babel/Flow syntax uses equivalent ES/TypeScript fixtures.
// This wrapper checks messages; Go asserts IDs, full ranges and absence of edits.
const filename = testFixturePath('named-rule/consumer.ts');

// named
new RuleTester().run('named', null as never, {
  valid: [
    // Skipped: The test Program rejects malformed dependencies before rule execution.
    // import "./malformed.js"
    { code: 'import { foo } from "./bar"' },
    { code: 'import { foo } from "./empty-module"' },
    { code: 'import bar from "./bar.js"' },
    { code: 'import bar, { foo } from "./bar.js"' },
    { code: 'import {a, b, d} from "./named-exports"' },
    { code: 'import {ExportedClass} from "./named-exports"' },
    { code: 'import { destructingAssign } from "./named-exports"' },
    { code: 'import { destructingRenamedAssign } from "./named-exports"' },
    { code: 'import { ActionTypes } from "./qc"' },
    { code: 'import {a, b, c, d} from "./re-export"' },
    { code: 'import {a, b, c} from "./re-export-common-star"' },
    { code: 'import {RuleTester} from "./re-export-node_modules"' },
    {
      code: 'import { jsxFoo } from "./jsx/AnotherComponent"',
      settings: { 'import/resolve': { extensions: ['.js', '.jsx'] } },
    },
    {
      code: 'import {a, b, d} from "./common"; // eslint-disable-line import/named',
    },
    { code: 'import { foo, bar } from "./re-export-names"' },
    {
      code: 'import { foo, bar } from "./common"',
      settings: { 'import/ignore': ['common'] },
    },
    { code: 'import { foo } from "crypto"' },
    { code: 'import { zoob } from "a"' },
    { code: 'import { someThing } from "./test-module"' },
    { code: 'export { foo } from "./bar"' },
    { code: 'export { foo as bar } from "./bar"' },
    { code: 'export { foo } from "./does-not-exist"' },
    // Skipped: Babel export-default-from declarations are not parsed by tsgo.
    // export bar, { foo } from "./bar"
    { code: 'import { foo, bar } from "./named-trampoline"' },
    { code: 'let foo; export { foo as bar }' },
    { code: 'import { destructuredProp } from "./named-exports"' },
    { code: 'import { arrayKeyProp } from "./named-exports"' },
    { code: 'import { deepProp } from "./named-exports"' },
    { code: 'import { deepSparseElement } from "./named-exports"' },
    { code: 'import type { MissingType } from "./flowtypes"' },
    // Skipped: Flow typeof imports are not parsed by tsgo.
    // import typeof { MissingType } from "./flowtypes"
    { code: 'import type { MyOpaqueType } from "./flowtypes"' },
    // Skipped: Flow typeof imports are not parsed by tsgo.
    // import typeof { MyOpaqueType } from "./flowtypes"
    { code: 'import { type MyOpaqueType, MyClass } from "./flowtypes"' },
    // Skipped: Flow typeof imports are not parsed by tsgo.
    // import { typeof MyOpaqueType, MyClass } from "./flowtypes"
    // Skipped: Flow typeof imports are not parsed by tsgo.
    // import typeof MissingType from "./flowtypes"
    // Skipped: Flow typeof imports are not parsed by tsgo.
    // import typeof * as MissingType from "./flowtypes"
    { code: 'export type { MissingType } from "./flowtypes"' },
    { code: 'export type { MyOpaqueType } from "./flowtypes"' },
    {
      code: '/*jsnext*/ import { createStore } from "./redux"',
      settings: { 'import/ignore': [] },
    },
    { code: '/*jsnext*/ import { createStore } from "./redux"' },
    { code: 'import { foo } from "./es6-module"' },
    { code: 'import { me, soGreat } from "./narcissist"' },
    { code: 'import { foo, bar, baz } from "./re-export-default"' },
    {
      code: 'import { common } from "./re-export-default"',
      settings: { 'import/ignore': ['common'] },
    },
    { code: 'import {a, b, d} from "./common"' },
    {
      code: 'import { baz } from "./bar"',
      settings: { 'import/ignore': ['bar'] },
    },
    { code: 'import { common } from "./re-export-default"' },
    {
      code: 'const { destructuredProp } = require("./named-exports")',
      options: [{ commonjs: true }],
    },
    {
      code: 'let { arrayKeyProp } = require("./named-exports")',
      options: [{ commonjs: true }],
    },
    {
      code: 'const { deepProp } = require("./named-exports")',
      options: [{ commonjs: true }],
    },
    {
      code: 'const { foo, bar } = require("./re-export-names")',
      options: [{ commonjs: true }],
    },
    { code: 'const { baz } = require("./bar")' },
    {
      code: 'const { baz } = require("./bar")',
      options: [{ commonjs: false }],
    },
    {
      code: 'const { default: defExport } = require("./bar")',
      options: [{ commonjs: true }],
    },
    { code: 'for (let { foo, bar } of baz) {}' },
    { code: 'for (let [ foo, bar ] of baz) {}' },
    { code: 'const { x, y } = bar' },
    { code: 'const { x, y, ...z } = bar' },
    { code: 'let x; export { x }' },
    { code: 'let x; export { x as y }' },
    { code: 'export const x = null' },
    { code: 'export var x = null' },
    { code: 'export let x = null' },
    { code: 'export default x' },
    { code: 'export default class x {}' },
    {
      code: 'import json from "./data.json"',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar.json";',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import foo from "./foobar";',
      settings: { 'import/extensions': ['.js'] },
    },
    {
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      settings: { 'import/ignore': ['foo'] },
    },
    {
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      settings: { 'import/ignore': ['foo'] },
    },
    { code: 'import * as a from "./commonjs-namespace/a"; a.b' },
    { code: 'import { foo } from "./ignore.invalid.extension"' },
    {
      code: "import { ExtfieldModel, Extfield2Model } from './models';",
      filename: testFixturePath('named-rule/export-star/downstream.js'),
    },
    {
      code: 'const { something } = require("./dynamic-import-in-commonjs")',
      options: [{ commonjs: true }],
    },
    { code: 'import { something } from "./dynamic-import-in-commonjs"' },
    { code: 'import { "foo" as foo } from "./bar"' },
    { code: 'import { "foo" as foo } from "./empty-module"' },
  ].map((test) => ({ filename, ...test })),
  invalid: [
    {
      code: 'import { somethingElse } from "./test-module"',
      errors: [{ message: "somethingElse not found in './test-module'" }],
    },
    {
      code: 'import { baz } from "./bar"',
      errors: [{ message: "baz not found in './bar'" }],
    },
    {
      code: 'import { baz, bop } from "./bar"',
      errors: [
        { message: "baz not found in './bar'" },
        { message: "bop not found in './bar'" },
      ],
    },
    {
      code: 'import {a, b, c} from "./named-exports"',
      errors: [{ message: "c not found in './named-exports'" }],
    },
    {
      code: 'import { a } from "./default-export"',
      errors: [{ message: "a not found in './default-export'" }],
    },
    {
      code: 'import { ActionTypess } from "./qc"',
      errors: [{ message: "ActionTypess not found in './qc'" }],
    },
    {
      code: 'import {a, b, c, d, e} from "./re-export"',
      errors: [{ message: "e not found in './re-export'" }],
    },
    {
      code: 'import { a } from "./re-export-names"',
      errors: [{ message: "a not found in './re-export-names'" }],
    },
    {
      code: 'export { bar } from "./bar"',
      errors: [{ message: "bar not found in './bar'" }],
    },
    // Skipped: Babel export-default-from declarations are not parsed by tsgo.
    // export bar2, { bar } from "./bar"
    {
      code: 'import { foo, bar, baz } from "./named-trampoline"',
      errors: [{ message: "baz not found in './named-trampoline'" }],
    },
    {
      code: 'import { baz } from "./broken-trampoline"',
      errors: [
        {
          message: 'baz not found via broken-trampoline.js -> named-exports.js',
        },
      ],
    },
    {
      code: 'const { baz } = require("./bar")',
      options: [{ commonjs: true }],
      errors: [{ message: "baz not found in './bar'" }],
    },
    {
      code: 'let { baz } = require("./bar")',
      options: [{ commonjs: true }],
      errors: [{ message: "baz not found in './bar'" }],
    },
    {
      code: 'const { baz: bar, bop } = require("./bar"), { a } = require("./re-export-names")',
      options: [{ commonjs: true }],
      errors: [
        { message: "baz not found in './bar'" },
        { message: "bop not found in './bar'" },
        { message: "a not found in './re-export-names'" },
      ],
    },
    {
      code: 'const { default: defExport } = require("./named-exports")',
      options: [{ commonjs: true }],
      errors: [{ message: "default not found in './named-exports'" }],
    },
    {
      code: 'import  { type MyOpaqueType, MyMissingClass } from "./flowtypes"',
      errors: [{ message: "MyMissingClass not found in './flowtypes'" }],
    },
    {
      code: '/*jsnext*/ import { createSnorlax } from "./redux"',
      settings: { 'import/ignore': [] },
      errors: [{ message: "createSnorlax not found in './redux'" }],
    },
    {
      code: '/*jsnext*/ import { createSnorlax } from "./redux"',
      errors: [{ message: "createSnorlax not found in './redux'" }],
    },
    {
      code: 'import { baz } from "./es6-module"',
      errors: [{ message: "baz not found in './es6-module'" }],
    },
    {
      code: 'import { foo, bar, bap } from "./re-export-default"',
      errors: [{ message: "bap not found in './re-export-default'" }],
    },
    {
      code: 'import { default as barDefault } from "./re-export"',
      errors: [{ message: "default not found in './re-export'" }],
    },
    {
      code: 'import { "somethingElse" as somethingElse } from "./test-module"',
      errors: [{ message: "somethingElse not found in './test-module'" }],
    },
    {
      code: 'import { "baz" as baz, "bop" as bop } from "./bar"',
      errors: [
        { message: "baz not found in './bar'" },
        { message: "bop not found in './bar'" },
      ],
    },
    {
      code: 'import { "default" as barDefault } from "./re-export"',
      errors: [{ message: "default not found in './re-export'" }],
    },
  ].map((test) => ({ filename, ...test })),
});

if (existsSync(testFixturePath('named-rule/Named-Exports.js'))) {
  // named (path case-insensitivity)
  new RuleTester().run('named', null as never, {
    valid: [{ code: 'import { b } from "./Named-Exports"' }].map((test) => ({
      filename,
      ...test,
    })),
    invalid: [
      {
        code: 'import { foo } from "./Named-Exports"',
        errors: [{ message: "foo not found in './Named-Exports'" }],
      },
    ].map((test) => ({ filename, ...test })),
  });
}

// named (export *)
new RuleTester().run('named', null as never, {
  valid: [{ code: 'import { foo } from "./export-all"' }].map((test) => ({
    filename,
    ...test,
  })),
  invalid: [
    {
      code: 'import { bar } from "./export-all"',
      errors: [{ message: "bar not found in './export-all'" }],
    },
  ].map((test) => ({ filename, ...test })),
});

// named [TypeScript]
new RuleTester().run('named', null as never, {
  valid: [
    { code: "import x from './typescript-export-assign-object'" },
    { code: 'import { MyType } from "./typescript"' },
    { code: 'import { Foo } from "./typescript"' },
    { code: 'import { Bar } from "./typescript"' },
    { code: 'import { getFoo } from "./typescript"' },
    { code: 'import { MyEnum } from "./typescript"' },
    {
      code: '\n              import { MyModule } from "./typescript"\n              MyModule.ModuleFunction()\n            ',
    },
    {
      code: '\n              import { MyNamespace } from "./typescript"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
    },
    { code: 'import { MyType } from "./typescript-declare"' },
    { code: 'import { Foo } from "./typescript-declare"' },
    { code: 'import { Bar } from "./typescript-declare"' },
    { code: 'import { getFoo } from "./typescript-declare"' },
    { code: 'import { MyEnum } from "./typescript-declare"' },
    {
      code: '\n              import { MyModule } from "./typescript-declare"\n              MyModule.ModuleFunction()\n            ',
    },
    {
      code: '\n              import { MyNamespace } from "./typescript-declare"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
    },
    { code: 'import { MyType } from "./typescript-export-assign-namespace"' },
    { code: 'import { Foo } from "./typescript-export-assign-namespace"' },
    { code: 'import { Bar } from "./typescript-export-assign-namespace"' },
    { code: 'import { getFoo } from "./typescript-export-assign-namespace"' },
    { code: 'import { MyEnum } from "./typescript-export-assign-namespace"' },
    {
      code: '\n              import { MyModule } from "./typescript-export-assign-namespace"\n              MyModule.ModuleFunction()\n            ',
    },
    {
      code: '\n              import { MyNamespace } from "./typescript-export-assign-namespace"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
    },
    {
      code: 'import { MyType } from "./typescript-export-assign-namespace-merged"',
    },
    {
      code: 'import { Foo } from "./typescript-export-assign-namespace-merged"',
    },
    {
      code: 'import { Bar } from "./typescript-export-assign-namespace-merged"',
    },
    {
      code: 'import { getFoo } from "./typescript-export-assign-namespace-merged"',
    },
    {
      code: 'import { MyEnum } from "./typescript-export-assign-namespace-merged"',
    },
    {
      code: '\n              import { MyModule } from "./typescript-export-assign-namespace-merged"\n              MyModule.ModuleFunction()\n            ',
    },
    {
      code: '\n              import { MyNamespace } from "./typescript-export-assign-namespace-merged"\n              MyNamespace.NSModule.NSModuleFunction()\n            ',
    },
  ].map((test) => ({ filename, ...test })),
  invalid: [
    {
      code: "import { NotExported } from './typescript-export-assign-object'",
      errors: [
        {
          message:
            "NotExported not found in './typescript-export-assign-object'",
        },
      ],
    },
    {
      code: "import { FooBar } from './typescript-export-assign-object'",
      errors: [
        { message: "FooBar not found in './typescript-export-assign-object'" },
      ],
    },
    {
      code: 'import { MissingType } from "./typescript"',
      errors: [{ message: "MissingType not found in './typescript'" }],
    },
    {
      code: 'import { NotExported } from "./typescript"',
      errors: [{ message: "NotExported not found in './typescript'" }],
    },
    {
      code: 'import { MissingType } from "./typescript-declare"',
      errors: [{ message: "MissingType not found in './typescript-declare'" }],
    },
    {
      code: 'import { NotExported } from "./typescript-declare"',
      errors: [{ message: "NotExported not found in './typescript-declare'" }],
    },
    {
      code: 'import { MissingType } from "./typescript-export-assign-namespace"',
      errors: [
        {
          message:
            "MissingType not found in './typescript-export-assign-namespace'",
        },
      ],
    },
    {
      code: 'import { NotExported } from "./typescript-export-assign-namespace"',
      errors: [
        {
          message:
            "NotExported not found in './typescript-export-assign-namespace'",
        },
      ],
    },
    {
      code: 'import { MissingType } from "./typescript-export-assign-namespace-merged"',
      errors: [
        {
          message:
            "MissingType not found in './typescript-export-assign-namespace-merged'",
        },
      ],
    },
    {
      code: 'import { NotExported } from "./typescript-export-assign-namespace-merged"',
      errors: [
        {
          message:
            "NotExported not found in './typescript-export-assign-namespace-merged'",
        },
      ],
    },
  ].map((test) => ({ filename, ...test })),
});

// documentation
new RuleTester().run('named', null as never, {
  valid: [
    { code: 'export const foo = "I\'m so foo"' },
    { code: "import { foo } from './foo'" },
    { code: "export { foo as bar } from './foo'" },
    // Match the docs' ignored CommonJS package even with React types installed.
    {
      code: "import { SomeNonsenseThatDoesntExist } from 'react'",
      settings: { 'import/ignore': ['node_modules'] },
    },
    {
      code: "import { notWhatever } from './whatever'",
      settings: { 'import/ignore': ['node_modules', '\\.coffee$'] },
    },
  ].map((test) => ({ filename, ...test })),
  invalid: [
    {
      code: "import { notFoo } from './foo'",
      errors: [{ message: "notFoo not found in './foo'" }],
    },
    {
      code: "export { notFoo as defNotBar } from './foo'",
      errors: [{ message: "notFoo not found in './foo'" }],
    },
    {
      code: "import { dontCreateStore } from './redux'",
      errors: [{ message: "dontCreateStore not found in './redux'" }],
    },
  ].map((test) => ({ filename, ...test })),
});
