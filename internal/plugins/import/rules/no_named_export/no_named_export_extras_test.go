package no_named_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_named_export"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNamedExportExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_named_export.NoNamedExportRule,
		[]rule_tester.ValidTestCase{
			// Non-exported declarations do not trigger modifier listeners.
			{
				Code: "const value = 1; function f() {} class C {} interface I {} type T = string; enum E {} namespace N.Inner {} import Alias = N;",
			},
			// Default declaration forms and expressions are allowed.
			{
				Code: "export default async function f() {}",
			},
			{
				Code: "export default abstract class C {}",
			},
			{
				Code: "export default interface I { value: string; }",
			},
			{
				Code: "export default (() => <div />);",
				Tsx:  true,
			},
			// CommonJS and TypeScript export assignment/global aliases are separate ESTree kinds.
			{
				Code: "exports.value = 1; module.exports = {}; export = value; export as namespace Library;",
			},
			// Aliases use the decoded exported name, including string literals.
			{
				Code: "const value = 1; export { value as \"def\\u0061ult\" };",
			},
			{
				Code: "export { default } from \"pkg\";",
			},
			{
				Code: "export type { T as default } from \"pkg\";",
			},
			{
				Code: "export { type T as default } from \"pkg\";",
			},
			{
				Code: `const value = 1; export { value as def\u0061ult };`,
			},
			{
				Code: `export { "original" as default } from "pkg";`,
			},
			{
				Code: `export type { default } from "pkg";`,
			},
			// JSDoc declarations and global export comments are not ES exports.
			{
				Code:     "/* exported value */\n/** @typedef {number} Value */\nconst value = 1;",
				FileName: "no-named-export.js", TSConfig: "tsconfig.allow-js.json",
			},
			{
				Code:     "/** @public */\nexport default class C {}",
				FileName: "no-named-export.js", TSConfig: "tsconfig.allow-js.json",
			},
			// Explicit source goals bypass the rule, even when tsgo accepts export syntax.
			{
				Code:            "export const value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            "export const value = 1;",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			// Inline directives can suppress this rule (the Go tester registers it as test).
			{
				Code: "// eslint-disable-next-line test\nexport const value = 1;",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Empty export lists report, including type-only and re-export forms.
			{
				Code:   "export {};\nexport {} from \"pkg\";\nexport type {};",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 11), noNamedExportError(2, 1, 2, 22), noNamedExportError(3, 1, 3, 16)},
			},
			// All star exports report, even when the namespace name is default.
			{
				Code:   "export * as ns from \"pkg\";\nexport * as default from \"pkg\";\nexport type * from \"pkg\";\nexport type * as default from \"pkg\";",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 27), noNamedExportError(2, 1, 2, 32), noNamedExportError(3, 1, 3, 26), noNamedExportError(4, 1, 4, 37)},
			},
			// A mixed list reports once, whether default occurs first or last.
			{
				Code:   "const value = 1, other = 2;\nexport { value as default, other };\nexport { value, other as default };",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 1, 2, 36), noNamedExportError(3, 1, 3, 36)},
			},
			// Imported names do not determine whether the export is default.
			{
				Code:   "export { default as value } from \"pkg\";\nexport { default as \"Default\" } from \"pkg\";\nexport { default as \"\" } from \"pkg\";",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 40), noNamedExportError(2, 1, 2, 44), noNamedExportError(3, 1, 3, 37)},
			},
			// TypeScript declarations and import aliases are named exports.
			{
				Code:   "export interface I {}\nexport const enum E {}\nexport import Alias = require(\"pkg\");\nexport declare const value: number;\nexport declare function f(): void;",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 22), noNamedExportError(2, 1, 2, 23), noNamedExportError(3, 1, 3, 38), noNamedExportError(4, 1, 4, 36), noNamedExportError(5, 1, 5, 35)},
			},
			// Both the exported dotted namespace and its explicit inner export are reported.
			{
				Code:   "export namespace Outer.Inner {\n  export const value = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 3, 2), noNamedExportError(2, 3, 2, 26)},
			},
			// Dotted namespace components are not implicit export declarations.
			{
				Code:   "namespace Outer.Inner { export const value = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 25, 1, 48)},
			},
			// Exports inside ambient modules are still visited.
			{
				Code:   "declare module \"pkg\" { export const value: number; }",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 24, 1, 51)},
			},
			// Leading decorators are outside the export range; following decorators are inside.
			{
				Code:   "@sealed\nexport class C {}\nexport @sealed class D {}",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 1, 2, 18), noNamedExportError(3, 1, 3, 26)},
			},
			// Each overload and implementation has its own export range.
			{
				Code:   "export function f(value: string): string;\nexport function f(value: string) { return value; }",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 42), noNamedExportError(2, 1, 2, 51)},
			},
			// Multiline ranges include the full list and import attributes.
			{
				Code:   "export {\n  default as data,\n  value,\n} from \"./data.json\" with { type: \"json\" };",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 4, 44)},
			},
			// Diagnostic columns use UTF-16, with comments and CRLF preserved.
			{
				Code:   "/* 😀 */ export const 名 = \"🌍\"; // end\r\nexport { 名 as \"值\" };",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 10, 1, 32), noNamedExportError(2, 1, 2, 21)},
			},
			// A semicolon after a declaration is outside its export range.
			{
				Code:   "export class C { #value = 1; [key]() { return this.#value; } };\nexport async function* f() {};",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 63), noNamedExportError(2, 1, 2, 30)},
			},
			// JS and JSX initializers do not affect declaration reporting.
			{
				Code:   "export const element = (<div />);",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 34)},
			},
			// Type specifiers report; a string-named default namespace is still a star export.
			{
				Code:   "export { type T } from \"pkg\";\nexport * as \"default\" from \"pkg\";",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 30), noNamedExportError(2, 1, 2, 34)},
			},
			// Ambient declarations and type import aliases retain their export wrappers.
			{
				Code:   "export declare namespace A.B.C {}\nexport module M.N {}\nexport declare global { interface Window {} }\nexport import type Alias = require(\"pkg\");",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(1, 1, 1, 34), noNamedExportError(2, 1, 2, 21), noNamedExportError(3, 1, 3, 46), noNamedExportError(4, 1, 4, 43)},
			},
			// JS input: hashbang, JSDoc, destructuring, and trivia before a named list.
			{
				Code:     "#!/usr/bin/env node\n/** @typedef {number} Value */\nexport const [one, ...rest] = values;\n/* 😀 */ export { one as \"named\" };",
				FileName: "no-named-export.js", TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(3, 1, 3, 38), noNamedExportError(4, 10, 4, 36)},
			},
			// BOM and Unicode line separators must not change complete export ranges.
			{
				Code:   "\uFEFF/* lead */\nexport {\n  default as value, other,\n} from \"pkg\";\u2028export {};",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(2, 1, 4, 14), noNamedExportError(5, 1, 5, 11)},
			},
			// Reporting ranges still respect block and line disable directives.
			{
				Code:   "/* eslint-disable test */\nexport { hidden } from \"pkg\";\n/* eslint-enable test */\nexport { visible } from \"pkg\";\nexport * from \"pkg\"; // eslint-disable-line test",
				Errors: []rule_tester.InvalidTestCaseError{noNamedExportError(4, 1, 4, 31)},
			},
		},
	)
}
