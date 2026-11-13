import { checkSkip } from '../../../../rslint-rule-tester/src/cases.ts';
import { test } from 'rstack/test';
// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('consistent-type-specifier-style', {
  valid: [
    {
      name: 'COMMON_TESTS / run 0 valid 0',
      code: "import Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 1',
      code: "import type Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 2',
      code: "import { Foo } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 3',
      code: "import { Foo as Bar } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 4',
      code: "import * as Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 5',
      code: "import 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 6',
      code: "import {} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 7',
      code: "import type {} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 8',
      code: "import type { Foo } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 9',
      code: "import type { Foo as Bar } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 10',
      code: "import type { Foo, Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 11',
      code: "import Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 12',
      code: "import type Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 13',
      code: "import { Foo } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 14',
      code: "import { Foo as Bar } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 15',
      code: "import * as Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 16',
      code: "import 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 17',
      code: "import {} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 18',
      code: "import type {} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 19',
      code: "import { type Foo } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 20',
      code: "import { type Foo as Bar } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS / run 0 valid 21',
      code: "import { type Foo, type Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'TS_ONLY / run 1 valid 0',
      code: "import type * as Foo from 'Foo';",
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 0',
      code: "import typeof Foo from 'Foo';",
      options: ['prefer-top-level'],
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 1',
      code: "import typeof { Foo, Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-top-level'],
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 2',
      code: "import typeof Foo from 'Foo';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 3',
      code: "import { typeof Foo } from 'Foo';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 4',
      code: "import { typeof Foo, typeof Bar, typeof Baz, typeof Bam } from 'Foo';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 valid 5',
      code: "import { type Foo, type Bar, typeof Baz, typeof Bam } from 'Foo';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    },
    {
      name: 'Documentation example 1 / run 3 valid 0',
      code: "import type Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1 / run 3 valid 1',
      code: "import type {Bar} from 'Bar';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1 / run 3 valid 2',
      code: "import type * as Bam from 'Bam';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1 (Flow syntax unsupported by tsgo) / run 4 valid 0',
      code: "import typeof Baz from 'Baz';",
      options: ['prefer-top-level'],
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'Documentation example 2 / run 5 valid 0',
      code: "import {type Foo} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 2 (Flow syntax unsupported by tsgo) / run 6 valid 0',
      code: "import {typeof Bar} from 'Bar';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    },
    {
      name: 'Documentation example 4 / run 9 valid 0',
      code: "import type {Foo} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 4 (Flow syntax unsupported by tsgo) / run 10 valid 0',
      code: "import type Foo, {Bar} from 'Foo';",
      options: ['prefer-top-level'],
      skip: 'Flow allows a type-only import with default and named bindings; TypeScript rejects that form (TS1363).',
    },
    {
      name: 'Documentation example 4 (Flow syntax unsupported by tsgo) / run 10 valid 1',
      code: "import typeof {Foo} from 'Foo';",
      options: ['prefer-top-level'],
      skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    },
    {
      name: 'Documentation example 6 / run 13 valid 0',
      code: "import {type Foo} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 6 / run 13 valid 1',
      code: "import Foo, {type Bar} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 6 (Flow syntax unsupported by tsgo) / run 14 valid 0',
      code: "import {typeof Foo} from 'Foo';",
      options: ['prefer-inline'],
      skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    },
  ],
  invalid: [
    {
      name: 'COMMON_TESTS / run 0 invalid 0',
      code: "import { type Foo } from 'Foo';",
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
      output: "import type {Foo} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 1',
      code: "import { type Foo as Bar } from 'Foo';",
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
      output: "import type {Foo as Bar} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 2',
      code: "import { type Foo, type Bar } from 'Foo';",
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
      output: "import type {Foo, Bar} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 3',
      code: "import { Foo, type Bar } from 'Foo';",
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
      output: "import { Foo  } from 'Foo';\nimport type {Bar} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 4',
      code: "import { type Foo, Bar } from 'Foo';",
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
      output: "import {  Bar } from 'Foo';\nimport type {Foo} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 5',
      code: "import Foo, { type Bar } from 'Foo';",
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
      output: "import Foo from 'Foo';\nimport type {Bar} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 6',
      code: "import Foo, { type Bar, Baz } from 'Foo';",
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
      output: "import Foo, {  Baz } from 'Foo';\nimport type {Bar} from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 7',
      code: 'import { Component, type ComponentProps } from "package-1";\nimport {\n  Component1,\n  Component2,\n  Component3,\n  Component4,\n  Component5,\n} from "package-2";',
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
      output:
        'import { Component  } from "package-1";\nimport type {ComponentProps} from "package-1";\nimport {\n  Component1,\n  Component2,\n  Component3,\n  Component4,\n  Component5,\n} from "package-2";',
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 8',
      code: "import type { Foo } from 'Foo';",
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
      output: "import  { type Foo } from 'Foo';",
    },
    {
      name: 'COMMON_TESTS / run 0 invalid 9',
      code: "import type { Foo, Bar, Baz } from 'Foo';",
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
      output: "import  { type Foo, type Bar, type Baz } from 'Foo';",
    },
    {
      name: 'Documentation example 3 / run 7 invalid 0',
      code: "import {type Foo} from 'Foo';",
      options: ['prefer-top-level'],
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
      output: "import type {Foo} from 'Foo';",
    },
    {
      name: 'Documentation example 3 / run 7 invalid 1',
      code: "import Foo, {type Bar} from 'Foo';",
      options: ['prefer-top-level'],
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
      output: "import Foo from 'Foo';\nimport type {Bar} from 'Foo';",
    },
    {
      name: 'Documentation example 5 / run 11 invalid 0',
      code: "import type {Foo} from 'Foo';",
      options: ['prefer-inline'],
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
      output: "import  {type Foo} from 'Foo';",
    },
  ],
});

