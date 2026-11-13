// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-named-default', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import bar from "./bar";',
    },
    {
      name: 'run 0 valid 1',
      code: 'import bar, { foo } from "./bar";',
    },
    {
      name: 'run 0 valid 2',
      code: 'import { type default as Foo } from "./bar";',
    },
    {
      name: 'run 0 valid 3',
      code: 'for (let { foo, bar } of baz) {}',
    },
    {
      name: 'run 0 valid 4',
      code: 'for (let [ foo, bar ] of baz) {}',
    },
    {
      name: 'run 0 valid 5',
      code: 'const { x, y } = bar',
    },
    {
      name: 'run 0 valid 6',
      code: 'const { x, y, ...z } = bar',
    },
    {
      name: 'run 0 valid 7',
      code: 'let x; export { x }',
    },
    {
      name: 'run 0 valid 8',
      code: 'let x; export { x as y }',
    },
    {
      name: 'run 0 valid 9',
      code: 'export const x = null',
    },
    {
      name: 'run 0 valid 10',
      code: 'export var x = null',
    },
    {
      name: 'run 0 valid 11',
      code: 'export let x = null',
    },
    {
      name: 'run 0 valid 12',
      code: 'export default x',
    },
    {
      name: 'run 0 valid 13',
      code: 'export default class x {}',
    },
    {
      name: 'run 0 valid 14',
      code: 'import json from "./data.json"',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 15',
      code: 'import foo from "./foobar.json";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 16',
      code: 'import foo from "./foobar";',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 17',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 18',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 19',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
    },
    {
      name: 'run 0 valid 20',
      code: 'import { foo } from "./ignore.invalid.extension"',
    },
    {
      name: 'run 0 valid 21',
      code: "// foo.js\nexport default 'foo';\nexport const bar = 'baz';",
    },
    {
      name: 'run 0 valid 22',
      code: "import foo from './foo.js';",
    },
    {
      name: 'run 0 valid 23',
      code: "import foo, { bar } from './foo.js';",
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import { default as bar } from "./bar";',
      errors: [
        {
          messageId: '',
          message: "Use default import syntax to import 'bar'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import { foo, default as bar } from "./bar";',
      errors: [
        {
          messageId: '',
          message: "Use default import syntax to import 'bar'.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import { "default" as bar } from "./bar";',
      errors: [
        {
          messageId: '',
          message: "Use default import syntax to import 'bar'.",
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import { default as foo } from './foo.js';",
      errors: [
        {
          messageId: '',
          message: "Use default import syntax to import 'foo'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import { default as foo, bar } from './foo.js';",
      errors: [
        {
          messageId: '',
          message: "Use default import syntax to import 'foo'.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
  ],
});
