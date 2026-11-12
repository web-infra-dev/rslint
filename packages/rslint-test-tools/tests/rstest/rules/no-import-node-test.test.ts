import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const ruleTester = new RuleTester(config);

ruleTester.run('no-import-node-test', {
  valid: [
    { code: `import { test } from '@rstest/core'` },
    { code: `import { test } from 'vitest'` },
    { code: `import { test } from 'node:testing'` },
    { code: `const assert = require('node:assert')` },
    { code: `import assert = require('node:assert')` },
  ],
  invalid: [
    {
      code: `import { test } from 'node:test'`,
      output: `import { test } from '@rstest/core'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `import { test } from "node:test"`,
      output: `import { test } from "@rstest/core"`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `import { test, describe } from 'node:test'`,
      output: `import { test, describe } from '@rstest/core'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: `import { test as nodeTest } from 'node:test'`,
      output: `import { test as nodeTest } from '@rstest/core'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: `import nodeTest from 'node:test'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `import * as nodeTest from 'node:test'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: `import 'node:test'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `import { mock } from 'node:test'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: `import { test, mock } from 'node:test'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `import { tap } from 'node:test/reporters'`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: `const nodeTest = require('node:test')`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: `import nodeTest = require('node:test')`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `const nodeTest = (require)('node:test')`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `const nodeTest = require(('node:test'))`,
      errors: [
        {
          messageId: 'noImportNodeTest',
          message: 'Do not import the Node test runner in Rstest test files',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
  ],
});
