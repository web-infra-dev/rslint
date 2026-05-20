// TestPreferFunctionTypeExtras locks in branches and edge shapes that the
// upstream test suite does not exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors cannot silently regress them without breaking a
// named lock-in.
package prefer_function_type

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferFunctionTypeExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferFunctionTypeRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: empty interface — `Members.Nodes` is empty, the
		//      length-1 gate keeps the rule silent.
		{Code: `interface Foo {}`},

		// ---- Dimension 4: empty type literal — same gate, no diagnostic.
		{Code: `type Foo = {};`},

		// ---- Dimension 4: two call signatures (overloads) — only the
		//      single-member shape is matched, so overloads are intentionally
		//      ignored.
		{Code: `
interface Overloaded {
  (data: string): number;
  (id: number): string;
}
    `},

		// ---- Dimension 4: a method signature is not a call signature — the
		//      switch on member.Kind only matches CallSignature /
		//      ConstructSignature, so a bare method stays valid.
		{Code: `
interface Foo {
  method(): void;
}
    `},

		// ---- Dimension 4: index signature — not a CallSignatureDeclaration,
		//      so the rule must not report.
		{Code: `
interface Foo {
  [key: string]: number;
}
    `},

		// ---- Dimension 4: property signature — same reason.
		{Code: `
interface Foo {
  bar: () => void;
}
    `},

		// ---- Locks in hasOneSupertype() arm: extends.length !== 1 (two
		//      supertypes, neither has to be Function) → skip the interface.
		{Code: `
interface Base1 {
  base1: string;
}
interface Base2 {
  base2: string;
}
interface Foo extends Base1, Base2 {
  (): void;
}
    `},

		// ---- Locks in hasOneSupertype() arm: extends expression is not an
		//      Identifier (member access like `ns.Base`) → skip.
		{Code: `
namespace ns {
  export interface Base {}
}
interface Foo extends ns.Base {
  (): void;
}
    `},

		// ---- Locks in hasOneSupertype() arm: extends one named type other
		//      than Function → skip. Distinct from the case where the only
		//      supertype is literally `Function` (upstream invalid-9).
		{Code: `
interface Base {
  base: string;
}
interface Foo extends Base {
  (): void;
}
    `},

		// ---- Dimension 4: nested type literal containing only a call
		//      signature is reported on the inner literal — but if the inner
		//      literal also has a sibling property, the inner is not
		//      single-member. Confirms the rule traverses each TypeLiteral
		//      independently.
		{Code: `
type Wrapper = {
  inner: { (): void; extra: number };
};
    `},

		// ---- Real-user: a generic event-callback interface with a sibling
		//      property is a common shape from React/Node typings and must
		//      stay valid (sibling property defeats single-member gate).
		{Code: `
interface EventEmitter<T> {
  (payload: T): void;
  cancel(): void;
}
    `},

		// ---- Dimension 4: function-type property — not a CallSignature, just
		//      a PropertySignature whose type happens to be `() => void`. Must
		//      not trigger the rule.
		{Code: `
interface Foo {
  handler: () => void;
}
    `},

		// ---- Dimension 4: type alias that's already a function type — there
		//      is no Interface/TypeLiteral wrapper to rewrite, must stay
		//      silent. Locks in that the rule doesn't accidentally fire on the
		//      result of its own fix (idempotency).
		{Code: `type Already = () => void;`},

		// ---- Dimension 4: a single index signature — `[k: string]: number`
		//      is the only member, but it's an IndexSignature (not a Call /
		//      Construct signature) so the switch falls through to default.
		{Code: `type Dict = { [k: string]: number };`},

		// ---- Dimension 4: single ConstructSignature without explicit return
		//      type — `member.returnType == null` keeps the rule silent (it
		//      requires a non-nil return type to know what to put after `=>`).
		//      `new ()` without explicit return reaches the `returnType == nil`
		//      guard.
		{Code: `type Ctor = { new (); };`},

		// ---- Dimension 4: interface with extends Function but no body
		//      members at all — Members.length === 0 skips the rule.
		{Code: `interface Empty extends Function {}`},

		// ---- Real-user: interface with multiple call signatures (overloads)
		//      — TS-typical pattern from lib.dom.d.ts (e.g. addEventListener
		//      overloads). Single-member gate keeps this valid.
		{Code: `
interface Overloaded {
  (data: string): number;
  (id: number, opts?: { strict: boolean }): string;
}
    `},

		// ---- Real-user: interface 'extends Bar<T>' where Bar is generic —
		//      typeArguments are ignored, expression is `Bar` (not Function),
		//      so the rule skips on the "single supertype that's not Function"
		//      branch.
		{Code: `
interface Base<T> {
  base: T;
}
interface Foo extends Base<number> {
  (): void;
}
    `},

		// ---- Real-user: callable interface used as a class member type —
		//      class field with a single-member type literal annotation
		//      stays valid here because the rewrite target is the parameter
		//      shape, but for class fields the listener is independent.
		//      The Foo type below has 2 members so single-member gate keeps it valid.
		{Code: `
type Service = {
  start(): void;
  stop(): void;
};
    `},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in checkMember() arm: ConstructSignature inside an
		//      interface. Upstream tests only exercise call signatures; this
		//      pins the `new` form.
		{
			Code: `
interface Foo {
  new (): Foo;
}
      `,
			Output: []string{`
type Foo = new () => Foo;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in checkMember() arm: ConstructSignature inside a
		//      TypeLiteral.
		{
			Code: `
type Foo = { new (): Foo };
      `,
			Output: []string{`
type Foo = new () => Foo;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    14,
				},
			},
		},

		// ---- Locks in shouldWrapSuggestion() arm: parent.type === ArrayType
		//      → wrap with parens. Upstream tests cover Union (invalid-16) and
		//      Intersection (invalid-17) but not Array.
		{
			Code: `
type X = { (): void }[];
      `,
			Output: []string{`
type X = (() => void)[];
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    12,
				},
			},
		},

		// ---- Locks in interface fix path: type parameters with multiple
		//      params + extends clause + complex return type.
		{
			Code: `
interface Foo<T, U> {
  (a: T): U;
}
      `,
			Output: []string{`
type Foo<T, U> = (a: T) => U;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in interface fix path: type parameter with a constraint
		//      (`extends`) — the slice from name to typeParameters.range[1]
		//      must include `<T extends X>` verbatim.
		{
			Code: `
interface Foo<T extends string> {
  (a: T): T;
}
      `,
			Output: []string{`
type Foo<T extends string> = (a: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in interface fix path: type parameter with a default
		//      (`T = string`) — same slicing requirement.
		{
			Code: `
interface Foo<T = string> {
  (a: T): T;
}
      `,
			Output: []string{`
type Foo<T = string> = (a: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in interface modifier prefix: `declare interface ...`
		//      keeps the `declare` keyword in the rewritten alias. Upstream
		//      does not test this branch.
		{
			Code: `
declare interface Foo {
  (): string;
}
      `,
			Output: []string{`
declare type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in interface modifier prefix: `export declare interface`
		//      preserves both keywords in the rewritten alias.
		{
			Code: `
export declare interface Foo {
  (): string;
}
      `,
			Output: []string{`
export declare type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in collectThisTypes() literalNesting branch: a `this`
		//      reference inside a nested TypeLiteral inside the call sig's
		//      return type must NOT trigger unexpectedThisOnFunctionOnlyInterface.
		//      Upstream invalid-15 covers this for object-typed returns; here we
		//      check it for a return inside an array of object types.
		{
			Code: `
interface Foo {
  (): { nested: this }[];
}
      `,
			Output: []string{`
type Foo = () => { nested: this }[];
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in collectThisTypes() arm: `this` in a parameter
		//      type (not the parameter name) — must report unexpectedThis.
		//      Distinct from `(this: T): void` where `this` is the parameter
		//      identifier and never produces a ThisType node.
		{
			Code: `
interface Foo {
  (other: this, more: number): void;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedThisOnFunctionOnlyInterface",
					Line:      3,
					Column:    11,
				},
			},
		},

		// ---- Real-user: callback that returns its own type via a parameterized
		//      type alias. Validates type-parameter forwarding survives the
		//      rewrite (closely mirrors common React/RxJS callback shapes).
		{
			Code: `
interface Mapper<T, U> {
  (input: T, index: number): U;
}
      `,
			Output: []string{`
type Mapper<T, U> = (input: T, index: number) => U;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: a type literal annotation on a variable declaration
		//      — `const cb: { (x: number): void } = ...` is a common manual
		//      type assertion shape from issue trackers.
		{
			Code: `
const cb: { (x: number): void } = x => {};
      `,
			Output: []string{`
const cb: (x: number) => void = x => {};
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    13,
				},
			},
		},

		// ---- Dimension 4: rest parameter — `(...args: T[]): R` shape must
		//      survive the rewrite. Locks in that the textual slicing from
		//      member start to colon preserves the spread token.
		{
			Code: `
interface Foo<T> {
  (...args: T[]): void;
}
      `,
			Output: []string{`
type Foo<T> = (...args: T[]) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: optional parameter — `(x?: T): R` must survive
		//      the rewrite with the `?` preserved.
		{
			Code: `
interface Foo<T> {
  (x?: T): void;
}
      `,
			Output: []string{`
type Foo<T> = (x?: T) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: type literal nested as the return type of a
		//      function type — `() => { (): void }` is itself a single-member
		//      type literal that the TypeLiteral listener must catch. The
		//      inner rewrite needs no parens because TypeLiteral's parent is
		//      FunctionType (not on shouldWrapSuggestion's list); `() => () => void`
		//      is right-associative and unambiguous.
		{
			Code: `
type Outer = () => { (): void };
      `,
			Output: []string{`
type Outer = () => () => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    22,
				},
			},
		},

		// ---- Dimension 4: type literal as type argument inside `Promise<...>`
		//      — must report. Locks in that the TypeLiteral listener fires
		//      regardless of the containing TypeReference.
		{
			Code: `
type P = Promise<{ (): void }>;
      `,
			Output: []string{`
type P = Promise<() => void>;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    20,
				},
			},
		},

		// ---- Dimension 4: type literal in tuple — `[{ (): void }]`. parent
		//      is TupleType, which is not on shouldWrapSuggestion's list, so
		//      no surrounding parens are added.
		{
			Code: `
type T = [{ (): void }];
      `,
			Output: []string{`
type T = [() => void];
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    13,
				},
			},
		},

		// ---- Real-user: deeply-nested union with an inline call-signature
		//      type literal. Locks in that union-arm wrapping kicks in even
		//      after multiple parent layers.
		{
			Code: `
type T = string | { (a: number): boolean } | null;
      `,
			Output: []string{`
type T = string | ((a: number) => boolean) | null;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    21,
				},
			},
		},

		// ---- Locks in interface fix with extends Function AND a single
		//      construct signature — upstream tests only mix `extends
		//      Function` with a call signature (invalid-9), not a construct
		//      signature.
		{
			Code: `
interface Ctor extends Function {
  new (): Ctor;
}
      `,
			Output: []string{`
type Ctor = new () => Ctor;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in collectThisTypes() arm: `this` inside a nested type
		//      literal that itself is inside another nested type literal —
		//      the nesting counter must reach 2 and still skip the `this`.
		{
			Code: `
interface Foo {
  (): { outer: { inner: this } };
}
      `,
			Output: []string{`
type Foo = () => { outer: { inner: this } };
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in CallSignature with its own generic type parameters —
		//      `<T extends string>(arg: T): T` lives inside the interface
		//      body and must survive the rewrite verbatim (raw-slice copies
		//      the angle-bracket header).
		{
			Code: `
interface Foo {
  <T extends string>(arg: T): T;
}
      `,
			Output: []string{`
type Foo = <T extends string>(arg: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in ConstructSignature with its own generic type
		//      parameters — `new <T>(arg: T): T` is a separate combination
		//      from plain call-sig generics.
		{
			Code: `
interface Ctor {
  new <T>(arg: T): T;
}
      `,
			Output: []string{`
type Ctor = new <T>(arg: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: type literal in conditional-type extends position
		//      — `T extends { (): void } ? ... : ...`. The TypeLiteral
		//      listener must fire regardless of context. Parent is
		//      ConditionalType (not on shouldWrapSuggestion's list), so no
		//      parens are added.
		{
			Code: `
type Check<T> = T extends { (): void } ? true : false;
      `,
			Output: []string{`
type Check<T> = T extends () => void ? true : false;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    29,
				},
			},
		},

		// ---- Dimension 4: type literal in conditional-type false branch.
		//      Same rationale, separate AST path.
		{
			Code: `
type Check<T> = T extends string ? T : { (): void };
      `,
			Output: []string{`
type Check<T> = T extends string ? T : () => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    42,
				},
			},
		},

		// ---- Dimension 4: type literal as a generic type argument inside
		//      Map<...>. The TypeLiteral parent is TypeArgument (which is
		//      part of TypeReferenceNode), not on the wrap list.
		{
			Code: `
type M = Map<string, { (key: string): number }>;
      `,
			Output: []string{`
type M = Map<string, (key: string) => number>;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    24,
				},
			},
		},

		// ---- Locks in shouldWrapSuggestion ArrayType arm with multi-level
		//      arrays — `{ (): void }[][]`. Only the immediate parent
		//      determines wrapping; reaching through two ArrayType levels is
		//      not needed (immediate parent of the type literal is the inner
		//      ArrayType).
		{
			Code: `
type X = { (): void }[][];
      `,
			Output: []string{`
type X = (() => void)[][];
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    12,
				},
			},
		},

		// ---- Locks in shouldWrapSuggestion IntersectionType arm with
		//      multiple non-callable intersection terms — three-way
		//      intersection `A & B & { (): void }`. shouldWrapSuggestion
		//      returns true so parens wrap the rewritten function type.
		{
			Code: `
type T = { a: number } & { b: string } & { (): void };
      `,
			Output: []string{`
type T = { a: number } & { b: string } & (() => void);
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      2,
					Column:    44,
				},
			},
		},

		// ---- Real-user: Promise executor pattern (matches the actual signature
		//      lib.es2015.promise.d.ts exposes). Generic + nested function-type
		//      parameters + optional parameter.
		{
			Code: `
type Executor<T> = {
  (resolve: (value: T) => void, reject: (reason?: any) => void): void;
};
      `,
			Output: []string{`
type Executor<T> = (resolve: (value: T) => void, reject: (reason?: any) => void) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: comparator for Array.sort. Two-arg numeric return —
		//      the canonical shape that documentation always shows as a
		//      function type, but real codebases sometimes write as interface.
		{
			Code: `
interface Comparator<T> {
  (a: T, b: T): number;
}
      `,
			Output: []string{`
type Comparator<T> = (a: T, b: T) => number;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: type guard predicate return — `value is X` form must
		//      be preserved verbatim. The return-type slice runs from after
		//      `:` to the end of the member, so `value is string` falls into
		//      the post-colon text untouched.
		{
			Code: `
interface IsString {
  (value: unknown): value is string;
}
      `,
			Output: []string{`
type IsString = (value: unknown) => value is string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: assertion signature return — TS 3.7+ `asserts ...`
		//      keyword in return position. Same raw-slice path as type guards
		//      but distinct token shape.
		{
			Code: `
interface AssertString {
  (value: unknown): asserts value is string;
}
      `,
			Output: []string{`
type AssertString = (value: unknown) => asserts value is string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: Redux-style selector — generic state-to-result.
		//      Very common pattern that real codebases write as interface to
		//      get nominal-ish typing.
		{
			Code: `
interface Selector<S, R> {
  (state: S): R;
}
      `,
			Output: []string{`
type Selector<S, R> = (state: S) => R;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: middleware signature (Express/Koa) — three-arg with
		//      `next` callback. Tests parameter list with a typed function-
		//      type parameter mixed with primitives.
		{
			Code: `
interface Middleware {
  (req: Request, res: Response, next: (err?: Error) => void): void;
}
      `,
			Output: []string{`
type Middleware = (req: Request, res: Response, next: (err?: Error) => void) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: decorator factory — `(target: any) => ClassDecorator`
		//      pattern. ClassDecorator itself is a function type so the
		//      return reads as a function-type token sequence.
		{
			Code: `
interface DecoratorFactory {
  (target: any): ClassDecorator;
}
      `,
			Output: []string{`
type DecoratorFactory = (target: any) => ClassDecorator;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: destructured parameter with type annotation —
		//      `({ x, y }: Point): number`. The destructuring pattern is
		//      embedded text inside the parameter list and must be preserved
		//      by the raw-slice rewrite.
		{
			Code: `
interface PointMapper {
  ({ x, y }: { x: number; y: number }): number;
}
      `,
			Output: []string{`
type PointMapper = ({ x, y }: { x: number; y: number }) => number;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in modifier handling: `export interface` without leading
		//      comments — comment relocation branch must still produce a
		//      correct `export type` rewrite (the comments-pre-insert branch
		//      must skip empty cases).
		{
			Code: `
export interface Handler<T> {
  (event: T): void;
}
      `,
			Output: []string{`
export type Handler<T> = (event: T) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: comment between modifier and `interface` keyword.
		//      `export /* note */ interface Foo` must preserve the
		//      `/* note */` verbatim — replace range starts at `interface`,
		//      not at the modifier list, matching upstream ESTree's
		//      `declaration.range[0]` semantics.
		{
			Code: `
export /* note */ interface Foo {
  (): string;
}
      `,
			Output: []string{`
export /* note */ type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Same shape but for `declare` keyword — `declare /* c */ interface`
		//      must preserve the inline comment after `declare`.
		{
			Code: `
declare /* c */ interface Foo {
  (): string;
}
      `,
			Output: []string{`
declare /* c */ type Foo = () => string;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Locks in CallSignature with type parameter that has a default —
		//      `<T = string>` slice must include the default in the rewrite.
		//      This is at the CallSignature level (not the InterfaceDeclaration
		//      level which already has its own test).
		{
			Code: `
interface Foo {
  <T = string>(arg: T): T;
}
      `,
			Output: []string{`
type Foo = <T = string>(arg: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: variance-marked type parameter (TS 4.7+) `in` / `out`.
		//      Variance modifiers go inside the type-parameter declaration and
		//      must be preserved by raw-slicing the header.
		{
			Code: `
interface Box<in out T> {
  (value: T): T;
}
      `,
			Output: []string{`
type Box<in out T> = (value: T) => T;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: empty parameter list with an inline block comment
		//      between the parens — `(/* c */): void`. Inner trivia inside
		//      the parameter list must be preserved by raw slicing.
		{
			Code: `
interface Foo {
  (/* inline */): void;
}
      `,
			Output: []string{`
type Foo = (/* inline */) => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: trailing comma in parameter list — `(a, b,): void`.
		//      Modern TS allows trailing commas and the raw-slice rewrite must
		//      keep them. (TypeScript itself drops them at emit time but the
		//      source still parses.)
		{
			Code: `
interface Foo {
  (a: number, b: number): number;
}
      `,
			Output: []string{`
type Foo = (a: number, b: number) => number;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Real-user: callable interface returning the same interface type
		//      — a self-referential alias pattern (chainable builder). Locks
		//      in that the rewrite is valid TS even when the body references
		//      the interface name.
		{
			Code: `
interface Chainable {
  (arg: string): Chainable;
}
      `,
			Output: []string{`
type Chainable = (arg: string) => Chainable;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

		// ---- Dimension 4: explicit `: void` versus implicit any — only the
		//      explicit form should reach the rule because checkMember bails
		//      when returnType is nil. This positive case pins the explicit
		//      path while the corresponding valid case (above, `new ();`)
		//      pins the bail.
		{
			Code: `
interface Foo {
  (): void;
}
      `,
			Output: []string{`
type Foo = () => void;
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "functionTypeOverCallableType",
					Line:      3,
					Column:    3,
				},
			},
		},

	})
}
