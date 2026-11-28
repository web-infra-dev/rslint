// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['function % () {}', '(function % () {});', 'declare function % ();'],
    options: {
      selector: 'function',
    },
    locations: [
      {
        type: 'Function',
        line: 1,
        column: 10,
        endLine: 1,
        endColumn: 11,
      },
      {
        type: 'Function',
        line: 2,
        column: 11,
        endLine: 2,
        endColumn: 12,
      },
      {
        type: 'Function',
        line: 3,
        column: 18,
        endLine: 3,
        endColumn: 19,
      },
    ],
  },
]);
