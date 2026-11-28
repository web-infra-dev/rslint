// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['interface % {}'],
    options: {
      selector: 'interface',
    },
    locations: [
      {
        type: 'Interface',
        line: 1,
        column: 11,
        endLine: 1,
        endColumn: 12,
      },
    ],
  },
]);
