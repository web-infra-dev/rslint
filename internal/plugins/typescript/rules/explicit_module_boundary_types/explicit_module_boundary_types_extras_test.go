// TestExplicitModuleBoundaryTypesExtras locks in branches and edge shapes
// that the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
package explicit_module_boundary_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExplicitModuleBoundaryTypesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExplicitModuleBoundaryTypesRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: Parenthesized receivers on a typed-context arrow ----
		// tsgo preserves ParenthesizedExpression that ESLint strips.
		// `(arrowFn)` as a typed VariableDeclaration RHS should still be
		// considered a typed function expression via WalkUpParenthesizedExpressions.
		{Code: `
export const x: Foo = ((): string => 'x');
		`},

		// ---- Dimension 4: Type assertion wrappers ----
		{Code: `
export const x = (() => {}) as Foo;
		`},
		{Code: `
export const x = <Foo>(() => {});
		`},

		// ---- Dimension 4: TS satisfies on an arrow returning `as const` ----
		// allowDirectConstAssertionInArrowFunctions defaults to true and the
		// helper peels both `as const` and `satisfies T` layers (tsgo
		// SatisfiesExpression). The upstream suite covers `satisfies R` once;
		// nesting through ParenthesizedExpression is rslint-specific because
		// tsgo doesn't flatten parens.
		{Code: `
type R = { type: string };
export const func = (value: number) =>
  (({ type: 'X', value }) as const satisfies R);
		`},

		// ---- Dimension 4: AccessorProperty (PropertyDeclaration + accessor modifier) ----
		// Upstream's getFunctionHeadLoc has no AccessorProperty case, so
		// `accessor bool = arg => body` reports arg before the head. We
		// mirror that by shrinking the head loc to the arrow itself when
		// the parent has `accessor`.
		{Code: `
export class Foo {
  accessor bar = (): void => {
    return;
  };
}
		`},

		// ---- Dimension 4: Optional method ----
		// tsgo treats `foo?(): void` as a body-less MethodDeclaration with
		// optional QuestionToken. Body-less + return-type-present must not
		// report missingReturnType.
		{Code: `
export interface I {
  foo?(): void;
}
		`},

		// ---- Dimension 4: Default param with type annotation on the pattern ----
		// Upstream's checkParameter early-returns for AssignmentPattern. In
		// tsgo, a default value lives on Parameter.Initializer; we mirror
		// the early-return.
		{Code: `
export function foo(arg = 1): void {}
		`},
		{Code: `
export function foo(arg: number = 1): void {}
		`},

		// ---- Dimension 4: Computed key with string literal ----
		// allowedNames resolves computed-key string literals to their value.
		{
			Code: `
export class Test {
  ['method']() {}
}
		`,
			Options: map[string]interface{}{"allowedNames": []interface{}{"method"}},
		},

		// ---- Real-user: issue #2134 — `new Proxy(obj, { get(): T {} })` ----
		// Methods inside an object passed to a `new` are not directly exported
		// (the outer function provides the typing).
		{Code: `
export function foo(): unknown {
  return new Proxy(apiInstance, {
    get: (target, property) => {
      // implementation
    },
  });
}
		`},

		// ---- Real-user: deep higher-order chains ----
		{Code: `
export const foo = (): ((n: number) => (m: number) => string) => n => m =>
  String(n + m);
		`},

		// ---- Locks in upstream isExportedHigherOrderFunction() arm: nested return ----
		// Mirrors the `Program:exit` higher-order discovery — a function
		// that's the inner of a chain of `return fn` should still be
		// detected when its outer wrapper IS exported.
		{Code: `
export function FunctionDeclaration() {
  return function FunctionExpression_Within_FunctionDeclaration(): () => () => number {
    return () =>
      (): number => 1;
  };
}
		`},

		// ---- Locks in upstream followReference: const → as Foo → export default ----
		// The reference walk must respect type-assertion wrappers.
		{Code: `
const foo = (arg => arg) as Foo;
export default foo;
		`},

		// ---- Locks in upstream Property → checkNode(node.value): object literal export ----
		{Code: `
const test = (): void => { return; };
export default { test };
		`},

		// ---- Locks in TypeChecker-based scope resolution: inner-block shadowing ----
		// `let foo` inside a block shadows the outer `let foo`. The
		// inner-block reassignment must NOT be attributed to the outer
		// `foo` (which is the one that gets exported). Hand-rolled
		// name-string matching would have falsely flagged the inner write;
		// symbol-based resolution gets this right.
		{Code: `
let foo: (x: number) => number = (x: number): number => x;
{
  let foo = (arg) => arg;
  foo = (a) => a;
}
export default foo;
		`},

		// ---- Locks in TypeChecker-based scope resolution: function-param shadowing ----
		// A function parameter `test` shadows the outer `test`. Reassigning
		// it inside the function body must not pollute the outer export.
		{Code: `
let test: (n: number) => number = (n: number): number => n;
function shadow(test: any) {
  test = (a) => a;
}
export default test;
		`},

		// ---- Locks in shouldCheckDefinition import-binding skip ----
		// `import { foo } from 'x'; export { foo }` — upstream's
		// DefinitionType.ImportBinding skip means we don't check the
		// import declaration. The import provides the contract; the local
		// binding has no observable function shape we can validate.
		{Code: `
import { foo } from './mod';
export { foo };
		`},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: Parameter property (TSParameterProperty equivalent) ----
		// `constructor(public foo)` — upstream recurses into the inner
		// Identifier and reports there. In tsgo we collapse them into a
		// single Parameter; the report range must still anchor on the
		// binding name, not the modifier.
		{
			Code: `
export class Test {
  constructor(public foo) {}
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 3, Column: 22, EndLine: 3, EndColumn: 25},
			},
		},

		// ---- Dimension 4: Rest with destructuring pattern → "Rest" label, not "Array pattern" ----
		// Upstream's report message for `...[a]` says "Rest argument should be
		// typed", not "Array pattern argument should be typed". We mirror that
		// even though tsgo represents both inside a single Parameter node.
		{
			Code: `
export function foo(...[a]): void {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArgTypeUnnamed",
					Message:   "Rest argument should be typed.",
					Line:      2,
					Column:    21,
				},
			},
		},
		{
			Code: `
export function foo(...{ a }): void {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArgTypeUnnamed",
					Message:   "Rest argument should be typed.",
					Line:      2,
					Column:    21,
				},
			},
		},

		// ---- Dimension 4: Multiple destructured params on the same function ----
		// Diagnostic ordering — multiple missingArgType reports must come
		// back in source order (sort by loc.start).
		{
			Code: `
export function foo({ a }, [b], c): void {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgTypeUnnamed", Line: 2, Column: 21},
				{MessageId: "missingArgTypeUnnamed", Line: 2, Column: 28},
				{MessageId: "missingArgType", Line: 2, Column: 33},
			},
		},

		// ---- Locks in upstream checkParameter "any" arm with allowArgumentsExplicitlyTypedAsAny default ----
		// Default is false. `any` on a non-named binding triggers the
		// anyTypedArgUnnamed variant.
		{
			Code: `
export function foo({ foo }: any): void {}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "anyTypedArgUnnamed",
					Message:   "Object pattern argument should be typed with a non-any type.",
					Line:      2,
					Column:    21,
				},
			},
		},

		// ---- Locks in upstream isAllowedName computed-key arm: numeric literal ----
		// Computed keys with numeric literals canonicalise to their textual
		// value. `[0]` matches allowedNames: ["0"] but not ["zero"].
		{
			Code: `
export class Test {
  [0]() {
    return;
  }
}
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingReturnType", Line: 3},
			},
			Options: map[string]interface{}{"allowedNames": []interface{}{"zero"}},
		},

		// ---- Locks in followReference + write reassignment chain ----
		// `let foo; foo = ...; foo = ...; export default foo;`
		// Every write expression must be checked. The second write supplies
		// types, but the first reassignment is still untyped → diagnostic.
		{
			Code: `
let foo: ((arg: number) => number) | undefined;
foo = (arg: number): number => arg;
foo = arg => arg;
export default foo;
			`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArgType", Line: 4},
				{MessageId: "missingReturnType", Line: 4},
			},
		},

		// ---- Locks in checkClassMember + accessor field arrow without parens ----
		// `accessor bool = arg => body` — diagnostic order: param before
		// missingReturnType (since head loc collapses to the arrow itself).
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
	})
}
