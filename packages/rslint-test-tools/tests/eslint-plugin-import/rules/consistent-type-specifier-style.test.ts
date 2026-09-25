import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/consistent-type-specifier-style.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/consistent-type-specifier-style.md
// This wrapper checks counts and messages; Go tests and differential validation
// also verify full ranges, fixes, and rslint's message IDs (upstream has none).
const ruleTester = new RuleTester();
const rule = null as never;

// Includes the upstream regression for issue #2753.
describe('COMMON_TESTS', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import Foo from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type Foo from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import { Foo } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import { Foo as Bar } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import * as Foo from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import {} from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type {} from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type { Foo } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type { Foo as Bar } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type { Foo, Bar, Baz, Bam } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import Foo from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import type Foo from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { Foo } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { Foo as Bar } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import * as Foo from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import {} from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import type {} from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { type Foo } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { type Foo as Bar } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { type Foo, type Bar, Baz, Bam } from 'Foo';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [
      {
        code: "import { type Foo } from 'Foo';",
        output: "import type {Foo} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        code: "import { type Foo as Bar } from 'Foo';",
        output: "import type {Foo as Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 39,
          },
        ],
      },
      {
        code: "import { type Foo, type Bar } from 'Foo';",
        output: "import type {Foo, Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
      {
        code: "import { Foo, type Bar } from 'Foo';",
        output: "import { Foo  } from 'Foo';\nimport type {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        code: "import { type Foo, Bar } from 'Foo';",
        output: "import {  Bar } from 'Foo';\nimport type {Foo} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 10,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        code: "import Foo, { type Bar } from 'Foo';",
        output: "import Foo from 'Foo';\nimport type {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        code: "import Foo, { type Bar, Baz } from 'Foo';",
        output:
          "import Foo, {  Baz } from 'Foo';\nimport type {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 15,
            endLine: 1,
            endColumn: 23,
          },
        ],
      },
      {
        code: 'import { Component, type ComponentProps } from "package-1";\nimport {\n  Component1,\n  Component2,\n  Component3,\n  Component4,\n  Component5,\n} from "package-2";',
        output:
          'import { Component  } from "package-1";\nimport type {ComponentProps} from "package-1";\nimport {\n  Component1,\n  Component2,\n  Component3,\n  Component4,\n  Component5,\n} from "package-2";',
        options: ['prefer-top-level'],
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 21,
            endLine: 1,
            endColumn: 40,
          },
        ],
      },
      {
        code: "import type { Foo } from 'Foo';",
        output: "import  { type Foo } from 'Foo';",
        options: ['prefer-inline'],
        errors: [
          {
            messageId: 'preferInline',
            message:
              'Prefer using inline type specifiers instead of a top-level type-only import.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 32,
          },
        ],
      },
      {
        code: "import type { Foo, Bar, Baz } from 'Foo';",
        output: "import  { type Foo, type Bar, type Baz } from 'Foo';",
        options: ['prefer-inline'],
        errors: [
          {
            messageId: 'preferInline',
            message:
              'Prefer using inline type specifiers instead of a top-level type-only import.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 42,
          },
        ],
      },
    ],
  });
});

describe('TS_ONLY', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import type * as Foo from 'Foo';",
      },
    ],
    invalid: [],
  });
});

