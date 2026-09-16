package exports_style

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional scope, AST and edit cases checked with eslint-plugin-n v18.3.0,
// ESLint 10.9.0 and @typescript-eslint/parser 8.65.0.
func TestExportsStyleExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ExportsStyleRule, []rule_tester.ValidTestCase{ // unconfigured globals
		{Code: "exports.x = 1; module.exports.x = 2;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "off", "module": "off"}, Options: []any{}},
		// inline globals off
		{Code: "/* global exports: off */ exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}},
		// disabled module with exports option
		{Code: "module.exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "off"}, Options: []any{"exports"}},
		// module-local binding
		{Code: "var exports = {}; exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}},
		// module source local binding
		{Code: "const module = {}; module.exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}},
		// nested shadowing
		{Code: "function f(exports) { exports.x = 1; } { let exports; exports.x = 2; } try {} catch (exports) { exports.x = 3; }", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}},
		// module binding shadowed
		{Code: "function f(module) { module.exports = {}; }", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}},
		// assignment patterns do not count as assignments
		{Code: "({value: exports = {}} = obj); [exports] = items;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}},
		// parentheses in batch assignment
		{Code: "((module).exports) = ((exports) = {});", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}},
		// compound batch assignment
		{Code: "exports ||= module.exports &&= {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}},
		// batch property assignments
		{Code: "module.exports.x = exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}},
		// repeated exports assignments
		{Code: "module.exports = exports = exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}},
		// repeated module assignments in module mode
		{Code: "exports = module.exports = module.exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"module.exports", map[string]any{"allowBatchAssign": true}}},
		// computed dynamic and private names
		{Code: "module[exports].x = 1; module[`ex${part}ports`].x = 2; class C { #exports; m() { module.#exports.x = 3; } }", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}},
		// disabled diagnostics
		{Code: "// eslint-disable-next-line\nmodule.exports = { a: 1 };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}},
		// typed literal access
		{Code: "module[\"exports\" as const].x; module[`exports`!].x;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": false}}},
		// type-only import
		{Code: "import type {exports} from \"x\"; exports.x;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"module.exports", map[string]any{"allowBatchAssign": false}}},
		// module type-only import
		{Code: "import type {module} from \"x\"; module.exports.x;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "module"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": false}}},
	}, []rule_tester.InvalidTestCase{
		// Legacy script for-in initializers and iteration each write the binding.
		{Code: "for (var exports = 1 in obj) {}", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
			{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
		}},
		// inline globals on
		{Code: "/* global exports: writable */ exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "off", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 32, EndLine: 1, EndColumn: 40}}},
		// script global declaration
		{Code: "var exports = {}; exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 5, EndLine: 1, EndColumn: 14}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27}}},
		// script declaration without configured globals
		{Code: "let exports; exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "off", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22}}},
		// script destructuring initializer
		{Code: "var {x: exports = {}} = obj; exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 9, EndLine: 1, EndColumn: 18}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 38}}},
		// script uninitialized loop binding
		{Code: "for (var exports of items) { exports.x = 1; }", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 20}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 30, EndLine: 1, EndColumn: 38}}},
		// script module declaration
		{Code: "var module = {}; module.exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 18, EndLine: 1, EndColumn: 33}}, Output: []string{"var module = {}; exports.x = 1;"}},
		// default parameter evaluates outside body binding
		{Code: "function f(a = exports.x) { var exports; }", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24}}},
		// object references and update
		{Code: "const data = {exports}; exports++; typeof exports;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 15, EndLine: 1, EndColumn: 23}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 43, EndLine: 1, EndColumn: 51}}},
		// destructuring assignment references
		{Code: "({value: exports = {}} = obj); [exports] = items;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 33, EndLine: 1, EndColumn: 41}}},
		// runtime assignment nested in default
		{Code: "[target = (exports = {})] = items;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedAssignment", Message: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.", Line: 1, Column: 12, EndLine: 1, EndColumn: 21}}},
		// default pattern cannot exempt runtime assignment
		{Code: "[module.exports = (exports = {})] = items;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 2, EndLine: 1, EndColumn: 18}, {MessageId: "unexpectedAssignment", Message: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29}}},
		// asymmetric repeated module assignment
		{Code: "module.exports = module.exports = exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 18, EndLine: 1, EndColumn: 34}}},
		// sequence breaks batch assignment
		{Code: "module.exports = (0, exports = {});", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}, {MessageId: "unexpectedAssignment", Message: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.", Line: 1, Column: 22, EndLine: 1, EndColumn: 31}}},
		// computed static names
		{Code: "module[\"exports\"].x = 1; module[`exports`].x = 2; module[(\"exports\")].x = 3;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 26, EndLine: 1, EndColumn: 44}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 51, EndLine: 1, EndColumn: 71}}, Output: []string{"exports.x = 1; exports.x = 2; exports.x = 3;"}},
		// escaped identifier names
		{Code: "mod\\u0075le.ex\\u0070orts.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}}, Output: []string{"exports.x = 1;"}},
		// optional members
		{Code: "module?.exports.foo; module.exports?.foo; (module?.exports).foo; (module.exports)?.foo;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 22, EndLine: 1, EndColumn: 38}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 44, EndLine: 1, EndColumn: 60}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 67, EndLine: 1, EndColumn: 82}}, Output: []string{"exports.foo; exports?.foo; (module?.exports).foo; (exports)?.foo;"}},
		// optional exports references
		{Code: "exports?.foo; (exports)?.foo;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24}}},
		// read references and final token
		{Code: "const a = module.exports; module.exports", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 11, EndLine: 1, EndColumn: 26}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 27, EndLine: 1, EndColumn: 41}}},
		// read exports at EOF
		{Code: "exports", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
		// multiline range with comment
		{Code: "/* 😀 */\nexports /* range */\n.foo = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 2, Column: 1, EndLine: 3, EndColumn: 2}}},
		// UTF-16 range
		{Code: "const 文 = \"😀\"; exports.foo = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25}}},
		// empty object removal
		{Code: "module.exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{";"}},
		// parenthesized object and values
		{Code: "(module.exports) = ({ a: (1), b: (function () {}) });", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 2, EndLine: 1, EndColumn: 17}}, Output: []string{"exports.a = 1;\n\nexports.b = function () {};;"}},
		// computed and numeric object keys
		{Code: "module.exports = { [(name)]: (value), 1: true, \"x-y\": null };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"exports[name] = value;\n\nexports[1] = true;\n\nexports[\"x-y\"] = null;;"}},
		// async generator method
		{Code: "module.exports = { async *items(x) { yield x; } };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"exports.items = async function* (x) { yield x; };;"}},
		// comments at property boundaries
		{Code: "module.exports = { /* first */ a: 1 /* after a */, /* second */ b: 2, /* trailing */ };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"/* first */\nexports.a = 1;\n/* after a */\n\n/* second */\nexports.b = 2;;"}},
		// comments in method signature
		{Code: "module.exports = { run /* before params */ ( /* param */ x ) { return x; } };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"exports.run = function ( /* param */ x ) { return x; };;"}},
		// function property and shorthand
		{Code: "module.exports = { named: async function f() {}, value };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"exports.named = async function f() {};\n\nexports.value = value;;"}},
		// unsupported accessor cancels whole fix
		{Code: "module.exports = { a: 1, get b() {} };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
		// top-level compound object assignment
		{Code: "module.exports += { a: 1 };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}, Output: []string{"exports.a = 1;;"}},
		// nested assignment is not expanded
		{Code: "const result = module.exports = { a: 1 };", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32}}},
		// JSX names versus expression members
		{Code: "const view = <module.exports value={module.exports.foo} />;", FileName: "input.jsx", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 52}}, Output: []string{"const view = <module.exports value={exports.foo} />;"}},
		// JSX exports object reference
		{Code: "const view = <exports.Widget value={exports.foo} />;", FileName: "input.jsx", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 45}}},
		// TypeScript wrappers preserve semantics
		{Code: "(module as any).exports.foo; module!.exports.foo; module.exports!.foo; (module.exports as any).foo;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 51, EndLine: 1, EndColumn: 66}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 73, EndLine: 1, EndColumn: 90}}},
		// TypeScript wrapper blocks batch allowance
		{Code: "module.exports = ((exports = {}) as object);", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{"allowBatchAssign": true}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}, {MessageId: "unexpectedAssignment", Message: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.", Line: 1, Column: 20, EndLine: 1, EndColumn: 29}}},
		// TypeScript method signature
		{Code: "module.exports = { async run<T>(x: T): Promise<T> { return x; } };", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}, Output: []string{"exports.run = async function <T>(x: T): Promise<T> { return x; };;"}},
		// TypeScript type and runtime names
		{Code: "type X = exports; type Y = typeof exports; interface Z extends module.exports {} const value = module.exports;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 64, EndLine: 1, EndColumn: 80}, {MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 96, EndLine: 1, EndColumn: 111}}},
		// TypeScript type references to exports
		{Code: "type X = exports; type Y = typeof exports;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 10, EndLine: 1, EndColumn: 18}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 35, EndLine: 1, EndColumn: 43}}},
		// TypeScript type-only shadow
		{Code: "function f() { type exports = object; exports.x = 1; }", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 39, EndLine: 1, EndColumn: 47}}},
		// TypeScript namespace shadow
		{Code: "namespace exports { export const x = 1; } exports.x;", FileName: "input.ts", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 43, EndLine: 1, EndColumn: 51}}},
		// JavaScript JSDoc does not declare a binding
		{Code: "/** @typedef {object} exports */ exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 34, EndLine: 1, EndColumn: 42}}},
		// JavaScript JSDoc wrapper
		{Code: "/** @type {object} */ (module).exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 23, EndLine: 1, EndColumn: 40}}, Output: []string{"/** @type {object} */ exports.x = 1;"}},
		// explicit default options
		{Code: "exports.x = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"module.exports", map[string]any{"allowBatchAssign": false}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
		// empty object options
		{Code: "exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports", map[string]any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedAssignment", Message: "Unexpected assignment to 'exports'. Don't modify 'exports' itself.", Line: 1, Column: 1, EndLine: 1, EndColumn: 10}}},
		// shorthand assignment default
		{Code: "({exports = {}} = obj);", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 3, EndLine: 1, EndColumn: 12}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 3, EndLine: 1, EndColumn: 12}}},
		// nested pattern defaults
		{Code: "([{ value: exports = 1 } = {}] = items);", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 12, EndLine: 1, EndColumn: 21}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 12, EndLine: 1, EndColumn: 21}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 12, EndLine: 1, EndColumn: 21}}},
		// nested declaration defaults
		{Code: "var [{ value: exports = 1 } = {}] = items;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}, {MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}}},
		// computed pattern key is a read
		{Code: "({ [exports]: target = {} } = obj);", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 5, EndLine: 1, EndColumn: 13}}},
		// module key in member access
		{Code: "obj[module].exports = 1; module[module.exports] = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{"exports"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedModuleExports", Message: "Unexpected access to 'module.exports'. Use 'exports' instead.", Line: 1, Column: 33, EndLine: 1, EndColumn: 48}}},
		// TypeScript JSX member references
		{Code: "const view = <exports.Widget value={exports.foo} />;", FileName: "input.tsx", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Globals: map[string]any{"exports": "writable", "module": "readonly"}, Options: []any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedExports", Message: "Unexpected access to 'exports'. Use 'module.exports' instead.", Line: 1, Column: 37, EndLine: 1, EndColumn: 45}}}})
}

func TestExportsStyleEditDemand(t *testing.T) {
	for _, test := range []struct {
		code    string
		mode    string
		fixable bool
	}{
		{"module.exports = { /* keep */ async *items(x) { yield x; } };", "exports", true},
		{"module.exports.foo = 1;", "exports", true},
		{"module.exports = { get foo() {} };", "exports", false},
		{"exports = {};", "exports", false},
		{"exports.foo = 1;", "module.exports", false},
	} {
		t.Run(test.code, func(t *testing.T) {
			program, file, err := rule_tester.NewProgramHelper(fixtures.GetRootDir()).CreateTestProgram(test.code, "input.cjs", "tsconfig.allowJs.json")
			if err != nil {
				t.Fatal(err)
			}
			run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
				var diagnostics []rule.RuleDiagnostic
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: file.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{Name: ExportsStyleRule.Name, Severity: rule.SeverityError,
							Environment: &rule.RuleEnvironment{LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},
							Run:         func(ctx rule.RuleContext) rule.RuleListeners { return ExportsStyleRule.Run(ctx, []any{test.mode}) },
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
				})
				return diagnostics
			}
			all := run(rule.EditDemandAll)
			if len(all) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(all))
			}
			if (all[0].FixesPtr != nil) != test.fixable {
				t.Fatalf("unexpected fix availability: %#v", all[0])
			}
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				got := run(demand)
				if len(got) != 1 {
					t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
				}
				if got[0].Range != all[0].Range || !reflect.DeepEqual(got[0].Message, all[0].Message) {
					t.Fatalf("demand %d changed diagnostic identity", demand)
				}
				if got[0].Suggestions != nil {
					t.Fatal("unexpected suggestions")
				}
				if demand&rule.EditDemandAutofix != 0 {
					if !reflect.DeepEqual(got[0].FixesPtr, all[0].FixesPtr) {
						t.Fatal("fix changed with demand")
					}
				} else if got[0].FixesPtr != nil {
					t.Fatal("fix created without autofix demand")
				}
			}
		})
	}
}

func TestExportsStyleSchema(t *testing.T) {
	for _, options := range [][]any{{"invalid"}, {true}, {"exports", true}, {"exports", map[string]any{"allowBatchAssign": "true"}}, {"exports", map[string]any{"unknown": true}}, {"exports", map[string]any{}, false}} {
		if err := ExportsStyleRule.Schema.Validate(options); err == nil {
			t.Errorf("expected invalid options: %#v", options)
		}
	}
}
