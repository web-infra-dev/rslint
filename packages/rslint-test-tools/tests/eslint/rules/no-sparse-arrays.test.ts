// Upstream: https://github.com/eslint/eslint/blob/v10.8.1/tests/lib/rules/no-sparse-arrays.js
/**
 * @fileoverview Disallow sparse arrays
 * @author Nicholas C. Zakas
 */

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

import { RuleTester } from '../rule-tester';

//------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------

const ruleTester = new RuleTester();

ruleTester.run('no-sparse-arrays', {
  valid: ['var a = [ 1, 2, ]'],

  invalid: [
    {
      code: 'var a = [,];',
      errors: [
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'var a = [ 1,, 2];',
      errors: [
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '[\r\n\t/* comment */,\n// comment\n ,];',
      errors: [
        {
          messageId: 'unexpectedSparseArray',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 16,
        },
        {
          messageId: 'unexpectedSparseArray',
          line: 4,
          column: 2,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: '[(( [a,] )),,,];',
      errors: [
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 14,
        },
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '[,(( [a,] )),,];',
      errors: [
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
        {
          messageId: 'unexpectedSparseArray',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
  ],
});
