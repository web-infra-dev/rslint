// cspell:ignore Unioned aorb
// TestNoUnnecessaryTypeAssertionUpstream migrates the complete valid/invalid
// suite from typescript-eslint v8.68.0. rslint-specific cases live in
// no_unnecessary_type_assertion_extras_test.go.
package no_unnecessary_type_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTypeAssertionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTypeAssertionRule,
		[]rule_tester.ValidTestCase{
			{Code: `
type RecursiveMethod<Options = {}> = (<NewOptions = {}>(
  options: NewOptions,
) => RecursiveMethod<Options & NewOptions>) &
  ((command: string) => Promise<void>);

declare const recursiveMethod: RecursiveMethod;
declare const value: object;

value as unknown as typeof recursiveMethod;
    `},
			{Code: `
function castToSubtype<T extends string>(value: string): T {
  return value as T;
}
    `},
			{Code: `
import { TSESTree } from '@typescript-eslint/utils';
declare const member: TSESTree.TSEnumMember;
if (
  member.id.type === AST_NODE_TYPES.Literal &&
  typeof member.id.value === 'string'
) {
  const name = member.id as TSESTree.StringLiteral;
}
    `},
			{Code: `
const c = 1;
let z = c as number;
    `},
			{Code: `
const c = 1;
let z = c as const;
    `},
			{Code: `
const c = 1;
let z = c as 1;
    `},
			{Code: `
type Bar = 'bar';
const data = {
  x: 'foo' as 'foo',
  y: 'bar' as Bar,
};
    `},
			{Code: `[1, 2, 3, 4, 5].map(x => [x, 'A' + x] as [number, string]);`},
			{Code: `
let x: Array<[number, string]> = [1, 2, 3, 4, 5].map(
  x => [x, 'A' + x] as [number, string],
);
    `},
			{Code: `let y = 1 as 1;`},
			{Code: `const foo = 3 as number;`},
			{Code: `const foo = <number>3;`},
			{Code: `
type Tuple = [3, 'hi', 'bye'];
const foo = [3, 'hi', 'bye'] as Tuple;
    `},
			{Code: `
type PossibleTuple = {};
const foo = {} as PossibleTuple;
    `},
			{Code: `
type PossibleTuple = { hello: 'hello' };
const foo = { hello: 'hello' } as PossibleTuple;
    `},
			{Code: `
type PossibleTuple = { 0: 'hello'; 5: 'hello' };
const foo = { 0: 'hello', 5: 'hello' } as PossibleTuple;
    `},
			{Code: `
let bar: number | undefined = x;
let foo: number = bar!;
    `},
			{Code: `
declare const a: { data?: unknown };

const x = a.data!;
    `},
			{Code: `
declare function foo(arg?: number): number | void;
const bar: number = foo()!;
    `},
			{Code: `
type Foo = number;
const foo = (3 + 5) as Foo;
      `, Options: []any{map[string]any{`typesToIgnore`: []any{`Foo`}}}},
			{Code: `const foo = (3 + 5) as any;`, Options: []any{map[string]any{`typesToIgnore`: []any{`any`}}}},
			{Code: `(Syntax as any).ArrayExpression = 'foo';`, Options: []any{map[string]any{`typesToIgnore`: []any{`any`}}}},
			{Code: `const foo = (3 + 5) as string;`, Options: []any{map[string]any{`typesToIgnore`: []any{`string`}}}},
			{Code: `
type Foo = number;
const foo = <Foo>(3 + 5);
      `, Options: []any{map[string]any{`typesToIgnore`: []any{`Foo`}}}},
			{Code: `
let bar: number;
bar! + 1;
    `},
			{Code: `
let bar: undefined | number;
bar! + 1;
    `},
			{Code: `
let bar: number, baz: number;
bar! + 1;
    `},
			{Code: `
function foo<T extends string | undefined>(bar: T) {
  return bar!;
}
    `},
			{Code: `
type Data<T> = { value?: T };
type ValueType<TData> = TData extends Data<infer T> ? T : never;

export const foo = <TData extends Data<any>>(data: TData) => {
  const getValue = () => data.value as ValueType<TData> | undefined;
  const value: ValueType<TData> = getValue()!;
  return value;
};
    `},
			{Code: `
function bar<T extends any>(value: T | undefined): T {
  return value!;
}
    `},
			{Code: `
declare function nonNull(s: string);
let s: string | null = null;
nonNull(s!);
    `},
			{Code: `
const x: number | null = null;
const y: number = x!;
    `},
			{Code: `
const x: number | null = null;
class Foo {
  prop: number = x!;
}
    `},
			{Code: `
class T {
  a = 'a' as const;
}
    `},
			{Code: `
class T {
  a = 3 as 3;
}
    `},
			{Code: "\nconst foo = 'foo';\n\nclass T {\n  readonly test = `${foo}` as const;\n}\n    "},
			{Code: `
class T {
  readonly a = { foo: 'foo' } as const;
}
    `},
			{Code: `
declare const y: number | null;
console.log(y!);
    `},
			{Code: `
declare function foo(str?: string): void;
declare const str: string | null;

foo(str!);
    `},
			{Code: `
declare function a(a: string): any;
declare const b: string | null;
class Mx {
  @a(b!)
  private prop = 1;
}
    `},
			{Code: `
function testFunction(_param: string | undefined): void {
  /* noop */
}
const value = 'test' as string | null | undefined;
testFunction(value!);
    `},
			{Code: `
function testFunction(_param: string | null): void {
  /* noop */
}
const value = 'test' as string | null | undefined;
testFunction(value!);
    `},
			{Code: `
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: null | string | number }) {
  return <div key={props.id!} />;
}
      `, Tsx: true},
			{Code: `
const a = [1, 2];
const b = [3, 4];
const c = [...a, ...b] as const;
      `},
			{Code: `const a = [1, 2] as const;`},
			{Code: `const a = { foo: 'foo' } as const;`},
			{Code: `
const a = [1, 2];
const b = [3, 4];
const c = <const>[...a, ...b];
      `},
			{Code: `const a = <const>[1, 2];`},
			{Code: `const a = <const>{ foo: 'foo' };`},
			{Code: `
let a: number | undefined;
let b: number | undefined;
let c: number;
a = b;
c = b!;
a! -= 1;
      `},
			{Code: `
let a: { b?: string } | undefined;
a!.b = '';
      `},
			{Code: `
let value: number | undefined;
let values: number[] = [];

value = values.pop()!;
    `},
			{Code: `
declare function foo(): number | undefined;
const a = foo()!;
    `},
			{Code: `
declare function foo(): number | undefined;
const a = foo() as number;
    `},
			{Code: `
declare function foo(): number | undefined;
const a = <number>foo();
    `},
			{Code: `
declare const arr: (object | undefined)[];
const item = arr[0]!;
    `},
			{Code: `
declare const arr: (object | undefined)[];
const item = arr[0] as object;
    `},
			{Code: `
declare const arr: (object | undefined)[];
const item = <object>arr[0];
    `},
			{Code: `
const array: object[] = [{}];

let nullish: object | undefined;
nullish ??= array[1] as object | undefined;

let falsy: object | undefined;
falsy ||= array[1] as object | undefined;

let truthy: object | undefined = {};
truthy &&= array[1] as object | undefined;
    `},
			{Code: `
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]!);
  }
}
      `, TSConfig: `tsconfig.noUncheckedIndexedAccess.json`},
			{Code: "\ndeclare const myString: 'foo';\nconst templateLiteral = `${myString}-somethingElse` as const;\n    "},
			{Code: "\ndeclare const myString: 'foo';\nconst templateLiteral = `${myString}-somethingElse` as const;\n      ", Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: "\ndeclare const myString: 'foo';\nconst templateLiteral = <const>`${myString}-somethingElse`;\n    "},
			{Code: "\nconst myString = 'foo';\nconst templateLiteral = `${myString}-somethingElse` as const;\n    "},
			{Code: "\ntype ValuePath = 'values' | `values.${string}`;\n\ndeclare function apply(paths: ValuePath[]): void;\n\nexport function update(ids: string[]) {\n  apply(ids.map(id => `values.${id}` as ValuePath));\n}\n    "},
			{Code: "let a = `a` as const;"},
			{Code: `
declare const foo: {
  a?: string;
};
const bar = foo.a as string;
      `, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a as string;
      `, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
declare const foo: {
  a: string;
};
const bar = foo.a as string | undefined;
      `, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
declare const foo: {
  a?: string | null | number;
};
const bar = foo.a as string | undefined;
      `, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
declare const foo: {
  a?: string | number;
};
const bar = foo.a as string | undefined | bigint;
      `, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
if (Math.random()) {
  {
    var x = 1;
  }
}
x!;
      `},
			{Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T.Value2;
      `},
			{Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T;
      `},
			{Code: `
enum T {
  Value1 = 0,
  Value2 = 1,
}

const b = 1 as T.Value2;
      `},
			{Code: `
const foo: unknown = {};
const baz: {} = foo!;
    `},
			{Code: `
const foo: unknown = {};
const bar: object = foo!;
    `},
			{Code: `
declare function foo<T extends unknown>(bar: T): T;
const baz: unknown = {};
foo(baz!);
    `},
			{Code: "const a = `a` as const;"},
			{Code: `const a = 'a' as const;`},
			{Code: `const a = <const>'a';`},
			{Code: `
class T {
  readonly a = 'a' as const;
}
      `},
			{Code: `
enum T {
  Value1,
  Value2,
}
declare const a: T.Value1;
const b = a as const;
      `},
			{Code: `
(() => {})() as undefined;
      `},
			{Code: `
const f = () => {};
f() as undefined;
      `},
			{Code: `
(function () {})() as undefined;
      `},
			{Code: `
interface Overloaded {
  (): undefined;
  (value: string): void;
}

((value => {}) as Overloaded)('') as undefined;
      `},
			{Code: `
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((() => {}) as Overloaded)() as undefined;
      `},
			{Code: `
interface GenericOverloaded {
  <T extends string>(value: T): void;
  (): undefined;
}
((value => {}) as GenericOverloaded)('') as undefined;
      `},
			{Code: `
interface Unioned {
  (): undefined | void;
}

((() => {}) as Unioned)() as undefined;
      `},
			{Code: `
function fn<T>(items: ReadonlyArray<T>) {}
fn([42] as const);
      `},
			{Code: `
declare const a: any;
declare function foo(arg: string): void;
foo(a as string);
    `},
			{Code: `
declare const a: object;
const b = a as { id?: number };
    `},
			{Code: `
declare const array: any[];
function foo(strings: string[]): void {}
foo(array as string[]);
    `},
			{Code: `
declare const record: Record<string, unknown>;
const obj = record as { id?: number };
    `},
			{Code: `
declare const obj: { [key: string]: unknown };
const foo = obj as {};
    `},
			{Code: `
interface Empty {}
declare function getAny(): any;
const result = getAny() as Empty;
    `},
			{Code: `
interface Empty {}
declare function getObject(): object;
const result = getObject() as Empty;
    `},
			{Code: `
interface Obj {
  id: number;
}
declare const obj: Readonly<Obj>;
const obj2 = obj as Obj;
    `},
			{Code: `
declare const record: Record<string, unknown>;
const obj = record as { [additionalProperties: string]: unknown; id?: number };
    `},
			{Code: `
interface PropsA {
  a?: number;
}
interface PropsB extends PropsA {
  b?: string;
}
declare const propsB: PropsB;
const propsA = propsB as PropsA;
    `},
			{Code: `
interface PropsA {
  a?: number;
}
interface PropsB extends PropsA {
  b?: string;
}
declare const propsB: PropsB[];
const propsA = propsB as PropsA[];
    `},
			{Code: `
class Box<T> {
  value: T;
}
class PairBox<T, U> {
  value: T;
}
declare const pairBox: PairBox<string, number>;
const box = pairBox as Box<string>;
    `},
			{Code: `
type ObjectLike = Record<string, unknown>;
declare const result: ObjectLike;
declare const key: string;
result[key] = { ...(result[key] as ObjectLike) };
    `},
			{Code: `
interface AST {
  comments: string[] | undefined;
}
const ast: AST = {
  comments: [],
};
const { comments } = ast as { comments: string[] };
    `},
			{Code: `
type Tuple = [string | undefined, number];
const tuple: Tuple = ['hello', 42];
const [first, second] = tuple as [string, number];
    `},
			{Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
const obj = { value: narrow as Wide } satisfies Record<string, Wide>;
    `},
			{Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
const value = narrow as Wide satisfies Wide;
    `},
			{Code: `
interface Wide {
  name?: string;
}
interface Narrow {
  name: string;
}
declare const narrow: Narrow;
declare function identity<T>(x: T): T;
const result = identity({ value: narrow as Wide }) satisfies { value: Wide };
    `},
			{Code: `
declare const x: string | number;
const result: { tag: string; value: string | number } | { value: number } = {
  value: x as number,
};
    `},
			{Code: `
declare const x: string | number;
function fn(): { tag: string; value: string | number } | { value: number } {
  return {
    value: x as number,
  };
}
    `},
			{Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare const a: A;
let result;
result = a as B;
result.b;
    `},
			{Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
interface C extends B {
  c: string;
}
declare let a: A;
declare let b: B;
const c = (a = b as C);
c.c;
    `},
			{Code: `
type NumberRecord = { readonly [P in number]: number };
function fn<T extends NumberRecord>(record: T) {
  for (const key of Object.keys(record)) {
    const index = +key as keyof T & number;
    record[index] = record[index] + 1;
  }
}
    `},
			{Code: `
interface ReadonlyMap<K, V> {
  get(key: K): V | undefined;
}
type T = { get<K>(key: K): K };
declare const x: ReadonlyMap<string, string>;
declare let y: T;
y = x as T;
    `},
			{Code: `
declare function find<T>(array: readonly T[] | undefined): T | undefined;
declare const array: string[] | number[];
find(array as (string | number)[]);
    `},
			{Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function mapDefined<T>(fn: () => T): T;
declare const b: B;
declare const arrayA: A[];
const a = mapDefined(() => b as A);
[a].concat(arrayA);
    `},
			{Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function mapDefined<T>(fn: () => T): T;
declare const b: B;
declare const arrayA: A[];
const a = mapDefined(() =>
  Math.random() > 0.5 ? (b as A) : (null as unknown as A),
);
[a].concat(arrayA);
    `},
			{Code: `
interface Params {
  a?: string;
  b?: string;
}
declare const params: Omit<Params, 'a'> & { c?: string };
(params as Params).a = 'c';
    `},
			{Code: `
const text: string | null = null as string | null;
if (text) {
  text.toLowerCase();
}
    `},
			{Code: `
const text: string | undefined = undefined as string | undefined;
if (text) {
  text.toLowerCase();
}
    `},
			{Code: `
type Infer<T> = T extends ObjectConstructor
  ? never
  : T extends () => infer V
    ? V
    : never;
declare function fn<T>(o: { p: T }): { [K in keyof T]: Infer<T[K]> };
const result = fn({ p: { a: Object as () => string } });
result.a.toLowerCase();
    `},
			{Code: `
type Accessor<T> = () => T;
declare function inner<T>(): Accessor<T>;
function outer<T>(): Accessor<T> {
  return inner<string>() as Accessor<any>;
}
    `},
			{Code: `
interface InjectionConstraint<T> {}
type InjectionKey<T> = symbol & InjectionConstraint<T>;
declare function inject<T>(key: InjectionKey<T>): T;
const context = Symbol('ctx') as InjectionKey<{ value: string }>;
inject(context).value;
    `},
			{Code: `
declare function fn<U>(g: (memo: U) => U, initial: U): U;
declare function fn<T>(g: (memo: T) => T): T | undefined;
enum E {
  A = 1,
}
const x: E = fn(n => n | 0, 0 as E);
    `},
			{Code: `
type BasePayload = { id: string };

abstract class AbstractHandler {
  constructor(_ctx: { token: number }) {}
}

abstract class AbstractPayloadHandler<
  TPayload extends BasePayload = BasePayload,
> extends AbstractHandler {}

type HandlerCtor = new (ctx: { token: number }) => AbstractHandler;

declare const registeredHandlers: HandlerCtor[];

function example<TItem extends BasePayload>(
  handlerClass: typeof AbstractPayloadHandler<TItem>,
) {
  registeredHandlers.includes(
    handlerClass as new (...args: any[]) => AbstractPayloadHandler<TItem>,
  );
}
    `},
			{Code: `
function fn<T extends { type: string }, K extends string, V>(
  node: T,
): T & Record<K, V> {
  return node as T & Record<K, V>;
}
    `},
			{Code: `
declare function fn<T extends boolean>(
  options: {
    a: T extends true ? never : unknown;
  } & {
    b: T;
  },
): void;

fn({
  a: true,
  b: true as any,
});
    `},
			{Code: `
declare const a: number[] | number | undefined;
const b: number[] = a ?? ([0] as any);
    `},
			{Code: `
const context: { meta: Record<string, unknown> | undefined } = { meta: {} };
const meta = context.meta as { schema?: object } | undefined;
    `},
			{Code: `
type Test<T extends Record<string, unknown>> = {};

function inferred<T extends Test<never>[]>(_input: {
  addons?: T;
}): {
  options: T extends Test<infer C>[] ? C : never;
} {
  return {
    options: {} as T extends Test<infer C>[] ? C : never,
  };
}

const test = inferred({
  addons: [{} as Test<{ parameters: { potato: boolean } }>],
});

console.log(test.options.parameters.potato);
    `},
			{Code: `
declare const items: string[] | undefined;

const counts = items?.reduce(
  (acc, item) => {
    acc[item] = (acc[item] ?? 0) + 1;
    return acc;
  },
  {} as Record<string, number>,
);
    `},
			{Code: `
declare const o:
  | {
      fn<U>(g: (memo: U) => U, initial: U): U;
      fn<T>(g: (memo: T) => T): T | undefined;
    }
  | undefined;
enum E {
  A = 1,
}
const x: E | undefined = o?.fn(n => n | 0, 0 as E);
    `},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `const foo = <3>3;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 13}}, Output: []string{`const foo = 3;`}},
			{Code: `const foo = 3 as 3;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 13}}, Output: []string{`const foo = 3;`}},
			{Code: `
type Foo = 3;
const foo = <Foo>3;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 13}}, Output: []string{`
type Foo = 3;
const foo = 3;
      `}},
			{Code: `
type Foo = 3;
const foo = 3 as Foo;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 13}}, Output: []string{`
type Foo = 3;
const foo = 3;
      `}},
			{Code: `
const foo = 3;
const bar = foo!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 13}}, Output: []string{`
const foo = 3;
const bar = foo;
      `}},
			{Code: `
const foo = (3 + 5) as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (3 + 5);
      `}},
			{Code: `
const foo = <number>(3 + 5);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (3 + 5);
      `}},
			{Code: `
type Foo = number;
const foo = (3 + 5) as Foo;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 13}}, Output: []string{`
type Foo = number;
const foo = (3 + 5);
      `}},
			{Code: `
type Foo = number;
const foo = <Foo>(3 + 5);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 13}}, Output: []string{`
type Foo = number;
const foo = (3 + 5);
      `}},
			{Code: `
let bar: number = 1;
bar! + 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
let bar: number = 1;
bar + 1;
      `}},
			{Code: `
let bar!: number;
bar! + 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
let bar!: number;
bar + 1;
      `}},
			{Code: `
let bar: number | undefined;
bar = 1;
bar! + 1;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 4}}, Output: []string{`
let bar: number | undefined;
bar = 1;
bar + 1;
      `}},
			{Code: `
declare const y: number;
console.log(y!);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
declare const y: number;
console.log(y);
      `}},
			{Code: `Proxy!;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`Proxy;`}},
			{Code: `
function foo<T extends string>(bar: T) {
  return bar!;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
function foo<T extends string>(bar: T) {
  return bar;
}
      `}},
			{Code: `
declare const foo: Foo;
const bar = <Foo>foo;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
declare const foo: Foo;
const bar = foo;
      `}},
			{Code: `
declare function nonNull(s: string | null);
let s: string | null = null;
nonNull(s!);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 4}}, Output: []string{`
declare function nonNull(s: string | null);
let s: string | null = null;
nonNull(s);
      `}},
			{Code: `
const x: number | null = null;
const y: number | null = x!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 3}}, Output: []string{`
const x: number | null = null;
const y: number | null = x;
      `}},
			{Code: `
const x: number | null = null;
class Foo {
  prop: number | null = x!;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 4}}, Output: []string{`
const x: number | null = null;
class Foo {
  prop: number | null = x;
}
      `}},
			{Code: `
declare function a(a: string): any;
const b = 'asdf';
class Mx {
  @a(b!)
  private prop = 1;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 5}}, Output: []string{`
declare function a(a: string): any;
const b = 'asdf';
class Mx {
  @a(b)
  private prop = 1;
}
      `}},
			{Code: `
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: string | number }) {
  return <div key={props.id!} />;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 9}}, Output: []string{`
declare namespace JSX {
  interface IntrinsicElements {
    div: { key?: string | number };
  }
}

function Test(props: { id?: string | number }) {
  return <div key={props.id} />;
}
      `}, Tsx: true},
			{Code: `
let x: number | undefined;
let y: number | undefined;
y = x!;
y! = 0;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 5}}, Output: []string{`
let x: number | undefined;
let y: number | undefined;
y = x!;
y = 0;
      `}},
			{Code: `
declare function foo(arg?: number): number | void;
const bar: number | void = foo()!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 3, Column: 28, EndColumn: 34}}, Output: []string{`
declare function foo(arg?: number): number | void;
const bar: number | void = foo();
      `}},
			{Code: `
declare function foo(): number;
const a = foo()!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 11, EndColumn: 17}}, Output: []string{`
declare function foo(): number;
const a = foo();
      `}},
			{Code: `
const b = new Date()!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2}}, Output: []string{`
const b = new Date();
      `}},
			{Code: `
const b = (1 + 1)!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 11, EndColumn: 19}}, Output: []string{`
const b = (1 + 1);
      `}},
			{Code: `
declare function foo(): number;
const a = foo() as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3, Column: 11}}, Output: []string{`
declare function foo(): number;
const a = foo();
      `}},
			{Code: `
declare function foo(): number;
const a = <number>foo();
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
declare function foo(): number;
const a = foo();
      `}},
			{Code: `
type RT = { log: () => void };
declare function foo(): RT;
(foo() as RT).log;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
type RT = { log: () => void };
declare function foo(): RT;
(foo()).log;
      `}},
			{Code: `
declare const arr: object[];
const item = arr[0]!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
declare const arr: object[];
const item = arr[0];
      `}},
			{Code: `
const foo = (  3 + 5  ) as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (  3 + 5  );
      `}},
			{Code: `
const foo = (  3 + 5  ) /*as*/ as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (  3 + 5  ) /*as*/;
      `}},
			{Code: `
const foo = (  3 + 5
  ) /*as*/ as //as
  (
    number
  );
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (  3 + 5
  ) /*as*/;
      `}},
			{Code: `
const foo = (3 + (5 as number) ) as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (3 + (5 as number) );
      `}},
			{Code: `
const foo = 3 + 5/*as*/ as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = 3 + 5/*as*/;
      `}},
			{Code: `
const foo = 3 + 5/*a*/ /*b*/ as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = 3 + 5/*a*/ /*b*/;
      `}},
			{Code: `
const foo = <(number)>(3 + 5);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (3 + 5);
      `}},
			{Code: `
const foo = < ( number ) >( 3 + 5 );
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = ( 3 + 5 );
      `}},
			{Code: `
const foo = <number> /* a */ (3 + 5);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo =  /* a */ (3 + 5);
      `}},
			{Code: `
const foo = <number /* a */>(3 + 5);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2, Column: 13}}, Output: []string{`
const foo = (3 + 5);
      `}},
			{Code: `
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]!);
  }
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 5, Column: 9}}, Output: []string{`
function foo(item: string) {}
function bar(items: string[]) {
  for (let i = 0; i < items.length; i++) {
    foo(items[i]);
  }
}
      `}},
			{Code: `
declare const foo: {
  a?: string;
};
const bar = foo.a as string | undefined;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 5, Column: 13}}, Output: []string{`
declare const foo: {
  a?: string;
};
const bar = foo.a;
      `}, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a as string | undefined;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 5, Column: 13}}, Output: []string{`
declare const foo: {
  a?: string | undefined;
};
const bar = foo.a;
      `}, TSConfig: `tsconfig.exactOptionalPropertyTypes.json`},
			{Code: `
varDeclarationFromFixture!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 2}}, Output: []string{`
varDeclarationFromFixture;
      `}},
			{Code: `
var x = 1;
x!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
var x = 1;
x;
      `}},
			{Code: `
var x = 1;
{
  x!;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 4}}, Output: []string{`
var x = 1;
{
  x;
}
      `}},
			{Code: `
class T {
  readonly a = 3 as 3;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
class T {
  readonly a = 3;
}
      `}},
			{Code: `
type S = 10;

class T {
  readonly a = 10 as S;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 5}}, Output: []string{`
type S = 10;

class T {
  readonly a = 10;
}
      `}},
			{Code: `
class T {
  readonly a = (3 + 5) as number;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
class T {
  readonly a = (3 + 5);
}
      `}},
			{Code: `
const a = '';
const b: string | undefined = (a ? undefined : a)!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
const a = '';
const b: string | undefined = (a ? undefined : a);
      `}},
			{Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as T.Value1;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a;
      `}},
			{Code: `
const foo: unknown = {};
const bar: unknown = foo!;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
const foo: unknown = {};
const bar: unknown = foo;
      `}},
			{Code: `
function foo(bar: unknown) {}
const baz: unknown = {};
foo(baz!);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
function foo(bar: unknown) {}
const baz: unknown = {};
foo(baz);
      `}},
			{Code: `const a = true as const;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = true;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = <const>true;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = true;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = 1 as const;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 1;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = <const>1;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 1;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = 1n as const;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 1n;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = <const>1n;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 1n;`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: "const a = `a` as const;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{"const a = `a`;"}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = 'a' as const;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 'a';`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `const a = <const>'a';`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1}}, Output: []string{`const a = 'a';`}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `
class T {
  readonly a = 'a' as const;
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 3}}, Output: []string{`
class T {
  readonly a = 'a';
}
      `}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a as const;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
