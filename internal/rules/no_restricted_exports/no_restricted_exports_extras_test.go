package no_restricted_exports

// TestNoRestrictedExportsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in. Upstream-migrated cases live in
// no_restricted_exports_upstream_test.go.

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRestrictedExportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedExportsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 1: TS-only declaration types are not checked by a
			// named export declaration (upstream's declaration.type switch only
			// matches FunctionDeclaration/ClassDeclaration/VariableDeclaration;
			// everything else silently falls through unchecked) ----
			{Code: `export interface Foo {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}}},
			{Code: `export type Foo = string;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}}},
			{Code: `export enum Foo { A }`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}}},

			// ---- Dimension 4: graceful degradation — empty destructuring
			// pattern binds nothing; must not crash or spuriously report ----
			{Code: `export const {} = {};`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export const [] = [];`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- Real-user: https://github.com/eslint/eslint/issues/15384 —
			// TypeScript's forgiving parser accepts `export class extends Base
			// {}` (a named, non-default export with no identifier), which
			// crashes upstream's `checkExportedName` on a null node. Must not
			// panic and must not report. ----
			{Code: `class Example {}
export class extends Example {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Example"}}}},

			// ---- Dimension 1: a function declaration with no body is a
			// TSDeclareFunction upstream, a type the declaration.type switch
			// does not match — an overload signature, a `declare function` and
			// every function in a `.d.ts` alike go unchecked, on both the set
			// and the pattern arm ----
			{Code: `export declare function bar(): void;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"bar"}}}},
			{Code: `export function foo();`, Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}}},
			{Code: `export declare function fooBar(): void;`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "Bar$"}}},

			// ---- Dimension 1: `export default interface Foo {}` is a default
			// export, so the named restriction never reaches its name ----
			{Code: `export default interface Foo {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}}},

			// ---- Dimension 1: the remaining TS-only declaration kinds a named
			// export can carry, none of them a type the declaration.type switch
			// matches ----
			{Code: `export const enum Foo { A }`, Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}}},
			{Code: `export declare namespace N {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"N"}}}},
			{Code: `export declare module N {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"N"}}}},

			// ---- Dimension 4: an export clause with no specifiers exports no
			// name to check ----
			{Code: `export {} from 'mod';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- checkSpecifierName: which restrictDefaultExports property
			// governs a `default` specifier depends on its local name and on
			// whether a source is specified, so the other properties leave it
			// alone ----
			{Code: `export { default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": true}}}},
			{Code: `export { foo as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}}},
			{Code: `export { default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"named": true}}}},
			{Code: `let foo; export { foo as default };`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": true}}}},
			{Code: `export * as default from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true, "named": true, "defaultFrom": true, "namedFrom": true}}}},

			// ---- parseOptions: an empty restrictDefaultExports object keeps
			// the rule running with every default branch off ----
			{Code: `export default foo;`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{}}}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: TS type-expression wrappers on the exported
			// expression must not affect the wrapping ExportAssignment /
			// declaration-level check for restrictDefaultExports.direct ----
			{
				Code:    `export default (1 as any);`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code:    `export default foo!;`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code:    `export default foo satisfies Bar;`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},

			// ---- Dimension 1: parenthesized function/class EXPRESSION default
			// exports (ExportAssignment) vs the DECLARATION form — upstream only
			// locks the declaration form (`export default function foo() {}`)
			// with direct:true; the expression form takes a different tsgo node
			// kind (ExportAssignment vs FunctionDeclaration) and must be
			// independently verified ----
			{
				Code:    `export default (function a() {});`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code:    `export default (class A {});`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},

			// ---- Dimension 2: TypeScript namespace-nested export — the rule's
			// listeners fire on AST kind alone with no module-level-only guard,
			// matching upstream's own naive (non-scope-aware) traversal ----
			{
				Code:    `namespace NS { export const a = 1; }`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 29},
				},
			},

			// ---- Dimension 4: graceful degradation — RestElement inside an
			// object binding pattern must still be checked like any other bound
			// name (CollectBindingNames doesn't special-case DotDotDotToken) ----
			{
				Code:    `export var { a, ...rest } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"rest"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("rest"), Line: 1, Column: 20},
				},
			},

			// ---- Dimension 1: `export type { Foo }` re-export specifiers go
			// through the same ExportSpecifier path as value exports — upstream
			// never checks specifier.exportKind, so type-only specifiers are
			// restricted identically to value specifiers ----
			{
				Code:    `interface Foo {} export type { Foo };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("Foo"), Line: 1, Column: 32},
				},
			},

			// ---- Locks in upstream checkExportedName() branch B: pattern-match
			// and set-match are independent OR arms. Upstream never configures
			// restrictedNamedExports and restrictedNamedExportsPattern in the
			// same test case, so the two branches firing on different names in
			// one file is otherwise unverified. ----
			{
				Code: `var a, xyz; export { a, xyz };`,
				Options: []any{map[string]any{
					"restrictedNamedExports":        []any{"a"},
					"restrictedNamedExportsPattern": "^xy",
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 22},
					{MessageId: "restrictedNamed", Message: namedMsg("xyz"), Line: 1, Column: 25},
				},
			},

			// ---- Real-user: https://github.com/eslint/eslint/issues/15617 —
			// `export { default, other } from 'mod'` (no "as") re-exports
			// "default" from another module just like `export { default as
			// default } from 'mod'`; closed as intended behavior upstream, so
			// the port must restrict it the same way. ----
			{
				Code:    `export { default, other } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 10},
				},
			},

			// ---- Upstream reads a statement's declared names out of scope
			// analysis, which knows one variable per name, so a name bound
			// twice by one statement is reported once, on its first binding ----
			{
				Code:    `export var a, a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Position assertions: full Line/Column/EndLine/EndColumn range
			// on the VariableStatement declared-name container ----
			{
				Code:    `export var a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},

			// ---- Position assertions: full range on the ExportSpecifier
			// container ----
			{
				Code:    `var a; export { a as b };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},

			// ---- Dimension 1: only the implementation of an overload set
			// carries a body, so one report lands on the implementation's name
			// rather than one per signature on the same exported name ----
			{
				Code: `export function foo(a: string): void;
export function foo(a: number): void;
export function foo(a: any) {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 3, Column: 17, EndLine: 3, EndColumn: 20},
				},
			},

			// ---- Dimension 1: a class keeps its ClassDeclaration type when
			// declared, so the bodiless-function skip must not reach it ----
			{
				Code:    `export declare class Foo {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("Foo"), Line: 1, Column: 22, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 1: an interface is the third declaration kind
			// TypeScript lets carry the default modifier ----
			{
				Code:    `export default interface Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},

			// ---- Dimension 1: the default path reports what is declared
			// without looking inside it, so a bodiless declaration is reported
			// there even though the named path skips one ----
			{
				Code:    `export default function foo(): void;`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},

			// ---- Position assertions: TSESTree folds a declarator's type
			// annotation, and its definite-assignment token, into the range of
			// the Identifier upstream reports on, so the reported range runs
			// past the name to the end of the annotation ----
			{
				Code:    `export const foo: number = 1;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `export declare const foo: (a: number) => void;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 22, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    `export let foo!: number;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 12, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: `export const foo: {
  a: number;
} = { a: 1 };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 14, EndLine: 3, EndColumn: 2},
				},
			},
			{
				Code:    `export const foo: number = 1, bar: string = '';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo", "bar"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
					{MessageId: "restrictedNamed", Message: namedMsg("bar"), Line: 1, Column: 31, EndLine: 1, EndColumn: 42},
				},
			},

			// ---- Position assertions: a binding pattern carries the
			// annotation itself and upstream reports the bound identifiers
			// inside it, which carry none, so the range stays on the name ----
			{
				Code:    `export const { foo }: { foo: number } = { foo: 1 };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `export const [foo]: number[] = [1];`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 15, EndLine: 1, EndColumn: 18},
				},
			},

			// ---- Position assertions: a decorator written ahead of `export`
			// belongs to the class it decorates in TSESTree, leaving the
			// ExportDefaultDeclaration upstream reports to start at the
			// `export` keyword, while tsgo keeps it inside the declaration ----
			{
				Code:    `@dec export default class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 6, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    `@dec1 @dec2 export default class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 13, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    `@dec /* c */ export default class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 14, EndLine: 1, EndColumn: 41},
				},
			},

			// ---- Position assertions: a decorator that follows `default`
			// already sits inside the range both sides report ----
			{
				Code:    `export default @dec class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},

			// ---- Position assertions: full range on the default-direct
			// declaration container, multi-line ----
			{
				Code: `export default class Foo {
  bar() {}
}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 3, EndColumn: 2},
				},
			},

			// ---- Position assertions: leading trivia stays out of the range,
			// and a comment between the modifiers stays in it ----
			{
				Code:    `/** doc */ export default class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 12, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    `export /*c*/ default /*c*/ class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: `export
default
class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 3, EndColumn: 13},
				},
			},

			// ---- Dimension 1: the remaining default-export shapes — an
			// anonymous class, an abstract one, and an expression that is not
			// a function or class ----
			{
				Code:    `export default class {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `export default abstract class Foo {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    `export default (foo, bar);`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},

			// ---- Position assertions: type parameters hang off the
			// declaration rather than off the name upstream reports, and a
			// decorator leaves a named export's name where it is ----
			{
				Code:    `export class Foo<T> {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("Foo"), Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `export function foo<T>(): void {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 17, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `@dec export class Foo {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"Foo"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("Foo"), Line: 1, Column: 19, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- Only names bound by the same statement collapse, and a name
			// that follows a collapsed one keeps its own report ----
			{
				Code: `export var a = 1;
export var a = 2;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 2, Column: 12, EndLine: 2, EndColumn: 13},
				},
			},
			{
				Code:    `export var a, b, a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
				},
			},

			// ---- Dimension 4: `using` and `await using` declarations reach
			// the same VariableStatement path as var/let/const ----
			{
				Code:    `export using a = res;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:    `export await using a = res;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- Position assertions: a nested binding pattern reports the
			// bound name wherever it sits, defaults and rest included ----
			{
				Code:    `export const { a: [{ b }] } = obj;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    `export const [a = 1, ...rest] = arr;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "rest"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15, EndLine: 1, EndColumn: 16},
					{MessageId: "restrictedNamed", Message: namedMsg("rest"), Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
				},
			},

			// ---- Position assertions: a column counts UTF-16 code units, so
			// an astral exported name spans two of them ----
			{
				Code:    `export { x as "👍" } from 'mod';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"👍"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("👍"), Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `export { "👍" as x, y } from 'mod';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"y"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("y"), Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},

			// ---- checkSpecifierName / checkExportAllName: a string-literal
			// `default` is the same exported name as the keyword one ----
			{
				Code:    `export { 'default' } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `export * as 'default' from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namespaceFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 13, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `let d; export { d as 'default' };`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"named": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},

			// ---- isRestrictedName: a name matched by both option arms is
			// still one report ----
			{
				Code: `export const foo = 1;`,
				Options: []any{map[string]any{
					"restrictedNamedExports":        []any{"foo"},
					"restrictedNamedExportsPattern": "^f",
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("foo"), Line: 1, Column: 14, EndLine: 1, EndColumn: 17},
				},
			},
		},
	)
}
