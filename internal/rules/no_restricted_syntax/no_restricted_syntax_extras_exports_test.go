package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Regression expectations were checked against ESLint 10.10.0, esquery 1.7.0,
// and @typescript-eslint/parser 8.69.0. tsgo folds `export` into the exported
// declaration's modifiers; ESTree wraps it in ExportNamedDeclaration.
func TestNoRestrictedSyntaxExportRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "export const a = 1;", Options: []any{"ExportNamedDeclaration[exportKind='type']"}},
			{Code: "export * from 'x';", Options: []any{"ExportNamedDeclaration"}},
			{Code: "export { z } from 'x';", Options: []any{"ExportAllDeclaration"}},
			{Code: "export const x = 1; sample();", Options: []any{"Program:not(:has(ExportNamedDeclaration, ExportDefaultDeclaration, ExportAllDeclaration)) > ExpressionStatement > CallExpression[callee.name=/^(sample|split)$/]"}},
			{Code: "export const a = 1;", Options: []any{"Program > VariableDeclaration"}},
			{Code: "export default function f() {}", Options: []any{"ExportNamedDeclaration"}},
			{Code: "export const a = 1;", Options: []any{":not(ExportNamedDeclaration) > VariableDeclaration"}},
			{Code: "export const a = 1;", Options: []any{"ExportNamedDeclaration > VariableDeclaration ~ *"}},
			{Code: "if (x) { sample(); }\nexport const a = 1;", Options: []any{"Program:not(:has(ExportNamedDeclaration)) CallExpression"}},
			{Code: "namespace A.B {}", Options: []any{"ExportNamedDeclaration"}},
			{Code: "namespace A.B.C {}", Options: []any{"ExportNamedDeclaration"}},
			{Code: "declare module A.B {}", Options: []any{"ExportNamedDeclaration"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export function f() {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "export class C {}",
				Options: []any{"ExportNamedDeclaration > ClassDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export interface I {}",
				Options: []any{"ExportNamedDeclaration[exportKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    "export type T = 1;",
				Options: []any{"ExportNamedDeclaration[exportKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export declare const a: number;",
				Options: []any{"ExportNamedDeclaration[exportKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export enum E { A }",
				Options: []any{"ExportNamedDeclaration > TSEnumDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "const y = 1; export { y };",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "export * from 'x';",
				Options: []any{"ExportAllDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export * as ns from 'x';",
				Options: []any{"ExportAllDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "export { z } from 'x';",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "sample();",
				Options: []any{"Program:not(:has(ExportNamedDeclaration, ExportDefaultDeclaration, ExportAllDeclaration)) > ExpressionStatement > CallExpression[callee.name=/^(sample|split)$/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration > VariableDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1; export const b = 2;",
				Options: []any{"ExportNamedDeclaration + ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 21, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code:    "const a = 1; export const b = 2;",
				Options: []any{"VariableDeclaration ~ ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration[declaration.type='VariableDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration[specifiers.length=0]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration[source=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "declare module 'm' { export const a: number; }",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"Program > ExportNamedDeclaration:first-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"Program:has(> ExportNamedDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration VariableDeclarator"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export class C { m() {} }",
				Options: []any{"ExportNamedDeclaration MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"VariableDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"[type=/^Export/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration > .declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{":matches(ExportNamedDeclaration, VariableDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export import A = B.C;",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "export abstract class C {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    "export class C {}",
				Options: []any{":matches(ExportNamedDeclaration, ClassDeclaration, ClassBody)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export class C {}",
				Options: []any{"[type='ExportNamedDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration[parent.type='Program']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"VariableDeclaration[parent.type='ExportNamedDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration > VariableDeclaration[kind='const']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "const a = 1; export const b = 2;",
				Options: []any{"Program > :nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "const a = 1; export const b = 2;",
				Options: []any{"Program > :last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:    "const a = 1;\nexport { a as default };",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 25},
				},
			},
			{
				Code:    "export type { T } from 'x';",
				Options: []any{"ExportNamedDeclaration[exportKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "export * from 'x';",
				Options: []any{"ExportAllDeclaration[exported=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "export * as ns from 'x';",
				Options: []any{"ExportAllDeclaration[exported.name='ns']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    "declare module 'm' { export const a: number; }",
				Options: []any{"Program:has(ExportNamedDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 47},
				},
			},
			{
				Code:    "export function f() { return 1; }",
				Options: []any{"ExportNamedDeclaration ReturnStatement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 23, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{":statement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{":declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export async function f() {}",
				Options: []any{"ExportNamedDeclaration > FunctionDeclaration[async=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"ExportNamedDeclaration.body"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;",
				Options: []any{"Program > .body"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "export namespace A.B {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "namespace A.B {} sample();",
				Options: []any{"Program:not(:has(ExportNamedDeclaration)) > ExpressionStatement > CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 18, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "namespace A { export namespace B {} }",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code:    "@dec export class C {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 6, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "@dec export abstract class C {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 6, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "@dec export class C {}",
				Options: []any{":matches(ExportNamedDeclaration, ClassDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 6, EndLine: 1, EndColumn: 23},
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "export @dec class C {}",
				Options: []any{":matches(ExportNamedDeclaration, ClassDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "@dec export class C {}",
				Options: []any{"ClassDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 13, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "@dec export class C {}",
				Options: []any{"ExportNamedDeclaration > ClassDeclaration > Decorator"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code:    "@dec\nexport class C {}",
				Options: []any{"ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 18},
				},
			},
		})
}
