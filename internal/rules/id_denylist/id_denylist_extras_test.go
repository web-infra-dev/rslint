// TestIdDenylistExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
// See id_denylist_upstream_test.go for the migrated upstream suite.
//
// N/A: Dimension 3 (autofix boundaries) — the rule emits neither a fix nor a
// suggestion, so there is no edit demand to keep invariant.
// N/A: Dimension 4 scope-container walks — the rule classifies an identifier
// from the expression around it and never looks for an enclosing function,
// method, arrow, or class static block, so no traversal can bleed across one.
// cspell:ignore axios
package id_denylist

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
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

			// Parentheses are transparent when upstream decides whether an identifier
			// is the callee or an argument of a call.
			{Code: `(foo)();`, Options: deny("foo")},
			{Code: `foo((bar));`, Options: deny("bar")},
			{Code: `new (foo)();`, Options: deny("foo")},
			{Code: `new foo((bar));`, Options: deny("bar")},
			{Code: `(((foo)))();`, Options: deny("foo")},
			{Code: `call((((bar))));`, Options: deny("bar")},
			{Code: `((foo))?.();`, Options: deny("foo")},
			{Code: `call?.(((bar)));`, Options: deny("bar")},

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

			// A denied name is matched against the private name without its `#`, so
			// denying the spelling with the `#` matches nothing.
			{Code: `class C { #bar; }`, Options: deny("#bar")},

			// A literal that only feeds a member access is read as a value, however
			// the result of that access is then destructured.
			{Code: `[[obj.b][0]] = d;`, Options: deny("b")},
			{Code: `[...[obj.b][0]] = d;`, Options: deny("b")},
			{Code: `({ x: { a: obj.b }.c } = d);`, Options: deny("b")},
			{Code: `({ a: { ...obj.b }.c } = d);`, Options: deny("b")},
			{Code: `[foo.bar]! = source;`, Options: deny("bar")},
			{Code: `({x: foo.bar})! = source;`, Options: deny("bar")},
			{Code: `[foo.bar!] = source;`, Options: deny("bar")},

			// JSDoc syntax is comment-only upstream, including declarations that tsgo
			// synthesizes into a script's global symbol table.
			{Code: `/** @type {Foo} */ let x;`, FileName: "jsdoc-type.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Foo")},
			{Code: `/** @typedef {number} Number */ Number;`, FileName: "jsdoc-global.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Number")},
			{Code: `/** @param {Foo} value */ function f(value) {}`, FileName: "jsdoc-param.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Foo")},
			{Code: `/** @template Foo */ function f(x) {}`, FileName: "jsdoc-template.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Foo")},
			{Code: `/** @typedef {Foo.Bar} Baz */ let x;`, FileName: "jsdoc-qualified.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Foo", "Bar", "Baz")},
			{Code: `/** @typedef {number} Array */ Array;`, FileName: "jsdoc-array.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Array")},
			{Code: `/** @typedef {number} Promise */ function f() { Promise; }`, FileName: "jsdoc-promise.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Promise")},
			{Code: `/** @import {Array} from "x" */ Array;`, FileName: "jsdoc-import.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Array")},

			// CommonJS wrapper globals have no authored definitions upstream.
			{Code: `exports.foo = 1;`, FileName: "wrapper.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("exports")},
			{Code: `module.foo; require.resolve; global.foo;`, FileName: "wrapper-globals.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("module", "require", "global")},

			// Parentheses around a dynamic import's options object, or around a nested
			// attribute value, still leave their keys import attributes.
			{Code: `import('x', ({ type: 'json' }));`, Options: deny("type")},
			{Code: `import('x', { with: ({ type: 'json' }) });`, Options: deny("type")},

			// ---- Dimension 4: JSX name positions, which upstream parses as JSXIdentifier
			// nodes that this rule's Identifier listener never visits ----
			{Code: `const x = <Foo.Bar />;`, Options: deny("Foo", "Bar"), Tsx: true},
			{Code: `const x = <svg xlink:href="a" />;`, Options: deny("xlink", "href", "svg"), Tsx: true},
			{Code: `const x = <a:b />;`, Options: deny("a", "b"), Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses are transparent only while they directly wrap a call/new
			// child. A member or TypeScript assertion between the identifier and the
			// call remains visible upstream; dynamic import is not a CallExpression.
			{Code: `(foo).bar();`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2)}},
			{Code: `call((foo).bar);`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 7)}},
			{Code: `import(((foo)));`, Options: deny("foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 10)}},
			{Code: `call((bar as any));`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 7)}},

			// A non-null wrapper around the pattern/member keeps a property read-only;
			// one inside the receiver does not stop the member itself being a target.
			{Code: `[foo.bar]! = source;`, Options: deny("foo", "bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2)}},
			{Code: `({x: foo.bar})! = source;`, Options: deny("foo", "bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 6)}},
			{Code: `[foo!.bar] = source;`, Options: deny("foo", "bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2), restricted("bar", 1, 7)}},
			{Code: `[foo.bar!] = source;`, Options: deny("foo", "bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 2)}},

			// Filtering JSDoc syntax must not hide the real declaration that follows
			// it, nor an authored declaration carrying a JSDoc tag.
			{Code: `/** @type {Foo} */ let Foo;`, FileName: "jsdoc-authored.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 24)}},
			{Code: `/** @enum {number} */ const Number = {x: 1}; Number;`, FileName: "jsdoc-enum.js", TSConfig: "tsconfig.allow-js.json", Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 29), restricted("Number", 1, 46)}},

			// `__filename` and `__dirname` are not globals in ESLint's CommonJS
			// environment, while authored declarations still shadow its real globals.
			{Code: `__filename; __dirname;`, FileName: "wrapper-paths.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("__filename", "__dirname"), Errors: []rule_tester.InvalidTestCaseError{restricted("__filename", 1, 1), restricted("__dirname", 1, 13)}},
			{Code: `arguments;`, FileName: "wrapper-arguments.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("arguments"), Errors: []rule_tester.InvalidTestCaseError{restricted("arguments", 1, 1)}},
			{Code: `const module = {}; module.foo;`, FileName: "authored-module.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("module"), Errors: []rule_tester.InvalidTestCaseError{restricted("module", 1, 7), restricted("module", 1, 20)}},

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

			// A literal that only feeds a member access is read as a value, so its own
			// members are checked as an object literal's rather than a pattern's.
			{Code: `[{ a: 1 }.b] = c;`, Options: deny("a"), Errors: []rule_tester.InvalidTestCaseError{restricted("a", 1, 4)}},
			{Code: `[{ a: obj.b }.c] = d;`, Options: deny("a", "b"), Errors: []rule_tester.InvalidTestCaseError{restricted("a", 1, 4)}},
			{Code: `({ a: { b: 1 }.c } = d);`, Options: deny("b"), Errors: []rule_tester.InvalidTestCaseError{restricted("b", 1, 9)}},
			{Code: `[{ Number }.x] = y;`, Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 4)}},

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
			// The nested-key recursion reaches the computed-key guard: a key
			// spelled `[w]` names nothing, so what it holds is not an attribute.
			{Code: `import('x', { [w]: { type: 1 } });`, Options: deny("type", "w"), Errors: []rule_tester.InvalidTestCaseError{restricted("w", 1, 16), restricted("type", 1, 22)}},

			// Locks in upstream's two roles for a shorthand property, which tsgo spells with
			// a single node: in an object literal the key half is checked and resolves to
			// nothing, in a pattern the key half is skipped and the reference half decides.
			{Code: `var bar = 1; x = { bar };`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5), restricted("bar", 1, 20)}},
			{Code: `var bar = 1; ({bar} = x);`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 5), restricted("bar", 1, 16)}},
			{Code: `x = { Number };`, Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 7)}},

			// Locks in upstream isReferenceToGlobalVariable() for a PrivateIdentifier,
			// which is never a variable reference however its parent reads — `#Number`
			// stays reported where the plain name would be excused as a global.
			{Code: `class C { #x; m(o) { return #x in o; } }`, Options: deny("x"), Errors: []rule_tester.InvalidTestCaseError{restrictedPrivate("x", 1, 11), restrictedPrivate("x", 1, 29)}},
			{Code: `class C { #Number; m(o) { return #Number in o; } }`, Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restrictedPrivate("Number", 1, 11), restrictedPrivate("Number", 1, 34)}},

			// A denied name is matched against the private name without its `#`, so the
			// two spellings form separate equivalence classes and never pair up.
			{Code: `class C { bar; #bar; }`, Options: deny("bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 11), restrictedPrivate("bar", 1, 16)}},

			// In a script every top-level declaration shares the global scope, so one of
			// them claims the name for the whole file even from another declaration
			// space; a nested one, and a module's top level, reach only their own scope.
			{Code: `const exports = {}; exports.foo = 1;`, FileName: "authored.cjs", TSConfig: "tsconfig.allow-js.json", Options: deny("exports"), Errors: []rule_tester.InvalidTestCaseError{restricted("exports", 1, 7), restricted("exports", 1, 21)}},
			{Code: "interface Number { q: string }\nNumber;", Options: deny("Number"), LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 11), restricted("Number", 2, 1)}},
			{Code: "const Number = 1;\nlet x: Number;", Options: deny("Number"), LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 7), restricted("Number", 2, 8)}},
			{Code: "namespace Number { export type A = 1; }\nNumber;", Options: deny("Number"), LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 11), restricted("Number", 2, 1)}},
			{Code: "function f() { interface Number { q: string } }\nNumber;", Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 26)}},
			{Code: "declare global { interface Number { q: string } }\nNumber;", Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 1, 28)}},
			{Code: "import {} from 'x';\ninterface Number { q: string }\nNumber;", Options: deny("Number"), Errors: []rule_tester.InvalidTestCaseError{restricted("Number", 2, 11)}},
			{
				Code:            "interface Promise {} Promise;",
				Options:         deny("Promise"),
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{restricted("Promise", 1, 11)},
			},
			{
				Code:            "namespace Number {} Number;",
				Options:         deny("Number"),
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{restricted("Number", 1, 11), restricted("Number", 1, 21)},
			},
			{
				Code:            "namespace Promise.Inner {} Promise;",
				Options:         deny("Promise"),
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{restricted("Promise", 1, 11)},
			},
			{
				Code:     "/** @import { Promise } from \"x\" */\nimport { Promise } from \"y\";\nPromise;",
				FileName: "jsdoc-and-authored-import.js",
				TSConfig: "tsconfig.allow-js.json",
				Options:  deny("Promise"),
				Errors:   []rule_tester.InvalidTestCaseError{restricted("Promise", 2, 10), restricted("Promise", 3, 1)},
			},

			// Locks in upstream isReferenceToGlobalVariable() on a qualified type name: the
			// right side names a member and is always checked, while the left side is a
			// reference and follows the global rule.
			{Code: `let x: Foo.Bar;`, Options: deny("Foo", "Bar"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 8), restricted("Bar", 1, 12)}},
			{Code: `let a: Array<Foo>;`, Options: deny("Array", "Foo"), Errors: []rule_tester.InvalidTestCaseError{restricted("Foo", 1, 14)}},

			// ---- Dimension 4: an annotated variable declarator or parameter is
			// reported over the name alone, where ESLint ends the range after the
			// annotation. A class field, an interface member and a type-literal
			// member carry an annotation too and do not differ ----
			{Code: `let x: Foo;`, Options: deny("x"), Errors: []rule_tester.InvalidTestCaseError{restricted("x", 1, 5)}},
			{Code: `function f(a: string) {}`, Options: deny("a"), Errors: []rule_tester.InvalidTestCaseError{restricted("a", 1, 12)}},
			{Code: `class C { data: string[] = []; }`, Options: deny("data"), Errors: []rule_tester.InvalidTestCaseError{restricted("data", 1, 11)}},
			{Code: `interface I { data: string }`, Options: deny("data"), Errors: []rule_tester.InvalidTestCaseError{restricted("data", 1, 15)}},

			// ---- Dimension 4: JSX name positions, which upstream parses as JSXIdentifier
			// nodes that this rule's Identifier listener never visits ----
			{Code: `const x = <div className={foo} />;`, Options: deny("div", "className", "foo"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("foo", 1, 27)}},
			{Code: `const x = <Foo>{bar}</Foo>;`, Options: deny("Foo", "bar"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 17)}},
			{Code: `const x = <Foo {...bar} />;`, Options: deny("Foo", "bar"), Tsx: true, Errors: []rule_tester.InvalidTestCaseError{restricted("bar", 1, 20)}},
		},
	)
}
