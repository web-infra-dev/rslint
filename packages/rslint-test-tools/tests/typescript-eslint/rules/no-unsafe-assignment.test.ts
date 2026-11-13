import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

new RuleTester({
  ...config,
  languageOptions: {
    ...config.languageOptions,
    parserOptions: {
      ...config.languageOptions?.parserOptions,
      project: './tsconfig.noImplicitThis.json',
    },
  },
}).run('no-unsafe-assignment', {
  valid: [
    'const x = 1;',
    'const x: number = 1;',
    `
const x = 1,
  y = 1;
    `,
    'let x;',
    `
let x = 1,
  y;
    `,
    'function foo(a = 1) {}',
    `
class Foo {
  constructor(private a = 1) {}
}
    `,
    `
class Foo {
  private a = 1;
}
    `,
    `
class Foo {
  accessor a = 1;
}
    `,
    'const x: Set<string> = new Set();',
    'const x: Set<string> = new Set<string>();',
    'const [x] = [1];',
    'const [x, y] = [1, 2] as number[];',
    'const [x, ...y] = [1, 2, 3, 4, 5];',
    'const [x, ...y] = [1];',
    'const [{ ...x }] = [{ x: 1 }] as [{ x: any }];',
    'function foo(x = 1) {}',
    'function foo([x] = [1]) {}',
    'function foo([x, ...y] = [1, 2, 3, 4, 5]) {}',
    'function foo([x, ...y] = [1]) {}',
    // this is not checked, because there's no annotation to compare it with
    'const x = new Set<any>();',
    'const x = { y: 1 };',
    // Invalid TypeScript syntax; typescript-go exposes its error type to the rule.
    'const x = { y = 1 };',
    'const x = { y(){} };',
    'const x: { y: number } = { y: 1 };',
    'const x = [...[1, 2, 3]];',
    'const [{ [`x${1}`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];',
    `
type T = [string, T[]];
const test: T = ['string', []] as T;
    `,
    {
      code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={'foo'} />;
      `,
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      filename: 'react.tsx',
    },
    {
      code: `
declare function Foo(props: { a: string }): never;
<Foo a="foo" />;
      `,
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      filename: 'react.tsx',
    },
    {
      code: `
declare function Foo(props: { a: string }): never;
<Foo a={} />;
      `,
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      filename: 'react.tsx',
    },
    'const x: unknown = y as any;',
    'const x: unknown[] = y as any[];',
    'const x: Set<unknown> = y as Set<any>;',
    // https://github.com/typescript-eslint/typescript-eslint/issues/2109
    'const x: Map<string, string> = new Map();',
    `
type Foo = { bar: unknown };
const bar: any = 1;
const foo: Foo = { bar };
    `,
  ],
  invalid: [
    {
      code: 'const x = 1 as any;',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const x = 1 as any,
  y = 1;
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: 'function foo(a = 1 as any) {}',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: `
class Foo {
  constructor(private a = 1 as any) {}
}
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      code: `
class Foo {
  private a = 1 as any;
}
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: `
class Foo {
  accessor a = 1 as any;
}
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
const [x] = spooky;
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an error typed value.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
const [[[x]]] = [spooky];
      `,
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an error typed value.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: `
const {
  x: { y: z },
} = { x: spooky };
      `,
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an error typed value.',
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 14,
        },
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an error typed value.',
          line: 4,
          column: 7,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: `
let value: number;

value = spooky;
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an error typed value.',
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
const [x] = 1 as any;
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
const [x] = [] as any[];
      `,
      errors: [
        {
          messageId: 'unsafeArrayPattern',
          message: 'Unsafe array destructuring of an `any` array value.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'const x: Set<string> = new Set<any>();',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `Set<any>` to a variable of type `Set<string>`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'const x: Map<string, string> = new Map<string, any>();',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `Map<string, any>` to a variable of type `Map<string, string>`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'const x: Set<string[]> = new Set<any[]>();',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `Set<any[]>` to a variable of type `Set<string[]>`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'const x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>();',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: 'const [x] = [1] as [any];',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'function foo([x] = [1] as [any]) {}',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '[x] = [1] as [any];',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'const [[[[x]]]] = [[[[1 as any]]]];',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'function foo([[[[x]]]] = [[[[1 as any]]]]) {}',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '[[[[x]]]] = [[[[1 as any]]]];',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: 'const [[[[x]]]] = [1 as any];',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: 'function foo([[[[x]]]] = [1 as any]) {}',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: 'const [{ x }] = [{ x: 1 }] as [{ x: any }];',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: 'function foo([{ x }] = [{ x: 1 }] as [{ x: any }]) {}',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '[{ x }] = [{ x: 1 }] as [{ x: any }];',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "const [{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "function foo([{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }]) {}",
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "[{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: 'const [{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'function foo([{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }]) {}',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '[{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    // TS treats the assignment pattern weirdly in this case
    {
      code: '[[[[x]]]] = [1 as any];',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `[any]` to a variable of type `[[[[any]]]]`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: `
const x = [...(1 as any)];
      `,
      errors: [
        {
          messageId: 'unsafeArraySpread',
          message: 'Unsafe spread of an `any` value in an array.',
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: `
const x = [...([] as any[])];
      `,
      errors: [
        {
          messageId: 'unsafeArraySpread',
          message: 'Unsafe spread of an `any` value in an array.',
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 28,
        },
      ],
    },
    {
      code: 'const { x } = { x: 1 } as { x: any };',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: 'function foo({ x } = { x: 1 } as { x: any }) {}',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '({ x } = { x: 1 } as { x: any });',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: 'const { x: y } = { x: 1 } as { x: any };',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'function foo({ x: y } = { x: 1 } as { x: any }) {}',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '({ x: y } = { x: 1 } as { x: any });',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: `
const {
  x: { y },
} = { x: { y: 1 } } as { x: { y: any } };
      `,
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: 'function foo({ x: { y } } = { x: { y: 1 } } as { x: { y: any } }) {}',
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: `
({
  x: { y },
} = { x: { y: 1 } } as { x: { y: any } });
      `,
      errors: [
        {
          messageId: 'unsafeObjectPattern',
          message:
            'Unsafe object destructuring of a property with an `any` value.',
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: `
const {
  x: [y],
} = { x: { y: 1 } } as { x: [any] };
      `,
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'function foo({ x: [y] } = { x: { y: 1 } } as { x: [any] }) {}',
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: `
({
  x: [y],
} = { x: { y: 1 } } as { x: [any] });
      `,
      errors: [
        {
          messageId: 'unsafeArrayPatternFromTuple',
          message:
            'Unsafe array destructuring of a tuple element with an `any` value.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'const x = { y: 1 as any };',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'const x = { y: { z: 1 as any } };',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'const x: { y: Set<Set<Set<string>>> } = { y: new Set<Set<Set<any>>>() };',
      errors: [
        {
          messageId: 'unsafeAssignment',
          message:
            'Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    // spreading an any widens the object type to any
    {
      code: 'const x = { ...(1 as any) };',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: `
type Props = { a: string };
declare function Foo(props: Props): never;
<Foo a={1 as any} />;
      `,
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
        },
      },
      filename: 'react.tsx',
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
function foo() {
  const bar = this;
}
      `,
      errors: [
        {
          messageId: 'anyAssignmentThis',
          message:
            'Unsafe assignment of an `any` value. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: `
type T = [string, T[]];
const test: T = ['string', []] as any;
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      code: `
type Foo = { bar: number };
const bar: any = 1;
const foo: Foo = { bar };
      `,
      errors: [
        {
          messageId: 'anyAssignment',
          message: 'Unsafe assignment of an `any` value.',
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
  ],
});
