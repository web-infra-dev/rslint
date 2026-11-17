// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-cycle', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import { rootValue as other } from "./consumer"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: 'import { rootValue as other } from "./consumer"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 2',
      code: 'import foo from "./foo.js"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import foo from "./foo.js"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 4',
      code: 'import _ from "lodash"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import _ from "lodash"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 6',
      code: 'import foo from "@scope/foo"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'import foo from "@scope/foo"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 8',
      code: 'import { noCycle } from "./no-cycle"; export const rootValue = noCycle; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: 'import { noCycle } from "./no-cycle"; export const rootValue = noCycle; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: 'var _ = require("lodash"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: 'var _ = require("lodash"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 12',
      code: 'var find = require("lodash.find"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'var find = require("lodash.find"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 14',
      code: 'var foo = require("./foo"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'var foo = require("./foo"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 16',
      code: 'var foo = require("../foo"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'var foo = require("../foo"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 18',
      code: 'var foo = require("foo"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 19',
      code: 'var foo = require("foo"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 20',
      code: 'var foo = require("./"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 21',
      code: 'var foo = require("./"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 22',
      code: 'var foo = require("@scope/foo"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 23',
      code: 'var foo = require("@scope/foo"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 24',
      code: 'var bar = require("./bar/index"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 25',
      code: 'var bar = require("./bar/index"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 26',
      code: 'var bar = require("./bar"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 27',
      code: 'var bar = require("./bar"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 28',
      code: 'const common = require("./commonjs-depth-one"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 29',
      code: 'const common = require("./commonjs-depth-one"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 30',
      code: 'define(["./amd-depth-one"], () => ({})); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 31',
      code: 'define(["./amd-depth-one"], () => ({})); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 32',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
        },
      ],
    },
    {
      name: 'run 0 valid 33',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 34',
      code: 'import { depthOne, depthTwo } from "./depth-two"; export const rootValue = depthOne || depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
        },
      ],
    },
    {
      name: 'run 0 valid 35',
      code: 'import { depthOne, depthTwo } from "./depth-two"; export const rootValue = depthOne || depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 36',
      code: 'import("./depth-two").then(({ depthTwo }) => depthTwo); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
        },
      ],
    },
    {
      name: 'run 0 valid 37',
      code: 'import("./depth-two").then(({ depthTwo }) => depthTwo); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 38',
      code: 'import type { RootType as LocalType } from "./type-only"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 39',
      code: 'import type { RootType as LocalType } from "./type-only"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 40',
      code: 'import type { RootType as LocalType, TypeOnly } from "./type-only"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 41',
      code: 'import type { RootType as LocalType, TypeOnly } from "./type-only"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 42',
      code: 'import { typeOnly } from "./type-only"; export const rootValue = typeOnly; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 43',
      code: 'import { typeOnly } from "./type-only"; export const rootValue = typeOnly; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 44',
      code: 'import { inlineTypeOnly } from "./inline-type-only"; export const rootValue = inlineTypeOnly; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 45',
      code: 'import { inlineTypeOnly } from "./inline-type-only"; export const rootValue = inlineTypeOnly; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 46',
      code: 'function bar(){ return import("./depth-one"); } export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
        },
      ],
    },
    {
      name: 'run 0 valid 47',
      code: 'function bar(){ return import("./depth-one"); } export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 48',
      code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
        },
      ],
    },
    {
      name: 'run 0 valid 49',
      code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 50',
      code: 'import { loadRoot } from "./dynamic-depth-one"; export const rootValue = loadRoot; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
        },
      ],
    },
    {
      name: 'run 0 valid 51',
      code: 'import { loadRoot } from "./dynamic-depth-one"; export const rootValue = loadRoot; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
          disableScc: true,
        },
      ],
    },
    {
      name: 'run 0 valid 52',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          esmodule: false,
        },
      ],
    },
    {
      name: 'run 0 valid 53',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          esmodule: false,
          disableScc: true,
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [{}],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 1,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'const { common } = require("./commonjs-depth-one"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'const { common } = require("./commonjs-depth-one"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          commonjs: true,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'require(["./amd-depth-one"], d1 => {}); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          amd: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'require(["./amd-depth-one"], d1 => {}); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          amd: true,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'define(["./amd-depth-one"], d1 => {}); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          amd: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'define(["./amd-depth-one"], d1 => {}); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          amd: true,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'import { rootValue as reexported } from "./reexport-depth-one"; export const rootValue = reexported; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'import { rootValue as reexported } from "./reexport-depth-one"; export const rootValue = reexported; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 2,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: 2,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'const { depthTwo } = require("./depth-two"); export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          commonjs: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 19',
      code: 'const { depthTwo } = require("./depth-two"); export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          commonjs: true,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 20',
      code: 'import { two } from "./depth-three-star"; export const rootValue = two.depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'run 0 invalid 21',
      code: 'import { two } from "./depth-three-star"; export const rootValue = two.depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'run 0 invalid 22',
      code: 'import one, { two, three } from "./depth-three-star"; export const rootValue = one || two || three; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      name: 'run 0 invalid 23',
      code: 'import one, { two, three } from "./depth-three-star"; export const rootValue = one || two || three; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      name: 'run 0 invalid 24',
      code: 'import { depthThreeIndirect } from "./depth-three-indirect"; export const rootValue = depthThreeIndirect; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      name: 'run 0 invalid 25',
      code: 'import { depthThreeIndirect } from "./depth-three-indirect"; export const rootValue = depthThreeIndirect; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      name: 'run 0 invalid 26',
      code: 'import { depthThree } from "./depth-three"; export const rootValue = depthThree; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 27',
      code: 'import { depthThree } from "./depth-three"; export const rootValue = depthThree; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      name: 'run 0 invalid 28',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: null,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 29',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: null,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 30',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: '∞',
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 31',
      code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          maxDepth: '∞',
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 32',
      code: 'import("./depth-three-star"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 33',
      code: 'import("./depth-three-star"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 34',
      code: 'import("./depth-three-indirect"); export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 35',
      code: 'import("./depth-three-indirect"); export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 36',
      code: 'function bar(){ return import("./depth-one"); } export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 37',
      code: 'function bar(){ return import("./depth-one"); } export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 38',
      code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      name: 'run 0 invalid 39',
      code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      name: 'run 0 invalid 40',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 41',
      code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          allowUnsafeDynamicCyclicDependency: true,
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 42',
      code: 'import "./reexport-type-only"; export const rootValue = 1; export type RootType = string;',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 43',
      code: 'import "./reexport-type-only"; export const rootValue = 1; export type RootType = string; // disableScc=true',
      filename: 'fixtures/no-cycle-rule/consumer.ts',
      options: [
        {
          disableScc: true,
        },
      ],
      errors: [
        {
          messageId: 'cycle',
          message: 'Dependency cycle detected.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
  ],
});
