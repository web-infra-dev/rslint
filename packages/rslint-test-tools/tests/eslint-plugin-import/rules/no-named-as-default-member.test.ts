// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-named-as-default-member', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import bar, {foo} from "./bar";',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: 'import bar from "./bar"; const baz = bar.baz',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: 'import {foo} from "./bar"; const baz = foo.baz;',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import * as named from "./named-exports"; const a = named.a',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: 'import foo from "./default-export-default-property"; const a = foo.default',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import bar, { foo } from "./export-default-string-and-named"',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 6',
      code: 'for (let { foo, bar } of baz) {}',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'for (let [ foo, bar ] of baz) {}',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: 'const { x, y } = bar',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: 'const { x, y, ...z } = bar',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 10',
      code: 'let x; export { x }',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 11',
      code: 'let x; export { x as y }',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'export const x = null',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'export var x = null',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'export let x = null',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'export default x',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 16',
      code: 'export default class x {}',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'import json from "./data.json"',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 18',
      code: 'import foo from "./foobar.json";',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 19',
      code: 'import foo from "./foobar";',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      settings: {
        'import/extensions': ['.js'],
      },
    },
    {
      name: 'run 0 valid 20',
      code: 'import { foo } from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 21',
      code: 'export * from "./issue-370-commonjs-namespace/bar"',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      settings: {
        'import/ignore': ['foo'],
      },
    },
    {
      name: 'run 0 valid 22',
      code: 'import * as a from "./commonjs-namespace/a"; a.b',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 23',
      code: 'import { foo } from "./ignore.invalid.extension"',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 24',
      code: "export default 'foo'; export const bar = 'baz';",
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 25',
      code: "import foo, {bar} from './foo.js';",
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import bar from "./bar"; const foo = bar.foo;',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import bar from "./bar"; bar.foo();',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import bar from "./bar"; const {foo} = bar;',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import bar from "./bar"; const {foo: foo2, baz} = bar;',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './bar'` instead.",
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import bar from "./export-default-string-and-named"; const foo = bar.foo;',
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `bar` also has a named export `foo`. Check if you meant to write `import {foo} from './export-default-string-and-named'` instead.",
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: "import foo from './foo.js';\nconst bar = foo.bar;",
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `foo` also has a named export `bar`. Check if you meant to write `import {bar} from './foo.js'` instead.",
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import foo from './foo.js';\nconst {bar} = foo;",
      filename: 'fixtures/no-named-as-default-member-rule/consumer.ts',
      errors: [
        {
          messageId: 'noNamedAsDefaultMember',
          message:
            "Caution: `foo` also has a named export `bar`. Check if you meant to write `import {bar} from './foo.js'` instead.",
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
  ],
});
