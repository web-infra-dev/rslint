package url_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Reference tracking and tsgo/ESTree shapes beyond the upstream suite.
// Expectations were checked with eslint-plugin-n v18.3.0 and typescript-eslint 8.65.0.
func TestPreferGlobalURLExtras(t *testing.T) {
	runURLTests(t,
		[]rule_tester.ValidTestCase{
			{
				Code:     "require('url'); require('url').URLSearchParams; require('other').URL",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "function f(require, process) { require('url').URL; process.getBuiltinModule('url').URL; }",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:            "import('url').then(url => url.URL)",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const key = 'URL'; require('url')[key]; require('url')[unknown]",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "class C { #URL; m() { return require('url').#URL; } }",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const URL = Custom; new URL(s); function f(URL) { return new URL(s); }",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "URL = Custom; new URL(s)",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "new URL(s)",
				Options:  []any{"never"},
				Globals:  map[string]any{"URL": "off"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "/** @type {URL} */ let value;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const { URL, URLSearchParams } = require('node:url'); new URL(s);",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:            "import { URL } from 'node:url'; new URL(s);",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:            "import type { URL } from 'node:url'; type T = URL;",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.ts",
			},
			{
				Code:     "/* global URL: off */ new URL(s)",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "require('url').url; url.URL;",
				Options:  []any{"always"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:            "import 'url'; import { parse } from 'url'; export { parse } from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     "const module = require('url'); const alias = module; new alias.URL(s); alias['URL'];",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 58, 1, 67),
					urlError("preferGlobal", 1, 72, 1, 84),
				},
			},
			{
				Code:     "require('url')['U' + 'RL']; process.getBuiltinModule('node:url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 27),
					urlError("preferGlobal", 1, 29, 1, 69),
				},
			},
			{
				Code:     "const { URL: LocalURL = fallback } = require('url');",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 9, 1, 33),
				},
			},
			{
				Code:     "let LocalURL; ({ URL: LocalURL } = require('url'));",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 18, 1, 31),
				},
			},
			{
				Code:     "const { URL } = process.getBuiltinModule('url');",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 9, 1, 12),
				},
			},
			{
				Code:     "require?.('url')?.URL; (require('url')).URL; (require('url')?.URL);",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 22),
					urlError("preferGlobal", 1, 24, 1, 44),
					urlError("preferGlobal", 1, 47, 1, 66),
				},
			},
			{
				Code:     "const { URL } = globalThis.require('url'); global.process.getBuiltinModule('url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 9, 1, 12),
					urlError("preferGlobal", 1, 44, 1, 86),
				},
			},
			{
				Code:     "require('url').URL = replacement;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 19),
				},
			},
			{
				Code:            "import { URL as LocalURL } from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 10, 1, 25),
				},
			},
			{
				Code:            "import url from 'node:url'; new url.URL(s);",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 33, 1, 40),
				},
			},
			{
				Code:            "import * as url from 'url'; url.URL; url.default.URL;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 29, 1, 36),
					urlError("preferGlobal", 1, 38, 1, 53),
				},
			},
			{
				Code:            "export { URL as LocalURL } from 'node:url'; export * from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 10, 1, 25),
					urlError("preferGlobal", 1, 45, 1, 65),
				},
			},
			{
				Code:            "export * as url from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js", TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 28),
				},
			},
			{
				Code:            "import type { URL } from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 15, 1, 18),
				},
			},
			{
				Code:     "(require('url') as typeof import('url')).URL;",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 45),
				},
			},
			{
				Code:     "require('url')!.URL; require('url') satisfies object;",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 20),
				},
			},
			{
				Code:     "const symbol = '😀'; require('url').URL;\nconst {\n  URL: LocalURL\n} = require('url');",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 22, 1, 40),
					urlError("preferGlobal", 3, 3, 3, 16),
				},
			},
			{
				Code:     "new (URL)(s); const LocalURL = URL; new LocalURL(s);",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 6, 1, 9),
					urlError("preferModule", 1, 32, 1, 35),
				},
			},
			{
				Code:     "globalThis.URL; global['URL']; self.URL; window.URL;",
				Options:  []any{"never"},
				Globals:  map[string]any{"self": "readonly", "window": "readonly"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 1, 1, 15),
					urlError("preferModule", 1, 17, 1, 30),
					urlError("preferModule", 1, 32, 1, 40),
					urlError("preferModule", 1, 42, 1, 52),
				},
			},
			{
				Code:     "const root = globalThis; root.URL; const { URL: LocalURL } = global;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 26, 1, 34),
					urlError("preferModule", 1, 44, 1, 57),
				},
			},
			{
				Code:     "type T = URL; type U = typeof URL; class C extends URL {}",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 10, 1, 13),
					urlError("preferModule", 1, 31, 1, 34),
					urlError("preferModule", 1, 52, 1, 55),
				},
			},
			{
				Code:     "const view = <URL />; const member = <globalThis.URL />;",
				Options:  []any{"never"},
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 15, 1, 18),
				},
			},
			{
				Code:     "function f() { return URL; } { let URL; new URL(s); }",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 23, 1, 26),
				},
			},
		},
	)
}

