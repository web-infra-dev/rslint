// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'class Ignored { private % }',
      'class Ignored { private "%" = 1 }',
      'class Ignored { private readonly % = 1 }',
      'class Ignored { private static % }',
      'class Ignored { private static readonly % = 1 }',
      'class Ignored { abstract % }',
      'class Ignored { declare % }',
      'class Ignored { #% }',
      'class Ignored { static #% }',
    ],
    options: {
      selector: 'classProperty',
    },
    locations: [
      {
        type: 'Class Property',
        line: 1,
        column: 25,
        endLine: 1,
        endColumn: 26,
      },
      {
        type: 'Class Property',
        line: 2,
        column: 25,
        endLine: 2,
        endColumn: 28,
      },
      {
        type: 'Class Property',
        line: 3,
        column: 34,
        endLine: 3,
        endColumn: 35,
      },
      {
        type: 'Class Property',
        line: 4,
        column: 32,
        endLine: 4,
        endColumn: 33,
      },
      {
        type: 'Class Property',
        line: 5,
        column: 41,
        endLine: 5,
        endColumn: 42,
      },
      {
        type: 'Class Property',
        line: 6,
        column: 26,
        endLine: 6,
        endColumn: 27,
      },
      {
        type: 'Class Property',
        line: 7,
        column: 25,
        endLine: 7,
        endColumn: 26,
      },
      {
        type: 'Class Property',
        line: 8,
        column: 17,
        endLine: 8,
        endColumn: 19,
      },
      {
        type: 'Class Property',
        line: 9,
        column: 24,
        endLine: 9,
        endColumn: 26,
      },
    ],
  },
  {
    code: ['const ignored = { % };', 'const ignored = { "%": 1 };'],
    options: {
      selector: 'objectLiteralProperty',
    },
    locations: [
      {
        type: 'Object Literal Property',
        line: 1,
        column: 19,
        endLine: 1,
        endColumn: 20,
      },
      {
        type: 'Object Literal Property',
        line: 2,
        column: 19,
        endLine: 2,
        endColumn: 22,
      },
    ],
  },
  {
    code: [
      'interface Ignored { % }',
      'interface Ignored { "%": string }',
      'type Ignored = { % }',
      'type Ignored = { "%": string }',
    ],
    options: {
      selector: 'typeProperty',
    },
    locations: [
      {
        type: 'Type Property',
        line: 1,
        column: 21,
        endLine: 1,
        endColumn: 22,
      },
      {
        type: 'Type Property',
        line: 2,
        column: 21,
        endLine: 2,
        endColumn: 24,
      },
      {
        type: 'Type Property',
        line: 3,
        column: 18,
        endLine: 3,
        endColumn: 19,
      },
      {
        type: 'Type Property',
        line: 4,
        column: 18,
        endLine: 4,
        endColumn: 21,
      },
    ],
  },
]);
