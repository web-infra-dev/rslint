package no_restricted_syntax

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Checked against ESLint 10.13.0 and @typescript-eslint/parser 8.65.0.
func TestNoRestrictedSyntaxExportSelectorForms(t *testing.T) {
	for _, declaration := range []struct {
		code     string
		nodeType string
	}{
		{"export default interface I {}", "ExportDefaultDeclaration"},
		{"export default function f() {}", "ExportDefaultDeclaration"},
		{"export default class C {}", "ExportDefaultDeclaration"},
		{"export default 1;", "ExportDefaultDeclaration"},
		{"export default value;", "ExportDefaultDeclaration"},
		{"export default () => {};", "ExportDefaultDeclaration"},
		{"export default (class {});", "ExportDefaultDeclaration"},
		{"export interface I {}", "ExportNamedDeclaration"},
		{"export type T = string;", "ExportNamedDeclaration"},
		{"export namespace N {}", "ExportNamedDeclaration"},
		{"export namespace A.B {}", "ExportNamedDeclaration"},
		{"export import A = B.C;", "ExportNamedDeclaration"},
		{`export import A = require("x");`, "ExportNamedDeclaration"},
		{`export import type A = require("x");`, "ExportNamedDeclaration"},
		{"export const value = 1;", "ExportNamedDeclaration"},
		{"export function f() {}", "ExportNamedDeclaration"},
		{"export class C {}", "ExportNamedDeclaration"},
		{"export enum E { A }", "ExportNamedDeclaration"},
		{"export { value };", "ExportNamedDeclaration"},
		{"export { value } from 'x';", "ExportNamedDeclaration"},
		{"export type { Value } from 'x';", "ExportNamedDeclaration"},
		{"export * from 'x';", "ExportAllDeclaration"},
		{"export * as ns from 'x';", "ExportAllDeclaration"},
	} {
		t.Run(declaration.code, func(t *testing.T) {
			var invalid []rule_tester.InvalidTestCase
			for _, selector := range []string{
				declaration.nodeType,
				"[type='" + declaration.nodeType + "']",
				"[type=" + declaration.nodeType + "]",
				"*[type='" + declaration.nodeType + "']",
				"[type='" + declaration.nodeType + "']:exit",
				":matches([type='" + declaration.nodeType + "'], Identifier[name='blocked'])",
				"[type=/^Export.*Declaration$/]",
				"[type=/^Export.*Declaration$/]:exit",
				":not([type!=/^Export.*Declaration$/])",
				"Program > [type=/^Export.*Declaration$/]",
				":statement[type=/^Export.*Declaration$/]",
				":declaration[type=/^Export.*Declaration$/]",
				"[type=/^Export.*Declaration$/]:first-child",
			} {
				invalid = append(invalid, rule_tester.InvalidTestCase{
					Code: declaration.code, Options: []any{selector},
					Errors: []rule_tester.InvalidTestCaseError{{
						MessageId: "restrictedSyntax", Line: 1, Column: 1,
						EndLine: 1, EndColumn: len(declaration.code) + 1,
					}},
				})
			}
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule, nil, invalid)
		})
	}
}

