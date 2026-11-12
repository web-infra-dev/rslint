import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('no-unnecessary-condition', {
  valid: [
    {
      code: "\ndeclare const b1: boolean;\ndeclare const b2: boolean;\nconst t1 = b1 && b2;\nconst t2 = b1 || b2;\nif (b1 && b2) {\n}\nwhile (b1 && b2) {}\nfor (let i = 0; b1 && b2; i++) {\n  break;\n}\nconst t1 = b1 && b2 ? 'yes' : 'no';\nif (b1 && b2) {\n}\nwhile (b1 && b2) {}\nfor (let i = 0; b1 && b2; i++) {\n  break;\n}\nconst t1 = b1 && b2 ? 'yes' : 'no';\nfor (;;) {}\nswitch (b1) {\n  case true:\n  default:\n}\n    ",
    },
    {
      code: "\ndeclare function foo(): number | void;\nconst result1 = foo() === undefined;\nconst result2 = foo() == null;\n    ",
    },
    {
      code: "\ndeclare function foo(): number | void;\nfoo() ?? 1;\n    ",
    },
    {
      code: "\ndeclare const bigInt: 0n | 1n;\nif (bigInt) {\n}\n    ",
    },
    {
      code: `
declare const b1: false | 5;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: boolean | "foo";
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: 0 | boolean;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: boolean | object;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: false | object;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: null | object;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: undefined | true;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: void | true;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & {};
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: number & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: boolean & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: bigint & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & {} & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & { __brandA: string } & { __brandB: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & { __brand: string } | number;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: (string | number) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & ({ __brand: string } | number);
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: ("" | "foo") & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: (string & { __brandA: string }) | (number & { __brandB: string });
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string }) & ("" | "foo"));
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: { __brandA: string} & (({ __brandB: string } & string) | ({ __brandC: string } & number));
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: (string | number) & ("foo" | 123 | { __brandA: string });
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: string & string;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: any;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: `
declare const b1: unknown;
declare const b2: boolean;
const t1 = b1 && b2;
`,
    },
    {
      code: "\nfunction test<T extends string>(t: T) {\n  return t ? 'yes' : 'no';\n}\n    ",
    },
    {
      code: "\n// Naked type param\nfunction test<T>(t: T) {\n  return t ? 'yes' : 'no';\n}\n    ",
    },
    {
      code: "\n// Naked type param in union\nfunction test<T>(t: T | []) {\n  return t ? 'yes' : 'no';\n}\n    ",
    },
    {
      code: "\nfunction test<T>(arg: T, key: keyof T) {\n  if (arg[key]?.toString()) {\n  }\n}\n    ",
    },
    {
      code: "\nfunction test<T>(arg: T, key: keyof T) {\n  if (arg?.toString()) {\n  }\n}\n    ",
    },
    {
      code: "\nfunction test<T>(arg: T | { value: string }) {\n  if (arg?.value) {\n  }\n}\n    ",
    },
    {
      code: "\nfunction test(a: string) {\n  const t1 = a === 'a';\n  const t2 = 'a' === a;\n}\n    ",
    },
    {
      code: "\nfunction test(a?: string) {\n  const t1 = a === undefined;\n  const t2 = undefined === a;\n  const t1 = a !== undefined;\n  const t2 = undefined !== a;\n}\n    ",
    },
    {
      code: "\nfunction test(a: null | string) {\n  const t1 = a === null;\n  const t2 = null === a;\n  const t1 = a !== null;\n  const t2 = null !== a;\n}\n    ",
    },
    {
      code: "\nfunction test(a?: null | string) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n}\n    ",
    },
    {
      code: "\nfunction test(a?: string) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n}\n    ",
    },
    {
      code: "\nfunction test(a: null | string) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n}\n    ",
    },
    {
      code: "\nfunction test(a: any) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n  const t9 = a === null;\n  const t10 = null === a;\n  const t11 = a !== null;\n  const t12 = null !== a;\n  const t13 = a === undefined;\n  const t14 = undefined === a;\n  const t15 = a !== undefined;\n  const t16 = undefined !== a;\n}\n    ",
    },
    {
      code: "\nfunction test(a: unknown) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n  const t9 = a === null;\n  const t10 = null === a;\n  const t11 = a !== null;\n  const t12 = null !== a;\n  const t13 = a === undefined;\n  const t14 = undefined === a;\n  const t15 = a !== undefined;\n  const t16 = undefined !== a;\n}\n    ",
    },
    {
      code: "\nfunction test<T>(a: T) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n  const t9 = a === null;\n  const t10 = null === a;\n  const t11 = a !== null;\n  const t12 = null !== a;\n  const t13 = a === undefined;\n  const t14 = undefined === a;\n  const t15 = a !== undefined;\n  const t16 = undefined !== a;\n}\n    ",
    },
    {
      code: "\nfunction foo<T extends object>(arg: T, key: keyof T): void {\n  arg[key] == null;\n}\n    ",
    },
    {
      code: "\n// with literal arrow function\n[0, 1, 2].filter(x => x);\n\n// filter with named function\nfunction length(x: string) {\n  return x.length;\n}\n['a', 'b', ''].filter(length);\n\n// with non-literal array\nfunction nonEmptyStrings(x: string[]) {\n  return x.filter(length);\n}\n\n// filter-like predicate\nfunction count(\n  list: string[],\n  predicate: (value: string, index: number, array: string[]) => unknown,\n) {\n  return list.filter(predicate).length;\n}\n    ",
    },
    {
      code: "\ndeclare const test: <T>() => T;\n\n[1, null].filter(test);\n    ",
    },
    {
      code: "\ndeclare const test: <T extends boolean>() => T;\n\n[1, null].filter(test);\n    ",
    },
    {
      code: "\n[1, null].filter(1 as any);\n    ",
    },
    {
      code: "\n[1, null].filter(1 as never);\n    ",
    },
    {
      code: "\nconst notArray = {\n  filter: (func: () => boolean) => func(),\n  find: (func: () => boolean) => func(),\n};\nnotArray.filter(() => true);\nnotArray.find(() => true);\n    ",
    },
    {
      code: "\nfunction test(a: string | null) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction test(a: string | undefined) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction test(a: string | null | undefined) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction test(a: unknown) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction test<T>(a: T) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction test<T extends string | null>(a: T) {\n  return a ?? 'default';\n}\n    ",
    },
    {
      code: "\nfunction foo<T extends object>(arg: T, key: keyof T): void {\n  arg[key] ?? 'default';\n}\n    ",
    },
    {
      code: "\ndeclare const arr: object[];\nif (arr[42]) {\n} // looks unnecessary from the types, but isn't\n\nconst tuple = [{}] as [object];\ndeclare const n: number;\nif (tuple[n]) {\n}\n    ",
    },
    {
      code: "\ndeclare const arr: Array<{ value: string } & (() => void)>;\nif (arr[42]?.value) {\n}\narr[41]?.();\n\n// An array access can \"infect\" deeper into the chain\ndeclare const arr2: Array<{ x: { y: { z: object } } }>;\narr2[42]?.x?.y?.z;\n\nconst tuple = ['foo'] as const;\ndeclare const n: number;\ntuple[n]?.toUpperCase();\n    ",
    },
    {
      code: "\ndeclare const arr: Array<{ value: string } & (() => void)>;\nif (arr[42]?.value) {\n}\narr[41]?.();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\nif (arr?.[42]) {\n}\n    ",
    },
    {
      code: "\ntype ItemA = { bar: string; baz: string };\ntype ItemB = { bar: string; qux: string };\ndeclare const foo: ItemA[] | ItemB[];\nfoo[0]?.bar;\n    ",
    },
    {
      code: "\ntype TupleA = [string, number];\ntype TupleB = [string, number];\n\ndeclare const foo: TupleA | TupleB;\ndeclare const index: number;\nfoo[index]?.toString();\n    ",
    },
    {
      code: "\ntype TupleA = [string, number];\ntype TupleB = [string, number];\n\ndeclare const foo: TupleA | TupleB;\ndeclare const index: number;\nfoo[index]?.toString();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ndeclare const returnsArr: undefined | (() => string[]);\nif (returnsArr?.()[42]) {\n}\nreturnsArr?.()[42]?.toUpperCase();\n    ",
    },
    {
      code: "\ndeclare const arr: string[][];\narr[x] ?? [];\n    ",
    },
    {
      code: "\ndeclare const arr: { foo: number }[];\nconst bar = arr[42]?.foo ?? 0;\n    ",
    },
    {
      code: "\ndeclare const b1: boolean;\ndeclare const b2: true;\nconst x = b1 && b2;\n      ",
    },
    {
      code: "\nwhile (true) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: "\nfor (; true; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: "\ndo {} while (true);\n      ",
      options: [
        {
          allowConstantLoopConditions: true,
        },
      ],
    },
    {
      code: "\nwhile (true) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "always",
        },
      ],
    },
    {
      code: "\nfor (; true; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "always",
        },
      ],
    },
    {
      code: "\nfor (; true; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\nfor (; 0; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\ndo {} while (0);\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\ndo {} while (true);\n      ",
      options: [
        {
          allowConstantLoopConditions: "always",
        },
      ],
    },
    {
      code: "\nwhile (true) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\nwhile (1) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\nwhile (false) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\nwhile (0) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
    },
    {
      code: "\nlet variable = 'abc' as string | void;\nvariable?.[0];\n    ",
    },
    {
      code: "\nlet foo: undefined | { bar: true };\nfoo?.bar;\n    ",
    },
    {
      code: "\nlet foo: null | { bar: true };\nfoo?.bar;\n    ",
    },
    {
      code: "\nlet foo: undefined;\nfoo?.bar;\n    ",
    },
    {
      code: "\nlet foo: undefined;\nfoo?.bar.baz;\n    ",
    },
    {
      code: "\nlet foo: null;\nfoo?.bar;\n    ",
    },
    {
      code: "\nlet anyValue: any;\nanyValue?.foo;\n    ",
    },
    {
      code: "\nlet unknownValue: unknown;\nunknownValue?.foo;\n    ",
    },
    {
      code: "\nlet foo: undefined | (() => {});\nfoo?.();\n    ",
    },
    {
      code: "\nlet foo: null | (() => {});\nfoo?.();\n    ",
    },
    {
      code: "\nlet foo: undefined;\nfoo?.();\n    ",
    },
    {
      code: "\nlet foo: undefined;\nfoo?.().bar;\n    ",
    },
    {
      code: "\nlet foo: null;\nfoo?.();\n    ",
    },
    {
      code: "\nlet anyValue: any;\nanyValue?.();\n    ",
    },
    {
      code: "\nlet unknownValue: unknown;\nunknownValue?.();\n    ",
    },
    {
      code: "const foo = [1, 2, 3][0];",
    },
    {
      code: "\ndeclare const foo: { bar?: { baz: { c: string } } } | null;\nfoo?.bar?.baz;\n    ",
    },
    {
      code: "\nfoo?.bar?.baz?.qux;\n    ",
    },
    {
      code: "\ndeclare const foo: { bar: { baz: string } };\nfoo.bar.qux?.();\n    ",
    },
    {
      code: "\ntype Foo = { baz: number } | null;\ntype Bar = { baz: null | string | { qux: string } };\ndeclare const foo: { fooOrBar: Foo | Bar } | null;\nfoo?.fooOrBar?.baz?.qux;\n    ",
    },
    {
      code: "\ntype Foo = { [key: string]: string } | null;\ndeclare const foo: Foo;\n\nconst key = '1';\nfoo?.[key]?.trim();\n    ",
    },
    {
      code: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = 'bar' | 'foo';\ndeclare const foo: Foo;\ndeclare const key: Key;\n\nfoo?.[key].trim();\n    ",
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n  };\n}\n\nfunction Foo(outer: Outer, key: string): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n    ",
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n    bar: 'bar';\n  };\n}\ntype Foo = 'foo';\n\nfunction Foo(outer: Outer, key: Foo): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n    ",
    },
    {
      code: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = 'bar' | 'foo' | 'baz';\ndeclare const foo: Foo;\ndeclare const key: Key;\n\nfoo?.[key]?.trim();\n    ",
    },
    {
      code: "\ntype BrandedKey = string & { __brand: string };\ntype Foo = { [key: BrandedKey]: string } | null;\ndeclare const foo: Foo;\nconst key = '1' as BrandedKey;\nfoo?.[key]?.trim();\n    ",
    },
    {
      code: "\ntype BrandedKey<S extends string> = S & { __brand: string };\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = BrandedKey<'bar'> | BrandedKey<'foo'>;\ndeclare const foo: Foo;\ndeclare const key: Key;\nfoo?.[key].trim();\n    ",
    },
    {
      code: "\ntype BrandedKey = string & { __brand: string };\ninterface Outer {\n  inner?: {\n    [key: BrandedKey]: string | undefined;\n  };\n}\nfunction Foo(outer: Outer, key: BrandedKey): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n    ",
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string & { __brand: string }]: string | undefined;\n    bar: 'bar';\n  };\n}\ntype Foo = 'foo' & { __brand: string };\nfunction Foo(outer: Outer, key: Foo): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n    ",
    },
    {
      code: "\ntype BrandedKey<S extends string> = S & { __brand: string };\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;\ndeclare const foo: Foo;\ndeclare const key: Key;\nfoo?.[key]?.trim();\n    ",
    },
    {
      code: "\ntype BrandedKey = string & { __brand: string };\ntype Foo = { [key: BrandedKey]: string } | null;\ndeclare const foo: Foo;\nconst key = '1' as BrandedKey;\nfoo?.[key]?.trim();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype BrandedKey<S extends string> = S & { __brand: string };\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = BrandedKey<'bar'> | BrandedKey<'foo'>;\ndeclare const foo: Foo;\ndeclare const key: Key;\nfoo?.[key].trim();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype BrandedKey = string & { __brand: string };\ninterface Outer {\n  inner?: {\n    [key: BrandedKey]: string | undefined;\n  };\n}\nfunction Foo(outer: Outer, key: BrandedKey): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string & { __brand: string }]: string | undefined;\n    bar: 'bar';\n  };\n}\ntype Foo = 'foo' & { __brand: string };\nfunction Foo(outer: Outer, key: Foo): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype BrandedKey<S extends string> = S & { __brand: string };\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = BrandedKey<'bar'> | BrandedKey<'foo'> | BrandedKey<'baz'>;\ndeclare const foo: Foo;\ndeclare const key: Key;\nfoo?.[key]?.trim();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype Foo = {\n  key?: Record<string, { key: string }>;\n};\ndeclare const foo: Foo;\nfoo.key?.someKey?.key;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype Foo = {\n  key?: {\n    [key: string]: () => void;\n  };\n};\ndeclare const foo: Foo;\nfoo.key?.value?.();\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype A = {\n  [name in Lowercase<string>]?: {\n    [name in Lowercase<string>]: {\n      a: 1;\n    };\n  };\n};\n\ndeclare const a: A;\n\na.a?.a?.a;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\nlet latencies: number[][] = [];\n\nfunction recordData(): void {\n  if (!latencies[0]) latencies[0] = [];\n  latencies[0].push(4);\n}\n\nrecordData();\n    ",
    },
    {
      code: "\nlet latencies: number[][] = [];\n\nfunction recordData(): void {\n  if (latencies[0]) latencies[0] = [];\n  latencies[0].push(4);\n}\n\nrecordData();\n    ",
    },
    {
      code: "\nfunction test(testVal?: boolean) {\n  if (testVal ?? true) {\n    console.log('test');\n  }\n}\n    ",
    },
    {
      code: "\ndeclare const x: string[];\nif (!x[0]) {\n}\n    ",
    },
    {
      code: "\nconst isEven = (val: number) => val % 2 === 0;\nif (!isEven(1)) {\n}\n    ",
    },
    {
      code: "\ndeclare const booleanTyped: boolean;\ndeclare const unknownTyped: unknown;\n\nif (!(booleanTyped || unknownTyped)) {\n}\n    ",
    },
    {
      code: "\ndeclare const x: string[] | null;\n// eslint-disable-next-line\nif (x) {\n}\n      ",
      options: [
        {
          allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.unstrict.json",
        },
      },
    },
    {
      code: "\ninterface Foo {\n  [key: string]: [string] | undefined;\n}\n\ntype OptionalFoo = Foo | undefined;\ndeclare const foo: OptionalFoo;\nfoo?.test?.length;\n    ",
    },
    {
      code: "\ninterface Foo {\n  [key: number]: [string] | undefined;\n}\n\ntype OptionalFoo = Foo | undefined;\ndeclare const foo: OptionalFoo;\nfoo?.[1]?.length;\n    ",
    },
    {
      code: "\ndeclare let foo: number | null;\nfoo ??= 1;\n    ",
    },
    {
      code: "\ndeclare let foo: number;\nfoo ||= 1;\n    ",
    },
    {
      code: "\ndeclare const foo: { bar: { baz?: number; qux: number } };\ntype Key = 'baz' | 'qux';\ndeclare const key: Key;\nfoo.bar[key] ??= 1;\n    ",
    },
    {
      code: "\nenum Keys {\n  A = 'A',\n  B = 'B',\n}\ntype Foo = {\n  [Keys.A]: number | null;\n  [Keys.B]: number;\n};\ndeclare const foo: Foo;\ndeclare const key: Keys;\nfoo[key] ??= 1;\n    ",
    },
    {
      code: "\ndeclare const foo: { bar?: number };\nfoo.bar ??= 1;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.exactOptionalPropertyTypes.json",
        },
      },
    },
    {
      code: "\ndeclare const foo: { bar: { baz?: number } };\nfoo['bar'].baz ??= 1;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.exactOptionalPropertyTypes.json",
        },
      },
    },
    {
      code: "\ndeclare const foo: { bar: { baz?: number; qux: number } };\ntype Key = 'baz' | 'qux';\ndeclare const key: Key;\nfoo.bar[key] ??= 1;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.exactOptionalPropertyTypes.json",
        },
      },
    },
    {
      code: "\ndeclare let foo: number;\nfoo &&= 1;\n    ",
    },
    {
      code: "\nfunction foo<T extends object>(arg: T, key: keyof T): void {\n  arg[key] ??= 'default';\n}\n    ",
    },
    {
      code: "\nfunction get<Obj, Key extends keyof Obj>(obj: Obj, key: Key) {\n  const value = obj[key];\n  if (value) {\n    return value;\n  }\n  throw new Error('BOOM!');\n}\n\nget({ foo: null }, 'foo');\n    ",
    },
    {
      code: "\nfunction getElem(dict: Record<string, { foo: string }>, key: string) {\n  if (dict[key]) {\n    return dict[key].foo;\n  } else {\n    return '';\n  }\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
    },
    {
      code: "\ntype Foo = { bar: () => number | undefined } | null;\ndeclare const foo: Foo;\nfoo?.bar()?.toExponential();\n    ",
    },
    {
      code: "\ntype Foo = (() => number | undefined) | null;\ndeclare const foo: Foo;\nfoo?.()?.toExponential();\n    ",
    },
    {
      code: "\ntype FooUndef = () => undefined;\ntype FooNum = () => number;\ntype Foo = FooUndef | FooNum | null;\ndeclare const foo: Foo;\nfoo?.()?.toExponential();\n    ",
    },
    {
      code: "\ntype Foo = { [key: string]: () => number | undefined } | null;\ndeclare const foo: Foo;\nfoo?.['bar']()?.toExponential();\n    ",
    },
    {
      code: "\ndeclare function foo(): void | { key: string };\nconst bar = foo()?.key;\n    ",
    },
    {
      code: "\ntype fn = () => void;\ndeclare function foo(): void | fn;\nconst bar = foo()?.();\n    ",
    },
    {
      code: "\ndeclare function maybe<T>(v: T): T | void;\nconst a = maybe({ key1: { key2: maybe({ i: 1 }) } })?.key1.key2?.i;\n    ",
    },
    {
      code: "\ntype Foo = { bar: { baz: number } | void } | null;\ndeclare const foo: Foo;\nfoo?.bar?.baz;\n    ",
    },
    {
      code: "\ndeclare function maybeFn(): { fn: (() => number) | void } | undefined;\nmaybeFn()?.fn?.();\n    ",
    },
    {
      code: "\nclass ConsistentRand {\n  #rand?: number;\n\n  getCachedRand() {\n    this.#rand ??= Math.random();\n    return this.#rand;\n  }\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.exactOptionalPropertyTypes.json",
        },
      },
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\n\nassert(Math.random() > 0.5);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown, y: unknown): asserts x;\n\nassert(Math.random() > 0.5, true);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert(true);\n      ",
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    {
      code: "\nclass ThisAsserter {\n  assertThis(this: unknown, arg2: unknown): asserts this {}\n}\n\nconst thisAsserter: ThisAsserter = new ThisAsserter();\nthisAsserter.assertThis(true);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\nclass ThisAsserter {\n  assertThis(this: unknown, arg2: unknown): asserts this {}\n}\n\nconst thisAsserter: ThisAsserter = new ThisAsserter();\nthisAsserter.assertThis(Math.random());\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert(...[]);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert(...[], {});\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function assertString(x: unknown): asserts x is string;\ndeclare const a: string;\nassertString(a);\n      ",
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    {
      code: "\ndeclare function isString(x: unknown): x is string;\ndeclare const a: string;\nisString(a);\n      ",
      options: [
        {
          checkTypePredicates: false,
        },
      ],
    },
    {
      code: "\ndeclare function assertString(x: unknown): asserts x is string;\nassertString('falafel');\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function isString(x: unknown): x is string;\nisString('falafel');\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ndeclare function isStringOrNumber(x: unknown): x is string | number;\ndeclare const s: string | number | bigint;\nif (isStringOrNumber(s)) {\n}\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ninterface Wider {\n  a: string;\n}\ninterface Narrower {\n  a: string;\n  b: number;\n}\ndeclare function isNarrower(x: unknown): x is Narrower;\ndeclare const w: Wider;\nif (isNarrower(w)) {\n}\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
    },
    {
      code: "\ntype A = { [name in Lowercase<string>]?: A };\ndeclare const a: A;\na.a?.a?.a;\n    ",
    },
    {
      code: "\ninterface T {\n  [name: Lowercase<string>]: {\n    [name: Lowercase<string>]: {\n      [name: Lowercase<string>]: {\n        value: 'value';\n      };\n    };\n  };\n  [name: Uppercase<string>]: null | {\n    [name: Uppercase<string>]: null | {\n      [name: Uppercase<string>]: null | {\n        VALUE: 'VALUE';\n      };\n    };\n  };\n}\n\ndeclare const t: T;\n\nt.a.a.a.value;\nt.A?.A?.A?.VALUE;\n    ",
    },
  ],
  invalid: [
    {
      code: "\nconst b1 = true;\ndeclare const b2: boolean;\nconst t1 = b1 && b2;\nconst t2 = b1 || b2;\nif (b1 && b2) {\n}\nif (b2 && b1) {\n}\nwhile (b1 && b2) {}\nwhile (b2 && b1) {}\nfor (let i = 0; b1 && b2; i++) {\n  break;\n}\nconst t1 = b1 && b2 ? 'yes' : 'no';\nconst t1 = b2 && b1 ? 'yes' : 'no';\nswitch (b1) {\n  case true:\n  default:\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 5,
          column: 12,
          endLine: 5,
          endColumn: 14,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 6,
          column: 5,
          endLine: 6,
          endColumn: 7,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 8,
          column: 11,
          endLine: 8,
          endColumn: 13,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 10,
          column: 8,
          endLine: 10,
          endColumn: 10,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 11,
          column: 14,
          endLine: 11,
          endColumn: 16,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 12,
          column: 17,
          endLine: 12,
          endColumn: 19,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 15,
          column: 12,
          endLine: 15,
          endColumn: 14,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 16,
          column: 18,
          endLine: 16,
          endColumn: 20,
        },
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `true === true` is true.",
          line: 18,
          column: 8,
          endLine: 18,
          endColumn: 12,
        },
      ],
    },
    {
      code: `
declare const b1: object;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: object | true;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: "" | false;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: "always truthy";
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: undefined;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: null;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: void;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: never;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "never",
          message: "Unnecessary conditional, value is `never`.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: string & number;
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "never",
          message: "Unnecessary conditional, value is `never`.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ndeclare const falseyBigInt: 0n;\nif (falseyBigInt) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 17,
        },
      ],
    },
    {
      code: "\ndeclare const posbigInt: 1n;\nif (posbigInt) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ndeclare const negBigInt: -2n;\nif (negBigInt) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ndeclare const b1: boolean;\ndeclare const b2: boolean;\nif (true && b1 && b2) {\n}\nif (b1 && false && b2) {\n}\nif (b1 || b2 || true) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 9,
        },
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 16,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 8,
          column: 17,
          endLine: 8,
          endColumn: 21,
        },
      ],
    },
    {
      code: "\nfunction test<T extends object>(t: T) {\n  return t ? 'yes' : 'no';\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test<T extends false>(t: T) {\n  return t ? 'yes' : 'no';\n}\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test<T extends 'a' | 'b'>(t: T) {\n  return t ? 'yes' : 'no';\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test(a: 'a') {\n  return a === 'a';\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `\"a\" === \"a\"` is true.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\ndeclare const a: '34';\ndeclare const b: '56';\na > b;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `\"34\" > \"56\"` is false.",
          line: 4,
          column: 1,
          endLine: 4,
          endColumn: 6,
        },
      ],
    },
    {
      code: "\nconst y = 1;\nif (y === 0) {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `1 === 0` is false.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\n// @ts-expect-error\nif (1 == '1') {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `1 == \"1\"` is true.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: "\n2.3 > 2.3;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `2.3 > 2.3` is false.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    {
      code: "\n2.3 >= 2.3;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `2.3 >= 2.3` is true.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n2n < 2n;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `2n < 2n` is false.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 8,
        },
      ],
    },
    {
      code: "\n2n <= 2n;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `2n <= 2n` is true.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\n-2n !== 2n;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `-2n !== 2n` is true.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n// @ts-expect-error\nif (1 == '2') {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `1 == \"2\"` is false.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: "\n// @ts-expect-error\nif (1 != '2') {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `1 != \"2\"` is true.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: "\nenum Foo {\n  a = 1,\n  b = 2,\n}\n\nconst x = Foo.a;\nif (x === Foo.a) {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `Foo.a === Foo.a` is true.",
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\nenum Foo {\n  a = 1,\n  b = 2,\n}\n\nconst x = Foo.a;\nif (x === 1) {\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `Foo.a === 1` is true.",
          line: 8,
          column: 5,
          endLine: 8,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\nfunction takesMaybeValue(a: null | object) {\n  if (a) {\n  } else if (a == undefined) {\n  }\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `null == undefined` is true.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    {
      code: "\nfunction takesMaybeValue(a: null | object) {\n  if (a) {\n  } else if (a === undefined) {\n  }\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `null === undefined` is false.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: "\nfunction takesMaybeValue(a: null | object) {\n  if (a) {\n  } else if (a != undefined) {\n  }\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `null != undefined` is false.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
    {
      code: "\nfunction takesMaybeValue(a: null | object) {\n  if (a) {\n  } else if (a !== undefined) {\n  }\n}\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `null !== undefined` is true.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
      ],
    },
    {
      code: "\ntrue === false;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `true === false` is false.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\ntrue === true;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always true, since `true === true` is true.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ntrue === undefined;\n      ",
      errors: [
        {
          messageId: "comparisonBetweenLiteralTypes",
          message: "Unnecessary conditional, comparison is always false, since `true === undefined` is false.",
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\nfunction test(a: string) {\n  const t1 = a === undefined;\n  const t2 = undefined === a;\n  const t3 = a !== undefined;\n  const t4 = undefined !== a;\n  const t5 = a === null;\n  const t6 = null === a;\n  const t7 = a !== null;\n  const t8 = null !== a;\n}\n      ",
      errors: [
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 24,
        },
      ],
    },
    {
      code: "\nfunction test(a?: string) {\n  const t1 = a === undefined;\n  const t2 = undefined === a;\n  const t3 = a !== undefined;\n  const t4 = undefined !== a;\n  const t5 = a === null;\n  const t6 = null === a;\n  const t7 = a !== null;\n  const t8 = null !== a;\n}\n      ",
      errors: [
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 24,
        },
      ],
    },
    {
      code: "\nfunction test(a: null | string) {\n  const t1 = a === undefined;\n  const t2 = undefined === a;\n  const t3 = a !== undefined;\n  const t4 = undefined !== a;\n  const t5 = a === null;\n  const t6 = null === a;\n  const t7 = a !== null;\n  const t8 = null !== a;\n}\n      ",
      errors: [
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 29,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 29,
        },
      ],
    },
    {
      code: "\nfunction test<T extends object>(a: T) {\n  const t1 = a == null;\n  const t2 = null == a;\n  const t3 = a != null;\n  const t4 = null != a;\n  const t5 = a == undefined;\n  const t6 = undefined == a;\n  const t7 = a != undefined;\n  const t8 = undefined != a;\n  const t9 = a === null;\n  const t10 = null === a;\n  const t11 = a !== null;\n  const t12 = null !== a;\n  const t13 = a === undefined;\n  const t14 = undefined === a;\n  const t15 = a !== undefined;\n  const t16 = undefined !== a;\n}\n      ",
      errors: [
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 23,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 23,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 23,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 6,
          column: 14,
          endLine: 6,
          endColumn: 23,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 7,
          column: 14,
          endLine: 7,
          endColumn: 28,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 8,
          column: 14,
          endLine: 8,
          endColumn: 28,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 9,
          column: 14,
          endLine: 9,
          endColumn: 28,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 10,
          column: 14,
          endLine: 10,
          endColumn: 28,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 11,
          column: 14,
          endLine: 11,
          endColumn: 24,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 12,
          column: 15,
          endLine: 12,
          endColumn: 25,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 13,
          column: 15,
          endLine: 13,
          endColumn: 25,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 14,
          column: 15,
          endLine: 14,
          endColumn: 25,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 15,
          column: 15,
          endLine: 15,
          endColumn: 30,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 16,
          column: 15,
          endLine: 16,
          endColumn: 30,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 17,
          column: 15,
          endLine: 17,
          endColumn: 30,
        },
        {
          messageId: "noOverlapBooleanExpression",
          message: "Unnecessary conditional, the types have no overlap.",
          line: 18,
          column: 15,
          endLine: 18,
          endColumn: 30,
        },
      ],
    },
    {
      code: "\nfunction test(a: string) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test(a: string | false) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test<T extends string>(a: T) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test(a: { foo: string }[]) {\n  return a[0].foo ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\nfunction test(a: null) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "alwaysNullish",
          message: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test(a: null[]) {\n  return a[0] ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "alwaysNullish",
          message: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\nfunction test<T extends null>(a: T) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "alwaysNullish",
          message: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test(a: never) {\n  return a ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "never",
          message: "Unnecessary conditional, value is `never`.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\nfunction test<T extends { foo: number }, K extends 'foo'>(num: T[K]) {\n  num ?? 'default';\n}\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: "\n[1, 3, 5].filter(() => true);\n[1, 2, 3].find(() => {\n  return false;\n});\n\n// with non-literal array\nfunction nothing(x: string[]) {\n  return x.filter(() => false);\n}\n// with readonly array\nfunction nothing2(x: readonly string[]) {\n  return x.filter(() => false);\n}\n// with tuple\nfunction nothing3(x: [string, string]) {\n  return x.filter(() => false);\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 28,
        },
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 15,
        },
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 9,
          column: 25,
          endLine: 9,
          endColumn: 30,
        },
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 13,
          column: 25,
          endLine: 13,
          endColumn: 30,
        },
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 17,
          column: 25,
          endLine: 17,
          endColumn: 30,
        },
      ],
    },
    {
      code: "\ndeclare const test: <T extends true>() => T;\n\n[1, null].filter(test);\n      ",
      errors: [
        {
          messageId: "alwaysTruthyFunc",
          message: "This callback should return a conditional, but return is always truthy.",
          line: 4,
          column: 18,
          endLine: 4,
          endColumn: 22,
        },
      ],
    },
    {
      code: "\ndeclare const dict: Record<string, object>;\nif (dict['mightNotExist']) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 26,
        },
      ],
    },
    {
      code: "\nconst x = [{}] as [{ foo: string }];\nif (x[0]) {\n}\nif (x[0]?.foo) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 9,
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nconst x = [{}] as [{ foo: string }];\nif (x[0]) {\n}\nif (x[0].foo) {\n}\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const arr: object[];\nif (arr.filter) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nfunction truthy() {\n  return [];\n}\nfunction falsy() {}\n[1, 3, 5].filter(truthy);\n[1, 2, 3].find(falsy);\n[1, 2, 3].findLastIndex(falsy);\n      ",
      errors: [
        {
          messageId: "alwaysTruthyFunc",
          message: "This callback should return a conditional, but return is always truthy.",
          line: 6,
          column: 18,
          endLine: 6,
          endColumn: 24,
        },
        {
          messageId: "alwaysFalsyFunc",
          message: "This callback should return a conditional, but return is always falsy.",
          line: 7,
          column: 16,
          endLine: 7,
          endColumn: 21,
        },
        {
          messageId: "alwaysFalsyFunc",
          message: "This callback should return a conditional, but return is always falsy.",
          line: 8,
          column: 25,
          endLine: 8,
          endColumn: 30,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nwhile (test) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nfor (; test; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\ndo {} while (test);\n      ",
      options: [
        {
          allowConstantLoopConditions: false,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nwhile (test) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "never",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nfor (; test; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "never",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\ndo {} while (test);\n      ",
      options: [
        {
          allowConstantLoopConditions: "never",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nwhile (test) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: 1;\n\nwhile (test) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\nfor (; test; ) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const test: true;\n\ndo {} while (test);\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 18,
        },
      ],
    },
    {
      code: "\nlet shouldRun = true;\n\nwhile ((shouldRun = true)) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 25,
        },
      ],
    },
    {
      code: "\nwhile (2) {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\nwhile ('truthy') {}\n      ",
      options: [
        {
          allowConstantLoopConditions: "only-allowed-literals",
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\nlet foo = { bar: true };\nfoo?.bar;\nfoo ?. bar;\nfoo ?.\n  bar;\nfoo\n  ?. bar;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = { bar: true };\nfoo.bar;\nfoo ?. bar;\nfoo ?.\n  bar;\nfoo\n  ?. bar;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = { bar: true };\nfoo?.bar;\nfoo . bar;\nfoo ?.\n  bar;\nfoo\n  ?. bar;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = { bar: true };\nfoo?.bar;\nfoo ?. bar;\nfoo .\n  bar;\nfoo\n  ?. bar;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = { bar: true };\nfoo?.bar;\nfoo ?. bar;\nfoo ?.\n  bar;\nfoo\n  . bar;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\nlet foo = () => {};\nfoo?.();\nfoo ?. ();\nfoo ?.\n  ();\nfoo\n  ?. ();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo();\nfoo ?. ();\nfoo ?.\n  ();\nfoo\n  ?. ();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.();\nfoo  ();\nfoo ?.\n  ();\nfoo\n  ?. ();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.();\nfoo ?. ();\nfoo \n  ();\nfoo\n  ?. ();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.();\nfoo ?. ();\nfoo ?.\n  ();\nfoo\n   ();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\nlet foo = () => {};\nfoo?.(bar);\nfoo ?. (bar);\nfoo ?.\n  (bar);\nfoo\n  ?. (bar);\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo(bar);\nfoo ?. (bar);\nfoo ?.\n  (bar);\nfoo\n  ?. (bar);\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.(bar);\nfoo  (bar);\nfoo ?.\n  (bar);\nfoo\n  ?. (bar);\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 5,
          endLine: 5,
          endColumn: 7,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.(bar);\nfoo ?. (bar);\nfoo \n  (bar);\nfoo\n  ?. (bar);\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 8,
          column: 3,
          endLine: 8,
          endColumn: 5,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet foo = () => {};\nfoo?.(bar);\nfoo ?. (bar);\nfoo ?.\n  (bar);\nfoo\n   (bar);\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "const foo = [1, 2, 3]?.[0];",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "const foo = [1, 2, 3][0];",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const x: { a?: { b: string } };\nx?.a?.b;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 4,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const x: { a?: { b: string } };\nx.a?.b;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const x: { a: { b?: { c: string } } };\nx.a?.b?.c;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 4,
          endLine: 3,
          endColumn: 6,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const x: { a: { b?: { c: string } } };\nx.a.b?.c;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\nlet x: { a?: string };\nx?.a;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 2,
          endLine: 3,
          endColumn: 4,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\nlet x: { a?: string };\nx.a;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const foo: { bar: { baz: { c: string } } } | null;\nfoo?.bar?.baz;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { c: string } } } | null;\nfoo?.bar.baz;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const foo: { bar?: { baz: { qux: string } } } | null;\nfoo?.bar?.baz?.qux;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar?: { baz: { qux: string } } } | null;\nfoo?.bar?.baz.qux;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const foo: { bar: { baz: { qux?: () => {} } } } | null;\nfoo?.bar?.baz?.qux?.();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { qux?: () => {} } } } | null;\nfoo?.bar.baz?.qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { qux?: () => {} } } } | null;\nfoo?.bar?.baz.qux?.();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const foo: { bar: { baz: { qux: () => {} } } } | null;\nfoo?.bar?.baz?.qux?.();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { qux: () => {} } } } | null;\nfoo?.bar.baz?.qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 14,
          endLine: 3,
          endColumn: 16,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { qux: () => {} } } } | null;\nfoo?.bar?.baz.qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 19,
          endLine: 3,
          endColumn: 21,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const foo: { bar: { baz: { qux: () => {} } } } | null;\nfoo?.bar?.baz?.qux();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype baz = () => { qux: () => {} };\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.().qux?.();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = () => { qux: () => {} };\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar.baz?.().qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 14,
          endLine: 4,
          endColumn: 16,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = () => { qux: () => {} };\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz().qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 24,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = () => { qux: () => {} };\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.().qux();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype baz = null | (() => { qux: () => {} });\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.().qux?.();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = null | (() => { qux: () => {} });\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar.baz?.().qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 24,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = null | (() => { qux: () => {} });\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.().qux();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype baz = null | (() => { qux: () => {} } | null);\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.()?.qux?.();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 11,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = null | (() => { qux: () => {} } | null);\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar.baz?.()?.qux?.();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 23,
          endLine: 4,
          endColumn: 25,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype baz = null | (() => { qux: () => {} } | null);\ndeclare const foo: { bar: { baz: baz } } | null;\nfoo?.bar?.baz?.()?.qux();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { baz: number };\ntype Bar = { baz: null | string | { qux: string } };\ndeclare const foo: { fooOrBar: Foo | Bar } | null;\nfoo?.fooOrBar?.baz?.qux;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 14,
          endLine: 5,
          endColumn: 16,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { baz: number };\ntype Bar = { baz: null | string | { qux: string } };\ndeclare const foo: { fooOrBar: Foo | Bar } | null;\nfoo?.fooOrBar.baz?.qux;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const x: { a: { b: number } }[];\nx[0].a?.b;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 9,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const x: { a: { b: number } }[];\nx[0].a.b;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = 'bar' | 'foo';\ndeclare const foo: Foo;\ndeclare const key: Key;\n\nfoo?.[key]?.trim();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 13,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ntype Key = 'bar' | 'foo';\ndeclare const foo: Foo;\ndeclare const key: Key;\n\nfoo?.[key].trim();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ndeclare const foo: Foo;\nconst key = 'bar';\nfoo?.[key]?.trim();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 13,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { [key: string]: string; foo: 'foo'; bar: 'bar' } | null;\ndeclare const foo: Foo;\nconst key = 'bar';\nfoo?.[key].trim();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n    bar: 'bar';\n  };\n}\n\nexport function test(outer: Outer): number | undefined {\n  const key = 'bar';\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 11,
          column: 28,
          endLine: 11,
          endColumn: 30,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n    bar: 'bar';\n  };\n}\n\nexport function test(outer: Outer): number | undefined {\n  const key = 'bar';\n  return outer.inner?.[key].charCodeAt(0);\n}\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n    bar: 'bar';\n  };\n}\ntype Bar = 'bar';\n\nfunction Foo(outer: Outer, key: Bar): number | undefined {\n  return outer.inner?.[key]?.charCodeAt(0);\n}\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 11,
          column: 28,
          endLine: 11,
          endColumn: 30,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface Outer {\n  inner?: {\n    [key: string]: string | undefined;\n    bar: 'bar';\n  };\n}\ntype Bar = 'bar';\n\nfunction Foo(outer: Outer, key: Bar): number | undefined {\n  return outer.inner?.[key].charCodeAt(0);\n}\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\nfunction test(testVal?: true) {\n  if (testVal ?? true) {\n    console.log('test');\n  }\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 22,
        },
      ],
    },
    {
      code: "\nconst a = null;\nif (!a) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: "\nconst a = true;\nif (!a) {\n}\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 7,
        },
      ],
    },
    {
      code: "\nfunction sayHi(): void {\n  console.log('Hi!');\n}\n\nlet speech: never = sayHi();\nif (!speech) {\n}\n      ",
      errors: [
        {
          messageId: "never",
          message: "Unnecessary conditional, value is `never`.",
          line: 7,
          column: 5,
          endLine: 7,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare const x: string[] | null;\nif (x) {\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.unstrict.json",
        },
      },
      errors: [
        {
          messageId: "noStrictNullCheck",
          message: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 1,
        },
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 5,
          endLine: 3,
          endColumn: 6,
        },
      ],
    },
    {
      code: "\ninterface Foo {\n  test: string;\n  [key: string]: [string] | undefined;\n}\n\ntype OptionalFoo = Foo | undefined;\ndeclare const foo: OptionalFoo;\nfoo?.test?.length;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 9,
          column: 10,
          endLine: 9,
          endColumn: 12,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface Foo {\n  test: string;\n  [key: string]: [string] | undefined;\n}\n\ntype OptionalFoo = Foo | undefined;\ndeclare const foo: OptionalFoo;\nfoo?.test.length;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\nfunction pick<Obj extends Record<string, 1 | 2 | 3>, Key extends keyof Obj>(\n  obj: Obj,\n  key: Key,\n): Obj[Key] {\n  const k = obj[key];\n  if (obj[key]) {\n    return obj[key];\n  }\n  throw new Error('Boom!');\n}\n\npick({ foo: 1, bar: 2 }, 'bar');\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 7,
          column: 7,
          endLine: 7,
          endColumn: 15,
        },
      ],
    },
    {
      code: "\nfunction getElem(dict: Record<string, { foo: string }>, key: string) {\n  if (dict[key]) {\n    return dict[key].foo;\n  } else {\n    return '';\n  }\n}\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ndeclare let foo: {};\nfoo ??= 1;\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: number;\nfoo ??= 1;\n      ",
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: null;\nfoo ??= null;\n      ",
      errors: [
        {
          messageId: "alwaysNullish",
          message: "Unnecessary conditional, left-hand side of `??` operator is always `null` or `undefined`.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: {};\nfoo ||= 1;\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: null;\nfoo ||= null;\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: {};\nfoo &&= 1;\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare let foo: null;\nfoo &&= null;\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 4,
        },
      ],
    },
    {
      code: "\ndeclare const foo: { bar: number };\nfoo.bar ??= 1;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.exactOptionalPropertyTypes.json",
        },
      },
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 8,
        },
      ],
    },
    {
      code: "\ntype Foo = { bar: () => number } | null;\ndeclare const foo: Foo;\nfoo?.bar()?.toExponential();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 13,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { bar: () => number } | null;\ndeclare const foo: Foo;\nfoo?.bar().toExponential();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { bar: null | { baz: () => { qux: number } } } | null;\ndeclare const foo: Foo;\nfoo?.bar?.baz()?.qux?.toExponential();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 18,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { bar: null | { baz: () => { qux: number } } } | null;\ndeclare const foo: Foo;\nfoo?.bar?.baz().qux?.toExponential();\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 21,
          endLine: 4,
          endColumn: 23,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { bar: null | { baz: () => { qux: number } } } | null;\ndeclare const foo: Foo;\nfoo?.bar?.baz()?.qux.toExponential();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = (() => number) | null;\ndeclare const foo: Foo;\nfoo?.()?.toExponential();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 10,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = (() => number) | null;\ndeclare const foo: Foo;\nfoo?.().toExponential();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { [key: string]: () => number } | null;\ndeclare const foo: Foo;\nfoo?.['bar']()?.toExponential();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 17,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { [key: string]: () => number } | null;\ndeclare const foo: Foo;\nfoo?.['bar']().toExponential();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ntype Foo = { [key: string]: () => number } | null;\ndeclare const foo: Foo;\nfoo?.['bar']?.()?.toExponential();\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 4,
          column: 17,
          endLine: 4,
          endColumn: 19,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype Foo = { [key: string]: () => number } | null;\ndeclare const foo: Foo;\nfoo?.['bar']?.().toExponential();\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\n        const a = true;\n        if (!!a) {\n        }\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert(true);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert(false);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown, y: unknown): asserts x;\n\nassert(true, Math.random() > 0.5);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 8,
          endLine: 4,
          endColumn: 12,
        },
      ],
    },
    {
      code: "\ndeclare function assert(x: unknown): asserts x;\nassert({});\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 3,
          column: 8,
          endLine: 3,
          endColumn: 10,
        },
      ],
    },
    {
      code: "\ndeclare function assertsString(x: unknown): asserts x is string;\ndeclare const a: string;\nassertsString(a);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the assertion function.",
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\ndeclare function isString(x: unknown): x is string;\ndeclare const a: string;\nisString(a);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\ndeclare function isString(x: unknown): x is string;\ndeclare const a: string;\nisString('fa' + 'lafel');\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 24,
        },
      ],
    },
    {
      code: "\ndeclare function isStringOrNumber(x: unknown): x is string | number;\ndeclare const s: string;\nif (isStringOrNumber(s)) {\n}\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 4,
          column: 22,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\ninterface Wider {\n  a: string;\n}\ninterface Narrower {\n  a: string;\n  b?: number;\n}\ndeclare function isWider(x: unknown): x is Wider;\ndeclare const n: Narrower;\nif (isWider(n)) {\n}\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ninterface Wider {\n  a: string;\n}\ninterface Narrower {\n  a: string;\n  b?: number;\n}\ndeclare function isNarrower(x: unknown): x is Narrower;\ndeclare const w: Wider;\nif (isNarrower(w)) {\n}\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 11,
          column: 16,
          endLine: 11,
          endColumn: 17,
        },
      ],
    },
    {
      code: `
declare const b1: "" & {};
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: "" & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: ("" | false) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & "";
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: ("foo" | "bar") & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: (123 | true) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: (string | number) & ("foo" | 123) & { __brand: string };
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & "foo";
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: `
declare const b1: ((string & { __brandA: string }) | (number & { __brandB: string })) & ("foo" | 123);
declare const b2: boolean;
const t1 = b1 && b2;
`,
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 12,
          endLine: 4,
          endColumn: 14,
        },
      ],
    },
    {
      code: "\ntype A = {\n  [name in Lowercase<string>]?: {\n    [name in Lowercase<string>]: {\n      a: 1;\n    };\n  };\n};\n\ndeclare const a: A;\n\na.a?.a?.a;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 12,
          column: 7,
          endLine: 12,
          endColumn: 9,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ntype A = {\n  [name in Lowercase<string>]?: {\n    [name in Lowercase<string>]: {\n      a: 1;\n    };\n  };\n};\n\ndeclare const a: A;\n\na.a?.a.a;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ninterface T {\n  [name: Lowercase<string>]: {\n    [name: Lowercase<string>]: {\n      [name: Lowercase<string>]: {\n        value: 'value';\n      };\n    };\n  };\n  [name: Uppercase<string>]: null | {\n    [name: Uppercase<string>]: null | {\n      [name: Uppercase<string>]: null | {\n        VALUE: 'VALUE';\n      };\n    };\n  };\n}\n\ndeclare const t: T;\n\nt.a?.a?.a?.value;\n      ",
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 21,
          column: 4,
          endLine: 21,
          endColumn: 6,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface T {\n  [name: Lowercase<string>]: {\n    [name: Lowercase<string>]: {\n      [name: Lowercase<string>]: {\n        value: 'value';\n      };\n    };\n  };\n  [name: Uppercase<string>]: null | {\n    [name: Uppercase<string>]: null | {\n      [name: Uppercase<string>]: null | {\n        VALUE: 'VALUE';\n      };\n    };\n  };\n}\n\ndeclare const t: T;\n\nt.a.a?.a?.value;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 21,
          column: 7,
          endLine: 21,
          endColumn: 9,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface T {\n  [name: Lowercase<string>]: {\n    [name: Lowercase<string>]: {\n      [name: Lowercase<string>]: {\n        value: 'value';\n      };\n    };\n  };\n  [name: Uppercase<string>]: null | {\n    [name: Uppercase<string>]: null | {\n      [name: Uppercase<string>]: null | {\n        VALUE: 'VALUE';\n      };\n    };\n  };\n}\n\ndeclare const t: T;\n\nt.a?.a.a?.value;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 21,
          column: 10,
          endLine: 21,
          endColumn: 12,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ninterface T {\n  [name: Lowercase<string>]: {\n    [name: Lowercase<string>]: {\n      [name: Lowercase<string>]: {\n        value: 'value';\n      };\n    };\n  };\n  [name: Uppercase<string>]: null | {\n    [name: Uppercase<string>]: null | {\n      [name: Uppercase<string>]: null | {\n        VALUE: 'VALUE';\n      };\n    };\n  };\n}\n\ndeclare const t: T;\n\nt.a?.a?.a.value;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const test: Array<{ a?: string }>;\n\nif (test[0]?.a) {\n  test[0]?.a;\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 5,
          column: 10,
          endLine: 5,
          endColumn: 12,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const test: Array<{ a?: string }>;\n\nif (test[0]?.a) {\n  test[0].a;\n}\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const arr2: Array<{ x: { y: { z: object } } }>;\narr2[42]?.x?.y?.z;\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
      errors: [
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 12,
          endLine: 3,
          endColumn: 14,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const arr2: Array<{ x: { y: { z: object } } }>;\narr2[42]?.x.y?.z;\n      ",
            },
          ],
        },
        {
          messageId: "neverOptionalChain",
          message: "Unnecessary optional chain on a non-nullish value.",
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 17,
          suggestions: [
            {
              messageId: "suggestRemoveOptionalChain",
              desc: "Remove unnecessary optional chain",
              output: "\ndeclare const arr2: Array<{ x: { y: { z: object } } }>;\narr2[42]?.x?.y.z;\n      ",
            },
          ],
        },
      ],
    },
    {
      code: "\ndeclare const arr: string[];\n\nif (arr[0]) {\n  arr[0] ?? 'foo';\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
      errors: [
        {
          messageId: "neverNullish",
          message: "Unnecessary conditional, expected left-hand side of `??` operator to be possibly null or undefined.",
          line: 5,
          column: 3,
          endLine: 5,
          endColumn: 9,
        },
      ],
    },
    {
      code: "\ndeclare const arr: object[];\n\nif (arr[42] && arr[42]) {\n}\n      ",
      languageOptions: {
        parserOptions: {
          project: "./tsconfig.noUncheckedIndexedAccess.json",
          projectService: false,
        },
      },
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 4,
          column: 16,
          endLine: 4,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\ndeclare function isString(x: unknown): x is string;\ndeclare const a: string;\nisString(a);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the type guard.",
          line: 4,
          column: 10,
          endLine: 4,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\ndeclare function assertsString(x: unknown): asserts x is string;\ndeclare const a: string;\nassertsString(a);\n      ",
      options: [
        {
          checkTypePredicates: true,
        },
      ],
      errors: [
        {
          messageId: "typeGuardAlreadyIsType",
          message: "Unnecessary conditional, expression already has the type being checked by the assertion function.",
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 16,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].filter(() => 'truthy-string');\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 24,
          endLine: 2,
          endColumn: 39,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].find(function () {\n  return 0;\n});\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].some(() => 1);\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 22,
          endLine: 2,
          endColumn: 23,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].every(() => null);\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 27,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].findIndex(function namedFalsy() {\n  return undefined;\n});\n      ",
      errors: [
        {
          messageId: "alwaysFalsy",
          message: "Unnecessary conditional, value is always falsy.",
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 19,
        },
      ],
    },
    {
      code: "\n[1, 2, 3].findLast(() => 'always');\n      ",
      errors: [
        {
          messageId: "alwaysTruthy",
          message: "Unnecessary conditional, value is always truthy.",
          line: 2,
          column: 26,
          endLine: 2,
          endColumn: 34,
        },
      ],
    },
  ],
});
