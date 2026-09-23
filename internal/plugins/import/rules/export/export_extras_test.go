package export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional AST and module-boundary cases, compared with eslint-plugin-import
// v2.32.0 and @typescript-eslint/parser 8.65.0. Upstream uses literal messages;
// rslint assigns stable IDs to its three diagnostic variants.
func TestExportExtras(t *testing.T) {
	rule_tester.RunRuleTester(exportRoot(t, "testdata/extras.txtar"), "tsconfig.json", t, &export.ExportRule,
		[]rule_tester.ValidTestCase{
			// A qualified namespace name has no simple exported identifier.
			{Code: `export namespace A.B {} export const A = 1;`},
			{Code: `export declare module "foo" {} const value = 1; export {value as "foo"};`},
			// RuleTester registers the rule as "test" for directive matching.
			{Code: `/* eslint-disable test */
export default 1; export default 2;`},
			// The existing export map terminates cyclic star traversal; upstream
			// v2.32.0's forEach recurses indefinitely on this dependency graph.
			{Code: `export * from "./cycle-a";`},
			// type-only star exports
			{Code: `export type * from "./named";`},
			// binding patterns
			{Code: `export const {nested: {value = 1}, [key]: computed, ...rest} = source; export const [first,, ...tail] = list;`},
			// namespace scopes remain separate
			{Code: `export namespace A { export const x = 1; } export namespace A { export const x = 2; } namespace B { export const x = 3; }`},
			// function namespace merge after overloads
			{Code: `export function f(x: string): void; export function f(x: number): void; export function f(x: any) {} export namespace f {}`},
			// function implementations merged upstream
			{Code: `export function f() {} export function f() {} export namespace f {}`},
			// namespace export aliases are ignored
			{Code: `export const ns = 1; export * as ns from "./named"; export * as ns from "./barrel";`},
			// export assignment is not a default declaration
			{Code: `const value = 1; export = value; export default value; export as namespace API;`},
			// unresolved and CommonJS sources
			{Code: `export * from "./not-found"; export * from "./common";`},
			// ignored star target
			{Code: `export const a = 1; export * from "./named";`,
				Settings: map[string]interface{}{"import/ignore": []interface{}{"named"}},
			},
			// one star node reached twice
			{Code: `export * from "./barrel";`},
			// namespace assignment names
			{Code: `export * from "./namespace";`},
		},
		[]rule_tester.InvalidTestCase{
			// Namespace export assignments include their import aliases.
			{Code: `export const Value = 1; export * from "./alias-namespace";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Value'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Value'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 59},
				},
			},
			// Exported import assignments are wrapped export declarations upstream.
			{Code: `export import Foo = require("./named"); export {Foo};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 18},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 49, EndLine: 1, EndColumn: 52},
				},
			},
			{Code: `export import type Foo = require("./named"); export {Foo};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 57},
				},
			},
			{Code: `namespace API { export import Foo = Other.Foo; export {Foo}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 34},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'Foo'.", Line: 1, Column: 56, EndLine: 1, EndColumn: 59},
				},
			},
			// Decode escaped identifiers and string aliases through existing helpers.
			{Code: `export const \u0061 = 1; export {a as "\u0061"};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 39, EndLine: 1, EndColumn: 47},
				},
			},
			// Preserve upstream's string prefix collision with a type export.
			{Code: `export type T = number; const v = 0; export {v as "type:T"};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'T'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'T'.", Line: 1, Column: 51, EndLine: 1, EndColumn: 59},
				},
			},
			// Suppressing one occurrence must not hide the other conflict.
			{Code: `// eslint-disable-next-line test
export const a = 1;
export {a};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 3, Column: 9, EndLine: 3, EndColumn: 10},
				},
			},
			// tsgo accepts top-level return. Espree instead emits a parser error
			// in the dependency; this documented difference remains rule-local.
			{
				Code: `export * from "./return";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './return'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			// string namespace alias follows upstream star handling
			{Code: `export * as "ns" from "./named"; export const a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 47, EndLine: 1, EndColumn: 48},
				},
			},
			// empty public name
			{Code: `const a = 1; export {a as ""}; export * from "./empty-name";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name ''.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29},
					{MessageId: "multipleNamed", Message: "Multiple exports of name ''.", Line: 1, Column: 32, EndLine: 1, EndColumn: 61},
				},
			},
			// decorated default declaration
			{Code: `@decorate export default class C {}
export default (value satisfies unknown);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 11, EndLine: 1, EndColumn: 36},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 2, Column: 1, EndLine: 2, EndColumn: 42},
				},
			},
			// Decorators after export remain inside the reported export range.
			{Code: `export default @decorate class C {}
export default 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 1, EndLine: 1, EndColumn: 36},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 2, Column: 1, EndLine: 2, EndColumn: 18},
				},
			},
			// dotted namespace assignment names
			{Code: `export * from "./dotted";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './dotted'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 25},
				},
			},
			// default expression and declaration
			{Code: `export default (value); export default class {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 25, EndLine: 1, EndColumn: 48},
				},
			},
			// named and literal defaults
			{Code: `const a = 1, b = 2; export {a as default, b as "default"};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 34, EndLine: 1, EndColumn: 41},
					{MessageId: "multipleDefault", Message: "Multiple default exports.", Line: 1, Column: 48, EndLine: 1, EndColumn: 57},
				},
			},
			// alias ranges and non-ASCII columns
			{Code: `const marker = "😀"; const café = 1; export {café as "你好"}; export {café as "你好"};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name '你好'.", Line: 1, Column: 54, EndLine: 1, EndColumn: 58},
					{MessageId: "multipleNamed", Message: "Multiple exports of name '你好'.", Line: 1, Column: 77, EndLine: 1, EndColumn: 81},
				},
			},
			// nested binding names
			{Code: `export const {nested: {value = 1}, [key]: renamed, ...rest} = source; export {value, renamed, rest};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 29},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'renamed'.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'rest'.", Line: 1, Column: 55, EndLine: 1, EndColumn: 59},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 79, EndLine: 1, EndColumn: 84},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'renamed'.", Line: 1, Column: 86, EndLine: 1, EndColumn: 93},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'rest'.", Line: 1, Column: 95, EndLine: 1, EndColumn: 99},
				},
			},
			// array binding holes and rest
			{Code: `export let [first,, {deep}, ...tail] = source; export {first, deep, tail};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'first'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 18},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'deep'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'tail'.", Line: 1, Column: 32, EndLine: 1, EndColumn: 36},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'first'.", Line: 1, Column: 56, EndLine: 1, EndColumn: 61},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'deep'.", Line: 1, Column: 63, EndLine: 1, EndColumn: 67},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'tail'.", Line: 1, Column: 69, EndLine: 1, EndColumn: 73},
				},
			},
			// typed variable identifier ranges
			{Code: `export let value!: number; export {value};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 12, EndLine: 1, EndColumn: 26},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 36, EndLine: 1, EndColumn: 41},
				},
			},
			// interface duplicates
			{Code: `export interface A {} export interface A {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 40, EndLine: 1, EndColumn: 41},
				},
			},
			// type alias and interface share type namespace
			{Code: `export type A = number; export interface A {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 42, EndLine: 1, EndColumn: 43},
				},
			},
			// type specifiers use the value bucket upstream
			{Code: `const A = 1; type T = string; export {A}; export type {T as A};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 61, EndLine: 1, EndColumn: 62},
				},
			},
			// inline type specifier
			{Code: `export const A = 1; export {type T as A} from "./named";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
				},
			},
			// class generic identifier
			{Code: `export class C<T> {} export {C};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'C'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'C'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			// ambient function signatures are omitted
			{Code: `export declare function f(): void; export const f = 1; export {f};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'f'.", Line: 1, Column: 49, EndLine: 1, EndColumn: 50},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'f'.", Line: 1, Column: 64, EndLine: 1, EndColumn: 65},
				},
			},
			// namespace and variable conflict
			{Code: `export const A = 1; export namespace A {} export namespace A {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 38, EndLine: 1, EndColumn: 39},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'A'.", Line: 1, Column: 60, EndLine: 1, EndColumn: 61},
				},
			},
			// module declarations are distinct scopes
			{Code: `declare module "a" { export const x = 1; export {x}; } declare module "b" { export const x = 2; export {x}; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 50, EndLine: 1, EndColumn: 51},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 90, EndLine: 1, EndColumn: 91},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 105, EndLine: 1, EndColumn: 106},
				},
			},
			// dotted namespace scope
			{Code: `export namespace A.B { export const x = 1; export {x}; } export namespace C.B { export const x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 37, EndLine: 1, EndColumn: 38},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'x'.", Line: 1, Column: 52, EndLine: 1, EndColumn: 53},
				},
			},
			// star export collisions and excludes default
			{Code: `export * from "./named"; export * from "./barrel"; export default 0;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'z'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'T'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'z'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 51},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 51},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'T'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 51},
				},
			},
			// named reexports still declare missing names
			{Code: `export const alias = 1; export * from "./missing-name";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'alias'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'alias'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 56},
				},
			},
			// star of unresolved star has no known names
			{Code: `export * from "./unknown";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './unknown'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 26},
				},
			},
			// empty ES module
			{Code: `export * from "./empty";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './empty'.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
				},
			},
			// mixed early and deferred diagnostics
			{Code: `export const a = 1; export {a}; export * from "./empty";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'a'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
					{MessageId: "noNamed", Message: "No named exports found in module './empty'.", Line: 1, Column: 47, EndLine: 1, EndColumn: 56},
				},
			},
			// namespace assignment collision
			{Code: `export const value = 1; export * from "./namespace";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
					{MessageId: "multipleNamed", Message: "Multiple exports of name 'value'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 53},
				},
			},
			// dotted namespace assignment collision
			{Code: `export const Nested = 1; export * from "./dotted";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noNamed", Message: "No named exports found in module './dotted'.", Line: 1, Column: 40, EndLine: 1, EndColumn: 50},
				},
			},
		},
	)
}
