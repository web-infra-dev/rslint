package prefer_module_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_module"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Expectations were checked against eslint-plugin-unicorn v79.0.0, except
// the documented suggestions withheld to avoid invalid syntax or comment loss.
func TestPreferModuleExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		{Code: "require(\"x\"); module.exports = exports; __dirname; __filename;", FileName: "extras.js", Globals: map[string]any{"exports": "off", "require": "off", "module": "off", "__filename": "off", "__dirname": "off"}},
		{Code: "function f(require, exports, module, __dirname, __filename) { require(\"x\"); exports.x = module; return __dirname + __filename; }", FileName: "extras.js"},
		{Code: "{ let require; require(\"x\"); } try {} catch (module) { module.exports = 1; }", FileName: "extras.js"},
		{Code: "const object = {module: 1, exports() {}, require: 2, __dirname: 3}; object.__filename;", FileName: "extras.js"},
		{Code: "class C { #module; module() { return this.#module; } require = 1; }", FileName: "extras.js"},
		{Code: "module: for (;;) { break module; }", FileName: "extras.js"},
		{Code: "const view = <module.exports require=\"__dirname\" />;", FileName: "extras.tsx"},
		{Code: "foo(); \"use strict\";", FileName: "extras.js"},
		{Code: "{\"use strict\";}", FileName: "extras.js"},
		{Code: "\"use\\x20strict\";", FileName: "extras.js"},
		{Code: "function f() { (\"use strict\"); }", FileName: "extras.js"},
		{Code: "class C { static { \"use strict\"; } }", FileName: "extras.js"},
		{Code: "function f(){return;} const g=()=>{return 1}; class C {get x(){return 1} f(){return 2}}", FileName: "extras.js"},
		{Code: "var require; require(\"a\");", FileName: "extras.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},

		// Review cases: tags, type references, JSX, escaped identifiers, and comments.
		{Code: "/** @type {typeof module} */ const x = 1;", FileName: "review.js"},
		{Code: "/** @type {typeof __dirname} */ const x = 1;", FileName: "review.js"},
		{Code: "const view = <module:exports />;", FileName: "review.tsx"},
		{Code: "const view = <__dirname />;", FileName: "review.tsx"},
	}
	invalid := []rule_tester.InvalidTestCase{
		{Code: "type Registry = typeof module; interface T {__dirname: string} type U = {__filename: string};", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
			}},
		{Code: "namespace N { \"use strict\"; }", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 15, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/use-strict-directive", Output: "namespace N { }"},
					}},
			}},
		{Code: "class C { static { \"use strict\"; } }", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 20, EndLine: 1, EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/use-strict-directive", Output: "class C { static { } }"},
					}},
			}},
		{Code: "\"use asm\"; \"use strict\"; \"use strict\"; run();", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 12, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/use-strict-directive", Output: "\"use asm\"; \"use strict\"; run();"},
					}},
				{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/use-strict-directive", Output: "\"use asm\"; \"use strict\"; run();"},
					}},
			}},
		{Code: "\"use strict\";\u00a0\u2028/* keep */run();", FileName: "extras.mjs", Output: []string{"/* keep */run();"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
			}},
		{Code: "if (ok) { /* return token */ return 1; }", FileName: "extras.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/global-return", Message: "\"return\" should be used inside a function.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
			}},
		{Code: "const object = {__dirname /* keep */, __filename};", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "const object = {__dirname: import.meta.dirname /* keep */, __filename};"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "const object = {__dirname: path.dirname(url.fileURLToPath(import.meta.url)) /* keep */, __filename};"},
					}},
				{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 39, EndLine: 1, EndColumn: 49,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-filename", Output: "const object = {__dirname /* keep */, __filename: import.meta.filename};"},
						{MessageId: "suggestion/import-meta-url-to-filename", Output: "const object = {__dirname /* keep */, __filename: url.fileURLToPath(import.meta.url)};"},
					}},
			}},
		{Code: "({__dirname = \"fallback\"} = object);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 3, EndLine: 1, EndColumn: 12,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "({__dirname: import.meta.dirname = \"fallback\"} = object);"},
					}},
			}},
		{Code: "({__filename} = object);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 3, EndLine: 1, EndColumn: 13,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-filename", Output: "({__filename: import.meta.filename} = object);"},
					}},
			}},
		{Code: "const 日本語 = \"😀\";\nconsole.log(__dirname);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 2, Column: 13, EndLine: 2, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "const 日本語 = \"😀\";\nconsole.log(import.meta.dirname);"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "const 日本語 = \"😀\";\nconsole.log(path.dirname(url.fileURLToPath(import.meta.url)));"},
					}},
			}},
		{Code: "const emoji = \"😀\"; __filename;", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 21, EndLine: 1, EndColumn: 31,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-filename", Output: "const emoji = \"😀\"; import.meta.filename;"},
						{MessageId: "suggestion/import-meta-url-to-filename", Output: "const emoji = \"😀\"; url.fileURLToPath(import.meta.url);"},
					}},
			}},
		{Code: "new (__dirname.constructor)(\"x\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname.constructor)(\"x\");"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url)).constructor)(\"x\");"},
					}},
			}},
		{Code: "new __filename[__dirname](\"x\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-filename", Output: "new import.meta.filename[__dirname](\"x\");"},
						{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url))[__dirname](\"x\");"},
					}},
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 16, EndLine: 1, EndColumn: 25,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "new __filename[import.meta.dirname](\"x\");"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new __filename[path.dirname(url.fileURLToPath(import.meta.url))](\"x\");"},
					}},
			}},
		{Code: "new (__dirname as any).constructor(\"x\");", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname as any).constructor(\"x\");"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url)) as any).constructor(\"x\");"},
					}},
			}},
		{Code: "new __filename<string>();", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename)<string>();"},
						{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url))<string>();"},
					}},
			}},
		{Code: "require(\"a\",);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "const value = require(\"a\",);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22},
			}},
		{Code: "const file = /** @type {string} */ (require(\"a\"));", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 37, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from /** @type {string} */ \"a\";"},
					}},
			}},
		{Code: "const file = (/** @type {Function} */ (require))(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 40, EndLine: 1, EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from /** @type {Function} */ \"a\";"},
					}},
			}},
		{Code: "const file = require(/** @type {string} */ (\"a\"));", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from /** @type {string} */ \"a\";"},
					}},
			}},
		{Code: "const file = require(/* lead */ ((\"a\" /* trail */)));", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from /* lead */ \"a\" /* trail */;"},
					}},
			}},
		{Code: "const { foo /* key */ : /* colon */ bar } /* binding */ = /* equal */ require(/* source */ \"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 71, EndLine: 1, EndColumn: 78,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import { foo/* key */  as /* colon */ bar }/* binding */  from /* equal */ /* source */ \"a\";"},
					}},
			}},
		{Code: "require?.(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "(require)?.(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 2, EndLine: 1, EndColumn: 9},
			}},
		{Code: "const x = require(1);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const x = require(null);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const x = require(/a/);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const [x] = require(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
			}},
		{Code: "const {x = 1} = require(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
			}},
		{Code: "const {x: {y}} = require(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
			}},
		{Code: "const {...rest} = require(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
			}},
		{Code: "const {\"x\": x} = require(\"a\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
			}},
		{Code: "const x = require(\"a.JSON\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import x from \"a.JSON\";"},
					}},
			}},
		{Code: "const x = require(\"a.json?query\");", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import x from \"a.json?query\";"},
					}},
			}},
		{Code: "const x = require<string>(\"a\");", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "export const x = require(\"a\");", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
			}},
		{Code: "const x = (require as Function)(\"a\");", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 12, EndLine: 1, EndColumn: 19},
			}},
		{Code: "const x = require!(\"a\");", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const x = require(\"a\") as X;", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const x = require(\"a\" as string);", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
			}},
		{Code: "const file = require(`./😀.js`);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from './😀.js';"},
					}},
			}},
		{Code: "const file = require(`./\\uD800.js`);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from './\\u{d800}.js';"},
					}},
			}},
		{Code: "const file = require(`./\\u0000\\u0001\\u007f\\u2028\\u2029.js`);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from './\\u{0}\\u{1}\\u{7f}\\u{2028}\\u{2029}.js';"},
					}},
			}},
		{Code: "const file = require(`./\\t\\r\\b\\f\\v\\\\.js`);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import file from './\\t\\r\\b\\f\\v\\\\.js';"},
					}},
			}},
		{Code: "exports /* keep */ = /* rhs */ (foo);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/export", Output: "export default /* keep */ /* rhs */ (foo);"},
					}},
			}},
		{Code: "module /* one */ .exports /* two */ = /* three */ (() => 1);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			}},
		{Code: "((exports.foo)) = ((bar));", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 3, EndLine: 1, EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/export", Output: "export {bar as foo};"},
					}},
			}},
		{Code: "exports.foo = /* rhs */ bar;", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "(/* before */ exports.foo = bar /* after */);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/export", Output: "/* before */ export {bar as foo} /* after */;"},
					}},
			}},
		{Code: "module.exports?.foo;", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			}},
		{Code: "module[\"exports\"].foo = bar;", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			}},
		{Code: "exports.foo ||= bar;", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "exports.foo = (bar as any);", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "exports.foo = (foo, bar);", FileName: "extras.js",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
			}},
		{Code: "require(\"a\"); module.exports = 1; __dirname;", FileName: "extras.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import", Output: "import \"a\"; module.exports = 1; __dirname;"},
					}},
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/export", Output: "require(\"a\"); export default 1; __dirname;"},
					}},
				{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 35, EndLine: 1, EndColumn: 44,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/import-meta-dirname", Output: "require(\"a\"); module.exports = 1; import.meta.dirname;"},
						{MessageId: "suggestion/import-meta-url-to-dirname", Output: "require(\"a\"); module.exports = 1; path.dirname(url.fileURLToPath(import.meta.url));"},
					}},
			}},
		{Code: "type module = number; module.exports = foo;", FileName: "extras.ts",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 23, EndLine: 1, EndColumn: 29,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestion/export", Output: "type module = number; export default foo;"},
					}},
			}},

		// Review cases checked against the pinned upstream implementation.
		{Code: "__dirname`text`;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-dirname", Output: "import.meta.dirname`text`;"},
				{MessageId: "suggestion/import-meta-url-to-dirname", Output: "path.dirname(url.fileURLToPath(import.meta.url))`text`;"},
			}},
		}},
		{Code: "__filename++;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-filename", Output: "import.meta.filename++;"},
			}},
		}},
		{Code: "for (__dirname of paths) {}", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-dirname", Output: "for (import.meta.dirname of paths) {}"},
			}},
		}},
		{Code: "type T = typeof __dirname; type U = typeof __filename;", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 26},
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 44, EndLine: 1, EndColumn: 54},
		}},
		{Code: "new __dirname`text`;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-dirname", Output: "new import.meta.dirname`text`;"},
				{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url)))`text`;"},
			}},
		}},
		{Code: "new (__filename)`text`;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename)`text`;"},
				{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url))`text`;"},
			}},
		}},
		{Code: "new __filename.tag`text`;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-filename", Output: "new import.meta.filename.tag`text`;"},
				{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url)).tag`text`;"},
			}},
		}},
		{Code: "new __filename!.constructor();", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename)!.constructor();"},
				{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url))!.constructor();"},
			}},
		}},
		{Code: "new __dirname<string>`text`;", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname)<string>`text`;"},
				{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url)))<string>`text`;"},
			}},
		}},
		{Code: "type T = module; interface X extends module.exports {}", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
			{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
		}},
		{Code: "type T = typeof module; type U = typeof require;", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 41, EndLine: 1, EndColumn: 48},
		}},
		{Code: "function f(value: __dirname): __filename {}", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 28},
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 31, EndLine: 1, EndColumn: 41},
		}},
		{Code: "const moduleName = \\u0072\\u0065\\u0071\\u0075\\u0069\\u0072\\u0065(\"x\");", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 20, EndLine: 1, EndColumn: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import", Output: "import moduleName from \"x\";"},
			}},
		}},
		{Code: "const value = 1; (exports /* middle */ .value = value);", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
		}},
		{Code: "/* before */ exports.value = value; /* after */", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/export", Output: "/* before */ export {value as value}; /* after */"},
			}},
		}},
		{Code: "const text = \"/* no comment */\"; exports.value = value;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 34, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/export", Output: "const text = \"/* no comment */\"; export {value as value};"},
			}},
		}},
		{Code: "const expression = /\\/\\*not-a-comment\\*\\//; exports.value = value;", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 45, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/export", Output: "const expression = /\\/\\*not-a-comment\\*\\//; export {value as value};"},
			}},
		}},
		{Code: "type T = __dirname; type U = __filename;", FileName: "review.ts", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
			{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 30, EndLine: 1, EndColumn: 40},
		}},
		{Code: "\"use strict\"; // keep\nrun();", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/use-strict-directive", Output: "// keep\nrun();"},
			}},
		}},
		{Code: "\"use strict\"; // keep\nrun();", FileName: "review.mjs", Output: []string{"// keep\nrun();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
		}},
		{Code: "require(\"a.json\",);", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
		}},
		{Code: "const {module: exported} = require(\"a\");", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 28, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import", Output: "import {module as exported} from \"a\";"},
			}},
		}},
		{Code: "class C { #exports; f() { return module.exports; } }", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 34, EndLine: 1, EndColumn: 40},
		}},

		// A comment between the literal and semicolon belongs to the directive.
		{Code: `"use strict" /* keep */; run();`, FileName: "review.mjs", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
		}},
		{Code: "\"use strict\" // keep\n; run();", FileName: "review.js", Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 2, EndColumn: 2},
		}},

		// CommonJS accepts these local names, but an ES module import cannot bind them.
		{Code: `const await = require("x");
const yield = require("y");
const eval = require("z");
const arguments = require("w");`, FileName: "review.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22},
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 2, Column: 15, EndLine: 2, EndColumn: 22},
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 3, Column: 14, EndLine: 3, EndColumn: 21},
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 4, Column: 19, EndLine: 4, EndColumn: 26},
		}},
		{Code: `const {yield} = require("x");`, FileName: "review.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 24},
		}},
		{Code: `const {foo: \u0061\u0077\u0061\u0069\u0074} = require("x");`, FileName: "review.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 47, EndLine: 1, EndColumn: 54},
		}},
		{Code: `const {await: wait, yield: async} = require("x");`, FileName: "review.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 37, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
				{MessageId: "suggestion/import", Output: `import {await as wait, yield as async} from "x";`},
			}},
		}},
	}
	for i := range valid {
		if valid[i].Globals == nil {
			valid[i].Globals = commonJSGlobals()
		}
	}
	for i := range invalid {
		if invalid[i].Globals == nil {
			invalid[i].Globals = commonJSGlobals()
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_module.PreferModuleRule, valid, invalid)

}

