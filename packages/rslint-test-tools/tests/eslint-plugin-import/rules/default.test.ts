// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('default', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "./malformed.js"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: 'import foo from "./empty-folder";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: 'import { foo } from "./default-export";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import foo from "./default-export";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: 'import foo from "./mixed-exports";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import bar from "./default-export";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 6',
      code: 'import CoolClass from "./default-class";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'import bar, { baz } from "./default-export";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: 'import crypto from "crypto";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: 'import common from "./common";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 10',
      code: 'export { default as bar } from "./bar"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: 'export { default as bar, foo } from "./bar"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'export {a} from "./named-exports"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'import twofer from "./trampoline"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'import foo from "./named-default-export"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'import connectedApp from "./redux"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 16',
      code: 'import MyCoolComponent from "./jsx/MyCoolComponent.jsx"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'import App from "./jsx/App"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 18',
      code: "import Foo from './jsx/FooES7.js';",
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 19',
      code: 'import bar from "./default-export-from";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 20',
      code: 'import bar from "./default-export-from-named";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 21',
      code: "import bar from './default-export-from-ignored.js';",
      filename: 'fixtures/default-rule/consumer.ts',
      settings: {
        'import/ignore': ['common'],
      },
    },
    {
      name: 'run 0 valid 22',
      code: 'export { "default" as bar } from "./bar"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 23',
      code: 'import foobar from "./typescript-default"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 24',
      code: 'import foobar from "./typescript-export-assign-default"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 25',
      code: 'import foobar from "./typescript-export-assign-function"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 26',
      code: 'import foobar from "./typescript-export-assign-mixed"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 27',
      code: 'import foobar from "./typescript-export-assign-default-reexport"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 28',
      code: 'import foobar from "./typescript-export-assign-property"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 29',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 30',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 31',
      code: 'const { x, y } = bar',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 32',
      code: 'const { x, y, ...z } = bar',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 33',
      code: 'let x; export { x }',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 34',
      code: 'let x; export { x as y }',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 35',
      code: 'export const x = null',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 36',
      code: 'export var x = null',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 37',
      code: 'export let x = null',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 38',
      code: 'export default x',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 39',
      code: 'export default class x {}',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 40',
      code: 'import json from "./data.json"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 41',
      code: 'import foo from "./foobar.json";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 42',
      code: 'import foo from "./foobar";',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 43',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 44',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 45',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'fixtures/default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 46',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'fixtures/default-rule/consumer.ts',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import baz from "./named-exports";',
      filename: 'fixtures/default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noDefault',
          message:
            'No default export found in imported module "./named-exports".',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import twofer from "./broken-trampoline"',
      filename: 'fixtures/default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noDefault',
          message:
            'No default export found in imported module "./broken-trampoline".',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import barDefault from "./re-export"',
      filename: 'fixtures/default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noDefault',
          message: 'No default export found in imported module "./re-export".',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import foobar from "./typescript"',
      filename: 'fixtures/default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noDefault',
          message: 'No default export found in imported module "./typescript".',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
  ],
});
