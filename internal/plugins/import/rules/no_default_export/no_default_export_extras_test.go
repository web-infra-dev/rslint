package no_default_export_test

import (
	"errors"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoDefaultExportExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestNoDefaultExportExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_default_export.NoDefaultExportRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration forms, named function export is allowed ----
			{Code: `export function foo() {}`},
			// ---- Dimension 4: declaration forms, named class export is allowed ----
			{Code: `export class Foo {}`},
			// ---- Dimension 4: declaration forms, named interface/type exports are allowed ----
			{Code: `export interface Foo {}`},
			{Code: `export type Foo = string`},
			{Code: `export enum Foo { A }`},
			// ---- Dimension 4: access/key forms, non-default string export name is allowed ----
			{Code: `const foo = 1; export { foo as "foo-bar" }`},
			// ---- Dimension 4: access/key forms, re-exported default under a named export is allowed ----
			{Code: `export { default as foo } from "./foo"`},
			// ---- Dimension 4: access/key forms, namespace and star exports are allowed ----
			{Code: `export * from "./foo"`},
			{Code: `export * as ns from "./foo"`},
			// ---- Dimension 4: receiver/expression wrappers N/A, export declarations have no receiver ----
			// ---- Dimension 4: computed keys N/A, module export names are identifier/string literal only ----
			// ---- Dimension 4: nesting/traversal boundaries, named exports inside ambient module are allowed ----
			{Code: `declare module "pkg" { export const foo: string }`},
			{Code: `namespace N { export const foo = 1 }`},
			{Code: `declare module "pkg" { export { foo } from "foo" }`},
			// ---- Dimension 4: graceful degradation, empty named export has no default specifier ----
			{Code: `export {}`},
			// ---- Dimension 4: graceful degradation, export equals is not a default export ----
			{Code: `const foo = 1; export = foo`},
			// Locks in upstream create() arm 1: CommonJS assignments are ignored.
			{Code: `module.exports = { foo: 1 }`},
			// Locks in upstream ExportNamedDeclaration arm 1: exported name is not default.
			{Code: `const foo = 1; export { foo as bar }`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver/expression wrappers, parenthesized default expression still reports ----
			{
				Code: `const foo = 1; export default (foo)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 4: receiver/expression wrappers, TS non-null expression still reports ----
			{
				Code: `const foo = 1; export default foo!`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 4: receiver/expression wrappers, TS assertion expression still reports ----
			{
				Code: `const foo = 1; export default foo as string`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 30},
				},
			},
			// ---- Dimension 4: receiver/expression wrappers, TS satisfies expression still reports ----
			{
				Code: `const foo = "x"; export default foo satisfies string`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- Dimension 4: declaration forms, default interface export reports ----
			{
				Code: `export default interface Foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 4: declaration forms, default abstract class export reports ----
			{
				Code: `export default abstract class Foo {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 4: declaration forms, async generator default function reports ----
			{
				Code: `export default async function* foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Dimension 4: declaration forms, multiline default export reports at default ----
			{
				Code: "export\n  default class Foo {}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 2, Column: 3, EndLine: 2, EndColumn: 10},
				},
			},
			// ---- Dimension 4: declaration forms, leading comments before export are ignored for report location ----
			{
				Code: `/* leading */ export default function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: declaration forms, comments between export and default are transparent ----
			{
				Code: `export /* comment */ default function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: access/key forms, string-literal default export name reports ----
			{
				Code: `const foo = 1; export { foo as "default" }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- Dimension 4: access/key forms, string-literal local name is preserved in alias message ----
			{
				Code: `export { "foo" as default } from "foo"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasUndefinedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Dimension 4: access/key forms, string-literal default shorthand matches upstream local.name output ----
			{
				Code: `export { "default" } from "foo"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasUndefinedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Dimension 4: access/key forms, comments after export are ignored for named export location ----
			{
				Code: "const foo = 1;\nexport\n/* comment */\n{ foo as default }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},
			// ---- Dimension 4: access/key forms, shorthand default re-export reports ----
			// ---- Real-user: import-js/eslint-plugin-import#3209 default re-exports are reported upstream ----
			{
				Code: `export { default } from "foo"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noAliasDefault",
						Message:   noAliasDefaultDefaultMessage,
						Line:      1,
						Column:    8,
						EndLine:   1,
						EndColumn: 9,
					},
				},
			},
			// ---- Dimension 4: access/key forms, type-only default alias reports like upstream's generic specifier walk ----
			{
				Code: `type Foo = string; export type { Foo as default }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noAliasDefault",
						Message:   "Do not alias `Foo` as `default`. Just export `Foo` itself instead.",
						Line:      1,
						Column:    27,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, ambient module default export still reports ----
			{
				Code: `declare module "pkg" { export default Foo }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 31, EndLine: 1, EndColumn: 38},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, namespace default export reports ----
			{
				Code: `namespace N { export default Foo }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, nested namespace default export reports ----
			{
				Code: `namespace A { namespace B { export default Foo } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, ambient module default re-export still reports ----
			{
				Code: `declare module "pkg" { export { default } from "foo" }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noAliasDefault",
						Message:   noAliasDefaultDefaultMessage,
						Line:      1,
						Column:    31,
						EndLine:   1,
						EndColumn: 32,
					},
				},
			},
			// ---- Dimension 4: nesting/traversal boundaries, nested namespace alias default reports ----
			{
				Code: `namespace A { namespace B { export { foo as default } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 37},
				},
			},
			// Locks in upstream ExportDefaultDeclaration arm: report direct default exports.
			// ---- Real-user: import-js/eslint-plugin-import#889 proposed direct default export bad case ----
			{
				Code: `export default function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "preferNamed", Message: preferNamedMessage, Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// Locks in upstream ExportNamedDeclaration arm 2: ExportSpecifier exported as default uses alias message.
			// ---- Real-user: import-js/eslint-plugin-import#889 proposed alias default bad case ----
			{
				Code: `function foo() {}; export {foo as default}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
				},
			},
			// Locks in upstream ExportNamedDeclaration filter: only exported default names are reported.
			{
				Code: `let foo, bar; export { foo as default, bar }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAliasDefault", Message: noAliasDefaultMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
		},
	)
}

func TestNoDefaultExportSkippedBabelReExportSyntaxIsNotParsedByTsgo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
	}{
		{name: "default re-export", code: `export foo from "foo.js"`},
		{name: "default plus named re-export", code: `export Memory, { MemoryValue } from './Memory'`},
		{name: "default from shorthand", code: `export default from "foo.js"`},
		{name: "default enum export", code: `export default enum Foo { A }`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootDir := fixtures.GetRootDir()
			fileName := "file.ts"
			fs := rslint_utils.NewOverlayVFSForFile(tspath.ResolvePath(rootDir, fileName), tc.code)
			host := rslint_utils.CreateCompilerHost(rootDir, fs)
			_, err := rslint_utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
			if err == nil {
				t.Fatalf("expected tsgo parse error for %q", tc.code)
			}
			var syntacticError *rslint_utils.SyntacticError
			if !errors.As(err, &syntacticError) {
				t.Fatalf("expected *utils.SyntacticError for %q, got %T: %v", tc.code, err, err)
			}
		})
	}
}
