// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: ['enum Ignored { % }', 'enum Ignored { "%" }'],
    options: {
      selector: 'enumMember',
    },
    locations: [
      {
        type: 'Enum Member',
        line: 1,
        column: 16,
        endLine: 1,
        endColumn: 17,
      },
      {
        type: 'Enum Member',
        line: 2,
        column: 16,
        endLine: 2,
        endColumn: 19,
      },
    ],
  },
]);
