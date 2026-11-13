// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/consistent-type-specifier-style.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/consistent-type-specifier-style.md
// Includes the upstream regression for issue #2753.
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { config } from '../config';

// Skipped cases retain their upstream records; re-enabling them also requires complete expectations.
new RuleTester(config).run('consistent-type-specifier-style', {
  valid: [
    {
      name: 'COMMON_TESTS',
      code: "import Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { Foo } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { Foo as Bar } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import * as Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import {} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type {} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type { Foo } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type { Foo as Bar } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type { Foo, Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { Foo } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { Foo as Bar } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import * as Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import {} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import type {} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { type Foo } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { type Foo as Bar } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'COMMON_TESTS',
      code: "import { type Foo, type Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'TS_ONLY',
      code: "import type * as Foo from 'Foo';",
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof { Foo, Bar, Baz, Bam } from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Foo from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo, typeof Bar, typeof Baz, typeof Bam } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { type Foo, type Bar, typeof Baz, typeof Bam } from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 1',
      code: "import type Foo from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1',
      code: "import type {Bar} from 'Bar';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1',
      code: "import type * as Bam from 'Bam';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 1 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof Baz from 'Baz';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 2',
      code: "import {type Foo} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 2 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import {typeof Bar} from 'Bar';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 4',
      code: "import type {Foo} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 4',
      code: "import type Foo, {Bar} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 4 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof {Foo} from 'Foo';",
      options: ['prefer-top-level'],
    },
    {
      name: 'Documentation example 6',
      code: "import {type Foo} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 6',
      code: "import Foo, {type Bar} from 'Foo';",
      options: ['prefer-inline'],
    },
    {
      name: 'Documentation example 6 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import {typeof Foo} from 'Foo';",
      options: ['prefer-inline'],
    },
  ],
  invalid: [
    {
      name: 'COMMON_TESTS',
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
      name: 'COMMON_TESTS',
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
      name: 'COMMON_TESTS',
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
      name: 'COMMON_TESTS',
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
      output: `import { Foo  } from 'Foo';
import type {Bar} from 'Foo';`,
    },
    {
      name: 'COMMON_TESTS',
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
      output: `import {  Bar } from 'Foo';
import type {Foo} from 'Foo';`,
    },
    {
      name: 'COMMON_TESTS',
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
      output: `import Foo from 'Foo';
import type {Bar} from 'Foo';`,
    },
    {
      name: 'COMMON_TESTS',
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
      output: `import Foo, {  Baz } from 'Foo';
import type {Bar} from 'Foo';`,
    },
    {
      name: 'COMMON_TESTS',
      code: `import { Component, type ComponentProps } from "package-1";
import {
  Component1,
  Component2,
  Component3,
  Component4,
  Component5,
} from "package-2";`,
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
      output: `import { Component  } from "package-1";
import type {ComponentProps} from "package-1";
import {
  Component1,
  Component2,
  Component3,
  Component4,
  Component5,
} from "package-2";`,
    },
    {
      name: 'COMMON_TESTS',
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
      name: 'COMMON_TESTS',
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
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportDeclaration',
        },
      ],
      output: "import typeof {Foo} from 'Foo';",
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo as Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportDeclaration',
        },
      ],
      output: "import typeof {Foo as Bar} from 'Foo';",
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { type Foo, typeof Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level type/typeof-only import instead of inline type/typeof specifiers.',
          type: 'ImportDeclaration',
        },
      ],
      output: `import type {Foo} from 'Foo';
import typeof {Bar} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo, typeof Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportDeclaration',
        },
      ],
      output: "import typeof {Foo, Bar} from 'Foo';",
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { Foo, typeof Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportSpecifier',
        },
      ],
      output: `import { Foo  } from 'Foo';
import typeof {Bar} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { typeof Foo, Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportSpecifier',
        },
      ],
      output: `import {  Bar } from 'Foo';
import typeof {Foo} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import { Foo, type Bar, typeof Baz } from 'Foo';",
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
      output: `import { Foo   } from 'Foo';
import type {Bar} from 'Foo';
import typeof {Baz} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import Foo, { typeof Bar } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportSpecifier',
        },
      ],
      output: `import Foo from 'Foo';
import typeof {Bar} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import Foo, { typeof Bar, Baz } from 'Foo';",
      options: ['prefer-top-level'],
      errors: [
        {
          message:
            'Prefer using a top-level typeof-only import instead of inline typeof specifiers.',
          type: 'ImportSpecifier',
        },
      ],
      output: `import Foo, {  Baz } from 'Foo';
import typeof {Bar} from 'Foo';`,
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof { Foo } from 'Foo';",
      options: ['prefer-inline'],
      errors: [
        {
          message:
            'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
          type: 'ImportDeclaration',
        },
      ],
      output: "import  { typeof Foo } from 'Foo';",
    } as unknown as TestCase,
    {
      name: 'FLOW_ONLY (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof { Foo, Bar, Baz } from 'Foo';",
      options: ['prefer-inline'],
      errors: [
        {
          message:
            'Prefer using inline typeof specifiers instead of a top-level typeof-only import.',
          type: 'ImportDeclaration',
        },
      ],
      output: "import  { typeof Foo, typeof Bar, typeof Baz } from 'Foo';",
    } as unknown as TestCase,
    {
      name: 'Documentation example 3',
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
      name: 'Documentation example 3',
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
      output: `import Foo from 'Foo';
import type {Bar} from 'Foo';`,
    },
    {
      name: 'Documentation example 3 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import {typeof Foo} from 'Foo';",
      options: ['prefer-top-level'],
      errors: 1,
    } as unknown as TestCase,
    {
      name: 'Documentation example 5',
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
    {
      name: 'Documentation example 5',
      code: "import type Foo, {Bar} from 'Foo';",
      options: ['prefer-inline'],
      errors: [
        {
          messageId: 'preferInline',
          message:
            'Prefer using inline type specifiers instead of a top-level type-only import.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
      output: "import   type Foo, {type type Bar} from 'Foo';",
    },
    {
      name: 'Documentation example 5 (Flow syntax unsupported by tsgo)',
      skip: 'The native TypeScript parser does not support Flow typeof import specifiers.',
      code: "import typeof {Foo} from 'Foo';",
      options: ['prefer-inline'],
      errors: 1,
    } as unknown as TestCase,
  ],
});
