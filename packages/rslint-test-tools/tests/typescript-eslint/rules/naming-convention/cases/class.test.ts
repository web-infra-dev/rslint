// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['class % {}', 'abstract class % {}', 'const ignored = class % {}'],
    options: {
      selector: 'class',
    },
    locations: [
      {
        type: 'Class',
        line: 1,
        column: 7,
        endLine: 1,
        endColumn: 8,
      },
      {
        type: 'Class',
        line: 2,
        column: 16,
        endLine: 2,
        endColumn: 17,
      },
      {
        type: 'Class',
        line: 3,
        column: 23,
        endLine: 3,
        endColumn: 24,
      },
    ],
  },
]);