enum T {
  Value1,
  Value2,
}

declare const a: T.Value1;
const b = a;
      `}, Options: []any{map[string]any{`checkLiteralConstAssertions`: true}}},
			{Code: `
((): undefined => {})() as undefined;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
((): undefined => {})();
      `}},
			{Code: `
(() => 1)() as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
(() => 1)();
      `}},
			{Code: `
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((value => {}) as Overloaded)('') as undefined;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
interface Overloaded {
  (): void;
  (value: string): undefined;
}

((value => {}) as Overloaded)('');
      `}},
			{Code: `
function doThing(a: number) {}
doThing(5 as any);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
function doThing(a: number) {}
doThing(5);
      `}},
			{Code: `
interface A {
  required: string;
  alsoRequired: number;
}
function doThing(a: A) {}
doThing({ required: 'yes', alsoRequired: 1 } as any);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
interface A {
  required: string;
  alsoRequired: number;
}
function doThing(a: A) {}
doThing({ required: 'yes', alsoRequired: 1 });
      `}},
			{Code: `const x = 5 as any as 5;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`const x = 5;`}},
			{Code: `
const v: number = 5;
const x = v as unknown as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const v: number = 5;
const x = v;
      `}},
			{Code: `
const v: number = 5;
const x = v as any as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const v: number = 5;
const x = v;
      `}},
			{Code: `
const x = (1 + 1) as any as number;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const x = 1 + 1;
      `}},
			{Code: `
const x = 2 * ((1 + 1) as any as number);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const x = 2 * (1 + 1);
      `}},
			{Code: `
const v: number = 5;
const x = <number>(<any>v);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const v: number = 5;
const x = v;
      `}},
			{Code: `
const obj = { id: '' };
const obj2 = obj as { id: string };
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `}},
			{Code: `
const obj = { id: '' };
const obj2 = obj as any as { id: string };
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `}},
			{Code: `
const obj = { id: '' };
const obj2 = obj as unknown as { id: string };
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const obj = { id: '' };
const obj2 = obj;
      `}},
			{Code: `
const array = ['a', 'b'];
const array2 = array as any as string[];
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const array = ['a', 'b'];
const array2 = array;
      `}},
			{Code: `
const array = ['a', 'b'];
const array2 = array as unknown as string[];
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
const array = ['a', 'b'];
const array2 = array;
      `}},
			{Code: `
type A = 'a';
type B = 'b';
type AorB = A | B;
function fn(aorb: AorB) {}
const a: A = 'a';
fn(a as AorB);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type A = 'a';
type B = 'b';
type AorB = A | B;
function fn(aorb: AorB) {}
const a: A = 'a';
fn(a);
      `}},
			{Code: `
interface Props {
  a: number;
}
const x = { a: 1 } as unknown as Props;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
interface Props {
  a: number;
}
const x = { a: 1 };
      `}},
			{Code: `
interface Props {
  a: number;
}
const x = { a: 1 } as Props;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
interface Props {
  a: number;
}
const x = { a: 1 };
      `}},
			{Code: `
interface Props {
  a: number;
}
const fn = (): Props => ({ a: 1 }) as unknown as Props;
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
interface Props {
  a: number;
}
const fn = (): Props => ({ a: 1 });
      `}},
			{Code: `
declare function fn(param: number): void;
fn(42 as unknown as number);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(param: number): void;
fn(42);
      `}},
			{Code: `
declare function fn(param: number): void;
fn(42 as any as number);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(param: number): void;
fn(42);
      `}},
			{Code: `
declare function fn(params: { param: number });
fn({ param: 42 as number });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(params: { param: number });
fn({ param: 42 });
      `}},
			{Code: `
declare function fn(params: { param: number });
fn({ param: 42 as any });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(params: { param: number });
fn({ param: 42 });
      `}},
			{Code: `
type StringOrNumber = string | number;
declare function fn(param: StringOrNumber);
fn(42 as any as StringOrNumber);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type StringOrNumber = string | number;
declare function fn(param: StringOrNumber);
fn(42);
      `}},
			{Code: `
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
const data = { a: 1 };
fn({ data: data as NumbersRecord });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
const data = { a: 1 };
fn({ data: data });
      `}},
			{Code: `
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
fn({
  data: {
    a: 1,
  } as NumbersRecord,
});
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type NumbersRecord = { [key: string]: number };
declare function fn(params: { data: NumbersRecord });
fn({
  data: {
    a: 1,
  },
});
      `}},
			{Code: `
type Json = string | number | boolean | null | { [key: string]: Json } | Json[];
type Tables<T extends 'my_table'> = { my_table: { my_column: Json } }[T];
declare const updatedColumn: Json;
const result = updatedColumn as unknown as Tables<'my_table'>['my_column'];
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
type Json = string | number | boolean | null | { [key: string]: Json } | Json[];
type Tables<T extends 'my_table'> = { my_table: { my_column: Json } }[T];
declare const updatedColumn: Json;
const result = updatedColumn;
      `}},
			{Code: `
interface T {
  a: string;
}
declare function fn<U extends T>(args: Pick<U, 'a'>): void;
fn<T>({ a: '' as string });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
interface T {
  a: string;
}
declare function fn<U extends T>(args: Pick<U, 'a'>): void;
fn<T>({ a: '' });
      `}},
			{Code: `
declare function update<T extends string>(value: T): void;
update('hi' as unknown as string);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function update<T extends string>(value: T): void;
update('hi');
      `}},
			{Code: `
declare function update<T extends string>(value: T): void;
update('hi' as string);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function update<T extends string>(value: T): void;
update('hi');
      `}},
			{Code: "\ndeclare function fn(param: string): void;\ndeclare const name_: string;\nfn(`hello ${name_}` as string);\n      ", Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{"\ndeclare function fn(param: string): void;\ndeclare const name_: string;\nfn(`hello ${name_}`);\n      "}},
			{Code: `
declare function fn(x: string[]): void;
fn(['hello'] as any);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(x: string[]): void;
fn(['hello']);
      `}},
			{Code: `
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update(values: { chat: Json[] }): void;
declare const chat: ChatMessage[];
update({ chat: chat as Json[] });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update(values: { chat: Json[] }): void;
declare const chat: ChatMessage[];
update({ chat: chat });
      `}},
			{Code: `
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update<Row extends { chat: Json[] }>(values: Row): void;
declare const chat: ChatMessage[];
update({ chat: chat as Json[] });
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
type ChatMessage = { message: string };
type Json = string | { [key: string]: Json };
declare function update<Row extends { chat: Json[] }>(values: Row): void;
declare const chat: ChatMessage[];
update({ chat: chat });
      `}},
			{Code: `
interface Node {
  parent: Node;
}
declare function fn<T extends Node>(node: T): T;
function fn2<T extends Node>(node: T): void {
  fn(node as NonNullable<T>);
}
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`
interface Node {
  parent: Node;
}
declare function fn<T extends Node>(node: T): T;
function fn2<T extends Node>(node: T): void {
  fn(node);
}
      `}},
			{Code: `
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function fn(a: A): A;
declare function fn(a: A): A | undefined;
declare const a: A;
fn(a as B);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
interface A {
  a: string;
}
interface B extends A {
  b: string;
}
declare function fn(a: A): A;
declare function fn(a: A): A | undefined;
declare const a: A;
fn(a);
      `}},
			{Code: `
declare const a: string[];
declare const b: readonly string[];
const fileNames: string[] = a.concat(b as string[]);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare const a: string[];
declare const b: readonly string[];
const fileNames: string[] = a.concat(b);
      `}},
			{Code: `
declare function fn(text: any): void;
declare const value: string | number;
fn(value as number);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn(text: any): void;
declare const value: string | number;
fn(value);
      `}},
			{Code: `
interface A {
  type: 'a';
  a: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a as string,
};
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
interface A {
  type: 'a';
  a: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a,
};
      `}},
			{Code: `
interface A {
  type: 'a';
  a?: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a as string,
};
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
interface A {
  type: 'a';
  a?: string;
}
interface B {
  type: 'b';
}
declare const a: '1' | '2';
const schema: A | B = {
  type: 'a',
  a: a,
};
      `}},
			{Code: `
declare function fn1<T>(fn: () => void): void;
declare function fn2(text: string): void;
fn1(() => {
  fn2('hi' as any);
});
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`
declare function fn1<T>(fn: () => void): void;
declare function fn2(text: string): void;
fn1(() => {
  fn2('hi');
});
      `}},
			{Code: `[].map(() => <{ a: false; b: false }>{ a: false, b: false });`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`[].map(() => ({ a: false, b: false }));`}},
			{Code: `[].map(() => /* 1 */ <{ a: false; b: false }> /* 2 */ { a: false, b: false } /* 3 */);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`[].map(() => /* 1 */ ( /* 2 */ { a: false, b: false }) /* 3 */);`}},
			{Code: `[].map(() => <{ a: false; b: false }>({ a: false, b: false }));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`[].map(() => ({ a: false, b: false }));`}},
			{Code: `[].map(() => (<{ a: false; b: false }>{ a: false, b: false }));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`}}, Output: []string{`[].map(() => ({ a: false, b: false }));`}},
			{Code: `<{ a: string }>{ a: 'foo' };`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`}}, Output: []string{`({ a: 'foo' });`}},
			{Code: `<{ a: string }>{ a: 'foo' } + 1;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}, Output: []string{`({ a: 'foo' }) + 1;`}},
			{Code: `<{ a: string }>{ a: 'foo' } && true;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}, Output: []string{`({ a: 'foo' }) && true;`}},
			{Code: `<{ a: string }>{ a: 'foo' } ? 1 : 2;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}, Output: []string{`({ a: 'foo' }) ? 1 : 2;`}},
			{Code: `<{ a: string }>{ a: 'foo' }, foo();`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}, Output: []string{`({ a: 'foo' }), foo();`}},
			{Code: `<{ a: string }>{ a: 'foo' } + 1 + 2;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 28}}, Output: []string{`({ a: 'foo' }) + 1 + 2;`}},
			{Code: `(<{ a: string }>{ a: 'foo' }, 1);`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 2, EndLine: 1, EndColumn: 29}}, Output: []string{`({ a: 'foo' }, 1);`}},
			{Code: `1 + <{ a: string }>{ a: 'foo' };`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 5, EndLine: 1, EndColumn: 32}}, Output: []string{`1 + { a: 'foo' };`}},
			{Code: `<number>{ lol: 32 as number }.lol;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 34}}, Output: []string{`({ lol: 32 as number }.lol);`}},
			{Code: `<number>function Fun() {}.length;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 33}}, Output: []string{`(function Fun() {}.length);`}},
			{Code: `<number>class Clazz {}.length;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 1, EndLine: 1, EndColumn: 30}}, Output: []string{`(class Clazz {}.length);`}},
			{Code: `const foo = () => <number>{ lol: 123 as number }.lol + 54321;`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `unnecessaryAssertion`, Line: 1, Column: 19, EndLine: 1, EndColumn: 53}}, Output: []string{`const foo = () => ({ lol: 123 as number }.lol) + 54321;`}},
			{Code: `
declare const maybeFn: ((arg: string | number) => void) | undefined;
declare const s: string;
maybeFn?.(s as string | number);
      `, Errors: []rule_tester.InvalidTestCaseError{{MessageId: `contextuallyUnnecessary`, Line: 4, Column: 11}}, Output: []string{`
declare const maybeFn: ((arg: string | number) => void) | undefined;
declare const s: string;
maybeFn?.(s);
      `}},
		},
	)
}
