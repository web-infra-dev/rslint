// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('namespace', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import "./malformed.js"',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 1',
      code: "import * as foo from './empty-folder';",
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 2',
      code: 'import * as names from "./named-exports"; console.log((names.b).c); ',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 3',
      code: 'import * as names from "./named-exports"; console.log(names.a);',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 4',
      code: 'import * as foo from "./jsx/re-export"; console.log(foo.jsxFoo);',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 5',
      code: 'import * as components from "./jsx/component-exports"; console.log(components.Baz1);',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 6',
      code: 'import * as names from "./named-exports"; const { a } = names',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 7',
      code: 'import * as names from "./named-exports"; function b(names) { const { c } = names }',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 8',
      code: 'export * as names from "./named-exports"',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 9',
      code: "import * as names from './named-exports'; console.log(names['a']);",
      filename: 'fixtures/namespace-rule/consumer.ts',
      options: [
        {
          allowComputed: true,
        },
      ],
    },
    {
      name: 'run 0 valid 10',
      code: 'import * as Names from "./named-exports"; const Foo = <Names.a/>',
      filename: 'fixtures/namespace-rule/consumer.tsx',
    },
    {
      name: 'run 0 valid 11',
      code: 'import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.d.e.f)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 12',
      code: 'import * as names from "./default-export-string"; console.log(names.default)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 13',
      code: 'import * as names from "./default-export-namespace-string"; console.log(names.default)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 14',
      code: 'import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.d.e)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 15',
      code: 'import * as aliasChain from "./namespace-alias"; console.log(aliasChain.myName.b);',
      filename:
        'fixtures/namespace-rule/export-namespace-alias-chain/consumer.ts',
    },
    {
      name: 'run 0 valid 16',
      code: 'import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.e)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
    {
      name: 'run 0 valid 17',
      code: 'import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.e)',
      filename: 'fixtures/namespace-rule/consumer.ts',
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import * as names from './named-exports'; console.log(names.c)",
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'notFound',
          message: "'c' not found in imported namespace 'names'.",
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import * as names from './named-exports'; console.log(names['a']);",
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'computed',
          message:
            "Unable to validate computed reference to imported namespace 'names'.",
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import * as foo from './assignment-target'; foo.foo = 'y';",
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'assignment',
          message: "Assignment to member of namespace 'foo'.",
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import * as names from "./named-exports"; const { c } = names',
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'notFound',
          message: "'c' not found in imported namespace 'names'.",
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "console.log(names.c); import * as names from './named-exports';",
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'notFound',
          message: "'c' not found in imported namespace 'names'.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import * as ree from "./re-export"; console.log(ree.default)',
      filename: 'fixtures/namespace-rule/consumer.ts',
      errors: [
        {
          messageId: 'notFound',
          message: "'default' not found in imported namespace 'ree'.",
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import * as Names from "./named-exports"; const Foo = <Names.e/>',
      filename: 'fixtures/namespace-rule/consumer.tsx',
      errors: [
        {
          messageId: 'notFound',
          message: "'e' not found in imported namespace 'Names'.",
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
  ],
});
