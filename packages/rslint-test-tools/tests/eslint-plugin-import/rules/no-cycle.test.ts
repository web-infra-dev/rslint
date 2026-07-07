import { test, testFixturePath } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
// Rslint-Modified: we don't require rules
// const rule = require('rules/no-cycle')
const rule = null as never;
// Rslint-Modified end

const filename = testFixturePath('./no-cycle-rule/consumer.ts');
const rootExports =
  'export const rootValue = 1; export type RootType = string;';

const errorDetected = {
  message: 'Dependency cycle detected.',
};

const errorViaDepthOne = {
  message: 'Dependency cycle via ./depth-one:1',
};

const errorViaTwoOne = {
  message: 'Dependency cycle via ./depth-two:1=>./depth-one:1',
};

const valid = [
  test({
    code: `import { rootValue as other } from "./consumer"; ${rootExports}`,
    filename,
  }),
  test({
    code: `import foo from "./foo.js"; ${rootExports}`,
    filename,
  }),
  test({
    code: `import _ from "lodash"; ${rootExports}`,
    filename,
  }),
  test({
    code: `import foo from "@scope/foo"; ${rootExports}`,
    filename,
  }),
  test({
    code: 'import { noCycle } from "./no-cycle"; export const rootValue = noCycle; export type RootType = string;',
    filename,
  }),
  test({
    code: `var _ = require("lodash"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var find = require("lodash.find"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var foo = require("./foo"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var foo = require("../foo"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var foo = require("foo"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var foo = require("./"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var foo = require("@scope/foo"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var bar = require("./bar/index"); ${rootExports}`,
    filename,
  }),
  test({
    code: `var bar = require("./bar"); ${rootExports}`,
    filename,
  }),
  test({
    code: `const common = require("./commonjs-depth-one"); ${rootExports}`,
    filename,
  }),
  test({
    code: `define(["./amd-depth-one"], () => ({})); ${rootExports}`,
    filename,
  }),
  test({
    code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
    filename,
    options: [{ maxDepth: 1 }],
  }),
  test({
    code: 'import { depthOne, depthTwo } from "./depth-two"; export const rootValue = depthOne || depthTwo; export type RootType = string;',
    filename,
    options: [{ maxDepth: 1 }],
  }),
  test({
    code: `import("./depth-two").then(({ depthTwo }) => depthTwo); ${rootExports}`,
    filename,
    options: [{ maxDepth: 1 }],
  }),
  test({
    code: 'import type { RootType as LocalType } from "./type-only"; export const rootValue = 1; export type RootType = string;',
    filename,
  }),
  test({
    code: 'import type { RootType as LocalType, TypeOnly } from "./type-only"; export const rootValue = 1; export type RootType = string;',
    filename,
  }),
  test({
    code: 'import { typeOnly } from "./type-only"; export const rootValue = typeOnly; export type RootType = string;',
    filename,
  }),
  test({
    code: 'import { inlineTypeOnly } from "./inline-type-only"; export const rootValue = inlineTypeOnly; export type RootType = string;',
    filename,
  }),
  test({
    code: `function bar(){ return import("./depth-one"); } ${rootExports}`,
    filename,
    options: [{ allowUnsafeDynamicCyclicDependency: true }],
  }),
  test({
    code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;',
    filename,
    options: [{ allowUnsafeDynamicCyclicDependency: true }],
  }),
  test({
    code: 'import { loadRoot } from "./dynamic-depth-one"; export const rootValue = loadRoot; export type RootType = string;',
    filename,
    options: [{ allowUnsafeDynamicCyclicDependency: true }],
  }),
  test({
    code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
    filename,
    options: [{ esmodule: false }],
  }),
];

const invalid = [
  test({
    code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
    filename,
    errors: [errorDetected],
  }),
  test({
    code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
    filename,
    options: [{}],
    errors: [errorDetected],
  }),
  test({
    code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
    filename,
    options: [{ maxDepth: 1 }],
    errors: [errorDetected],
  }),
  test({
    code: `const { common } = require("./commonjs-depth-one"); ${rootExports}`,
    filename,
    options: [{ commonjs: true }],
    errors: [errorDetected],
  }),
  test({
    code: `require(["./amd-depth-one"], d1 => {}); ${rootExports}`,
    filename,
    options: [{ amd: true }],
    errors: [errorDetected],
  }),
  test({
    code: `define(["./amd-depth-one"], d1 => {}); ${rootExports}`,
    filename,
    options: [{ amd: true }],
    errors: [errorDetected],
  }),
  test({
    code: 'import { rootValue as reexported } from "./reexport-depth-one"; export const rootValue = reexported; export type RootType = string;',
    filename,
    errors: [errorDetected],
  }),
  test({
    code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
    filename,
    errors: [errorViaDepthOne],
  }),
  test({
    code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
    filename,
    options: [{ maxDepth: 2 }],
    errors: [errorViaDepthOne],
  }),
  test({
    code: 'const { depthTwo } = require("./depth-two"); export const rootValue = depthTwo; export type RootType = string;',
    filename,
    options: [{ commonjs: true }],
    errors: [errorViaDepthOne],
  }),
  test({
    code: 'import { two } from "./depth-three-star"; export const rootValue = two.depthTwo; export type RootType = string;',
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: 'import one, { two, three } from "./depth-three-star"; export const rootValue = one || two || three; export type RootType = string;',
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: 'import { depthThreeIndirect } from "./depth-three-indirect"; export const rootValue = depthThreeIndirect; export type RootType = string;',
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: 'import { depthThree } from "./depth-three"; export const rootValue = depthThree; export type RootType = string;',
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
    filename,
    options: [{ maxDepth: Infinity }],
    errors: [errorViaDepthOne],
  }),
  test({
    code: 'import { depthTwo } from "./depth-two"; export const rootValue = depthTwo; export type RootType = string;',
    filename,
    options: [{ maxDepth: '∞' }],
    errors: [errorViaDepthOne],
  }),
  test({
    code: `import("./depth-three-star"); ${rootExports}`,
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: `import("./depth-three-indirect"); ${rootExports}`,
    filename,
    errors: [errorViaTwoOne],
  }),
  test({
    code: `function bar(){ return import("./depth-one"); } ${rootExports}`,
    filename,
    errors: [errorDetected],
  }),
  test({
    code: 'import { dynamicDepthOne } from "./depth-one-dynamic"; export const rootValue = dynamicDepthOne; export type RootType = string;',
    filename,
    errors: [errorDetected],
  }),
  test({
    code: 'import { depthOne } from "./depth-one"; export const rootValue = depthOne; export type RootType = string;',
    filename,
    options: [{ allowUnsafeDynamicCyclicDependency: true }],
    errors: [errorDetected],
  }),
  test({
    code: `import "./reexport-type-only"; ${rootExports}`,
    filename,
    errors: [errorDetected],
  }),
];

function withDisableScc<
  T extends { code: string; options?: Record<string, unknown>[] },
>(cases: T[]): T[] {
  return cases.flatMap((item) => [
    item,
    {
      ...item,
      code: `${item.code} // disableScc=true`,
      options: [{ ...(item.options?.[0] ?? {}), disableScc: true }],
    },
  ]);
}

ruleTester.run('no-cycle', rule, {
  valid: withDisableScc(valid),
  invalid: withDisableScc(invalid),
});
