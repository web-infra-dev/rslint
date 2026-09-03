// TestNoUnreadableNewExpressionExtras locks in tsgo/ESTree shape differences,
// every reachable upstream branch, and representative real-world expressions
// discussed in the upstream issue tracker. The direct v74.0.0 migration lives
// in no_unreadable_new_expression_upstream_test.go.
package no_unreadable_new_expression_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	no_unreadable_new_expression "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_unreadable_new_expression"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreadableNewExpressionExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_unreadable_new_expression.NoUnreadableNewExpressionRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isSimpleConstructor() identifier arm. ESTree
			// omits source parentheses around the callee.
			{Code: `const value = new (((Foo)))();`, FileName: "file.mjs"},

			// Locks in upstream isStaticMemberExpression() recursive-object arm,
			// including parentheses that ESTree does not expose.
			{Code: `const value = new (((namespace).models).Widget)();`, FileName: "file.mjs"},
			{Code: `const value = new (/** @type {typeof namespace} */ (namespace).Widget)();`, FileName: "file.js"},

			// Locks in upstream MemberExpression listener early-return arm: the
			// object is not a NewExpression.
			{Code: `const value = factory().Widget;`, FileName: "file.mjs"},

			// ---- Dimension 4: authored TypeScript wrappers remain visible. ----
			// TSESTree exposes the assertion as the member object, so this is not
			// direct member access from a NewExpression.
			{Code: `const value = (new Foo() as Foo).bar;`, FileName: "file.ts"},
			{Code: `const value = (new Foo() satisfies Foo).bar;`, FileName: "file.ts"},

			// Type arguments do not make an otherwise simple constructor complex.
			{Code: `const value = new namespace.Widget<string>();`, FileName: "file.ts"},

			// tsgo represents dotted JSX tag names with PropertyAccessExpression,
			// but neither link has a NewExpression object.
			{Code: `const value = <Namespace.Widget />;`, FileName: "file.tsx", Tsx: true},

			// N/A: declaration-name/key shapes, destructuring, function/class
			// variants, missing bodies, options, fixes, and suggestions are not
			// inspected by this expression-only rule.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream MemberExpression listener report arm. tsgo splits
			// ESTree's one member kind into property and element access listeners.
			invalid(`const value = new Foo().property;`, "file.mjs", memberAccessError(`const value = new Foo().property;`, "property", 0)),
			invalid(`const value = new Foo()[property];`, "file.mjs", memberAccessError(`const value = new Foo()[property];`, "property", 0)),
			invalid(`const value = new Foo()["property"];`, "file.mjs", memberAccessError(`const value = new Foo()["property"];`, `"property"`, 0)),
			invalid(`class A { #value; method() { return new A().#value; } }`, "file.mjs", memberAccessError(`class A { #value; method() { return new A().#value; } }`, "#value", 1)),

			// ---- Dimension 4: optional-chain member access still has a direct
			// NewExpression object after ESTree removes parentheses. ----
			invalid(`const value = (new Foo())?.property;`, "file.mjs", memberAccessError(`const value = (new Foo())?.property;`, "property", 0)),

			// ---- Dimension 4: optional chains and computed access used as a
			// constructor are complex, not static member chains. These lock in
			// upstream isStaticMemberExpression() rejection and the NewExpression
			// listener report arm. ----
			invalid(`const value = new (namespace?.Widget)();`, "file.mjs", complexConstructorError(`const value = new (namespace?.Widget)();`, "namespace?.Widget", 0)),
			invalid(`const value = new (namespace[Widget])();`, "file.mjs", complexConstructorError(`const value = new (namespace[Widget])();`, "namespace[Widget]", 0)),
			invalid(`class A { #Ctor; method() { return new this.#Ctor(); } }`, "file.mjs", complexConstructorError(`class A { #Ctor; method() { return new this.#Ctor(); } }`, "this.#Ctor", 0)),

			// ---- Dimension 4: JavaScript JSDoc casts are absent from ESTree
			// and therefore remain transparent around a NewExpression object. ----
			invalid(`const value = /** @type {Foo} */ (new Foo()).property;`, "file.js", memberAccessError(`const value = /** @type {Foo} */ (new Foo()).property;`, "property", 0)),

			// ---- Dimension 4: authored TypeScript assertions in a constructor
			// path remain visible and make the constructor complex. ----
			invalid(`const value = new ((namespace as typeof other).Widget)();`, "file.ts", complexConstructorError(`const value = new ((namespace as typeof other).Widget)();`, `(namespace as typeof other).Widget`, 0)),
			invalid(`const value = new (namespace!.Widget)();`, "file.ts", complexConstructorError(`const value = new (namespace!.Widget)();`, "namespace!.Widget", 0)),

			// Same-kind nesting: only the first member directly owned by the new
			// expression is reported, not later links in the chain.
			invalid(`const value = new Foo().first.second.third;`, "file.mjs", memberAccessError(`const value = new Foo().first.second.third;`, "first", 0)),

			// A complex outer constructor can contain another new expression
			// whose direct member access is independently reported.
			invalid(
				`const value = new (new Factory().Builder)();`,
				"file.mjs",
				complexConstructorError(`const value = new (new Factory().Builder)();`, `new Factory().Builder`, 0),
				memberAccessError(`const value = new (new Factory().Builder)();`, `Builder`, 0),
			),

			// Comments and line breaks do not change the direct member boundary.
			invalid("const value = new Foo()\n\t/* keep */ .property;", "file.mjs", memberAccessError("const value = new Foo()\n\t/* keep */ .property;", "property", 0)),

			// ---- Real-user: upstream #1603/#3266 examples. ----
			invalid(`const year = new Date(commit.committedDate).getFullYear();`, "file.mjs", memberAccessError(`const year = new Date(commit.committedDate).getFullYear();`, "getFullYear", 0)),
			invalid(`function check() { return new URL(location.href).searchParams.get('w') === '1'; }`, "file.mjs", memberAccessError(`function check() { return new URL(location.href).searchParams.get('w') === '1'; }`, "searchParams", 0)),
		},
	)
}
