import { RuleTester } from '@typescript-eslint/rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('explicit-module-boundary-types', {
  valid: [
    `
function test(): void {
  return;
}
    `,
    `
export function test(): void {
  return;
}
    `,
    `
export var fn = function (): number {
  return 1;
};
    `,
    `
export var arrowFn = (): string => 'test';
    `,
    // not exported — no diagnostics regardless of body
    `
class Test {
  constructor(one) {}
  get prop(one) {
    return 1;
  }
  set prop(one) {}
  method(one) {
    return;
  }
  arrow = one => 'arrow';
  abstract abs(one);
}
    `,
    `
export class Test {
  constructor(one: string) {}
  get prop(one: string): void {
    return 1;
  }
  set prop(one: string): void {}
  method(one: string): void {
    return;
  }
  arrow = (one: string): string => 'arrow';
  abstract abs(one: string): void;
}
    `,
    // private members are skipped
    `
export class Test {
  private constructor(one) {}
  private get prop(one) {
    return 1;
  }
  private set prop(one) {}
  private method(one) {
    return;
  }
  private arrow = one => 'arrow';
  private abstract abs(one);
}
    `,
    // # private members are skipped
    `
export class PrivateProperty {
  #property = () => null;
}
    `,
    `
export class PrivateMethod {
  #method() {}
}
    `,
    // overload signatures with implementation that has a body+return type
    `
export class Test {
  constructor();
  constructor(value?: string) {
    console.log(value);
  }
}
    `,
    // declare class — body-less constructor needs no return type
    `
declare class MyClass {
  constructor(options?: MyClass.Options);
}
export { MyClass };
    `,
    // allowTypedFunctionExpressions: variable annotation supplies the type
    {
      code: `
export var arrowFn: Foo = () => 'test';
      `,
      options: [{ allowTypedFunctionExpressions: true }],
    },
    {
      code: `
export var funcExpr: Foo = function () {
  return 'test';
};
      `,
      options: [{ allowTypedFunctionExpressions: true }],
    },
    {
      code: 'const x = (() => {}) as Foo;',
      options: [{ allowTypedFunctionExpressions: true }],
    },
    {
      code: `
export default () => (): void => {};
      `,
      options: [{ allowHigherOrderFunctions: true }],
    },
    // allowDirectConstAssertionInArrowFunctions
    {
      code: `
export const func1 = (value: number) => ({ type: 'X', value }) as const;
      `,
      options: [{ allowDirectConstAssertionInArrowFunctions: true }],
    },
    // allowedNames
    {
      code: `
export const func1 = (value: string) => value;
export const func2 = (value: number) => ({ type: 'X', value });
      `,
      options: [{ allowedNames: ['func1', 'func2'] }],
    },
    // higher-order
    `
export function foo(): (n: number) => string {
  return n => String(n);
}
    `,
    // explicit `as` cast covers the value
    `
const foo = (arg => arg) as Foo;
export default foo;
    `,
    // overload — allowOverloadFunctions: true
    {
      code: `
export function test(a: string): string;
export function test(a: number): number;
export function test(a: unknown) {
  return a;
}
      `,
      options: [{ allowOverloadFunctions: true }],
    },
    // allowArgumentsExplicitlyTypedAsAny
    {
      code: `
export function foo(foo: any): void {}
      `,
      options: [{ allowArgumentsExplicitlyTypedAsAny: true }],
    },
  ],
  invalid: [
    {
      code: `
export function test(a: number, b: number) {
  return;
}
      `,
      errors: [
        {
          column: 8,
          endColumn: 21,
          endLine: 2,
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
    },
    {
      code: `
export function test() {
  return;
}
      `,
      errors: [
        {
          column: 8,
          endColumn: 21,
          endLine: 2,
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
    },
    {
      code: `
export var fn = function () {
  return 1;
};
      `,
      errors: [
        {
          column: 17,
          endColumn: 26,
          endLine: 2,
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
    },
    {
      code: `
export var arrowFn = () => 'test';
      `,
      errors: [
        {
          column: 25,
          endColumn: 27,
          endLine: 2,
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
    },
    // destructured parameters
    {
      code: `
export function foo({ foo }): void {}
      `,
      errors: [
        {
          column: 21,
          line: 2,
          messageId: 'missingArgTypeUnnamed',
        },
      ],
    },
    {
      code: `
export function foo([bar]): void {}
      `,
      errors: [
        {
          column: 21,
          line: 2,
          messageId: 'missingArgTypeUnnamed',
        },
      ],
    },
    {
      code: `
export function foo(...bar): void {}
      `,
      errors: [
        {
          column: 21,
          line: 2,
          messageId: 'missingArgType',
        },
      ],
    },
    {
      code: `
export function foo(...[a]): void {}
      `,
      errors: [
        {
          column: 21,
          line: 2,
          messageId: 'missingArgTypeUnnamed',
        },
      ],
    },
    // allowArgumentsExplicitlyTypedAsAny: false
    {
      code: `
export function foo(foo: any): void {}
      `,
      errors: [
        {
          column: 21,
          line: 2,
          messageId: 'anyTypedArg',
        },
      ],
      options: [{ allowArgumentsExplicitlyTypedAsAny: false }],
    },
    // allowedNames excluding func2
    {
      code: `
export const func1 = (value: number) => value;
export const func2 = (value: number) => value;
      `,
      errors: [
        {
          column: 38,
          endColumn: 40,
          endLine: 2,
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
      options: [{ allowedNames: ['func2'] }],
    },
    // followReference: untyped const default-exported by identifier
    {
      code: `
const foo = arg => arg;
export default foo;
      `,
      errors: [
        {
          data: { name: 'arg' },
          line: 2,
          messageId: 'missingArgType',
        },
        {
          line: 2,
          messageId: 'missingReturnType',
        },
      ],
    },
    // higher-order: untyped both
    {
      code: `
export function foo(outer) {
  return function (inner) {};
}
      `,
      errors: [
        {
          data: { name: 'outer' },
          line: 2,
          messageId: 'missingArgType',
        },
        {
          line: 3,
          messageId: 'missingReturnType',
        },
        {
          data: { name: 'inner' },
          line: 3,
          messageId: 'missingArgType',
        },
      ],
      options: [{ allowHigherOrderFunctions: true }],
    },
    // overload — implementation needs return type by default
    {
      code: `
export function test(a: string): string;
export function test(a: number): number;
export function test(a: unknown) {
  return a;
}
      `,
      errors: [
        {
          column: 8,
          endColumn: 21,
          line: 4,
          messageId: 'missingReturnType',
        },
      ],
    },
  ],
});
