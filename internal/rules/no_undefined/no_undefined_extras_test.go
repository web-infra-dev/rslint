package no_undefined

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUndefinedExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing
// at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
func TestNoUndefinedExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUndefinedRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: receiver/expression wrappers (never skipped by
			// design — undefined stays a reference through every wrapper) ----
			// covered on the invalid side below; these are the *non*-flagged
			// TS-type-keyword counterparts, which never surface as Identifier
			// nodes at all because tsgo parses bare `undefined` in type
			// position as KindUndefinedKeyword, not KindIdentifier.
			{Code: `let x: undefined;`},
			{Code: `const y: any = 1; const z = y as undefined;`},
			{Code: `declare const w: any; const s = w satisfies undefined;`},
			{Code: `function f<T = undefined>() {}`},

			// ---- Dimension 4: access/key forms — non-computed member/key
			// names never bind or reference a variable, regardless of the
			// surrounding declaration kind. Locks in each
			// isNonBindingUndefinedPosition() switch arm individually. ----

			// Locks in isNonBindingUndefinedPosition() KindPropertySignature arm
			{Code: `interface I { undefined: string }`},
			// Locks in isNonBindingUndefinedPosition() KindMethodSignature arm
			{Code: `interface I { undefined(): void }`},
			// Locks in isNonBindingUndefinedPosition() KindPropertyDeclaration arm
			{Code: `class C { undefined = 1; }`},
			{Code: `class C { static undefined = 1; }`},
			// Locks in isNonBindingUndefinedPosition() KindGetAccessor arm
			{Code: `class C { get undefined() { return 1; } }`},
			// Locks in isNonBindingUndefinedPosition() KindSetAccessor arm
			{Code: `class C { set undefined(v) {} }`},
			// Locks in isNonBindingUndefinedPosition() KindNamedTupleMember arm
			{Code: `type T = [undefined: string];`},
			// Locks in isNonBindingUndefinedPosition() KindJsxAttribute arm
			{Code: `const el = <div undefined="x" />;`, Tsx: true},
			// PrivateIdentifier is a distinct Kind from Identifier, so `#undefined`
			// never reaches the ast.KindIdentifier listener at all.
			{Code: `class C { #undefined = 1; get() { return this.#undefined; } }`},

			// ---- Dimension 4: access/key forms — abstract/declare member
			// names stay non-computed keys even without a body ----
			{Code: `abstract class C { abstract undefined(): void; }`},

			// ---- Dimension 4: declaration/container forms ----

			// Locks in isNonBindingUndefinedPosition() KindBindingElement arm
			// (PropertyName side — renamed destructuring key)
			{Code: `var { undefined: foo } = bar;`},
			// Locks in isNonBindingUndefinedPosition() KindImportSpecifier arm
			// (PropertyName side — aliased *from* undefined)
			{Code: "import { undefined as localName } from 'foo';\nconsole.log(localName);"},

			// ---- Dimension 4: nesting/traversal boundaries — optional chain
			// member access never binds, only the call argument does ----
			{Code: `declare const a: any; a?.undefined;`},

			// ---- Dimension 4: graceful degradation — import attributes,
			// labels, qualified type names, and namespace re-exports are pure
			// syntax and never variable positions ----

			// Locks in isNonBindingUndefinedPosition() KindImportAttribute arm
			{Code: `import x from 'y' with { undefined: 'z' };`},
			// Locks in isNonBindingUndefinedPosition() KindLabeledStatement /
			// KindBreakStatement / KindContinueStatement arms
			{Code: `undefined: for(;;) { break undefined; }`},
			// Locks in isNonBindingUndefinedPosition() KindQualifiedName arm
			{Code: `declare namespace A {} let y: A.undefined;`},
			// Locks in isNonBindingUndefinedPosition() KindNamespaceExport arm
			{Code: `export * as undefined from 'foo';`},
			// Locks in isImportTypeSyntax() qualifier branch
			{Code: `type T = import('foo').undefined;`},
			// Locks in isInJSDocSyntax(): a JSDoc comment's synthetic AST
			// subtree is invisible to this rule, matching upstream (which
			// never sees JSDoc content as executable syntax at all).
			{Code: "/** @param {undefined} x */\nfunction f(x) {}"},

			// ---- Dimension 4: declaration/container forms — export
			// specifier name-slot classification, independent of the
			// type-only keyword ----

			// Local, aliased export: the exported label is a pure name, not a
			// reference to the local binding named "undefined".
			{Code: "const a = 1;\nexport { a as undefined };"},
			{Code: "const undefined2 = 1;\nexport { undefined2 as undefined };"},
			// Ambient module declared with a string literal name: the name
			// node is a StringLiteral, never an Identifier, so it can never
			// reach this rule's listener regardless of its text.
			{Code: `module "undefined" {}`},

			// ---- Real-user: common default-parameter and config-object
			// shapes that use `undefined` as an explicit sentinel value ----
			{Code: "function fetchData(callback) { return callback; }\nfetchData(function done(result) { return result; });"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver/expression wrappers ----
			{
				Code:   `(undefined)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 2}},
			},
			{
				Code:   `((undefined))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 3}},
			},
			{
				Code:   `undefined!;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1}},
			},
			{
				Code:   `const a: any = 1; const b = a as undefined ? undefined : 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 46}},
			},
			{
				Code:   `declare const a: any; a?.(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 27}},
			},
			{
				Code:   `typeof undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 8}},
			},
			{
				Code:   `typeof (undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 9}},
			},

			// ---- Dimension 4: access/key forms — computed keys are always
			// value expressions, regardless of container ----
			{
				Code:   `class Foo { [undefined] = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 14, EndLine: 1, EndColumn: 23}},
			},
			{
				Code:   `interface I { [undefined]: string; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},

			// ---- Dimension 4: declaration/container forms — generics on
			// every construct that carries a type-parameter list; each
			// locks in that TypeParameterDeclaration.Name() is a genuine
			// binding, unlike the surrounding container's own key/name ----
			{
				Code:   `class Foo<undefined> {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 11}},
			},
			{
				Code:   `interface Foo<undefined> {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 15}},
			},
			{
				Code:   `type Foo<undefined> = string;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 10}},
			},
			{
				Code:   `function f<undefined>(x: undefined) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 12}},
			},
			// Locks in the mapped-type key parameter binding as a
			// TypeParameterDeclaration name, same as any other generic.
			{
				Code:   `type M = { [undefined in string]: number };`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 13}},
			},

			// ---- Dimension 4: declaration/container forms — enum members
			// are genuine value bindings, unlike interface/class members ----
			{
				Code:   `enum E { undefined }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 10}},
			},
			{
				Code:   `enum E { undefined = 1 }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 10}},
			},
			{
				Code:   `const enum E { undefined }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},

			// ---- Dimension 4: declaration/container forms — TypeScript
			// namespace/module, ambient, and import-equals declarations
			// create real bindings just like their JS counterparts ----
			{
				Code:   `type undefined = string; const x: undefined = "a";`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 6}},
			},
			{
				Code:   `namespace undefined { export const a = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 11}},
			},
			{
				Code:   `declare var undefined: any;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 13}},
			},
			{
				Code:   `declare module "m" { const undefined: any; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 28}},
			},
			{
				Code:   `import undefined = require('foo');`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 8}},
			},
			{
				Code:   `export as namespace undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 21}},
			},

			// ---- Dimension 4: declaration/container forms — TS import-type
			// and type-only import/export specifiers still create/reference
			// real local bindings ----
			{
				Code:   `import type { undefined } from 'foo';`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 15}},
			},
			{
				Code: "type undefined = string;\nexport type { undefined };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 6},
					{MessageId: "unexpectedUndefined", Line: 2, Column: 15},
				},
			},

			// ---- Dimension 4: declaration/container forms — export
			// specifier's local-referring slot, independent of aliasing.
			// Locks in the isNonBindingExportSpecifierName() PropertyName
			// branch (aliased) and the no-PropertyName branch (shorthand). ----
			{
				Code: "const undefined = 1;\nexport { undefined as a };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 7},
					{MessageId: "unexpectedUndefined", Line: 2, Column: 10},
				},
			},
			{
				Code: "const undefined = 1;\nexport { undefined };",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 7},
					{MessageId: "unexpectedUndefined", Line: 2, Column: 10},
				},
			},

			// ---- Dimension 4: declaration/container forms — function
			// overload signatures each carry their own real parameter
			// binding, even though the signature has no body ----
			{
				Code: `function f(undefined: number): void; function f(undefined: number) { return undefined; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 12},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 49},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 77},
				},
			},
			// Constructor parameter property: still a real Parameter binding.
			{
				Code:   `class Foo { constructor(public undefined: number) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 32}},
			},
			// A decorator expression is a genuine value reference.
			{
				Code: "function undefined() {}\n@undefined\nclass Foo {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 10},
					{MessageId: "unexpectedUndefined", Line: 2, Column: 2},
				},
			},

			// ---- Dimension 4: nesting/traversal boundaries — shadowing at
			// any depth still resolves to *some* variable literally named
			// "undefined", so every occurrence is independently reportable
			// regardless of which scope declares it. ----
			{
				Code: `function outer(undefined) { function inner() { return undefined; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 16},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 55},
				},
			},
			{
				Code: `let undefined = 1; { let undefined = 2; console.log(undefined); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 5},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 26},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 53},
				},
			},
			{
				Code:   `try {} catch { try {} catch(undefined) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 29}},
			},
			{
				Code: `const f = (undefined = 1) => undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 12},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 30},
				},
			},
			{
				Code:   `class Foo { static { console.log(undefined); } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 34}},
			},

			// ---- Dimension 4: graceful degradation — rest/spread positions
			// in both object and array literals/patterns are value
			// expressions or real bindings, never pure syntax ----
			{
				Code: `const { a: undefined } = obj; console.log(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 12},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 43},
				},
			},
			{
				Code:   `for (const undefined of []) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 12}},
			},
			{
				Code: `const {...undefined} = obj; console.log(undefined);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUndefined", Line: 1, Column: 11},
					{MessageId: "unexpectedUndefined", Line: 1, Column: 41},
				},
			},
			{
				Code:   `const o = {...undefined};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 15}},
			},
			{
				Code:   `export default undefined;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 16}},
			},
			{
				Code:   `const [a = undefined] = b;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 12}},
			},
			{
				Code:   "tag`${undefined}`;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 7}},
			},

			// ---- Position assertions (Line/Column/EndLine/EndColumn) across
			// containers, including a multi-line case ----
			{
				Code: `const obj = {
  a: 1,
  b: undefined,
};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 3, Column: 6, EndLine: 3, EndColumn: 15}},
			},
			{
				Code:   `undefined`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}},
			},

			// ---- Intentional divergence from ESLint: a scope-manager
			// artifact where a named class declaration's self-referencing
			// inner scope duplicates the outer declaration's variable,
			// causing upstream to report the same identifier twice. This
			// rule reports each occurrence once. See no_undefined.md
			// "Differences from ESLint". ----
			{
				Code:   `class undefined {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 7}},
			},
			{
				Code:   `const C = class undefined {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 1, Column: 17}},
			},
			// Same divergence category: an assignment pattern's shorthand
			// property with a default value. Upstream's scope manager
			// reports this identifier twice (once as a definition, once as
			// an implicit reference produced by the default-value
			// initializer); this rule reports it once.
			{
				Code:   "declare const foo: any;\n({ undefined = 1 } = foo);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedUndefined", Line: 2, Column: 4}},
			},
		},
	)
}
