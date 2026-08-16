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
		},
	)
}
