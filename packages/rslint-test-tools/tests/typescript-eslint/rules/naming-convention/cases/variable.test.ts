// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'const % = 1;',
      'let % = 1;',
      'var % = 1;',
      'const {%} = {ignored: 1};',
      'const {% = 2} = {ignored: 1};',
      'const {...%} = {ignored: 1};',
      'const [%] = [1];',
      'const [% = 1] = [1];',
      'const [...%] = [1];',
    ],
    options: {
      selector: 'variable',
    },
    locations: [
      {
        type: 'Variable',
        line: 1,
        column: 7,
        endLine: 1,
        endColumn: 8,
      },
      {
        type: 'Variable',
        line: 2,
        column: 5,
        endLine: 2,
        endColumn: 6,
      },
      {
        type: 'Variable',
        line: 3,
        column: 5,
        endLine: 3,
        endColumn: 6,
      },
      {
        type: 'Variable',
        line: 4,
        column: 8,
        endLine: 4,
        endColumn: 9,
      },
      {
        type: 'Variable',
        line: 5,
        column: 8,
        endLine: 5,
        endColumn: 9,
      },
      {
        type: 'Variable',
        line: 6,
        column: 11,
        endLine: 6,
        endColumn: 12,
      },
      {
        type: 'Variable',
        line: 7,
        column: 8,
        endLine: 7,
        endColumn: 9,
      },
      {
        type: 'Variable',
        line: 8,
        column: 8,
        endLine: 8,
        endColumn: 9,
      },
      {
        type: 'Variable',
        line: 9,
        column: 11,
        endLine: 9,
        endColumn: 12,
      },
    ],
  },
]);
