// Semantic mirror of eslint-plugin-unicorn v77.0.0 test/import-style.js.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const options = {
  checkExportFrom: true,
  styles: {
    unassigned: { unassigned: true, named: false },
    default: { default: true, named: false },
    namespace: { namespace: true, named: false },
    named: { named: true },
  },
};

const bannedOptions = [
  {
    styles: {
      banned: {
        unassigned: false,
        default: false,
        namespace: false,
        named: false,
      },
    },
    extendDefaultStyles: false,
  },
];

const typeOnlyOptions = [
  {
    extendDefaultStyles: false,
    styles: { chalk: { named: true }, named: { default: true } },
  },
];

const withOptions = (code: string) => ({ code, options: [options] });
const styleError = (allowed: string, moduleName: string) => ({
  messageId: 'importStyle',
  message: `Use ${allowed} import for module \`${moduleName}\`.`,
});
const invalid = (
  code: string,
  allowed: string,
  moduleName: string,
  caseOptions: any[] = [options],
) => ({
  code,
  options: caseOptions,
  errors: [styleError(allowed, moduleName)],
});
const banned = (code: string, caseOptions: any[] = bannedOptions) => ({
  code,
  options: caseOptions,
  errors: [
    {
      messageId: 'importStyleBanned',
      message:
        `All import styles are disabled for module \`banned\`. ` +
        'Use the `no-restricted-imports` rule to disallow a module.',
    },
  ],
});

const valid = [
  ...[
    `require('unassigned')`,
    `const {} = require('unassigned')`,
    `import 'unassigned'`,
    `import {} from 'unassigned'`,
    `import('unassigned')`,
    `export {} from 'unassigned'`,
    `const x = require('default')`,
    `const {default: x} = require('default')`,
    `const [] = require("default")`,
    `import x from 'default'`,
    `async () => { const {default: x} = await import('default'); }`,
    `export {default} from 'default'`,
    `const x = require('namespace')`,
    `const [] = require("namespace")`,
    `import * as x from 'namespace'`,
    `async () => { const x = await import('namespace'); }`,
    `export * from 'namespace'`,
    `const {x} = require('named')`,
    `const {...rest} = require("named")`,
    `const {x: y} = require('named')`,
    `import {x} from 'named'`,
    `import {x as y} from 'named'`,
    `async () => { const {x} = await import('named'); }`,
    `async () => { const {x: y} = await import('named'); }`,
    `export {x} from 'named'`,
    `export {x as y} from 'named'`,
    `const foo = 1; export {foo}`,
    `export const foo = 1;`,
    `export function foo() {}`,
  ].map(withOptions),
  { code: `import {inspect} from 'util'`, options: [] },
  { code: `import {inspect} from 'node:util'`, options: [] },
  { code: `const {inspect} = require('util')`, options: [] },
  { code: `const {inspect} = require('node:util')`, options: [] },
  { code: `import chalk from 'chalk'`, options: [] },
  { code: `import {default as chalk} from 'chalk'`, options: [] },
  { code: `export {promisify, callbackify} from 'util'`, options: [] },
  { code: `export {promisify, callbackify} from 'node:util'`, options: [] },
  {
    code: `require('chalk')`,
    options: [{ styles: {}, extendDefaultStyles: false }],
  },
  {
    code: `import {join} from 'path'`,
    options: [{ styles: { path: { named: true } } }],
  },
  { code: `import 'chalk'`, options: [{ checkImport: false }] },
  {
    code: `async () => { const {red} = await import('chalk'); }`,
    options: [{ checkDynamicImport: false }],
  },
  { code: `import('chalk')`, options: [{ checkDynamicImport: false }] },
  { code: `require('chalk')`, options: [{ checkRequire: false }] },
  {
    code: `const {red} = require('chalk')`,
    options: [{ checkRequire: false }],
  },
  {
    code: `import util, {inspect} from 'named-or-default'`,
    options: [
      { styles: { 'named-or-default': { named: true, default: true } } },
    ],
  },
  ...[
    `import fs from 'node:fs'`,
    `import * as fs from 'node:fs'`,
    `import {readFile} from 'node:fs'`,
    `import fsPromises from 'node:fs/promises'`,
    `import path from 'node:path'`,
    `import {inspect} from 'node:util'`,
    `async () => { const {inspect} = await import('node:util'); }`,
    `import fs from 'fs'`,
    `import unknown from 'node:unknown'`,
    `const fs = require('node:fs')`,
    `import('node:unknown')`,
  ].map(withOptions),
  {
    code: `import * as fs from 'node:fs'`,
    options: [{ styles: { fs: { namespace: true } } }],
  },
  ...[
    `require(1, 2, 3)`,
    `require(variable)`,
    `const x = require(variable)`,
    `const x = require('unassigned').x`,
    `async () => { const {red} = await import(variable); }`,
  ].map(withOptions),
  {
    code: `
      import util from "node:util";
      import * as util2 from "node:util";
      import {foo} from "node:util";
    `,
    options: [{ styles: { util: false } }],
  },
  {
    code: `
      import util from "node:util";
      import * as util2 from "node:util";
      import {foo} from "node:util";
    `,
    options: [{ styles: { util: { named: false } } }],
  },
  { code: `import type chalk from 'chalk'`, options: typeOnlyOptions },
  { code: `import type {x} from 'named'`, options: typeOnlyOptions },
  { code: `import type {ChalkInstance} from 'chalk'`, options: [] },
  { code: `import {type ChalkInstance} from 'chalk'`, options: [] },
  {
    code: `import chalk, {type ChalkInstance} from 'chalk'`,
    options: [],
  },
  { code: `let a` },
  ...[
    `require();`,
    `const a = require();`,
    `const {a} = require();`,
    `const [a] = require();`,
    `export const a = require();`,
  ].map((code) => ({ code })),
];