describe.skip('FLOW_ONLY (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import typeof Foo from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import typeof { Foo, Bar, Baz, Bam } from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import typeof Foo from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { typeof Foo } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { typeof Foo, typeof Bar, typeof Baz, typeof Bam } from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import { type Foo, type Bar, typeof Baz, typeof Bam } from 'Foo';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [
      {
        code: "import { typeof Foo } from 'Foo';",
        output: "import typeof {Foo} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportDeclaration',
          },
        ],
      },
      {
        code: "import { typeof Foo as Bar } from 'Foo';",
        output: "import typeof {Foo as Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportDeclaration',
          },
        ],
      },
      {
        code: "import { type Foo, typeof Bar } from 'Foo';",
        output:
          "import type {Foo} from 'Foo';\nimport typeof {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level type/typeof-only import instead of inline type/typeof specifiers.',
            type: 'ImportDeclaration',
          },
        ],
      },
      {
        code: "import { typeof Foo, typeof Bar } from 'Foo';",
        output: "import typeof {Foo, Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportDeclaration',
          },
        ],
      },
      {
        code: "import { Foo, typeof Bar } from 'Foo';",
        output: "import { Foo  } from 'Foo';\nimport typeof {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportSpecifier',
          },
        ],
      },
      {
        code: "import { typeof Foo, Bar } from 'Foo';",
        output: "import {  Bar } from 'Foo';\nimport typeof {Foo} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportSpecifier',
          },
        ],
      },
      {
        code: "import { Foo, type Bar, typeof Baz } from 'Foo';",
        output:
          "import { Foo   } from 'Foo';\nimport type {Bar} from 'Foo';\nimport typeof {Baz} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            type: 'ImportSpecifier',
          },
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportSpecifier',
          },
        ],
      },
      {
        code: "import Foo, { typeof Bar } from 'Foo';",
        output: "import Foo from 'Foo';\nimport typeof {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportSpecifier',
          },
        ],
      },
      {
        code: "import Foo, { typeof Bar, Baz } from 'Foo';",
        output:
          "import Foo, {  Baz } from 'Foo';\nimport typeof {Bar} from 'Foo';",
        options: ['prefer-top-level'],
        errors: [
          {
            message:
              'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
            type: 'ImportSpecifier',
          },
        ],
      },
      {
        code: "import typeof { Foo } from 'Foo';",
        output: "import  { typeof Foo } from 'Foo';",
        options: ['prefer-inline'],
        errors: [
          {
            message:
              'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
            type: 'ImportDeclaration',
          },
        ],
      },
      {
        code: "import typeof { Foo, Bar, Baz } from 'Foo';",
        output: "import  { typeof Foo, typeof Bar, typeof Baz } from 'Foo';",
        options: ['prefer-inline'],
        errors: [
          {
            message:
              'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
            type: 'ImportDeclaration',
          },
        ],
      },
    ],
  });
});

describe('Documentation example 1', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import type Foo from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type {Bar} from 'Bar';",
        options: ['prefer-top-level'],
      },
      {
        code: "import type * as Bam from 'Bam';",
        options: ['prefer-top-level'],
      },
    ],
    invalid: [],
  });
});

describe.skip('Documentation example 1 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import typeof Baz from 'Baz';",
        options: ['prefer-top-level'],
      },
    ],
    invalid: [],
  });
});

describe('Documentation example 2', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import {type Foo} from 'Foo';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [],
  });
});

describe.skip('Documentation example 2 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import {typeof Bar} from 'Bar';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [],
  });
});

describe('Documentation example 3', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [],
    invalid: [
      {
        code: "import {type Foo} from 'Foo';",
        options: ['prefer-top-level'],
        output: "import type {Foo} from 'Foo';",
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
      },
      {
        code: "import Foo, {type Bar} from 'Foo';",
        options: ['prefer-top-level'],
        output: "import Foo from 'Foo';\nimport type {Bar} from 'Foo';",
        errors: [
          {
            messageId: 'preferTopLevel',
            message:
              'Prefer using a top-level type-only import instead of inline type specifiers.',
            line: 1,
            column: 14,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
    ],
  });
});

describe.skip('Documentation example 3 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [],
    invalid: [
      {
        code: "import {typeof Foo} from 'Foo';",
        options: ['prefer-top-level'],
        errors: 1,
      },
    ],
  });
});

describe('Documentation example 4', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import type {Foo} from 'Foo';",
        options: ['prefer-top-level'],
      },
    ],
    invalid: [],
  });
});

describe.skip('Documentation example 4 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import type Foo, {Bar} from 'Foo';",
        options: ['prefer-top-level'],
      },
      {
        code: "import typeof {Foo} from 'Foo';",
        options: ['prefer-top-level'],
      },
    ],
    invalid: [],
  });
});

describe('Documentation example 5', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [],
    invalid: [
      {
        code: "import type {Foo} from 'Foo';",
        options: ['prefer-inline'],
        output: "import  {type Foo} from 'Foo';",
        errors: [
          {
            messageId: 'preferInline',
            message:
              'Prefer using inline type specifiers instead of a top-level type-only import.',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 30,
          },
        ],
      },
    ],
  });
});

describe.skip('Documentation example 5 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [],
    invalid: [
      {
        code: "import type Foo, {Bar} from 'Foo';",
        options: ['prefer-inline'],
        errors: 1,
      },
      {
        code: "import typeof {Foo} from 'Foo';",
        options: ['prefer-inline'],
        errors: 1,
      },
    ],
  });
});

describe('Documentation example 6', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import {type Foo} from 'Foo';",
        options: ['prefer-inline'],
      },
      {
        code: "import Foo, {type Bar} from 'Foo';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [],
  });
});

describe.skip('Documentation example 6 (Flow syntax unsupported by tsgo)', () => {
  ruleTester.run('consistent-type-specifier-style', rule, {
    valid: [
      {
        code: "import {typeof Foo} from 'Foo';",
        options: ['prefer-inline'],
      },
    ],
    invalid: [],
  });
});
