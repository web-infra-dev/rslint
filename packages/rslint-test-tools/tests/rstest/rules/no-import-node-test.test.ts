import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-import-node-test', {} as never, {
  valid: [
    { code: `import { test } from '@rstest/core'` },
    { code: `import { test } from 'vitest'` },
    { code: `import { test } from 'node:testing'` },
    { code: `const assert = require('node:assert')` },
  ],
  invalid: [
    {
      code: `import { test } from 'node:test'`,
      output: `import { test } from '@rstest/core'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
        },
      ],
    },
    {
      code: `import { test } from "node:test"`,
      output: `import { test } from "@rstest/core"`,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import { test, describe } from 'node:test'`,
      output: `import { test, describe } from '@rstest/core'`,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import { test as nodeTest } from 'node:test'`,
      output: `import { test as nodeTest } from '@rstest/core'`,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import nodeTest from 'node:test'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import * as nodeTest from 'node:test'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import 'node:test'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import { mock } from 'node:test'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import { test, mock } from 'node:test'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `import { tap } from 'node:test/reporters'`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
    {
      code: `const nodeTest = require('node:test')`,
      output: null,
      errors: [{ messageId: 'noImportNodeTest' }],
    },
  ],
});
