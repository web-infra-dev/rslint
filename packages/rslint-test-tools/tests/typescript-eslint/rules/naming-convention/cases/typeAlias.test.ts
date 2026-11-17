// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['type % = {};', 'type % = 1;'],
    options: {
      selector: 'typeAlias',
    },
    locations: [
      {
        type: 'Type Alias',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      },
      {
        type: 'Type Alias',
        line: 2,
        column: 6,
        endLine: 2,
        endColumn: 7,
      },
    ],
  },
]);
