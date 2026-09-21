package url_search_params_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional cases compared with eslint-plugin-n v18.3.0 and @typescript-eslint/parser v8.65.0.
func TestURLSearchParamsExtras(t *testing.T) {
	runURLSearchParamsTests(t,
		[]rule_tester.ValidTestCase{
			// Other modules, other exports, and dynamic names are ignored.
			{
				Code: "require('other').URLSearchParams; require('url').URL; require(moduleName).URLSearchParams; require('url')[key]; process.getBuiltinModule(moduleName).URLSearchParams;",
			},
			// Locally shadowed loaders are ignored.
			{
				Code: "function f(require, process) { require('url').URLSearchParams; process.getBuiltinModule('url').URLSearchParams; }",
			},
			// Disabled loaders are ignored.
			{
				Code:    "require('url').URLSearchParams; process.getBuiltinModule('url').URLSearchParams;",
				Globals: map[string]any{"require": "off", "process": "off"},
			},
			// Global loader writes disable tracking throughout the file.
			{
				Code: "require('url').URLSearchParams; require = load; process.getBuiltinModule('url').URLSearchParams; process = other;",
			},
			// Dynamic imports and imported process do not root tracking.
			{
				Code: "import process from 'node:process'; process.getBuiltinModule('url').URLSearchParams; const url = await import('node:url'); url.URLSearchParams;",
			},
			// Local constructor bindings shadow the global.
			{
				Code:    "function f(URLSearchParams) { return new URLSearchParams(); } { const URLSearchParams = Custom; new URLSearchParams(); }",
				Options: []any{"never"},
			},
			// A disabled bare global is ignored.
			{
				Code:    "new URLSearchParams();",
				Options: []any{"never"},
				Globals: map[string]any{"URLSearchParams": "off"},
			},
			// A global write disables tracking throughout the file.
			{
				Code:    "new URLSearchParams(); URLSearchParams = Custom;",
				Options: []any{"never"},
			},
			// Rest bindings and private properties do not select the export.
			{
				Code: "const {...url} = require('url'); url.URLSearchParams; class C { #URLSearchParams; m() { require('url').#URLSearchParams; } }",
			},
			// Computed identifiers are not resolved from declarations.
			{
				Code: "const key = 'URLSearchParams'; require('url')[key];",
			},
			// JSX member tags do not read tracked properties.
			{
				Code:     "import url from 'node:url'; const el = <url.URLSearchParams />;",
				FileName: "input.tsx",
			},
			// Import-equals and qualified type queries do not create module reads.
			{
				Code:     "import url = require('url'); url.URLSearchParams; import type * as ns from 'url'; type T = typeof ns.URLSearchParams;",
				FileName: "input.ts",
			},
			// A script var declaration shadows the configured global.
			{
				Code:            "var URLSearchParams; new URLSearchParams();",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			},
			// Empty and spread loader arguments are not statically known.
			{
				Code: "require(); process.getBuiltinModule(); require(...['url']).URLSearchParams; process.getBuiltinModule(...['url']).URLSearchParams;",
			},
			// Global object writes disable tracking throughout the file.
			{
				Code:    "globalThis.URLSearchParams; globalThis = other;",
				Options: []any{"never"},
			},
			// Plain object keys do not read the global.
			{
				Code:    "const obj = {URLSearchParams: 1}; obj.URLSearchParams;",
				Options: []any{"never"},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Parentheses, optional chains, and static computed properties preserve ranges.
			{
				Code: "(require('url'))?.URLSearchParams; require('node:url')[('URL' + 'SearchParams')]; process?.getBuiltinModule?.('node:url')?.URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 1, 1, 34),
					urlSearchParamsAt("preferGlobal", 1, 36, 1, 81),
					urlSearchParamsAt("preferGlobal", 1, 83, 1, 139),
				},
			},
			// Destructuring bindings, assignments, and parameter defaults report their property.
			{
				Code: "const {URLSearchParams: Params = fallback} = process.getBuiltinModule('node:url');\nlet P; ({URLSearchParams: P} = require('url'));\nfunction f({URLSearchParams: S} = require('url')) { return S; }",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 8, 1, 42),
					urlSearchParamsAt("preferGlobal", 2, 10, 2, 28),
					urlSearchParamsAt("preferGlobal", 3, 13, 3, 31),
				},
			},
			// Named, default, and namespace imports expose the CommonJS export.
			{
				Code: "import { URLSearchParams as Params } from 'node:url';\nimport url from 'url'; import * as ns from 'node:url';\nurl.URLSearchParams; ns.URLSearchParams; ns.default.URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 10, 1, 35),
					urlSearchParamsAt("preferGlobal", 3, 1, 3, 20),
					urlSearchParamsAt("preferGlobal", 3, 22, 3, 40),
					urlSearchParamsAt("preferGlobal", 3, 42, 3, 68),
				},
			},
			// Re-exports report their specifier or complete export-all statement.
			{
				Code: "export {URLSearchParams as Params} from 'url';\nexport * from 'node:url';\nexport * as url from 'url';\nexport {'URLSearchParams' as SearchParams} from 'node:url';",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 9, 1, 34),
					urlSearchParamsAt("preferGlobal", 2, 1, 2, 26),
					urlSearchParamsAt("preferGlobal", 3, 1, 3, 28),
					urlSearchParamsAt("preferGlobal", 4, 9, 4, 42),
				},
			},
			// Aliased and computed loaders follow static module names.
			{
				Code: "const {require: load} = globalThis; load('node:' + 'url').URLSearchParams;\nconst p = globalThis['pro' + 'cess']; p['getBuiltinModule']('url').URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 37, 1, 74),
					urlSearchParamsAt("preferGlobal", 2, 39, 2, 83),
				},
			},
			// Cyclic aliases terminate and independent conditional paths retain reports.
			{
				Code: "let a = require('url'); let b = a; a = b; b.URLSearchParams;\nconst url = flag ? require('url') : require('node:url'); url.URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 43, 1, 60),
					urlSearchParamsAt("preferGlobal", 2, 58, 2, 77),
					urlSearchParamsAt("preferGlobal", 2, 58, 2, 77),
				},
			},
			// Unicode and multiline member reads use UTF-16 columns.
			{
				Code:    "\"😀\"; require(\"url\")\n  .URLSearchParams;\nrequire(\"url\")[\"URLSearchParam\\u0073\"];",
				Options: []any{"always"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 7, 2, 19),
					urlSearchParamsAt("preferGlobal", 3, 1, 3, 39),
				},
			},
			// Global object aliases and destructuring remain tracked with the bare global disabled.
			{
				Code:    "const root = globalThis; root?.URLSearchParams; const {URLSearchParams: Params} = global; new Params();",
				Options: []any{"never"},
				Globals: map[string]any{"URLSearchParams": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 26, 1, 47),
					urlSearchParamsAt("preferModule", 1, 56, 1, 79),
				},
			},
			// Bare reads, shorthand values, and computed keys read the global; constructor aliases do not.
			{
				Code:    "const Params = URLSearchParams; new Params();\nconst value = {URLSearchParams, [URLSearchParams]: URLSearchParams};",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 16, 1, 31),
					urlSearchParamsAt("preferModule", 2, 16, 2, 31),
					urlSearchParamsAt("preferModule", 2, 34, 2, 49),
					urlSearchParamsAt("preferModule", 2, 52, 2, 67),
				},
			},
			// Member assignments and deletions still read module properties.
			{
				Code: "require('url').URLSearchParams = Other; delete require('url').URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 1, 1, 31),
					urlSearchParamsAt("preferGlobal", 1, 48, 1, 78),
				},
			},
			// Global object writes and configured browser roots remain tracked.
			{
				Code:    "global.URLSearchParams = Other; delete globalThis.URLSearchParams;\nwindow.URLSearchParams; self[\"URLSearchParams\"];",
				Options: []any{"never"},
				Globals: map[string]any{"window": "readonly", "self": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 1, 1, 23),
					urlSearchParamsAt("preferModule", 1, 40, 1, 66),
					urlSearchParamsAt("preferModule", 2, 1, 2, 23),
					urlSearchParamsAt("preferModule", 2, 25, 2, 48),
				},
			},
			// JavaScript JSX reports the opening tag and expression containers.
			{
				Code:     "const el = <URLSearchParams value={URLSearchParams}>{new URLSearchParams()}</URLSearchParams>;",
				Options:  []any{"never"},
				FileName: "input.jsx",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 13, 1, 28),
					urlSearchParamsAt("preferModule", 1, 36, 1, 51),
					urlSearchParamsAt("preferModule", 1, 58, 1, 73),
				},
			},
			// TypeScript JSX also references the closing component tag.
			{
				Code:     "const el = <URLSearchParams value={URLSearchParams}>{new URLSearchParams()}</URLSearchParams>;",
				Options:  []any{"never"},
				FileName: "input.tsx",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 13, 1, 28),
					urlSearchParamsAt("preferModule", 1, 36, 1, 51),
					urlSearchParamsAt("preferModule", 1, 58, 1, 73),
					urlSearchParamsAt("preferModule", 1, 78, 1, 93),
				},
			},
			// Type assertions and heritage clauses preserve module reads.
			{
				Code:     "(require('url') as any).URLSearchParams; require('url')!.URLSearchParams; (require('url') satisfies any).URLSearchParams;\nimport url from 'url'; class C extends url.URLSearchParams {} interface I extends url.URLSearchParams {} class D implements url.URLSearchParams {}",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 1, 1, 40),
					urlSearchParamsAt("preferGlobal", 1, 42, 1, 73),
					urlSearchParamsAt("preferGlobal", 1, 75, 1, 121),
					urlSearchParamsAt("preferGlobal", 2, 40, 2, 59),
					urlSearchParamsAt("preferGlobal", 2, 83, 2, 102),
					urlSearchParamsAt("preferGlobal", 2, 125, 2, 144),
				},
			},
			// Type-only imports and re-exports still read the named export.
			{
				Code:     "import type { URLSearchParams } from 'url'; export type { URLSearchParams as P } from 'node:url'; export type * from 'url';",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 15, 1, 30),
					urlSearchParamsAt("preferGlobal", 1, 59, 1, 79),
					urlSearchParamsAt("preferGlobal", 1, 99, 1, 124),
				},
			},
			// Direct type references read the global; qualified type queries do not read a property.
			{
				Code:     "type P = URLSearchParams; type T = typeof URLSearchParams; type G = typeof globalThis.URLSearchParams; interface I extends URLSearchParams {}",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 10, 1, 25),
					urlSearchParamsAt("preferModule", 1, 43, 1, 58),
					urlSearchParamsAt("preferModule", 1, 124, 1, 139),
				},
			},
			// Value declarations do not shadow type namespace references.
			{
				Code:     "type P = typeof URLSearchParams; interface I { params: URLSearchParams } declare const URLSearchParams: any; new URLSearchParams();",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 56, 1, 71),
				},
			},
			// JSDoc casts preserve module references.
			{
				Code: "(/** @type {any} */ (require(\"url\"))).URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 1, 1, 54),
				},
			},
			// Inline global declarations enable the bare global and preserve Unicode ranges.
			{
				Code:    "/* global URLSearchParams */\n\"😀\"; URLSearchParams;\nglobalThis\n  [\"URLSearchParams\"];",
				Options: []any{"never"},
				Globals: map[string]any{"URLSearchParams": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 2, 7, 2, 22),
					urlSearchParamsAt("preferModule", 3, 1, 4, 22),
				},
			},
			// Parameter defaults and outer reads remain global despite body-local declarations.
			{
				Code:    "function f(value = URLSearchParams) { var URLSearchParams; }\nfunction g() { new URLSearchParams(); var URLSearchParams; } new URLSearchParams();",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 20, 1, 35),
					urlSearchParamsAt("preferModule", 2, 66, 2, 81),
				},
			},
			// Reports from imports and loaders are ordered by source position.
			{
				Code: "import {URLSearchParams} from 'node:url';\nprocess.getBuiltinModule('url').URLSearchParams;\nrequire('url').URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 9, 1, 24),
					urlSearchParamsAt("preferGlobal", 2, 1, 2, 48),
					urlSearchParamsAt("preferGlobal", 3, 1, 3, 31),
				},
			},
			// Escaped bare identifiers and global object names use their decoded names.
			{
				Code:    "new URLSearchParam\\u0073(); globalTh\\u0069s.URLSearchParam\\u0073;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 5, 1, 25),
					urlSearchParamsAt("preferModule", 1, 29, 1, 65),
				},
			},
			// String-named imports select the URLSearchParams export.
			{
				Code: "import { 'URLSearchParams' as Params } from 'node:url'; new Params();",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 10, 1, 37),
				},
			},
			// Comma and logical expressions preserve only possible module values.
			{
				Code: "(0, require('url')).URLSearchParams; (require('url'), other).URLSearchParams; (flag && require('url')).URLSearchParams; (flag || require('url')).URLSearchParams; (flag ?? require('url')).URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 1, 1, 36),
					urlSearchParamsAt("preferGlobal", 1, 79, 1, 119),
					urlSearchParamsAt("preferGlobal", 1, 121, 1, 161),
					urlSearchParamsAt("preferGlobal", 1, 163, 1, 203),
				},
			},
			// Static blocks do not shadow global reads in class methods.
			{
				Code:    "class C { static { const URLSearchParams = Other; new URLSearchParams(); } method() { return new URLSearchParams(); } }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 98, 1, 113),
				},
			},
			// Global object shadowing applies only within its function.
			{
				Code:    "function f(globalThis) { return globalThis.URLSearchParams; } globalThis.URLSearchParams;",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 63, 1, 89),
				},
			},
			// Global annotations can disable bare references without disabling member reads.
			{
				Code:    "/* global URLSearchParams: off */ new URLSearchParams(); globalThis?.[\"URLSearchParams\"];",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 58, 1, 89),
				},
			},
			// Named default imports expose the module export.
			{
				Code: "import {default as url} from 'url'; url.URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 37, 1, 56),
				},
			},
			// Namespace-local declarations preserve outside global references.
			{
				Code:     "namespace Local { export class URLSearchParams {} new URLSearchParams(); } new URLSearchParams();",
				Options:  []any{"never"},
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 80, 1, 95),
				},
			},
			// A nested binding default can read the export.
			{
				Code: "const {url = require('url')} = source; url.URLSearchParams;",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 40, 1, 59),
				},
			},
			// Annotated bindings preserve references to the module.
			{
				Code:     "const url: any = require('url'); const {URLSearchParams: Params}: any = url;",
				FileName: "input.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferGlobal", 1, 41, 1, 64),
				},
			},
			// Global references survive a with scope.
			{
				Code:            "with (other) { new URLSearchParams(); }",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 20, 1, 35),
				},
			},
			// Deleting an unbound global in script mode remains a read.
			{
				Code:            "delete URLSearchParams;",
				Options:         []any{"never"},
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 8, 1, 23),
				},
			},
			// Class fields and computed method names read the global.
			{
				Code:    "class C { [URLSearchParams] = URLSearchParams; static [URLSearchParams]() { return URLSearchParams; } }",
				Options: []any{"never"},
				Errors: []rule_tester.InvalidTestCaseError{
					urlSearchParamsAt("preferModule", 1, 12, 1, 27),
					urlSearchParamsAt("preferModule", 1, 31, 1, 46),
					urlSearchParamsAt("preferModule", 1, 56, 1, 71),
					urlSearchParamsAt("preferModule", 1, 84, 1, 99),
				},
			},
		},
	)
}
