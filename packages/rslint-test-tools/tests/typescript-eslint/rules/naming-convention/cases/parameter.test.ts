// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'function ignored(%) {}',
      '(function (%) {});',
      'declare function ignored(%);',
      'function ignored({%}) {}',
      'function ignored(...%) {}',
      'function ignored({% = 1}) {}',
      'function ignored({...%}) {}',
      'function ignored([%]) {}',
      'function ignored([% = 1]) {}',
      'function ignored([...%]) {}',
    ],
    options: {
      selector: 'parameter',
    },
    locations: [
      {
        type: 'Parameter',
        line: 1,
        column: 18,
        endLine: 1,
        endColumn: 19,
      },
      {
        type: 'Parameter',
        line: 2,
        column: 12,
        endLine: 2,
        endColumn: 13,
      },
      {
        type: 'Parameter',
        line: 3,
        column: 26,
        endLine: 3,
        endColumn: 27,
      },
      {
        type: 'Parameter',
        line: 4,
        column: 19,
        endLine: 4,
        endColumn: 20,
      },
      {
        type: 'Parameter',
        line: 5,
        column: 21,
        endLine: 5,
        endColumn: 22,
      },
      {
        type: 'Parameter',
        line: 6,
        column: 19,
        endLine: 6,
        endColumn: 20,
      },
      {
        type: 'Parameter',
        line: 7,
        column: 22,
        endLine: 7,
        endColumn: 23,
      },
      {
        type: 'Parameter',
        line: 8,
        column: 19,
        endLine: 8,
        endColumn: 20,
      },
      {
        type: 'Parameter',
        line: 9,
        column: 19,
        endLine: 9,
        endColumn: 20,
      },
      {
        type: 'Parameter',
        line: 10,
        column: 22,
        endLine: 10,
        endColumn: 23,
      },
    ],
  },
]);
