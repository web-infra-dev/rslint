import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('no-unsafe-assignment', {
  valid: [
    {
      code: "const x = 1;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: number = 1;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\nconst x = 1,\n  y = 1;\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "let x;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\nlet x = 1,\n  y;\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "function foo(a = 1) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\nclass Foo {\n  constructor(private a = 1) {}\n}\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\nclass Foo {\n  private a = 1;\n}\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\nclass Foo {\n  accessor a = 1;\n}\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: Set<string> = new Set();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: Set<string> = new Set<string>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [x] = [1];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [x, y] = [1, 2] as number[];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [x, ...y] = [1, 2, 3, 4, 5];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [x, ...y] = [1];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [{ ...x }] = [{ x: 1 }] as [{ x: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "function foo(x = 1) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "function foo([x] = [1]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "function foo([x, ...y] = [1, 2, 3, 4, 5]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "function foo([x, ...y] = [1]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x = new Set<any>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x = { y: 1 };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x = { y = 1 };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x = { y(){} };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: { y: number } = { y: 1 };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x = [...[1, 2, 3]];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const [{ [`x${1}`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\ntype T = [string, T[]];\nconst test: T = ['string', []] as T;\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\ntype Props = { a: string };\ndeclare function Foo(props: Props): never;\n<Foo a={'foo'} />;\n      ",
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      filename: "react.tsx",
    },
    {
      code: "\ndeclare function Foo(props: { a: string }): never;\n<Foo a=\"foo\" />;\n      ",
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      filename: "react.tsx",
    },
    {
      code: "\ndeclare function Foo(props: { a: string }): never;\n<Foo a={} />;\n      ",
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      filename: "react.tsx",
    },
    {
      code: "const x: unknown = y as any;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: unknown[] = y as any[];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: Set<unknown> = y as Set<any>;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "const x: Map<string, string> = new Map();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
    {
      code: "\ntype Foo = { bar: unknown };\nconst bar: any = 1;\nconst foo: Foo = { bar };\n    ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
    },
  ],
  invalid: [
    {
      code: "const x = 1 as any;",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\nconst x = 1 as any,\n  y = 1;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: "function foo(a = 1 as any) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "\nclass Foo {\n  constructor(private a = 1 as any) {}\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 35,
        },
      ],
    },
    {
      code: "\nclass Foo {\n  private a = 1 as any;\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 24,
        },
      ],
    },
    {
      code: "\nclass Foo {\n  accessor a = 1 as any;\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 25,
        },
      ],
    },
    {
      code: "\nconst [x] = spooky;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an error typed value.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\nconst [[[x]]] = [spooky];\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an error typed value.",
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: "\nconst {\n  x: { y: z },\n} = { x: spooky };\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an error typed value.",
          line: 3,
          column: 6,
          endLine: 3,
          endColumn: 14,
        },
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an error typed value.",
          line: 4,
          column: 7,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\nlet value: number;\n\nvalue = spooky;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an error typed value.",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nconst [x] = 1 as any;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\nconst [x] = [] as any[];\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPattern",
          message: "Unsafe array destructuring of an `any` array value.",
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: "const x: Set<string> = new Set<any>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `Set<any>` to a variable of type `Set<string>`.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: "const x: Map<string, string> = new Map<string, any>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `Map<string, any>` to a variable of type `Map<string, string>`.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: "const x: Set<string[]> = new Set<any[]>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `Set<any[]>` to a variable of type `Set<string[]>`.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: "const x: Set<Set<Set<string>>> = new Set<Set<Set<any>>>();",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    {
      code: "const [x] = [1] as [any];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: "function foo([x] = [1] as [any]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: "[x] = [1] as [any];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: "const [[[[x]]]] = [[[[1 as any]]]];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "function foo([[[[x]]]] = [[[[1 as any]]]]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: "[[[[x]]]] = [[[[1 as any]]]];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: "const [[[[x]]]] = [1 as any];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: "function foo([[[[x]]]] = [1 as any]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "const [{ x }] = [{ x: 1 }] as [{ x: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: "function foo([{ x }] = [{ x: 1 }] as [{ x: any }]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "[{ x }] = [{ x: 1 }] as [{ x: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "const [{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "function foo([{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "[{ ['x']: x }] = [{ ['x']: 1 }] as [{ ['x']: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "const [{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: "function foo([{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }]) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: "[{ [`x`]: x }] = [{ [`x`]: 1 }] as [{ [`x`]: any }];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: "[[[[x]]]] = [1 as any];",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `[any]` to a variable of type `[[[[any]]]]`.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\nconst x = [...(1 as any)];\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArraySpread",
          message: "Unsafe spread of an `any` value in an array.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 25,
        },
      ],
    },
    {
      code: "\nconst x = [...([] as any[])];\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArraySpread",
          message: "Unsafe spread of an `any` value in an array.",
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 28,
        },
      ],
    },
    {
      code: "const { x } = { x: 1 } as { x: any };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: "function foo({ x } = { x: 1 } as { x: any }) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: "({ x } = { x: 1 } as { x: any });",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: "const { x: y } = { x: 1 } as { x: any };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: "function foo({ x: y } = { x: 1 } as { x: any }) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: "({ x: y } = { x: 1 } as { x: any });",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: "\nconst {\n  x: { y },\n} = { x: { y: 1 } } as { x: { y: any } };\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: "function foo({ x: { y } } = { x: { y: 1 } } as { x: { y: any } }) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: "\n({\n  x: { y },\n} = { x: { y: 1 } } as { x: { y: any } });\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeObjectPattern",
          message: "Unsafe object destructuring of a property with an `any` value.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\nconst {\n  x: [y],\n} = { x: { y: 1 } } as { x: [any] };\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: "function foo({ x: [y] } = { x: { y: 1 } } as { x: [any] }) {}",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\n({\n  x: [y],\n} = { x: { y: 1 } } as { x: [any] });\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeArrayPatternFromTuple",
          message: "Unsafe array destructuring of a tuple element with an `any` value.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: "const x = { y: 1 as any };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "const x = { y: { z: 1 as any } };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: "const x: { y: Set<Set<Set<string>>> } = { y: new Set<Set<Set<any>>>() };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "unsafeAssignment",
          message: "Unsafe assignment of type `Set<Set<Set<any>>>` to a variable of type `Set<Set<Set<string>>>`.",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 70,
        },
      ],
    },
    {
      code: "const x = { ...(1 as any) };",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: "\ntype Props = { a: string };\ndeclare function Foo(props: Props): never;\n<Foo a={1 as any} />;\n      ",
      languageOptions: {
        parserOptions: {
          ecmaFeatures: {
            jsx: true,
          },
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      filename: "react.tsx",
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 17,
        },
      ],
    },
    {
      code: "\nfunction foo() {\n  const bar = this;\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignmentThis",
          message: "Unsafe assignment of an `any` value. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\ntype T = [string, T[]];\nconst test: T = ['string', []] as any;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 38,
        },
      ],
    },
    {
      code: "\ntype Foo = { bar: number };\nconst bar: any = 1;\nconst foo: Foo = { bar };\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noImplicitThis.json",
        },
      },
      errors: [
        {
          messageId: "anyAssignment",
          message: "Unsafe assignment of an `any` value.",
          line: 4,
          column: 20,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
  ],
});
