// Upstream tests and documentation: eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-module.js
// All cases are retained in their upstream groups.
package prefer_module_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_module"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferModuleUpstream(t *testing.T) {
	t.Run("Strict mode", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "const foo = \"use strict\";", FileName: "upstream.js"},
			{Code: "(\"use strict\")", FileName: "upstream.js"},
			{Code: "\"use strong\";", FileName: "upstream.js"},
			{Code: "eval(\"'use strict'; var x = 42; x;\");", FileName: "upstream.js"},
			{Code: "new Function(\"'use strict'; var x = 42; return x;\");", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "'use strict';\n\nconsole.log(1);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/use-strict-directive", Output: "console.log(1);"},
						}},
				}},
			{Code: "\"use strict\";\n\nconsole.log(1);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/use-strict-directive", Output: "console.log(1);"},
						}},
				}},
			{Code: "function foo () {\n\t\"use strict\";\n\tconsole.log(1);\n}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 2, Column: 2, EndLine: 2, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/use-strict-directive", Output: "function foo () {\n\tconsole.log(1);\n}"},
						}},
				}},
			{Code: "'use strict';\n\nconsole.log(1);", FileName: "example.mjs", Output: []string{"console.log(1);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
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

	})
	t.Run("Global return", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "function a() {\n\treturn;\n}", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "if (foo) {\n\treturn;\n}", FileName: "upstream.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/global-return", Message: "\"return\" should be used inside a function.", Line: 2, Column: 2, EndLine: 2, EndColumn: 8},
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

	})
	t.Run("__dirname and __filename", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "const __filename = 1;\nconst foo = __filename;", FileName: "upstream.js"},
			{Code: "const __dirname = 1;\nconst foo = __dirname;", FileName: "upstream.js"},
			{Code: "import {__filename as filename} from \"foo.mjs\"", FileName: "upstream.js"},
			{Code: "const foo = 1;export {foo as __dirname}", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "const dirname = __dirname;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "const dirname = import.meta.dirname;"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "const dirname = path.dirname(url.fileURLToPath(import.meta.url));"},
						}},
				}},
			{Code: "const dirname = __filename;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "const dirname = import.meta.filename;"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "const dirname = url.fileURLToPath(import.meta.url);"},
						}},
				}},
			{Code: "const foo = { __dirname};", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "const foo = { __dirname: import.meta.dirname};"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "const foo = { __dirname: path.dirname(url.fileURLToPath(import.meta.url))};"},
						}},
				}},
			{Code: "const foo = {__filename, };", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "const foo = {__filename: import.meta.filename, };"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "const foo = {__filename: url.fileURLToPath(import.meta.url), };"},
						}},
				}},
			{Code: "if (__dirname.startsWith(\"/project/src/\")) {}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "if (import.meta.dirname.startsWith(\"/project/src/\")) {}"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "if (path.dirname(url.fileURLToPath(import.meta.url)).startsWith(\"/project/src/\")) {}"},
						}},
				}},
			{Code: "if (__filename.endsWith(\".js\")) {}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "if (import.meta.filename.endsWith(\".js\")) {}"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "if (url.fileURLToPath(import.meta.url).endsWith(\".js\")) {}"},
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

	})
	t.Run("require", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "import {createRequire} from 'module';\nconst require = createRequire(import.meta.url);\nconst foo = require(\"foo\");", FileName: "upstream.js"},
			{Code: "const require = function require(require, id) {\n\treturn require(id);\n}", FileName: "upstream.js"},
			{Code: "const require = class A {\n\trequire(require) {\n\t\treturn require(id);\n\t}\n}", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import \"foo\";"},
						}},
				}},
			{Code: "require('foo');", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import 'foo';"},
						}},
				}},
			{Code: "require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import  \"foo\" ;"},
						}},
				}},
			{Code: "((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 3, EndLine: 1, EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import \"foo\";"},
						}},
				}},
			{Code: "(( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 4, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: " import \"foo\" ;"},
						}},
				}},
			{Code: "const foo=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from \"foo\";"},
						}},
				}},
			{Code: "const foo = require.resolve(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				}},
			{Code: "const foo\n\t=\n\trequire(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 3, Column: 2, EndLine: 3, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from \"foo\";"},
						}},
				}},
			{Code: "const foo = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from \"foo\";"},
						}},
				}},
			{Code: "const foo = require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from  \"foo\" ;"},
						}},
				}},
			{Code: "const foo = ((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from \"foo\";"},
						}},
				}},
			{Code: "const foo = (( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 16, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import foo from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo}=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo}\n\t=\n\trequire(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 3, Column: 2, EndLine: 3, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo} = require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo} = ((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo} = (( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo: foo}=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo: foo}\n\t=\n\trequire(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 3, Column: 2, EndLine: 3, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo: foo} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 20, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo: foo} = require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 20, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo: foo} = ((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 22, EndLine: 1, EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from \"foo\";"},
						}},
				}},
			{Code: "const {foo: foo} = (( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 23, EndLine: 1, EndColumn: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as foo} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo:bar}=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 17, EndLine: 1, EndColumn: 24,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from \"foo\";"},
						}},
				}},
			{Code: "const {foo:bar}\n\t=\n\trequire(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 3, Column: 2, EndLine: 3, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from \"foo\";"},
						}},
				}},
			{Code: "const {foo:bar} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from \"foo\";"},
						}},
				}},
			{Code: "const {foo:bar} = require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {foo:bar} = ((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 21, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from \"foo\";"},
						}},
				}},
			{Code: "const {foo:bar} = (( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 22, EndLine: 1, EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {a   :foo, b:   bar, default   :   baz}=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 47, EndLine: 1, EndColumn: 54,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {a as foo, b as bar, default as baz} from \"foo\";"},
						}},
				}},
			{Code: "const {\n\ta   :foo,\n\tb:   bar,\n\tdefault   :   baz,\n}\n\t=\n\trequire(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 7, Column: 2, EndLine: 7, EndColumn: 9,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {\n\ta as foo,\n\tb as bar,\n\tdefault as baz,\n} from \"foo\";"},
						}},
				}},
			{Code: "const {a   :foo, b:   bar, default   :   baz} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 49, EndLine: 1, EndColumn: 56,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {a as foo, b as bar, default as baz} from \"foo\";"},
						}},
				}},
			{Code: "const {a   :foo, b:   bar, default   :   baz} = require( ((\"foo\")) );", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 49, EndLine: 1, EndColumn: 56,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {a as foo, b as bar, default as baz} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {a   :foo, b:   bar, default   :   baz} = ((require))(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 51, EndLine: 1, EndColumn: 58,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {a as foo, b as bar, default as baz} from \"foo\";"},
						}},
				}},
			{Code: "const {a   :foo, b:   bar, default   :   baz} = (( require(\"foo\") ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 52, EndLine: 1, EndColumn: 59,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {a as foo, b as bar, default as baz} from  \"foo\" ;"},
						}},
				}},
			{Code: "const {} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 12, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {} from \"foo\";"},
						}},
				}},
			{Code: "const{   }=require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 12, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import{   } from \"foo\";"},
						}},
				}},
			{Code: "const r = require;\nconst foo = r(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				}},
			{Code: "new require(\"foo\")", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				}},
			{Code: "require(\"foo\", extraArgument)", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "const a = require()", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				}},
			{Code: "require(...\"foo\")", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "require(\"../\" + \"file.js\")", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "require(file)", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "a = require(\"foo\")", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				}},
			{Code: "function a(a = require(\"foo\")) {}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 16, EndLine: 1, EndColumn: 23},
				}},
			{Code: "let foo = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 11, EndLine: 1, EndColumn: 18},
				}},
			{Code: "const foo = require(\"foo\"), bar = require(\"bar\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 35, EndLine: 1, EndColumn: 42},
				}},
			{Code: "const {[foo]: foo} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 22, EndLine: 1, EndColumn: 29},
				}},
			{Code: "const {[\"foo\"]: foo} = require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 24, EndLine: 1, EndColumn: 31},
				}},
			{Code: "if (foo) require(\"foo\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				}},
			{Code: "const foo = require`foo`;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				}},
			{Code: "function loadModule() {\n\treturn require(\"foo\");\n}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 2, Column: 9, EndLine: 2, EndColumn: 16},
				}},
			{Code: "function loadModule() {\n\tconst foo = require(\"foo\");\n\treturn foo;\n}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 2, Column: 14, EndLine: 2, EndColumn: 21},
				}},
			{Code: "const foo = require(\"foo\"), bar = 1;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
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

	})
	t.Run("exports and module", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "function exports(exports) {}", FileName: "upstream.js"},
			{Code: "function module(module) {}", FileName: "upstream.js"},
			{Code: "const exports = foo;\nexports.bar = bar;", FileName: "upstream.js"},
			{Code: "const exports = 1;", FileName: "upstream.js"},
			{Code: "const module = foo;\nmodule.exports = bar;\nmodule.exports.bar = bar;", FileName: "upstream.js"},
			{Code: "const module = 1;", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "exports = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export default foo;"},
						}},
				}},
			{Code: "module.exports = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export default foo;"},
						}},
				}},
			{Code: "(( ((exports)) = ((foo)) ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: " export default  ((foo)) ;"},
						}},
				}},
			{Code: "(( ((module.exports)) = ((foo)) ));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 12,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: " export default  ((foo)) ;"},
						}},
				}},
			{Code: "const foo = 1;exports.foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "const foo = 1;export {foo as foo};"},
						}},
				}},
			{Code: "const foo = 1;module.exports.foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "const foo = 1;export {foo as foo};"},
						}},
				}},
			{Code: "exports[\"foo\"] = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "module.exports[\"foo\"] = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				}},
			{Code: "const foo = exports;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				}},
			{Code: "const foo = exports.foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				}},
			{Code: "const foo = module.exports;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				}},
			{Code: "const foo = module.exports.foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				}},
			{Code: "module[\"exports\"] = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				}},
			{Code: "module[exports] = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				}},
			{Code: "module.exports.foo.bar = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				}},
			{Code: "const foo = 1;exports.default = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "const foo = 1;export {foo as default};"},
						}},
				}},
			{Code: "const foo = 1;module.exports.default = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "const foo = 1;export {foo as default};"},
						}},
				}},
			{Code: "exports.foo.bar = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "exports = 1;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export default 1;"},
						}},
				}},
			{Code: "exports.foo = [];", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "module.exports = function() {};", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export default function() {};"},
						}},
				}},
			{Code: "module.exports.foo = foo || bar;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				}},
			{Code: "exports += foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				}},
			{Code: "const foo = module.children", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
				}},
			{Code: "const parentModule = module.parent", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
				}},
			{Code: "function foo() {\n\texports.foo = foo;\n\tmodule.exports.foo = foo;\n}", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 2, Column: 2, EndLine: 2, EndColumn: 9},
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 3, Column: 2, EndLine: 3, EndColumn: 8},
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

	})
	t.Run("TypeScript", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "type module = number[];", FileName: "upstream.ts"},
			{Code: "type ModuleRegistry = { [module: string]: string };", FileName: "upstream.ts"},
			{Code: "const module = 1; type ModuleRegistry = { [module: string]: string };", FileName: "upstream.ts"},
			{Code: "type module = number[]; type ModuleRegistry = { [module: string]: string };", FileName: "upstream.ts"},
			{Code: "type Data = { [module in keyof string]: number; };", FileName: "upstream.ts"},
			{Code: "type ModuleRegistry = { [exports: string]: string };", FileName: "upstream.ts"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "const foo: Foo = require(\"foo\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 18, EndLine: 1, EndColumn: 25},
				}},
			{Code: "const {foo}: Foo = require(\"foo\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 20, EndLine: 1, EndColumn: 27},
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

	})
	t.Run("JSON modules", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "const data = require(\"./a.json\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import data from \"./a.json\" with {type: \"json\"};"},
						}},
				}},
			{Code: "require(\"./a.json\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import \"./a.json\" with {type: \"json\"};"},
						}},
				}},
			{Code: "const data = (require(('./a.json')));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import data from './a.json' with {type: 'json'};"},
						}},
				}},
			{Code: "const {foo} = require(\"./a.json\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 15, EndLine: 1, EndColumn: 22},
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

	})
	t.Run(".cjs files", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "__dirname", FileName: "foo.cjs"},
			{Code: "__dirname", FileName: "foo.cjS",
				// Compiler file selection rejects .cjS; extras directly check the rule's suffix handling.
				Skip: true},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "__filename", FileName: "foo.mjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 11,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "import.meta.filename"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "url.fileURLToPath(import.meta.url)"},
						}},
				}},
			{Code: "require(\"lodash\")", FileName: "foo.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import \"lodash\""},
						}},
				}},
			{Code: "require(\"lodash\")", FileName: "foo.cjs/foo.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import \"lodash\""},
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

	})
	t.Run("Comments in named exports", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "const foo = 1;\nexports /* note */ .foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				}},
			{Code: "const foo = 1;\nexports./* c */foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				}},
			{Code: "const foo = 1;\nexports.foo /* c */ = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				}},
			{Code: "const foo = 1;\nmodule.exports /* c */ .foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 2, Column: 1, EndLine: 2, EndColumn: 7},
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

	})
	t.Run("Constructor callees", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "new __filename.constructor(\"text\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "new import.meta.filename.constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url)).constructor(\"text\");"},
						}},
				}},
			{Code: "new (__filename).constructor(\"text\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename).constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url)).constructor(\"text\");"},
						}},
				}},
			{Code: "new __filename.constructor(\"text\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename).constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url)).constructor(\"text\");"},
						}},
				}},
			{Code: "new __filename!.constructor(\"text\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "new (import.meta.filename)!.constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "new (url.fileURLToPath(import.meta.url))!.constructor(\"text\");"},
						}},
				}},
			{Code: "new __dirname.constructor(\"text\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "new import.meta.dirname.constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url))).constructor(\"text\");"},
						}},
				}},
			{Code: "new (__dirname).constructor(\"text\");", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 6, EndLine: 1, EndColumn: 15,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname).constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url))).constructor(\"text\");"},
						}},
				}},
			{Code: "new __dirname.constructor(\"text\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname).constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url))).constructor(\"text\");"},
						}},
				}},
			{Code: "new __dirname!.constructor(\"text\");", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 5, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "new (import.meta.dirname)!.constructor(\"text\");"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "new (path.dirname(url.fileURLToPath(import.meta.url)))!.constructor(\"text\");"},
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

	})
	t.Run("Static template module sources", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "const file = require(`file`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from 'file';"},
						}},
				}},
			{Code: "require(`file`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import 'file';"},
						}},
				}},
			{Code: "const file = require(`./file.json`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from './file.json' with {type: 'json'};"},
						}},
				}},
			{Code: "const {foo: bar} = require(`file`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 20, EndLine: 1, EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {foo as bar} from 'file';"},
						}},
				}},
			{Code: "const file = require(`./it's.js`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from './it\\'s.js';"},
						}},
				}},
			{Code: "const file = require(`./line\nfile.js`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from './line\\nfile.js';"},
						}},
				}},
			{Code: "const file = require(`./back\\`tick.js`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from './back`tick.js';"},
						}},
				}},
			{Code: "const file = require(`./fi\\u006ce`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from './file';"},
						}},
				}},
			{Code: "const file = require(/* keep */ (`file`));", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import file from /* keep */ 'file';"},
						}},
				}},
			{Code: "const file = require(`${name}`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				}},
			{Code: "const file = require(tag`file`);", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				}},
			{Code: "const file: Foo = require(`file`);", FileName: "upstream.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 19, EndLine: 1, EndColumn: 26},
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

	})
	t.Run("Documentation examples", func(t *testing.T) {
		valid := []rule_tester.ValidTestCase{
			{Code: "const __dirname = import.meta.dirname;\nconst __filename = import.meta.filename;", FileName: "upstream.js"},
			{Code: "import {fileURLToPath} from 'node:url';\nimport path from 'node:path';\nconst __filename = fileURLToPath(import.meta.url);\nconst __dirname = path.dirname(fileURLToPath(import.meta.url));", FileName: "upstream.js"},
			{Code: "import {fileURLToPath} from 'node:url';\nconst foo = fileURLToPath(new URL('foo.js', import.meta.url));", FileName: "upstream.js"},
			{Code: "const foo = new URL('foo.js', import.meta.url);", FileName: "upstream.js"},
			{Code: "function run() {\n\tif (foo) {\n\t\treturn;\n\t}\n}\nrun();", FileName: "upstream.js"},
			{Code: "const file = path.join(import.meta.dirname, 'foo.js');", FileName: "upstream.js"},
			{Code: "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');", FileName: "upstream.js"},
			{Code: "const content = fs.readFileSync(import.meta.filename, 'utf8');", FileName: "upstream.js"},
			{Code: "import {fromPairs} from 'lodash-es';", FileName: "upstream.js"},
			{Code: "export default foo;", FileName: "upstream.js"},
			{Code: "const foo = 1;export {foo};", FileName: "upstream.js"},
		}
		invalid := []rule_tester.InvalidTestCase{
			{Code: "'use strict';", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/use-strict-directive", Message: "Do not use \"use strict\" directive.", Line: 1, Column: 1, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/use-strict-directive", Output: ""},
						}},
				}},
			{Code: "if (foo) {\n\treturn;\n}", FileName: "upstream.js", LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/global-return", Message: "\"return\" should be used inside a function.", Line: 2, Column: 2, EndLine: 2, EndColumn: 8},
				}},
			{Code: "const file = path.join(__dirname, 'foo.js');", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__dirname\".", Line: 1, Column: 24, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-dirname", Output: "const file = path.join(import.meta.dirname, 'foo.js');"},
							{MessageId: "suggestion/import-meta-url-to-dirname", Output: "const file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), 'foo.js');"},
						}},
				}},
			{Code: "const content = fs.readFileSync(__filename, 'utf8');", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"__filename\".", Line: 1, Column: 33, EndLine: 1, EndColumn: 43,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import-meta-filename", Output: "const content = fs.readFileSync(import.meta.filename, 'utf8');"},
							{MessageId: "suggestion/import-meta-url-to-filename", Output: "const content = fs.readFileSync(url.fileURLToPath(import.meta.url), 'utf8');"},
						}},
				}},
			{Code: "const {fromPairs} = require('lodash');", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"require\".", Line: 1, Column: 21, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/import", Output: "import {fromPairs} from 'lodash';"},
						}},
				}},
			{Code: "module.exports = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"module\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export default foo;"},
						}},
				}},
			{Code: "exports.foo = foo;", FileName: "upstream.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "error/identifier", Message: "Do not use \"exports\".", Line: 1, Column: 1, EndLine: 1, EndColumn: 8,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "suggestion/export", Output: "export {foo as foo};"},
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

	})
}
func commonJSGlobals() map[string]any {
	return map[string]any{
		"exports": "readonly", "require": "readonly", "module": "readonly", "__filename": "readonly", "__dirname": "readonly",
	}
}