// Preserved upstream inputs/assertions, not executable native expectations.
const skippedInvalidCases = [
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 0',
    code: "import { typeof Foo } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import typeof {Foo} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 1',
    code: "import { typeof Foo as Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import typeof {Foo as Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 2',
    code: "import { type Foo, typeof Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level type/typeof-only import instead of inline type/typeof specifiers.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import type {Foo} from 'Foo';\nimport typeof {Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 3',
    code: "import { typeof Foo, typeof Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import typeof {Foo, Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 4',
    code: "import { Foo, typeof Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportSpecifier',
      },
    ],
    output: "import { Foo  } from 'Foo';\nimport typeof {Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 5',
    code: "import { typeof Foo, Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportSpecifier',
      },
    ],
    output: "import {  Bar } from 'Foo';\nimport typeof {Foo} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 6',
    code: "import { Foo, type Bar, typeof Baz } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
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
    output:
      "import { Foo   } from 'Foo';\nimport type {Bar} from 'Foo';\nimport typeof {Baz} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 7',
    code: "import Foo, { typeof Bar } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportSpecifier',
      },
    ],
    output: "import Foo from 'Foo';\nimport typeof {Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 8',
    code: "import Foo, { typeof Bar, Baz } from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: [
      {
        message:
          'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
        type: 'ImportSpecifier',
      },
    ],
    output: "import Foo, {  Baz } from 'Foo';\nimport typeof {Bar} from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 9',
    code: "import typeof { Foo } from 'Foo';",
    options: ['prefer-inline'],
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    errors: [
      {
        message:
          'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import  { typeof Foo } from 'Foo';",
  },
  {
    name: 'FLOW_ONLY (Flow syntax unsupported by tsgo) / run 2 invalid 10',
    code: "import typeof { Foo, Bar, Baz } from 'Foo';",
    options: ['prefer-inline'],
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    errors: [
      {
        message:
          'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
        type: 'ImportDeclaration',
      },
    ],
    output: "import  { typeof Foo, typeof Bar, typeof Baz } from 'Foo';",
  },
  {
    name: 'Documentation example 3 (Flow syntax unsupported by tsgo) / run 8 invalid 0',
    code: "import {typeof Foo} from 'Foo';",
    options: ['prefer-top-level'],
    skip: 'TypeScript-Go does not support Flow typeof import specifiers.',
    errors: 1,
  },
  {
    name: 'Documentation example 5 (Flow syntax unsupported by tsgo) / run 12 invalid 0',
    code: "import type Foo, {Bar} from 'Foo';",
    options: ['prefer-inline'],
    skip: 'Flow allows a type-only import with default and named bindings; TypeScript rejects that form (TS1363).',
    errors: 1,
  },
  {
    name: 'Documentation example 5 (Flow syntax unsupported by tsgo) / run 12 invalid 1',
    code: "import typeof {Foo} from 'Foo';",
    options: ['prefer-inline'],
    skip: 'TypeScript-Go does not support Flow import typeof declarations.',
    errors: 1,
  },
];
for (const item of skippedInvalidCases) {
  checkSkip(item);
  test.skip(`invalid ${item.name}: ${item.skip}`, () => {});
}
