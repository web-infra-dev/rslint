import { RuleTester } from '@typescript-eslint/rule-tester';


import { getFixturesRootDir } from '../RuleTester';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: rootPath,
    },
  },
});

describe('basic assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const a: string;
  a as string | number;
      `,
      `
  declare const a: string;
  <string | number>a;
      `,
      `
  declare const a: string;
  a as string | number as string | number | boolean;
      `,
      `
  declare const a: string;
  a as string;
      `,
      `
  declare const a: { hello: 'world' };
  a as { hello: string };
      `,
      `
  'hello' as const;
      `,
      `
  function foo<T extends boolean>(a: T) {
  return a as T | number;
  }
      `,
    ],
    invalid: [
      {
        code: `
  declare const a: string | number;
  a as string;
        `,
        errors: [
          {
            data: {
              type: 'string',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: string | number;
  a satisfies string as string;
        `,
        errors: [
          {
            data: {
              type: 'string',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: string | number;
  <string>a;
        `,
        errors: [
          {
            data: {
              type: 'string',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: string | undefined;
  a as string | boolean;
        `,
        errors: [
          {
            data: {
              type: 'string | boolean',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      // multiple failures
      {
        code: `
  declare const a: string;
  a as 'foo' as 'bar';
        `,
        errors: [
          {
            data: {
              type: '"bar"',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
          {
            data: {
              type: '"foo"',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      // type constraint
      {
        code: `
  function foo<T extends boolean>(a: T) {
  return a as true;
  }
        `,
        errors: [
          {
            data: {
              type: 'true',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      // long/complex asserted type
      {
        code: `
  declare const a: string;
  a as Omit<Required<Readonly<{ hello: 'world'; foo: 'bar' }>>, 'foo'>;
        `,
        errors: [
          {
            data: {
              type: 'Omit<Required<Readonly<{ hello: "world"; foo: "bar"; }>>, "foo">',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const foo: readonly number[];
  const bar = foo as number[];
        `,
        errors: [
          {
            data: {
              type: 'number[]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('any assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const _any_: any;
  _any_ as any;
      `,
      `
  declare const _any_: any;
  _any_ as unknown;
      `,
    ],
    invalid: [
      {
        code: `
  declare const _any_: any;
  _any_ as string;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _unknown_: unknown;
  _unknown_ as any;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _any_: any;
  _any_ as Function;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _any_: any;
  _any_ as never;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  'foo' as any;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 2,
            line: 2,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      // an error type `any`
      {
        code: `
  const bar = foo as number;
        `,
        errors: [
          {
            data: {
              type: 'error typed',
            },
            endLine: 2,
            line: 2,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  const bar = 'foo' as errorType;
        `,
        errors: [
          {
            data: {
              type: 'error typed',
            },
            endLine: 2,
            line: 2,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('never assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const _never_: never;
  _never_ as never;
      `,
      `
  declare const _never_: never;
  _never_ as unknown;
      `,
    ],
    invalid: [
      {
        code: `
  declare const _never_: never;
  _never_ as any;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _string_: string;
  _string_ as never;
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('function assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const _function_: Function;
  _function_ as Function;
      `,
      `
  declare const _function_: Function;
  _function_ as unknown;
      `,
    ],
    invalid: [
      {
        code: `
  declare const _function_: Function;
  _function_ as () => void;
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _function_: Function;
  _function_ as any;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const _function_: Function;
  _function_ as never;
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('object assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  // additional properties should be allowed
  export const foo = { bar: 1, bazz: 1 } as {
  bar: number;
  };
      `,
      `
  declare const a: { hello: string } & { world: string };
  a as { hello: string };
      `,
      `
  declare const a: { hello: any };
  a as { hello: unknown };
      `,
      `
  declare const a: { hello: string };
  a as { hello?: string };
      `,
      `
  declare const a: { hello: string };
  a satisfies Record<string, string> as { hello?: string };
      `,
    ],
    invalid: [
      {
        code: `
  var foo = {} as {
  bar: number;
  bas: string;
  };
        `,
        errors: [
          {
            data: {
              type: '{ bar: number; bas: string; }',
            },
            endLine: 5,
            line: 2,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: { hello: string };
  a satisfies Record<string, string> as { hello: string; world: string };
        `,
        errors: [
          {
            data: {
              type: '{ hello: string; world: string; }',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: { hello?: string };
  a as { hello: string };
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('array assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const a: string[];
  a as (string | number)[];
      `,
      `
  declare const a: number[];
  a as unknown[];
      `,
      `
  declare const a: { hello: 'world'; foo: 'bar' }[];
  a as { hello: 'world' }[];
      `,
    ],
    invalid: [
      {
        code: `
  declare const a: (string | number)[];
  a as string[];
        `,
        errors: [
          {
            data: {
              type: 'string[]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: any[];
  a as number[];
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: number[];
  a as any[];
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: unknown[];
  a as number[];
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: number[];
  a as never[];
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('tuple assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const a: [string];
  a as [string | number];
      `,
      `
  declare const a: [string, number];
  a as [string, string | number];
      `,
      `
  declare const a: [string];
  a as [unknown];
      `,
      `
  declare const a: [{ hello: 'world'; foo: 'bar' }];
  a as [{ hello: 'world' }];
      `,
    ],
    invalid: [
      {
        code: `
  declare const a: [string | number];
  a as [string];
        `,
        errors: [
          {
            data: {
              type: '[string]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [string, number];
  a as [string, string];
        `,
        errors: [
          {
            data: {
              type: '[string, string]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [string];
  a as [string, number];
        `,
        errors: [
          {
            data: {
              type: '[string, number]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [string, number];
  a as [string];
        `,
        errors: [
          {
            data: {
              type: '[string]',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [any];
  a as [number];
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [number, any];
  a as [number, number];
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [number];
  a as [any];
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [unknown];
  a as [number];
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [number];
  a as [never];
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: [Promise<string | number>];
  a as [Promise<string>];
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('promise assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  declare const a: Promise<string>;
  a as Promise<string | number>;
      `,
      `
  declare const a: Promise<number>;
  a as Promise<unknown>;
      `,
      `
  declare const a: Promise<{ hello: 'world'; foo: 'bar' }>;
  a as Promise<{ hello: 'world' }>;
      `,
      `
  declare const a: Promise<string>;
  a as Promise<string> | string;
      `,
    ],
    invalid: [
      {
        code: `
  declare const a: Promise<string | number>;
  a as Promise<string>;
        `,
        errors: [
          {
            data: {
              type: 'Promise<string>',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: Promise<any>;
  a as Promise<number>;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: Promise<number>;
  a as Promise<any>;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: Promise<number[]>;
  a as Promise<any[]>;
        `,
        errors: [
          {
            data: {
              type: '`any`',
            },
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: Promise<unknown>;
  a as Promise<number>;
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  declare const a: Promise<number>;
  a as Promise<never>;
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('class assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  class Foo {}
  declare const a: Foo;
  a as Foo | number;
      `,
      `
  class Foo {}
  class Bar {}
  declare const a: Foo;
  a as Bar;
      `,
      `
  class Foo {
  hello() {}
  }
  class Bar {}
  declare const a: Foo;
  a as Bar;
      `,
      `
  class Foo {
  hello() {}
  }
  class Bar extends Foo {}
  declare const a: Bar;
  a as Foo;
      `,
      `
  class Foo {
  hello() {}
  }
  class Bar extends Foo {}
  declare const a: Foo;
  a as Bar;
      `,
    ],
    invalid: [
      {
        code: `
  class Foo {
  hello() {}
  }
  class Bar extends Foo {
  world() {}
  }
  declare const a: Foo;
  a as Bar;
        `,
        errors: [
          {
            endLine: 9,
            line: 9,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});

describe('generic assertions', () => {
  ruleTester.run('no-unsafe-type-assertion', {
    valid: [
      `
  type Obj = { foo: string };
  function func<T extends Obj>(a: T) {
  const b = a as T;
  }
      `,
      `
  function parameterExtendsOtherParameter<T extends string | number, V extends T>(
  x: T,
  y: V,
  ) {
  y as T;
  }
      `,
      `
  function parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {
  y as T;
  }
      `,
      `
  function unconstrainedToUnknown<T>(x: T) {
  x as unknown;
  }
      `,
      `
  function stringToWider<T extends string>(x: T) {
  x as number | string; // allowed
  }
      `,
    ],
    invalid: [
      {
        code: `
  type Obj = { foo: string };
  function func<T extends Obj>() {
  const myObj = { foo: 'hi' } as T;
  }
        `,
        errors: [
          {
            endLine: 4,
            line: 4,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  type Obj = { foo: string };
  function func<T extends Obj>() {
  const o: Obj = { foo: 'hi' };
  const myObj = o as T;
  }
        `,
        errors: [
          {
            endLine: 5,
            line: 5,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      // https://github.com/typescript-eslint/typescript-eslint/issues/10453#issuecomment-2520964068
      // the custom error message should only occur if the expression type is
      // *actually* assignable to the constraint of the asserted type
      {
        code: `
  export function myfunc<CustomObjectT extends string>(
  input: number,
  ): CustomObjectT {
  const newCustomObject = input as CustomObjectT;
  return newCustomObject;
  }
        `,
        errors: [
          {
            endLine: 5,
            line: 5,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      // https://github.com/typescript-eslint/typescript-eslint/pull/10461#discussion_r1873887553
      // 1. non-parameter -> parameter assertions
      {
        code: `
  function unknownConstraint<T extends unknown>(x: T, y: string) {
  y as T; // banned; generic arbitrary subtype
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  function unconstrained<T>(x: T, y: string) {
  y as T;
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeToUnconstrainedTypeAssertion',
          },
        ],
      },
      {
        code: `
  // constraint of any functions like constraint of \`unknown\`
  // (even the TS error message has this verbiage)
  function anyConstraint<T extends any>(x: T, y: string) {
  y as T; // banned; generic arbitrary subtype
  }
        `,
        errors: [
          {
            endLine: 5,
            line: 5,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  function constraintWiderThanUncastType<T extends string | number>(
  x: T,
  y: string,
  ) {
  y as T; // banned; assignable to constraint
  }
        `,
        errors: [
          {
            endLine: 6,
            line: 6,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  function constraintEqualUncastType<T extends string>(x: T, y: string) {
  y as T; // banned; assignable to constraint
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  function constraintNarrowerThanUncastType<T extends string>(
  x: T,
  y: string | number,
  ) {
  y as T; // banned; *not* assignable to constraint
  }
        `,
        errors: [
          {
            endLine: 6,
            line: 6,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  function assertFromAny<T extends string | number>(x: T, y: any) {
  y as T; // banned; just an \`any\` complaint. Not a generic subtype.
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeOfAnyTypeAssertion',
          },
        ],
      },
      // 2. parameter -> parameter assertions
      {
        code: `
  function parameterExtendsOtherParameter<T extends string | number, V extends T>(
  x: T,
  y: V,
  ) {
  x as V; // banned; assignable to constraint
  }
        `,
        errors: [
          {
            endLine: 6,
            line: 6,
            messageId: 'unsafeTypeAssertionAssignableToConstraint',
          },
        ],
      },
      {
        code: `
  function parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {
  x as V; // banned; unconstrained arbitrary type
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeToUnconstrainedTypeAssertion',
          },
        ],
      },
      {
        code: `
  function twoUnconstrained<T, V>(x: T, y: V) {
  y as T;
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeToUnconstrainedTypeAssertion',
          },
        ],
      },
      // 2. parameter -> non-parameter assertions
      {
        code: `
  function toNarrower<T>(x: T, y: string) {
  x as string; // banned; ordinary 'string' narrower than 'T'.
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
      {
        code: `
  function unconstrainedToAny<T>(x: T) {
  x as any;
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  function stringToAny<T extends string>(x: T) {
  x as any;
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeToAnyTypeAssertion',
          },
        ],
      },
      {
        code: `
  function stringToNarrower<T extends string>(x: T) {
  x as 'a' | 'b';
  }
        `,
        errors: [
          {
            endLine: 3,
            line: 3,
            messageId: 'unsafeTypeAssertion',
          },
        ],
      },
    ],
  });
});
