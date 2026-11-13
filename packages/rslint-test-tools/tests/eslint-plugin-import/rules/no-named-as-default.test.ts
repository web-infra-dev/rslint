// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-named-as-default', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import bar, { foo } from "./bar";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: 'import bar, { foo } from "./empty-folder";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: 'import bar, { foo } from "./export-default-string-and-named"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import something from "./no-named-as-default/re-exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: 'import { something } from "./no-named-as-default/re-exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import myOwnNameForVariable from "./no-named-as-default/exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 6',
      code: 'import { variable } from "./no-named-as-default/exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'import variable from "./no-named-as-default/misleading-re-exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: 'import foobar from "./no-named-as-default/no-default-export.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: 'export { something } from "./no-named-as-default/re-exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 10',
      code: 'export { variable } from "./no-named-as-default/exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'const { x, y } = bar',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'const { x, y, ...z } = bar',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'let x; export { x }',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 16',
      code: 'let x; export { x as y }',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'export const x = null',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 18',
      code: 'export var x = null',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 19',
      code: 'export let x = null',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 20',
      code: 'export default x',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 21',
      code: 'export default class x {}',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 22',
      code: 'import json from "./data.json"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 23',
      code: 'import foo from "./foobar.json";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 24',
      code: 'import foo from "./foobar";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 25',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 26',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 27',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 28',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 29',
      code: "export default 'foo'; export const bar = 'baz';",
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 30',
      code: "import foo from './foo.js';",
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import foo from "./bar";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'foo' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import foo, { foo as bar } from "./bar";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'foo' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import foo from "./export-default-string-and-named"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'foo' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import foo, { foo as bar } from "./export-default-string-and-named"',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'foo' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import something from "./no-named-as-default/misleading-re-exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'something' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import variable from "./no-named-as-default/exports.js";',
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'variable' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import bar from './foo.js';",
      filename: 'fixtures/no-named-as-default-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefault',
          message:
            "Using exported name 'bar' as identifier for default import.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
  ],
});
