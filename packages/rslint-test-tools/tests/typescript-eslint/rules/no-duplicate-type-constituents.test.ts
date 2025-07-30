import { RuleTester } from '@typescript-eslint/rule-tester';
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
    'type T = string | number;',
    'type T = string | number | boolean;',
    'type T = string | number | boolean | object;',

    // Basic intersection types - no duplicates
    'type T = string & { length: number };',
    'type T = { a: string } & { b: number };',
    'type T = { a: string } & { b: number } & { c: boolean };',

    // Different but similar types
    'type T = string | String;',
    'type T = number | Number;',
    'type T = boolean | Boolean;',
    'type T = object | Object;',

    // Generic types with different parameters
    'type T = Array<string> | Array<number>;',
    'type T = Promise<string> | Promise<number>;',
    'type T = Record<string, string> | Record<string, number>;',

    // Object types with different properties
    'type T = { a: string } | { b: number };',
    'type T = { a: string; b: number } | { a: string; c: boolean };',

    // Function types with different signatures
    'type T = ((a: string) => void) | ((a: number) => void);',
    'type T = (() => string) | (() => number);',

    // Nested unions/intersections without duplicates
    'type T = (string | number) | (boolean | object);',
    'type T = (string & { length: number }) | (number & { toString(): string });',

    // Union with undefined on required parameter
    'function f(x: string | undefined): void {}',

    // Different literal types
    'type T = "a" | "b" | "c";',
    'type T = 1 | 2 | 3;',
    'type T = true | false;',

    // Template literal types
    'type T = `prefix-${string}` | `suffix-${string}`;',

    // Conditional types
    'type T<U> = U extends string ? string | number : number | boolean;',

    // Tuple types - different tuples
    'type T = [string, number] | [number, string];',
    'type T = [string] | [string, number];',

    // Mapped types
    'type T<K extends keyof any> = { [P in K]: string } | { [P in K]: number };',

    // Index access types
    'type T = { a: string }["a"] | { b: number }["b"];',

    // keyof types
    'type T = keyof { a: string } | keyof { b: number };',

    // typeof types
    'type T = typeof String | typeof Number;',

    // ignoreUnions option
    {
      code: 'type T = string | string;',
      options: [{ ignoreUnions: true }],
    },
    {
      code: 'type T = number | number | number;',
      options: [{ ignoreUnions: true }],
    },

    // ignoreIntersections option
    {
      code: 'type T = string & string;',
      options: [{ ignoreIntersections: true }],
    },
    {
      code: 'type T = { a: string } & { a: string };',
      options: [{ ignoreIntersections: true }],
    },
  ],

  invalid: [
    // Basic union duplicates
    {
      code: 'type T = string | string;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: 'type T = string | number | string;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: 'type T = number | string | number | boolean;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Basic intersection duplicates
    {
      code: 'type T = string & string;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Multiple duplicates in one type
    {
      code: 'type T = string | number | string | boolean | number;',
      errors: [
        {
          messageId: 'duplicate',
        },
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Parenthesized types
    {
      code: 'type T = (string) | (string);',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Object type duplicates
    {
      code: 'type T = { a: string } | { a: string };',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Array type duplicates
    {
      code: 'type T = string[] | number[] | string[];',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Generic type duplicates
    {
      code: 'type T = Array<string> | Promise<number> | Array<string>;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Function type duplicates
    {
      code: 'type T = ((a: string) => void) | ((a: string) => void);',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Literal type duplicates
    {
      code: 'type T = "a" | "b" | "a";',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: 'type T = 1 | 2 | 1;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Nested duplicates
    {
      code: 'type T = string | (number | string);',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: 'type T = (string | number) | (boolean | string);',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Mixed union and intersection
    {
      code: 'type T = (string & { length: number }) | (string & { length: number });',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Optional parameter with undefined (unnecessary explicit undefined)
    {
      code: 'function f(x?: string | undefined): void {}',
      errors: [
        {
          messageId: 'unnecessary',
        },
      ],
    },
    {
      code: 'function f(x?: number | boolean | undefined): void {}',
      errors: [
        {
          messageId: 'unnecessary',
        },
      ],
    },
    {
      code: 'const f = (x?: string | undefined) => {}',
      errors: [
        {
          messageId: 'unnecessary',
        },
      ],
    },

    // Complex nested cases
    {
      code: 'type T = string | (number | string) | boolean;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
    {
      code: 'type T = (string | number) | (string | boolean);',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Interface types
    {
      code: `
        interface I { a: string }
        type T = I | number | I;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Class types
    {
      code: `
        class C { a!: string }
        type T = C | number | C;
      `,
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Tuple type duplicates
    {
      code: 'type T = [string, number] | [boolean] | [string, number];',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Template literal duplicates
    {
      code: 'type T = `prefix-${string}` | `prefix-${number}` | `prefix-${string}`;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // keyof duplicates
    {
      code: 'type T = keyof { a: string } | keyof { b: number } | keyof { a: string };',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // typeof duplicates
    {
      code: 'type T = typeof String | typeof Number | typeof String;',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Complex intersection duplicates
    {
      code: 'type T = { a: string } & { b: number } & { a: string };',
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // Multiple nested duplicates
    {
      code: 'type T = (string | (number | string)) | (boolean | (string | object));',
      errors: [
        {
          messageId: 'duplicate',
        },
        {
          messageId: 'duplicate',
        },
      ],
    },

    // When ignoreIntersections is false (default), intersection duplicates should error
    {
      code: 'type T = string & string;',
      options: [{ ignoreIntersections: false }],
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },

    // When ignoreUnions is false (default), union duplicates should error
    {
      code: 'type T = string | string;',
      options: [{ ignoreUnions: false }],
      errors: [
        {
          messageId: 'duplicate',
        },
      ],
    },
  ],
});
