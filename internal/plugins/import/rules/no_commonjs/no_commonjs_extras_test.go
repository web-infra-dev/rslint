package no_commonjs_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_commonjs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCommonjsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.allow-js.json", t, &no_commonjs.NoCommonjsRule,
		[]rule_tester.ValidTestCase{
			// A catch binding is visible while its destructuring default is evaluated.
			{
				Code: `try {} catch ({exports = exports.x}) {}`,
			},
			// The exemption also applies to a read directly on the right of a destructuring assignment.
			{
				Code:    `({ value } = module.exports); [value] = module.exports;`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
			{
				Code:    `if (ok) require("x"); module.exports = () => 1`,
				Options: []any{map[string]any{"allowPrimitiveModules": true, "allowRequire": true, "allowConditionalRequire": false}},
			},
			// Only the innermost scope is inspected for exports declarations.
			{
				Code: `const exports = {}; exports.x = 1`,
			},
			{
				Code: `{ const exports = {}; exports.x = 1; }`,
			},
			{
				Code: `function f(exports) { exports.x = 1; }`,
			},
			{
				Code: `function f({exports}) { exports.x = 1; }`,
			},
			{
				Code: `exports.x = 1; var exports;`,
			},
			{
				Code: `import { exports } from "x"; exports.x = 1;`,
			},
			{
				Code: `function f(x = require("x")) { require("y"); } (() => require("z"))();`,
			},
			{
				Code: `class C { x = require("field"); static x = require("static"); static { require("block"); } method() { require("method"); } }`,
			},
			// Every ancestor conditional counts, including the test, catch and finally clauses.
			{
				Code: `if (require("test")) {} else require("else"); require("left") || fallback; value ?? require("right"); require("condition") ? require("yes") : require("no");`,
			},
			{
				Code: `try {} catch (e) { require("catch"); } finally { require("finally"); }`,
			},
			{
				Code: `(require as any)("x"); require!("x"); require("x" as string); (module as any).exports = 1; module[exports!];`,
			},
			{
				Code: `require(); require(...["x"]); require(null); require(/x/); new require("x"); import x = require("x");`,
			},
			{
				Code: "module[\"exports\"] = 1; module[`exports`] = 2; module[other] = 3; other.exports = 4;",
			},
			{
				Code:    `target = module.exports; (module.exports) = (function () {}); module.exports += value; module.exports = []; module.exports = null;`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
			{
				Code:    `module.exports = {} as object;`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
			},
			// JSX tag names and ordinary type names are not MemberExpression nodes.
			{
				Code:     `const el = <module.exports><exports.Item /></module.exports>;`,
				FileName: "element.tsx",
			},
			{
				Code: `type T = module.exports; type U = typeof exports.item;`,
			},
			{
				Code: `type exports = {}; exports.x = 1;`,
			},
			// Configured and inline globals suppress exports only in an actual global scope.
			{
				Code:            `require("x")`,
				FileName:        "script.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Options:         []any{map[string]any{"allowConditionalRequire": false}},
			},
			{
				Code:            `exports.x = 1;`,
				FileName:        "script.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"exports": "readonly"},
			},
			{
				Code:            `/* global exports */ exports.x = 1;`,
				FileName:        "script.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			{
				Code:            `exports.x = 1;`,
				FileName:        "common.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Escaped identifiers use their decoded names.
			{
				Code: strings.NewReplacer("i", `\u0069`, "o", `\u006f`).Replace(`require('x'); module.exports = {}; exports.x = 1;`),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 44},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 51, EndLine: 1, EndColumn: 65},
				},
			},
			// Decorators on classes and members run outside the method; parameter decorators stay inside it.
			{
				Code: `class C { @require("field") field = 1; @require("method") method(@require("param") p) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: importMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
				},
			},
			{
				Code: `@require("class") class C {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 9},
				},
			},
			// A named function expression binds its name outside its function scope.
			{
				Code: `const f = function exports() { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 32, EndLine: 1, EndColumn: 41},
				},
			},
			// A var declaration outside the loop body does not declare exports in the body block.
			{
				Code: `for (var exports of list) { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
				},
			},
			// TypeScript wrappers around the member interrupt the primitive assignment exemption.
			{
				Code:    `(module.exports! as any) = {};`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 16},
				},
			},
			// Members inside assignment patterns do not have the assignment as their immediate parent.
			{
				Code:    `[module.exports] = values; ({value: module.exports} = values);`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 2, EndLine: 1, EndColumn: 16},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 51},
				},
			},
			// JavaScript JSDoc casts preserve an object literal assignment.
			{
				Code:     `module.exports = /** @type {Object} */ ({});`,
				FileName: "case.js",
				Options:  []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
				},
			},
			// A TypeScript instantiation wrapper around the callee is distinct from call type arguments.
			{
				Code: `require<string>("x"); (require<string>)("x");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			// Class field and static block scopes do not own the surrounding class name.
			{
				Code: `class exports { field = exports.x; static { exports.x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 45, EndLine: 1, EndColumn: 54},
				},
			},
			// Omitted options and explicit defaults have the same behavior.
			{
				Code:    `require("x"); module.exports = "x"; exports.x = 1`,
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 29},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    `require("x"); module.exports = "x"; exports.x = 1`,
				Options: []any{map[string]any{"allowPrimitiveModules": false, "allowRequire": false, "allowConditionalRequire": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 29},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 46},
				},
			},
			{
				Code:    `module.exports = {}; exports.x = 1; require("x")`,
				Options: []any{map[string]any{"allowPrimitiveModules": true, "allowRequire": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: `const exports = {}; { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 32},
				},
			},
			{
				Code: `function f(exports) { { exports.x = 1; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code: `try {} catch (exports) { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 26, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `function f() { exports.x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: `function f(module) { module.exports = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 36},
				},
			},
			{
				Code: `const require = fn; require("x")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 28},
				},
			},
			// require is checked in the module variable scope, including blocks and computed method names.
			{
				Code: `{ require("x"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 3, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code: `class C { [require("method")]() {} [require("field")] = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: importMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code: `const obj = { [require("key")]() {} };`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `class C extends require("base") {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `try {} catch (e) { require("catch"); } finally { require("finally"); }`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
					{MessageId: "", Message: importMessage, Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
				},
			},
			{
				Code:    `ok && require("and"); ok || require("or"); ok ?? require("nullish"); ok ? require("yes") : require("no");`,
				Options: []any{map[string]any{"allowConditionalRequire": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 7, EndLine: 1, EndColumn: 14},
					{MessageId: "", Message: importMessage, Line: 1, Column: 29, EndLine: 1, EndColumn: 36},
					{MessageId: "", Message: importMessage, Line: 1, Column: 50, EndLine: 1, EndColumn: 57},
					{MessageId: "", Message: importMessage, Line: 1, Column: 75, EndLine: 1, EndColumn: 82},
					{MessageId: "", Message: importMessage, Line: 1, Column: 92, EndLine: 1, EndColumn: 99},
				},
			},
			{
				Code: `value ||= require("or"); value &&= require("and"); value ??= require("nullish");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
					{MessageId: "", Message: importMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 43},
					{MessageId: "", Message: importMessage, Line: 1, Column: 62, EndLine: 1, EndColumn: 69},
				},
			},
			{
				Code: `while (ok) require("while"); for (;;) { require("for"); break; } switch (x) { case 1: require("switch"); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
					{MessageId: "", Message: importMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
					{MessageId: "", Message: importMessage, Line: 1, Column: 87, EndLine: 1, EndColumn: 94},
				},
			},
			// Parentheses and JS JSDoc casts are transparent; authored TS assertions are not.
			{
				Code: `((require))((("x"))); (module).exports = ({ a: 1 }); (exports)[key] = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 3, EndLine: 1, EndColumn: 10},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 23, EndLine: 1, EndColumn: 39},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 54, EndLine: 1, EndColumn: 68},
				},
			},
			{
				Code:     `(/** @type {*} */ (require))("x"); (/** @type {*} */ (module)).exports = 1;`,
				FileName: "casts.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 36, EndLine: 1, EndColumn: 71},
				},
			},
			// Computed identifier names match; string and template property values do not.
			{
				Code: `module[exports] = 1; module[(exports)] = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 22, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code: "exports[\"x\"] = 1; exports[key] = 2; exports[`x`] = 3;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 19, EndLine: 1, EndColumn: 31},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 37, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code: `class C { #exports; read(module) { return module.#exports; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 43, EndLine: 1, EndColumn: 58},
				},
			},
			{
				Code: `class C { #x; read(exports) { { return exports.#x; } } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 40, EndLine: 1, EndColumn: 50},
				},
			},
			// Optional calls and members still match, but ChainExpression interrupts the primitive assignment exemption.
			{
				Code: `require?.("x"); module?.exports; exports?.x;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 17, EndLine: 1, EndColumn: 32},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 34, EndLine: 1, EndColumn: 44},
				},
			},
			{
				Code:    `target = module?.exports;`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 10, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `module.exports = ({}); module.exports ||= {}; use(module.exports); module.exports.x = 1;`,
				Options: []any{map[string]any{"allowPrimitiveModules": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 15},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 38},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 51, EndLine: 1, EndColumn: 65},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 68, EndLine: 1, EndColumn: 82},
				},
			},
			{
				Code:     `const el = <div value={module.exports}>{exports.item}</div>;`,
				FileName: "element.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 24, EndLine: 1, EndColumn: 38},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 41, EndLine: 1, EndColumn: 53},
				},
			},
			{
				Code: `interface I extends module.exports {} class C implements exports.Item {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 21, EndLine: 1, EndColumn: 35},
					{MessageId: "", Message: exportMessage, Line: 1, Column: 58, EndLine: 1, EndColumn: 70},
				},
			},
			{
				Code: `namespace N { require("x"); } enum E { A = require("y") }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code:            `require("x"); exports.x = 1;`,
				FileName:        "common.cjs",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:            `exports.x = 1;`,
				FileName:        "script.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"exports": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:            `{ exports.x = 1; }`,
				FileName:        "script.js",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Globals:         map[string]any{"exports": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 3, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `exports.x = 1;`,
				Globals: map[string]any{"exports": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:            `exports.x = 1;`,
				FileName:        "common.ts",
				LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
				Globals:         map[string]any{"exports": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: exportMessage, Line: 1, Column: 1, EndLine: 1, EndColumn: 10},
				},
			},
			// Diagnostic ranges use UTF-16 columns and include the whole member, even across lines.
			{
				Code: `"😀";
  (require)("x");
  module
    .exports = {};
"😀"; exports.名 = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "", Message: importMessage, Line: 2, Column: 4, EndLine: 2, EndColumn: 11},
					{MessageId: "", Message: exportMessage, Line: 3, Column: 3, EndLine: 4, EndColumn: 13},
					{MessageId: "", Message: exportMessage, Line: 5, Column: 7, EndLine: 5, EndColumn: 16},
				},
			},
		},
	)
}
