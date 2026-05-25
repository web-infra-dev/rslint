// TestNoUnnecessaryQualifierExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST
// quirk it covers, so future refactors can't silently regress them
// without breaking a named lock-in.
package no_unnecessary_qualifier

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryQualifierExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryQualifierRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parenthesized receiver — not matched per documented divergence ----
		// In tsgo, `(A).B` is PAE { Expression: ParenthesizedExpression(...) }.
		// IsEntityNameExpression rejects ParenthesizedExpression, so the rule
		// doesn't fire. Upstream (ESTree elides parens) DOES fire, but its
		// `--fix` produces invalid syntax: the fix range
		// `[qualifier.start, name.start)` covers `A).` and yields `(x` (verified
		// against @typescript-eslint/eslint-plugin@8.59.3 on the scratch repo).
		// We conservatively stay silent so output is always syntactically valid.
		// See rule .md "Differences from ESLint".
		{Code: `
namespace A {
  export const x = 3;
  export const y = (A).x;
}
    `},
		// ---- Dimension 4: non-null assertion on receiver ----
		// `A!.x` — Expression is NonNullExpression, not Identifier / PAE.
		// isEntityNameExpression returns false; no report.
		{Code: `
namespace A {
  export const x = 3;
  export const y = A!.x;
}
    `},
		// ---- Dimension 4: type-assertion wrappers on receiver ----
		// `(A as any).x` / `(<any>A).x` — receiver wraps the identifier in an
		// AsExpression or TypeAssertionExpression inside parens; both fail
		// isEntityNameExpression and are silently ignored.
		{Code: `
namespace A {
  export const x = 3;
  export const y = (A as any).x;
}
    `},
		// ---- Dimension 4: element-access (computed) form ----
		// `A['x']` is KindElementAccessExpression, not KindPropertyAccessExpression;
		// the rule listener never fires.
		{Code: `
namespace A {
  export const x = 3;
  export const y = A['x'];
}
    `},
		// ---- Dimension 4: numeric element-access key ----
		// Same as above but with a numeric key — confirms the rule doesn't
		// reach into ElementAccessExpression regardless of key kind.
		{Code: `
namespace A {
  export const x = 3;
  export const y = A[0 + 'x'];
}
    `},
		// ---- Dimension 4: call-expression receiver ----
		// `foo().B` — Expression is CallExpression; not an entity name and
		// not a namespace value anyway. Rule must not fire.
		{Code: `
namespace A {
  export const x = 3;
  function foo() {
    return A;
  }
  export const y = foo().x;
}
    `},
		// ---- Dimension 4: dotted namespace — outer levels NOT pushed ----
		// `namespace A.B.C { ... }` desugars to three nested
		// ModuleDeclarations but only the INNERMOST has a real block; we
		// push only on the ModuleBlock visit. Accessing `A.foo` from inside
		// C's body — where `foo` is declared in a SEPARATE `namespace A {}`
		// — must NOT fire, because none of A's declarations are on the
		// push stack. Locks in the innermost-only push contract.
		{Code: `
namespace A {
  export const foo = 1;
}
namespace A.B.C {
  const x = A.foo;
}
    `},
		// ---- Real-user: TypeQuery `typeof Foo.x` inside same namespace ----
		// Type-position access through TypeQueryNode is a common real shape.
		// Upstream tests don't include it but it must work: outer fires on Foo.
		// Locked here as the EXPECTED valid form (qualified inside Foo's body
		// is unnecessary), which we cover with an invalid case in the Invalid
		// block below. The valid pair is the cross-namespace lookup.
		{Code: `
namespace Foo {
  export const x = 3;
}
namespace Bar {
  type T = typeof Foo.x;
}
    `},
		// ---- Real-user: namespaced type argument from a different namespace ----
		// `Array<Foo.T>` where Foo is OUTSIDE current scope — qualifier is
		// load-bearing. Locks in the cross-namespace negative for the
		// QualifiedName path inside TypeReference type arguments.
		{Code: `
namespace Foo {
  export type T = number;
}
namespace Bar {
  type U = Array<Foo.T>;
}
    `},
		// ---- Locks in upstream symbolIsNamespaceInScope() arm 3: no alias, decl not in stack ----
		// A namespace symbol whose declaration is outside the enclosing
		// stack. Confirms the "not in scope and not aliased" arm returns
		// false (no report).
		{Code: `
namespace Outer {
  export type T = number;
}
namespace Other {
  type Use = Outer.T;
}
    `},
		// ---- Locks in upstream symbolsAreEqual() inequality arm: same name, different symbols ----
		// Inner T shadows Outer.T; accessing Outer.T inside Inner must NOT
		// be reported because the in-scope T is a different symbol.
		// Mirrors upstream valid case 4 with a minimal shape to keep
		// the shadowing branch locked even if that case is ever pruned.
		{Code: `
namespace Outer {
  export type T = number;
  namespace Inner {
    type T = string;
    const x: Outer.T = 0;
  }
}
    `},
		// ---- Config contract: upstream schema is [] — any options shape must be ignored ----
		// rslint's CLI / JS config passes options through several shapes
		// (bare object, [{...}], null, []); none of them should change
		// behavior or crash. Locks in that the rule's `Run` truly ignores
		// its `options any` parameter, matching upstream's empty `schema`
		// and `defaultOptions: []` contract.
		{Code: `const x: A.B = 3;`, Options: nil},
		{Code: `const x: A.B = 3;`, Options: []interface{}{}},
		{Code: `const x: A.B = 3;`, Options: map[string]interface{}{"someUnknownOption": true}},
		{Code: `const x: A.B = 3;`, Options: []interface{}{map[string]interface{}{"checkX": false}}},
		// ---- Real-user: ImportType qualifier (`import('./foo').T`) outside scope ----
		// `import('./foo').T` is an ImportTypeNode whose Qualifier is an
		// EntityName — for `import('./foo').N.T` the Qualifier is a
		// QualifiedName, which DOES enter our listener. Here the imported
		// namespace is NOT on our push stack, so the rule must not fire.
		// Locks in cross-module reference negative.
		{Code: `
type X = import('./foo').T;
    `},
		// ---- Real-user: same-name namespace + enum, accessing enum via namespace ----
		// `namespace Foo { export enum E { One } }`. Outside Foo, `Foo.E.One`
		// must not be flagged (Foo isn't pushed; we're at top level).
		{Code: `
namespace Foo {
  export enum E {
    One,
  }
}
const v = Foo.E.One;
    `},
		// ---- Real-user: namespace member referenced via different namespace ----
		// `namespace A { export const x = 1 } namespace B { const y = A.x }`
		// — A's declaration isn't on the stack while inside B; qualifier is
		// load-bearing. Different from upstream valid 1 (which uses types);
		// this nails down the value-side equivalent.
		{Code: `
namespace A {
  export const x = 1;
}
namespace B {
  const y = A.x;
}
    `},
		// ---- Real-user: empty namespace body shouldn't crash the visitor ----
		// `namespace A {}` produces a ModuleDeclaration with an empty
		// ModuleBlock. Our push/pop must handle an empty body cleanly
		// (no nodes inside to visit). No diagnostic expected.
		{Code: `namespace A {}`},
		// ---- Real-user: declare global { namespace N { } } pattern ----
		// `declare global` is itself a ModuleDeclaration with a ModuleBlock
		// body containing more namespaces. We push global; inner N also
		// pushes. Inside N's body, qualified access works the same way
		// as a non-declare namespace — locked here as a no-fire reference
		// to confirm declare ambient context doesn't cause crashes.
		{Code: `
declare global {
  namespace N {
    const x: number;
  }
}
export {};
    `},
		// ---- Locks in upstream isPropertyAccessExpression() narrowing: PrivateIdentifier name ----
		// `o.#x` is a PAE whose Name is a PrivateIdentifier (KindPrivateIdentifier),
		// not an Identifier. The rule listener's Name-kind guard returns
		// early without invoking the qualifier logic — even though the
		// receiver `o` would pass IsEntityNameExpression. No diagnostic.
		{Code: `
class C {
  #x = 1;
  read(o: C) {
    return o.#x;
  }
}
    `},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in upstream visitNamespaceAccess() arm 1: currentFailedNamespaceExpression guard ----
		// Triple-nested QualifiedName chain `A.B.C.D` inside the C namespace
		// body. Outer QN fires once on `A.B.C`; the inner QN `A.B` and the
		// even-inner `A` must NOT fire (the cursor blocks them). Then the
		// cursor resets on the outer node's exit so the SIBLING access
		// `A.B` further down still gets visited and reported independently.
		{
			Code: `
namespace A {
  export namespace B {
    export type T = number;
    export namespace C {
      export type D = number;
      const u: A.B.C.D = 0;
      const w: A.B.T = 0;
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 7, Column: 16},
				{MessageId: "unnecessaryQualifier", Line: 8, Column: 16},
			},
			Output: []string{`
namespace A {
  export namespace B {
    export type T = number;
    export namespace C {
      export type D = number;
      const u: D = 0;
      const w: T = 0;
    }
  }
}
      `},
		},
		// ---- Real-user: TypeQuery `typeof A.x` inside the same namespace ----
		// Type-position value reference through TypeQueryNode whose expression
		// is a PropertyAccessExpression. The PAE listener fires inside the
		// TypeQuery, removing the qualifier.
		{
			Code: `
namespace Foo {
  export const x = 3;
  type T = typeof Foo.x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 19},
			},
			Output: []string{`
namespace Foo {
  export const x = 3;
  type T = typeof x;
}
      `},
		},
		// ---- Real-user: type argument in TypeReference inside same namespace ----
		// `Array<Foo.T>` resolved from within `Foo` — exercises the QN
		// visitor reached via TypeArguments rather than a top-level
		// QualifiedName.
		{
			Code: `
namespace Foo {
  export type T = number;
  type U = Array<Foo.T>;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 18},
			},
			Output: []string{`
namespace Foo {
  export type T = number;
  type U = Array<T>;
}
      `},
		},
		// ---- Locks in upstream isEntityNameExpression() recursion arm: PAE-of-PAE chain ----
		// Confirms the recursive descent through `A.B.C` (PAE-of-PAE-of-Id)
		// classifies the receiver as an entity name and fires on the outer.
		{
			Code: `
namespace A {
  export namespace B {
    export const C = { x: 3 };
    const y = A.B.C.x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 15},
			},
			Output: []string{`
namespace A {
  export namespace B {
    export const C = { x: 3 };
    const y = C.x;
  }
}
      `},
		},
		// ---- Dimension 4: optional-chain qualifier on namespace value ----
		// `A?.x` is legal syntax even though `A` (a namespace value) is
		// never undefined. The PAE listener doesn't filter on QuestionDotToken
		// (mirrors upstream's MemberExpression[computed=false] selector,
		// which also matches optional chains). The fix removes `A?.`.
		{
			Code: `
namespace A {
  export const x = 3;
  export const y = A?.x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 20},
			},
			Output: []string{`
namespace A {
  export const x = 3;
  export const y = x;
}
      `},
		},
		// ---- Diagnostic contract: exact message text per upstream template ----
		// Upstream: `Qualifier is unnecessary since '{{ name }}' is in scope.`
		// Locks the substituted-name form so future refactors can't drift the
		// message wording or the placeholder value silently.
		{
			Code: `
namespace A {
  export type LongName = number;
  const x: A.LongName = 0;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessaryQualifier",
					Message:   "Qualifier is unnecessary since 'LongName' is in scope.",
					Line:      4,
					Column:    12,
				},
			},
			Output: []string{`
namespace A {
  export type LongName = number;
  const x: LongName = 0;
}
      `},
		},
		// ---- Real-user: interface extends QualifiedName inside same namespace ----
		// Heritage clauses thread through ExpressionWithTypeArguments whose
		// Expression resolves via the PAE/QN visitor. Common pattern in
		// declaration-merging shapes.
		{
			Code: `
namespace A {
  export interface I {
    x: number;
  }
  interface J extends A.I {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 6, Column: 23},
			},
			Output: []string{`
namespace A {
  export interface I {
    x: number;
  }
  interface J extends I {}
}
      `},
		},
		// ---- Real-user: class extends QualifiedName ----
		// `class C extends A.Base {}` from inside A. The Base is referenced
		// via PAE (ExpressionWithTypeArguments wraps the class identifier
		// access). Locks in heritage-clause traversal for value-side too.
		{
			Code: `
namespace A {
  export class Base {}
  class Sub extends A.Base {}
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 21},
			},
			Output: []string{`
namespace A {
  export class Base {}
  class Sub extends Base {}
}
      `},
		},
		// ---- Real-user: generic type-parameter constraint with QualifiedName ----
		// `class C<T extends A.X>` exercises QN reached through a
		// TypeParameter's Constraint slot rather than a top-level TypeReference.
		{
			Code: `
namespace A {
  export type X = string;
  class C<T extends A.X> {
    constructor(public v: T) {}
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 21},
			},
			Output: []string{`
namespace A {
  export type X = string;
  class C<T extends X> {
    constructor(public v: T) {}
  }
}
      `},
		},
		// ---- Real-user: mapped type key reference uses qualified name ----
		// Mapped type `[K in keyof A.M]` reaches QN through a MappedTypeNode's
		// `Type` slot. Confirms the visitor isn't gated on a specific parent.
		{
			Code: `
namespace A {
  export type M = { a: 1; b: 2 };
  type Keys = { [K in keyof A.M]: 1 };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 29},
			},
			Output: []string{`
namespace A {
  export type M = { a: 1; b: 2 };
  type Keys = { [K in keyof M]: 1 };
}
      `},
		},
		// ---- Real-user: conditional type with qualifier ----
		// `T extends A.U ? ...` — conditional type's CheckType / ExtendsType
		// slots are independently traversed; the rule fires on each side
		// that has a redundant qualifier.
		{
			Code: `
namespace A {
  export type U = number;
  type C<T> = T extends A.U ? T : never;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 25},
			},
			Output: []string{`
namespace A {
  export type U = number;
  type C<T> = T extends U ? T : never;
}
      `},
		},
		// ---- Locks in upstream symbolIsNamespaceInScope() alias-recursion arm ----
		// `import * as Foo from './foo'` makes Foo an alias to the imported
		// namespace. From inside `declare module './foo'`, Foo's symbol has
		// SymbolFlagsAlias set; symbolIsNamespaceInScope recurses through
		// getAliasedSymbol and finds the aliased module's declaration on
		// the stack. This is the upstream invalid #9 shape, locked in as
		// extras with explicit position + alias-arm comment so it can't
		// silently regress if Layer 1 is ever pruned.
		{
			Code: `
import * as Foo from './foo';
declare module './foo' {
  type Use = Foo.T;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 14},
			},
			Output: []string{`
import * as Foo from './foo';
declare module './foo' {
  type Use = T;
}
      `},
		},
		// ---- Locks in cursor reset after exit: sibling chain in different child ----
		// First QN chain fires and sets cursor; on exit, cursor resets; the
		// second sibling chain inside the SAME statement list fires
		// independently. Triple chains here exercise depth + reset + sibling
		// interplay in a tighter setting than the larger lock-in above.
		{
			Code: `
namespace N {
  export type X = number;
  type A = N.X;
  type B = N.X;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 12},
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 12},
			},
			Output: []string{`
namespace N {
  export type X = number;
  type A = X;
  type B = X;
}
      `},
		},
		// ---- Real-user: nested enum inside namespace, qualified through both ----
		// `namespace N { export enum E { A, B = N.E.A } }`. Three things
		// happen: enter namespace N (push), then enter enum E (push), then
		// the PAE chain `N.E.A` in B's initializer must fire on `N.E`
		// (which resolves to the enum, currently pushed) and reduce to `A`.
		{
			Code: `
namespace N {
  export enum E {
    A,
    B = N.E.A,
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 9},
			},
			Output: []string{`
namespace N {
  export enum E {
    A,
    B = A,
  }
}
      `},
		},
		// ---- Config contract: invalid case carries options too (CLI-shape) ----
		// Same rule semantics regardless of options shape. CLI ships a bare
		// map after config.go unwraps single-element arrays; pass that
		// shape directly to confirm the unwrap path doesn't accidentally
		// gate the listeners on an option that doesn't exist.
		{
			Code: `
namespace A {
  export type B = number;
  const x: A.B = 3;
}
      `,
			Options: map[string]interface{}{"anyKey": "anyValue"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 12},
			},
			Output: []string{`
namespace A {
  export type B = number;
  const x: B = 3;
}
      `},
		},
		// ---- Real-user: type position in function parameter ----
		{
			Code: `
namespace N {
  export type T = number;
  function f(x: N.T) {
    return x;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 17},
			},
			Output: []string{`
namespace N {
  export type T = number;
  function f(x: T) {
    return x;
  }
}
      `},
		},
		// ---- Real-user: type position in function return ----
		{
			Code: `
namespace N {
  export type T = number;
  function f(): N.T {
    return 0 as N.T;
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 17},
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 17},
			},
			Output: []string{`
namespace N {
  export type T = number;
  function f(): T {
    return 0 as T;
  }
}
      `},
		},
		// ---- Real-user: array element type position ----
		{
			Code: `
namespace N {
  export type T = number;
  const xs: N.T[] = [];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 13},
			},
			Output: []string{`
namespace N {
  export type T = number;
  const xs: T[] = [];
}
      `},
		},
		// ---- Real-user: union & intersection type positions ----
		{
			Code: `
namespace N {
  export type A = 1;
  export type B = 2;
  type U = N.A | N.B;
  type I = N.A & N.B;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 12},
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 18},
				{MessageId: "unnecessaryQualifier", Line: 6, Column: 12},
				{MessageId: "unnecessaryQualifier", Line: 6, Column: 18},
			},
			Output: []string{`
namespace N {
  export type A = 1;
  export type B = 2;
  type U = A | B;
  type I = A & B;
}
      `},
		},
		// ---- Real-user: tuple element type position ----
		{
			Code: `
namespace N {
  export type X = 1;
  export type Y = 2;
  type T = [N.X, N.Y];
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 13},
				{MessageId: "unnecessaryQualifier", Line: 5, Column: 18},
			},
			Output: []string{`
namespace N {
  export type X = 1;
  export type Y = 2;
  type T = [X, Y];
}
      `},
		},
		// ---- Real-user: object-literal type field ----
		{
			Code: `
namespace N {
  export type T = number;
  type O = { v: N.T };
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 17},
			},
			Output: []string{`
namespace N {
  export type T = number;
  type O = { v: T };
}
      `},
		},
		// ---- Real-user: generic default parameter ----
		// `class C<T = N.X>` — TypeParameter's Default slot holds a
		// TypeReference whose TypeName is a QualifiedName.
		{
			Code: `
namespace N {
  export type X = string;
  class C<T = N.X> {
    constructor(public v: T) {}
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 4, Column: 15},
			},
			Output: []string{`
namespace N {
  export type X = string;
  class C<T = X> {
    constructor(public v: T) {}
  }
}
      `},
		},
		// ---- Locks in upstream qualifierIsUnnecessary() arm 5: 4-level nesting ----
		// Quadruple chain `A.B.C.D.X` — outer QN must fire once on `A.B.C.D`,
		// suppressing all 3 inner QN matches via the cursor. Output should
		// reduce to bare `X`. Confirms cursor survives arbitrary depth.
		{
			Code: `
namespace A {
  export namespace B {
    export namespace C {
      export namespace D {
        export type X = number;
        const v: A.B.C.D.X = 0;
      }
    }
  }
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 7, Column: 18},
			},
			Output: []string{`
namespace A {
  export namespace B {
    export namespace C {
      export namespace D {
        export type X = number;
        const v: X = 0;
      }
    }
  }
}
      `},
		},
		// ---- Real-user: declaration merging across multiple namespace blocks ----
		// `namespace N { export const x = 1 } namespace N { ... N.x ... }`.
		// N's symbol has TWO declarations (both ModuleDeclarations); the
		// second one is on the stack while visiting its body. Inside, the
		// `N.x` qualifier resolves via N's symbol, whose declarations
		// include the currently-pushed second block — so it fires.
		{
			Code: `
namespace N {
  export const x = 1;
}
namespace N {
  export const y = N.x;
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unnecessaryQualifier", Line: 6, Column: 20},
			},
			Output: []string{`
namespace N {
  export const x = 1;
}
namespace N {
  export const y = x;
}
      `},
		},
	})
}
