// TestNoUnsafeDeclarationMergingExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Every expected output here was
// verified against upstream's reference implementation (ESLint +
// @typescript-eslint/eslint-plugin) — diagnostic count, position, and
// messageId match byte-for-byte. The cases extend the upstream Layer 1
// migration along three axes:
//
//   - Dimension 4 universal edge shapes: receiver wrappers, declaration
//     container forms (class expression vs class declaration, anonymous
//     default export, generics with / without constraints), modifiers
//     (declare, abstract, export), and nesting boundaries (function /
//     arrow body, class static block, deeply nested namespaces, different
//     namespaces with the same names).
//   - Upstream branch lock-ins: every reachable branch in the upstream
//     listener (`if (node.id)`, `defs.length <= 1`, `defs.some(... ===
//     unsafeKind)`, the type-parameter-scope detour on the interface side)
//     has at least one input that exercises it.
//   - Real-user shapes: cross-`declare module` blocks, ambient-module
//     augmentation of external classes, and the default-export class
//     binding pattern (`export default class Foo {}`) where tsgo binds the
//     class to the synthetic `__default` symbol — all forms that show up
//     in real codebases (rsbuild / rspack augmentation patterns) and
//     would silently regress if the same-container filtering or the
//     default-export local-binding fallback were broken.
package no_unsafe_declaration_merging

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeDeclarationMergingExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnsafeDeclarationMergingRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: Declaration container forms — class expression must not trigger ----
		// KindClassExpression is not KindClassDeclaration; only the declaration listener runs.
		{Code: `
interface Foo {}
const A = class Foo {};
`},

		// ---- Dimension 4: Declaration container forms — anonymous default-export class ----
		// Upstream `if (node.id)` short-circuits when the class has no name; rslint's name == nil check does the same.
		{Code: `
interface Foo {}
export default class {}
`},

		// ---- Dimension 4: Nesting boundaries — class inside function body, interface at file level ----
		// Different enclosing LocalsContainer (function body vs SourceFile); not a merge.
		{Code: `
interface Foo {}
function bar() {
  class Foo {}
  return Foo;
}
`},

		// ---- Dimension 4: Nesting boundaries — interface inside function body, class at file level ----
		{Code: `
class Foo {}
function bar() {
  interface Foo {
    x: number;
  }
  return null as unknown as Foo;
}
`},

		// ---- Dimension 4: Nesting boundaries — interface inside an arrow body ----
		{Code: `
class Foo {}
const bar = () => {
  interface Foo { x: number }
  return null as unknown as Foo;
};
`},

		// ---- Dimension 4: Nesting boundaries — interface inside a class static block ----
		{Code: `
class Holder {
  static {
    interface Foo {}
    type _ = Foo;
  }
}
class Foo {}
`},

		// ---- Dimension 4: Nesting boundaries — different namespaces with the same names ----
		// Each namespace block is its own LocalsContainer.
		{Code: `
namespace A {
  class Foo {
    x = 1;
  }
}
namespace B {
  interface Foo {
    y: number;
  }
}
`},

		// ---- Branch lock-in: defs.length <= 1 (class side) — single class with no peer.
		{Code: `
class Foo {
  bar() {}
}
`},

		// ---- Branch lock-in: defs.length <= 1 (interface side).
		{Code: `
interface Foo {
  x: number;
}
`},

		// ---- Branch lock-in: defs > 1 but no class peer — class + namespace merge is safe.
		{Code: `
class Foo {}
namespace Foo {
  export const bar = 1;
}
`},

		// ---- Branch lock-in: defs > 1 but no class peer (interface side) — interface + namespace.
		{Code: `
interface Foo {
  x: number;
}
namespace Foo {
  export const y = 1;
}
`},

		// ---- Branch lock-in: multiple interfaces merging into one symbol with no class — safe.
		{Code: `
interface Foo {
  x: number;
}
interface Foo {
  y: string;
}
`},

		// ---- Branch lock-in: function + class in different namespaces, no merge.
		{Code: `
namespace ns {
  function Foo() {}
}
namespace ns {
  class Bar {}
}
`},

		// ---- Branch lock-in: enum + class with different names — distinct symbols.
		{Code: `
enum Status {}
class Worker {}
`},

		// ---- Branch lock-in: type alias does not participate in declaration merging with classes.
		{Code: `
type Bar = number;
class Foo {}
`},

		// ---- Real-user: class implements a separately-named interface — the canonical safe form.
		{Code: `
interface IFoo {
  bar(): void;
}
class Foo implements IFoo {
  bar() {}
}
`},

		// ---- Real-user: cross-block `declare module` with the same module name ----
		// Each `declare module "X"` block is an independent LocalsContainer. Matches upstream's
		// scope-manager behavior of treating each block as its own scope.
		{Code: `
declare module "lib" {
  interface Foo {}
}
declare module "lib" {
  class Foo {}
}
`},

		// ---- Real-user: cross-block `declare module` with different module names ----
		{Code: `
declare module "a" {
  interface Foo {}
}
declare module "b" {
  class Foo {}
}
`},

		// ---- Real-user: ambient-module augmentation of an external class (rspack/rsbuild pattern) ----
		// Adding an interface declaration to a third-party module's class is idiomatic for
		// native-binding-style packages. The class lives in another file/module, so the
		// interface and class are in different LocalsContainers — not a merge for this rule.
		{Code: `
declare module "@external/lib" {
  class Compiler {}
}
declare module "@external/lib" {
  interface Compiler {
    __augmented?: string;
  }
}
`},

		// ---- Real-user: declared variable + function — no class/interface involvement.
		{Code: `
declare const Foo: number;
function foo() {}
`},

		// ---- Real-user: declare global { interface Foo {} } + module-local class Foo ----
		// Upstream test #8 covers this; reproduced here in case the listener gate changes.
		{Code: `
declare global {
  interface Foo {}
}

class Foo {}
`},

		// ---- Real-user: declare global { class Foo {} } at the top + interface in nested namespace ----
		{Code: `
declare global {
  class Foo {}
}

namespace nested {
  interface Foo { x: number }
}
`},
	}, []rule_tester.InvalidTestCase{
		// ---- Branch lock-in: class + interface inside the same namespace ----
		{
			Code: `
namespace Bar {
  interface Foo {}
  class Foo {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 3, Column: 13},
				{MessageId: "unsafeMerging", Line: 4, Column: 9},
			},
		},

		// ---- Dimension 4: Nesting — three-level nested namespace; same-container filter still finds the merge.
		{
			Code: `
namespace Outer {
  namespace Middle {
    namespace Inner {
      interface Foo {}
      class Foo {}
    }
  }
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 5, Column: 17},
				{MessageId: "unsafeMerging", Line: 6, Column: 13},
			},
		},

		// ---- Dimension 4: Graceful degradation — multiple interfaces all reported, class reported once.
		{
			Code: `
interface Foo {
  x: number;
}
interface Foo {
  y: string;
}
class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 11},
				{MessageId: "unsafeMerging", Line: 5, Column: 11},
				{MessageId: "unsafeMerging", Line: 8, Column: 7},
			},
		},

		// ---- Real-user: class + interface inside a single `declare module` block — merges within the block.
		{
			Code: `
declare module "foo" {
  interface Bar {}
  class Bar {}
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 3, Column: 13},
				{MessageId: "unsafeMerging", Line: 4, Column: 9},
			},
		},

		// ---- Dimension 4: Modifiers — exported class + exported interface.
		{
			Code: `
export interface Foo {}
export class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 18},
				{MessageId: "unsafeMerging", Line: 3, Column: 14},
			},
		},

		// ---- Dimension 4: Modifiers — `declare class` still participates.
		{
			Code: `
declare class Foo {}
interface Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 15},
				{MessageId: "unsafeMerging", Line: 3, Column: 11},
			},
		},

		// ---- Dimension 4: Modifiers — `abstract class` still participates.
		{
			Code: `
abstract class Foo {}
interface Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 16},
				{MessageId: "unsafeMerging", Line: 3, Column: 11},
			},
		},

		// ---- D4: generic interface + generic class — only the class side is reported ----
		// Upstream calls sourceCode.getScope(InterfaceDeclaration), which for a generic interface
		// returns the type-parameter scope. `Foo` is not in that scope, so the lookup fails and
		// the interface listener silently returns. The class listener uses .upper and is not
		// affected. rslint mirrors this asymmetry: interface listener returns when the interface
		// has TypeParameters.
		{
			Code: `
interface Foo<T> {
  x: T;
}
class Foo<T> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 5, Column: 7},
			},
		},

		// ---- D4: generic interface (one type param) + non-generic class — same asymmetry.
		{
			Code: `
interface Foo<T> {}
class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 3, Column: 7},
			},
		},

		// ---- D4: generic interface with two type parameters — same asymmetry.
		{
			Code: `
interface Foo<T, U> {}
class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 3, Column: 7},
			},
		},

		// ---- D4: generic interface with a constrained type parameter and heritage clauses.
		// `<T extends Base> extends IBase` still triggers the type-parameter-scope detour on the
		// interface side, so only the class is reported.
		{
			Code: `
class Base {}
interface IBase {}
interface Foo<T extends Base> extends IBase {
  x: T;
}
class Foo<T extends Base> extends Base implements IBase {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 7, Column: 7},
			},
		},

		// ---- D4: non-generic interface + generic class — interface side IS reported ----
		// The asymmetry only kicks in when the interface itself carries type parameters; class
		// generics are irrelevant. Both sides are reported here.
		{
			Code: `
interface Foo {}
class Foo<T> {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 11},
				{MessageId: "unsafeMerging", Line: 3, Column: 7},
			},
		},

		// ---- Real-user: class declared before interface (reversed order), both with bodies.
		{
			Code: `
class Foo {
  bar() {}
}
interface Foo {
  baz: number;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 7},
				{MessageId: "unsafeMerging", Line: 5, Column: 11},
			},
		},

		// ---- Branch lock-in: `export default class Foo {}` + `interface Foo {}` ----
		// tsgo binds the named default-export class to the synthetic `__default` symbol; the
		// module-scope `Foo` local (which actually merges) is a separate symbol in the enclosing
		// LocalsContainer. The rule's fallback retrieves it via ast.GetLocals(container)[name].
		{
			Code: `
export default class Foo {}
interface Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 22},
				{MessageId: "unsafeMerging", Line: 3, Column: 11},
			},
		},

		// Same merge but with interface first — class still recovered via local-binding fallback.
		{
			Code: `
interface Foo {}
export default class Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 11},
				{MessageId: "unsafeMerging", Line: 3, Column: 22},
			},
		},

		// Same merge but with multiple interfaces — each one is reported plus the default-export class.
		{
			Code: `
export default class Foo {}
interface Foo {}
interface Foo {}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unsafeMerging", Line: 2, Column: 22},
				{MessageId: "unsafeMerging", Line: 3, Column: 11},
				{MessageId: "unsafeMerging", Line: 4, Column: 11},
			},
		},
	})
}
