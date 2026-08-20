// TestIdDenylistExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
// See id_denylist_upstream_test.go for the migrated upstream suite.
//
// N/A: Dimension 3 (autofix boundaries) — the rule emits neither a fix nor a
// suggestion, so there is no edit demand to keep invariant.
// N/A: Dimension 4 ancestor walks — every decision reads the identifier's
// immediate parent, so no traversal can bleed across an arrow body, a method
// body, or a class static block.
// cspell:ignore axios
package id_denylist

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestIdDenylistExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&IdDenylistRule,
		[]rule_tester.ValidTestCase{
			// An empty deny list denies nothing.
			{Code: `var foo = bar;`},

			// ---- Dimension 4: parenthesized receiver and parenthesized assignment target ----
			{Code: `(foo).bar`, Options: deny("bar")},

			// ---- Dimension 4: TS non-null assertion receiver ----
			{Code: `const x = y!.bar;`, Options: deny("bar")},

			// ---- Dimension 4: optional chain ----
			{Code: `foo?.bar;`, Options: deny("bar")},
			{Code: `foo?.bar();`, Options: deny("bar")},
			{Code: `foo?.(bar);`, Options: deny("bar")},
			{Code: `Number?.NaN;`, Options: deny("Number")},

			// ---- Dimension 4: object key forms; only an identifier or a private name
			// is a name this rule can read, and a computed key is an expression ----
			{Code: `var foo = { "bar": 1 };`, Options: deny("bar")},
			{Code: `var foo = { 0: 1 };`, Options: deny("0")},

			// ---- Dimension 4: element access, where only an identifier subscript is a name ----
			{Code: `foo['bar'] = 1;`, Options: deny("bar")},
			{Code: `foo[0] = 1;`, Options: deny("0")},
			{Code: `foo[Symbol.iterator] = 1;`, Options: deny("Symbol", "iterator")},

			// ---- Dimension 4: spread, rest, body-absent and empty forms ----
			{Code: `[...foo.bar];`, Options: deny("bar")},

			// Locks in upstream shouldCheck() arms 4a and 4b: every child of a call or a
			// `new` is skipped, arguments included, while a tagged template and a dynamic
			// import (an ImportExpression upstream) keep theirs checked.
			{Code: `foo(bar);`, Options: deny("bar")},
			{Code: `new foo(bar);`, Options: deny("bar")},

			// Locks in upstream isAssignmentTarget(): the five arms it does have, and the
			// update-expression and bare for-in/of targets it deliberately does not.
			{Code: `foo.bar++;`, Options: deny("bar")},
			{Code: `++foo.bar;`, Options: deny("bar")},
			{Code: `for (foo.bar of x) {}`, Options: deny("bar")},
			{Code: `for (foo.bar in x) {}`, Options: deny("bar")},
			{Code: `({foo: obj.bar});`, Options: deny("bar")},

			// Locks in upstream isImportAttributeKey() arm 2: the object literal has to be
			// the second argument of a dynamic import, and the key has to be a name.
			{Code: `import('x', { get with() { return 1; } });`, Options: deny("with")},

			// Locks in upstream's two roles for a shorthand property, which tsgo spells with
			// a single node: in an object literal the key half is checked and resolves to
			// nothing, in a pattern the key half is skipped and the reference half decides.
			{Code: `({ Number } = x);`, Options: deny("Number")},

			// ---- Dimension 4: JSX name positions, which upstream parses as JSXIdentifier
			// nodes that this rule's Identifier listener never visits ----
			{Code: `const x = <Foo.Bar />;`, Options: deny("Foo", "Bar"), Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized receiver and parenthesized assignment target ----
			{Code: `(foo).bar = 1`, Options: deny("bar", "foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2), restricted("bar", 1, 7)}},
			{Code: `((foo)).bar = 1`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 9)}},
			{Code: `((foo.bar)) = 1`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 7)}},
			{Code: `(foo.bar) = 1`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 6)}},

			// ---- Dimension 4: TS non-null assertion receiver ----
			{Code: `y!.bar = 1;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 4)}},

			// ---- Dimension 4: TS type-expression wrappers ----
			{Code: `(y as any).bar = 1;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 12)}},
			{Code: `(y satisfies any).bar = 1;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 19)}},
			{Code: `const x = y as Foo;`, Options: deny("Foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 16)}},
			{Code: `foo satisfies Bar;`, Options: deny("foo", "Bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 1), restricted("Bar", 1, 15)}},

			// ---- Dimension 4: optional chain ----
			{Code: `foo?.[bar];`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 7)}},

			// ---- Dimension 4: object key forms; only an identifier or a private name
			// is a name this rule can read, and a computed key is an expression ----
			{Code: `var foo = { [bar]: 1 };`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 14)}},
			{Code: `var foo = { bar: 1 };`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 13)}},
			{Code: `class C { #bar; }`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restrictedPrivate("bar", 1, 11)}},

			// ---- Dimension 4: declaration and container forms ----
			{Code: `class foo {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 7)}},
			{Code: `var x = class foo {};`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 15)}},
			{Code: `var foo = () => {};`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5)}},
			{Code: `class C { foo() {} }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 11)}},
			{Code: `class C { foo = () => {}; }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 11)}},
			{Code: `async function foo() {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 16)}},
			{Code: `function* foo() {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 11)}},
			{Code: `async function* foo() {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 17)}},
			{Code: `class C { static #foo() {} }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restrictedPrivate("foo", 1, 18)}},

			// ---- Dimension 4: same-kind nesting, where both the outer and the inner name are reported ----
			{Code: `function foo() { function foo() {} }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10), restricted("foo", 1, 27)}},
			{Code: `class foo { m() { class foo {} } }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 7), restricted("foo", 1, 25)}},
			{Code: `const {a: {b: {c}}} = d;`, Options: deny("a", "b", "c"), Errors: []rule_tester.InvalidTestCaseError{restricted("c", 1, 16)}},

			// ---- Dimension 4: spread, rest, body-absent and empty forms ----
			{Code: `[...foo.bar] = baz;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 9)}},
			{Code: `const {...bar} = baz;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `const [...bar] = baz;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `const {} = foo;`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 12)}},
			{Code: `function foo() {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `declare function foo(): void;`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 18)}},
			{Code: `abstract class C { abstract foo(): void; }`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 29)}},
			{Code: `function foo(a: string): void;
function foo(a: number): void;
function foo(a: any) {}`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10), restricted("foo", 2, 10), restricted("foo", 3, 10)}},

			// ---- Real-user: ESLint issue 16221, an option object passed to a call keeps
			// its keys checked even though the call's own children are not ----
			{Code: `axios({ url: '/api/send', method: 'post', data: {} });`, Options: deny("data", "method"), Errors: []rule_tester.InvalidTestCaseError{restricted("method", 1, 27), restricted("data", 1, 43)}},

			// ---- Real-user: ESLint issue 15188, TS interface members named after primitives ----
			{Code: `interface Foo { number: string; }`, Options: deny("number", "string"), Errors: []rule_tester.InvalidTestCaseError{restricted("number", 1, 17)}},

			// ---- Real-user: ESLint issue 15504, TS enum members named after primitives ----
			{Code: `export enum ValuesTypesIDs { number = 'NUMBER', string = 'STRING' }`, Options: deny("number", "string"), Errors: []rule_tester.InvalidTestCaseError{restricted("number", 1, 30), restricted("string", 1, 49)}},

			// Locks in upstream shouldCheck() arms 4a and 4b: every child of a call or a
			// `new` is skipped, arguments included, while a tagged template and a dynamic
			// import (an ImportExpression upstream) keep theirs checked.
			{Code: "foo`x`;", Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 1)}},
			{Code: `var foo; import(foo);`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 5), restricted("foo", 1, 17)}},

			// Locks in upstream isAssignmentTarget(): the five arms it does have, and the
			// update-expression and bare for-in/of targets it deliberately does not.
			{Code: `for ([foo.bar] of x) {}`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11)}},
			{Code: `for ({a: foo.bar} of x) {}`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 14)}},
			{Code: `foo.bar += 1;`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5)}},

			// Locks in upstream isImportAttributeKey() arm 2: the object literal has to be
			// the second argument of a dynamic import, and the key has to be a name.
			{Code: `f('x', { with: { type: 1 } });`, Options: deny("type"), Errors: []rule_tester.InvalidTestCaseError{restricted("type", 1, 18)}},
			{Code: `import({ with: { type: 1 } });`, Options: deny("type"), Errors: []rule_tester.InvalidTestCaseError{restricted("type", 1, 18)}},
			{Code: `import('x', { [w]: 1 });`, Options: deny("w"), Errors: []rule_tester.InvalidTestCaseError{restricted("w", 1, 16)}},
			{Code: `let obj = { get with() { return 1; } };`, Options: deny("with"), Errors: []rule_tester.InvalidTestCaseError{restricted("with", 1, 17)}},

			// Locks in upstream's two roles for a shorthand property, which tsgo spells with
			// a single node: in an object literal the key half is checked and resolves to
			// nothing, in a pattern the key half is skipped and the reference half decides.
			{Code: `var bar = 1; x = { bar };`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5), restricted("bar", 1, 20)}},
			{Code: `var bar = 1; ({bar} = x);`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5), restricted("bar", 1, 16)}},
			{Code: `x = { Number };`, Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 7)}},

			// Locks in upstream isReferenceToGlobalVariable() for a PrivateIdentifier, which
			// is never a variable reference however its parent reads.
			{Code: `class C { #x; m(o) { return #x in o; } }`, Options: deny("x"), Errors: []rule_tester.InvalidTestCaseError{restrictedPrivate("x", 1, 11), restrictedPrivate("x", 1, 29)}},

			// Locks in upstream isReferenceToGlobalVariable() on a qualified type name: the
			// right side names a member and is always checked, while the left side is a
			// reference and follows the global rule.
			{Code: `let x: Foo.Bar;`, Options: deny("Foo", "Bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 8), restricted("Bar", 1, 12)}},
			{Code: `let a: Array<Foo>;`, Options: deny("Array", "Foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 14)}},

			// ---- Dimension 4: an annotated declaration name is reported over the name
			// alone, where ESLint's TypeScript parser stretches the range over the
			// annotation too ----
			{Code: `let x: Foo;`, Options: deny("x"), Errors: []rule_tester.InvalidTestCaseError{restricted("x", 1, 5)}},
			{Code: `function f(a: string) {}`, Options: deny("a"), Errors: []rule_tester.InvalidTestCaseError{restricted("a", 1, 12)}},
			{Code: `class C { data: string[] = []; }`, Options: deny("data"), Errors: []rule_tester.InvalidTestCaseError{restricted("data", 1, 11)}},

			// ---- Dimension 4: JSX name positions, which upstream parses as JSXIdentifier
			// nodes that this rule's Identifier listener never visits ----
			{Code: `const x = <div className={foo} />;`, Options: deny("div", "className", "foo"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 27)}},
			{Code: `const x = <Foo>{bar}</Foo>;`, Options: deny("Foo", "bar"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `const x = <Foo {...bar} />;`, Options: deny("Foo", "bar"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 20)}},
		},
	)
}
