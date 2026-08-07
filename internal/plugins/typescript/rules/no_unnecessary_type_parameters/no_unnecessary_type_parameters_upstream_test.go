// TestNoUnnecessaryTypeParametersUpstream migrates the full valid/invalid
// suite from upstream typescript-eslint's
// packages/eslint-plugin/tests/rules/no-unnecessary-type-parameters.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the
// no_unnecessary_type_parameters_extras_test.go file.
package no_unnecessary_type_parameters

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeParametersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeParametersRule, []rule_tester.ValidTestCase{
		{Code: `
class ClassyArray<T> {
  arr: T[];
}
`},
		{Code: `
class ClassyArray<T> {
  value1: T;
  value2: T;
}
`},
		{Code: `
class ClassyArray<T> {
  arr: T[];
  constructor(arr: T[]) {
    this.arr = arr;
  }
}
`},
		{Code: `
class ClassyArray<T> {
  arr: T[];
  workWith(value: T) {
    this.arr.indexOf(value);
  }
}
`},
		{Code: `
abstract class ClassyArray<T> {
  arr: T[];
  abstract workWith(value: T): void;
}
`},
		{Code: `
class Box<T> {
  val: T | null = null;
  get() {
    return this.val;
  }
}
`},
		{Code: `
class Joiner<T extends string | number> {
  join(els: T[]) {
    return els.map(el => '' + el).join(',');
  }
}
`},
		{Code: `
declare class Foo {
  getProp<T>(this: Record<'prop', T>): T;
}
`},
		{Code: `type Fn = <T>(input: T) => T;`},
		{Code: `type Fn = <T extends string>(input: T) => T;`},
		{Code: "type Fn = <T extends string>(input: T) => `a${T}b`;"},
		{Code: `type Fn = new <T>(input: T) => T;`},
		{Code: `type Fn = <T>(input: T) => typeof input;`},
		{Code: `type Fn = <T>(input: T) => keyof typeof input;`},
		{Code: `type Fn = <T>(input: Partial<T>) => typeof input;`},
		{Code: `type Fn = <T>(input: Partial<T>) => input is T;`},
		{Code: `type Fn = <T>(input: T) => { [K in keyof T]: K };`},
		{Code: `type Fn = <T>(input: T) => { [K in keyof T as K]: string };`},
		{Code: "type Fn = <T>(input: T) => { [K in keyof T as `${K & string}`]: string };"},
		{Code: `type Fn = <T>(input: T) => Partial<T>;`},
		{Code: `type Fn = <T>(input: { [i: number]: T }) => T;`},
		{Code: `type Fn = <T>(input: { [i: number]: T }) => Partial<T>;`},
		{Code: `type Fn = <T>(input: { [i: string]: T }) => Partial<T>;`},
		{Code: `type Fn = <T>(input: T) => { [i: number]: T };`},
		{Code: `type Fn = <T>(input: T) => { [i: string]: T };`},
		{Code: "type Fn = <T extends unknown[]>(input: T) => Omit<T, 'length'>;"},
		{Code: `
interface I {
  <T>(value: T): T;
}
`},
		{Code: `
interface I {
  new <T>(value: T): T;
}
`},
		{Code: `
function identity<T>(arg: T): T {
  return arg;
}
`},
		{Code: `
function printProperty<T>(obj: T, key: keyof T) {
  console.log(obj[key]);
}
`},
		{Code: `
function getProperty<T, K extends keyof T>(obj: T, key: K) {
  return obj[key];
}
`},
		{Code: `
function box<T>(val: T) {
  return { val };
}
`},
		{Code: `
function doStuff<K, V>(map: Map<K, V>, key: K) {
  let v = map.get(key);
  v = 1;
  map.set(key, v);
  return v;
}
`},
		{Code: `
function makeMap<K, V>() {
  return new Map<K, V>();
}
`},
		{Code: `
function makeMap<K, V>(ks: K[], vs: V[]) {
  const r = new Map<K, V>();
  ks.forEach((k, i) => {
    r.set(k, vs[i]);
  });
  return r;
}
`},
		{Code: `
function arrayOfPairs<T>() {
  return [] as [T, T][];
}
`},
		{Code: `
function isNonNull<T>(v: T): v is Exclude<T, null> {
  return v !== null;
}
`},
		{Code: `
function both<Args extends unknown[]>(
  fn1: (...args: Args) => void,
  fn2: (...args: Args) => void,
): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
`},
		{Code: `
function lengthyIdentity<T extends { length: number }>(x: T) {
  return x;
}
`},
		{Code: `
interface Lengthy {
  length: number;
}
function lengthyIdentity<T extends Lengthy>(x: T) {
  return x;
}
`},
		{Code: `
function ItemComponent<T>(props: { item: T; onSelect: (item: T) => void }) {}
`},
		{Code: `
interface ItemProps<T> {
  item: readonly T;
  onSelect: (item: T) => void;
}
function ItemComponent<T>(props: ItemProps<T>) {}
`},
		{Code: `
declare namespace React {
  interface RefObject<T> {
    current: T | null;
  }
}
function useFocus<T extends HTMLOrSVGElement>(): [
  React.RefObject<T>,
  () => void,
];
`},
		{Code: `
function findFirstResult<U>(
  inputs: unknown[],
  getResult: (t: unknown) => U | undefined,
): U | undefined;
`},
		{Code: `
function findFirstResult<T, U>(
  inputs: T[],
  getResult: (t: T) => () => [U | undefined],
): () => [U | undefined];
`},
		{Code: `
function getData<T>(url: string): Promise<T | null> {
  return Promise.resolve(null);
}
`},
		{Code: `
function getData<T>(url: string): Promise<T extends null ? T : null> {
  return Promise.resolve(null);
}
`},
		{Code: "function getData<T extends string>(url: string): Promise<`a${T}b`> {\n  return Promise.resolve(null);\n}\n"},
		{Code: `
async function getData<T>(url: string): Promise<T | null> {
  return null;
}
`},
		{Code: `declare function get(): void;`},
		{Code: `declare function get<T>(param: T[]): T;`},
		{Code: `declare function box<T>(val: T): { val: T };`},
		{Code: `declare function identity<T>(param: T): T;`},
		{Code: `declare function compare<T>(param1: T, param2: T): boolean;`},
		{Code: `declare function example<T>(a: Set<T>): T;`},
		{Code: `declare function example<T>(a: Set<T>, b: T[]): void;`},
		{Code: `declare function example<T>(a: Map<T, T>): void;`},
		{Code: `declare function example<T, U extends T>(t: T, u: U): U;`},
		{Code: `declare function makeSet<K>(): Set<K>;`},
		{Code: `declare function makeSet<K>(): [Set<K>];`},
		{Code: `declare function makeSets<K>(): Set<K>[];`},
		{Code: `declare function makeSets<K>(): [Set<K>][];`},
		{Code: `declare function makeMap<K, V>(): Map<K, V>;`},
		{Code: `declare function makeMap<K, V>(): [Map<K, V>];`},
		{Code: `declare function makeArray<T>(): T[];`},
		{Code: `declare function makeArrayNullish<T>(): (T | null)[];`},
		{Code: `declare function makeTupleMulti<T>(): [T | null, T | null];`},
		{Code: `declare function takeTupleMulti<T>(input: [T, T]): void;`},
		{Code: `declare function takeTupleMultiNullish<T>(input: [T | null, T | null]): void;`},
		{Code: `declare function arrayOfPairs<T>(): [T, T][];`},
		{Code: `declare function fetchJson<T>(url: string): Promise<T>;`},
		{Code: `declare function fetchJsonTuple<T>(url: string): Promise<[T]>;`},
		{Code: `declare function fn<T>(input: T): 0 extends 0 ? T : never;`},
		{Code: `
declare namespace React {
  interface RefObject<T> {
    current: T | null;
  }
}
declare function useFocus<T extends HTMLOrSVGElement>(): [React.RefObject<T>];
`},
		{Code: `
declare namespace React {
  interface RefObject<T> {
    current: T | null;
  }
}
declare function useFocus<T extends HTMLOrSVGElement>(): {
  ref: React.RefObject<T>;
};
`},
		{Code: `
interface TwoMethods<T> {
  a(x: T): void;
  b(x: T): void;
}

declare function two<T>(props: TwoMethods<T>): void;
`},
		{Code: `
type Obj = { a: string };

declare function hasOwnProperty<K extends keyof Obj>(
  obj: Obj,
  key: K,
): obj is Obj & { [key in K]-?: Obj[key] };
`},
		{Code: `
type AsMutable<T extends readonly unknown[]> = {
  -readonly [Key in keyof T]: T[Key];
};

declare function makeMutable<T>(input: T): MakeMutable<T>;
`},
		{Code: `
type AsMutable<T extends readonly unknown[]> = {
  -readonly [Key in keyof T]: T[Key];
};

declare function makeMutable<T>(input: T): MakeMutable<typeof input>;
`},
		{Code: `
type ValueNulls<U extends string> = {} & {
  [P in U]: null;
};

declare function invert<T extends string>(obj: T): ValueNulls<T>;
`},
		{Code: `
interface Middle {
  inner: boolean;
}

type Conditional<T extends Middle> = {} & (T['inner'] extends true ? {} : {});

function withMiddle<T extends Middle = Middle>(options: T): Conditional<T> {
  return options;
}
`},
		{Code: `
import * as ts from 'typescript';

declare function forEachReturnStatement<T>(
  body: ts.Block,
  visitor: (stmt: ts.ReturnStatement) => T,
): T | undefined;
`},
		{Code: `
declare enum AST_NODE_TYPES {
  A = 'A',
  B = 'B',
}
declare namespace TSESTree {
  interface NodeA {
    type: AST_NODE_TYPES.A;
    a: string;
  }
  interface NodeB {
    type: AST_NODE_TYPES.B;
    b: number;
  }
  type Node = NodeA | NodeB;
}

declare const isNodeOfType: <NodeType extends AST_NODE_TYPES>(
  nodeType: NodeType,
) => node is Extract<TSESTree.Node, { type: NodeType }>;
`},
		{Code: `
declare enum AST_NODE_TYPES {
  A = 'A',
  B = 'B',
}
declare namespace TSESTree {
  interface NodeA {
    type: AST_NODE_TYPES.A;
    a: string;
  }
  interface NodeB {
    type: AST_NODE_TYPES.B;
    b: number;
  }
  type Node = NodeA | NodeB;
}

const isNodeOfType =
  <NodeType extends AST_NODE_TYPES>(nodeType: NodeType) =>
  (
    node: TSESTree.Node | null,
  ): node is Extract<TSESTree.Node, { type: NodeType }> =>
    node?.type === nodeType;
`},
		{Code: `
declare enum AST_TOKEN_TYPES {
  A = 'A',
  B = 'B',
}
declare namespace TSESTree {
  interface TokenA {
    type: AST_TOKEN_TYPES.A;
    a: string;
  }
  interface TokenB {
    type: AST_TOKEN_TYPES.B;
    b: number;
  }
  type Token = TokenA | TokenB;
}

export const isNotTokenOfTypeWithConditions =
  <
    TokenType extends AST_TOKEN_TYPES,
    ExtractedToken extends Extract<TSESTree.Token, { type: TokenType }>,
    Conditions extends Partial<ExtractedToken>,
  >(
    tokenType: TokenType,
    conditions: Conditions,
  ): ((
    token: TSESTree.Token | null | undefined,
  ) => token is Exclude<TSESTree.Token, Conditions & ExtractedToken>) =>
  (token): token is Exclude<TSESTree.Token, Conditions & ExtractedToken> =>
    tokenType in conditions;
`},
		{Code: `
type Foo<T, S> = S extends 'somebody'
  ? T extends 'once'
    ? 'told'
    : 'me'
  : never;

declare function foo<T>(data: T): <S>(other: S) => Foo<T, S>;
`},
		{Code: `
type Foo<T, S> = S extends 'somebody'
  ? T extends 'once'
    ? 'told'
    : 'me'
  : never;

declare function foo<T>(data: T): <S>(other: S) => Foo<S, T>;
`},
		{Code: `
declare function mapObj<K extends string, V>(
  obj: { [key in K]?: V },
  fn: (key: K, val: V) => number,
): number[];
`},
		{Code: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return mappedReturnType(x);
}
`},
		{Code: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return () => mappedReturnType(x);
}
`},
		{Code: `
declare function mappedReturnType<T extends string>(
  x: T,
): { [K in T]: Capitalize<K> };

function inferredMappedReturnType<T extends string>(x: T) {
  return [{ value: () => mappedReturnType(x) }];
}
`},
		{Code: `
type Identity<T> = T;

type Mapped<T, Value> = Identity<{ [P in keyof T]: Value }>;

declare function sillyFoo<Data, Value>(
  c: Value,
): (data: Data) => Mapped<Data, Value>;
`},
		{Code: `
type Silly<T> = { [P in keyof T]: T[P] };

type SillyFoo<T, Value> = Silly<{ [P in keyof T]: Value }>;

type Foo<T, Value> = { [P in keyof T]: Value };

declare function foo<T, Constant>(data: T, c: Constant): Foo<T, Constant>;
declare function foo<T, Constant>(c: Constant): (data: T) => Foo<T, Constant>;

declare function sillyFoo<T, Constant>(
  data: T,
  c: Constant,
): SillyFoo<T, Constant>;
declare function sillyFoo<T, Constant>(
  c: Constant,
): (data: T) => SillyFoo<T, Constant>;
`},
		{Code: `
const f = <T,>(setValue: (v: T) => void, getValue: () => NoInfer<T>) => {};
`},
		{Code: `
const f = <T,>(
  setValue: (v: T) => NoInfer<T>,
  getValue: (v: NoInfer<T>) => NoInfer<T>,
) => {};
`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `const func = <T,>(param: T) => null;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const func = (param: unknown) => null;`,
				}},
			}},
		},
		{
			Code: `const func = <T,>(param: [T]) => null;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const func = (param: [unknown]) => null;`,
				}},
			}},
		},
		{
			Code: `const func = <T,>(param: T[]) => null;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const func = (param: unknown[]) => null;`,
				}},
			}},
		},
		{
			Code: `const f1 = <T,>(): T => {};`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the function signature.",
				Line:      1,
				Column:    13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `const f1 = (): unknown => {};`,
				}},
			}},
		},
		{
			Code: `
interface I {
  <T>(value: T): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    4,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface I {
  (value: unknown): void;
}
`,
				}},
			}},
		},
		{
			Code: `
interface I {
  m<T>(x: T): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    5,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface I {
  m(x: unknown): void;
}
`,
				}},
			}},
		},
		{
			Code: `
class Joiner<T extends string | number> {
  join(el: T, other: string) {
    return [el, other].join(',');
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is used only once in the class signature.",
				Line:      2,
				Column:    14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Joiner {
  join(el: string | number, other: string) {
    return [el, other].join(',');
  }
}
`,
				}},
			}},
		},
		{
			Code: `
declare class C<V> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter V is never used in the class signature.",
				Line:      2,
				Column:    17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare class C {}
`,
				}},
			}},
		},
		{
			Code: `
declare class C<T, U> {
  method(param: T): U;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      2,
					Column:    17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
declare class C<U> {
  method(param: unknown): U;
}
`,
					}},
				},
				{
					MessageId: "sole",
					Line:      2,
					Column:    20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
declare class C<T> {
  method(param: T): unknown;
}
`,
					}},
				},
			},
		},
		{
			Code: `
declare class C {
  method<T, U>(param: T): U;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      3,
					Column:    10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
declare class C {
  method<U>(param: unknown): U;
}
`,
					}},
				},
				{
					MessageId: "sole",
					Line:      3,
					Column:    13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
declare class C {
  method<T>(param: T): unknown;
}
`,
					}},
				},
			},
		},
		{
			Code: `
declare class C {
  prop: <P>() => P;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    10,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare class C {
  prop: () => unknown;
}
`,
				}},
			}},
		},
		{
			Code: `
declare class Foo {
  foo<T>(this: T): void;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    7,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare class Foo {
  foo(this: unknown): void;
}
`,
				}},
			}},
		},
		{
			Code: `
function third<A, B, C>(a: A, b: B, c: C): C {
  return c;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      2,
					Column:    16,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
function third<B, C>(a: unknown, b: B, c: C): C {
  return c;
}
`,
					}},
				},
				{
					MessageId: "sole",
					Line:      2,
					Column:    19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
function third<A, C>(a: A, b: unknown, c: C): C {
  return c;
}
`,
					}},
				},
			},
		},
		{
			Code: `
function foo<T>(_: T) {
  const x: T = null!;
  const y: T = null!;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function foo(_: unknown) {
  const x: unknown = null!;
  const y: unknown = null!;
}
`,
				}},
			}},
		},
		{
			Code: `
function foo<T>(_: T): void {
  const x: T = null!;
  const y: T = null!;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function foo(_: unknown): void {
  const x: unknown = null!;
  const y: unknown = null!;
}
`,
				}},
			}},
		},
		{
			Code: `
function foo<T>(_: T): <T>(input: T) => T {
  const x: T = null!;
  const y: T = null!;
  return null!;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function foo(_: unknown): <T>(input: T) => T {
  const x: unknown = null!;
  const y: unknown = null!;
  return null!;
}
`,
				}},
			}},
		},
		{
			Code: `
function foo<T>(_: T) {
  function withX(): T {
    return null!;
  }
  function withY(): T {
    return null!;
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    14,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function foo(_: unknown) {
  function withX(): unknown {
    return null!;
  }
  function withY(): unknown {
    return null!;
  }
}
`,
				}},
			}},
		},
		{
			Code: `
function parseYAML<T>(input: string): T {
  return input as any as T;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function parseYAML(input: string): unknown {
  return input as any as unknown;
}
`,
				}},
			}},
		},
		{
			Code: `
function printProperty<T, K extends keyof T>(obj: T, key: K) {
  console.log(obj[key]);
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    27,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function printProperty<T>(obj: T, key: keyof T) {
  console.log(obj[key]);
}
`,
				}},
			}},
		},
		{
			Code: `
function fn<T>(param: string) {
  let v: T = null!;
  return v;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    13,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function fn(param: string) {
  let v: unknown = null!;
  return v;
}
`,
				}},
			}},
		},
		{
			Code: `
function both<
  Args extends unknown[],
  CB1 extends (...args: Args) => void,
  CB2 extends (...args: Args) => void,
>(fn1: CB1, fn2: CB2): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      4,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
function both<
  Args extends unknown[],
  CB2 extends (...args: Args) => void,
>(fn1: (...args: Args) => void, fn2: CB2): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
`,
					}},
				},
				{
					MessageId: "sole",
					Line:      5,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
function both<
  Args extends unknown[],
  CB1 extends (...args: Args) => void,
>(fn1: CB1, fn2: (...args: Args) => void): (...args: Args) => void {
  return function (...args: Args) {
    fn1(...args);
    fn2(...args);
  };
}
`,
					}},
				},
			},
		},
		{
			Code: `
function getLength<T extends { length: number }>(x: T) {
  return x.length;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function getLength(x: { length: number }) {
  return x.length;
}
`,
				}},
			}},
		},
		{
			Code: `
interface Lengthy {
  length: number;
}
function getLength<T extends Lengthy>(x: T) {
  return x.length;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      5,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface Lengthy {
  length: number;
}
function getLength(x: Lengthy) {
  return x.length;
}
`,
				}},
			}},
		},
		{
			Code: `declare function get<T>(): unknown;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is never used in the function signature.",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): unknown;`,
				}},
			}},
		},
		{
			Code: `declare function get<T>(): T;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): unknown;`,
				}},
			}},
		},
		{
			Code: `declare function get<T extends object>(): T;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get(): object;`,
				}},
			}},
		},
		{
			Code: `declare function take<T>(param: T): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function take(param: unknown): void;`,
				}},
			}},
		},
		{
			Code: `declare function take<T extends object>(param: T): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function take(param: object): void;`,
				}},
			}},
		},
		{
			Code: `declare function take<T, U = T>(param1: T, param2: U): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function take<T>(param1: T, param2: unknown): void;`,
				}},
			}},
		},
		{
			Code: `declare function take<T, U extends T>(param: T): U;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function take<T>(param: T): T;`,
				}},
			}},
		},
		{
			Code: `declare function take<T, U extends T>(param: U): U;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    23,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function take<U extends unknown>(param: U): U;`,
				}},
			}},
		},
		{
			Code: `declare function get<T, U = T>(param: U): U;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    22,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get<U = unknown>(param: U): U;`,
				}},
			}},
		},
		{
			Code: `declare function get<T, U extends T = T>(param: T): U;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    25,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function get<T>(param: T): T;`,
				}},
			}},
		},
		{
			Code: `declare function compare<T, U extends T>(param1: T, param2: U): boolean;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    29,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function compare<T>(param1: T, param2: T): boolean;`,
				}},
			}},
		},
		{
			Code: `declare function get<T>(param: <U, V>(param: U) => V): T;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      1,
					Column:    22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get(param: <U, V>(param: U) => V): unknown;`,
					}},
				},
				{
					MessageId: "sole",
					Line:      1,
					Column:    33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get<T>(param: <V>(param: unknown) => V): T;`,
					}},
				},
				{
					MessageId: "sole",
					Line:      1,
					Column:    36,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get<T>(param: <U>(param: U) => unknown): T;`,
					}},
				},
			},
		},
		{
			Code: `declare function get<T>(param: <T, U>(param: T) => U): T;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      1,
					Column:    22,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get(param: <T, U>(param: T) => U): unknown;`,
					}},
				},
				{
					MessageId: "sole",
					Line:      1,
					Column:    33,
					EndColumn: 34,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get<T>(param: <U>(param: unknown) => U): T;`,
					}},
				},
				{
					MessageId: "sole",
					Line:      1,
					Column:    36,
					EndColumn: 37,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output:    `declare function get<T>(param: <T>(param: T) => unknown): T;`,
					}},
				},
			},
		},
		{
			Code: `declare function makeReadonlyArray<T>(): readonly T[];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function makeReadonlyArray(): readonly unknown[];`,
				}},
			}},
		},
		{
			Code: `declare function makeReadonlyTuple<T>(): readonly [T];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    36,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function makeReadonlyTuple(): readonly [unknown];`,
				}},
			}},
		},
		{
			Code: `declare function makeReadonlyTupleNullish<T>(): readonly [T | null];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function makeReadonlyTupleNullish(): readonly [unknown | null];`,
				}},
			}},
		},
		{
			Code: `declare function takeArray<T>(input: T[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function takeArray(input: unknown[]): void;`,
				}},
			}},
		},
		{
			Code: `declare function takeArrayNullish<T>(input: (T | null)[]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    35,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function takeArrayNullish(input: (unknown | null)[]): void;`,
				}},
			}},
		},
		{
			Code: `declare function takeTuple<T>(input: [T]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    28,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function takeTuple(input: [unknown]): void;`,
				}},
			}},
		},
		{
			Code: `declare function takeTupleMultiUnrelated<T>(input: [T, number]): void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    42,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `declare function takeTupleMultiUnrelated(input: [unknown, number]): void;`,
				}},
			}},
		},
		{
			Code: `
declare function takeTupleMultiUnrelatedNullish<T>(
  input: [T | null, null],
): void;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    49,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare function takeTupleMultiUnrelatedNullish(
  input: [unknown | null, null],
): void;
`,
				}},
			}},
		},
		{
			Code: `type Fn = <T>() => T;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Fn = () => unknown;`,
				}},
			}},
		},
		{
			Code: `type Fn = <T>() => [];`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is never used in the function signature.",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Fn = () => [];`,
				}},
			}},
		},
		{
			Code: `
type Other = 0;
type Fn = <T>() => Other;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type Other = 0;
type Fn = () => Other;
`,
				}},
			}},
		},
		{
			Code: `
type Other = 0 | 1;
type Fn = <T>() => Other;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type Other = 0 | 1;
type Fn = () => Other;
`,
				}},
			}},
		},
		{
			Code: `type Fn = <U>(param: U) => void;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Fn = (param: unknown) => void;`,
				}},
			}},
		},
		{
			Code: `type Ctr = new <T>() => T;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    17,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Ctr = new () => unknown;`,
				}},
			}},
		},
		{
			Code: `type Fn = <T>() => { [K in keyof T]: K };`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Fn = () => { [K in keyof unknown]: K };`,
				}},
			}},
		},
		{
			Code: "type Fn = <T>() => { [K in 'a']: T };",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    "type Fn = () => { [K in 'a']: unknown };",
				}},
			}},
		},
		{
			Code: `type Fn = <T>(value: unknown) => value is T;`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    `type Fn = (value: unknown) => value is unknown;`,
				}},
			}},
		},
		{
			Code: "type Fn = <T extends string>() => `a${T}b`;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      1,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output:    "type Fn = () => `a${string}b`;",
				}},
			}},
		},
		{
			Code: `
declare function mapObj<K extends string, V>(
  obj: { [key in K]?: V },
  fn: (key: K) => number,
): number[];
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    43,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare function mapObj<K extends string>(
  obj: { [key in K]?: unknown },
  fn: (key: K) => number,
): number[];
`,
				}},
			}},
		},
		{
			Code: `
declare function setItem<T>(T): T;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    26,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
declare function setItem(T): unknown;
`,
				}},
			}},
		},
		{
			Code: `
interface StorageService {
  setItem<T>({ key: string, value: T }): Promise<void>;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Message:   "Type parameter T is never used in the function signature.",
				Line:      3,
				Column:    11,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
interface StorageService {
  setItem({ key: string, value: T }): Promise<void>;
}
`,
				}},
			}},
		},
		{
			// Not an important case in isolation, but used as a doc example of
			// code that gets flagged despite arguably not needing to be: see
			// the rule's "Differences from ESLint" note.
			Code: `
type Compute<A> = A extends Function ? A : { [K in keyof A]: Compute<A[K]> };
type Equal<X, Y> =
  (<T1>() => T1 extends Compute<X> ? 1 : 2) extends
    (<T2>() => T2 extends Compute<Y> ? 1 : 2)
  ? true
  : false;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "sole",
					Line:      4,
					Column:    5,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
type Compute<A> = A extends Function ? A : { [K in keyof A]: Compute<A[K]> };
type Equal<X, Y> =
  (() => unknown extends Compute<X> ? 1 : 2) extends
    (<T2>() => T2 extends Compute<Y> ? 1 : 2)
  ? true
  : false;
`,
					}},
				},
				{
					MessageId: "sole",
					Line:      5,
					Column:    7,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceUsagesWithConstraint",
						Output: `
type Compute<A> = A extends Function ? A : { [K in keyof A]: Compute<A[K]> };
type Equal<X, Y> =
  (<T1>() => T1 extends Compute<X> ? 1 : 2) extends
    (() => unknown extends Compute<Y> ? 1 : 2)
  ? true
  : false;
`,
					}},
				},
			},
		},
		{
			Code: `
function f<T extends any>(x: T): void {
  // @ts-expect-error
  x.notAMethod();
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    12,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function f(x: unknown): void {
  // @ts-expect-error
  x.notAMethod();
}
`,
				}},
			}},
		},
		{
			Code: `
class Joiner {
  join<T extends number>(els: T[]) {
    return els.map(el => '' + el).join(',');
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      3,
				Column:    8,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
class Joiner {
  join(els: number[]) {
    return els.map(el => '' + el).join(',');
  }
}
`,
				}},
			}},
		},
		{
			Code: `
function join<T extends string | number>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function join(els: (string | number)[]) {
  return els.map(el => '' + el).join(',');
}
`,
				}},
			}},
		},
		{
			Code: `
function join<T extends string & number>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function join(els: (string & number)[]) {
  return els.map(el => '' + el).join(',');
}
`,
				}},
			}},
		},
		{
			Code: `
function join<T extends (string & number) | boolean>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function join(els: ((string & number) | boolean)[]) {
  return els.map(el => '' + el).join(',');
}
`,
				}},
			}},
		},
		{
			Code: `
function join<T extends (string | number)>(els: T[]) {
  return els.map(el => '' + el).join(',');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function join(els: (string | number)[]) {
  return els.map(el => '' + el).join(',');
}
`,
				}},
			}},
		},
		{
			Code: `
function join<T extends { hoge: string } | { hoge: number }>(els: T['hoge'][]) {
  return els.map(el => '' + el).join(',');
}
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      2,
				Column:    15,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
function join(els: ({ hoge: string } | { hoge: number })['hoge'][]) {
  return els.map(el => '' + el).join(',');
}
`,
				}},
			}},
		},
		{
			Code: `
type A = string;
type B = string;
type C = string;
declare function f<T extends A | B>(): T & C;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      5,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type A = string;
type B = string;
type C = string;
declare function f(): (A | B) & C;
`,
				}},
			}},
		},
		{
			Code: `
type A = string;
type B = string;
type C = string;
type D = string;
declare function f<T extends (A extends B ? C : D)>(): T | null;
`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "sole",
				Line:      6,
				Column:    20,
				Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
					MessageId: "replaceUsagesWithConstraint",
					Output: `
type A = string;
type B = string;
type C = string;
type D = string;
declare function f(): (A extends B ? C : D) | null;
`,
				}},
			}},
		},
	})
}