const invalidCases = [
  ...[
    `const {x} = require('unassigned')`,
    `const {default: x} = require('unassigned')`,
    `import x from 'unassigned'`,
    `async () => { const {default: x} = await import('unassigned'); }`,
    `const x = require('unassigned')`,
    `import * as x from 'unassigned'`,
    `async () => { const x = await import('unassigned'); }`,
    `const {x: y} = require('unassigned')`,
    `import {x} from 'unassigned'`,
    `import {x as y} from 'unassigned'`,
    `async () => { const {x} = await import('unassigned'); }`,
    `async () => { const {x: y} = await import('unassigned'); }`,
    `const {...rest} = require("unassigned")`,
    `const [] = require("unassigned")`,
    `export * from 'unassigned'`,
    `export {x} from 'unassigned'`,
    `export {x as y} from 'unassigned'`,
    `export {default} from 'unassigned'`,
  ].map((code) => invalid(code, 'unassigned', 'unassigned')),
  ...[
    `require('default')`,
    `const {} = require('default')`,
    `const {...rest} = require("default")`,
    `import 'default'`,
    `import {} from 'default'`,
    `import('default')`,
    `import * as x from 'default'`,
    `async () => { const x = await import('default'); }`,
    `const {x} = require('default')`,
    `const {x: y} = require('default')`,
    `import {x} from 'default'`,
    `import {x as y} from 'default'`,
    `async () => { const {x} = await import('default'); }`,
    `async () => { const {x: y} = await import('default'); }`,
    `export * from 'default'`,
    `export {x} from 'default'`,
    `export {x as y} from 'default'`,
  ].map((code) => invalid(code, 'default', 'default')),
  ...[
    `require('namespace')`,
    `const {} = require('namespace')`,
    `import 'namespace'`,
    `import {} from 'namespace'`,
    `import('namespace')`,
    `const {default: x} = require('namespace')`,
    `const {...rest} = require("namespace")`,
    `import x from 'namespace'`,
    `const {x} = require('namespace')`,
    `const {x: y} = require('namespace')`,
    `import {x} from 'namespace'`,
    `import {x as y} from 'namespace'`,
    `async () => { const {x} = await import('namespace'); }`,
    `async () => { const {x: y} = await import('namespace'); }`,
    `export {x} from 'namespace'`,
    `export {x as y} from 'namespace'`,
    `export {default} from 'namespace'`,
  ].map((code) => invalid(code, 'namespace', 'namespace')),
  ...[
    `require('named')`,
    `const {} = require('named')`,
    `const [] = require("named")`,
    `import 'named'`,
    `import {} from 'named'`,
    `import('named')`,
    `const x = require('named')`,
    `const {default: x} = require('named')`,
    `import x from 'named'`,
    `async () => { const {default: x} = await import('named'); }`,
    `async () => { const [x] = await import('named'); }`,
    `import * as x from 'named'`,
    `async () => { const x = await import('named'); }`,
    `export * from 'named'`,
    `export {default} from 'named'`,
    `import util, {inspect} from 'named'`,
  ].map((code) => invalid(code, 'named', 'named')),
  invalid(`import util, {inspect} from 'default'`, 'default', 'default'),
  invalid(`import * as path from 'node:path'`, 'default', 'node:path', []),
  invalid(`import util from 'node:util'`, 'named', 'node:util', []),
  invalid(`import * as util from 'node:util'`, 'named', 'node:util', []),
  invalid(`import('node:util')`, 'named', 'node:util', []),
  invalid(
    `async () => { const util = await import('node:util'); }`,
    'named',
    'node:util',
    [],
  ),
  invalid(`export * from 'node:util'`, 'named', 'node:util'),
  invalid(`import * as fs from 'node:fs'`, 'default', 'node:fs', [
    { styles: { fs: { default: true } } },
  ]),
  ...[
    [`import util from 'util'`, 'named', 'util'],
    [`import * as util from 'util'`, 'named', 'util'],
    [`import util from 'node:util'`, 'named', 'node:util'],
    [`const util = require('util')`, 'named', 'util'],
    [`const util = require('node:util')`, 'named', 'node:util'],
    [`require('util')`, 'named', 'util'],
    [`require('node:util')`, 'named', 'node:util'],
    [`require('ut' + 'il')`, 'named', 'util'],
    [`require('node:' + 'util')`, 'named', 'node:util'],
    [`import {red} from 'chalk'`, 'default', 'chalk'],
    [`import {red as green} from 'chalk'`, 'default', 'chalk'],
    [
      `async () => { const {red} = await import('chalk'); }`,
      'default',
      'chalk',
    ],
  ].map(([code, allowed, moduleName]) =>
    invalid(code, allowed, moduleName, []),
  ),
  invalid(
    `require('no-unassigned')`,
    'named, namespace, or default',
    'no-unassigned',
    [
      {
        styles: {
          'no-unassigned': { named: true, namespace: true, default: true },
        },
      },
    ],
  ),
  invalid(
    `import * as util from "node:util";`,
    'named or default',
    'node:util',
    [{ styles: { util: { default: true } } }],
  ),
  invalid(`import {promisify} from "node:util";`, 'default', 'node:util', [
    { styles: { util: { default: true, named: false } } },
  ]),
  {
    code: `import {type ChalkInstance, red} from 'chalk'`,
    options: [],
    errors: [styleError('default', 'chalk')],
  },
  banned(`import type {Foo} from 'banned'`),
  banned(`import {type Foo} from 'banned'`),
  // Upstream snapshot group (duplicates intentionally retained).
  ...[
    [`import util from 'util'`, 'named', 'util'],
    [`import * as util from 'util'`, 'named', 'util'],
    [`import util from 'node:util'`, 'named', 'node:util'],
    [`const util = require('util')`, 'named', 'util'],
    [`const util = require('node:util')`, 'named', 'node:util'],
    [`require('util')`, 'named', 'util'],
    [`require('node:util')`, 'named', 'node:util'],
    [`import {red} from 'chalk'`, 'default', 'chalk'],
    [`import {red as green} from 'chalk'`, 'default', 'chalk'],
    [
      `async () => { const {red} = await import('chalk'); }`,
      'default',
      'chalk',
    ],
  ].map(([code, allowed, moduleName]) =>
    invalid(code, allowed, moduleName, []),
  ),
  ...[
    `import 'banned'`,
    `import foo from 'banned'`,
    `import * as foo from 'banned'`,
    `import {foo} from 'banned'`,
    `async () => { const foo = await import('banned'); }`,
    `import('banned')`,
    `const foo = require('banned')`,
    `require('banned')`,
  ].map((code) => banned(code)),
  banned(`export {foo} from 'banned'`, [
    { checkExportFrom: true, ...bannedOptions[0] },
  ]),
  banned(`export * from 'banned'`, [
    { checkExportFrom: true, ...bannedOptions[0] },
  ]),
  banned(`import 'banned'`, [
    {
      styles: {
        banned: {
          unassigned: false,
          default: false,
          namespace: false,
          named: false,
        },
      },
    },
  ]),
  invalid(`require("chalk");`, 'default', 'chalk', []),
  invalid(`const {red} = require("chalk");`, 'default', 'chalk', []),
];

ruleTester.run('import-style', null as never, {
  valid,
  invalid: invalidCases,
});
