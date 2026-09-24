package exports_last_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/exports_last"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestExportsLastExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &exports_last.ExportsLastRule,
		[]rule_tester.ValidTestCase{
			// Empty input and an empty export still form a valid file.
			{
				Code: "",
			},
			{
				Code: "export {};",
			},
			// Comments and directives are not trailing statements.
			{
				Code: "\"use strict\";\nconst value = 1;\nexport { value }; // trailing comment\n/* done */",
			},
			{
				Code: "#!/usr/bin/env node\nexport default true;\n// done",
			},
			// Only the Program body is checked, not namespace or module bodies.
			{
				Code: "namespace N { export const value = 1; const later = 2; }\nconst tail = 3;",
			},
			{
				Code: "declare module \"pkg\" { export const value: number; const later: number; }",
			},
			// CommonJS and TypeScript assignment/global exports are not ES exports.
			{
				Code: "module.exports = {}; exports.name = 1; const later = 2;",
			},
			{
				Code: "export = value;\nconst later = 1;",
			},
			{
				Code: "export as namespace Library;\ndeclare const value: number;",
			},
			// TypeScript exported declarations share the trailing export group.
			{
				Code: "export type T = string;\nexport interface I {}\nexport enum E {}\nexport namespace N {}\nexport import Alias = N;\nexport declare function call(): void;",
			},
			// Statements inside an exported declaration do not break the group.
			{
				Code: "export function f() { const x = 1; return x; }\nexport class C { static { const x = 1; } }\nexport namespace N { export const x = 1; const y = 2; }",
			},
			// A semicolon terminates these export expressions, not a separate statement.
			{Code: "export default (function () {});"},
			{Code: "export default (class {});"},
			// The Go RuleTester registers this rule as test. Reporting during Run
			// must still respect the context's disable directives.
			{Code: "// eslint-disable-next-line test\nexport const value = 1;\nvoid value;"},
			{
				Code:     "export default /** @type {unknown} */ (value?.[key]);",
				FileName: "exports-last.js", TSConfig: "tsconfig.allow-js.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Report only exports before the last non-export statement.
			{
				Code:   "const value = 1;\nexport { value };\nconsole.log(value);\nexport default value;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 1, 2, 18)},
			},
			// All earlier groups are reported, in source order.
			{
				Code:   "export const first = 1;\nvoid first;\nexport const second = 2;\nvoid second;\nexport const last = 3;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 24), exportsLastError(3, 1, 3, 25)},
			},
			// Named, namespace, and type re-exports.
			{
				Code:   "export { value as renamed } from \"pkg\";\nexport * as ns from \"pkg\";\nexport type { T } from \"pkg\";\nexport type * from \"pkg\";\nimport \"side-effect\";",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 40), exportsLastError(2, 1, 2, 27), exportsLastError(3, 1, 3, 30), exportsLastError(4, 1, 4, 26)},
			},
			// Every declaration represented by an export modifier in tsgo.
			{
				Code:   "export const { value } = source;\nexport async function f() {}\nexport class C {}\nexport type T = string;\nexport interface I {}\nexport const enum E {}\nexport namespace N {}\nexport import Alias = N;\nexport declare function call(): void;\nconst later = 1;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 33), exportsLastError(2, 1, 2, 29), exportsLastError(3, 1, 3, 18), exportsLastError(4, 1, 4, 24), exportsLastError(5, 1, 5, 22), exportsLastError(6, 1, 6, 23), exportsLastError(7, 1, 7, 22), exportsLastError(8, 1, 8, 25), exportsLastError(9, 1, 9, 38)},
			},
			// Multiline default declaration range.
			{
				Code:   "export default function f() {\n  return 1;\n}\nf();",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 3, 2)},
			},
			{
				Code:   "export default class C { #value = 1; [key]() { return this.#value; } }\nvoid C;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 71)},
			},
			{
				Code:   "export default interface I { value: string; }\nconst later = 1;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 46)},
			},
			// An empty export and an extra empty statement both affect ordering.
			{
				Code:   "export {};\n;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 11)},
			},
			// A final CommonJS assignment is a non-export statement.
			{
				Code:   "export const value = 1;\nmodule.exports = value;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 24)},
			},
			// TypeScript export assignment and global alias also end an ES export group.
			{
				Code:   "export const value = 1;\nexport = value;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 24)},
			},
			{
				Code:   "export const value = 1;\nexport as namespace Library;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 24)},
			},
			// A non-exported type declaration counts as a statement.
			{
				Code:   "export default true;\ninterface Local {}",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 21)},
			},
			// Only the enclosing exported namespace is reported.
			{
				Code:   "export namespace N {\n  export const value = 1;\n  const later = 2;\n}\nconst tail = 3;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 4, 2)},
			},
			// Parentheses and JSX stay inside the default export range.
			{
				Code:   "export default (\n  () => <div />\n);\nrender();",
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 3, 3)},
			},
			// Leading decorators are outside the ESTree export range.
			{
				Code:   "@sealed\nexport class C {}\nconst later = 1;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 1, 2, 18)},
			},
			// Decorators following export are inside the range.
			{
				Code:   "export @sealed class C {}\nconst later = 1;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 26)},
			},
			// UTF-16 columns, CRLF, and leading/trailing comments.
			{
				Code:   "/* 😀 */ export const 名 = \"🌍\"; // end\r\nvoid 名;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 10, 1, 32)},
			},
			// A semicolon after a function/class declaration is a separate statement.
			{
				Code:   "export default function () {};",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 30)},
			},
			{
				Code:   "export default class {};",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 24)},
			},
			// Attributes belong to the reported export; a type import still ends the group.
			{
				Code:   "export { default as data } from \"./data.json\" with { type: \"json\" };\nexport type * as types from \"./types\";\nimport type { T } from \"./types\";",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 69), exportsLastError(2, 1, 2, 39)},
			},
			// Each overload and its implementation are separate exports.
			{
				Code:   "export function f(value: string): string;\nexport function f(value: string) { return value; }\nvoid f;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 42), exportsLastError(2, 1, 2, 51)},
			},
			{
				Code:   "@sealed export default class {}\nrun();",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 9, 1, 32)},
			},
			{
				Code:   "export import Alias = require(\"pkg\");\nvoid Alias;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(1, 1, 1, 38)},
			},
			// JS input: BOM, JSDoc cast, optional/computed access, and a Unicode newline.
			{
				Code:     "\uFEFF/** start */\nexport default /** @type {unknown} */ (value?.[key]);\u2028void value;",
				FileName: "exports-last.js", TSConfig: "tsconfig.allow-js.json",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(2, 1, 2, 54)},
			},
			// Suppressing one report must not hide another export before the final statement.
			{
				Code:   "// eslint-disable-next-line test\nexport const first = 1;\nexport const second = 2;\nvoid second;",
				Errors: []rule_tester.InvalidTestCaseError{exportsLastError(3, 1, 3, 25)},
			},
		},
	)
}
