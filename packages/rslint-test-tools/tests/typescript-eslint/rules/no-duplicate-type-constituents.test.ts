// Complete upstream suite: typescript-eslint v8.72.2.
// https://github.com/typescript-eslint/typescript-eslint/blob/v8.72.2/packages/eslint-plugin/tests/rules/no-duplicate-type-constituents.test.ts
// Go upstream tests verify each autofix pass; this suite verifies the native
// multi-pass output, complete diagnostics and absence of suggestions.
import { RuleTester } from '@rslint/rule-tester';

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'file.ts',
  fixtureFiles: {
    'tsconfig.json': JSON.stringify({
      compilerOptions: { strict: true, target: 'esnext' },
      include: ['**/*.ts'],
    }),
  },
  languageOptions: { parserOptions: { project: './tsconfig.json' } },
});

ruleTester.run('no-duplicate-type-constituents', {
  valid: [
    {
      code: 'type T = 1 | 2;',
    },
    {
      code: "type T = 1 | '1';",
    },
    {
      code: 'type T = true & boolean;',
    },
    {
      code: 'type T = null | undefined;',
    },
    {
      code: 'type T = any | unknown;',
    },
    {
      code: 'type T = { a: string } | { b: string };',
    },
    {
      code: 'type T = { a: string; b: number } | { b: number; a: string };',
    },
    {
      code: 'type T = { a: string | number };',
    },
    {
      code: 'type T = Set<string> | Set<number>;',
    },
    {
      code: 'type T = Class<string> | Class<number>;',
    },
    {
      code: 'type T = string[] | number[];',
    },
    {
      code: 'type T = string[][] | string[];',
    },
    {
      code: 'type T = [1, 2, 3] | [1, 2, 4];',
    },
    {
      code: 'type T = [1, 2, 3] | [1, 2, 3, 4];',
    },
    {
      code: "type T = 'A' | string[];",
    },
    {
      code: 'type T = (() => string) | (() => void);',
    },
    {
      code: 'type T = () => string | void;',
    },
    {
      code: 'type T = () => null | undefined;',
    },
    {
      code: 'type T = (arg: string | number) => void;',
    },
    {
      code: 'type T = A | A;',
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = A | B;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
const a: A | B = 'A';
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = A | /* comment */ B;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = 'A' | 'B';
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | C;
      `,
    },
    {
      code: 'type T = readonly string[] | string[];',
    },
    {
      code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type T = (A | B) | (C | D);
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A & B);
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = Record<string, A | B>;
      `,
    },
    {
      code: 'type T = A | A;',
      options: [
        {
          ignoreUnions: true,
        },
      ],
    },
    {
      code: 'type T = A & A;',
      options: [
        {
          ignoreIntersections: true,
        },
      ],
    },
    {
      code: 'type T = Class<string> | Class<string>;',
    },
    {
      code: 'type T = A | A | string;',
    },
    { code: '(a: string | undefined) => {};' },
  ],
  invalid: [
    {
      code: 'type T = 1 | 1;',
      errors: [
        {
          column: 14,
          message: "Union type constituent is duplicated with " + '1' + '.',
          endColumn: 15,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = 1  ;`,
    },
    {
      code: 'type T = true & true;',
      errors: [
        {
          column: 17,
          message: "Intersection type constituent is duplicated with " + 'true' + '.',
          endColumn: 21,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = true  ;`,
    },
    {
      code: 'type T = null | null;',
      errors: [
        {
          column: 17,
          message: "Union type constituent is duplicated with " + 'null' + '.',
          endColumn: 21,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = null  ;`,
    },
    {
      code: 'type T = any | any;',
      errors: [
        {
          column: 16,
          message: "Union type constituent is duplicated with " + 'any' + '.',
          endColumn: 19,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = any  ;`,
    },
    {
      code: 'type T = { a: string | string };',
      errors: [
        {
          column: 24,
          message: "Union type constituent is duplicated with " + 'string' + '.',
          endColumn: 30,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = { a: string   };`,
    },
    {
      code: 'type T = { a: string } | { a: string };',
      errors: [
        {
          column: 26,
          message: "Union type constituent is duplicated with " + '{ a: string }' + '.',
          endColumn: 39,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = { a: string }  ;`,
    },
    {
      code: 'type T = { a: string; b: number } | { a: string; b: number };',
      errors: [
        {
          column: 37,
          message: "Union type constituent is duplicated with " + '{ a: string; b: number }' + '.',
          endColumn: 61,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = { a: string; b: number }  ;`,
    },
    {
      code: 'type T = Set<string> | Set<string>;',
      errors: [
        {
          column: 24,
          message: "Union type constituent is duplicated with " + 'Set<string>' + '.',
          endColumn: 35,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = Set<string>  ;`,
    },
    {
      code: `
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number> | IsArray<string>;
      `,
      errors: [
        {
          column: 45,
          message: "Union type constituent is duplicated with " + 'IsArray<number>' + '.',
          endColumn: 60,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number>  ;
      `,
    },
    {
      code: 'type T = string[] | string[];',
      errors: [
        {
          column: 21,
          message: "Union type constituent is duplicated with " + 'string[]' + '.',
          endColumn: 29,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = string[]  ;`,
    },
    {
      code: 'type T = string[][] | string[][];',
      errors: [
        {
          column: 23,
          message: "Union type constituent is duplicated with " + 'string[][]' + '.',
          endColumn: 33,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = string[][]  ;`,
    },
    {
      code: 'type T = [1, 2, 3] | [1, 2, 3];',
      errors: [
        {
          column: 22,
          message: "Union type constituent is duplicated with " + '[1, 2, 3]' + '.',
          endColumn: 31,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = [1, 2, 3]  ;`,
    },
    {
      code: 'type T = () => string | string;',
      errors: [
        {
          column: 25,
          message: "Union type constituent is duplicated with " + 'string' + '.',
          endColumn: 31,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = () => string  ;`,
    },
    {
      code: 'type T = () => null | null;',
      errors: [
        {
          column: 23,
          message: "Union type constituent is duplicated with " + 'null' + '.',
          endColumn: 27,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = () => null  ;`,
    },
    {
      code: 'type T = (arg: string | string) => void;',
      errors: [
        {
          column: 25,
          message: "Union type constituent is duplicated with " + 'string' + '.',
          endColumn: 31,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = (arg: string  ) => void;`,
    },
    {
      code: "type T = 'A' | 'A';",
      errors: [
        {
          column: 16,
          message: "Union type constituent is duplicated with " + "'A'" + '.',
          endColumn: 19,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: `type T = 'A'  ;`,
    },
    {
      code: `
type A = 'A';
type T = A | A;
      `,
      errors: [
        {
          column: 14,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 15,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type T = A  ;
      `,
    },
    {
      code: `
type A = 'A';
const a: A | A = 'A';
      `,
      errors: [
        {
          column: 14,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 15,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
const a: A   = 'A';
      `,
    },
    {
      code: `
type A = 'A';
type T = A | /* comment */ A;
      `,
      errors: [
        {
          column: 28,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 29,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type T = A  /* comment */ ;
      `,
    },
    {
      code: `
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1 | A2 | A3;
      `,
      errors: [
        {
          column: 15,
          message: "Union type constituent is duplicated with " + 'A1' + '.',
          endColumn: 17,
          endLine: 5,
          line: 5,
          messageId: 'duplicate',
        },
        {
          column: 20,
          message: "Union type constituent is duplicated with " + 'A1' + '.',
          endColumn: 22,
          endLine: 5,
          line: 5,
          messageId: 'duplicate',
        },
      ],
      output: `
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1    ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = A | B | A;
      `,
      errors: [
        {
          column: 18,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 19,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type T = A | B  ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = A | B | A | B;
      `,
      errors: [
        {
          column: 18,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 19,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
        {
          column: 22,
          message: "Union type constituent is duplicated with " + 'B' + '.',
          endColumn: 23,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = A | B | A | A;
      `,
      errors: [
        {
          column: 18,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 19,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
        {
          column: 22,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 23,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | A | C;
      `,
      errors: [
        {
          column: 18,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 19,
          endLine: 5,
          line: 5,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B   | C;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A | B);
      `,
      errors: [
        {
          column: 21,
          message: "Union type constituent is duplicated with " + 'A | B' + '.',
          endColumn: 27,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type T = (A | B)  ;
      `,
    },
    {
      code: `
type A = 'A';
type T = A | (A | A);
      `,
      errors: [
        {
          column: 15,
          message: "Union type constituent is duplicated with " + `A` + '.',
          endColumn: 21,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type T = A  ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B) | (A | B) | ((C | D) & (A | B)) | (A | B);
      `,
      errors: [
        {
          column: 21,
          message: "Union type constituent is duplicated with " + 'A | B' + '.',
          endColumn: 27,
          endLine: 6,
          line: 6,
          messageId: 'duplicate',
        },
        {
          column: 53,
          message: "Union type constituent is duplicated with " + 'A | B' + '.',
          endColumn: 59,
          endLine: 6,
          line: 6,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B)   | ((C | D) & (A | B))  ;
      `,
    },
    {
      code: `
type A = 'A';
type B = 'B';
type C = (A | B) | A | B | (A | B);
      `,
      errors: [
        {
          column: 20,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 21,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
        {
          column: 24,
          message: "Union type constituent is duplicated with " + 'B' + '.',
          endColumn: 25,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
        {
          column: 29,
          message: "Union type constituent is duplicated with " + 'A | B' + '.',
          endColumn: 35,
          endLine: 4,
          line: 4,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type B = 'B';
type C = (A | B)      ;
      `,
    },
    {
      code: 'type A = (number | string) | number | string;',
      errors: [
        {
          column: 30,
          message: "Union type constituent is duplicated with " + 'number' + '.',
          endColumn: 36,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
        {
          column: 39,
          message: "Union type constituent is duplicated with " + 'string' + '.',
          endColumn: 45,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: 'type A = (number | string)    ;',
    },
    {
      code: 'type A = (number | (string | null)) | (string | (null | number));',
      errors: [
        {
          column: 40,
          message: "Union type constituent is duplicated with " + 'number | (string | null)' + '.',
          endColumn: 65,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: 'type A = (number | (string | null))  ;',
    },
    {
      code: 'type A = (number & string) & number & string;',
      errors: [
        {
          column: 30,
          message: "Intersection type constituent is duplicated with " + 'number' + '.',
          endColumn: 36,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
        {
          column: 39,
          message: "Intersection type constituent is duplicated with " + 'string' + '.',
          endColumn: 45,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: 'type A = (number & string)    ;',
    },
    {
      code: 'type A = number & string & (number & string);',
      errors: [
        {
          column: 29,
          message: "Intersection type constituent is duplicated with " + 'number' + '.',
          endColumn: 35,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
        {
          column: 38,
          message: "Intersection type constituent is duplicated with " + 'string' + '.',
          endColumn: 44,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: [
        'type A = number & string & (  string);',
        'type A = number & string    ;',
      ].at(-1)!,
    },
    {
      code: `
type A = 'A';
type T = Record<string, A | A>;
      `,
      errors: [
        {
          column: 29,
          message: "Union type constituent is duplicated with " + 'A' + '.',
          endColumn: 30,
          endLine: 3,
          line: 3,
          messageId: 'duplicate',
        },
      ],
      output: `
type A = 'A';
type T = Record<string, A  >;
      `,
    },
    {
      code: 'type T = A | A | string | string;',
      errors: [
        {
          column: 27,
          message: "Union type constituent is duplicated with " + 'string' + '.',
          endColumn: 33,
          endLine: 1,
          line: 1,
          messageId: 'duplicate',
        },
      ],
      output: 'type T = A | A | string  ;',
    },
    {
      code: '(a?: string | undefined) => {};',
      errors: [
        {
          column: 15,
          endColumn: 24,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: '(a?: string  ) => {};',
    },
    {
      code: `
type T = undefined;
(arg?: T | string) => {};
      `,
      errors: [
        {
          column: 8,
          endColumn: 9,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
type T = undefined;
(arg?:   string) => {};
      `,
    },
    {
      code: `
interface F {
  (a?: string | undefined): void;
}
      `,
      errors: [
        {
          column: 17,
          endColumn: 26,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
interface F {
  (a?: string  ): void;
}
      `,
    },
    {
      code: 'type fn = new (a?: string | undefined) => void;',
      errors: [
        {
          column: 29,
          endColumn: 38,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'type fn = new (a?: string  ) => void;',
    },
    {
      code: 'function f(a?: string | undefined) {}',
      errors: [
        {
          column: 25,
          endColumn: 34,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'function f(a?: string  ) {}',
    },
    {
      code: 'f = function (a?: string | undefined) {};',
      errors: [
        {
          column: 28,
          endColumn: 37,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'f = function (a?: string  ) {};',
    },
    {
      code: 'declare function f(a?: string | undefined): void;',
      errors: [
        {
          column: 33,
          endColumn: 42,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'declare function f(a?: string  ): void;',
    },
    {
      code: `
declare class bb {
  f(a?: string | undefined): void;
}
      `,
      errors: [
        {
          column: 18,
          endColumn: 27,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
declare class bb {
  f(a?: string  ): void;
}
      `,
    },
    {
      code: `
interface ee {
  f(a?: string | undefined): void;
}
      `,
      errors: [
        {
          column: 18,
          endColumn: 27,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
interface ee {
  f(a?: string  ): void;
}
      `,
    },
    {
      code: `
interface ee {
  new (a?: string | undefined): void;
}
      `,
      errors: [
        {
          column: 21,
          endColumn: 30,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
interface ee {
  new (a?: string  ): void;
}
      `,
    },
    {
      code: 'type fn = (a?: string | undefined) => void;',
      errors: [
        {
          column: 25,
          endColumn: 34,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'type fn = (a?: string  ) => void;',
    },
    {
      code: 'type fn = (a?: string | (undefined | number)) => void;',
      errors: [
        {
          column: 26,
          endColumn: 35,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'type fn = (a?: string | (  number)) => void;',
    },
    {
      code: 'type fn = (a?: (undefined | number) | string) => void;',
      errors: [
        {
          column: 17,
          endColumn: 26,
          endLine: 1,
          line: 1,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: 'type fn = (a?: (  number) | string) => void;',
    },
    {
      code: `
abstract class cc {
  abstract f(a?: string | undefined): void;
}
      `,
      errors: [
        {
          column: 27,
          endColumn: 36,
          endLine: 3,
          line: 3,
          messageId: 'unnecessary',
          message: 'Explicit undefined is unnecessary on an optional parameter.',
        },
      ],
      output: `
abstract class cc {
  abstract f(a?: string  ): void;
}
      `,
    },
  ],
});
