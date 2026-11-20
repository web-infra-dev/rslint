// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'class Ignored<%> {}',
      'function ignored<%>() {}',
      'type Ignored<%> = { ignored: % };',
      'interface Ignored<%> extends Ignored<string> {}',
    ],
    options: {
      selector: 'typeParameter',
    },
    locations: [
      {
        type: 'Type Parameter',
        line: 1,
        column: 15,
        endLine: 1,
        endColumn: 16,
      },
      {
        type: 'Type Parameter',
        line: 2,
        column: 18,
        endLine: 2,
        endColumn: 19,
      },
      {
        type: 'Type Parameter',
        line: 3,
        column: 14,
        endLine: 3,
        endColumn: 15,
      },
      {
        type: 'Type Parameter',
        line: 4,
        column: 19,
        endLine: 4,
        endColumn: 20,
      },
    ],
  },
]);
