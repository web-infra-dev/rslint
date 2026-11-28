// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'class Ignored { constructor(private %) {} }',
      'class Ignored { constructor(readonly %) {} }',
      'class Ignored { constructor(private readonly %) {} }',
    ],
    options: {
      selector: 'parameterProperty',
    },
    locations: [
      {
        type: 'Parameter Property',
        line: 1,
        column: 37,
        endLine: 1,
        endColumn: 38,
      },
      {
        type: 'Parameter Property',
        line: 2,
        column: 38,
        endLine: 2,
        endColumn: 39,
      },
      {
        type: 'Parameter Property',
        line: 3,
        column: 46,
        endLine: 3,
        endColumn: 47,
      },
    ],
  },
  {
    code: ['class Ignored { constructor(private readonly %) {} }'],
    options: {
      modifiers: ['readonly'],
      selector: 'parameterProperty',
    },
    locations: [
      {
        type: 'Parameter Property',
        line: 1,
        column: 46,
        endLine: 1,
        endColumn: 47,
      },
    ],
  },
]);
