import { test, testFixturePath } from '../utils.js';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
// Rslint-Modified: we don't require rules
// const rule = require('rules/namespace')
const rule = null as never;
// Rslint-Modified end

const filename = testFixturePath('./namespace-rule/consumer.ts');
const tsxFilename = testFixturePath('./namespace-rule/consumer.tsx');

ruleTester.run('namespace', rule, {
  valid: [
    test({
      code: 'import "./malformed.js"',
      filename,
    }),
    test({
      code: "import * as foo from './empty-folder';",
      filename,
    }),
    test({
      code: 'import * as names from "./named-exports"; console.log((names.b).c); ',
      filename,
    }),
    test({
      code: 'import * as names from "./named-exports"; console.log(names.a);',
      filename,
    }),
    test({
      code: 'import * as foo from "./jsx/re-export"; console.log(foo.jsxFoo);',
      filename,
    }),
    test({
      code: 'import * as components from "./jsx/component-exports"; console.log(components.Baz1);',
      filename,
    }),
    test({
      code: 'import * as names from "./named-exports"; const { a } = names',
      filename,
    }),
    test({
      code: 'import * as names from "./named-exports"; function b(names) { const { c } = names }',
      filename,
    }),
    test({
      code: 'export * as names from "./named-exports"',
      filename,
    }),
    test({
      code: "import * as names from './named-exports'; console.log(names['a']);",
      filename,
      options: [{ allowComputed: true }],
    }),
    test({
      code: 'import * as Names from "./named-exports"; const Foo = <Names.a/>',
      filename: tsxFilename,
    }),
    test({
      code: 'import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.d.e.f)',
      filename,
    }),
    test({
      code: 'import * as names from "./default-export-string"; console.log(names.default)',
      filename,
    }),
    test({
      code: 'import * as names from "./default-export-namespace-string"; console.log(names.default)',
      filename,
    }),
    test({
      code: 'import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.d.e)',
      filename,
    }),
    test({
      code: 'import * as aliasChain from "./namespace-alias"; console.log(aliasChain.myName.b);',
      filename: testFixturePath(
        './namespace-rule/export-namespace-alias-chain/consumer.ts',
      ),
    }),
    test({
      code: 'import * as a from "./deep-namespace-chain/entry"; console.log(a.b.c.e)',
      filename,
    }),
    test({
      code: 'import { "b" as b } from "./deep-namespace-chain/entry"; console.log(b.c.e)',
      filename,
    }),
  ],

  invalid: [
    test({
      code: "import * as names from './named-exports'; console.log(names.c)",
      filename,
      errors: [{ message: "'c' not found in imported namespace 'names'." }],
    }),
    test({
      code: "import * as names from './named-exports'; console.log(names['a']);",
      filename,
      errors: [
        "Unable to validate computed reference to imported namespace 'names'.",
      ],
    }),
    test({
      code: "import * as foo from './assignment-target'; foo.foo = 'y';",
      filename,
      errors: [{ message: "Assignment to member of namespace 'foo'." }],
    }),
    test({
      code: 'import * as names from "./named-exports"; const { c } = names',
      filename,
      errors: [{ message: "'c' not found in imported namespace 'names'." }],
    }),
    test({
      code: "console.log(names.c); import * as names from './named-exports';",
      filename,
      errors: [{ message: "'c' not found in imported namespace 'names'." }],
    }),
    test({
      code: 'import * as ree from "./re-export"; console.log(ree.default)',
      filename,
      errors: [{ message: "'default' not found in imported namespace 'ree'." }],
    }),
    test({
      code: 'import * as Names from "./named-exports"; const Foo = <Names.e/>',
      filename: tsxFilename,
      errors: [{ message: "'e' not found in imported namespace 'Names'." }],
    }),
  ],
});