// The compiler's program file selection excludes mixed-case extensions. Parse
// them directly so that upstream's case-insensitive filename gate is tested.
func TestPreferModuleFilenameSuffix(t *testing.T) {
	for _, filename := range []string{"/example.cjs", "/example.cjS", "/example.CJS"} {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: filename, Path: tspath.Path(filename)}, `"use strict"; require("x"); __dirname;`, core.ScriptKindJS)
		if listeners := prefer_module.PreferModuleRule.Run(rule.RuleContext{SourceFile: file}, nil); len(listeners) != 0 {
			t.Errorf("%s: expected no listeners", filename)
		}
	}
}

func TestPreferModuleArtifactsFollowDemand(t *testing.T) {
	for _, test := range []struct {
		code, filename, output string
		suggestions            int
	}{
		{`"use strict"; run();`, "/demand.MJS", "run();", 0},
		{`"use strict"; run();`, "/demand.js", "", 1},
		{`const file = __filename;`, "/demand.js", "", 2},
		{`const dir = __dirname;`, "/demand.js", "", 2},
		{`const file = require("x");`, "/demand.js", "", 1},
		{`module.exports = value;`, "/demand.js", "", 1},
		{`exports.value = value;`, "/demand.js", "", 1},
		{`let file = require("x");`, "/demand.js", "", 0},
	} {
		t.Run(test.code+test.filename, func(t *testing.T) {
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: test.filename, Path: tspath.Path(test.filename)}, test.code, core.ScriptKindJS)
			globals := make(map[string]utils.GlobalAccess)
			for name := range commonJSGlobals() {
				globals[name] = utils.GlobalAccessReadonly
			}
			init, refsInit, language := rule.ResolveLanguageDefaults(test.filename, rule.LanguageOptions{})
			demands := []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll}
			diagnostics := make([]rule.RuleDiagnostic, len(demands))
			for index, demand := range demands {
				comments := rule.NewCommentStore(file)
				var found []rule.RuleDiagnostic
				ctx := rule.RuleContext{
					SourceFile: file, Comments: comments,
					Refs:           rule.NewRefStore(file, &core.CompilerOptions{}, nil, refsInit),
					Globals:        rule.NewGlobals(language, init, globals, nil, nil),
					DisableManager: rule.NewDisableManager(file, comments),
				}.WithDiagnosticConsumer(prefer_module.PreferModuleRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
					Demand: demand, Report: func(d rule.RuleDiagnostic) { found = append(found, d) },
				})
				listeners := prefer_module.PreferModuleRule.Run(ctx, nil)
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return node.ForEachChild(visit)
				}
				visit(file.AsNode())
				if len(found) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(found))
				}
				diagnostics[index] = found[0]
			}
			all := diagnostics[len(diagnostics)-1]
			count := 0
			if all.Suggestions != nil {
				count = len(*all.Suggestions)
			}
			if count != test.suggestions {
				t.Fatalf("got %d suggestions, want %d", count, test.suggestions)
			}
			for index, demand := range demands {
				diagnostic := diagnostics[index]
				if diagnostic.Range != all.Range || !reflect.DeepEqual(diagnostic.Message, all.Message) {
					t.Errorf("demand %d changed diagnostic identity", demand)
				}
				wantFix := test.output != "" && demand&rule.EditDemandAutofix != 0
				wantSuggestion := test.suggestions > 0 && demand&rule.EditDemandSuggestion != 0
				if (diagnostic.FixesPtr != nil) != wantFix || (diagnostic.Suggestions != nil) != wantSuggestion {
					t.Errorf("demand %d: unexpected edit artifacts", demand)
				}
				if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) || wantSuggestion && !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Errorf("demand %d changed edit artifacts", demand)
				}
			}
			if test.output != "" {
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{all})
				if !fixed || output != test.output {
					t.Errorf("autofix = %q, want %q", output, test.output)
				}
			}
		})
	}
}
