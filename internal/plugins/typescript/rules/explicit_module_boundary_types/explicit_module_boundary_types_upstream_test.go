// TestExplicitModuleBoundaryTypesUpstream migrates the full valid/invalid
// suite from upstream packages/eslint-plugin/tests/rules/explicit-module-boundary-types.test.ts
// 1:1. Position assertions cover line/column for every invalid case. rslint-
// specific lock-in cases live in the explicit_module_boundary_types_extras_test.go file.
package explicit_module_boundary_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitModuleBoundaryTypesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitModuleBoundaryTypesRule, []rule_tester.ValidTestCase{
		// non-exported function: rule never fires
		{Code: `
function test(): void {
  return;
}
		`},
		// directly-exported function with explicit return type
		{Code: `
export function test(): void {
  return;
}
		`},
		// exported `var fn = function () { ... }`
		{Code: `
export var fn = function (): number {
  return 1;
};
		`},
		// exported arrow with return type
		{Code: `
export var arrowFn = (): string => 'test';
		`},
		// non-exported class: nothing to report even if members are untyped
		{Code: `
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
		`},
		// exported class with all members typed
		{Code: `
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
		`},
		// exported class but every member is `private` — skipped
		{Code: `
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
		`},
		// #private property — skipped
		{Code: `
export class PrivateProperty {
  #property = () => null;
}
		`},
		// #private method — skipped
		{Code: `
export class PrivateMethod {
  #method() {}
}
		`},
		// constructor overload — only implementation has body; signature is skipped
		{Code: `
export class Test {
  constructor();
  constructor(value?: string) {
    console.log(value);
  }
}
		`},
		// `declare class` member with no body — type-only, no check needed
		{Code: `
declare class MyClass {
  constructor(options?: MyClass.Options);
}
export { MyClass };
		`},
		// nested function declarations inside an exported function are not exported
		{Code: `
export function test(): void {
  nested();
  return;

  function nested() {}
}
		`},
		// nested arrow inside exported function — not exported itself
		{Code: `
export function test(): string {
  const nested = () => 'value';
  return nested();
}
		`},
		// nested class inside exported function — not exported itself
		{Code: `
export function test(): string {
  class Nested {
    public method() {
      return 'value';
    }
  }
  return new Nested().method();
}
		`},
		// allowTypedFunctionExpressions — variable annotation supplies the type
		{
			Code: `
export var arrowFn: Foo = () => 'test';
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export var funcExpr: Foo = function () {
  return 'test';
};
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// type assertion supplies the type
		{
			Code:    `const x = (() => {}) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// angle-bracket type assertion
		{
			Code:    `const x = <Foo>(() => {});`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// trailing `as Foo` on an object literal — every nested property is "typed"
		{
			Code: `
export const x = {
  foo: () => {},
} as Foo;
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export const x = <Foo>{
  foo: () => {},
};
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export const x: Foo = {
  foo: () => {},
};
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// nested object property — typed because the outer object is typed
		{
			Code: `
export const x = {
  foo: { bar: () => {} },
} as Foo;
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export const x = <Foo>{
  foo: { bar: () => {} },
};
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export const x: Foo = {
  foo: { bar: () => {} },
};
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// class property has a type annotation — arrow initializer is typed
		{
			Code: `
type MethodType = () => void;

export class App {
  public method: MethodType = () => {};
}
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// object literal setter parameter has a type — return type unneeded for set
		{Code: `
export const myObj = {
  set myProp(val: number) {
    this.myProp = val;
  },
};
		`},
		// allowHigherOrderFunctions — bodyless outer arrow → inner arrow
		{
			Code: `
export default () => (): void => {};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => function (): void {};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => {
  return (): void => {};
};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => {
  return function (): void {};
};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export function fn() {
  return (): void => {};
}
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export function fn() {
  return function (): void {};
}
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// deep higher-order: only the innermost arrow needs `:number`
		{
			Code: `
export function FunctionDeclaration() {
  return function FunctionExpression_Within_FunctionDeclaration() {
    return function FunctionExpression_Within_FunctionExpression() {
      return () => {
        // ArrowFunctionExpression_Within_FunctionExpression
        return () =>
          // ArrowFunctionExpression_Within_ArrowFunctionExpression
          (): number =>
            1; // ArrowFunctionExpression_Within_ArrowFunctionExpression_WithNoBody
      };
    };
  };
}
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => () => {
  return (): void => {
    return;
  };
};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => () => {
  const foo = 'foo';
  return (): void => {
    return;
  };
};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => () => {
  const foo = () => (): string => 'foo';
  return (): void => {
    return;
  };
};
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// non-exported invocation — `new Accumulator().accumulate(() => 1)` is not exported
		{
			Code: `
export class Accumulator {
  private count: number = 0;

  public accumulate(fn: () => number): void {
    this.count += fn();
  }
}

new Accumulator().accumulate(() => 1);
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// allowDirectConstAssertionInArrowFunctions — body is `as const`
		{
			Code: `
export const func1 = (value: number) => ({ type: 'X', value }) as const;
export const func2 = (value: number) => ({ type: 'X', value }) as const;
export const func3 = (value: number) => x as const;
export const func4 = (value: number) => x as const;
		`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// `as const satisfies R`
		{
			Code: `
interface R {
  type: string;
  value: number;
}

export const func = (value: number) =>
  ({ type: 'X', value }) as const satisfies R;
		`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		{
			Code: `
interface R {
  type: string;
  value: number;
}

export const func = (value: number) =>
  ({ type: 'X', value }) as const satisfies R satisfies R;
		`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		{
			Code: `
interface R {
  type: string;
  value: number;
}

export const func = (value: number) =>
  ({ type: 'X', value }) as const satisfies R satisfies R satisfies R;
		`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// allowedNames
		{
			Code: `
export const func1 = (value: string) => value;
export const func2 = (value: number) => ({ type: 'X', value });
		`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"func1", "func2"}},
		},
		{
			Code: `
export function func1() {
  return 0;
}
export const foo = {
  func2() {
    return 0;
  },
};
		`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"func1", "func2"}},
		},
		// allowedNames with computed/quoted keys
		{
			Code: `
export class Test {
  get prop() {
    return 1;
  }
  set prop() {}
  method() {
    return;
  }
  // prettier-ignore
  'method'() {}
  ['prop']() {}
  [` + "`prop`" + `]() {}
  [null]() {}
  [` + "`${v}`" + `](): void {}

  foo = () => {
    bar: 5;
  };
}
		`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"prop", "method", "null", "foo"}},
		},
		// higher-order: outer typed, inner typed
		{
			Code: `
        export function foo(outer: string) {
          return function (inner: string): void {};
        }
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// allowTypedFunctionExpressions — variable type covers the arrow
		{
			Code: `
        export type Ensurer = (blocks: TFBlock[]) => TFBlock[];

        export const myEnsurer: Ensurer = blocks => {
          return blocks;
        };
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// JSX — children are JSXExpressionContainer arguments → typed
		{
			Code: `
export const Foo: FC = () => (
  <div a={e => {}} b={function (e) {}} c={function foo(e) {}}></div>
);
		`,
			Tsx: true,
		},
		{
			Code: `
export const Foo: JSX.Element = (
  <div a={e => {}} b={function (e) {}} c={function foo(e) {}}></div>
);
		`,
			Tsx: true,
		},
		// followReference: const test arrow with type
		{Code: `
const test = (): void => {
  return;
};
export default test;
		`},
		{Code: `
function test(): void {
  return;
}
export default test;
		`},
		{Code: `
const test = (): void => {
  return;
};
export default [test];
		`},
		{Code: `
function test(): void {
  return;
}
export default [test];
		`},
		{Code: `
const test = (): void => {
  return;
};
export default { test };
		`},
		{Code: `
function test(): void {
  return;
}
export default { test };
		`},
		// as-cast supplies the type
		{Code: `
const foo = (arg => arg) as Foo;
export default foo;
		`},
		// let with reassignment — both writes carry type
		{Code: `
let foo = (arg => arg) as Foo;
foo = 3;
export default foo;
		`},
		// class default export — referenced inside `export default { Foo }`
		{Code: `
class Foo {
  bar = (arg: string): string => arg;
}
export default { Foo };
		`},
		{Code: `
class Foo {
  bar(): void {
    return;
  }
}
export default { Foo };
		`},
		// accessor property
		{Code: `
export class Foo {
  accessor bar = (): void => {
    return;
  };
}
		`},
		// allowHigherOrderFunctions default behavior — inner arrow types itself via return-type annotation
		{Code: `
export function foo(): (n: number) => string {
  return n => String(n);
}
		`},
		{Code: `
export const foo = (a: string): ((n: number) => string) => {
  return function (n) {
    return String(n);
  };
};
		`},
		// inner functions NOT exported — no diagnostics
		{Code: `
export function a(): void {
  function b() {}
  const x = () => {};
  (function () {});

  function c() {
    return () => {};
  }

  return;
}
		`},
		{Code: `
export function a(): void {
  function b() {
    function c() {}
  }
  const x = () => {
    return () => 100;
  };
  (function () {
    (function () {});
  });

  function c() {
    return () => {
      (function () {});
    };
  }

  return;
}
		`},
		// higher-order, allowHigherOrderFunctions, inner typed
		{
			Code: `
export function a() {
  return function b(): () => void {
    return function c() {};
  };
}
		`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// default `allowHigherOrderFunctions: true`
		{Code: `
export var arrowFn = () => (): void => {};
		`},
		{Code: `
export function fn() {
  return function (): void {};
}
		`},
		{Code: `
export function foo(outer: string) {
  return function (inner: string): void {};
}
		`},
		// `new Proxy(apiInstance, { get: (target, property) => {} })` — methods nested inside
		// an object passed to `new` are NOT directly exported — see upstream issue #2134
		{Code: `
export function foo(): unknown {
  return new Proxy(apiInstance, {
    get: (target, property) => {
      // implementation
    },
  });
}
		`},
		// IIFE
		{
			Code:    `export default (() => true)();`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
		},
		// explicit assertions remain allowed
		{
			Code:    `export const x = (() => {}) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
		},
		{
			Code: `
interface Foo {}
export const x = {
  foo: () => {},
} as Foo;
		`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
		},
		// allowArgumentsExplicitlyTypedAsAny
		{
			Code: `
export function foo(foo: any): void {}
		`,
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": true},
		},
		{
			Code: `
export function foo({ foo }: any): void {}
		`,
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": true},
		},
		{
			Code: `
export function foo([bar]: any): void {}
		`,
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": true},
		},
		{
			Code: `
export function foo(...bar: any): void {}
		`,
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": true},
		},
		{
			Code: `
export function foo(...[a]: any): void {}
		`,
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": true},
		},
		// assignment patterns: ignored
		{Code: `
export function foo(arg = 1): void {}
		`},
		// higher-order chained returns
		{Code: `
export const foo = (): ((n: number) => string) => n => String(n);
		`},
		// see upstream issue #2173
		{Code: `
export function foo(): (n: number) => (m: number) => string {
  return function (n) {
    return function (m) {
      return String(n + m);
    };
  };
}
		`},
		{Code: `
export const foo = (): ((n: number) => (m: number) => string) => n => m =>
  String(n + m);
		`},
		{Code: `
export const bar: () => (n: number) => string = () => n => String(n);
		`},
		{Code: `
type Buz = () => (n: number) => string;

export const buz: Buz = () => n => String(n);
		`},
		// abstract set accessor — body-less, no return-type required for set
		{Code: `
export abstract class Foo<T> {
  abstract set value(element: T);
}
		`},
		// declare class — body-less set accessor
		{Code: `
export declare class Foo {
  set time(seconds: number);
}
		`},
		// exported class self-referencing
		{Code: `
export class A {
  b = A;
}
		`},
		// typed array of typed object — every nested arrow gets a type via context
		{Code: `
interface Foo {
  f: (x: boolean) => boolean;
}

export const a: Foo[] = [
  {
    f: (x: boolean) => x,
  },
];
		`},
		{Code: `
interface Foo {
  f: (x: boolean) => boolean;
}

export const a: Foo = {
  f: (x: boolean) => x,
};
		`},
		// allowOverloadFunctions
		{
			Code: `
export function test(a: string): string;
export function test(a: number): number;
export function test(a: unknown) {
  return a;
}
		`,
			Options: map[string]interface{}{"allowOverloadFunctions": true},
		},
		{
			Code: `
export default function test(a: string): string;
export default function test(a: number): number;
export default function test(a: unknown) {
  return a;
}
		`,
			Options: map[string]interface{}{"allowOverloadFunctions": true},
		},
		{
			Code: `
export default function (a: string): string;
export default function (a: number): number;
export default function (a: unknown) {
  return a;
}
		`,
			Options: map[string]interface{}{"allowOverloadFunctions": true},
		},
		{
			Code: `
export class Test {
  test(a: string): string;
  test(a: number): number;
  test(a: unknown) {
    return a;
  }
}
		`,
			Options: map[string]interface{}{"allowOverloadFunctions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// missing return type
		{
			Code: `
export function test(a: number, b: number) {
  return;
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 8, EndLine: 2, EndColumn: 21},
			},
		},
		{
			Code: `
export function test() {
  return;
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 8, EndLine: 2, EndColumn: 21},
			},
		},
		{
			Code: `
export var fn = function () {
  return 1;
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 17, EndLine: 2, EndColumn: 26},
			},
		},
		{
			Code: `
export var arrowFn = () => 'test';
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 25, EndLine: 2, EndColumn: 27},
			},
		},
		// class members that need types
		{
			Code: `
export class Test {
  constructor() {}
  get prop() {
    return 1;
  }
  set prop(value) {}
  method() {
    return;
  }
  arrow = arg => 'arrow';
  private method() {
    return;
  }
  abstract abs(arg);
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3, EndLine: 4, EndColumn: 11},
				{MessageId: "missingArgType", Line: 7, Column: 12, EndLine: 7, EndColumn: 17},
				{MessageId: "missingReturnType", Line: 8, Column: 3, EndLine: 8, EndColumn: 9},
				{MessageId: "missingReturnType", Line: 11, Column: 3, EndLine: 11, EndColumn: 11},
				{MessageId: "missingArgType", Line: 11, Column: 11, EndLine: 11, EndColumn: 14},
				{MessageId: "missingReturnType", Line: 15, Column: 15, EndLine: 15, EndColumn: 21},
				{MessageId: "missingArgType", Line: 15, Column: 16, EndLine: 15, EndColumn: 19},
			},
		},
		// public/static class fields with untyped arrows / functions
		{
			Code: `
export class Foo {
  public a = () => {};
  public b = function () {};
  public c = function test() {};

  static d = () => {};
  static e = function () {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 3, EndLine: 3, EndColumn: 14},
				{MessageId: "missingReturnType", Line: 4, Column: 3, EndLine: 4, EndColumn: 23},
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndLine: 5, EndColumn: 27},
				{MessageId: "missingReturnType", Line: 7, Column: 3, EndLine: 7, EndColumn: 14},
				{MessageId: "missingReturnType", Line: 8, Column: 3, EndLine: 8, EndColumn: 23},
			},
		},
		// ternary branch with untyped arrow
		{
			Code: `export default () => (true ? () => {} : (): void => {});`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 1, Column: 19, EndLine: 1, EndColumn: 21},
			},
		},
		// missing return type on arrow even with allowTypedFunctionExpressions
		{
			Code: `export var arrowFn = () => 'test';`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
			},
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
export var funcExpr = function () {
  return 'test';
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 23, EndLine: 2, EndColumn: 32},
			},
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// allowTypedFunctionExpressions: false — typed parent doesn't help
		{
			Code: `
interface Foo {}
export const x: Foo = {
  foo: () => {},
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3, EndLine: 4, EndColumn: 8},
			},
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
		},
		// allowHigherOrderFunctions but inner untyped
		{
			Code: `export default () => () => {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 1, Column: 25, EndLine: 1, EndColumn: 27},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `export default () => function () {};`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => {
  return () => {};
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 13, EndLine: 3, EndColumn: 15},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => {
  return function () {};
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 10, EndLine: 3, EndColumn: 19},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export function fn() {
  return () => {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 13, EndLine: 3, EndColumn: 15},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export function fn() {
  return function () {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 10, EndLine: 3, EndColumn: 19},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// deeply nested higher-order: only innermost needs return type
		{
			Code: `
export function FunctionDeclaration() {
  return function FunctionExpression_Within_FunctionDeclaration() {
    return function FunctionExpression_Within_FunctionExpression() {
      return () => {
        // ArrowFunctionExpression_Within_FunctionExpression
        return () =>
          // ArrowFunctionExpression_Within_ArrowFunctionExpression
          () =>
            1; // ArrowFunctionExpression_Within_ArrowFunctionExpression_WithNoBody
      };
    };
  };
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 9, Column: 14, EndLine: 9, EndColumn: 16},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export default () => () => {
  return () => {
    return;
  };
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 13, EndLine: 3, EndColumn: 15},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// as-cast not `as const` — still needs return type
		{
			Code: `
export const func1 = (value: number) => ({ type: 'X', value }) as any;
export const func2 = (value: number) => ({ type: 'X', value }) as Action;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 38, EndLine: 2, EndColumn: 40},
				{MessageId: "missingReturnType", Line: 3, Column: 38, EndLine: 3, EndColumn: 40},
			},
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// allowDirectConstAssertionInArrowFunctions: false
		{
			Code: `
export const func = (value: number) => ({ type: 'X', value }) as const;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 37, EndLine: 2, EndColumn: 39},
			},
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": false},
		},
		{
			Code: `
interface R {
  type: string;
  value: number;
}

export const func = (value: number) =>
  ({ type: 'X', value }) as const satisfies R;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 7, Column: 37, EndLine: 7, EndColumn: 39},
			},
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": false},
		},
		// allowedNames: ['prop'] — `method`/`foo` still need types
		{
			Code: `
export class Test {
  constructor() {}
  get prop() {
    return 1;
  }
  set prop() {}
  method() {
    return;
  }
  arrow = (): string => 'arrow';
  foo = () => 'bar';
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 8, Column: 3, EndLine: 8, EndColumn: 9},
				{MessageId: "missingReturnType", Line: 12, Column: 3, EndLine: 12, EndColumn: 9},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{"prop"}},
		},
		// parameter property without a type
		{
			Code: `
export class Test {
  constructor(public foo) {}
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3, Column: 22},
			},
		},
		// allowedNames excluding only func2 — func1 still flagged
		{
			Code: `
export const func1 = (value: number) => value;
export const func2 = (value: number) => value;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 38, EndLine: 2, EndColumn: 40},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{"func2"}},
		},
		// missing arg type
		{
			Code: `
export function fn(test): string {
  return '123';
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2, Column: 20, EndLine: 2, EndColumn: 24},
			},
		},
		{
			Code: `
export const fn = (one: number, two): string => '123';
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2, Column: 33, EndLine: 2, EndColumn: 36},
			},
		},
		// higher-order untyped both
		{
			Code: `
export function foo(outer) {
  return function (inner) {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// allowDirectConstAssertionInArrowFunctions doesn't help untyped arg
		{
			Code: `export const baz = arg => arg as const;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 1},
			},
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// followReference: untyped const followed by export default
		{
			Code: `
const foo = arg => arg;
export default foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		{
			Code: `
const foo = arg => arg;
export = foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		// followReference with reassignment
		{
			Code: `
let foo = (arg: number): number => arg;
foo = arg => arg;
export default foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
				{MessageId: "missingReturnType", Line: 3},
			},
		},
		// export default [foo]
		{
			Code: `
const foo = arg => arg;
export default [foo];
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		{
			Code: `
const foo = arg => arg;
export default { foo };
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		// function declaration default-exported via identifier
		{
			Code: `
function foo(arg) {
  return arg;
}
export default foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
				{MessageId: "missingArgType", Line: 2},
			},
		},
		{
			Code: `
function foo(arg) {
  return arg;
}
export default [foo];
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
				{MessageId: "missingArgType", Line: 2},
			},
		},
		{
			Code: `
function foo(arg) {
  return arg;
}
export default { foo };
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
				{MessageId: "missingArgType", Line: 2},
			},
		},
		{
			Code: `
const bar = function foo(arg) {
  return arg;
};
export default { bar };
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
				{MessageId: "missingArgType", Line: 2},
			},
		},
		// class default-exported
		{
			Code: `
class Foo {
  bool(arg) {
    return arg;
  }
}
export default Foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  bool = arg => {
    return arg;
  };
}
export default Foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  bool = function (arg) {
    return arg;
  };
}
export default Foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},
		// accessor field
		{
			Code: `
class Foo {
  accessor bool = arg => {
    return arg;
  };
}
export default Foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
				{MessageId: "missingReturnType", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  accessor bool = function (arg) {
    return arg;
  };
}
export default Foo;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},
		{
			Code: `
class Foo {
  bool = function (arg) {
    return arg;
  };
}
export default [Foo];
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},
		// let reassignment then export default
		{
			Code: `
let test = arg => argl;
test = (): void => {
  return;
};
export default test;
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		{
			Code: `
let test = arg => argl;
test = (): void => {
  return;
};
export { test };
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},
		// allowHigherOrderFunctions: false
		{
			Code: `
export const foo =
  () =>
  (a: string): ((n: number) => string) => {
    return function (n) {
      return String(n);
    };
  };
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 6},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": false},
		},
		// inner arrow untyped
		{
			Code: `
export var arrowFn = () => () => {};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 31},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
export function fn() {
  return function () {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 10},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// outer and inner untyped args
		{
			Code: `
export function foo(outer) {
  return function (inner): void {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2, Column: 21},
				{MessageId: "missingArgType", Line: 3, Column: 20},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// only one of the returned branches is a function — not higher-order
		{
			Code: `
export function foo(outer: boolean) {
  if (outer) {
    return 'string';
  }
  return function (inner): void {};
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 8},
			},
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// destructuring patterns
		{
			Code: `
export function foo({ foo }): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgTypeUnnamed", Line: 2, Column: 21},
			},
		},
		{
			Code: `
export function foo([bar]): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgTypeUnnamed", Line: 2, Column: 21},
			},
		},
		{
			Code: `
export function foo(...bar): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2, Column: 21},
			},
		},
		{
			Code: `
export function foo(...[a]): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgTypeUnnamed", Line: 2, Column: 21},
			},
		},
		// allowArgumentsExplicitlyTypedAsAny: false (default)
		{
			Code: `
export function foo(foo: any): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyTypedArg", Line: 2, Column: 21},
			},
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": false},
		},
		{
			Code: `
export function foo({ foo }: any): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyTypedArgUnnamed", Line: 2, Column: 21},
			},
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": false},
		},
		{
			Code: `
export function foo([bar]: any): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyTypedArgUnnamed", Line: 2, Column: 21},
			},
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": false},
		},
		{
			Code: `
export function foo(...bar: any): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyTypedArg", Line: 2, Column: 21},
			},
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": false},
		},
		{
			Code: `
export function foo(...[a]: any): void {}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "anyTypedArgUnnamed", Line: 2, Column: 21},
			},
			Options: map[string]interface{}{"allowArgumentsExplicitlyTypedAsAny": false},
		},
		// allowedNames: [] (default — restated)
		{
			Code: `
export function func1() {
  return 0;
}
export const foo = {
  func2() {
    return 0;
  },
};
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 8, EndLine: 2, EndColumn: 22},
				{MessageId: "missingReturnType", Line: 6, Column: 3, EndLine: 6, EndColumn: 8},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{}},
		},
		// allowOverloadFunctions: default false — implementation needs return type
		{
			Code: `
export function test(a: string): string;
export function test(a: number): number;
export function test(a: unknown) {
  return a;
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 8, EndColumn: 21},
			},
		},
		{
			Code: `
export default function test(a: string): string;
export default function test(a: number): number;
export default function test(a: unknown) {
  return a;
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 16, EndColumn: 29},
			},
		},
		{
			Code: `
export default function (a: string): string;
export default function (a: number): number;
export default function (a: unknown) {
  return a;
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 16, EndColumn: 25},
			},
		},
		{
			Code: `
export class Test {
  test(a: string): string;
  test(a: number): number;
  test(a: unknown) {
    return a;
  }
}
		`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndColumn: 7},
			},
		},
	})
}
