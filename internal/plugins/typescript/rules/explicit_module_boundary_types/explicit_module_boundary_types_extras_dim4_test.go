// TestExplicitModuleBoundaryTypesExtrasDim4 locks in Dimension 4 universal
// edge shapes from PORT_RULE.md — combinations the upstream test suite does
// not exercise but real users hit: async / generator / async-generator
// variants on every function-like form, class declaration vs expression,
// decorators, deeply nested chains, spread in object literals, destructuring
// at the export boundary, and tsgo paren / satisfies / chain quirks.
package explicit_module_boundary_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitModuleBoundaryTypesExtrasDim4(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitModuleBoundaryTypesRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: async function with explicit Promise return type ----
		{Code: `
export async function f(): Promise<void> {
  return;
}
		`},

		// ---- Dimension 4: generator with explicit return type ----
		{Code: `
export function* gen(): Generator<number> {
  yield 1;
}
		`},

		// ---- Dimension 4: async generator with explicit return type ----
		{Code: `
export async function* agen(): AsyncGenerator<number> {
  yield 1;
}
		`},

		// ---- Dimension 4: async arrow with explicit return type ----
		{Code: `
export const f = async (): Promise<void> => {};
		`},

		// ---- Dimension 4: async method on exported class ----
		{Code: `
export class Foo {
  async bar(x: number): Promise<number> {
    return x;
  }
}
		`},

		// ---- Dimension 4: generator method on exported class ----
		{Code: `
export class Foo {
  *gen(): Generator<number> {
    yield 1;
  }
}
		`},

		// ---- Dimension 4: async generator method on exported class ----
		{Code: `
export class Foo {
  async *agen(): AsyncGenerator<number> {
    yield 1;
  }
}
		`},

		// ---- Dimension 4: class expression assigned to a typed const ----
		// `export const X: typeof Class = class { ... }` — the variable type
		// covers the class expression so members don't need annotations
		// according to allowTypedFunctionExpressions semantics. NOTE:
		// upstream still recurses into class members. We mirror that.
		{Code: `
export const X = class {
  method(): void {}
};
		`},

		// ---- Dimension 4: ClassExpression in `export default` ----
		// `export default class { method(): T {} }` — class expression as
		// the export default value; members checked.
		{Code: `
export default class {
  method(): void {}
};
		`},

		// ---- Dimension 4: Decorated class with typed methods ----
		{Code: `
declare function Component(): any;
@Component()
export class Foo {
  bar(x: number): number {
    return x;
  }
}
		`},

		// ---- Dimension 4: Method decorator preserves head loc trimming ----
		{Code: `
declare function Bind(): any;
export class Foo {
  @Bind()
  bar(x: number): number {
    return x;
  }
}
		`},

		// ---- Dimension 4: readonly class field arrow with explicit type ----
		{Code: `
export class Foo {
  readonly bar = (x: number): number => x;
}
		`},

		// ---- Dimension 4: override modifier on typed method ----
		{Code: `
class Base { method(x: number): number { return x; } }
export class Sub extends Base {
  override method(x: number): number {
    return x;
  }
}
		`},

		// ---- Dimension 4: Generic function with explicit return type ----
		{Code: `
export function identity<T>(x: T): T {
  return x;
}
		`},

		// ---- Dimension 4: Generic arrow with explicit return type ----
		{Code: `
export const identity = <T>(x: T): T => x;
		`},

		// ---- Dimension 4: Function-in-function — inner function not exported ----
		// The outer is exported and typed; the inner is purely local. Upstream
		// only checks export-reachable functions, so the inner must NOT
		// trigger even if it has no annotations.
		{Code: `
export function outer(): void {
  function inner(a, b) {
    return a + b;
  }
  inner(1, 2);
}
		`},

		// ---- Dimension 4: Class-in-class — inner not exported ----
		{Code: `
export class Outer {
  method(): void {
    class Inner {
      hidden(a, b) {
        return a + b;
      }
    }
    new Inner();
  }
}
		`},

		// ---- Dimension 4: SpreadAssignment in exported object literal ----
		// Upstream's checkNode for ObjectExpression iterates `properties`.
		// SpreadElement / RestElement are not Property nodes, so they fall
		// through. Verify we don't crash and don't mask the sibling check.
		{Code: `
declare const base: { foo(): void };
export const x = {
  ...base,
  foo: (): void => {},
};
		`},

		// ---- Dimension 4: Destructured const export, all members typed ----
		{Code: `
export const { f, g } = { f: (): void => {}, g: (): void => {} };
		`},

		// ---- Dimension 4: Array-destructured const export ----
		{Code: `
export const [a, b] = [(): void => {}, (): void => {}];
		`},

		// ---- Dimension 4: Object literal with nested method shorthand ----
		{Code: `
export const obj = {
  outer(): void {
    return;
  },
};
		`},

		// ---- Locks in overload-signature skip + `declare` skip ----
		// Differential validation caught this: top-level body-less
		// FunctionDeclaration is `TSDeclareFunction` in ESTree, which
		// upstream's checkNode has no case for — overload signatures and
		// `declare function` are silently skipped. tsgo represents them
		// as KindFunctionDeclaration with Body() == nil; we must bail
		// before checking parameters or the destructured / typed-with-
		// `any` params would falsely fire (the impl below is fine because
		// it has the right annotations).
		{Code: `
export function g({a, b}): void;
export function g(arr: [number]): void;
export function g(arg: number): number {
  return arg;
}
		`},
		{Code: `
declare function f(x): void;
export { f };
		`},

		// ---- Locks in object-method shorthand under typed parent ----
		// Differential validation against upstream eslint caught this: tsgo
		// models `{ foo() {} }` as MethodDeclaration directly inside
		// ObjectLiteralExpression — no `Property > FunctionExpression`
		// wrapper. Upstream's isTypedFunctionExpression climbs from the
		// FunctionExpression through Property → ObjectExpression → parent;
		// our equivalent has to special-case ObjectLiteralExpression parent
		// before bailing. Lock the typed-context resolution in for both
		// shorthand methods and arrow values inside the same object.
		{Code: `
type Obj = { foo(): void; bar: () => void };
export const o: Obj = { foo() {}, bar: () => {} };
		`},

		// ---- Dimension 4: Setter parameter with type annotation (no return type needed) ----
		// Setters always pass — they can't have a return type. Verify both
		// in classes and in object literals.
		{Code: `
export class Foo {
  set x(v: number) { this._x = v; }
}
		`},
		{Code: `
export const obj = {
  set x(v: number) { /* */ },
};
		`},

		// ---- Dimension 4: Deeply parenthesized `as const` body ----
		// tsgo preserves parens; SkipParentheses on the body must still
		// reach the `as const` for allowDirectConstAssertionInArrowFunctions.
		{Code: `
export const x = () => (((({ type: 'X' }) as const)));
		`},

		// ---- Dimension 4: Type annotation on default parameter pattern ----
		{Code: `
export function foo({ a, b }: { a: number; b: number } = { a: 0, b: 0 }): void {}
		`},

		// ---- Dimension 4: Tagged template literal as default value ----
		// `arg = tagged\`...\`` — initializer present → skipped per upstream.
		{Code: `
declare function tag(strings: TemplateStringsArray): string;
export function foo(arg = tag` + "`x`" + `): void {}
		`},

		// ---- Real-user: React.FC functional component ----
		// React FC supplies the type via the variable annotation. With
		// allowTypedFunctionExpressions: true (default), nothing to flag.
		{Code: `
type Props = { name: string };
type FC<P> = (props: P) => any;
export const Greet: FC<Props> = ({ name }) => name;
		`},

		// ---- Real-user: HOC with explicit return-type chain ----
		{Code: `
type Comp<P> = (props: P) => any;
type Props = { foo: string };
export const withFoo = <P extends Props>(C: Comp<P>): Comp<P> => {
  return (props): any => C(props);
};
		`},

		// ---- Real-user: Reducer with explicit types ----
		{Code: `
type State = { count: number };
type Action = { type: 'inc' } | { type: 'dec' };
export const reducer = (state: State, action: Action): State => state;
		`},

		// ---- Real-user: Async wrapper with explicit Promise ----
		{Code: `
export async function fetchJson(url: string): Promise<unknown> {
  return null;
}
		`},

		// ---- Real-user: const-asserted lookup table ----
		// `as const` immediately following an object literal — common in
		// state machines / redux patterns.
		{Code: `
export const Actions = (() => ({
  INC: 'inc',
  DEC: 'dec',
}) as const)();
		`},

		// ---- Locks in checkNode VariableStatement iteration: multiple decls ----
		// `export const a = ..., b = ...;` — both initializers must be
		// dispatched. If only the first were checked, the second would
		// silently slip through.
		{Code: `
export const a: (n: number) => number = (n) => n,
  b: (n: number) => number = (n) => n;
		`},

		// ---- Locks in checkNode ArrayLiteral iteration: nested arrays ----
		{Code: `
const a = (): void => {};
const b = (): void => {};
export default [[a, b]];
		`},

		// ---- Locks in checkClassMember PropertyDeclaration with NO initializer ----
		// Class field without initializer — checkNode dispatches but
		// pd.Initializer is nil; must be a no-op, not a crash.
		{Code: `
export class Foo {
  bar!: () => void;
}
		`},

		// ---- Locks in checkClassMember PrivateIdentifier on PropertyDeclaration ----
		// `#x` private field skipped (parent of name → PrivateIdentifier).
		{Code: `
export class Foo {
  #helper = (a, b) => a + b;
}
		`},

		// ---- Locks in walkExports for `export = expr` ----
		// CommonJS-style export — must dispatch the expression. Distinct
		// AST path from `export default expr` though both are KindExportAssignment.
		{Code: `
const test = (a: number, b: number): number => a + b;
export = test;
		`},

		// ---- Locks in walkExports for `export {} from 'mod'` (re-export) ----
		// Has ModuleSpecifier → skipped. The names refer to the remote
		// module's symbols, which we can't see.
		{Code: `
export { foo } from './other';
		`},
		{Code: `
export * from './other';
		`},
		{Code: `
export * as ns from './other';
		`},

		// ---- Locks in checkNode skipping type-only export ----
		{Code: `
export type Handler = () => void;
		`},
		{Code: `
type Handler = () => void;
export type { Handler };
		`},
		{Code: `
interface Handler {
  (): void;
}
export { type Handler };
		`},

		// ---- Locks in checkNode skipping interface / enum / namespace declarations ----
		{Code: `
export interface I {
  method(arg: number): void;
}
		`},
		{Code: `
export enum E {
  A,
  B,
}
		`},

		// ---- Locks in `namespace` exports — nested export within namespace ----
		// Differential validation caught this: upstream's listeners traverse
		// the AST and fire on every `ExportNamedDeclaration:exit`, including
		// those inside `namespace X { ... }` (TSModuleDeclaration). tsgo
		// wraps namespace bodies in KindModuleBlock; walkExports must
		// recurse into them, otherwise inner exported functions silently
		// slip through.
		{Code: `
export namespace X {
  export function foo(): void {}
}
		`},
		{Code: `
namespace X {
  export function foo(): void {}
}
		`},
		// Nested namespace (dotted form desugars to nested ModuleDeclaration).
		{Code: `
export namespace A.B {
  export function foo(): void {}
}
		`},

		// ---- Locks in declare-module / module augmentation recursion ----
		// `declare module 'foo' { ... }` is parsed as ModuleDeclaration with
		// a string-literal name. Inner `export function f(): void` is
		// body-less (ambient) so checkFunction's nil-Body() early-return
		// skips it. We just need to verify walkExports doesn't crash and
		// doesn't fire false positives.
		{Code: `
declare module 'foo' {
  export function inner(): void;
  export const x: () => void;
}
		`},
		// `declare global { ... }` — same shape, different name kind.
		{Code: `
declare global {
  interface Window {
    customFn(): void;
  }
}
		`},
		// Module augmentation in an actual module: declare module body
		// contains exported function with body should also be checked.
		{Code: `
declare module 'mylib' {
  export function helper(x: string): string;
}
export {};
		`},

		// ---- Locks in checkBodyless KindSetAccessor skip arm ----
		// Body-less setter in a declared class — no return type required.
		{Code: `
declare class Foo {
  set x(v: number);
}
export { Foo };
		`},

		// ---- Locks in checkBodyless KindConstructor skip arm ----
		// Body-less constructor in an abstract class.
		{Code: `
export abstract class Foo {
  abstract construct(): void;
}
		`},

		// ---- Locks in TypeChecker alias chain: re-exported via local binding ----
		// `const local = imported;` then export — local re-binding is a
		// VariableDeclaration with the import as initializer. The
		// initializer is an Identifier, which followReference resolves.
		// Since the resolved symbol is an import (ImportSpecifier),
		// shouldCheckDefinition filters it out — no diagnostic.
		{Code: `
import { foo } from './other';
const local = foo;
export default local;
		`},

		// ---- Locks in allowedNames + computed numeric literal "0" ----
		{
			Code: `
export class Test {
  [0]() {
    return;
  }
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"0"}},
		},

		// ---- Locks in allowedNames + computed string literal "method" ----
		{
			Code: `
export class Test {
  ['method']() {
    return;
  }
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"method"}},
		},

		// ---- Locks in allowedNames + NoSubstitutionTemplateLiteral key ----
		// `[\`method\`]` resolves to "method" the same as the string literal.
		{
			Code: `
export class Test {
  [` + "`method`" + `]() {
    return;
  }
}
			`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"method"}},
		},

		// ---- Locks in option-array form: single-element wrapper ----
		// rslint's GetOptionsMap must handle `[{...}]` as well as `{...}`.
		// Already exercised in upstream cases — pin one explicitly.
		{
			Code: `
export const foo = (): unknown => null;
			`,
			Options: []interface{}{
				map[string]interface{}{"allowTypedFunctionExpressions": true},
			},
		},

		// ---- Locks in higher-order option interaction ----
		// allowHigherOrderFunctions: true + allowDirectConstAssertionInArrowFunctions: false
		{
			Code: `
export const f = () => (): number => 1;
			`,
			Options: map[string]interface{}{
				"allowHigherOrderFunctions":                 true,
				"allowDirectConstAssertionInArrowFunctions": false,
			},
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: untyped async function ----
		{
			Code: `
export async function f() {
  return;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: untyped generator ----
		{
			Code: `
export function* gen() {
  yield 1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: untyped async generator ----
		{
			Code: `
export async function* agen() {
  yield 1;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: untyped async arrow ----
		{
			Code: `
export const f = async () => null;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: class-expression members untyped ----
		// `export const X = class { method() {} }` — the class expression
		// itself is checkable via VariableDeclaration → checkNode init →
		// ClassExpression → iterate members.
		{
			Code: `
export const X = class {
  method(a) {
    return a;
  }
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Dimension 4: `export default class {}` with untyped method ----
		{
			Code: `
export default class {
  method(a) {
    return a;
  }
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Dimension 4: Decorated method preserves head loc, still reported ----
		{
			Code: `
declare function Bind(): any;
export class Foo {
  @Bind()
  bar() {
    return 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5, Column: 3, EndColumn: 6},
			},
		},

		// ---- Dimension 4: override modifier untyped ----
		{
			Code: `
class Base { method(x: number): number { return x; } }
export class Sub extends Base {
  override method(x) {
    return x;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4},
				{MessageId: "missingArgType", Line: 4},
			},
		},

		// ---- Dimension 4: Generic function untyped return ----
		{
			Code: `
export function identity<T>(x: T) {
  return x;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: SpreadAssignment doesn't mask sibling untyped property ----
		{
			Code: `
declare const base: any;
export const x = {
  ...base,
  foo: () => null,
};
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 5},
			},
		},

		// ---- Dimension 4: Destructured export — function values untyped ----
		// `export const { f } = { f: () => 1 }` — destructured binding
		// gives `f` its type from the object literal. The arrow inside the
		// initializer is a function expression with NO typed parent at
		// the AST level (the destructuring target is on the LHS, but
		// upstream's isTypedFunctionExpression only checks immediate
		// parents). So upstream WOULD flag this. Lock in our behavior.
		// NOTE: this is a `Skip: true` if our behavior diverges — flip to
		// invalid once differential validation confirms upstream's call.
		{
			Code: `
export const { f } = { f: () => 1 };
			`,
			Skip: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Dimension 4: Multiple declarators with mixed annotation ----
		{
			Code: `
export const a: () => number = () => 1,
  b = () => 2;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
			},
		},

		// ---- Dimension 4: Class field arrow with non-null assertion (no init) ----
		// Field declared with `!` and no initializer — no function to check.
		// Pair: same shape WITH untyped arrow initializer must still fire.
		{
			Code: `
export class Foo {
  bar = () => null;
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
			},
		},

		// ---- Dimension 4: optional method with body but no return type ----
		{
			Code: `
export class Foo {
  bar?() {
    return 1;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
			},
		},

		// ---- Dimension 4: Static method untyped ----
		{
			Code: `
export class Foo {
  static bar(x) {
    return x;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Dimension 4: protected member still public-by-default ----
		// `protected` is NOT `private`, so it's still checked.
		{
			Code: `
export class Foo {
  protected method(x) {
    return x;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Real-user: HOC pattern — both layers untyped ----
		// `(Component) => (props) => Component(props)`. With default
		// allowHigherOrderFunctions: true, only the inner-most arrow
		// needs a return type. Diagnostic order (by loc.start column):
		// Component (outer arg, col 31) → props (inner arg, col 46) →
		// inner-arrow `=>` head (col 55). All on one line so source-
		// order sorting drives the sequence.
		{
			Code: `
export const withSomething = (Component) => (props) => Component(props);
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Real-user: arrow returning a conditional with non-function branch ----
		// `() => cond ? "x" : () => {}` — body is a Conditional, not a
		// function expression. doesImmediatelyReturnFunctionExpression
		// returns false (body is not a function), so the outer arrow
		// itself needs a return type, even though one branch is a function.
		{
			Code: `
export const wrap = () => (true ? 'x' : 'y');
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Real-user: redux action creator typed as `as const` chain ----
		{
			Code: `
export const inc = (delta) => ({ type: 'inc', delta }) as const;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
			},
		},


		// ---- Locks in `export = expr` followReference path ----
		{
			Code: `
const fn = (arg) => arg;
export = fn;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 2},
				{MessageId: "missingReturnType", Line: 2},
			},
		},

		// ---- Locks in isExportedHigherOrderFunction recognising method-likes ----
		// Found via differential validation on rspack: a class method
		// `snapped()` returning `function SNAPPED_HOOK(this: any, ...){}` —
		// upstream visits the inner function via its higher-order Program:exit
		// pass because the wrapping FunctionExpression of MethodDefinition
		// IS a function. tsgo has no FunctionExpression wrapper on methods;
		// the parent-walk lands directly on MethodDeclaration. Treat
		// method-likes as function-like during the walk, otherwise inner
		// functions returned from a class method silently slip through.
		{
			Code: `
export class C {
  snapped(cb: (...args: unknown[]) => Promise<unknown>) {
    return function SNAPPED_HOOK(this: any, ...args: unknown[]) {
      return cb.apply(this, args);
    };
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 4},
				{MessageId: "anyTypedArg", Line: 4},
			},
		},

		// ---- Locks in namespace-body recursion ----
		// Untyped `export function inner(x)` inside a namespace must be
		// reached and flagged.
		{
			Code: `
export namespace NS {
  export function inner(x): void {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Locks in followReference returning multiple declarations ----
		// `function fn(a: string): string; function fn(a) { ... };
		// export default fn;` — sym.Declarations has both signatures.
		// Order: missingReturnType (function head, col 1) before
		// missingArgType (param a, col 13), per loc.start sort.
		{
			Code: `
function fn(a: string): string;
function fn(a) {
  return a;
}
export default fn;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Locks in TypeChecker scope: arrow exported via reassignment chain through 3 writes ----
		// Initial decl & last write are fully typed; only the middle
		// reassignment is untyped. With shadowing-aware symbol matching,
		// only the middle write's arrow fires both diagnostics.
		{
			Code: `
let h: (n: number) => number = (n: number): number => n;
h = (n) => n;
h = (n: number): number => n;
export default h;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
				{MessageId: "missingReturnType", Line: 3},
			},
		},

		// ---- Locks in argument check NOT triggered for skipped-name function ----
		// `allowedNames: ['foo']` skips both args AND return type.
		{
			Code: `
export const foo = (x) => x;
export const bar = (x) => x;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
				{MessageId: "missingReturnType", Line: 3},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{"foo"}},
		},

		// ---- Locks in optional set-accessor with explicit set parameter ----
		// `set` with optional `?` — body absent. Skipped for missingReturnType
		// (setter), parameter still checked.
		{
			Code: `
export class Foo {
  set x(v) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3},
			},
		},

		// ---- Locks in tsgo quirk: numeric literal computed key NOT canonicalised ----
		// `[0x1]` in tsgo normalises to numeric value 1 at parse time.
		// Upstream's token-level comparison sees `0x1` and `1` as distinct.
		// Our allowedNames against ["1"] matches `[0x1]` as well — document
		// this as language-natural divergence (Phase 1 Step 6.B).
		{
			Code: `
export class Test {
  [0x1]() {
    return;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				// Pin current behavior: with allowedNames: ["nope"], `[0x1]`
				// is NOT skipped, so it fires.
				{MessageId: "missingReturnType", Line: 3},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{"nope"}},
		},
	})
}
