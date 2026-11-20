// Upstream: typescript-eslint v8.67.0. Locations verified with the upstream rule.
import { createTestCases } from './createTestCases';

createTestCases([
  {
    code: [
      'const ignored = { get %() {} };',
      'const ignored = { set "%"(ignored) {} };',
      'class Ignored { private get %() {} }',
      'class Ignored { private set "%"(ignored) {} }',
      'class Ignored { private static get %() {} }',
      'class Ignored { static get #%() {} }',
      'abstract class Ignored { abstract get %(): number }',
      'abstract class Ignored { abstract set %(ignored: number) }',
    ],
    options: {
      selector: 'classicAccessor',
    },
    locations: [
      {
        type: 'Classic Accessor',
        line: 1,
        column: 23,
        endLine: 1,
        endColumn: 24,
      },
      {
        type: 'Classic Accessor',
        line: 2,
        column: 23,
        endLine: 2,
        endColumn: 26,
      },
      {
        type: 'Classic Accessor',
        line: 3,
        column: 29,
        endLine: 3,
        endColumn: 30,
      },
      {
        type: 'Classic Accessor',
        line: 4,
        column: 29,
        endLine: 4,
        endColumn: 32,
      },
      {
        type: 'Classic Accessor',
        line: 5,
        column: 36,
        endLine: 5,
        endColumn: 37,
      },
      {
        type: 'Classic Accessor',
        line: 6,
        column: 28,
        endLine: 6,
        endColumn: 30,
      },
      {
        type: 'Classic Accessor',
        line: 7,
        column: 39,
        endLine: 7,
        endColumn: 40,
      },
      {
        type: 'Classic Accessor',
        line: 8,
        column: 39,
        endLine: 8,
        endColumn: 40,
      },
    ],
  },
]);
