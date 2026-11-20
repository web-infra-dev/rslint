// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['enum % {}'],
    options: {
      selector: 'enum',
    },
    locations: [
      {
        type: 'Enum',
        line: 1,
        column: 6,
        endLine: 1,
        endColumn: 7,
      },
    ],
  },
]);
