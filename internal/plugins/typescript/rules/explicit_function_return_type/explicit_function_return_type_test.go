package explicit_function_return_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitFunctionReturnTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// return statement at top level
		{Code: `return;`},
		// Functions with explicit return types
		{Code: `
function test(): void {
  return;
}
		`},
		{Code: `
var fn = function(): number {
  return 1;
};
		`},
		{Code: `
var arrowFn = (): string => 'test';
		`},
		// Class with constructor, getter, setter, methods with return types
		{Code: `
class Test {
  constructor() {}
  get prop(): number {
    return 1;
  }
  set prop() {}
  method(): void {
    return;
  }
  arrow = (): string => 'arrow';
}
		`},
		// allowExpressions: true
		{
			Code:    `fn(() => {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		{
			Code:    `fn(function() {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		{
			Code:    `[function() {}, () => {}];`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		{
			Code:    `(function() {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		{
			Code:    `(() => {})();`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		{
			Code:    `export default (): void => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// allowTypedFunctionExpressions: true
		{
			Code:    `var arrowFn: Foo = () => 'test';`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
var funcExpr: Foo = function() {
  return 'test';
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code:    `const x = (() => {}) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code:    `const x = <Foo>(() => {});`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
const x = {
  foo: () => {},
} as Foo;
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
const x = <Foo>{
  foo: () => {},
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
const x: Foo = {
  foo: () => {},
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Nested object with type assertion
		{
			Code: `
const x = {
  foo: { bar: () => {} },
} as Foo;
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
const x = <Foo>{
  foo: { bar: () => {} },
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
const x: Foo = {
  foo: { bar: () => {} },
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Class property with type annotation
		{
			Code: `
type MethodType = () => void;

class App {
  private method: MethodType = () => {};
}
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Setter in object literal
		{Code: `
const myObj = {
  set myProp(val) {
    this.myProp = val;
  },
};
		`},
		// allowHigherOrderFunctions: true
		{
			Code:    `() => (): void => {};`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code:    `() => function(): void {};`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
() => {
  return (): void => {};
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
() => {
  return function(): void {};
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
function fn() {
  return (): void => {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
function fn() {
  return function(): void {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		{
			Code: `
function FunctionDeclaration() {
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
() => () => {
  return (): void => {
    return;
  };
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// allowTypedFunctionExpressions: call expression arguments
		{
			Code: `
declare function foo(arg: () => void): void;
foo(() => 1);
foo(() => {});
foo(() => null);
foo(() => true);
foo(() => '');
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Optional chain call expression arguments
		{
			Code: `
declare function foo(arg: () => void): void;
foo?.(() => 1);
foo?.bar(() => {});
foo?.bar?.(() => null);
foo.bar?.(() => true);
foo?.(() => '');
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// New expression arguments
		{
			Code: `
new Promise(resolve => {});
new Foo(1, () => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Typed function expression in call
		{
			Code: `
class Accumulator {
  private count: number = 0;

  public accumulate(fn: () => number): void {
    this.count += fn();
  }
}

new Accumulator().accumulate(() => 1);
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// Object method in typed call expression
		{
			Code: `
declare function foo(arg: { meth: () => number }): void;
foo({
  meth() {
    return 1;
  },
});
foo({
  meth: function() {
    return 1;
  },
});
foo({
  meth: () => {
    return 1;
  },
});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// allowDirectConstAssertionInArrowFunctions: true
		{
			Code: `
const func = (value: number) => ({ type: 'X', value }) as const;
const func = (value: number) => ({ type: 'X', value }) as const;
const func = (value: number) => x as const;
const func = (value: number) => x as const;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// as const satisfies
		{
			Code: `
interface R {
  type: string;
  value: number;
}

const func = (value: number) => ({ type: 'X', value }) as const satisfies R;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// nested satisfies
		{
			Code: `
interface R {
  type: string;
  value: number;
}

const func = (value: number) =>
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

const func = (value: number) =>
  ({ type: 'X', value }) as const satisfies R satisfies R satisfies R;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
		},
		// allowConciseArrowFunctionExpressionsStartingWithVoid: true
		{
			Code:    `const log = (message: string) => void console.log(message);`,
			Options: map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": true},
		},
		// allowFunctionsWithoutTypeParameters: true
		{
			Code:    `const log = (a: string) => a;`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code:    `const log = <A,>(a: A): A => a;`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code: `
function log<A>(a: A): A {
  return a;
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code: `
function log(a: string) {
  return a;
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code: `
const log = function <A>(a: A): A {
  return a;
};
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		{
			Code: `
const log = function(a: A): string {
  return a;
};
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
		},
		// allowedNames
		{
			Code: `
function test1() {
  return;
}

const foo = function test2() {
  return;
};
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"test1", "test2"}},
		},
		{
			Code: `
const test1 = function() {
  return;
};
const foo = function() {
  return function test2() {};
};
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"test1", "test2"}},
		},
		{
			Code: `
const test1 = () => {
  return;
};
export const foo = {
  test2() {
    return 0;
  },
};
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"test1", "test2"}},
		},
		{
			Code: `
class Test {
  constructor() {}
  get prop() {
    return 1;
  }
  set prop() {}
  method() {
    return;
  }
  arrow = () => 'arrow';
  private method() {
    return;
  }
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"prop", "method", "arrow"}},
		},
		{
			Code: `
const x = {
  arrowFn: () => {
    return;
  },
  fn: function() {
    return;
  },
};
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"arrowFn", "fn"}},
		},
		// Combined higher order + typed expressions
		{
			Code: `
type HigherOrderType = () => (arg1: string) => (arg2: number) => string;
const x: HigherOrderType = () => arg1 => arg2 => 'foo';
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions":     true,
				"allowTypedFunctionExpressions": true,
			},
		},
		{
			Code: `
type HigherOrderType = () => (arg1: string) => (arg2: number) => string;
const x: HigherOrderType = () => arg1 => arg2 => 'foo';
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions":     false,
				"allowTypedFunctionExpressions": true,
			},
		},
		// allowIIFEs: true
		{
			Code: `
let foo = function(): number {
  return 1;
};
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
const foo = (function() {
  return 1;
})();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
const foo = (() => {
  return 1;
})();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
const foo = (() => (() => 'foo')())();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
let foo = (() => (): string => {
  return 'foo';
})()();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
let foo = (() => (): string => {
  return 'foo';
})();
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions": false,
				"allowIIFEs":                true,
			},
		},
		{
			Code: `
let foo = (() => (): void => {})()();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		{
			Code: `
let foo = (() => (() => {})())();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		// Class property with typed initializer
		{Code: `
class Bar {
  bar: Foo = {
    foo: x => x + 1,
  };
}
		`},
		// Default parameter with type annotation
		{
			Code: `
type CallBack = () => void;

function f(gotcha: CallBack = () => {}): void {}
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
type CallBack = () => void;

const f = (gotcha: CallBack = () => {}): void => {};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		{
			Code: `
type ObjectWithCallback = { callback: () => void };

const f = (gotcha: ObjectWithCallback = { callback: () => {} }): void => {};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
	}, []rule_tester.InvalidTestCase{
		// Basic missing return types
		{
			Code: `
function test(a: number, b: number) {
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 1}},
		},
		{
			Code: `
function test() {
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 1}},
		},
		{
			Code: `
var fn = function() {
  return 1;
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 10}},
		},
		{
			Code:   `var arrowFn = () => 'test';`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 18}},
		},
		// Class with missing return types
		{
			Code: `
class Test {
  constructor() {}
  get prop() {
    return 1;
  }
  set prop() {}
  method() {
    return;
  }
  arrow = () => 'arrow';
  private method() {
    return;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3},
				{MessageId: "missingReturnType", Line: 8, Column: 3},
				{MessageId: "missingReturnType", Line: 11, Column: 3},
				{MessageId: "missingReturnType", Line: 12, Column: 3},
			},
		},
		// allowExpressions: true - declarations still need types
		{
			Code: `
function test() {
  return;
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 1}},
		},
		{
			Code:    `const foo = () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		{
			Code:    `const foo = function() {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 13}},
		},
		{
			Code:    `export default () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 19}},
		},
		{
			Code:    `export default function() {}`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 16}},
		},
		// Class properties with allowExpressions
		{
			Code: `
class Foo {
  public a = () => {};
  public b = function() {};
  public c = function test() {};

  static d = () => {};
  static e = function() {};
}
			`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 3},
				{MessageId: "missingReturnType", Line: 4, Column: 3},
				{MessageId: "missingReturnType", Line: 5, Column: 3},
				{MessageId: "missingReturnType", Line: 7, Column: 3},
				{MessageId: "missingReturnType", Line: 8, Column: 3},
			},
		},
		// allowTypedFunctionExpressions: true - untyped expressions still need types
		{
			Code:    `var arrowFn = () => 'test';`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 18}},
		},
		{
			Code: `
var funcExpr = function() {
  return 'test';
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 16}},
		},
		// allowTypedFunctionExpressions: false - type assertions don't help
		{
			Code:    `const x = (() => {}) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 15}},
		},
		// Higher order functions with missing inner types
		{
			Code:    `() => () => {};`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 10}},
		},
		{
			Code:    `() => function() {};`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 7}},
		},
		{
			Code: `
() => {
  return () => {};
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 13}},
		},
		{
			Code: `
() => {
  return function() {};
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 10}},
		},
		{
			Code: `
function fn() {
  return () => {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 13}},
		},
		{
			Code: `
function fn() {
  return function() {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 10}},
		},
		// Higher order: not all returns are functions
		{
			Code: `
function fn(arg: boolean) {
  if (arg) return 'string';
  return function(): void {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 1}},
		},
		// allowTypedFunctionExpressions: false - call args don't help
		{
			Code: `
declare function foo(arg: () => void): void;
foo(() => 1);
foo(() => {});
foo(() => null);
foo(() => true);
foo(() => '');
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 8},
				{MessageId: "missingReturnType", Line: 4, Column: 8},
				{MessageId: "missingReturnType", Line: 5, Column: 8},
				{MessageId: "missingReturnType", Line: 6, Column: 8},
				{MessageId: "missingReturnType", Line: 7, Column: 8},
			},
		},
		// allowTypedFunctionExpressions: false - class accumulate
		{
			Code: `
class Accumulator {
  private count: number = 0;

  public accumulate(fn: () => number): void {
    this.count += fn();
  }
}

new Accumulator().accumulate(() => 1);
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 10, Column: 33}},
		},
		// allowDirectConstAssertionInArrowFunctions: non-const assertions
		{
			Code: `
const func = (value: number) => ({ type: 'X', value }) as any;
const func = (value: number) => ({ type: 'X', value }) as Action;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 30},
				{MessageId: "missingReturnType", Line: 3, Column: 30},
			},
		},
		// allowDirectConstAssertionInArrowFunctions: false
		{
			Code: `
const func = (value: number) => ({ type: 'X', value }) as const;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 30}},
		},
		// allowConciseArrowFunctionExpressionsStartingWithVoid: false
		{
			Code:    `const log = (message: string) => void console.log(message);`,
			Options: map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 31}},
		},
		// allowConciseArrowFunctionExpressionsStartingWithVoid: true - block body still needs type
		{
			Code: `
        const log = (message: string) => {
          void console.log(message);
        };
			`,
			Options: map[string]interface{}{"allowConciseArrowFunctionExpressionsStartingWithVoid": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 39}},
		},
		// allowFunctionsWithoutTypeParameters: true - generic functions still need types
		{
			Code:    `const log = <A,>(a: A) => a;`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
		{
			Code: `
function log<A>(a: A) {
  return a;
}
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
		{
			Code: `
const log = function <A>(a: A) {
  return a;
};
			`,
			Options: map[string]interface{}{"allowFunctionsWithoutTypeParameters": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType"}},
		},
		// allowedNames - non-matching names still need types
		{
			Code: `
function hoge() {
  return;
}
const foo = () => {
  return;
};
const baz = function() {
  return;
};
let [test, test] = function() {
  return;
};
class X {
  [test] = function() {
    return;
  };
}
const x = {
  1: function() {
    reutrn; // cspell:disable-line
  },
};
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"test", "1"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 1},
				{MessageId: "missingReturnType", Line: 5, Column: 16},
				{MessageId: "missingReturnType", Line: 8, Column: 13},
				{MessageId: "missingReturnType", Line: 11, Column: 20},
				{MessageId: "missingReturnType", Line: 15, Column: 3},
				{MessageId: "missingReturnType", Line: 20, Column: 3},
			},
		},
		// allowedNames with computed property
		{
			Code: `
const ignoredName = 'notIgnoredName';
class Foo {
  [ignoredName]() {}
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"ignoredName"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 3}},
		},
		// Untyped class property array
		{
			Code: `
class Bar {
  bar = [
    {
      foo: x => x + 1,
    },
  ];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 5, Column: 7}},
		},
		// IIFE: false
		{
			Code: `
const foo = (function() {
  return 'foo';
})();
			`,
			Options: map[string]interface{}{"allowIIFEs": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 14}},
		},
		// IIFE: inner function still needs type
		{
			Code: `
const foo = (function() {
  return () => {
    return 1;
  };
})();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 13}},
		},
		// Non-IIFE function expression with allowIIFEs
		{
			Code: `
let foo = function() {
  return 'foo';
};
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 11}},
		},
		// IIFE returning arrow
		{
			Code: `
let foo = (() => () => {})()();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 21}},
		},
		// Default parameter with allowTypedFunctionExpressions: false
		{
			Code: `
type CallBack = () => void;

function f(gotcha: CallBack = () => {}): void {}
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 34}},
		},
		{
			Code: `
type CallBack = () => void;

const f = (gotcha: CallBack = () => {}): void => {};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 34}},
		},
	})
}

