import { checkSkip } from '../../../../rslint-rule-tester/src/cases.ts';
import { test } from 'rstack/test';
// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, resolveFixturePaths } from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/no_extraneous_dependencies/testdata/modules.txtar',
  ),
).run('no-extraneous-dependencies', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "@org/package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 1',
      code: 'import foo, { bar } from "@org/package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 2',
      code: 'require("@org/package")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 3',
      code: 'var foo = require("@org/package")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 4',
      code: 'export { foo } from "@org/package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 5',
      code: 'export * from "@org/package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 6',
      code: 'import "esm-package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 7',
      code: 'import foo, { bar } from "esm-package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 8',
      code: 'require("esm-package")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 9',
      code: 'var foo = require("esm-package")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 10',
      code: 'export { foo } from "esm-package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 11',
      code: 'export * from "esm-package"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 12',
      code: 'import "jquery"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 13',
      code: 'import foo, { bar } from "jquery"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 14',
      code: 'require("jquery")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 15',
      code: 'var foo = require("jquery")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 16',
      code: 'export { foo } from "jquery"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 17',
      code: 'export * from "jquery"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 18',
      code: 'import "lodash.cond"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 19',
      code: 'import foo, { bar } from "lodash.cond"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 20',
      code: 'require("lodash.cond")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 21',
      code: 'var foo = require("lodash.cond")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 22',
      code: 'export { foo } from "lodash.cond"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 23',
      code: 'export * from "lodash.cond"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 24',
      code: 'import "find-up"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 25',
      code: 'import foo, { bar } from "find-up"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 26',
      code: 'require("find-up")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 27',
      code: 'var foo = require("find-up")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 28',
      code: 'export { foo } from "find-up"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 29',
      code: 'export * from "find-up"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 30',
      code: 'import "rxjs"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 31',
      code: 'import foo, { bar } from "rxjs"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 32',
      code: 'require("rxjs")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 33',
      code: 'var foo = require("rxjs")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 34',
      code: 'export { foo } from "rxjs"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 35',
      code: 'export * from "rxjs"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 36',
      code: 'import "glob"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 37',
      code: 'import foo, { bar } from "glob"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 38',
      code: 'require("glob")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 39',
      code: 'var foo = require("glob")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 40',
      code: 'export { foo } from "glob"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 41',
      code: 'export * from "glob"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 42',
      code: 'import "eslint"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 43',
      code: 'import foo, { bar } from "eslint"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 44',
      code: 'require("eslint")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 45',
      code: 'var foo = require("eslint")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 46',
      code: 'export { foo } from "eslint"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 47',
      code: 'export * from "eslint"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 48',
      code: 'import "eslint/lib/api"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 49',
      code: 'import "fs"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 50',
      code: 'import "./foo"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 51',
      code: 'import "electron"',
      filename: 'tests/files/foo.js',
      settings: {
        'import/core-modules': ['electron'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 52',
      code: 'import "eslint"',
      filename: 'tests/files/foo.js',
      options: [
        {
          peerDependencies: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 53',
      code: 'import "importType"',
      filename: 'tests/files/foo.js',
      settings: {
        'import/resolver': {
          node: {
            paths: ['./tests/files'],
          },
        },
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 54',
      code: 'import chai from "chai"',
      filename: 'foo.spec.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.spec.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 55',
      code: 'import chai from "chai"',
      filename: 'foo.spec.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.spec.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 56',
      code: 'import chai from "chai"',
      filename: 'foo.spec.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.test.js', '<fixture>/*.spec.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 57',
      code: 'require(6)',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 58',
      code: 'import "doctrine"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 59',
      code: 'import type MyType from "myflowtyped";',
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/with-flow-typed',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 60',
      code: "\n        // @flow\n        import typeof TypeScriptModule from 'typescript';\n      ",
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/with-flow-typed',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'run 0 valid 61',
      code: 'import react from "react";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/monorepo/packages/nested-package',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 62',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/monorepo/packages/nested-package',
                '<fixture>/tests/files/monorepo',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 63',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 64',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/empty-folder',
                '<fixture>/tests/files/monorepo',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 65',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/monorepo',
                '<fixture>/tests/files/empty-folder',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 66',
      code: 'import react from "react";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/monorepo',
                '<fixture>/tests/files/monorepo/packages/nested-package',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 67',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/monorepo',
                '<fixture>/tests/files/monorepo/packages/nested-package',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 68',
      code: 'import rightpad from "right-pad";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: [
                '<fixture>/tests/files/monorepo',
                '<fixture>/tests/files/monorepo/packages/nested-package',
              ],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 69',
      code: 'import foo from "@generated/foo"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 70',
      code: 'import foo from "@generated/foo"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/bundled-dependencies/as-array-bundle-deps',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 71',
      code: 'import foo from "@generated/foo"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/bundled-dependencies/as-object',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 72',
      code: 'import foo from "@generated/foo"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/bundled-dependencies/race-condition',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 73',
      code: 'export function getToken() {}',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 74',
      code: 'export class Component extends React.Component {}',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 75',
      code: 'export function Component() {}',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 76',
      code: 'export const Component = () => {}',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 77',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/monorepo/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      settings: {
        'import/core-modules': ['not-a-dependency'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 78',
      code: 'import "@generated/bar/module"',
      filename: 'tests/files/foo.js',
      settings: {
        'import/core-modules': ['@generated/bar'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 79',
      code: 'import "@generated/bar/and/sub/path"',
      filename: 'tests/files/foo.js',
      settings: {
        'import/core-modules': ['@generated/bar'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 80',
      code: 'import "rxjs/operators"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 81',
      code: 'import "esm-package/esm-module";',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 82',
      code: '\n        import "alias/esm-package/esm-module";\n        import \'expose-loader?exposes[]=$&exposes[]=jQuery!jquery\';\n      ',
      filename: 'tests/files/foo.js',
      settings: {
        'import/resolver': 'webpack',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      skip: 'The case requires a webpack JavaScript resolver for its configured aliases or loader requests; the native Import runtime supports Node and TypeScript resolvers.',
    },
    {
      name: 'run 0 valid 83',
      skip: 'The case requires a webpack JavaScript resolver for its configured aliases or loader requests; the native Import runtime supports Node and TypeScript resolvers.',
      create(root: string) {
        const code = 'import "@custom-internal-alias/api/service";';
        return {
          code: code,
          filename: 'tests/files/foo.js',
          settings: resolveFixturePaths(
            {
              'import/resolver': {
                webpack: {
                  config: {
                    resolve: {
                      alias: {
                        '@custom-internal-alias':
                          '<fixture>/tests/files/internal-modules',
                      },
                    },
                  },
                },
              },
            },
            root,
          ),
          languageOptions: {
            parserOptions: {
              project: false,
              projectService: false,
            },
          },
        };
      },
    },
    {
      name: 'run 1 valid 0',
      code: 'import type T from "a";',
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/with-typescript-dev-dependencies',
              devDependencies: false,
            },
          ],
          root,
        ),
      settings: {
        'import/parsers': {
          typescript: ['.ts'],
        },
        'import/resolver': ['node', 'typescript'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 1 valid 1',
      code: 'import type { T } from "a"; export type { T };',
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/with-typescript-dev-dependencies',
              devDependencies: false,
            },
          ],
          root,
        ),
      settings: {
        'import/parsers': {
          typescript: ['.ts'],
        },
        'import/resolver': ['node', 'typescript'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 1 valid 2',
      code: 'export type { T } from "a";',
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/with-typescript-dev-dependencies',
              devDependencies: false,
            },
          ],
          root,
        ),
      settings: {
        'import/parsers': {
          typescript: ['.ts'],
        },
        'import/resolver': ['node', 'typescript'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 2 valid 0',
      code: 'import type MyType from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 2 valid 1',
      code: 'import type { MyType } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 2 valid 2',
      code: 'import { type MyType } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 2 valid 3',
      code: 'import { type MyType, type OtherType } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 0',
      code: "var path = require('path');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 1',
      code: "var foo = require('./foo');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 2',
      code: "import test from 'ava';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 3',
      code: "import find from 'lodash.find';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 4',
      code: "import isArray from 'lodash.isarray';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 5',
      code: 'import foo from \'"@generated/foo"\';',
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 6',
      code: "import type { MyType } from 'foo';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 7',
      code: "import react from 'react';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 8',
      code: "import foo from './foo';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 9',
      code: "import react from 'react';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 10',
      code: "import foo from './foo';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              includeInternal: true,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 11',
      code: "var foo = require('./foo');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              includeInternal: true,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 12',
      code: 'import foo from \'"@generated/foo"\';',
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              bundledDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 3 valid 13',
      code: 'var foo = require(\'"@generated/foo"\');',
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              bundledDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/monorepo/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/monorepo/packages/nested-package/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'var donthaveit = require("@org/not-a-dependency")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'@org/not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S @org/not-a-dependency' to add it",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'var donthaveit = require("@org/not-a-dependency/foo")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'@org/not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S @org/not-a-dependency' to add it",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import "eslint"',
      filename: 'tests/files/foo.js',
      options: [
        {
          devDependencies: false,
          peerDependencies: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'eslint' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import "lodash.isarray"',
      filename: 'tests/files/foo.js',
      options: [
        {
          optionalDependencies: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash.isarray' should be listed in the project's dependencies, not optionalDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'var foo = require("not-a-dependency")',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'var glob = require("glob")',
      filename: 'tests/files/foo.js',
      options: [
        {
          devDependencies: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'glob' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'import chai from "chai"',
      filename: 'foo.tes.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.test.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'chai' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'import chai from "chai"',
      filename: 'foo.tes.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.test.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'chai' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'import chai from "chai"',
      filename: 'foo.tes.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.test.js', '<fixture>/*.spec.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'chai' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'import chai from "chai"',
      filename: 'foo.tes.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              devDependencies: ['<fixture>/*.test.js', '<fixture>/*.spec.js'],
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'chai' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'var eslint = require("lodash.isarray")',
      filename: 'tests/files/foo.js',
      options: [
        {
          optionalDependencies: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash.isarray' should be listed in the project's dependencies, not optionalDependencies.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'import "bar"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/src/rules/doesn-exist/',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'The package.json file could not be found.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'import foo from "foo"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/with-syntax-error',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "The package.json file could not be parsed: invalid character '{' looking for beginning of object key string",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'import leftpad from "left-pad";',
      filename: 'tests/files/monorepo/packages/nested-package/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/monorepo/packages/nested-package',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'left-pad' should be listed in the project's dependencies. Run 'npm i -S left-pad' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'import react from "react";',
      filename: 'tests/files/monorepo/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'react' should be listed in the project's dependencies. Run 'npm i -S react' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'import react from "react";',
      filename: 'tests/files/monorepo/packages/nested-package/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/monorepo',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'react' should be listed in the project's dependencies. Run 'npm i -S react' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: 'import "react";',
      filename: 'tests/files/empty/index.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/tests/files/empty',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'react' should be listed in the project's dependencies. Run 'npm i -S react' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: 'import bar from "@generated/bar"',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'@generated/bar' should be listed in the project's dependencies. Run 'npm i -S @generated/bar' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: 'import foo from "@generated/foo"',
      filename: 'tests/files/foo.js',
      options: [
        {
          bundledDependencies: false,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'@generated/foo' should be listed in the project's dependencies. Run 'npm i -S @generated/foo' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 24',
      code: 'import bar from "@generated/bar"',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/bundled-dependencies/race-condition',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'@generated/bar' should be listed in the project's dependencies. Run 'npm i -S @generated/bar' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 25',
      code: 'export { foo } from "not-a-dependency";',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 26',
      code: 'export * from "not-a-dependency";',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'run 0 invalid 28',
      code: 'import "esm-package-not-in-pkg-json/esm-module";',
      filename: 'tests/files/foo.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'esm-package-not-in-pkg-json' should be listed in the project's dependencies. Run 'npm i -S esm-package-not-in-pkg-json' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      name: 'run 0 invalid 29',
      code: 'import "not-a-dependency"',
      filename: 'tests/files/foo.js',
      options: [
        {
          includeInternal: true,
        },
      ],
      settings: {
        'import/resolver': {
          node: {
            paths: ['./tests/files'],
          },
        },
        'import/internal-regex': '^not-a-dependency.*',
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 1 invalid 0',
      code: 'import T from "a";',
      filename: 'tests/files/foo.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/with-typescript-dev-dependencies',
              devDependencies: false,
            },
          ],
          root,
        ),
      settings: {
        'import/parsers': {
          typescript: ['.ts'],
        },
        'import/resolver': ['node', 'typescript'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'a' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      name: 'run 1 invalid 1',
      code: 'import type T from "a";',
      filename: 'tests/files/foo.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir:
                '<fixture>/tests/files/with-typescript-dev-dependencies',
              devDependencies: false,
              includeTypes: true,
            },
          ],
          root,
        ),
      settings: {
        'import/parsers': {
          typescript: ['.ts'],
        },
        'import/resolver': ['node', 'typescript'],
      },
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'a' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 2 invalid 0',
      code: 'import type { MyType } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      options: [
        {
          includeTypes: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      name: 'run 2 invalid 1',
      code: "import type { Foo } from 'not-a-dependency';",
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      options: [
        {
          includeTypes: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 2 invalid 2',
      code: 'import Foo, { type MyType } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'run 2 invalid 3',
      code: 'import { type MyType, Foo } from "not-a-dependency";',
      filename: 'tests/files/no-unused-modules/typescript/file-ts-a.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'not-a-dependency' should be listed in the project's dependencies. Run 'npm i -S not-a-dependency' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      name: 'run 3 invalid 0',
      code: "var _ = require('lodash');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash' should be listed in the project's dependencies. Run 'npm i -S lodash' to add it",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 3 invalid 1',
      code: "import _ from 'lodash';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash' should be listed in the project's dependencies. Run 'npm i -S lodash' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 3 invalid 2',
      code: "import test from 'ava';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              devDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'ava' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 3 invalid 3',
      code: "var test = require('ava');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              devDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'ava' should be listed in the project's dependencies, not devDependencies.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 3 invalid 4',
      code: "import isArray from 'lodash.isarray';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              optionalDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash.isarray' should be listed in the project's dependencies, not optionalDependencies.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 3 invalid 5',
      code: "var isArray = require('lodash.isarray');",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              optionalDependencies: false,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'lodash.isarray' should be listed in the project's dependencies, not optionalDependencies.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 3 invalid 6',
      code: "import type { MyType } from 'foo';",
      filename: 'docs/example.ts',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              packageDir: '<fixture>/docs',
              includeTypes: true,
            },
          ],
          root,
        ),
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "'foo' should be listed in the project's dependencies. Run 'npm i -S foo' to add it",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
  ],
});

// Preserved upstream inputs/assertions, not executable native expectations.
const skippedInvalidCases = [
  {
    name: 'run 0 invalid 27',
    code: 'import chai from "alias/chai";',
    filename: 'tests/files/foo.js',
    settings: {
      'import/resolver': 'webpack',
    },
    languageOptions: {
      parserOptions: {
        project: false,
        projectService: false,
      },
    },
    skip: 'The case requires a webpack JavaScript resolver for its configured aliases or loader requests; the native Import runtime supports Node and TypeScript resolvers.',
    errors: [
      {
        messageId: '',
        message:
          "'chai' should be listed in the project's dependencies. Run 'npm i -S chai' to add it",
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: 31,
      },
    ],
  },
];
for (const item of skippedInvalidCases) {
  checkSkip(item);
  test.skip(`invalid ${item.name}: ${item.skip}`, () => {});
}
