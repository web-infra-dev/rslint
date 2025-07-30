import { noFormat, RuleTester } from '@typescript-eslint/rule-tester';
import { getFixturesRootDir } from '../RuleTester';

const rootDir = getFixturesRootDir();
const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootDir,
    },
  },
});

ruleTester.run('no-duplicate-type-constituents', {
  valid: [
    // Basic union types - no duplicates
    `
type T = number | boolean;
    `,
    `
type T = string | number;
    `,
    `
type T = string | number | boolean;
    `,

    // Basic intersection types - no duplicates
    `
type T = string & { length: number };
    `,
    `
type T = { a: string } & { b: number };
    `,

    // Different but similar types
    `
type T = string | String;
    `,
    `
type T = number | Number;
    `,

    // Generic types with different parameters
    `
type T = Array<string> | Array<number>;
    `,

    // Object types with different properties
    `
type T = { a: string } | { b: number };
    `,

    // Function types with different signatures
    `
type T = ((a: string) => void) | ((a: number) => void);
    `,

    // Different literal types
    `
type T = "a" | "b" | "c";
    `,
    `
type T = 1 | 2 | 3;
    `,

    // Template literal types
    `
type T = \`prefix-\${string}\` | \`suffix-\${string}\`;
    `,

    // ignoreUnions option
    {
      code: `
type T = string | string;
      `,
      options: [{ ignoreUnions: true }],
    },

    // ignoreIntersections option
    {
      code: `
type T = string & string;
      `,
      options: [{ ignoreIntersections: true }],
    },
  ],

  invalid: [
    // Basic union duplicates
    {
      code: `
type T = string | string;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: `
type T = string | number | string;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: `
type T = number | string | number | boolean;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Basic intersection duplicates
    {
      code: `
type T = string & string;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Multiple duplicates in one type
    {
      code: `
type T = string | number | string | boolean | number;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Literal type duplicates
    {
      code: `
type T = "a" | "b" | "a";
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: `
type T = 1 | 2 | 1;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Optional parameter with undefined (unnecessary explicit undefined)
    {
      code: `
function f(x?: string | undefined): void {}
      `,
      errors: [
        {
          messageId: 'unnecessary',
        },
      ],
    },
  ],
});