// TestEdgeCases covers exhaustive edge cases across all dimensions.
func TestEdgeCases(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitFunctionReturnTypeRule, []rule_tester.ValidTestCase{
		// ============================================================
		// Dimension 1: All function-like AST node types with return types
		// ============================================================
		{Code: `function foo(): void {}`},
		{Code: `const foo = function(): void {}`},
		{Code: `const foo = (): void => {}`},
		{Code: `class A { method(): void {} }`},
		{Code: `class A { get prop(): number { return 1; } }`},
		{Code: `class A { set prop(v: number) {} }`},
		{Code: `class A { constructor() {} }`},
		{Code: `const obj = { method(): void {} }`},
		{Code: `const obj = { get prop(): number { return 1; } }`},
		{Code: `const obj = { set prop(v: number) {} }`},
		// Async variants
		{Code: `async function foo(): Promise<void> {}`},
		{Code: `const foo = async (): Promise<void> => {}`},
		{Code: `const foo = async function(): Promise<void> {}`},
		{Code: `class A { async method(): Promise<void> {} }`},
		// Generator variants
		{Code: `function* foo(): Generator { yield 1; }`},

		// ============================================================
		// Dimension 2: allowTypedFunctionExpressions — every typed context
		// ============================================================
		// 2a. Type assertion: `as T`
		{
			Code:    `const x = (() => {}) as Foo;`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2b. Type assertion: `<T>x`
		{
			Code:    `const x = <Foo>(() => {});`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2c. Variable declarator with type annotation
		{
			Code:    `const x: Foo = () => {};`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2d. Class property with type annotation
		{
			Code: `
class App {
  handler: () => void = () => {};
}
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2e. Call expression argument (function is typed by parameter)
		{
			Code: `
declare function foo(cb: () => void): void;
foo(() => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2f. New expression argument
		{
			Code: `
new Promise(resolve => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2g. Default parameter with type annotation
		{
			Code: `
type Fn = () => void;
function f(cb: Fn = () => {}): void {}
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2h. Property of typed object — nested 1 level
		{
			Code: `
const x: Foo = { bar: () => {} };
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2i. Property of typed object — nested 2 levels
		{
			Code: `
const x: Foo = { a: { b: () => {} } };
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2j. Property of typed object — nested 3 levels
		{
			Code: `
const x: Foo = { a: { b: { c: () => {} } } };
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2k. Object typed via `as` with nested properties
		{
			Code: `
const x = { a: { b: { c: () => {} } } } as Foo;
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2l. Object in call argument with method shorthand
		{
			Code: `
declare function foo(arg: { fn: () => void }): void;
foo({ fn() {} });
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},
		// 2m. Object in new expression argument — direct arrow arg is typed, not nested object prop
		// NOTE: `new Foo({ callback: () => {} })` is intentionally NOT here —
		// isConstructorArgument is only checked at the top level, not inside isPropertyOfObjectWithType.
		// This matches ESLint behavior.
		// 2n. Optional call expression argument
		{
			Code: `
declare function foo(cb: () => void): void;
foo?.(() => {});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": true},
		},

		// ============================================================
		// Dimension 3: ancestorHasReturnType — function returns object with arrows
		// ============================================================
		// 3a. Function with return type returning object with arrow property
		{
			Code: `
interface Foo { bar: () => string; }
function foo(): Foo {
  return { bar: () => 'test' };
}
			`,
			Options: map[string]interface{}{
				"allowTypedFunctionExpressions": true,
				"allowHigherOrderFunctions":     true,
			},
		},
		// 3b. Arrow with return type returning object with arrow property
		{
			Code: `
interface Foo { bar: () => string; }
const foo = (): Foo => ({ bar: () => 'test' });
			`,
			Options: map[string]interface{}{
				"allowTypedFunctionExpressions": true,
				"allowHigherOrderFunctions":     true,
			},
		},
		// 3c. Typed variable with higher-order function returning arrow in object
		{
			Code: `
interface Foo { bar: () => string; }
const foo: () => Foo = () => ({ bar: () => 'test' });
			`,
			Options: map[string]interface{}{
				"allowTypedFunctionExpressions": true,
				"allowHigherOrderFunctions":     true,
			},
		},

		// ============================================================
		// Dimension 4: allowHigherOrderFunctions — deeply nested
		// ============================================================
		// 4a. Three levels of HOF with final typed return
		{
			Code: `
function a() {
  return function b() {
    return (): void => {};
  };
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// 4b. HOF: arrow with block body and multiple returns all returning functions
		{
			Code: `
function fn(arg: boolean) {
  if (arg) {
    return () => (): number => 1;
  } else {
    return function(): string { return 'foo'; };
  }
  return function(): void {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},
		// 4c. HOF: mixed with extra statements between returns
		{
			Code: `
() => {
  const foo = 'foo';
  return function(): string {
    return foo;
  };
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
		},

		// ============================================================
		// Dimension 5: allowIIFEs — complex nesting
		// ============================================================
		// 5a. IIFE with explicit return type on inner
		{
			Code: `
const foo = ((arg: number): number => {
  return arg;
})(0);
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		// 5b. Nested IIFE inside IIFE
		{
			Code: `
const foo = (() => (() => 'foo')())();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
		},
		// 5c. IIFE + higher order: outer is IIFE, inner has return type
		{
			Code: `
let foo = (() => (): string => {
  return 'foo';
})()();
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions": true,
				"allowIIFEs":                true,
			},
		},

		// ============================================================
		// Dimension 6: allowExpressions — expression contexts
		// ============================================================
		// 6a. Ternary expression
		{
			Code:    `const x = true ? () => {} : () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// 6b. Logical expression
		{
			Code:    `const x = false || (() => {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},
		// 6c. Comma expression
		{
			Code:    `(0, () => {});`,
			Options: map[string]interface{}{"allowExpressions": true},
		},

		// ============================================================
		// Dimension 7: allowedNames — with class methods and object methods
		// ============================================================
		// 7a. Class getter with allowed name
		{
			Code: `
class Foo {
  get bar() { return 1; }
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"bar"}},
		},
		// 7b. Object method shorthand with allowed name
		{
			Code: `
const obj = { foo() { return 1; } };
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"foo"}},
		},
		// 7c. Named function expression matches allowed name
		{
			Code: `
const x = function namedFn() { return 1; };
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"namedFn"}},
		},

		// ============================================================
		// Dimension 8: Class property typed array with nested objects
		// ============================================================
		{Code: `
class Bar {
  bar: Foo[] = [
    {
      foo: x => x + 1,
    },
  ];
}
		`},
	}, []rule_tester.InvalidTestCase{
		// ============================================================
		// Dimension 1: Basic function types without return types
		// ============================================================
		// 1a. Async function without return type
		{
			Code:   `async function foo() {}`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 1}},
		},
		// 1b. Async arrow without return type
		{
			Code:   `const foo = async () => {};`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 22}},
		},
		// 1c. Async method without return type
		{
			Code: `
class A {
  async method() {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// 1d. Object getter without return type
		{
			Code: `
const obj = {
  get foo() { return 1; },
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// 1e. Object method shorthand without return type
		{
			Code: `
const obj = {
  foo() { return 1; },
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		// 1f. Generator function without return type
		{
			Code:   `function* foo() { yield 1; }`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 1}},
		},

		// ============================================================
		// Dimension 2: allowTypedFunctionExpressions: false — every context is invalid
		// ============================================================
		// 2a. Object property in typed context
		{
			Code: `
interface Foo {}
const x: Foo = {
  foo: () => {},
};
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 3}},
		},
		// 2b. Object method shorthand in typed call
		{
			Code: `
declare function foo(arg: { meth: () => number }): void;
foo({
  meth() {
    return 1;
  },
});
foo({
  meth: function() {
    return 1;
  },
});
foo({
  meth: () => {
    return 1;
  },
});
			`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4, Column: 3},
				{MessageId: "missingReturnType", Line: 9, Column: 3},
				{MessageId: "missingReturnType", Line: 14, Column: 3},
			},
		},
		// 2c. IIFE with allowTypedFunctionExpressions: false
		{
			Code:    `(() => true)();`,
			Options: map[string]interface{}{"allowTypedFunctionExpressions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 5}},
		},

		// ============================================================
		// Dimension 3: allowHigherOrderFunctions — edge cases
		// ============================================================
		// 3a. Not all returns are function expressions
		{
			Code: `
function fn() {
  const bar = () => (): number => 1;
  const baz = () => () => 'baz';
  return function(): void {};
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 24}},
		},
		// 3b. Deeply nested HOF: innermost function still needs return type
		{
			Code: `
function FunctionDeclaration() {
  return function FunctionExpression_Within_FunctionDeclaration() {
    return function FunctionExpression_Within_FunctionExpression() {
      return () => {
        return () =>
          () =>
            1;
      };
    };
  };
}
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 7, Column: 14}},
		},
		// 3c. HOF where inner function returns non-function in some branches
		{
			Code: `
() => () => {
  return () => {
    return;
  };
};
			`,
			Options: map[string]interface{}{"allowHigherOrderFunctions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 13}},
		},

		// ============================================================
		// Dimension 4: allowTypedFunctionExpressions — inner arrow in higher-order typed context
		// ============================================================
		// 4a. Inner arrow of higher-order typed function without return type
		{
			Code: `
type HigherOrderType = () => (arg1: string) => (arg2: number) => string;
const x: HigherOrderType = () => arg1 => arg2 => 'foo';
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions":     true,
				"allowTypedFunctionExpressions": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 47},
			},
		},
		// 4b. All three levels invalid when both options are false.
		// Exit listeners fire innermost-first, so errors are reported in reverse position order.
		{
			Code: `
type HigherOrderType = () => (arg1: string) => (arg2: number) => string;
const x: HigherOrderType = () => arg1 => arg2 => 'foo';
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions":     false,
				"allowTypedFunctionExpressions": false,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3, Column: 47},
				{MessageId: "missingReturnType", Line: 3, Column: 39},
				{MessageId: "missingReturnType", Line: 3, Column: 31},
			},
		},

		// ============================================================
		// Dimension 5: Class property arrow in typed context that IS reportable
		// ============================================================
		// 5a. Class property without type annotation — arrow inside is invalid
		{
			Code: `
class Foo {
  foo = () => () => {
    return console.log('foo');
  };
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 18}},
		},
		// 5b. Nested arrow inside untyped variable assignment
		{
			Code: `
function foo(): any {
  const bar = () => () => console.log('aa');
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 24}},
		},
		// 5c. Nested arrow inside anyValue assignment
		{
			Code: `
let anyValue: any;
function foo(): any {
  anyValue = () => () => console.log('aa');
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 4, Column: 23}},
		},

		// ============================================================
		// Dimension 6: allowIIFEs edge cases
		// ============================================================
		// 6a. IIFE returning another function — inner needs type
		{
			Code: `
let foo = (() => () => {})()();
			`,
			Options: map[string]interface{}{"allowIIFEs": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 2, Column: 21}},
		},

		// ============================================================
		// Dimension 7: allowDirectConstAssertionInArrowFunctions edge cases
		// ============================================================
		// 7a. as const satisfies with allowDirectConstAssertionInArrowFunctions: false
		{
			Code: `
interface R {
  type: string;
  value: number;
}

const func = (value: number) => ({ type: 'X', value }) as const satisfies R;
			`,
			Options: map[string]interface{}{"allowDirectConstAssertionInArrowFunctions": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 7, Column: 30}},
		},

		// ============================================================
		// Dimension 8: allowExpressions — declaration contexts are still invalid
		// ============================================================
		// 8a. export default with allowExpressions still reports on assignment arrow
		{
			Code:    `export default () => {};`,
			Options: map[string]interface{}{"allowExpressions": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 1, Column: 19}},
		},

		// ============================================================
		// Dimension 9: Untyped class property — nested object with arrow
		// ============================================================
		{
			Code: `
class Bar {
  bar = [
    {
      foo: x => x + 1,
    },
  ];
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 5, Column: 7}},
		},

		// ============================================================
		// Dimension 10: Object property with function expression — no typed context
		// ============================================================
		{
			Code: `
const x = {
  foo: () => {},
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},
		{
			Code: `
const x = {
  foo: function() {},
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingReturnType", Line: 3, Column: 3}},
		},

		// ============================================================
		// Dimension 11: Multiple options interacting
		// ============================================================
		// 11a. allowExpressions: true — declaration (var assignment) still needs type
		{
			Code: `
const foo = () => {};
			`,
			Options: map[string]interface{}{
				"allowExpressions":          true,
				"allowHigherOrderFunctions": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2, Column: 16},
			},
		},
	})
}
