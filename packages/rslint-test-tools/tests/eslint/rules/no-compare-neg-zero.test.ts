// Upstream: https://github.com/eslint/eslint/blob/v9.39.1/tests/lib/rules/no-compare-neg-zero.js
/**
 * @fileoverview The rule should warn against code that tries to compare against -0.
 * @author Aladdin-ADD<hh_2013@foxmail.com>
 */

//------------------------------------------------------------------------------
// Requirements
//------------------------------------------------------------------------------

import { RuleTester } from '../rule-tester';

//------------------------------------------------------------------------------
// Tests
//------------------------------------------------------------------------------

const ruleTester = new RuleTester();

ruleTester.run('no-compare-neg-zero', {
  valid: [
    'x === 0',
    '0 === x',
    'x == 0',
    '0 == x',
    "x === '0'",
    "'0' === x",
    "x == '0'",
    "'0' == x",
    "x === '-0'",
    "'-0' === x",
    "x == '-0'",
    "'-0' == x",
    'x === -1',
    '-1 === x',
    'x < 0',
    '0 < x',
    'x <= 0',
    '0 <= x',
    'x > 0',
    '0 > x',
    'x >= 0',
    '0 >= x',
    'x != 0',
    '0 != x',
    'x !== 0',
    '0 !== x',
    'Object.is(x, -0)',
  ],

  invalid: [
    {
      code: 'x === -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '===' },
        },
      ],
    },
    {
      code: '-0 === x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '===' },
        },
      ],
    },
    {
      code: 'x == -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '==' },
        },
      ],
    },
    {
      code: '-0 == x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '==' },
        },
      ],
    },
    {
      code: 'x > -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '>' },
        },
      ],
    },
    {
      code: '-0 > x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '>' },
        },
      ],
    },
    {
      code: 'x >= -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '>=' },
        },
      ],
    },
    {
      code: '-0 >= x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '>=' },
        },
      ],
    },
    {
      code: 'x < -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '<' },
        },
      ],
    },
    {
      code: '-0 < x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '<' },
        },
      ],
    },
    {
      code: 'x <= -0',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '<=' },
        },
      ],
    },
    {
      code: '-0 <= x',
      errors: [
        {
          messageId: 'unexpected',
          data: { operator: '<=' },
        },
      ],
    },
  ],
});
