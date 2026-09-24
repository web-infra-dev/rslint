package unambiguous_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/unambiguous"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Expectations checked against eslint-plugin-import v2.32.0 with ESLint 9.39.4
// (Espree for JS) and @typescript-eslint/parser 8.65.0 (for TS).
func TestUnambiguousExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &unambiguous.UnambiguousRule,
		[]rule_tester.ValidTestCase{
			// Additional syntax, suppression, BOM, and line-ending regressions.
			{
				Code: "import data from \"data\" with { type: \"json\" };",
			},
			{
				Code: "export type * from \"pkg\";",
			},
			{
				Code: "export type {};",
			},
			{
				Code: "@sealed\nexport class C {}",
			},
			{
				Code: "export default ((value?.[key]));",
			},
			{
				Code:     "export {};\ndeclare global { interface Window { value: number } }",
				FileName: "global.d.ts",
			},
			{
				Code: "/* eslint-disable test */\nconst value = 1;",
			},
			{
				Code: "// eslint-disable-next-line test\nconst value = 1;",
			},
			{
				Code:     "const value = 1; // eslint-disable-line test",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
			},
			// Source goals and extension defaults.
			{
				Code:            "module.exports = {};",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
			{
				Code:            "function script() {}",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            "",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:     "module.exports = {};",
				FileName: "file.cjs",
				TSConfig: "tsconfig.allow-js.json",
			},
			// ES and TypeScript import/export forms.
			{Code: "export default 42;"},
			{Code: "export default function () {}"},
			{Code: "export default class {}"},
			{Code: "export const value = 1;"},
			{Code: "export class C {}"},
			{Code: "export * from \"pkg\";"},
			{Code: "import type { T } from \"pkg\";"},
			{Code: "export type { T } from \"pkg\";"},
			{Code: "export type T = string;"},
			{Code: "export interface I {}"},
			{Code: "export enum E {}"},
			{Code: "export namespace N {}"},
			{Code: "export declare function f(): void;"},
			{Code: "export default interface I {}"},
			{Code: "export import Alias = require(\"pkg\");"},
			{Code: "namespace N {} export import Alias = N;"},
			{Code: "const value = {}; export = value;"},
			{
				Code:     "const view = <div />; export default view;",
				FileName: "file.tsx",
			},
			{
				Code:     "/** @type {number} */ const value = 1; export { value };",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Additional syntax, suppression, BOM, and line-ending regressions.
			{
				Code:     "declare global { interface Window { value: number } }",
				FileName: "global.d.ts",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 54)},
			},
			{
				Code:     "declare const value: number;",
				FileName: "types.d.mts",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 29)},
			},
			{
				Code:     "module.exports = {};",
				FileName: "file.cts",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 21)},
			},
			{
				Code:   "/* eslint-disable test */\n/* eslint-enable test */\nconst value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(3, 1, 3, 17)},
			},
			{
				Code:   "\uFEFFconst value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 17)},
			},
			{
				Code:     "\uFEFF// comment\nconst value = 1;",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 17)},
			},
			{
				Code:     "#!/usr/bin/env node\n",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 2, 1)},
			},
			{
				Code:     "/** @typedef {number} Tools.Value */\nconst value = 1;",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 17)},
			},
			{
				Code:     "/** @callback Handler\n * @param {string} value\n * @returns {void}\n */\nconst handler = value => {};",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(5, 1, 5, 29)},
			},
			{
				Code:     "// header\u2028  const value = \"🙂\";\u2029",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 2, 22)},
			},
			{
				Code:   "// header\u2028  const value = \"🙂\";\u2029",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 3, 1)},
			},
			// Source goals and extension defaults.
			{
				Code:            "module.exports = {};",
				FileName:        "file.cjs",
				TSConfig:        "tsconfig.allow-js.json",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 21)},
			},
			{
				Code:     "const value = 1;",
				FileName: "file.mjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 17)},
			},
			{
				Code:     "const value = 1;",
				FileName: "file.mts",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 17)},
			},
			// Module-like syntax outside upstream declaration kinds.
			{
				Code:   "\"use strict\"; function y() {}",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 30)},
			},
			{
				Code:   "module.exports = {}; exports.value = 1;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 40)},
			},
			{
				Code:   "import(\"pkg\");",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 15)},
			},
			{
				Code:   "const url = import.meta.url;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 29)},
			},
			{
				Code:   "await task();",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 14)},
			},
			{
				Code:   "import Alias = require(\"pkg\");",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 31)},
			},
			{
				Code:   "namespace N {} import Alias = N;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 33)},
			},
			{
				Code:   "export as namespace Library;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 29)},
			},
			{
				Code:   "type T = import(\"pkg\").T;",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 26)},
			},
			{
				Code:   "namespace Outer.Inner { export const value = 1; }",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 50)},
			},
			{
				Code:   "declare module \"pkg\" { export const value: number; }",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 53)},
			},
			{
				Code:     "const view = <div title=\"import\" />;",
				FileName: "file.tsx",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 37)},
			},
			{
				Code:     "/** @import { T } from \"pkg\" */\nconst value = 1;",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 17)},
			},
			{
				Code:     "/** @typedef {number} Value */\nconst value = 1;",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 17)},
			},
			{
				Code:     "/** @import { T } from \"pkg\" */\n",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 2, 1)},
			},
			{
				Code:     "const value = 1;\n/** @typedef {number} Value */",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 17)},
			},
			{
				Code:   "// export {}\nconst text = \"import x from 'pkg'\";",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 36)},
			},
			// Whole-Program ranges, including empty files and trivia.
			{
				Code:   "",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 1)},
			},
			{
				Code:   "   ",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(1, 4, 1, 4)},
			},
			{
				Code:   "// only\n",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 1, 2, 1)},
			},
			{
				Code:   "// head\n  f(); // tail\n",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 3, 1)},
			},
			{
				Code:   "#!/usr/bin/env node\r\n  const café = \"🙂\";\r\n",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 3, 1)},
			},
			{
				Code:   "/* header */\n  const café = \"🙂\"; // tail",
				Errors: []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 2, 29)},
			},
			{
				Code:     "",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 1)},
			},
			{
				Code:     "   ",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 4)},
			},
			{
				Code:     "// only\n",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 2, 1)},
			},
			{
				Code:     "// head\n  f(); // tail\n",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 2, 7)},
			},
			{
				Code:     "/* header */\n  const café = \"🙂\"; // tail",
				FileName: "file.js",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(2, 3, 2, 21)},
			},
			{
				Code:     "const view = <div />; // tail\n",
				FileName: "file.jsx",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{unambiguousError(1, 1, 1, 22)},
			},
		},
	)
}
