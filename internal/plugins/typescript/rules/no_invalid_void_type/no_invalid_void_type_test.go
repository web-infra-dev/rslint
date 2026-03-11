package no_invalid_void_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidVoidTypeRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidVoidTypeRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// === allowInGenericTypeArguments: false ===
			{Code: `type Generic<T> = [T];`, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},
			{Code: `
function foo(): void | never {
  throw new Error('Test');
}
      `, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},
			{Code: `type voidNeverUnion = void | never;`, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},
			{Code: `type neverVoidUnion = never | void;`, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},

			// === allowInGenericTypeArguments: true (default) ===
			{Code: `function func(): void {}`},
			{Code: `type NormalType = () => void;`},
			{Code: `interface Callable { (...args: string[]): void; }`},
			{Code: `interface Constructable { new (...args: string[]): void; }`},
			{Code: `let normalArrow = (): void => {};`},
			{Code: `let voidPromise: Promise<void> = new Promise<void>(() => {});`},
			{Code: `let voidMap: Map<string, void> = new Map<string, void>();`},
			{Code: `async function returnsVoidPromiseAsync(): Promise<void> {}`},
			{Code: `type GenericVoid = Generic<void>;`},
			{Code: `type voidPromiseUnion = void | Promise<void>;`},
			{Code: `type promiseNeverUnion = Promise<void> | never;`},
			{Code: `const arrowGeneric1 = <T = void,>(arg: T) => {};`},
			{Code: `declare function functionDeclaration1<T = void>(arg: T): void;`},
			{Code: `type FunctionType = (x: number) => void;`},
			{Code: `interface Foo { method(): void; }`},
			// Callable signatures should be valid even with allowInGenericTypeArguments: false
			{Code: `interface Callable { (...args: string[]): void; }`, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},
			{Code: `interface Constructable { new (...args: string[]): void; }`, Options: map[string]interface{}{"allowInGenericTypeArguments": false}},
			{Code: `interface GenericCallable { <T>(arg: T): void; }`},

			// === void in heritage clause type arguments (extends/implements) ===
			{Code: `
interface IObservable<T, U> {}
export interface IObservableSignal<TChange> extends IObservable<void, TChange> {}
`},
			{Code: `
interface Base<T> {}
class Foo implements Base<void> {}
`},
			// Multiple extends with void
			{Code: `
interface A<T> {}
interface B<T> {}
interface C extends A<void>, B<void> {}
`},
			// Nested generics in extends
			{Code: `
interface Wrapper<T> {}
interface Foo extends Wrapper<Promise<void>> {}
`},
			// Multiple type arguments including void
			{Code: `
interface Multi<A, B, C> {}
interface Foo extends Multi<string, void, number> {}
`},
			// Class implements multiple interfaces with void
			{Code: `
interface A<T> {}
interface B<T> {}
class Foo implements A<void>, B<void> {}
`},
			// Deep nesting: extends with generic of generic of void
			{Code: `
interface Deep<T> {}
interface Foo extends Deep<Map<string, void>> {}
`},

			// === type alias with void as generic type argument ===
			{Code: `type Foo = Promise<void>;`},
			{Code: `type Foo = Map<string, void>;`},
			{Code: `type Foo = Promise<void> & Map<string, void>;`},
			{Code: `type Foo = Array<void>;`},
			{Code: `type Foo = Readonly<Promise<void>>;`},

			// === allowInGenericTypeArguments: whitelist ===
			{Code: `type AllowedVoid = Allowed<void>;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Allowed"}}},
			{Code: `type voidPromiseUnion = void | Promise<void>;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Promise"}}},
			{Code: `type promiseVoidUnion = Promise<void> | void;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Promise"}}},
			{Code: `type promiseNeverUnion = Promise<void> | never;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Promise"}}},
			{Code: `type voidPromiseNeverUnion = void | Promise<void> | never;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Promise"}}},
			// Heritage clause with whitelist - in whitelist
			{Code: `
interface Base<T> {}
interface Foo extends Base<void> {}
`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Base"}}},
			{Code: `
interface A<T> {}
interface B<T> {}
interface Foo extends A<void>, B<void> {}
`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"A", "B"}}},

			// === allowAsThisParameter: true ===
			{Code: `function f(this: void) {}`, Options: map[string]interface{}{"allowAsThisParameter": true}},
			{Code: `
class Test {
  public static helper(this: void) {}
  method(this: void) {}
}
      `, Options: map[string]interface{}{"allowAsThisParameter": true}},

			// === void operator expressions (not type) ===
			{Code: `let ughThisThing = void 0;`},
			{Code: `function takeThing(thing: undefined) {}`},
			{Code: `takeThing(void 0);`},

			// === accessor property with non-void type ===
			{Code: `
class ClassName {
  accessor propName: number;
}
`},

			// === Function overloads ===
			{Code: `
function f(): void;
function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
`},
			{Code: `
class SomeClass {
  f(): void;
  f(x: string): string;
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
class SomeClass {
  ['f'](): void;
  ['f'](x: string): string;
  ['f'](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
class SomeClass {
  [Symbol.iterator](): void;
  [Symbol.iterator](x: string): string;
  [Symbol.iterator](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
class SomeClass {
  'f'(): void;
  'f'(x: string): string;
  'f'(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
class SomeClass {
  1(): void;
  1(x: string): string;
  1(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
const staticSymbol = Symbol.for('static symbol');

class SomeClass {
  [staticSymbol](): void;
  [staticSymbol](x: string): string;
  [staticSymbol](x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
declare module foo {
  function f(): void;
  function f(x: string): string;
  function f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
{
  function f(): void;
  function f(x: string): string;
  function f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
function f(): Promise<void>;
function f(x: string): Promise<string>;
async function f(x?: string): Promise<void | string> {
  if (x !== undefined) { return x; }
}
`},
			{Code: `
class SomeClass {
  f(): Promise<void>;
  f(x: string): Promise<string>;
  async f(x?: string): Promise<void | string> {
    if (x !== undefined) { return x; }
  }
}
`},
			{Code: `
function f(): void;

const a = 5;

function f(x: string): string;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
`},
			{Code: `
export default function (): void;
export default function (x: string): string;
export default function (x?: string): string | void {
  if (x !== undefined) { return x; }
}
`},
			{Code: `
export function f(): void;
export function f(x: string): string;
export function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
`},
			{Code: `
export {};

export function f(): void;
export function f(x: string): string;
export function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}
`},

			// === Dotted generic names with whitelist ===
			{Code: `type AllowedVoid = Ex.Mx.Tx<void>;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Ex.Mx.Tx"}}},
			{Code: `type AllowedVoid = Ex . Mx . Tx<void>;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Ex.Mx.Tx"}}},
			{Code: `type AllowedVoid = Ex.Mx.Tx<void>;`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Ex . Mx . Tx"}}},

			// === void in function parameter type within whitelist ===
			{Code: `
async function foo(bar: () => void | Promise<void>) {
  await bar();
}
`, Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Promise"}}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// === allowInGenericTypeArguments: false ===
			{
				Code:    `type GenericVoid = Generic<void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code:    `function takeVoid(thing: void) {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code:    `let voidPromise: Promise<void> = new Promise<void>(() => {});`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code:    `let voidMap: Map<string, void> = new Map<string, void>();`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code:    `type invalidVoidUnion = void | number;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code: `
interface Base<T> {}
interface Foo extends Base<void> {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			// Multiple extends with void, allowInGenericTypeArguments: false
			{
				Code: `
interface A<T> {}
interface B<T> {}
interface C extends A<void>, B<void> {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			// Class implements with void, allowInGenericTypeArguments: false
			{
				Code: `
interface A<T> {}
class Foo implements A<void> {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			// Type alias with void generic, allowInGenericTypeArguments: false
			{
				Code:    `type Foo = Promise<void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},
			{
				Code:    `type Foo = Map<string, void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturn"},
				},
			},

			// === allowInGenericTypeArguments: true (default) ===
			{
				Code: `function takeVoid(thing: void) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `const arrowGeneric = <T extends void>(arg: T) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `const arrowGeneric2 = <T extends void = void>(arg: T) => {};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `function functionGeneric<T extends void>(arg: T) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `function functionGeneric2<T extends void = void>(arg: T) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `declare function functionDeclaration<T extends void>(arg: T): void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `declare function functionDeclaration2<T extends void = void>(arg: T): void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `functionGeneric<void>(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `let letVoid: void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `type VoidType = void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `type UnionType2 = string | number | void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},
			{
				Code: `type IntersectionType = string & number & void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `
interface Interface {
  lambda: () => void;
  voidProp: void;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `
class ClassName {
  private readonly propName: void;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `type invalidVoidUnion = void | Map<string, number>;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},
			{
				Code: `type invalidVoidUnion = void | Map;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},
			// Void in callable signature parameter position should be invalid
			{
				Code: `interface Foo { (arg: void): string; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === allowInGenericTypeArguments: whitelist ===
			{
				Code:    `type BannedVoid = Banned<void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Allowed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidForGeneric"},
				},
			},
			{
				Code:    `function takeVoid(thing: void) {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Allowed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			// Heritage clause with whitelist - not in whitelist
			{
				Code: `
interface Base<T> {}
interface Foo extends Base<void> {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Allowed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidForGeneric"},
				},
			},
			// Heritage clause with whitelist - multiple extends, one allowed one not
			{
				Code: `
interface Allowed<T> {}
interface Banned<T> {}
interface Foo extends Allowed<void>, Banned<void> {}`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Allowed"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidForGeneric"},
				},
			},

			// === allowAsThisParameter: true ===
			{
				Code:    `type alias = void;`,
				Options: map[string]interface{}{"allowAsThisParameter": true, "allowInGenericTypeArguments": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrThisParamOrGeneric"},
				},
			},
			{
				Code:    `type alias = void;`,
				Options: map[string]interface{}{"allowAsThisParameter": true, "allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrThisParam"},
				},
			},
			{
				Code:    `type alias = Array<void>;`,
				Options: map[string]interface{}{"allowAsThisParameter": true, "allowInGenericTypeArguments": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrThisParam"},
				},
			},

			// === void in array types ===
			{
				Code: `declare function voidArray(args: void[]): void[];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === void in type assertions ===
			{
				Code: `let value = undefined as void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `let value = <void>undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === void in rest parameter ===
			{
				Code: `function takesThings(...things: void[]): void {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === keyof void ===
			{
				Code: `type KeyofVoid = keyof void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === accessor property with void type ===
			{
				Code: `
class ClassName {
  accessor propName: void;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === void type alias with class usage ===
			{
				Code: `
type VoidType = void;
class OtherClassName {
  private propName: VoidType;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === nested union with void ===
			{
				Code: `type UnionType3 = string | ((number & any) | (string | void));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === declared function return union ===
			{
				Code: `declare function test(): number | void;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === void in union inside type parameter constraint ===
			{
				Code: `declare function test<T extends number | void>(): T;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === mapped type ===
			{
				Code: `
type MappedType<T> = {
  [K in keyof T]: void;
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === conditional type ===
			{
				Code: `
type ConditionalType<T> = {
  [K in keyof T]: T[K] extends string ? void : string;
};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === readonly void array ===
			{
				Code: `type ManyVoid = readonly void[];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},
			{
				Code: `function foo(arr: readonly void[]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidNotReturnOrGeneric"},
				},
			},

			// === Class method without overloads - invalid ===
			{
				Code: `
class SomeClass {
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === Export default function without overloads - invalid ===
			{
				Code: `export default function (x?: string): string | void {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === Export function without overloads - invalid ===
			{
				Code: `export function f(x?: string): string | void {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === Overload where intermediate signature has void union - invalid ===
			{
				Code: `
function f(): void;
function f(x: string): string | void;
function f(x?: string): string | void {
  if (x !== undefined) { return x; }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === Class method overload where intermediate signature has void union - invalid ===
			{
				Code: `
class SomeClass {
  f(): void;
  f(x: string): string | void;
  f(x?: string): string | void {
    if (x !== undefined) { return x; }
  }
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidUnionConstituent"},
				},
			},

			// === Whitelist invalid with dotted name ===
			{
				Code:    `type BannedVoid = Ex.Mx.Tx<void>;`,
				Options: map[string]interface{}{"allowInGenericTypeArguments": []interface{}{"Tx"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidVoidForGeneric"},
				},
			},
		},
	)
}
