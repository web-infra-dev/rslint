import { test, testFixturePath } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
// Rslint-Modified: we don't require rules
// const rule = require('rules/default')
const rule = null as never;
// Rslint-Modified end

ruleTester.run('default', rule, {
  valid: [
    test({
      code: 'import "./malformed.js"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./empty-folder";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import { foo } from "./default-export";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./default-export";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./mixed-exports";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import bar from "./default-export";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import CoolClass from "./default-class";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import bar, { baz } from "./default-export";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import crypto from "crypto";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import common from "./common";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export { default as bar } from "./bar"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export { default as bar, foo } from "./bar"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export {a} from "./named-exports"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import twofer from "./trampoline"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./named-default-export"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import connectedApp from "./redux"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import MyCoolComponent from "./jsx/MyCoolComponent.jsx"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import App from "./jsx/App"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: "import Foo from './jsx/FooES7.js';",
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import bar from "./default-export-from";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import bar from "./default-export-from-named";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: "import bar from './default-export-from-ignored.js';",
      filename: testFixturePath('./default-rule/consumer.ts'),
      settings: { 'import/ignore': ['common'] },
    }),
    test({
      code: 'export { "default" as bar } from "./bar"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-default"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-export-assign-default"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-export-assign-function"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-export-assign-mixed"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-export-assign-default-reexport"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foobar from "./typescript-export-assign-property"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'for (let { foo, bar } of baz) {}',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'const { x, y } = bar',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'const { x, y, ...z } = bar',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'let x; export { x }',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'let x; export { x as y }',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export const x = null',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export var x = null',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export let x = null',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export default x',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export default class x {}',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import json from "./data.json"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./foobar.json";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import foo from "./foobar";',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
    test({
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: testFixturePath('./default-rule/consumer.ts'),
    }),
  ],

  invalid: [
    test({
      code: 'import baz from "./named-exports";',
      filename: testFixturePath('./default-rule/consumer.ts'),
      errors: [
        {
          message:
            'No default export found in imported module "./named-exports".',
        },
      ],
    }),
    test({
      code: 'import twofer from "./broken-trampoline"',
      filename: testFixturePath('./default-rule/consumer.ts'),
      errors: [
        {
          message:
            'No default export found in imported module "./broken-trampoline".',
        },
      ],
    }),
    test({
      code: 'import barDefault from "./re-export"',
      filename: testFixturePath('./default-rule/consumer.ts'),
      errors: [
        {
          message: 'No default export found in imported module "./re-export".',
        },
      ],
    }),
    test({
      code: 'import foobar from "./typescript"',
      filename: testFixturePath('./default-rule/consumer.ts'),
      errors: [
        {
          message: 'No default export found in imported module "./typescript".',
        },
      ],
    }),
  ],
});