func TestNoRestrictedSyntaxExportSelectorBoundaries(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "interface I {} type T = string; namespace A.B {} import C = D.E;", Options: []any{"[type=/^Export/]"}},
			{Code: "namespace A.B.C {}", Options: []any{"[type='ExportNamedDeclaration']"}},
			{Code: "export default interface I {}", Options: []any{"[type='exportDefaultDeclaration']"}},
			{Code: "export default interface I {}", Options: []any{"[type='UnknownDeclaration']"}},
			{Code: "export default interface I {}", Options: []any{"[type='ExportNamedDeclaration']"}},
			{Code: "export interface I {}", Options: []any{"[type='ExportDefaultDeclaration']"}},
			{Code: `export import type A = require("x");`, Options: []any{"[exportKind='type']"}},
			{Code: "export = value;", Options: []any{"[type='ExportDefaultDeclaration']", "[type=/^Export/]", "[declaration]", "[exportKind]"}},
			{Code: "export default (value as number);", Options: []any{"[type='ExportDefaultDeclaration'][declaration.type='Identifier']"}},
			{Code: "// eslint-disable-next-line test\nexport default interface I {}", Options: []any{"[type='ExportDefaultDeclaration']"}},
			{Code: "export type T = string; // eslint-disable-line test", Options: []any{"[type=/^Export/]:exit"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "namespace N { export interface I {} }", Options: []any{"[type=/^Export/]:first-child", "[type=/^Export/]:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 36},
					{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: "namespace N { export interface A {} export type B = A; }", Options: []any{"[type=/^Export/] + [type=/^Export/]", "[type=/^Export/]:nth-child(2)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 37, EndLine: 1, EndColumn: 55},
					{MessageId: "restrictedSyntax", Line: 1, Column: 37, EndLine: 1, EndColumn: 55},
				},
			},
			{
				Code: "export { type T };", Options: []any{"[type='ExportSpecifier'][exportKind='type']"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 16}},
			},
			{
				Code: "export type { T };", Options: []any{"[type='ExportSpecifier'][exportKind='value']"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 15, EndLine: 1, EndColumn: 16}},
			},
			{
				Code: "export { type T };", Options: []any{"[type=/^Export/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "export default (value as number);", Options: []any{"[type='ExportDefaultDeclaration'][declaration.type='TSAsExpression']"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 34}},
			},
			{
				Code: "export default interface I {}", Options: []any{":not(Program, Identifier)"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}},
			},
			{
				Code: "export default 1;", Options: []any{"[type='ExportDefaultDeclaration'][declaration.type='Literal'][exportKind='value']"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}},
			},
			{
				Code: "export default (value);", Options: []any{"ExportDefaultDeclaration > Identifier.declaration"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 22}},
			},
			{
				Code: "export { value };", Options: []any{"[type='ExportNamedDeclaration'][declaration=null]"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}},
			},
			{
				Code: "export interface A {}\nexport type B = A;", Options: []any{"[type=/^Export/] + [type=/^Export/]"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 19}},
			},
			{
				Code: "export default interface I {}", Options: []any{"[type=/^Export/]:has(Identifier[name='I'])"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}},
			},
			{
				Code: "export default interface I {}", Options: []any{"[type='Program']:has([type='ExportDefaultDeclaration'])"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30}},
			},
			{
				Code: "export default interface I {}", Options: []any{"[type='ExportDefaultDeclaration'] Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 26, EndLine: 1, EndColumn: 27}},
			},
			{
				// Adding export hosts must neither expose unsupported physical
				// declarations nor add virtual diagnostics to the bare wildcard.
				Code: "export default interface I {}", Options: []any{"*", "[type=/^Export/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
					{MessageId: "restrictedSyntax", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
		})
}

func TestNoRestrictedSyntaxExportAttributeSubjects(t *testing.T) {
	// Unlike a named selector, an attribute selector can also receive the
	// left-hand node for esquery's reverse sibling-subject matching.
	for _, selector := range []string{
		"!ExpressionStatement ~ [type='ExportDefaultDeclaration']",
		"ExpressionStatement + ![type='ExportDefaultDeclaration']",
		":matches(!ExpressionStatement ~ [type='ExportDefaultDeclaration'], Identifier[name='missing'])",
	} {
		t.Run(selector, func(t *testing.T) {
			rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule, nil,
				[]rule_tester.InvalidTestCase{{
					Code: "value(); export default function f() {}", Options: []any{selector},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
						{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 40},
					},
				}})
		})
	}
	t.Run("preserve typed dispatch", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule, nil,
			[]rule_tester.InvalidTestCase{
				{
					Code: "[a && b, c + d];", Options: []any{"!LogicalExpression ~ BinaryExpression[operator='+']"},
					Errors: []rule_tester.InvalidTestCaseError{{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15}},
				},
				{
					Code: "[a && b, c + d];", Options: []any{"!LogicalExpression ~ [type='BinaryExpression'][operator='+']"},
					Errors: []rule_tester.InvalidTestCaseError{
						{MessageId: "restrictedSyntax", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
						{MessageId: "restrictedSyntax", Line: 1, Column: 10, EndLine: 1, EndColumn: 15},
					},
				},
			})
	})
}

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
			{Code: "const x = 1; export = x;", Options: []any{"ExportDefaultDeclaration"}},
			{Code: "export default function f() {}", Options: []any{"Program > FunctionDeclaration"}},
			{Code: "export default function f() {} sample();", Options: []any{"Program:not(:has(ExportNamedDeclaration, ExportDefaultDeclaration, ExportAllDeclaration)) > ExpressionStatement > CallExpression"}},
			{Code: "sample(); export default class {}", Options: []any{"Program:not(:has(ExportDefaultDeclaration)) CallExpression"}},
			{Code: "export default function f() {}", Options: []any{"ExportNamedDeclaration"}},
			{Code: "const f = 1;\nexport { f as default };", Options: []any{"ExportDefaultDeclaration"}},
			{Code: "export default function f() {}", Options: []any{":not(ExportDefaultDeclaration) > FunctionDeclaration"}},
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
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function () {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "export default async function f() {}",
				Options: []any{"ExportDefaultDeclaration > FunctionDeclaration[async=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 37},
				},
			},
			{
				Code:    "export default class {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default class C {}",
				Options: []any{"ExportDefaultDeclaration > ClassDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "export default abstract class C {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    "export default interface I {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "const x = 1; export default x;",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "const x = 1; export default x;",
				Options: []any{"ExportDefaultDeclaration > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "export default 1;",
				Options: []any{"ExportDefaultDeclaration > Literal"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default class C {}",
				Options: []any{"ClassDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "@dec export default class C {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 6, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "@dec export default class C {}",
				Options: []any{"ClassDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 21, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default class C { m() {} }",
				Options: []any{"ExportDefaultDeclaration MethodDefinition"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 26, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code:    "export default function f() { sample(); }",
				Options: []any{"ExportDefaultDeclaration CallExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 31, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration[declaration.type='FunctionDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration[exportKind='value']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration[parent.type='Program']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"FunctionDeclaration[parent.type='ExportDefaultDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration > .declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "const a = 1; export default function f() {}",
				Options: []any{"VariableDeclaration + ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 14, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"Program > ExportDefaultDeclaration:last-child"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{":matches(ExportDefaultDeclaration, FunctionDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration:exit"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"[type=/^Export/]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{":statement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
					{MessageId: "restrictedSyntax", Line: 1, Column: 29, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f(): void;\nexport default function f() {}",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 31},
				},
			},
			{
				Code:    "export default class {}",
				Options: []any{"ClassDeclaration[id=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "export default function () {}",
				Options: []any{"FunctionDeclaration[id=null]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code:    "declare module 'm' { export default function f(): void; }",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"Program:has(> ExportDefaultDeclaration)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}\nexport const a = 1;",
				Options: []any{"ExportDefaultDeclaration ~ ExportNamedDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 20},
				},
			},
			{
				Code:    "export const a = 1;\nexport default function f() {}",
				Options: []any{"ExportNamedDeclaration + ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration[declaration.id.name='f']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"[type='ExportDefaultDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default class C { static {} }",
				Options: []any{"ExportDefaultDeclaration StaticBlock"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 26, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{"ExportDefaultDeclaration > FunctionDeclaration > Identifier"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "export default function f() {}",
				Options: []any{":declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default (function () {});",
				Options: []any{"ExportDefaultDeclaration > FunctionExpression"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 17, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "export default class {}",
				Options: []any{"ExportDefaultDeclaration > ClassDeclaration > ClassBody"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 24},
				},
			},
		})
}

// A FunctionDeclaration without a body is TSDeclareFunction in ESTree, including
// overload signatures under ExportDefaultDeclaration / ExportNamedDeclaration.
func TestNoRestrictedSyntaxDeclareFunctionRegressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedSyntaxRule,
		[]rule_tester.ValidTestCase{
			{Code: "declare function g(): void;", Options: []any{"FunctionDeclaration"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    "export default function f(x: string): string;\nexport default function f(x: any) { return x; }",
				Options: []any{"ExportDefaultDeclaration[declaration.type='FunctionDeclaration']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 48},
				},
			},
			{
				Code:    "export default function f(x: string): string;\nexport default function f(x: any) { return x; }",
				Options: []any{"ExportDefaultDeclaration[declaration.type='TSDeclareFunction']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    "export default function f(x: string): string;\nexport default function f(x: any) { return x; }",
				Options: []any{"ExportDefaultDeclaration > FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 16, EndLine: 2, EndColumn: 48},
				},
			},
			{
				Code:    "export default function f(x: string): string;\nexport default function f(x: any) { return x; }",
				Options: []any{"ExportDefaultDeclaration > TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    "export default function f(x: string): string;\nexport default function f(x: any) { return x; }",
				Options: []any{"ExportDefaultDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 46},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 48},
				},
			},
			{
				Code:    "export function f(x: string): string;\nexport function f(x: any) { return x; }",
				Options: []any{"ExportNamedDeclaration[declaration.type='TSDeclareFunction']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
				},
			},
			{
				Code:    "export function f(x: string): string;\nexport function f(x: any) { return x; }",
				Options: []any{"ExportNamedDeclaration > FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 8, EndLine: 2, EndColumn: 40},
				},
			},
			{
				Code:    "export function f(x: string): string;\nexport function f(x: any) { return x; }",
				Options: []any{"ExportNamedDeclaration[exportKind='value']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 38},
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 40},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{"FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{"TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{"Program > TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{"TSDeclareFunction[id.name='f']"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{":function"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{":statement"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
					{MessageId: "restrictedSyntax", Line: 2, Column: 20, EndLine: 2, EndColumn: 33},
					{MessageId: "restrictedSyntax", Line: 2, Column: 22, EndLine: 2, EndColumn: 31},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{":declaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "function f(x: string): string;\nfunction f(x: any) { return x; }",
				Options: []any{"TSDeclareFunction + FunctionDeclaration"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code:    "declare function g(): void;",
				Options: []any{"TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "export declare function g(): void;",
				Options: []any{"ExportNamedDeclaration[exportKind='type'] > TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 8, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    "declare module 'm' { function h(): void; }",
				Options: []any{"TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 22, EndLine: 1, EndColumn: 41},
				},
			},
			{
				Code:    "async function f(): Promise<void>;\nasync function f() {}",
				Options: []any{"TSDeclareFunction[async=true]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code:    "export default function (): void;\nexport default function () {}",
				Options: []any{"ExportDefaultDeclaration > TSDeclareFunction"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 1, Column: 16, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "function f(): void;\nfunction f() {}",
				Options: []any{"FunctionDeclaration[body]"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedSyntax", Line: 2, Column: 1, EndLine: 2, EndColumn: 16},
				},
			},
		})
}