// Additional scope, module and syntax boundaries checked against the pinned upstream rule.
func TestPreferGlobalURLBoundaries(t *testing.T) {
	runURLTests(t,
		[]rule_tester.ValidTestCase{
			{
				Code:            "export { default as URL } from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js",
				TSConfig:        "tsconfig.allowJs.json",
			},
			{
				Code:     "const { ...rest } = require('url'); rest.URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const module = 'url'; require(module).URL; require(); require(123).URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:     "const { [key]: U, ...rest } = require('url'); rest.URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
			},
			{
				Code:            "var URL; new URL(s);",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				FileName:        "input.js",
				TSConfig:        "tsconfig.allowJs.json",
			},
			{
				Code:            "function URL() {} new URL(s);",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				FileName:        "input.js",
				TSConfig:        "tsconfig.allowJs.json",
			},
			{
				Code:    "declare const URL: unknown; new URL(s);",
				Options: []any{"never"},
			},
			{
				Code:    "type URL = string; new URL(s);",
				Options: []any{"never"},
			},
			{
				Code:    "interface URL { href: string }; new URL(s);",
				Options: []any{"never"},
			},
			{
				Code:            "import URL = require('url'); URL.URL;",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:            "import { 'URL' as U } from 'node:url'; export { 'URL' as U2 } from 'url';",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				FileName:        "input.js",
				TSConfig:        "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 10, 1, 20),
					urlError("preferGlobal", 1, 49, 1, 60),
				},
			},
			{
				Code:     "let load = require; load('url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 21, 1, 36),
				},
			},
			{
				Code:     "const load = process.getBuiltinModule; load('url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 40, 1, 55),
				},
			},
			{
				Code:     "function f({ URL: U } = require('url')) {}",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 14, 1, 20),
				},
			},
			{
				Code:     "(flag ? require('url') : require('node:url')).URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 50),
					urlError("preferGlobal", 1, 1, 1, 50),
				},
			},
			{
				Code:     "(flag && require('url')).URL; (0, require('url')).URL; (require('url'), other).URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 29),
					urlError("preferGlobal", 1, 31, 1, 54),
				},
			},
			{
				Code:     "require(`url`)[`URL`]; require('u' + 'rl')['U\\u0052L'];",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 22),
					urlError("preferGlobal", 1, 24, 1, 55),
				},
			},
			{
				Code:     "delete require('url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 8, 1, 26),
				},
			},
			{
				Code:     "globalThis.URL = Custom; globalThis.URL;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 1, 1, 15),
					urlError("preferModule", 1, 26, 1, 40),
				},
			},
			{
				Code:     "({ URL } = globalThis); new URL(s);",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 4, 1, 7),
				},
			},
			{
				Code:     "{ let globalThis; globalThis.URL; } globalThis.URL;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 37, 1, 51),
				},
			},
			{
				Code:     "globalThis?.['URL']; typeof URL;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 1, 1, 20),
					urlError("preferModule", 1, 29, 1, 32),
				},
			},
			{
				Code:     "(/** @type {URL} */ (globalThis)).URL;",
				Options:  []any{"never"},
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 1, 1, 38),
				},
			},
			{
				Code:     "(/** @type {typeof import('url')} */ (require('url'))).URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 1, 1, 1, 59),
				},
			},
			{
				Code:    "namespace N { export const u = new URL(s); }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferModule", 1, 36, 1, 39),
				},
			},
			{
				// RunRuleTester registers the rule as "test".
				Code:     "// eslint-disable-next-line test\nrequire('url').URL;\nrequire('url').URL;",
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlError("preferGlobal", 3, 1, 3, 19),
				},
			},
		},
	)
}
