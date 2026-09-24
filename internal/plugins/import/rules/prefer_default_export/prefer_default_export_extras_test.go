package prefer_default_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/prefer_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDefaultExportExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_default_export.PreferDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    "export const { item: {} } = source;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "const value = 1; export { value as def\\u0061ult };",
			},
			// Disable comments suppress the final report without changing export counts.
			{
				Code: "// eslint-disable-next-line test\nexport const value = 1;",
			},
			{
				Code:    "export const first = 1;\nexport const last = 2; // eslint-disable-line test",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "/* eslint-disable test */\nexport const hidden = 1;\n/* eslint-enable test */\nexport const visible = 2;",
			},
			{
				Code:    "export const first = 1, second = 2;",
				Options: map[string]any{"target": "single"},
			},
			{
				Code:    "module.exports = { value: 1 };",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "const value = 1; export = value;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export as namespace Library;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export {};",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export {} from \"values\";",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export const {} = value; export const [] = values;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export type Value = number; export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export const value = 1; export interface Value {}",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export type * from \"values\"; export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export * as values from \"values\"; export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "export * as default from \"values\"; export const value = 1;",
			},
			{
				Code:    "export const value = 1; export { default } from \"values\";",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "const value = 1; export { value as \"\\u0064efault\" };",
			},
			{
				Code:    "export default class Value {}; export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export default interface Value {}; export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code:    "export const value = 1; export default (value);",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "export const [first, ...rest] = values;",
			},
			{
				Code: "export const [[first, second]] = values;",
			},
			// Array holes count as exports upstream.
			{
				Code: "export const [, value] = values;",
			},
			{
				Code: "export namespace Outer { export const value = 1; }",
			},
			{
				Code:    "declare namespace Values { export type Value = number; } export const value = 1;",
				Options: map[string]any{"target": "any"},
			},
			{
				Code: "export function value(input: string): string; export function value(input: any) { return input; }",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Types without their own export do not exempt the containing file.
			{
				Code:   "export namespace Values { interface Value {} }",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 47)},
			},
			{
				Code:   "declare namespace Values { interface Value {} } export const value=1;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 49, 1, 70)},
			},
			{
				Code:   "export const {} = source, value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 37)},
			},
			{
				Code:   "export const [...[]] = values;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 31)},
			},
			{
				Code:   "\uFEFF// comment\r\nexport const value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 1, 2, 24)},
			},
			// JSDoc types and namespace tags are comments, not ES exports.
			{
				Code:     "/** @typedef {number} Value */ export const value = 1;",
				FileName: "prefer-default-export.js", TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 32, 1, 55)},
			},
			{
				Code:     "/** @namespace */ export const values = {};",
				FileName: "prefer-default-export.js", TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 19, 1, 44)},
			},
			// Disabling a later empty export must not hide the earlier report site.
			{
				Code:   "export const value = 1;\n/* eslint-disable test */\nexport {};",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 24)},
			},
			{
				Code:    "/* eslint-disable test */\nexport const hidden = 1;\n/* eslint-enable test */\nexport const visible = 2;",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 4, 1, 4, 26)},
			},
			// An empty options object uses the default target.
			{
				Code:    "export let value = 1;",
				Options: map[string]any{},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 22)},
			},
			{
				Code:    "export class Value {}",
				Options: map[string]any{"target": "single"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 22)},
			},
			{
				Code:    "export const first = 1, second = 2;",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 1, 1, 36)},
			},
			// An empty declaration still becomes the last report site.
			{
				Code:   "export const value = 1; export const {} = source;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 25, 1, 50)},
			},
			// An empty export list does not replace the last report site.
			{
				Code:    "export const value = 1; export {};",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 1, 1, 24)},
			},
			// Only directly exported type declarations exempt the file.
			{
				Code:   "type Value = number; export { Value };",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 31, 1, 36)},
			},
			{
				Code:   "export type { Value } from \"values\";",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 15, 1, 20)},
			},
			{
				Code:   "export { type Value as Renamed } from \"values\";",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 10, 1, 31)},
			},
			{
				Code:   "export const { value: renamed = 1 } = source;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 46)},
			},
			{
				Code:   "export const { [key()]: renamed } = source;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 44)},
			},
			{
				Code:   "export const { ...rest } = source;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 35)},
			},
			{
				Code:   "export const [{ value }] = values;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 35)},
			},
			// Defaulted nested patterns count once, not once per bound identifier.
			{
				Code:   "export const [{ first, second } = {}] = values;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 48)},
			},
			{
				Code:   "export const { item: { first, second } = {} } = source;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 56)},
			},
			// A rest element also counts once even when it binds a nested pattern.
			{
				Code:   "export const [...[first, second]] = values;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 44)},
			},
			{
				Code:   "export const [,] = values;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 27)},
			},
			{
				Code:    "export const [,,] = values;",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 1, 1, 28)},
			},
			{
				Code:   "export enum State { Ready }",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 28)},
			},
			{
				Code:   "export declare function value(): void;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 39)},
			},
			{
				Code:   "export declare class Value {}",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 30)},
			},
			{
				Code:   "export import Value = Namespace.Value;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 39)},
			},
			// Ignore the synthetic export modifier on a dotted namespace segment.
			{
				Code:   "export namespace Outer.Inner {}",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 32)},
			},
			{
				Code:    "export namespace Outer { export const value = 1; }",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 26, 1, 49)},
			},
			{
				Code:   "declare module \"values\" { export const value: number; }",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 27, 1, 54)},
			},
			{
				Code:   "@decorator\nexport class Value {}",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 2, 1, 2, 22)},
			},
			{
				Code:   "export @decorator class Value {}",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 33)},
			},
			{
				Code:   "/* comment */ export const value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 15, 1, 38)},
			},
			{
				Code:   "export /* before declaration */ const value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 49)},
			},
			{
				Code:   "const value = 1;\nexport {\n  /* before specifier */ value as renamed\n};",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 3, 26, 3, 42)},
			},
			{
				Code:   "const 𐐀 = 1; export { 𐐀 as \"值\" };",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 24, 1, 33)},
			},
			// Initializers, including JSX and optional access, do not affect export counts.
			{
				Code:     "export const view = <Widget value={item?.value} />;",
				FileName: "file.tsx",
				Errors:   []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 52)},
			},
			{
				Code:   "export const value = (source?.value as number)!;",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 49)},
			},
			{
				Code:   "export class Value { #value = 1; [key()]() { return this.#value; } }",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 1, 1, 69)},
			},
			{
				Code:   "export { value as \"display name\" } from \"values\";",
				Errors: []rule_tester.InvalidTestCaseError{preferDefaultExportError(singleExportMessage, 1, 10, 1, 33)},
			},
			{
				Code:    "export { first } from \"one\"; export { second as last } from \"two\";",
				Options: map[string]any{"target": "any"},
				Errors:  []rule_tester.InvalidTestCaseError{preferDefaultExportError(anyExportMessage, 1, 39, 1, 53)},
			},
		},
	)
}
