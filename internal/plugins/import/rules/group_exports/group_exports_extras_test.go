package group_exports

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Cases cover AST adaptation, additional branches, and documented corrections
// to eslint-plugin-import v2.32.0. Diagnostic ranges follow
// @typescript-eslint/parser 8.65.0.
func TestGroupExportsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &GroupExportsRule,
		[]rule_tester.ValidTestCase{
			// Reads, updates, destructuring targets, and unrelated assignments do not count.
			{Code: "module.exports; exports.x++; module.exports.x--; [exports.a, exports.b] = values; ({x: exports.c} = value); for (exports.x of values) {} other.x = 1;"},
			// Destructuring defaults are assignment patterns, not export assignments.
			{Code: `({ a: exports.x = 1, b: exports.y = 2 } = source);
[exports.a = 1, ...[exports.b = 2]] = source;
for ([exports.a = 1, exports.b = 2] of source) {}
for ({a: exports.a = 1, b: exports.b = 2} in source) {}`},
			// Default declarations and export assignments are not named exports.
			{Code: "export default class C {} export const value = 1; export = value; export as namespace Library;"},
			// A default function is also ignored.
			{Code: "export default function f() {} export const value = 1;"},
			// Star exports, including namespace and type-only stars, are ignored.
			{Code: "export * from \"m\"; export * as ns from \"m\"; export type * from \"m\"; export type * as types from \"m\"; export {a} from \"m\";"},
			// The spelling of the source, not its resolved file, determines the group.
			{Code: "export {a} from \"./m\"; export {b} from \"./m.js\"; export {c} from \"m\";"},
			// An empty source is distinct from a local export.
			{Code: "const a = 1; export {a}; export {b} from \"\";"},
			// Dynamic module properties and deeper mutations are not export assignments.
			{Code: "const exports = \"state\"; module[exports].a = 1; module[(exports)].b = 2; module[exports as string] = {}; module.exports.deep.x = 1; exports.deep.x = 2; module.exports = {};"},
			// Assertions and non-null expressions on the whole target are not member expressions.
			{Code: "(module.exports as object) = {}; module.exports! = {}; exports.x = 1;"},
			// Unrelated namespace bodies cannot consolidate their exports.
			{Code: "namespace A { export const a = 1 } namespace B { export const b = 2 }"},
			{Code: `export namespace A.B { export const a = 1; }
namespace C { export const b = 2; }`},
			// Separate declarations of the same namespace also have separate bodies.
			{Code: "namespace A { export const a = 1; } namespace A { export const b = 2; }"},
			// Local values/types and re-exports all respect their containing module body.
			{Code: `declare module "a" { export {a} from "m"; export type {A} from "m"; }
declare module "b" { export {b} from "m"; export type {B} from "m"; }
export {c} from "m"; export type {C} from "m";`},
			{Code: `namespace A { export const a = 1; export type T = string; }
namespace B { export const b = 2; export type U = string; }
export const top = 3; export type V = string;`},
			// Private members are not public CommonJS exports, at either level.
			{Code: "class C { #exports; #x; m(module, exports) { module.#exports = {}; module.#exports.a = 1; exports.#x = 2; module.exports.#x = 3; module.exports = {}; } }"},
			// Calls and optional-chain receivers cannot supply a CommonJS root.
			{Code: `getBox().exports.a = 1; getBox().exports.b = 2;
factory().module.exports = {};
factory().exports.x = 1;
(factory?.()).exports.y = 2;
(obj?.value).exports.z = 3;`},
			// Authored TypeScript wrappers do not create partial CommonJS chains.
			{Code: `(module as M).exports = {};
(module as M).exports.x = 1;
module[exports as string].x = 2;
exports.y = 3;`},
			// An export type and an inline type specifier belong to different groups.
			{Code: "type A = string; type B = number; export type {A}; export {type B};"},
		},
		[]rule_tester.InvalidTestCase{
			// __proto__ is an ordinary source name for both value and type re-exports.
			{Code: `export {a} from "__proto__";
export {b} from "__proto__";
export type {A} from "__proto__";
export type {B} from "__proto__";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 29},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 34},
					{MessageId: "", Message: namedMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 34},
				}},
			// Exports in one namespace still group together, separately from its declaration.
			{Code: `export namespace A.B {
  export const a = 1;
  export const b = 2;
}
export const outside = 0;
namespace C { export const c = 3; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 3, EndLine: 2, EndColumn: 22},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 3, EndLine: 3, EndColumn: 22},
					{MessageId: "", Message: namedMessage, Line: 5, Column: 1, EndLine: 5, EndColumn: 26},
				}},
			// Value and type re-exports group by source within their module body.
			{Code: `declare module "a" {
  export {a} from "m";
  export {b} from "m";
  export type {A} from "m";
  export type {B} from "m";
}
declare module "b" { export {c} from "m"; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 3, EndLine: 2, EndColumn: 23},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 3, EndLine: 3, EndColumn: 23},
					{MessageId: "", Message: namedMessage, Line: 4, Column: 3, EndLine: 4, EndColumn: 28},
					{MessageId: "", Message: namedMessage, Line: 5, Column: 3, EndLine: 5, EndColumn: 28},
				}},
			// Literal exports access is CommonJS; the exported member can be computed.
			{Code: "module[\"exports\"] = {};\nmodule[(`exports`)].one = 1;\nexports[key] = 2;\nmodule.exports[other] = 3;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 28},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 17},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 26},
				}},
			// Computed keys and default-value expressions still contain real assignments.
			{Code: "({ [exports.key = 1]: exports.x = (exports.y = 2) } = source);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 5, EndLine: 1, EndColumn: 20},
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
				}},
			{Code: "for ([exports.x = (exports.y = 1)] of source) { exports.z = 2; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 33},
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 49, EndLine: 1, EndColumn: 62},
				}},
			// Assignments in array values are expressions rather than patterns.
			{Code: "const values = [exports.a = 1, exports.b = 2];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 30},
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 32, EndLine: 1, EndColumn: 45},
				}},
			// Import-equals declarations stay in the value group regardless of modifiers.
			{Code: `export import type A = require("m");
export declare import B = NS.B;
export const value = 1;
export type T = string;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 37},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 32},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 24},
				}},
			// Exported TypeScript import-equals declarations also count as named exports.
			{Code: "export import A = NS.A; export import B = require(\"m\"); export const value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 56},
					{MessageId: "", Message: namedMessage, Line: 1, Column: 57, EndLine: 1, EndColumn: 80},
				}},
			// Every declaration is reported, with no fix or suggestion.
			{Code: `export let a = 1, b = 2;
export var c = 3;
export const {d, ...rest} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 18},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 33},
				}},
			// Functions, classes, and enums all count as value declarations.
			{Code: `export async function* values() {}
export class Service {}
export const enum State { Ready }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 34},
				}},
			// Interfaces, aliases, and explicit ambient declarations share the type group.
			{Code: `export interface A {}
export type B = string;
export declare const c: number;
export declare function f(): void;
export declare class D {}
export declare enum E { One }
export const value = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 32},
					{MessageId: "", Message: namedMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 35},
					{MessageId: "", Message: namedMessage, Line: 5, Column: 1, EndLine: 5, EndColumn: 26},
					{MessageId: "", Message: namedMessage, Line: 6, Column: 1, EndLine: 6, EndColumn: 30},
				}},
			// Ambient namespaces join the type group; dotted namespaces count once.
			{Code: `export declare namespace A.B {}
export type C = string;
export const value = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				}},
			// Named default aliases, string export names, and empty exports are named declarations.
			{Code: "const value = 1; export {value as default}; export {value as \"other\"}; export {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 44},
					{MessageId: "", Message: namedMessage, Line: 1, Column: 45, EndLine: 1, EndColumn: 71},
					{MessageId: "", Message: namedMessage, Line: 1, Column: 72, EndLine: 1, EndColumn: 82},
				}},
			// Inline type specifiers do not make the declaration an export type.
			{Code: "type A = string; type B = number; export {type A}; export {type B}; export type C = boolean;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 35, EndLine: 1, EndColumn: 51},
					{MessageId: "", Message: namedMessage, Line: 1, Column: 52, EndLine: 1, EndColumn: 68},
				}},
			// Local values, local types, and each source are independent groups.
			{Code: `export const a = 1;
export type A = string;
export {b} from "m";
export type {B} from "m";
exports.one = 1;
export const c = 2;
export type C = number;
export {d} from "m";
export type {D} from "m";
exports.two = 2;
export {other} from "other";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 20},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 21},
					{MessageId: "", Message: namedMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 26},
					{MessageId: "", Message: commonJSMessage, Line: 5, Column: 1, EndLine: 5, EndColumn: 16},
					{MessageId: "", Message: namedMessage, Line: 6, Column: 1, EndLine: 6, EndColumn: 20},
					{MessageId: "", Message: namedMessage, Line: 7, Column: 1, EndLine: 7, EndColumn: 24},
					{MessageId: "", Message: namedMessage, Line: 8, Column: 1, EndLine: 8, EndColumn: 21},
					{MessageId: "", Message: namedMessage, Line: 9, Column: 1, EndLine: 9, EndColumn: 26},
					{MessageId: "", Message: commonJSMessage, Line: 10, Column: 1, EndLine: 10, EndColumn: 16},
				}},
			// Source values are decoded before grouping, including Unicode escapes.
			{Code: `export {a} from "m";
export {b} from "\u006d";
export {c} from "constructor";
export {d} from "constructor";
export {e} from "__proto__";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 26},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 31},
					{MessageId: "", Message: namedMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 31},
				}},
			// Import attributes do not separate declarations from the same source.
			{Code: `export {a} from "m" with {type: "json"};
export {b} from "m" with {type: "json"};`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 41},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 41},
				}},
			// Parentheses are transparent; report the complete assignment, not its statement.
			{Code: `(module.exports) = {};
((module).exports).one = ((1));
(exports)[("two")] = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 31},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 23},
				}},
			// All assignment operators count, but updates and equality comparisons do not.
			{Code: `module.exports ||= {};
exports.x += 1;
exports.y &&= 2;
exports.z ??= 3;
exports.x === 1;
exports.y++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 22},
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 15},
					{MessageId: "", Message: commonJSMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 16},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 16},
				}},
			// Chained assignments each produce a diagnostic with their own range.
			{Code: "module.exports = exports.x = {};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 18, EndLine: 1, EndColumn: 32},
				}},
			// Bindings and function scope do not suppress syntactic CommonJS exports.
			{Code: `function first(module) { module.exports = {}; }
function second(exports) { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 45},
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 28, EndLine: 2, EndColumn: 41},
				}},
			// Assignments inside JSX expressions count; JSX tag names and attributes do not.
			{Code: "const view = <exports.Component prop={exports.a = 1}>{module.exports.b = 2}</exports.Component>;", Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 39, EndLine: 1, EndColumn: 52},
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 55, EndLine: 1, EndColumn: 75},
				}},
			// Source-authored JSDoc parentheses are transparent to the accessor check.
			{Code: `(/** @type {any} */ (module)).exports = {};
exports.a = 1;`, FileName: "test.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 43},
					{MessageId: "", Message: commonJSMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 14},
				}},
			// Export ranges start after preceding decorators.
			{Code: `@decorate
export class A {}
export @decorate class B {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 2, EndColumn: 18},
					{MessageId: "", Message: namedMessage, Line: 3, Column: 1, EndLine: 3, EndColumn: 28},
				}},
			// Multiline exports preserve UTF-16 columns and CRLF positions.
			{Code: "/* 😀 */ export const 名 = \"😀\";\r\nexport const second = {\r\n  名,\r\n};",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: namedMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 32},
					{MessageId: "", Message: namedMessage, Line: 2, Column: 1, EndLine: 4, EndColumn: 3},
				}},
			// Assignment ranges include multiline right-hand sides and non-ASCII text.
			{Code: `/* 😀 */ exports.名 = {
  value: "😀",
};
module.exports.second = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: commonJSMessage, Line: 1, Column: 10, EndLine: 3, EndColumn: 2},
					{MessageId: "", Message: commonJSMessage, Line: 4, Column: 1, EndLine: 4, EndColumn: 26},
				}},
		},
	)
}
